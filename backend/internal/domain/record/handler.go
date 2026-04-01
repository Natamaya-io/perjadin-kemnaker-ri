package record

import (
	"context"
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
	// mapTravelToDocument maps TravelRecord data to placeholders used in DOCX templates.
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

	// SPD sequence from ID-SPJ-001 -> 001
	spdSubNumber := fmt.Sprintf("%03d", globalIndex)
	
	extractNumericID := func(id string) string {
		if id == "" {
			return spdSubNumber
		}
		parts := strings.Split(id, "-")
		if len(parts) > 0 {
			suffix := parts[len(parts)-1]
			if len(suffix) < 3 {
				return fmt.Sprintf("%03s", suffix)
			}
			return suffix
		}
		return id
	}

	noSurat := record.SuratTugasNumber
	tglSurat := record.SuratTugasDate
	
	if noSurat == "" {
		noSurat = extractNumericID(record.SPDNumber)
	}

	noSpd := record.SPDNumber
	if noSpd == "" {
		noSpd = spdSubNumber
	}

	maksud := strings.TrimSpace(record.Purpose)
	stakeholder := strings.TrimSpace(record.Stakeholder)
	
	redundantPhrases := []string{"Kunjungan Kerja", "kunjungan kerja", "KUNJUNGAN KERJA"}
	
	cleanMaksud := maksud
	cleanStakeholder := stakeholder

	// Remove "kunjungan kerja" from both variables because the template 
	// provides it as static text between {{tujuan_perjalanan}} and {{stakeholder}}.
	for _, phrase := range redundantPhrases {
		re := regexp.MustCompile("(?i)\\s*" + regexp.QuoteMeta(phrase) + "\\s*")
		cleanMaksud = re.ReplaceAllString(cleanMaksud, " ")
		cleanStakeholder = re.ReplaceAllString(cleanStakeholder, " ")
	}

	cleanMaksud = regexp.MustCompile(`\s+`).ReplaceAllString(cleanMaksud, " ")
	cleanMaksud = strings.TrimSpace(cleanMaksud)
	
	cleanStakeholder = regexp.MustCompile(`\s+`).ReplaceAllString(cleanStakeholder, " ")
	cleanStakeholder = strings.TrimSpace(cleanStakeholder)

	// Signatory Logic
	namaPpk := h.cfg.Signatory.PPKName
	nipPpk := h.cfg.Signatory.PPKNIP
	namaBendahara := h.cfg.Signatory.BendaharaName
	nipBendahara := h.cfg.Signatory.BendaharaNIP
	jabPpk := "Pejabat Pembuat Komitmen"
	jabBendahara := "Bendahara Pengeluaran Pembantu"

	if h.masterSvc != nil {
		globalSettings, _ := h.masterSvc.GetSettings(context.Background())
		if v, ok := globalSettings["ppk_name"]; ok && v != "" {
			namaPpk = v
		}
		if v, ok := globalSettings["ppk_nip"]; ok && v != "" {
			nipPpk = v
		}
		if v, ok := globalSettings["bendahara_name"]; ok && v != "" {
			namaBendahara = v
		}
		if v, ok := globalSettings["bendahara_nip"]; ok && v != "" {
			nipBendahara = v
		}
	}

	if record.Report != nil {
		if record.Report.PPKName != "" {
			namaPpk = record.Report.PPKName
		}
		if record.Report.PPKNIP != "" {
			nipPpk = record.Report.PPKNIP
		}
		if record.Report.BendaharaName != "" {
			namaBendahara = record.Report.BendaharaName
		}
		if record.Report.BendaharaNIP != "" {
			nipBendahara = record.Report.BendaharaNIP
		}
		// Future proof: User might input titles in report too
		// if record.Report.PPKTitle != "" { jabPpk = record.Report.PPKTitle }
	}

	vars := map[string]interface{}{
		"no_spd":               noSpd,
		"no_surat":            noSurat,
		"bulan_no_surat":      utils.GetRomanMonths()[int(tglSurat.Month())],
		"tahun_no_surat":      tglSurat.Year(),
		"tanggal_no_surat":    utils.FormatIndonesianDate(tglSurat),
		"bulan_pembayaran":    utils.GetIndonesianMonths()[int(time.Now().Month())],
		"tahun_pembayaran":    time.Now().Year(),
		"bulan":               "",
		"bulan_romawi":        utils.GetRomanMonths()[int(record.StartDate.Month())],
		"tahun":               "",
		"tahun_saat_ini":      fmt.Sprintf("%d", time.Now().Year()),
		"nama":                record.Employee.Name,
		"nama_petugas":        record.Employee.Name,
		"nip":                 record.Employee.NIP,
		"nip_petugas":         record.Employee.NIP,
		"pangkat_gol":         fmt.Sprintf("%s (%s)", record.Employee.Pangkat, record.Employee.Golongan),
		"pangkat":             record.Employee.Pangkat,
		"golongan":            record.Employee.Golongan,
		"jabatan":             record.Employee.Jabatan,
		"tingkat_biaya":       record.Employee.TingkatBiaya,
		"maksud_perjalanan":   cleanMaksud,
		"tujuan_perjalanan":  cleanMaksud,
		"nama_stakeholder":    cleanStakeholder,
		"stakeholder":        cleanStakeholder,
		"tujuan":              dest,
		"kota_atau_kabupaten": dest,
		"kota":                dest,
		"provinsi":            prov,
		"transportasi":        transportMode,
		"tgl_berangkat":       utils.FormatIndonesianDate(record.StartDate),
		"tanggal_berangkat":   utils.FormatIndonesianDate(record.StartDate),
		"tanggal_mulai":      utils.FormatIndonesianDate(record.StartDate),
		"tgl_kembali":         utils.FormatIndonesianDate(record.EndDate),
		"tanggal_selesai":     utils.FormatIndonesianDate(record.EndDate),
		"lama_hari":           days,
		"lama_perjalanan":     days,
		"terbilang":           titleCaser.String(terbilang.FormatTerbilang(days)),
		"tgl_surat_tugas":     utils.FormatIndonesianDate(tglSurat),
		"tanggal_dikeluarkan": utils.FormatIndonesianDate(time.Now()),
		"tgl_cetak":           utils.FormatIndonesianDate(time.Now()),
		"nama_ppk":            namaPpk,
		"nip_ppk":             nipPpk,
		"nama ppk":            namaPpk,
		"nip ppk":             nipPpk,
		"nama_bendahara":      namaBendahara,
		"nip_bendahara":       nipBendahara,
		"jabatan_ppk":         jabPpk,
		"jabatan_bendahara":   jabBendahara,
		"keterangan":          "-",
		"tiket_pesawat":     "0",
		"transport_lokal":   "0",
		"transport_daerah":  "0",
		"sbm":               "0",
		"total_sbm":         "0",
		"penginapan":        "0",
		"total_penginapan":  "0",
		"total_biaya":       "0",
		"total_keseluruhan": "0",
		"total_kesuluruhan": "0",
		"total_akhir":       "0",
		"jumlah_total":      "0",
		"total_rupiah":      "0",
		"total_terbilang":   "Nol Rupiah",
		"durasi":            fmt.Sprintf("%d", days),
		"jml_sbm":           "0",
		"jml_penginapan":    "0",
		"ditetapkan_sejumlah": "0",
		"dibayar_semula":      "0",
		"sisa_kurang_lebih":   "0",
	}

	ordinals := []string{"satu", "dua", "tiga", "empat", "lima", "enam", "tujuh", "delapan", "sembilan", "sepuluh"}

	type Detail struct {
		TransportMode   string  `json:"transportMode"`
		TicketGo        float64 `json:"ticketGo"`
		TicketBack      float64 `json:"ticketBack"`
		HotelDays       int     `json:"hotelDays"`
		HotelRate       float64 `json:"hotelRate"`
		TransportAmount float64 `json:"transportAmount"`
		AdditionalCosts []struct {
			Name   string  `json:"name"`
			Amount float64 `json:"amount"`
		} `json:"additionalCosts"`
	}

	var costDetails []Detail
	if record.Cost != nil && len(record.Cost.Details) > 0 {
		json.Unmarshal(record.Cost.Details, &costDetails)
	}

	for i, ordinal := range ordinals {
		if i < len(record.Locations) {
			loc := record.Locations[i]
			locDays := int(loc.EndDate.Sub(loc.StartDate).Hours()/24) + 1

			vars[fmt.Sprintf("no_%s", ordinal)] = i + 1
			vars[fmt.Sprintf("tujuan_%s", ordinal)] = titleCaser.String(loc.Location)
			vars[fmt.Sprintf("provinsi_%s", ordinal)] = titleCaser.String(loc.Province)
			vars[fmt.Sprintf("tgl_pergi_%s", ordinal)] = utils.FormatIndonesianDate(loc.StartDate)
			vars[fmt.Sprintf("tgl_pulang_%s", ordinal)] = utils.FormatIndonesianDate(loc.EndDate)
			vars[fmt.Sprintf("hari_%s", ordinal)] = locDays
			vars[fmt.Sprintf("lama_hari_%s", ordinal)] = locDays

			var sbmVal float64
			var sbmRate float64
			if record.Cost != nil {
				sbmRate = record.Cost.DailyAllowanceRate
				sbmVal = sbmRate * float64(locDays)
			}
			vars[fmt.Sprintf("sbm_rate_%s", ordinal)] = utils.FormatRupiah(sbmRate)
			vars[fmt.Sprintf("sbm_%s", ordinal)] = utils.FormatRupiah(sbmVal)
			vars[fmt.Sprintf("tarif_harian_%s", ordinal)] = utils.FormatRupiah(sbmRate)
			vars[fmt.Sprintf("total_harian_%s", ordinal)] = utils.FormatRupiah(sbmVal)
			vars[fmt.Sprintf("uang_harian_%s", ordinal)] = utils.FormatRupiah(sbmVal)

			if i < len(costDetails) {
				cd := costDetails[i]
				totalHotel := float64(cd.HotelDays) * cd.HotelRate
				totalTransport := cd.TransportAmount + cd.TicketGo + cd.TicketBack
				var totalAdd float64
				for _, ac := range cd.AdditionalCosts {
					totalAdd += ac.Amount
				}

				vars[fmt.Sprintf("hotel_days_%s", ordinal)] = cd.HotelDays
				vars[fmt.Sprintf("hotel_hari_%s", ordinal)] = cd.HotelDays
				vars[fmt.Sprintf("hotel_rate_%s", ordinal)] = utils.FormatRupiah(cd.HotelRate)
				vars[fmt.Sprintf("tarif_penginapan_%s", ordinal)] = utils.FormatRupiah(cd.HotelRate)
				vars[fmt.Sprintf("hotel_%s", ordinal)] = utils.FormatRupiah(totalHotel)
				vars[fmt.Sprintf("total_penginapan_%s", ordinal)] = utils.FormatRupiah(totalHotel)
				vars[fmt.Sprintf("penginapan_%s", ordinal)] = utils.FormatRupiah(totalHotel)
				
				vars[fmt.Sprintf("transport_%s", ordinal)] = utils.FormatRupiah(totalTransport)
				vars[fmt.Sprintf("total_transport_%s", ordinal)] = utils.FormatRupiah(totalTransport)
				vars[fmt.Sprintf("tiket_%s", ordinal)] = utils.FormatRupiah(cd.TicketGo + cd.TicketBack)
				vars[fmt.Sprintf("tambahan_%s", ordinal)] = utils.FormatRupiah(totalAdd)
				vars[fmt.Sprintf("total_tambahan_%s", ordinal)] = utils.FormatRupiah(totalAdd)
			} else {
				vars[fmt.Sprintf("hotel_days_%s", ordinal)] = "0"
				vars[fmt.Sprintf("hotel_hari_%s", ordinal)] = "0"
				vars[fmt.Sprintf("hotel_rate_%s", ordinal)] = "0"
				vars[fmt.Sprintf("tarif_penginapan_%s", ordinal)] = "0"
				vars[fmt.Sprintf("hotel_%s", ordinal)] = "0"
				vars[fmt.Sprintf("total_penginapan_%s", ordinal)] = "0"
				vars[fmt.Sprintf("penginapan_%s", ordinal)] = "0"
				vars[fmt.Sprintf("transport_%s", ordinal)] = "0"
				vars[fmt.Sprintf("total_transport_%s", ordinal)] = "0"
				vars[fmt.Sprintf("tiket_%s", ordinal)] = "0"
				vars[fmt.Sprintf("tambahan_%s", ordinal)] = "0"
				vars[fmt.Sprintf("total_tambahan_%s", ordinal)] = "0"
			}
		} else {
			vars[fmt.Sprintf("no_%s", ordinal)] = ""
			vars[fmt.Sprintf("tujuan_%s", ordinal)] = "__REMOVE_ROW__"
			vars[fmt.Sprintf("provinsi_%s", ordinal)] = ""
			vars[fmt.Sprintf("tgl_pergi_%s", ordinal)] = ""
			vars[fmt.Sprintf("tgl_pulang_%s", ordinal)] = ""
			vars[fmt.Sprintf("hari_%s", ordinal)] = ""
			vars[fmt.Sprintf("lama_hari_%s", ordinal)] = ""
			vars[fmt.Sprintf("sbm_rate_%s", ordinal)] = ""
			vars[fmt.Sprintf("sbm_%s", ordinal)] = ""
			vars[fmt.Sprintf("tarif_harian_%s", ordinal)] = ""
			vars[fmt.Sprintf("total_harian_%s", ordinal)] = ""
			vars[fmt.Sprintf("uang_harian_%s", ordinal)] = ""
			vars[fmt.Sprintf("hotel_days_%s", ordinal)] = ""
			vars[fmt.Sprintf("hotel_hari_%s", ordinal)] = ""
			vars[fmt.Sprintf("hotel_rate_%s", ordinal)] = ""
			vars[fmt.Sprintf("tarif_penginapan_%s", ordinal)] = ""
			vars[fmt.Sprintf("hotel_%s", ordinal)] = ""
			vars[fmt.Sprintf("total_penginapan_%s", ordinal)] = ""
			vars[fmt.Sprintf("penginapan_%s", ordinal)] = ""
			vars[fmt.Sprintf("transport_%s", ordinal)] = ""
			vars[fmt.Sprintf("total_transport_%s", ordinal)] = ""
			vars[fmt.Sprintf("tiket_%s", ordinal)] = ""
			vars[fmt.Sprintf("tambahan_%s", ordinal)] = ""
			vars[fmt.Sprintf("total_tambahan_%s", ordinal)] = ""
		}
	}

	if record.Report != nil {
		vars["isi_laporan"] = record.Report.Text
		vars["tgl_laporan"] = utils.FormatIndonesianDate(record.Report.SubmittedAt)
	}

	if record.Cost != nil {
		total := record.TotalCost
		vars["total_biaya"] = fmt.Sprintf("Rp %s", utils.FormatRupiah(total))
		vars["total_keseluruhan"] = utils.FormatRupiah(total)
		vars["total_kesuluruhan"] = utils.FormatRupiah(total)
		vars["total_akhir"] = utils.FormatRupiah(total)
		vars["jumlah_total"] = utils.FormatRupiah(total)
		vars["total_rupiah"] = utils.FormatRupiah(total)
		vars["terbilang_biaya"] = fmt.Sprintf("%s Rupiah", titleCaser.String(terbilang.FormatTerbilang(int(total))))
		vars["total_terbilang"] = fmt.Sprintf("%s Rupiah", titleCaser.String(terbilang.FormatTerbilang(int(total))))

		// AGGREGATE COSTS FROM DETAILS (for multi-location support)
		var aggTicket, aggLokal, aggDaerah, aggSbm, aggHotel float64
		var totalDays, hotelDays int
		
		for _, cd := range costDetails {
			aggTicket += cd.TicketGo + cd.TicketBack
			aggLokal += cd.TransportAmount
			aggHotel += float64(cd.HotelDays) * cd.HotelRate
			hotelDays += cd.HotelDays
		}
		
		for _, loc := range record.Locations {
			d := int(loc.EndDate.Sub(loc.StartDate).Hours()/24) + 1
			totalDays += d
			aggSbm += record.Cost.DailyAllowanceRate * float64(d)
		}

		// Fallback to top-level if aggregation is 0
		if aggTicket == 0 { aggTicket = record.Cost.TicketGo + record.Cost.TicketBack }
		if aggLokal == 0 { aggLokal = record.Cost.LocalTransport }
		if aggDaerah == 0 { aggDaerah = record.Cost.RegionalTransport }
		if aggSbm == 0 { aggSbm = record.Cost.DailyAllowanceRate * float64(record.Cost.DailyAllowanceDays) }
		if aggHotel == 0 { aggHotel = record.Cost.HotelRate * float64(record.Cost.HotelDays) }
		if totalDays == 0 { totalDays = record.Cost.DailyAllowanceDays }
		if hotelDays == 0 { hotelDays = record.Cost.HotelDays }

		vars["tiket_pesawat"] = utils.FormatRupiah(aggTicket)
		vars["transport_lokal"] = utils.FormatRupiah(aggLokal)
		vars["transport_daerah"] = utils.FormatRupiah(aggDaerah)
		vars["sbm"] = utils.FormatRupiah(record.Cost.DailyAllowanceRate)
		vars["total_sbm"] = utils.FormatRupiah(aggSbm)
		vars["jml_sbm"] = utils.FormatRupiah(aggSbm)
		vars["penginapan"] = utils.FormatRupiah(record.Cost.HotelRate)
		vars["total_penginapan"] = utils.FormatRupiah(aggHotel)
		vars["jml_penginapan"] = utils.FormatRupiah(aggHotel)
		
		vars["durasi_sbm"] = fmt.Sprintf("%d", totalDays)
		vars["tarif_sbm"] = utils.FormatRupiah(record.Cost.DailyAllowanceRate)
		vars["durasi_hotel"] = fmt.Sprintf("%d", hotelDays)
		vars["tarif_hotel"] = utils.FormatRupiah(record.Cost.HotelRate)

		vars["biaya_harian"] = fmt.Sprintf("Rp %s", utils.FormatRupiah(aggSbm))
		vars["biaya_hotel"] = fmt.Sprintf("Rp %s", utils.FormatRupiah(aggHotel))
		vars["biaya_pesawat"] = fmt.Sprintf("Rp %s", utils.FormatRupiah(aggTicket))

		vars["ditetapkan_sejumlah"] = utils.FormatRupiah(total)
		vars["sisa_kurang_lebih"] = utils.FormatRupiah(total)
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
		globalIndex = 1
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

func (h *Handler) ExportSpdPDF(c echo.Context) error {
	return h.exportDocument(c, "Berkas Luar Kota - SPD.docx", "SPD")
}

func (h *Handler) ExportSpdDocx(c echo.Context) error {
	return h.exportDocumentDocx(c, "Berkas Luar Kota - SPD.docx", "SPD")
}

func (h *Handler) ExportLaporanPDF(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Format UUID tidak valid")
	}

	record, err := h.svc.GetRecordByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Data tidak ditemukan")
	}

	allRecords, err := h.svc.GetRecords(map[string]interface{}{})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Gagal mengambil data records")
	}

	var spdGroupRecords []models.TravelRecord
	for _, r := range allRecords {
		if r.SPDNumber == record.SPDNumber {
			spdGroupRecords = append(spdGroupRecords, r)
		}
	}

	sort.Slice(spdGroupRecords, func(i, j int) bool {
		ti := spdGroupRecords[i].CreatedAt.Unix()
		tj := spdGroupRecords[j].CreatedAt.Unix()
		if ti != tj {
			return ti < tj
		}
		return spdGroupRecords[i].ID.String() < spdGroupRecords[j].ID.String()
	})

	namaPpk := h.cfg.Signatory.PPKName
	nipPpk := h.cfg.Signatory.PPKNIP

	if h.masterSvc != nil {
		globalSettings, _ := h.masterSvc.GetSettings(context.Background())
		if v, ok := globalSettings["ppk_name"]; ok && v != "" {
			namaPpk = v
		}
		if v, ok := globalSettings["ppk_nip"]; ok && v != "" {
			nipPpk = v
		}
	}

	if record.Report != nil {
		if record.Report.PPKName != "" {
			namaPpk = record.Report.PPKName
		}
		if record.Report.PPKNIP != "" {
			nipPpk = record.Report.PPKNIP
		}
	}

	noSuratTugas := record.SuratTugasNumber
	if noSuratTugas == "" {
		noSuratTugas = "-"
	}

	bulanNoSurat := ""
	tanggalNoSurat := ""
	if !record.SuratTugasDate.IsZero() {
		bulanNoSurat = utils.GetRomanMonths()[int(record.SuratTugasDate.Month())]
		tanggalNoSurat = fmt.Sprintf("%d", record.SuratTugasDate.Day())
	}

	isiLaporan := "-"
	tanggalLaporan := utils.FormatIndonesianDate(time.Now())
	if record.Report != nil {
		if record.Report.Text != "" {
			isiLaporan = record.Report.Text
		}
		if !record.Report.SubmittedAt.IsZero() {
			tanggalLaporan = utils.FormatIndonesianDate(record.Report.SubmittedAt)
		}
	}

	vars := map[string]interface{}{
		"kota":               record.Location,
		"provinsi":           record.Province,
		"tanggal_mulai":      utils.FormatIndonesianDate(record.StartDate),
		"tanggal_selesai":    utils.FormatIndonesianDate(record.EndDate),
		"bulan":              utils.GetRomanMonths()[int(record.StartDate.Month())],
		"tahun":              record.StartDate.Year(),
		"no_surat":           noSuratTugas,
		"bulan_no_surat":     bulanNoSurat,
		"tanggal_no_surat":   tanggalNoSurat,
		"tujuan_perjalanan":  record.Purpose,
		"stakeholder":        record.Stakeholder,
		"isi_laporan":        isiLaporan,
		"tanggal_dikeluarkan": tanggalLaporan,
		"nama_ppk":           namaPpk,
		"nip_ppk":            nipPpk,
	}

	ordinals := []string{"satu", "dua", "tiga", "empat", "lima", "enam", "tujuh", "delapan", "sembilan", "sepuluh"}
	for i, ordinal := range ordinals {
		if i < len(spdGroupRecords) {
			emp := spdGroupRecords[i].Employee
			vars[fmt.Sprintf("no_urut_%s", ordinal)] = fmt.Sprintf("%d.", i+1)
			vars[fmt.Sprintf("nama_petugas_%s", ordinal)] = emp.Name
			vars[fmt.Sprintf("nip_petugas_%s", ordinal)] = emp.NIP
			vars[fmt.Sprintf("no_urut_ttd_%s", ordinal)] = fmt.Sprintf("%d.", i+1)
		} else {
			vars[fmt.Sprintf("no_urut_%s", ordinal)] = ""
			vars[fmt.Sprintf("nama_petugas_%s", ordinal)] = ""
			vars[fmt.Sprintf("nip_petugas_%s", ordinal)] = ""
			vars[fmt.Sprintf("no_urut_ttd_%s", ordinal)] = ""
		}
	}

	imageOrdinals := []string{"satu", "dua", "tiga", "empat", "lima", "enam"}
	images := make(map[string]document.ImageData)
	if record.Report != nil && len(record.Report.Files) > 0 {
		var reportFiles []struct {
			Name      string `json:"name"`
			Type      string `json:"type"`
			Data      string `json:"data"`
			Timestamp string `json:"timestamp"`
		}
		if err := json.Unmarshal(record.Report.Files, &reportFiles); err == nil {
			for i, f := range reportFiles {
				if i >= len(imageOrdinals) {
					break
				}
				docKey := fmt.Sprintf("foto_dokumentasi_%s", imageOrdinals[i])
				images[docKey] = document.ImageData{
					Data:     f.Data,
					MimeType: f.Type,
				}
			}
		}
	}

	for _, ordinal := range imageOrdinals {
		key := fmt.Sprintf("foto_dokumentasi_%s", ordinal)
		if _, ok := images[key]; !ok {
			vars[key] = ""
		}
	}

	payload := document.DocumentRequest{
		TemplateName: "Berkas Luar Kota - Laporan.docx",
		Variables:    vars,
	}

	var pdfBytes []byte
	var errPDF error
	if len(images) > 0 {
		pdfBytes, errPDF = h.docGen.GenerateWithImages(c.Request().Context(), payload, images)
	} else {
		pdfBytes, errPDF = h.docGen.Generate(c.Request().Context(), payload)
	}
	if errPDF != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Gagal membuat dokumen laporan: %v", errPDF))
	}

	filename := fmt.Sprintf("Laporan_%s_%s.pdf", record.Employee.Name, record.SPDNumber)
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Response().Header().Set(echo.HeaderContentType, "application/pdf")

	return c.Blob(http.StatusOK, "application/pdf", pdfBytes)
}

