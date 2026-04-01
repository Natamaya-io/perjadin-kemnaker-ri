package terbilang

import "strings"

var (
	huruf = []string{"", "Satu", "Dua", "Tiga", "Empat", "Lima", "Enam", "Tujuh", "Delapan", "Sembilan", "Sepuluh", "Sebelas"}
)

func Terbilang(n int) string {
	if n < 0 {
		return "Minus " + Terbilang(-n)
	}
	if n < 12 {
		return huruf[n]
	} else if n < 20 {
		return Terbilang(n-10) + " Belas"
	} else if n < 100 {
		return Terbilang(n/10) + " Puluh " + Terbilang(n%10)
	} else if n < 200 {
		return "Seratus " + Terbilang(n-100)
	} else if n < 1000 {
		return Terbilang(n/100) + " Ratus " + Terbilang(n%100)
	} else if n < 2000 {
		return "Seribu " + Terbilang(n-1000)
	} else if n < 1000000 {
		return Terbilang(n/1000) + " Ribu " + Terbilang(n%1000)
	} else if n < 1000000000 {
		return Terbilang(n/1000000) + " Juta " + Terbilang(n%1000000)
	} else if n < 1000000000000 {
		return Terbilang(n/1000000000) + " Miliar " + Terbilang(n%1000000000)
	}
	return ""
}

func FormatTerbilang(n int) string {
	return strings.TrimSpace(Terbilang(n))
}
