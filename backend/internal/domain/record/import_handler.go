package record

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/labstack/echo/v4"
	"github.com/xuri/excelize/v2"
)

// ImportResult holds the summary of an import operation
type ImportResult struct {
	TotalRows int            `json:"totalRows"`
	Imported  int            `json:"imported"`
	Skipped   int            `json:"skipped"`
	Failed    int            `json:"failed"`
	Details   []ImportDetail `json:"details"`
}

type ImportDetail struct {
	Row     int    `json:"row"`
	SPJID   string `json:"spjId"`
	Name    string `json:"name"`
	Status  string `json:"status"` // "imported", "skipped_duplicate", "skipped_thr", "skipped_no_user", "failed"
	Message string `json:"message,omitempty"`
}

// ImportExcel handles bulk import of travel records from an Excel file
func (h *Handler) ImportExcel(c echo.Context) error {
	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "File tidak ditemukan. Silakan upload file .xlsx")
	}

	// Validate file extension
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".xlsx") {
		return echo.NewHTTPError(http.StatusBadRequest, "Format file harus .xlsx")
	}

	src, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Gagal membuka file")
	}
	defer src.Close()

	// Parse Excel
	f, err := excelize.OpenReader(src)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Gagal membaca file Excel: %v", err))
	}
	defer f.Close()

	// Get the first sheet
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "File Excel kosong")
	}

	// Use RawCellValue to bypass local formatting like "Rp", "." and "," for thousands/decimals
	rows, err := f.GetRows(sheetName, excelize.Options{RawCellValue: true})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Gagal membaca data: %v", err))
	}

	if len(rows) < 3 {
		return echo.NewHTTPError(http.StatusBadRequest, "File Excel tidak memiliki data (minimal 3 baris: header + data)")
	}

	// Get creator ID from JWT (the logged-in superadmin)
	creatorID := uuid.Nil
	if creatorIDStr, ok := c.Get("user_id").(string); ok && creatorIDStr != "" {
		creatorID, _ = uuid.Parse(creatorIDStr)
	}
	if creatorID == uuid.Nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User tidak terautentikasi")
	}

	// Load all users and build name->UUID map (case-insensitive)
	allUsers, err := h.userRepo.GetUsers()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Gagal memuat data pegawai")
	}
	userMap := make(map[string]uuid.UUID)
	for _, u := range allUsers {
		normalizedName := strings.TrimSpace(strings.ToLower(u.Name))
		userMap[normalizedName] = u.ID
	}

	// Load existing records to check duplicates (by SPD number + employee name)
	existingRecords, err := h.svc.GetRecords(map[string]interface{}{})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Gagal memuat data record existing")
	}
	existingSet := make(map[string]bool)
	for _, r := range existingRecords {
		key := fmt.Sprintf("%s|%s", strings.ToLower(r.SPDNumber), strings.ToLower(r.Employee.Name))
		existingSet[key] = true
	}

	result := ImportResult{
		Details: []ImportDetail{},
	}

	// Process each data row (skip row 1=empty, row 2=headers, start from row 3 = index 2)
	for rowIdx := 2; rowIdx < len(rows); rowIdx++ {
		row := rows[rowIdx]
		result.TotalRows++

		// Helper to safely get cell value
		getCell := func(colIdx int) string {
			if colIdx < len(row) {
				return strings.TrimSpace(row[colIdx])
			}
			return ""
		}

		// Column mapping (0-indexed):
		// A=0: ID SPJ, B=1: No, C=2: Nama, D=3: Daerah
		// F=5: Tiket, G=6: Jumlah Hari, I=8: Uang Harian
		// K=10: Jml Hari Menginap, L=11: Hotel
		// N=13: Transport Lokal, O=14: Transport Daerah, P=15: Jumlah
		// R=17: KETERANGAN, S=18: Provinsi
		// V=21: Alat Angkut, W=22: Mulai, X=23: Selesai
		// AB=27: Status

		idSPJ := getCell(0)
		if idSPJ == "" {
			continue // Skip empty rows
		}

		name := getCell(2)
		daerah := getCell(3)
		statusCol := getCell(27)
		province := getCell(18)
		purpose := getCell(17)
		transportMode := getCell(21)

		rowNum := rowIdx + 1 // 1-indexed for user display

		// Skip THR records
		if strings.ToUpper(strings.TrimSpace(statusCol)) == "THR" {
			result.Skipped++
			result.Details = append(result.Details, ImportDetail{
				Row: rowNum, SPJID: idSPJ, Name: name,
				Status: "skipped_thr", Message: "Record THR, dilewati",
			})
			continue
		}

		// Check if user exists
		normalizedName := strings.TrimSpace(strings.ToLower(name))
		employeeID, userFound := userMap[normalizedName]
		if !userFound {
			result.Skipped++
			result.Details = append(result.Details, ImportDetail{
				Row: rowNum, SPJID: idSPJ, Name: name,
				Status: "skipped_no_user", Message: fmt.Sprintf("Pegawai '%s' tidak ditemukan di database", name),
			})
			continue
		}

		// Check duplicates
		dupKey := fmt.Sprintf("%s|%s", strings.ToLower(idSPJ), normalizedName)
		if existingSet[dupKey] {
			result.Skipped++
			result.Details = append(result.Details, ImportDetail{
				Row: rowNum, SPJID: idSPJ, Name: name,
				Status: "skipped_duplicate", Message: "Data sudah ada di database",
			})
			continue
		}

		// Parse dates - since we use RawCellValue, we can actually read the raw unformatted serial or standard text
		colW, _ := excelize.ColumnNumberToName(23)
		colX, _ := excelize.ColumnNumberToName(24)
		startDate := readExcelDate(f, sheetName, fmt.Sprintf("%s%d", colW, rowNum))
		endDate := readExcelDate(f, sheetName, fmt.Sprintf("%s%d", colX, rowNum))

		if startDate.IsZero() || endDate.IsZero() {
			result.Failed++
			result.Details = append(result.Details, ImportDetail{
				Row: rowNum, SPJID: idSPJ, Name: name,
				Status: "failed", Message: "Tanggal mulai atau selesai tidak valid",
			})
			continue
		}

		// Parse cost fields
		ticketCost := safeParseFloat(getCell(5))      // F: Tiket
		dailyRate := safeParseFloat(getCell(8))        // I: Uang Harian
		dailyDays := safeParseInt(getCell(6))          // G: Jumlah Hari
		hotelNights := safeParseInt(getCell(10))       // K: Jml Hari Menginap
		hotelRate := safeParseFloat(getCell(11))       // L: Hotel
		localTransport := safeParseFloat(getCell(13))  // N: Transport Lokal
		regionalTransport := safeParseFloat(getCell(14)) // O: Transport Daerah
		totalCost := safeParseFloat(getCell(15))       // P: Jumlah

		// Check if cell has a background fill color (e.g. red highlight = incomplete)
		cellA := fmt.Sprintf("A%d", rowNum)
		styleID, err := f.GetCellStyle(sheetName, cellA)
		
		recordStatus := "Approved"
		reportStatus := "Completed"
		paymentStatus := "Paid"
		
		if err == nil {
			style, _ := f.GetStyle(styleID)
			// Pattern 1 is solid fill
			if style != nil && style.Fill.Pattern == 1 && len(style.Fill.Color) > 0 {
				color := strings.ToUpper(style.Fill.Color[0])
				// Ignore white/black
				if !strings.HasSuffix(color, "FFFFFF") && !strings.HasSuffix(color, "000000") {
					recordStatus = "Assigned"
					reportStatus = "Pending"
					paymentStatus = "Pending"
				}
			}
		}

		// Create the travel record (bypass service.CreateRecord to skip overlap check & WA notification)
		record := &models.TravelRecord{
			SPDNumber:     idSPJ,
			EmployeeID:    employeeID,
			CreatorID:     creatorID,
			StartDate:     startDate,
			EndDate:       endDate,
			Location:      daerah,
			Province:      province,
			Type:          "luar_kota",
			Purpose:       purpose,
			Stakeholder:   "",
			Agenda:        "",
			Status:        recordStatus,
			ReportStatus:  reportStatus,
			PaymentStatus: paymentStatus,
			TotalCost:     totalCost,
			Locations: []models.TravelLocation{
				{
					Location:  daerah,
					Province:  province,
					StartDate: startDate,
					EndDate:   endDate,
				},
			},
			Cost: &models.TravelCost{
				TicketGo:           ticketCost,
				DailyAllowanceDays: dailyDays,
				DailyAllowanceRate: dailyRate,
				HotelDays:          hotelNights,
				HotelRate:          hotelRate,
				LocalTransport:     localTransport,
				RegionalTransport:  regionalTransport,
				TransportMode:      transportMode,
			},
		}

		// Use CreateRecordDirect to bypass overlap check and WA notification
		if err := h.svc.CreateRecordDirect(record); err != nil {
			result.Failed++
			result.Details = append(result.Details, ImportDetail{
				Row: rowNum, SPJID: idSPJ, Name: name,
				Status: "failed", Message: fmt.Sprintf("Gagal menyimpan: %v", err),
			})
			continue
		}

		// Mark as existing to prevent duplicate within same import batch
		existingSet[dupKey] = true

		result.Imported++
		result.Details = append(result.Details, ImportDetail{
			Row: rowNum, SPJID: idSPJ, Name: name,
			Status: "imported", Message: "Berhasil diimport",
		})
	}

	// Invalidate cache once at the end
	h.svc.InvalidateAllCache()

	return c.JSON(http.StatusOK, result)
}

