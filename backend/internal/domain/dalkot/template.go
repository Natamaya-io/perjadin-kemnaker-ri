package dalkot

import (
	"bytes"
	"html/template"
	"strings"

	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/kemnaker/perjadin-backend/internal/utils"
)

var dalkotLaporanTmpl = `
<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Laporan Perjadin Dalkot {{ .Record.SPDNumber }}</title>
    <style>
@page { size: A4; margin: 0; }
* { box-sizing: border-box; }
body { margin: 0; color: #000000; background: #ffffff; font-family: Arial, Helvetica, sans-serif; font-size: 11px; line-height: 1.4; }
.paper { width: 100%; max-width: 210mm; padding: 16mm 18mm 18mm; margin: 0 auto; min-height: 297mm; }
.wide-paper { max-width: 100%; }
.print-header { display: flex; align-items: center; gap: 12px; border-bottom: 2px solid #0f2d52; padding-bottom: 18px; margin-bottom: 6px; }
.print-header .print-logo { width: 48px; height: auto; object-fit: contain; }
.print-header strong { display: block; font-size: 12px; color: #0f2d52; }
.print-header span { display: block; font-weight: 700; color: #374151; font-size: 10px; }
.print-title { text-align: center; margin: 22px 0 18px; }
.print-title h1 { margin: 0 0 4px; font-size: 15px; color: #0f2d52; }
.print-title p { margin: 0; font-size: 10px; color: #4b5563; }
.left-title { text-align: left; margin-bottom: 14px; }
.info-table, .cost-table { border-collapse: collapse; width: 100%; }
.info-table th, .info-table td, .cost-table th, .cost-table td { border: 1px solid #1f2937; padding: 7px 8px; vertical-align: top; }
.info-table th { width: 18%; text-align: left; background: #eef3f8; color: #0f2d52; }
.cost-table th { background: #dce9f8; color: #0f2d52; text-align: center; font-weight: 700; }
.cost-table .total td, .cost-table tr.total td { background: #eef3f8; font-weight: 700; }
.print-section { margin-top: 18px; }
.print-section h2 { font-size: 11px; color: #0f2d52; margin: 0 0 7px; }
.print-section p { margin: 0; }
.report-text { min-height: 64px; border: 1px solid #9ca3af; padding: 8px; }
.signature { display: flex; justify-content: space-between; gap: 30px; margin-top: 28px; text-align: center; }
.signature > div { flex: 1; }
.signature p { margin: 0; }
.signature-space { height: 58px; }
.signature-single { justify-content: flex-end; }
.signature-single > div { flex: 0 0 220px; }
[contenteditable="true"] { outline: none; }
@media print { [contenteditable="true"] { border-bottom-color: transparent; } }
    </style>
</head>
<body onload="window.print()">
<main class="paper">
    <header class="print-header">
        <img src="/kemnaker-ri-logo-only.webp" alt="Logo Kemnaker" class="print-logo">
        <div><strong>KEMENTERIAN KETENAGAKERJAAN REPUBLIK INDONESIA</strong><span>SEKRETARIAT JENDERAL</span></div>
    </header>
    <div class="print-title">
        <h1>LAPORAN PERJALANAN DINAS DALAM KOTA</h1>
        <p>Nomor Pengajuan: {{ .Record.SPDNumber }}</p>
    </div>

    <table class="info-table">
        <tr>
            <th>Tanggal Pelaksanaan</th>
            <td>{{ formatIndoDate .Record.ExecutionDate }}</td>
            <th>Pejabat Didampingi</th>
            <td>{{ .Record.Official }}</td>
        </tr>
        <tr>
            <th>Lokasi</th>
            <td>{{ .Record.Location }}</td>
            <th>Kategori Dalkot</th>
            <td>{{ .Record.Category }}</td>
        </tr>
        <tr>
            <th>Jenis Dalkot</th>
            <td colspan="3">{{ .Record.DalkotType }}</td>
        </tr>
    </table>

    <section class="print-section">
        <h2>Nama Kegiatan</h2>
        <p>{{ nl2br .Record.ActivityName }}</p>
    </section>

    {{ if .HasSPJ }}
    <section class="print-section">
        <h2>Rincian Petugas SPJ</h2>
        <table class="cost-table">
            <thead>
                <tr>
                    <th>No.</th>
                    <th>Nama Petugas</th>
                    <th>Biaya SPJ</th>
                    <th>Status</th>
                </tr>
            </thead>
            <tbody>
                {{ range $i, $a := .SPJAssignments }}
                <tr>
                    <td>{{ inc $i }}</td>
                    <td>{{ if $a.User }}{{ $a.User.Name }}{{ else }}-{{ end }}</td>
                    <td>{{ formatRupiah $a.SPJCost }}</td>
                    <td>{{ $a.Status }}</td>
                </tr>
                {{ end }}
                <tr class="total">
                    <td colspan="2">TOTAL SPJ</td>
                    <td>{{ formatRupiah .TotalSPJ }}</td>
                    <td></td>
                </tr>
            </tbody>
        </table>
    </section>
    {{ end }}

    <section class="print-section">
        <h2>Rincian Petugas Riil</h2>
        <table class="cost-table">
            <thead>
                <tr>
                    <th>No.</th>
                    <th>Nama Petugas</th>
                    <th>Biaya Riil</th>
                    <th>Status</th>
                </tr>
            </thead>
            <tbody>
                {{ range $i, $a := .RiilAssignments }}
                <tr>
                    <td>{{ inc $i }}</td>
                    <td>{{ if $a.User }}{{ $a.User.Name }}{{ else }}-{{ end }}</td>
                    <td>{{ formatRupiah $a.ActualCost }}</td>
                    <td>{{ $a.Status }}</td>
                </tr>
                {{ end }}
                <tr class="total">
                    <td colspan="2">TOTAL RIIL</td>
                    <td>{{ formatRupiah .TotalRiil }}</td>
                    <td></td>
                </tr>
            </tbody>
        </table>
    </section>

    {{ if eq .Record.DalkotType "SPJ RIIL" }}
    <section class="print-section">
        <h2>Isi Laporan</h2>
        <p class="report-text">{{ if .Record.ReportContent }}{{ nl2br .Record.ReportContent }}{{ else }}-{{ end }}</p>
    </section>
    {{ end }}

    <section class="signature">
        <div>
            <p>Mengetahui/Menyetujui,</p>
            <p>Pejabat Pembuat Komitmen</p>
            <div class="signature-space"></div>
            <strong contenteditable="true" style="display: block;">{{ if .PPKName }}{{ .PPKName }}{{ else }}____________________________{{ end }}</strong>
            <span contenteditable="true">NIP. {{ if .PPKNIP }}{{ .PPKNIP }}{{ else }}____________________________{{ end }}</span>
        </div>
        <div>
            <p>Jakarta, {{ formatIndoDate .Record.ExecutionDate }}</p>
            <p>Petugas Pelaksana</p>
            <div class="signature-space"></div>
            <strong contenteditable="true" style="display: block;">{{ .PetugasPenandatangan }}</strong>
            <span contenteditable="true">NIP. {{ if .PetugasNIP }}{{ .PetugasNIP }}{{ else }}-{{ end }}</span>
        </div>
    </section>
</main>
</body>
</html>
`

