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

func addWorkingDays(t time.Time, days int) time.Time {
	for i := 0; i < days; i++ {
		t = t.AddDate(0, 0, 1)
		for t.Weekday() == time.Saturday || t.Weekday() == time.Sunday {
			t = t.AddDate(0, 0, 1)
		}
	}
	return t
}

func (h *Handler) mapTravelToDocument(record *models.TravelRecord, _ int) map[string]interface{} {
	localIndex := 1
	if allRecords, err := h.svc.GetRecords(map[string]interface{}{}); err == nil {
		sort.Slice(allRecords, func(i, j int) bool {
			timeI := allRecords[i].CreatedAt.UnixNano()
			timeJ := allRecords[j].CreatedAt.UnixNano()
			if timeI != timeJ {
				return timeI < timeJ
			}
			spdI := allRecords[i].SPDNumber
			spdJ := allRecords[j].SPDNumber
			if spdI != spdJ {
				return spdI < spdJ
			}
			nameI := ""
			if allRecords[i].Employee.Name != "" {
				nameI = allRecords[i].Employee.Name
			}
			nameJ := ""
			if allRecords[j].Employee.Name != "" {
				nameJ = allRecords[j].Employee.Name
			}
			if nameI != nameJ {
				return nameI < nameJ
			}
			return allRecords[i].ID.String() < allRecords[j].ID.String()
		})
		for i, r := range allRecords {
			if r.ID == record.ID {
				localIndex = i + 1
				break
			}
		}
	}

	titleCaser := cases.Title(language.Indonesian)
	days := 0
	if !record.StartDate.IsZero() && !record.EndDate.IsZero() {
		days = int(record.EndDate.Sub(record.StartDate).Hours()/24) + 1
	}

	formatRupiahWithRp := func(amount float64) string {
		if amount == 0 { return "-" }
		return fmt.Sprintf("Rp %s", utils.FormatRupiah(amount))
	}
	formatRupiahNoRp := func(amount float64) string {
		if amount == 0 { return "-" }
		return utils.FormatRupiah(amount)
	}
	formatNumber := func(n int) string {
		if n == 0 { return "-" }
		return fmt.Sprintf("%d", n)
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
	spdSubNumber := fmt.Sprintf("%03d", localIndex)
	extractNumericID := func(id string) string {
		if id == "" { return spdSubNumber }
		parts := strings.Split(id, "-")
		if len(parts) > 0 {
			suffix := parts[len(parts)-1]
			// Ensure it's a numeric suffix, if not return the whole suffix or spdSubNumber
			if len(suffix) < 3 { return fmt.Sprintf("%03s", suffix) }
			return suffix
		}
		return id
	}

	noSurat := strings.TrimSpace(record.SuratTugasNumber)
	tglSurat := record.SuratTugasDate
	numberGap := "\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0"
	dateGap := "\u00A0\u00A0\u00A0\u00A0\u00A0"
	if noSurat == "" {
		noSurat = numberGap // Ruang kosong untuk diisi manual
	} else {
		groupID := extractNumericID(record.SPDNumber)
		if groupID != "" && groupID != spdSubNumber {
			noSurat = strings.Replace(noSurat, groupID, spdSubNumber, 1)
		}
	}
	noSpd := record.SPDNumber
	cleanMaksud := regexp.MustCompile(`\s+`).ReplaceAllString(record.Purpose, " ")
	reKunjungan := regexp.MustCompile(`(?i)\s*kunjungan kerja\s*`)
	cleanMaksud = strings.TrimSpace(reKunjungan.ReplaceAllString(cleanMaksud, " "))
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

	terbilangHari := "-"
	if days > 0 {
		terbilangHari = strings.ToLower(terbilang.FormatTerbilang(days))
	}

	tingkatBiaya := record.Employee.TingkatBiaya
	if tingkatBiaya == "" {
		tingkatBiaya = "C"
	}

	pejabatBerwenang := "KPA Biro Umum Sekretariat Jenderal Kemnaker"
	instansi := "Kementerian Ketenagakerjaan RI"
	akunAnggaran := "-" // Can be updated if database field is added

	pangkat := record.Employee.Pangkat
	if pangkat == "" { pangkat = "-" }
	golongan := record.Employee.Golongan
	if golongan == "" { golongan = "-" }

	now := time.Now()
	refDate := now
	tanggalLaporan := now
	tanggalRincian := now
	if !record.StartDate.IsZero() {
		refDate = record.StartDate
	}
	if !record.EndDate.IsZero() {
		tanggalLaporan = addWorkingDays(record.EndDate, 1)
		tanggalRincian = addWorkingDays(record.EndDate, 3)
	}

	tanggalNoSuratStr := fmt.Sprintf("%s %s %d", dateGap, utils.GetIndonesianMonths()[int(refDate.Month())], refDate.Year())
	bulanRomawiST := utils.GetRomanMonths()[int(refDate.Month())]
	tahunST := refDate.Year()
	if !tglSurat.IsZero() {
		tanggalNoSuratStr = utils.FormatIndonesianDate(tglSurat)
		bulanRomawiST = utils.GetRomanMonths()[int(tglSurat.Month())]
		tahunST = tglSurat.Year()
	}

	vars := map[string]interface{}{
		"bulan_romawi_st": bulanRomawiST,
		"tahun_st":        tahunST,
		"no_spd": spdSubNumber, "id_spj": noSpd, "no_surat": noSurat,
		"pejabat_berwenang": pejabatBerwenang,
		"tingkat_biaya": tingkatBiaya,
		"instansi": instansi,
		"akun_anggaran": akunAnggaran,
		"tempat_berangkat": "Jakarta",
		"tempat_tujuan": dest,
		"kota_atau_kabupaten": dest,
		"keterangan": "-",
		"lama_perjalanan": formatNumber(days),
		"nama_stakeholder": cleanStakeholder,
		"bulan_no_surat": utils.GetRomanMonths()[int(refDate.Month())], "tahun_no_surat": refDate.Year(),
		"bulan_romawi": utils.GetRomanMonths()[int(refDate.Month())],
		"tanggal_no_surat": tanggalNoSuratStr, "bulan_pembayaran": utils.GetIndonesianMonths()[int(refDate.Month())],
		"tahun_pembayaran": refDate.Year(), "tahun_saat_ini": fmt.Sprintf("%d", refDate.Year()),
		"nama": record.Employee.Name, "nama_petugas": record.Employee.Name, "nip": record.Employee.NIP, "nip_petugas": record.Employee.NIP,
		"pangkat": pangkat, "golongan": golongan,
		"pangkat_gol": fmt.Sprintf("%s (%s)", pangkat, golongan), "jabatan": record.Employee.Jabatan,
		"maksud_perjalanan": cleanMaksud, "tujuan_perjalanan": cleanMaksud, "stakeholder": cleanStakeholder,
		"tujuan": dest, "kota": dest, "provinsi": prov, "transportasi": transportMode,
		"tanggal_berangkat": utils.FormatIndonesianDate(record.StartDate), "tanggal_mulai": tanggalMulaiStr,
		"tanggal_selesai": utils.FormatIndonesianDate(record.EndDate), "lama_hari": formatNumber(days),
		"terbilang": terbilangHari,
		"bulan": utils.GetIndonesianMonths()[int(record.StartDate.Month())], "tahun": fmt.Sprintf("%d", record.StartDate.Year()),
		"tgl_cetak": utils.FormatIndonesianDate(now), "nama_ppk": namaPpk, "nip_ppk": nipPpk,
		"nama_bendahara": namaBendahara, "nip_bendahara": nipBendahara, "jabatan_ppk": jabPpk, "jabatan_bendahara": jabBendahara,
		"isi_laporan": isiLaporan, "tanggal_dikeluarkan": utils.FormatIndonesianDate(now),
		"tanggal_laporan": utils.FormatIndonesianDate(tanggalLaporan), "tanggal_rincian": utils.FormatIndonesianDate(tanggalRincian),
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
			HotelDays int `json:"hotelDays"`
			HotelRate float64 `json:"hotelRate"`
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
			vars[fmt.Sprintf("hari_%s", ordinal)] = formatNumber(locDays)
			var sbmRate float64
			if record.Cost != nil { sbmRate = record.Cost.DailyAllowanceRate }
			vars[fmt.Sprintf("sbm_rate_%s", ordinal)] = formatRupiahNoRp(sbmRate)
			if i < len(costDetails) {
				cd := costDetails[i]
				vars[fmt.Sprintf("hotel_rate_%s", ordinal)] = formatRupiahNoRp(cd.HotelRate)
				vars[fmt.Sprintf("tiket_%s", ordinal)] = formatRupiahNoRp(cd.TicketGo + cd.TicketBack)
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
				} else if ac.Name == "Extend Penginapan" {
					aggHotel += float64(ac.HotelDays) * ac.HotelRate
					hotelDays += ac.HotelDays
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
		if hotelDays == 0 { hotelDays = record.Cost.HotelDays }
		aggSbm := record.Cost.DailyAllowanceRate * float64(totalDays)
		totalAgg := aggTicket + aggLokal + aggSbm + aggHotel + aggTambahan
		
		penginapanRate := record.Cost.HotelRate
		if hotelDays > 0 { penginapanRate = aggHotel / float64(hotelDays) }

		vars["total_biaya"] = formatRupiahWithRp(totalAgg)
		vars["biaya_hotel"] = formatRupiahWithRp(aggHotel)
		vars["biaya_pesawat"] = formatRupiahWithRp(aggTicket)
		if totalAgg > 0 {
			vars["terbilang_biaya"] = fmt.Sprintf("%s RUPIAH", strings.ToUpper(terbilang.FormatTerbilang(int(totalAgg))))
		} else {
			vars["terbilang_biaya"] = "-"
		}

		// Rincian specific variables
		vars["tiket_pesawat"] = formatRupiahNoRp(aggTicket)
		vars["transport_lokal"] = formatRupiahNoRp(aggTambahan)
		vars["transport_daerah"] = formatRupiahNoRp(aggLokal)
		vars["sbm"] = formatRupiahNoRp(record.Cost.DailyAllowanceRate)
		vars["total_sbm"] = formatRupiahNoRp(aggSbm)
		
		if hotelDays > 0 {
			vars["durasi_hotel"] = hotelDays
		} else {
			vars["durasi_hotel"] = "-"
		}
		
		vars["p"] = formatRupiahNoRp(penginapanRate)
		vars["total_penginapan"] = formatRupiahNoRp(aggHotel)
		vars["total"] = formatRupiahNoRp(totalAgg)
		
		if totalAgg > 0 {
			vars["total_terbilang"] = strings.ToUpper(terbilang.FormatTerbilang(int(totalAgg))) + " RUPIAH"
		} else {
			vars["total_terbilang"] = "-"
		}
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
		nameI := strings.ToLower(strings.TrimSpace(spdGroupRecords[i].Employee.Name))
		nameJ := strings.ToLower(strings.TrimSpace(spdGroupRecords[j].Employee.Name))

		getPriority := func(name string) int {
			if strings.Contains(name, "auditya hermawan") {
				return 1
			}
			if strings.Contains(name, "mochamad gufron") {
				return 2
			}
			if strings.Contains(name, "muhammad isa") {
				return 3
			}
			return 4
		}

		pI := getPriority(nameI)
		pJ := getPriority(nameJ)

		if pI != pJ {
			return pI < pJ
		}

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
	vars["tanggal_dikeluarkan"] = vars["tanggal_laporan"]
	vars["tgl_cetak"] = vars["tanggal_laporan"]
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
	vars["tanggal_dikeluarkan"] = vars["tanggal_laporan"]
	vars["tgl_cetak"] = vars["tanggal_laporan"]
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

	vars := h.mapTravelToDocument(record, 1)
	if prefix == "Laporan" {
		vars["tanggal_dikeluarkan"] = vars["tanggal_laporan"]
		vars["tgl_cetak"] = vars["tanggal_laporan"]
	} else if prefix == "Rincian" {
		vars["tanggal_dikeluarkan"] = vars["tanggal_rincian"]
		vars["tgl_cetak"] = vars["tanggal_rincian"]
		vars["bulan"] = ""
		vars["tahun"] = ""
	}
	payload := document.DocumentRequest{TemplateName: templateName, Variables: vars}
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

	vars := h.mapTravelToDocument(record, 1)
	if prefix == "Laporan" {
		vars["tanggal_dikeluarkan"] = vars["tanggal_laporan"]
		vars["tgl_cetak"] = vars["tanggal_laporan"]
	} else if prefix == "Rincian" {
		vars["tanggal_dikeluarkan"] = vars["tanggal_rincian"]
		vars["tgl_cetak"] = vars["tanggal_rincian"]
		vars["bulan"] = ""
		vars["tahun"] = ""
	}
	payload := document.DocumentRequest{TemplateName: templateName, Variables: vars}
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
	userIDStr := c.Get("user_id").(string)
	
	fmt.Printf("[DEBUG] GetRecords called by user=%s, role=%s\n", userIDStr, role)
	
	recs, err := h.svc.GetRecords(nil)
	if err != nil {
		fmt.Printf("[ERROR] GetRecords service failed: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	
	fmt.Printf("[DEBUG] Total records found in DB: %d\n", len(recs))

	if role == "protokol" {
		uid, err := uuid.Parse(userIDStr)
		if err != nil {
			fmt.Printf("[ERROR] Failed to parse userID from context: %v\n", err)
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid User ID")
		}
		
		spds := make(map[string]bool)
		for _, r := range recs { 
			// Protokol can see records where they are the employee OR the creator
			if r.EmployeeID == uid || r.CreatorID == uid { 
				spds[r.SPDNumber] = true 
			} 
		}
		
		res := make([]models.TravelRecord, 0)
		for _, r := range recs { 
			if spds[r.SPDNumber] { 
				res = append(res, r) 
			} 
		}
		
		fmt.Printf("[DEBUG] Filtered records for protokol %s: %d\n", userIDStr, len(res))
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
