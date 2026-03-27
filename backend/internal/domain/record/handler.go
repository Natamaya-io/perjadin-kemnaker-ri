package record

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/config"
	"github.com/kemnaker/perjadin-backend/internal/domain/user"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/kemnaker/perjadin-backend/internal/utils"
	"github.com/kemnaker/perjadin-backend/internal/utils/document"
	"github.com/kemnaker/perjadin-backend/internal/utils/terbilang"
	"github.com/labstack/echo/v4"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Handler struct {
	svc      Service
	userRepo user.Repository
	cfg      *config.Config
	docGen   *document.Generator
}

func NewHandler(s Service, userRepo user.Repository, cfg *config.Config) *Handler {
	gotenbergURL := os.Getenv("GOTENBERG_URL")
	if gotenbergURL == "" {
		gotenbergURL = "http://gotenberg:3000"
	}

	return &Handler{
		svc:      s,
		userRepo: userRepo,
		cfg:      cfg,
		docGen:   document.NewGenerator(gotenbergURL, "templates"),
	}
}

func (h *Handler) notifyEmployee(record *models.TravelRecord) {
	if h.cfg.Fonnte.Token == "" {
		return
	}

	u, err := h.userRepo.GetUserByID(record.EmployeeID)
	if err != nil || u == nil || u.NomorHP == "" {
		return
	}

	locationsStr := ""
	if len(record.Locations) > 0 {
		for _, loc := range record.Locations {
			locationsStr += fmt.Sprintf("\n📍 *Tujuan:* %s, %s\n📅 *Tanggal:* %s s/d %s",
				loc.Location, loc.Province, loc.StartDate.Format("02 Jan 2006"), loc.EndDate.Format("02 Jan 2006"))
		}
	} else {
		locationsStr = fmt.Sprintf("\n📍 *Tujuan:* %s, %s\n📅 *Tanggal:* %s s/d %s",
			record.Location, record.Province, record.StartDate.Format("02 Jan 2006"), record.EndDate.Format("02 Jan 2006"))
	}

	message := fmt.Sprintf("Halo *%s*,\n\nAnda telah ditugaskan untuk melaksanakan perjalanan dinas dengan rincian sebagai berikut:\n%s\n\n🎯 *Kegiatan:* %s\n\nHarap persiapkan diri Anda dan cek aplikasi untuk detail lebih lanjut.\n\n_Pesan ini dikirim otomatis oleh Sistem Perjadin Protokol Kemnaker RI_",
		u.Name, locationsStr, record.Purpose)

	err = utils.SendWhatsAppMessage(h.cfg, u.NomorHP, message)
	if err != nil {
		fmt.Printf("Failed to send WhatsApp message to %s: %v\n", u.NomorHP, err)
	} else {
		fmt.Printf("Successfully sent WhatsApp notification to %s\n", u.NomorHP)
	}
}

