package chatbot

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type Service interface {
	Ask(ctx context.Context, req AskRequest) (ChatResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func formatRupiah(amount int64) string {
	p := message.NewPrinter(language.Indonesian)
	if amount < 0 {
		return p.Sprintf("-Rp%d", -amount)
	}
	return p.Sprintf("Rp%d", amount)
}

func containsAny(haystack string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(haystack, needle) {
			return true
		}
	}
	return false
}

func (s *service) Ask(ctx context.Context, req AskRequest) (ChatResponse, error) {
	q := strings.ToLower(strings.TrimSpace(req.Message))
	wantsChart := containsAny(q, []string{"chart", "grafik", "diagram", "batang"})
	year := req.Year
	if year == 0 {
		year = time.Now().Year()
	}

	gupSnap, err := s.repo.GetGupSnapshot(ctx, year)
	if err != nil {
		return ChatResponse{}, err
	}
	dalkotSnap, err := s.repo.GetDalkotSnapshot(ctx, year)
	if err != nil {
		return ChatResponse{}, err
	}

	// 1. Detect GUP Menus
	for _, m := range gupSnap.Menus {
		if strings.Contains(q, strings.ToLower(m.Name)) {
			return s.answerGupMenu(m, wantsChart), nil
		}
	}

	if containsAny(q, []string{"petugas spj", "laporan spj"}) {
		return s.answerDalkotOfficers("SPJ", dalkotSnap.PetugasSPJ, wantsChart), nil
	}
	if containsAny(q, []string{"petugas riil", "petugas rill", "laporan riil", "laporan rill"}) {
		return s.answerDalkotOfficers("RIIL", dalkotSnap.PetugasRiil, wantsChart), nil
	}
	if containsAny(q, []string{"proses", "selesai", "sukses", "status dalkot"}) {
		return s.answerDalkotStatus(dalkotSnap, wantsChart), nil
	}
	if containsAny(q, []string{"dalkot", "perjalanan dinas dalam kota"}) {
		return s.answerDalkotFinance(dalkotSnap, wantsChart), nil
	}
	if containsAny(q, []string{"12 menu", "menu gup", "jenis gup", "pilihan gup"}) {
		return s.answerGupMenuList(gupSnap, wantsChart), nil
	}
	if containsAny(q, []string{"gup", "anggaran", "realisasi", "serapan", "keuangan"}) {
		return s.answerGupOverview(gupSnap, wantsChart), nil
	}

	return s.answerCapabilities(gupSnap, dalkotSnap), nil
}

func (s *service) answerGupOverview(snap GupSnapshot, chart bool) ChatResponse {
	var serapan float64
	if snap.TotalPagu > 0 {
		serapan = float64(snap.TotalRealisasi) / float64(snap.TotalPagu) * 100
	}
	ans := fmt.Sprintf(`Berikut adalah ringkasan keseluruhan anggaran GUP:
• Total Transaksi: %d
• Pagu Anggaran: %s
• Total Realisasi: %s
• Sisa Anggaran: %s
• Tingkat Serapan: %.2f%%`,
		snap.TotalTransaksi, formatRupiah(snap.TotalPagu), formatRupiah(snap.TotalRealisasi), formatRupiah(snap.TotalSisa), serapan)

	var c *ChartData
	if chart {
		labels := []string{"Pagu", "Realisasi"}
		values := []interface{}{snap.TotalPagu, snap.TotalRealisasi}
		c = &ChartData{Type: "bar", Title: "Overview GUP", Labels: labels, Values: values, Format: "currency"}
	}

	return ChatResponse{
		Intent: "gup_overview", Title: "Analisis Keuangan GUP", Answer: ans,
		Metrics: []Metric{
			{Label: "Pagu", Value: formatRupiah(snap.TotalPagu)},
			{Label: "Realisasi", Value: formatRupiah(snap.TotalRealisasi)},
		},
		Recommendations: []string{"Telaah menu dengan serapan tinggi."},
		Actions: []Action{
			{Label: "Laporan GUP Keseluruhan", Url: "/print/chatbot?type=ringkasan"},
		},
		Chart:           c,
		GeneratedAt:     time.Now(),
	}
}

func (s *service) answerGupMenuList(snap GupSnapshot, chart bool) ChatResponse {
	lines := []string{}
	labels := []string{}
	values := []interface{}{}
	for i, m := range snap.Menus {
		var serapan float64
		if m.Pagu > 0 {
			serapan = float64(m.Realisasi) / float64(m.Pagu) * 100
		}
		lines = append(lines, fmt.Sprintf("%d. %s — realisasi %s, serapan %.2f%%.", i+1, m.Name, formatRupiah(m.Realisasi), serapan))
		labels = append(labels, m.Name)
		values = append(values, m.Realisasi)
	}

	var c *ChartData
	if chart {
		c = &ChartData{Type: "bar", Title: "Realisasi 12 Menu GUP", Labels: labels, Values: values, Format: "currency"}
	}

	return ChatResponse{
		Intent: "gup_menus", Title: "12 Menu Analisis GUP", Answer: strings.Join(lines, "\n"),
		Recommendations: []string{"Sebutkan nama menu untuk analisis spesifik."},
		Chart:           c, GeneratedAt: time.Now(),
	}
}

func (s *service) answerGupMenu(m GupMenuStat, chart bool) ChatResponse {
	var serapan float64
	if m.Pagu > 0 {
		serapan = float64(m.Realisasi) / float64(m.Pagu) * 100
	}
	risk := "Serapan masih dalam batas pagu."
	if m.Pagu <= 0 {
		risk = "Pagu belum tersedia."
	} else if m.Realisasi > m.Pagu {
		risk = "Realisasi melebihi pagu."
	} else if serapan >= 90 {
		risk = "Serapan tinggi."
	}

	ans := fmt.Sprintf(`Berikut adalah analisis untuk menu pengadaan %s:
• Jumlah Transaksi: %d
• Pagu Tersedia: %s
• Realisasi: %s
• Sisa Anggaran: %s
• Persentase Serapan: %.2f%%

💡 Rekomendasi/Status: %s`,
		m.Name, m.TotalTransaksi, formatRupiah(m.Pagu), formatRupiah(m.Realisasi), formatRupiah(m.Sisa), serapan, risk)

	var c *ChartData
	if chart {
		c = &ChartData{Type: "bar", Title: "Pagu vs Realisasi - " + m.Name,
			Labels: []string{"Pagu", "Realisasi", "Sisa"},
			Values: []interface{}{m.Pagu, m.Realisasi, m.Sisa}, Format: "currency"}
	}

	return ChatResponse{
		Intent: "gup_menu", Title: "Analisis GUP — " + m.Name, Answer: ans,
		Metrics: []Metric{
			{Label: "Transaksi", Value: fmt.Sprintf("%d", m.TotalTransaksi)},
			{Label: "Realisasi", Value: formatRupiah(m.Realisasi)},
			{Label: "Serapan", Value: fmt.Sprintf("%.2f%%", serapan)},
		},
		Recommendations: []string{risk}, Chart: c, GeneratedAt: time.Now(),
		Actions: []Action{
			{Label: "Buka Laporan " + m.Name, Url: "/print/chatbot?type=gup_menu&menu=" + url.QueryEscape(m.Name)},
		},
	}
}

func (s *service) answerDalkotFinance(snap DalkotSnapshot, chart bool) ChatResponse {
	ans := fmt.Sprintf(`Berikut adalah ringkasan laporan keuangan perjalanan dinas dalam kota (Dalkot):
• Total Pengajuan: %d
• Biaya SPJ: %s
• Biaya Riil: %s
• Grand Total Biaya: %s

Sebanyak %d pengajuan sedang diproses dan %d telah selesai.`,
		snap.JumlahPengajuan, formatRupiah(snap.TotalSPJ), formatRupiah(snap.TotalRiil), formatRupiah(snap.TotalBiaya), snap.Draft+snap.Pending, snap.Selesai)

	var c *ChartData
	if chart {
		c = &ChartData{Type: "bar", Title: "Komponen Biaya Dalkot", Labels: []string{"SPJ", "Riil"}, Values: []interface{}{snap.TotalSPJ, snap.TotalRiil}, Format: "currency"}
	}

	return ChatResponse{
		Intent: "dalkot_finance", Title: "Analisis Keuangan Dalkot", Answer: ans,
		Metrics: []Metric{
			{Label: "Total Biaya", Value: formatRupiah(snap.TotalBiaya)},
			{Label: "Dalam Proses", Value: fmt.Sprintf("%d", snap.Draft+snap.Pending)},
			{Label: "Selesai", Value: fmt.Sprintf("%d", snap.Selesai)},
		},
		Recommendations: []string{"Bandingkan biaya SPJ dan riil serta selesaikan pengajuan Draft/Pending sebelum rekap final."},
		Actions: []Action{
			{Label: "Laporan Dalkot", Url: "/print/chatbot?type=dalkot"},
			{Label: "Laporan Status Dalkot", Url: "/print/chatbot?type=dalkot_status"},
		},
		Chart:   c, GeneratedAt: time.Now(),
	}
}

func (s *service) answerDalkotStatus(snap DalkotSnapshot, chart bool) ChatResponse {
	ans := fmt.Sprintf(`Berikut adalah rincian status seluruh pengajuan Dalkot:
• Draft: %d
• Pending (Diproses): %d
• Selesai: %d

Total sedang diproses: %d
Total sudah selesai: %d`,
		snap.Draft, snap.Pending, snap.Selesai, snap.Draft+snap.Pending, snap.Selesai)

	var c *ChartData
	if chart {
		c = &ChartData{Type: "bar", Title: "Status Dalkot", Labels: []string{"Draft", "Pending", "Selesai"}, Values: []interface{}{snap.Draft, snap.Pending, snap.Selesai}, Format: "number"}
	}

	return ChatResponse{
		Intent: "dalkot_status", Title: "Laporan Status Dalkot", Answer: ans, Chart: c, GeneratedAt: time.Now(),
		Recommendations: []string{"Prioritaskan penyelesaian pengajuan Pending sebelum Draft karena sudah lebih dekat ke tahap final."},
		Actions: []Action{
			{Label: "Buka Laporan Proses & Selesai", Url: "/print/chatbot?type=dalkot_status"},
		},
	}
}

func (s *service) answerDalkotOfficers(typ string, officers []PetugasStat, chart bool) ChatResponse {
	var totalAssignments int
	var totalCost int64
	for _, o := range officers {
		totalAssignments += o.JumlahTugas
		totalCost += o.TotalBiaya
	}
	
	lines := []string{
		fmt.Sprintf("Terdapat %d petugas %s dengan %d penugasan dan total biaya %s.", len(officers), typ, totalAssignments, formatRupiah(totalCost)),
	}
	
	count := 8
	if len(officers) < 8 {
		count = len(officers)
	}
	for _, o := range officers[:count] {
		lines = append(lines, fmt.Sprintf("%s: %d tugas, %s, selesai %d, proses %d.", o.Nama, o.JumlahTugas, formatRupiah(o.TotalBiaya), o.Selesai, o.Pending))
	}

	var c *ChartData
	if chart && len(officers) > 0 {
		labels := []string{}
		values := []interface{}{}
		for i, o := range officers {
			if i >= 10 {
				break
			}
			labels = append(labels, o.Nama)
			values = append(values, o.JumlahTugas)
		}
		c = &ChartData{Type: "bar", Title: "Tugas Petugas " + typ, Labels: labels, Values: values, Format: "number"}
	}

	reportType := "dalkot_spj"
	if typ == "RIIL" {
		reportType = "dalkot_riil"
	}

	return ChatResponse{
		Intent: "dalkot_officers", Title: "Laporan Petugas " + typ, Answer: strings.Join(lines, "\n"), Chart: c, GeneratedAt: time.Now(),
		Metrics: []Metric{
			{Label: "Petugas", Value: fmt.Sprintf("%d", len(officers))},
			{Label: "Penugasan", Value: fmt.Sprintf("%d", totalAssignments)},
			{Label: "Total Biaya", Value: formatRupiah(totalCost)},
		},
		Recommendations: []string{"Gunakan laporan detail untuk melihat kegiatan, lokasi, tanggal, dan status masing-masing penugasan."},
		Actions: []Action{
			{Label: "Buka Laporan Petugas " + typ, Url: "/print/chatbot?type=" + reportType},
		},
	}
}

func (s *service) answerCapabilities(gup GupSnapshot, dalkot DalkotSnapshot) ChatResponse {
	ans := "Saya dapat menganalisis:\n• Keuangan GUP secara keseluruhan.\n• Masing-masing menu pengadaan GUP.\n• Keuangan perjalanan dinas (Dalkot).\n• Status pengajuan Dalkot.\n• Beban kerja petugas.\n• Saya juga bisa menampilkan Chart batang jika Anda memintanya."
	return ChatResponse{
		Intent: "capabilities", Title: "Pilihan Analisis Chatbot", Answer: ans, GeneratedAt: time.Now(),
	}
}