func (h *Handler) ExportLaporanDocx(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Format UUID tidak valid")
	}

	record, err := h.svc.GetRecordByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Data tidak ditemukan")
	}

	allRecords, err := h.svc.GetRecords(map[string]interface{}{})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Gagal mengambil data records")
	}

	var spdGroupRecords []models.TravelRecord
	for _, r := range allRecords {
		if r.SPDNumber == record.SPDNumber {
			spdGroupRecords = append(spdGroupRecords, r)
		}
	}

	sort.Slice(spdGroupRecords, func(i, j int) bool {
		ti := spdGroupRecords[i].CreatedAt.Unix()
		tj := spdGroupRecords[j].CreatedAt.Unix()
		if ti != tj {
			return ti < tj
		}
		return spdGroupRecords[i].ID.String() < spdGroupRecords[j].ID.String()
	})

	namaPpk := h.cfg.Signatory.PPKName
	nipPpk := h.cfg.Signatory.PPKNIP

	if h.masterSvc != nil {
		globalSettings, _ := h.masterSvc.GetSettings(context.Background())
		if v, ok := globalSettings["ppk_name"]; ok && v != "" {
			namaPpk = v
		}
		if v, ok := globalSettings["ppk_nip"]; ok && v != "" {
			nipPpk = v
		}
	}

	if record.Report != nil {
		if record.Report.PPKName != "" {
			namaPpk = record.Report.PPKName
		}
		if record.Report.PPKNIP != "" {
			nipPpk = record.Report.PPKNIP
		}
	}

	noSuratTugas := record.SuratTugasNumber
	if noSuratTugas == "" {
		noSuratTugas = "-"
	}

	bulanNoSurat := ""
	tanggalNoSurat := ""
	if !record.SuratTugasDate.IsZero() {
		bulanNoSurat = utils.GetRomanMonths()[int(record.SuratTugasDate.Month())]
		tanggalNoSurat = fmt.Sprintf("%d", record.SuratTugasDate.Day())
	}

	isiLaporan := "-"
	tanggalLaporan := utils.FormatIndonesianDate(time.Now())
	if record.Report != nil {
		if record.Report.Text != "" {
			isiLaporan = record.Report.Text
		}
		if !record.Report.SubmittedAt.IsZero() {
			tanggalLaporan = utils.FormatIndonesianDate(record.Report.SubmittedAt)
		}
	}

	vars := map[string]interface{}{
		"kota":               record.Location,
		"provinsi":           record.Province,
		"tanggal_mulai":      utils.FormatIndonesianDate(record.StartDate),
		"tanggal_selesai":    utils.FormatIndonesianDate(record.EndDate),
		"bulan":              utils.GetRomanMonths()[int(record.StartDate.Month())],
		"tahun":              record.StartDate.Year(),
		"no_surat":           noSuratTugas,
		"bulan_no_surat":     bulanNoSurat,
		"tanggal_no_surat":   tanggalNoSurat,
		"tujuan_perjalanan":  record.Purpose,
		"stakeholder":        record.Stakeholder,
		"isi_laporan":        isiLaporan,
		"tanggal_dikeluarkan": tanggalLaporan,
		"nama_ppk":           namaPpk,
		"nip_ppk":            nipPpk,
	}

	ordinals := []string{"satu", "dua", "tiga", "empat", "lima", "enam", "tujuh", "delapan", "sembilan", "sepuluh"}
	for i, ordinal := range ordinals {
		if i < len(spdGroupRecords) {
			emp := spdGroupRecords[i].Employee
			vars[fmt.Sprintf("no_urut_%s", ordinal)] = fmt.Sprintf("%d.", i+1)
			vars[fmt.Sprintf("nama_petugas_%s", ordinal)] = emp.Name
			vars[fmt.Sprintf("nip_petugas_%s", ordinal)] = emp.NIP
			vars[fmt.Sprintf("no_urut_ttd_%s", ordinal)] = fmt.Sprintf("%d.", i+1)
		} else {
			vars[fmt.Sprintf("no_urut_%s", ordinal)] = ""
			vars[fmt.Sprintf("nama_petugas_%s", ordinal)] = ""
			vars[fmt.Sprintf("nip_petugas_%s", ordinal)] = ""
			vars[fmt.Sprintf("no_urut_ttd_%s", ordinal)] = ""
		}
	}

	imageOrdinals := []string{"satu", "dua", "tiga", "empat", "lima", "enam"}
	images := make(map[string]document.ImageData)
	if record.Report != nil && len(record.Report.Files) > 0 {
		var reportFiles []struct {
			Name      string `json:"name"`
			Type      string `json:"type"`
			Data      string `json:"data"`
			Timestamp string `json:"timestamp"`
		}
		if err := json.Unmarshal(record.Report.Files, &reportFiles); err == nil {
			for i, f := range reportFiles {
				if i >= len(imageOrdinals) {
					break
				}
				docKey := fmt.Sprintf("foto_dokumentasi_%s", imageOrdinals[i])
				images[docKey] = document.ImageData{
					Data:     f.Data,
					MimeType: f.Type,
				}
			}
		}
	}

	for _, ordinal := range imageOrdinals {
		key := fmt.Sprintf("foto_dokumentasi_%s", ordinal)
		if _, ok := images[key]; !ok {
			vars[key] = ""
		}
	}

	payload := document.DocumentRequest{
		TemplateName: "Berkas Luar Kota - Laporan.docx",
		Variables:    vars,
	}

	var docxBytes []byte
	var errDocx error
	if len(images) > 0 {
		docxBytes, errDocx = h.docGen.GenerateDocxWithImages(c.Request().Context(), payload, images)
	} else {
		docxBytes, errDocx = h.docGen.GenerateDocx(c.Request().Context(), payload)
	}
	if errDocx != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Gagal membuat dokumen laporan: %v", errDocx))
	}

	filename := fmt.Sprintf("Laporan_%s_%s.docx", record.Employee.Name, record.SPDNumber)
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Response().Header().Set(echo.HeaderContentType, "application/vnd.openxmlformats-officedocument.wordprocessingml.document")

	return c.Blob(http.StatusOK, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", docxBytes)
}

