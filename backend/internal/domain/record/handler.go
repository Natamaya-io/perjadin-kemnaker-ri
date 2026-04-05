package record

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/config"
	"github.com/kemnaker/perjadin-backend/internal/domain/master"
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
	svc       Service
	userRepo  user.Repository
	masterSvc master.Service
	cfg       *config.Config
	docGen    *document.Generator
}

func NewHandler(s Service, userRepo user.Repository, cfg *config.Config, masterSvc master.Service) *Handler {
	gotenbergURL := os.Getenv("GOTENBERG_URL")
	if gotenbergURL == "" {
		gotenbergURL = "http://gotenberg:3000"
	}

	return &Handler{
		svc:       s,
		userRepo:  userRepo,
		masterSvc: masterSvc,
		cfg:       cfg,
		docGen:    document.NewGenerator(gotenbergURL, "templates"),
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
	titleCaser := cases.Title(language.Indonesian)
	days := 0
	if !record.StartDate.IsZero() && !record.EndDate.IsZero() {
		days = int(record.EndDate.Sub(record.StartDate).Hours()/24) + 1
	}
	dest := record.Location
	prov := record.Province
	if len(record.Locations) > 0 {
		var locs []string
		var provs []string
		seenProvs := make(map[string]bool)
		for _, l := range record.Locations {
			locs = append(locs, titleCaser.String(l.Location))
			pName := titleCaser.String(l.Province)
			if pName != "" && !seenProvs[pName] {
				provs = append(provs, pName)
				seenProvs[pName] = true
			}
		}
		dest = strings.Join(locs, " & ")
		prov = strings.Join(provs, " & ")
	}
	transportMode := "-"
	if record.Cost != nil && record.Cost.TransportMode != "" {
		transportMode = record.Cost.TransportMode
	}
	spdSubNumber := fmt.Sprintf("%03d", globalIndex)
	extractNumericID := func(id string) string {
		if id == "" { return spdSubNumber }
		parts := strings.Split(id, "-")
		if len(parts) > 0 {
			suffix := parts[len(parts)-1]
			if len(suffix) < 3 { return fmt.Sprintf("%03s", suffix) }
			return suffix
		}
		return id
	}
	noSurat := record.SuratTugasNumber
	tglSurat := record.SuratTugasDate
	if noSurat == "" {
		noSurat = spdSubNumber
	} else {
		groupID := extractNumericID(record.SPDNumber)
		if groupID != "" && groupID != spdSubNumber {
			noSurat = strings.Replace(noSurat, groupID, spdSubNumber, 1)
		}
	}
	noSpd := spdSubNumber
	cleanMaksud := regexp.MustCompile(`\s+`).ReplaceAllString(record.Purpose, " ")
	cleanStakeholder := regexp.MustCompile(`\s+`).ReplaceAllString(record.Stakeholder, " ")
	namaPpk := h.cfg.Signatory.PPKName
	nipPpk := h.cfg.Signatory.PPKNIP
	namaBendahara := h.cfg.Signatory.BendaharaName
	nipBendahara := h.cfg.Signatory.BendaharaNIP
	jabPpk := "Pejabat Pembuat Komitmen"
	jabBendahara := "Bendahara Pengeluaran Pembantu"
	if h.masterSvc != nil {
		gs, _ := h.masterSvc.GetSettings(context.Background())
		if gs["ppk_name"] != "" { namaPpk = gs["ppk_name"] }
		if gs["ppk_nip"] != "" { nipPpk = gs["ppk_nip"] }
		if gs["bendahara_name"] != "" { namaBendahara = gs["bendahara_name"] }
		if gs["bendahara_nip"] != "" { nipBendahara = gs["bendahara_nip"] }
	}
	if record.Report != nil {
		if record.Report.PPKName != "" { namaPpk = record.Report.PPKName }
		if record.Report.PPKNIP != "" { nipPpk = record.Report.PPKNIP }
		if record.Report.BendaharaName != "" { namaBendahara = record.Report.BendaharaName }
		if record.Report.BendaharaNIP != "" { nipBendahara = record.Report.BendaharaNIP }
	}
	if record.Cost == nil {
		record.Cost = &models.TravelCost{DailyAllowanceRate: 0, DailyAllowanceDays: days}
	}
	tglCetak := tglSurat
	if tglCetak.IsZero() { tglCetak = time.Now() }
	tanggalMulaiStr := "-"
	if !record.StartDate.IsZero() && !record.EndDate.IsZero() {
		if record.StartDate.Month() == record.EndDate.Month() && record.StartDate.Year() == record.EndDate.Year() {
			tanggalMulaiStr = fmt.Sprintf("%d", record.StartDate.Day())
		} else {
			tanggalMulaiStr = utils.FormatIndonesianDate(record.StartDate)
		}
	}
	isiLaporan := "-"
	if record.Report != nil && record.Report.Text != "" {
		isiLaporan = record.Report.Text
	}

	vars := map[string]interface{}{
		"no_spd": noSpd, "id_spj": noSpd, "no_surat": noSurat,
		"bulan_no_surat": utils.GetRomanMonths()[int(tglSurat.Month())], "tahun_no_surat": tglSurat.Year(),
		"bulan_romawi": utils.GetRomanMonths()[int(tglSurat.Month())],
		"tanggal_no_surat": utils.FormatIndonesianDate(tglSurat), "bulan_pembayaran": utils.GetIndonesianMonths()[int(tglCetak.Month())],
		"tahun_pembayaran": tglCetak.Year(), "tahun_saat_ini": fmt.Sprintf("%d", time.Now().Year()),
		"nama": record.Employee.Name, "nama_petugas": record.Employee.Name, "nip": record.Employee.NIP, "nip_petugas": record.Employee.NIP,
		"pangkat_gol": fmt.Sprintf("%s (%s)", record.Employee.Pangkat, record.Employee.Golongan), "jabatan": record.Employee.Jabatan,
		"maksud_perjalanan": cleanMaksud, "tujuan_perjalanan": cleanMaksud, "stakeholder": cleanStakeholder,
		"tujuan": dest, "kota": dest, "provinsi": prov, "transportasi": transportMode,
		"tanggal_berangkat": utils.FormatIndonesianDate(record.StartDate), "tanggal_mulai": tanggalMulaiStr,
		"tanggal_selesai": utils.FormatIndonesianDate(record.EndDate), "lama_hari": days,
		"tgl_cetak": utils.FormatIndonesianDate(tglCetak), "nama_ppk": namaPpk, "nip_ppk": nipPpk,
		"nama_bendahara": namaBendahara, "nip_bendahara": nipBendahara, "jabatan_ppk": jabPpk, "jabatan_bendahara": jabBendahara,
		"isi_laporan": isiLaporan, "tanggal_dikeluarkan": utils.FormatIndonesianDate(tglCetak),
	}
	type Detail struct {
		TransportMode string `json:"transportMode"`
		TicketGo float64 `json:"ticketGo"`
		TicketBack float64 `json:"ticketBack"`
		HotelDays int `json:"hotelDays"`
		HotelRate float64 `json:"hotelRate"`
		TransportAmount float64 `json:"transportAmount"`
		AdditionalCosts []struct {
			Name string `json:"name"`
			Amount float64 `json:"amount"`
			TicketGo float64 `json:"ticketGo"`
			TicketBack float64 `json:"ticketBack"`
		} `json:"additionalCosts"`
	}
	var costDetails []Detail
	if record.Cost != nil && len(record.Cost.Details) > 0 { json.Unmarshal(record.Cost.Details, &costDetails) }
	ordinals := []string{"satu", "dua", "tiga", "empat", "lima", "enam", "tujuh", "delapan", "sembilan", "sepuluh"}
	for i, ordinal := range ordinals {
		if i < len(record.Locations) {
			loc := record.Locations[i]
			locDays := int(loc.EndDate.Sub(loc.StartDate).Hours()/24) + 1
			vars[fmt.Sprintf("tujuan_%s", ordinal)] = titleCaser.String(loc.Location)
			vars[fmt.Sprintf("hari_%s", ordinal)] = locDays
			var sbmRate float64
			if record.Cost != nil { sbmRate = record.Cost.DailyAllowanceRate }
			vars[fmt.Sprintf("sbm_rate_%s", ordinal)] = utils.FormatRupiah(sbmRate)
			if i < len(costDetails) {
				cd := costDetails[i]
				vars[fmt.Sprintf("hotel_rate_%s", ordinal)] = utils.FormatRupiah(cd.HotelRate)
				vars[fmt.Sprintf("tiket_%s", ordinal)] = utils.FormatRupiah(cd.TicketGo + cd.TicketBack)
			}
		}
	}
	if record.Cost != nil {
		var aggTicket, aggLokal, aggHotel, aggTambahan float64
		var totalDays, hotelDays int
		for _, cd := range costDetails {
			aggTicket += cd.TicketGo + cd.TicketBack
			aggLokal += cd.TransportAmount
			aggHotel += float64(cd.HotelDays) * cd.HotelRate
			hotelDays += cd.HotelDays
			for _, ac := range cd.AdditionalCosts {
				if ac.Name == "Extend Tiket" {
					aggTicket += ac.TicketGo + ac.TicketBack
					aggTambahan += ac.Amount
				} else { aggTambahan += ac.Amount }
			}
		}
		for _, loc := range record.Locations {
			d := int(loc.EndDate.Sub(loc.StartDate).Hours()/24) + 1
			totalDays += d
		}
		if aggTicket == 0 { aggTicket = record.Cost.TicketGo + record.Cost.TicketBack }
		if aggLokal == 0 { aggLokal = record.Cost.LocalTransport }
		if aggHotel == 0 { aggHotel = record.Cost.HotelRate * float64(record.Cost.HotelDays) }
		aggSbm := record.Cost.DailyAllowanceRate * float64(totalDays)
		totalAgg := aggTicket + aggLokal + aggSbm + aggHotel + aggTambahan
		vars["total_biaya"] = fmt.Sprintf("Rp %s", utils.FormatRupiah(totalAgg))
		vars["terbilang_biaya"] = fmt.Sprintf("%s RUPIAH", strings.ToUpper(terbilang.FormatTerbilang(int(totalAgg))))
		vars["biaya_hotel"] = fmt.Sprintf("Rp %s", utils.FormatRupiah(aggHotel))
		vars["biaya_pesawat"] = fmt.Sprintf("Rp %s", utils.FormatRupiah(aggTicket))
	}
	return vars
}

func (h *Handler) ExportSpdPDF(c echo.Context) error {
	return h.exportDocument(c, "Berkas Luar Kota - SPD.docx", "SPD")
}

func (h *Handler) ExportSpdDocx(c echo.Context) error {
	return h.exportDocumentDocx(c, "Berkas Luar Kota - SPD.docx", "SPD")
}

func (h *Handler) ExportLaporanPDF(c echo.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("panic: %v", r))
		}
	}()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid record ID")
	}
	record, err := h.svc.GetRecordByID(id)
	if err != nil || record == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Record not found")
	}

	allRecords, _ := h.svc.GetRecords(map[string]interface{}{})
	var spdGroupRecords []models.TravelRecord
	for _, r := range allRecords {
		if r.SPDNumber == record.SPDNumber { spdGroupRecords = append(spdGroupRecords, r) }
	}
	sort.Slice(spdGroupRecords, func(i, j int) bool {
		nipI := strings.TrimSpace(spdGroupRecords[i].Employee.NIP)
		nipJ := strings.TrimSpace(spdGroupRecords[j].Employee.NIP)
		
		hasNIPI := nipI != "" && nipI != "-"
		hasNIPJ := nipJ != "" && nipJ != "-"

		if hasNIPI && !hasNIPJ {
			return true
		} else if !hasNIPI && hasNIPJ {
			return false
		}
		return spdGroupRecords[i].CreatedAt.Unix() < spdGroupRecords[j].CreatedAt.Unix()
	})

	vars := h.mapTravelToDocument(record, 1)
	ordinals := []string{"satu", "dua", "tiga", "empat", "lima", "enam", "tujuh", "delapan", "sembilan", "sepuluh"}
	for i, ordinal := range ordinals {
		if i < len(spdGroupRecords) {
			vars[fmt.Sprintf("nama_petugas_%s", ordinal)] = spdGroupRecords[i].Employee.Name
			vars[fmt.Sprintf("nip_petugas_%s", ordinal)] = spdGroupRecords[i].Employee.NIP
			vars[fmt.Sprintf("no_urut_%s", ordinal)] = fmt.Sprintf("%d.", i+1)
		} else { 
			vars[fmt.Sprintf("nama_petugas_%s", ordinal)] = "__REMOVE_ROW__" 
			vars[fmt.Sprintf("nip_petugas_%s", ordinal)] = ""
			vars[fmt.Sprintf("no_urut_%s", ordinal)] = ""
		}
	}

	fotoOrdinals := []string{"satu", "dua", "tiga", "empat"}

	images := make(map[string]document.ImageData)
	var pdfAttachments [][]byte
	lampiranIndex := 1
	fotoIndex := 0

	addLamp := func(data, mime string, isFoto bool) {
		if data == "" { return }
		if strings.HasPrefix(mime, "image/") {
			if isFoto {
				if fotoIndex < len(fotoOrdinals) {
					images[fmt.Sprintf("foto_dokumentasi_%s", fotoOrdinals[fotoIndex])] = document.ImageData{Data: data, MimeType: mime}
					fotoIndex++
				}
			} else {
				if lampiranIndex <= 100 {
					images[fmt.Sprintf("lampiran_%d", lampiranIndex)] = document.ImageData{Data: data, MimeType: mime}
					lampiranIndex++
				}
			}
		} else if mime == "application/pdf" {
			raw := data
			if i := strings.Index(raw, ","); i != -1 { raw = raw[i+1:] }
			b, err := base64.StdEncoding.DecodeString(raw)
			if err == nil { pdfAttachments = append(pdfAttachments, b) }
		}
	}

	type FileObj struct { Data string `json:"data"`; Type string `json:"type"` }
	processedFoto := false
	for _, r := range spdGroupRecords {
		if !processedFoto && r.Report != nil && len(r.Report.Files) > 0 {
			var files []FileObj
			if err := json.Unmarshal(r.Report.Files, &files); err == nil {
				for _, f := range files { addLamp(f.Data, f.Type, true) }
				processedFoto = true
			}
		}
		if r.Cost != nil && len(r.Cost.Details) > 0 {
			var details []struct {
				TicketGoFile *FileObj `json:"ticketGoFile"`; TicketBackFile *FileObj `json:"ticketBackFile"`
				HotelFile *FileObj `json:"hotelFile"`; TransportFile *FileObj `json:"transportFile"`
				BoardingPassFiles []FileObj `json:"boardingPassFiles"`
				AdditionalCosts []struct {
					File *FileObj `json:"file"`; TicketGoFile *FileObj `json:"ticketGoFile"`
					TicketBackFile *FileObj `json:"ticketBackFile"`; BoardingPassFiles []FileObj `json:"boardingPassFiles"`
				} `json:"additionalCosts"`
			}
			if err := json.Unmarshal(r.Cost.Details, &details); err == nil {
				for _, d := range details {
					if d.TicketGoFile != nil { addLamp(d.TicketGoFile.Data, d.TicketGoFile.Type, false) }
					if d.TicketBackFile != nil { addLamp(d.TicketBackFile.Data, d.TicketBackFile.Type, false) }
					for _, bp := range d.BoardingPassFiles { addLamp(bp.Data, bp.Type, false) }
					if d.HotelFile != nil { addLamp(d.HotelFile.Data, d.HotelFile.Type, false) }
					if d.TransportFile != nil { addLamp(d.TransportFile.Data, d.TransportFile.Type, false) }
					for _, ac := range d.AdditionalCosts {
						if ac.File != nil { addLamp(ac.File.Data, ac.File.Type, false) }
						if ac.TicketGoFile != nil { addLamp(ac.TicketGoFile.Data, ac.TicketGoFile.Type, false) }
						if ac.TicketBackFile != nil { addLamp(ac.TicketBackFile.Data, ac.TicketBackFile.Type, false) }
						for _, abp := range ac.BoardingPassFiles { addLamp(abp.Data, abp.Type, false) }
					}
				}
			}
		}
	}

	for _, ordinal := range fotoOrdinals {
		key := fmt.Sprintf("foto_dokumentasi_%s", ordinal)
		if _, ok := images[key]; !ok {
			vars[key] = "__REMOVE_ROW__"
		}
	}

	for i := 1; i <= 100; i++ {
		key := fmt.Sprintf("lampiran_%d", i)
		if _, ok := images[key]; !ok { vars[key] = "__REMOVE_P__" }
	}

	payload := document.DocumentRequest{TemplateName: "Berkas Luar Kota - Laporan.docx", Variables: vars}
	pdfBytes, err := h.docGen.GenerateWithImages(c.Request().Context(), payload, images)
	if err != nil { return echo.NewHTTPError(http.StatusInternalServerError, err.Error()) }

	if len(pdfAttachments) > 0 {
		merged, err := h.docGen.MergePDFs(c.Request().Context(), append([][]byte{pdfBytes}, pdfAttachments...))
		if err == nil { pdfBytes = merged }
	}

	filename := fmt.Sprintf("Laporan_%s_%s.pdf", record.Employee.Name, record.SPDNumber)
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=\"%s\"", filename))
	return c.Blob(http.StatusOK, "application/pdf", pdfBytes)
}