func (h *Handler) mapTravelToDocument(record *models.TravelRecord, globalIndex int) map[string]interface{} {
	// mapTravelToDocument maps TravelRecord data to placeholders used in DOCX templates.
	// Ensure keys match the {{placeholder}} names in 'templates/Berkas Luar Kota - SPD.docx'
	titleCaser := cases.Title(language.Indonesian)
	
	formatDate := func(t time.Time) string {
		if t.IsZero() {
			return "-"
		}
		months := []string{"", "Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
		return fmt.Sprintf("%d %s %d", t.Day(), months[int(t.Month())], t.Year())
	}

	days := int(record.EndDate.Sub(record.StartDate).Hours()/24) + 1
	
	dest := record.Location
	prov := record.Province
	if len(record.Locations) > 0 {
		var locs []string
		for _, l := range record.Locations {
			locs = append(locs, titleCaser.String(l.Location))
		}
		dest = strings.Join(locs, " & ")
		prov = titleCaser.String(record.Locations[0].Province)
	}

	transportMode := "-"
	if record.Cost != nil && record.Cost.TransportMode != "" {
		transportMode = record.Cost.TransportMode
	}

	// Roman numeral months
	romanMonths := []string{"", "I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X", "XI", "XII"}
	spdSubNumber := fmt.Sprintf("%03d", globalIndex)

	vars := map[string]interface{}{
		"no_spd":               spdSubNumber,
		"no_surat":            spdSubNumber,
		"bulan":               romanMonths[int(record.StartDate.Month())],
		"tahun":               record.StartDate.Year(),
		"nama":                record.Employee.Name,
		"nama_petugas":        record.Employee.Name,
		"nip":                 record.Employee.NIP,
		"nip_petugas":         record.Employee.NIP,
		"pangkat_gol":         fmt.Sprintf("%s (%s)", record.Employee.Pangkat, record.Employee.Golongan),
		"pangkat":             record.Employee.Pangkat,
		"golongan":            record.Employee.Golongan,
		"jabatan":             record.Employee.Jabatan,
		"tingkat_biaya":       record.Employee.TingkatBiaya,
		"maksud_perjalanan":   record.Purpose,
		"nama_stakeholder":    record.Stakeholder,
		"tujuan":              dest,
		"kota_atau_kabupaten": dest,
		"provinsi":            prov,
		"transportasi":        transportMode,
		"tgl_berangkat":       formatDate(record.StartDate),
		"tanggal_berangkat":   formatDate(record.StartDate),
		"tgl_kembali":         formatDate(record.EndDate),
		"tanggal_selesai":     formatDate(record.EndDate),
		"lama_hari":           fmt.Sprintf("%d ( %s )", days, terbilang.FormatTerbilang(days)),
		"lama_perjalanan":     days,
		"terbilang":           terbilang.FormatTerbilang(days),
		"tgl_surat_tugas":     formatDate(record.SuratTugasDate),
		"tanggal_dikeluarkan": formatDate(record.SuratTugasDate),
		"tgl_cetak":           formatDate(time.Now()),
		"nama_ppk":            "Arief Hafidiyanto",
		"nip_ppk":             "19720827 200312 1 002",
		"nama ppk":            "Arief Hafidiyanto",
		"nip ppk":             "19720827 200312 1 002",
		"keterangan":          "-",
	}

	if record.Report != nil {
		vars["isi_laporan"] = record.Report.Text
		vars["tgl_laporan"] = formatDate(record.Report.SubmittedAt)
	}

	if record.Cost != nil {
		total := record.TotalCost
		vars["total_biaya"] = fmt.Sprintf("Rp %s", utils.FormatRupiah(total))
		vars["terbilang_biaya"] = fmt.Sprintf("%s Rupiah", titleCaser.String(terbilang.FormatTerbilang(int(total))))
		vars["biaya_harian"] = fmt.Sprintf("Rp %s", utils.FormatRupiah(record.Cost.DailyAllowanceRate * float64(record.Cost.DailyAllowanceDays)))
		vars["biaya_hotel"] = fmt.Sprintf("Rp %s", utils.FormatRupiah(record.Cost.HotelRate * float64(record.Cost.HotelDays)))
		vars["biaya_pesawat"] = fmt.Sprintf("Rp %s", utils.FormatRupiah(record.Cost.TicketGo + record.Cost.TicketBack))
	}

	return vars
}

func (h *Handler) exportDocument(c echo.Context, templateName, prefix string) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Format UUID tidak valid")
	}

	record, err := h.svc.GetRecordByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Data tidak ditemukan")
	}

	// Compute global index using the same deterministic sort as the frontend admin page
	allRecords, err := h.svc.GetRecords(map[string]interface{}{})
	globalIndex := 0
	if err == nil && len(allRecords) > 0 {
		sort.Slice(allRecords, func(i, j int) bool {
			ti := allRecords[i].CreatedAt.Unix()
			tj := allRecords[j].CreatedAt.Unix()
			if ti != tj {
				return ti < tj
			}
			si := allRecords[i].SPDNumber
			sj := allRecords[j].SPDNumber
			if si != sj {
				return si < sj
			}
			return allRecords[i].ID.String() < allRecords[j].ID.String()
		})
		for idx, r := range allRecords {
			if r.ID == record.ID {
				globalIndex = idx + 1
				break
			}
		}
	}
	if globalIndex == 0 {
		globalIndex = 1 // fallback
	}

	payload := document.DocumentRequest{
		TemplateName: templateName,
		Variables:    h.mapTravelToDocument(record, globalIndex),
	}

	pdfBytes, err := h.docGen.Generate(c.Request().Context(), payload)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Gagal membuat dokumen: %v", err))
	}

	filename := fmt.Sprintf("%s_%s_%s.pdf", prefix, record.Employee.Name, record.SPDNumber)
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Response().Header().Set(echo.HeaderContentType, "application/pdf")

	return c.Blob(http.StatusOK, "application/pdf", pdfBytes)
}