// readExcelDate reads a date value from an Excel cell reference (e.g. "W3")
func readExcelDate(f *excelize.File, sheet, cellRef string) time.Time {
	rawVal, err := f.GetCellValue(sheet, cellRef, excelize.Options{RawCellValue: true})
	if err != nil || rawVal == "" {
		return time.Time{}
	}

	// Try to parse as Excel serial date number
	serial, err := strconv.ParseFloat(rawVal, 64)
	if err == nil && serial > 0 {
		t, err := excelize.ExcelDateToTime(serial, false)
		if err == nil {
			return t
		}
	}

	formattedVal, _ := f.GetCellValue(sheet, cellRef)
	if formattedVal == "" {
		return time.Time{}
	}

	formats := []string{
		"01-02-06",
		"1/2/2006",
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"02-Jan-2006",
	}
	for _, layout := range formats {
		if t, err := time.Parse(layout, formattedVal); err == nil {
			return t
		}
	}

	return time.Time{}
}

// safeParseFloat safely parses a string to float64
func safeParseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" {
		return 0
	}
	
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return val
}

// safeParseInt safely parses a string to int
func safeParseInt(s string) int {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" {
		return 0
	}
	
	val, err := strconv.Atoi(s)
	if err != nil {
		// Try parsing as float first (Excel might return "2.0" for integer cells)
		f, ferr := strconv.ParseFloat(s, 64)
		if ferr == nil {
			return int(f)
		}
		return 0
	}
	return val
}