func (h *Handler) ExportRincianPDF(c echo.Context) error {
	return h.exportDocument(c, "Berkas Luar Kota - rincian pembayaran.docx", "Rincian")
}

func (h *Handler) ExportRincianDocx(c echo.Context) error {
	return h.exportDocumentDocx(c, "Berkas Luar Kota - rincian pembayaran.docx", "Rincian")
}

func (h *Handler) exportDocumentDocx(c echo.Context, templateName string, prefix string) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Format UUID tidak valid")
	}

	record, err := h.svc.GetRecordByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Data tidak ditemukan")
	}

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
		globalIndex = 1
	}

	payload := document.DocumentRequest{
		TemplateName: templateName,
		Variables:    h.mapTravelToDocument(record, globalIndex),
	}

	docxBytes, err := h.docGen.GenerateDocx(c.Request().Context(), payload)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Gagal membuat dokumen: %v", err))
	}

	filename := fmt.Sprintf("%s_%s_%s.docx", prefix, record.Employee.Name, record.SPDNumber)
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Response().Header().Set(echo.HeaderContentType, "application/vnd.openxmlformats-officedocument.wordprocessingml.document")

	return c.Blob(http.StatusOK, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", docxBytes)
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
	return c.JSON(http.StatusOK, map[string]string{"path": filename})
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

	// NEW: Role-based filtering for Protokol users
	// Allow protokol users to see all records within any SPD group they are part of.
	role := c.Get("role").(string)
	if role == "protokol" {
		userIDStr := c.Get("user_id").(string)
		userID, _ := uuid.Parse(userIDStr)
		
		// 1. Get all records to find which SPDs the user belongs to
		allRecs, err := h.svc.GetRecords(map[string]interface{}{})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		
		userSpds := make(map[string]bool)
		for _, r := range allRecs {
			if r.EmployeeID == userID {
				userSpds[r.SPDNumber] = true
			}
		}
		
		// 2. Filter all records to include only those in the user's SPDs
		var filtered []models.TravelRecord
		for _, r := range allRecs {
			if userSpds[r.SPDNumber] {
				filtered = append(filtered, r)
			}
		}
		
		// If status filter is present, apply it manually
		if status != "" {
			var statusFiltered []models.TravelRecord
			for _, r := range filtered {
				if r.Status == status {
					statusFiltered = append(statusFiltered, r)
				}
			}
			filtered = statusFiltered
		}
		
		return c.JSON(http.StatusOK, filtered)
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
	if err := c.Bind(record); err != nil {
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