// ExportSpdPDF handles the request to export a TravelRecord as an SPD PDF.
// It uses the docx template and converts it to PDF via Gotenberg.
func (h *Handler) ExportSpdPDF(c echo.Context) error {
	return h.exportDocument(c, "Berkas Luar Kota - SPD.docx", "SPD")
}

func (h *Handler) ExportLaporanPDF(c echo.Context) error {
	return h.exportDocument(c, "Berkas Luar Kota - Laporan.docx", "Laporan")
}

func (h *Handler) ExportRincianPDF(c echo.Context) error {
	return h.exportDocument(c, "Berkas Luar Kota - rincian pembayaran.docx", "Rincian")
}

func (h *Handler) UploadFile(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "File not found in request")
	}

	src, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer src.Close()

	ext := filepath.Ext(file.Filename)
	filename := uuid.New().String() + "_" + time.Now().Format("20060102150405") + ext

	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create upload directory")
	}

	dstPath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"path": filename,
	})
}

func (h *Handler) CreateRecord(c echo.Context) error {
	var record models.TravelRecord
	if err := c.Bind(&record); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	var creatorID uuid.UUID
	creatorIDInterface := c.Get("user_id")
	if creatorIDInterface != nil {
		if creatorIDStr, ok := creatorIDInterface.(string); ok {
			if id, err := uuid.Parse(creatorIDStr); err == nil {
				creatorID = id
			}
		}
	}

	if record.SPDNumber == "" {
		spjNumber, err := h.svc.GenerateSpdNumber()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate SPJ number")
		}
		record.SPDNumber = spjNumber
	}

	if len(record.EmployeeIDs) > 0 {
		var createdRecords []models.TravelRecord
		for _, empID := range record.EmployeeIDs {
			newRecord := record
			newRecord.ID = uuid.Nil
			newRecord.EmployeeID = empID
			newRecord.CreatorID = creatorID
			newRecord.EmployeeIDs = nil
			
			if len(record.Locations) > 0 {
				newRecord.Locations = make([]models.TravelLocation, len(record.Locations))
				copy(newRecord.Locations, record.Locations)
			}

			if err := h.svc.CreateRecord(&newRecord); err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			}
			go h.notifyEmployee(&newRecord)
			createdRecords = append(createdRecords, newRecord)
		}
		return c.JSON(http.StatusCreated, createdRecords)
	}

	record.CreatorID = creatorID
	if err := h.svc.CreateRecord(&record); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	go h.notifyEmployee(&record)

	return c.JSON(http.StatusCreated, record)
}

func (h *Handler) GetRecords(c echo.Context) error {
	filters := make(map[string]interface{})
	status := c.QueryParam("status")
	if status != "" {
		filters["status"] = status
	}

	records, err := h.svc.GetRecords(filters)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, records)
}

func (h *Handler) GetRecordByID(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid UUID format")
	}

	record, err := h.svc.GetRecordByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Record not found")
	}

	return c.JSON(http.StatusOK, record)
}

func (h *Handler) UpdateRecord(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid UUID format")
	}

	record, err := h.svc.GetRecordByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Record not found")
	}

	if err := c.Bind(&record); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	if err := h.svc.UpdateRecord(record); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, record)
}

func (h *Handler) DeleteRecord(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid UUID format")
	}

	if err := h.svc.DeleteRecord(id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) DeleteRecordsBySpd(c echo.Context) error {
	spd := c.Param("spd")
	if spd == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "SPD number is required")
	}

	if err := h.svc.DeleteRecordsBySpd(c.Request().Context(), spd); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Successfully deleted records for SPD " + spd,
	})
}