func (h *Handler) ExportLaporanDocx(c echo.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("panic: %v", r))
		}
	}()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid record ID")
	}
	record, err := h.svc.GetRecordByID(id)
	if err != nil || record == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Record not found")
	}

	vars := h.mapTravelToDocument(record, 1)
	payload := document.DocumentRequest{TemplateName: "Berkas Luar Kota - Laporan.docx", Variables: vars}
	docxBytes, _ := h.docGen.GenerateDocx(c.Request().Context(), payload)
	filename := fmt.Sprintf("Laporan_%s_%s.docx", record.Employee.Name, record.SPDNumber)
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=\"%s\"", filename))
	return c.Blob(http.StatusOK, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", docxBytes)
}

func (h *Handler) ExportRincianPDF(c echo.Context) error {
	return h.exportDocument(c, "Berkas Luar Kota - rincian pembayaran.docx", "Rincian")
}

func (h *Handler) ExportRincianDocx(c echo.Context) error {
	return h.exportDocumentDocx(c, "Berkas Luar Kota - rincian pembayaran.docx", "Rincian")
}

func (h *Handler) exportDocument(c echo.Context, templateName, prefix string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("panic: %v", r))
		}
	}()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid record ID")
	}
	record, err := h.svc.GetRecordByID(id)
	if err != nil || record == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Record not found")
	}

	payload := document.DocumentRequest{TemplateName: templateName, Variables: h.mapTravelToDocument(record, 1)}
	pdfBytes, err := h.docGen.Generate(c.Request().Context(), payload)
	if err != nil { return echo.NewHTTPError(http.StatusInternalServerError, err.Error()) }
	filename := fmt.Sprintf("%s_%s_%s.pdf", prefix, record.Employee.Name, record.SPDNumber)
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=\"%s\"", filename))
	return c.Blob(http.StatusOK, "application/pdf", pdfBytes)
}