func generateLaporanHTML(record *models.DalkotRecord, ppkName string, ppkNIP string) (string, error) {
	funcMap := template.FuncMap{
		"formatIndoDate": utils.FormatIndonesianDate,
		"formatRupiah": func(f float64) string {
			return "Rp " + utils.FormatRupiah(f)
		},
		"nl2br": func(text string) template.HTML {
			return template.HTML(strings.ReplaceAll(template.HTMLEscapeString(text), "\n", "<br>"))
		},
		"inc": func(i int) int {
			return i + 1
		},
	}

	tmpl, err := template.New("laporan").Funcs(funcMap).Parse(dalkotLaporanTmpl)
	if err != nil {
		return "", err
	}

	var spjAssignments []models.DalkotAssignment
	var riilAssignments []models.DalkotAssignment
	var totalSPJ float64
	var totalRiil float64

	petugasPenandatangan := ""
	petugasNIP := ""

	for _, a := range record.Assignments {
		if a.AssignmentType == "SPJ" {
			spjAssignments = append(spjAssignments, a)
			totalSPJ += a.SPJCost
		} else {
			riilAssignments = append(riilAssignments, a)
			totalRiil += a.ActualCost
		}
		if petugasPenandatangan == "" && a.User != nil {
			petugasPenandatangan = a.User.Name
			petugasNIP = a.User.NIP
		}
	}

	data := struct {
		Record               *models.DalkotRecord
		SPJAssignments       []models.DalkotAssignment
		RiilAssignments      []models.DalkotAssignment
		TotalSPJ             float64
		TotalRiil            float64
		HasSPJ               bool
		PetugasPenandatangan string
		PetugasNIP           string
		PPKName              string
		PPKNIP               string
	}{
		Record:               record,
		SPJAssignments:       spjAssignments,
		RiilAssignments:      riilAssignments,
		TotalSPJ:             totalSPJ,
		TotalRiil:            totalRiil,
		HasSPJ:               len(spjAssignments) > 0,
		PetugasPenandatangan: petugasPenandatangan,
		PetugasNIP:           petugasNIP,
		PPKName:              ppkName,
		PPKNIP:               ppkNIP,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

var dalkotDprTmpl = `
<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>DPR Dalkot {{ .Record.SPDNumber }}</title>
    <style>
@page { size: A4; margin: 0; }
* { box-sizing: border-box; }
body { margin: 0; color: #000000; background: #ffffff; font-family: Arial, Helvetica, sans-serif; font-size: 11px; line-height: 1.4; }
.paper { width: 100%; max-width: 210mm; padding: 16mm 18mm 18mm; margin: 0 auto; min-height: 297mm; }
.print-title { text-align: center; margin: 22px 0 18px; }
.print-title h1 { margin: 0 0 4px; font-size: 15px; color: #0f2d52; text-decoration: underline; }
.print-title p { margin: 0; font-size: 10px; color: #4b5563; }
.info-table, .cost-table { border-collapse: collapse; width: 100%; margin-top: 15px; }
.info-table th, .info-table td, .cost-table th, .cost-table td { border: 1px solid #1f2937; padding: 7px 8px; vertical-align: top; }
.info-table th { width: 30%; text-align: left; background: #eef3f8; color: #0f2d52; }
.cost-table th { background: #dce9f8; color: #0f2d52; text-align: center; font-weight: 700; }
.cost-table .total td { background: #eef3f8; font-weight: 700; }
.print-section { margin-top: 18px; }
.signature { display: flex; justify-content: space-between; gap: 30px; margin-top: 40px; text-align: center; }
.signature > div { flex: 1; }
.signature p { margin: 0; }
.signature-space { height: 58px; }
.page-break { page-break-after: always; }
[contenteditable="true"] { outline: none; }
@media print { [contenteditable="true"] { border-bottom-color: transparent; } }
    </style>
</head>
<body onload="window.print()">
{{ range $index, $assignment := .RiilAssignments }}
<main class="paper {{ if not (isLast $index $.TotalRiilCount) }}page-break{{ end }}">
    <div class="print-title">
        <h1>DAFTAR PENGELUARAN RIIL</h1>
        <p>Nomor: {{ $.Record.SPDNumber }}</p>
    </div>
    
    <p>Yang bertanda tangan di bawah ini:</p>
    <table class="info-table" style="border: none;">
        <tr style="border: none;">
            <td style="border: none; width: 25%;">Nama</td>
            <td style="border: none;" contenteditable="true">: {{ if $assignment.User }}{{ $assignment.User.Name }}{{ else }}-{{ end }}</td>
        </tr>
        <tr style="border: none;">
            <td style="border: none;">NIP</td>
            <td style="border: none;" contenteditable="true">: {{ if $assignment.User }}{{ if $assignment.User.NIP }}{{ $assignment.User.NIP }}{{ else }}-{{ end }}{{ else }}-{{ end }}</td>
        </tr>
        <tr style="border: none;">
            <td style="border: none;">Jabatan</td>
            <td style="border: none;" contenteditable="true">: {{ if $assignment.User }}{{ if $assignment.User.Jabatan }}{{ $assignment.User.Jabatan }}{{ else }}-{{ end }}{{ else }}-{{ end }}</td>
        </tr>
    </table>
    
    <p style="margin-top: 15px;">Berdasarkan Surat Tugas / SPD Nomor {{ $.Record.SPDNumber }} tanggal {{ formatIndoDate $.Record.ExecutionDate }}, dengan ini kami menyatakan dengan sesungguhnya bahwa:</p>
    <ol>
        <li>Biaya transport pegawai dan/atau biaya penginapan di bawah ini yang tidak dapat diperoleh bukti-bukti pengeluarannya, meliputi:</li>
    </ol>
    
    <table class="cost-table">
        <thead>
            <tr>
                <th>No.</th>
                <th>Uraian Pengeluaran</th>
                <th>Jumlah (Rp)</th>
            </tr>
        </thead>
        <tbody>
            <tr>
                <td style="text-align: center;">1</td>
                <td>Biaya Operasional (Transportasi, dll) - Dalam Kota</td>
                <td style="text-align: right;">{{ formatRupiah $assignment.ActualCost }}</td>
            </tr>
            <tr class="total">
                <td colspan="2" style="text-align: center;">Jumlah</td>
                <td style="text-align: right;">{{ formatRupiah $assignment.ActualCost }}</td>
            </tr>
        </tbody>
    </table>
    
    <ol start="2" style="margin-top: 15px;">
        <li>Jumlah uang tersebut pada angka 1 di atas benar-benar dikeluarkan untuk pelaksanaan Perjalanan Dinas dimaksud dan apabila di kemudian hari terdapat kelebihan atas pembayaran, kami bersedia untuk menyetorkan kelebihan tersebut ke Kas Negara.</li>
    </ol>
    
    <p>Demikian pernyataan ini kami buat dengan sebenarnya, untuk dipergunakan sebagaimana mestinya.</p>
    
    <section class="signature">
        <div>
            <p>Mengetahui/Menyetujui,</p>
            <p>Pejabat Pembuat Komitmen</p>
            <div class="signature-space"></div>
            <strong contenteditable="true" style="display: block;">{{ if $.PPKName }}{{ $.PPKName }}{{ else }}____________________________{{ end }}</strong>
            <span contenteditable="true">NIP. {{ if $.PPKNIP }}{{ $.PPKNIP }}{{ else }}____________________________{{ end }}</span>
        </div>
        <div>
            <p>Jakarta, {{ formatIndoDate $.Record.ExecutionDate }}</p>
            <p>Petugas Pelaksana</p>
            <div class="signature-space"></div>
            <strong contenteditable="true" style="display: block;">{{ if $assignment.User }}{{ $assignment.User.Name }}{{ else }}-{{ end }}</strong>
            <span contenteditable="true">NIP. {{ if $assignment.User }}{{ if $assignment.User.NIP }}{{ $assignment.User.NIP }}{{ else }}-{{ end }}{{ else }}-{{ end }}</span>
        </div>
    </section>
</main>
{{ end }}
</body>
</html>
`

func generateDprHTML(record *models.DalkotRecord, ppkName string, ppkNIP string) (string, error) {
	funcMap := template.FuncMap{
		"formatIndoDate": utils.FormatIndonesianDate,
		"formatRupiah": func(f float64) string {
			return utils.FormatRupiah(f) // without Rp for right align
		},
		"isLast": func(index, total int) bool {
			return index == total-1
		},
	}

	tmpl, err := template.New("dpr").Funcs(funcMap).Parse(dalkotDprTmpl)
	if err != nil {
		return "", err
	}

	var riilAssignments []models.DalkotAssignment

	for _, a := range record.Assignments {
		if a.AssignmentType == "RIIL" && a.ActualCost > 0 {
			riilAssignments = append(riilAssignments, a)
		}
	}

	data := struct {
		Record          *models.DalkotRecord
		RiilAssignments []models.DalkotAssignment
		TotalRiilCount  int
		PPKName         string
		PPKNIP          string
	}{
		Record:          record,
		RiilAssignments: riilAssignments,
		TotalRiilCount:  len(riilAssignments),
		PPKName:         ppkName,
		PPKNIP:          ppkNIP,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func ExportDprHTMLMock(record *models.DalkotRecord) (string, error) {
	return generateDprHTML(record, "", "")
}