func (h *Handler) exportDocumentDocx(c echo.Context, templateName string, prefix string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("panic: %v", r))
		}
	}()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid record ID")
	}
	record, err := h.svc.GetRecordByID(id)
	if err != nil || record == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Record not found")
	}

	payload := document.DocumentRequest{TemplateName: templateName, Variables: h.mapTravelToDocument(record, 1)}
	docxBytes, _ := h.docGen.GenerateDocx(c.Request().Context(), payload)
	filename := fmt.Sprintf("%s_%s_%s.docx", prefix, record.Employee.Name, record.SPDNumber)
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=\"%s\"", filename))
	return c.Blob(http.StatusOK, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", docxBytes)
}

func (h *Handler) UploadFile(c echo.Context) error {
	file, _ := c.FormFile("file")
	src, _ := file.Open()
	defer src.Close()
	filename := uuid.New().String() + "_" + filepath.Base(file.Filename)
	dst, _ := os.Create(filepath.Join("uploads", filename))
	defer dst.Close()
	io.Copy(dst, src)
	return c.JSON(http.StatusOK, map[string]string{"path": filename})
}

func (h *Handler) CreateRecord(c echo.Context) error {
	var r models.TravelRecord
	if err := c.Bind(&r); err != nil {
		fmt.Printf("Error binding record: %v\n", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	
	if creatorIDStr, ok := c.Get("user_id").(string); ok && creatorIDStr != "" {
		r.CreatorID, _ = uuid.Parse(creatorIDStr)
	}

	if r.SPDNumber == "" { r.SPDNumber, _ = h.svc.GenerateSpdNumber() }
	if len(r.EmployeeIDs) > 0 {
		for _, eid := range r.EmployeeIDs {
			nr := r; nr.ID = uuid.Nil; nr.EmployeeID = eid
			if err := h.svc.CreateRecord(&nr); err != nil {
				fmt.Printf("Error creating record for employee %s: %v\n", eid, err)
				return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			}
			go h.notifyEmployee(&nr)
		}
		return c.NoContent(http.StatusCreated)
	}
	
	if r.EmployeeID == uuid.Nil {
		r.EmployeeID = r.CreatorID
	}
	
	if err := h.svc.CreateRecord(&r); err != nil {
		fmt.Printf("Error creating record: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	go h.notifyEmployee(&r)
	return c.JSON(http.StatusCreated, r)
}

func (h *Handler) GetRecords(c echo.Context) error {
	role := c.Get("role").(string)
	recs, _ := h.svc.GetRecords(nil)
	if role == "protokol" {
		uid, _ := uuid.Parse(c.Get("user_id").(string))
		spds := make(map[string]bool)
		for _, r := range recs { if r.EmployeeID == uid { spds[r.SPDNumber] = true } }
		var res []models.TravelRecord
		for _, r := range recs { if spds[r.SPDNumber] { res = append(res, r) } }
		return c.JSON(http.StatusOK, res)
	}
	return c.JSON(http.StatusOK, recs)
}

func (h *Handler) GetRecordByID(c echo.Context) error {
	id, _ := uuid.Parse(c.Param("id"))
	r, _ := h.svc.GetRecordByID(id)
	return c.JSON(http.StatusOK, r)
}

func (h *Handler) UpdateRecord(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "ID tidak valid")
	}

	// Read body once for debugging if needed, but Bind is preferred
	r, err := h.svc.GetRecordByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Data tidak ditemukan")
	}

	if err := c.Bind(r); err != nil {
		fmt.Printf("UpdateRecord Bind Error: %v\n", err)
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Gagal memproses data: %v", err))
	}

	r.ID = id // Ensure ID stays correct
	
	fmt.Printf("Updating record %s, payload size approx: %d bytes\n", id, c.Request().ContentLength)

	if err := h.svc.UpdateRecord(r); err != nil {
		fmt.Printf("UpdateRecord Service Error: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Gagal menyimpan ke database: %v", err))
	}

	return c.JSON(http.StatusOK, r)
}

func (h *Handler) DeleteRecord(c echo.Context) error {
	id, _ := uuid.Parse(c.Param("id"))
	h.svc.DeleteRecord(id)
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) DeleteRecordsBySpd(c echo.Context) error {
	h.svc.DeleteRecordsBySpd(c.Request().Context(), c.Param("spd"))
	return c.JSON(http.StatusOK, map[string]string{"message": "deleted"})
}
