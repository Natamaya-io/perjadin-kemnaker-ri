package terbilang

import "testing"

func TestTerbilang(t *testing.T) {
	tests := []struct {
		name     string
		number   int
		expected string
	}{
		{"Zero", 0, ""}, // Based on the current logic, 0 returns ""
		{"Single digit", 5, "Lima"},
		{"Eleven", 11, "Sebelas"},
		{"Teens", 15, "Lima Belas"},
		{"Tens", 20, "Dua Puluh "},
		{"Tens with units", 45, "Empat Puluh Lima"},
		{"One Hundred", 100, "Seratus "},
		{"Hundreds with tens", 150, "Seratus Lima Puluh "},
		{"Hundreds with units", 505, "Lima Ratus Lima"},
		{"One Thousand", 1000, "Seribu "},
		{"Thousands", 2026, "Dua Ribu Dua Puluh Enam"},
		{"Millions", 1500000, "Satu Juta Lima Ratus  Ribu "}, // Kept double space as this is raw unformatted output
		{"Negative", -15, "Minus Lima Belas"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := Terbilang(tc.number)
			if result != tc.expected {
				t.Errorf("Terbilang(%d) expected %q, got %q", tc.number, tc.expected, result)
			}
		})
	}
}

func TestFormatTerbilang(t *testing.T) {
	tests := []struct {
		name     string
		number   int
		expected string
	}{
		{"Tens trim space", 20, "Dua Puluh"},
		{"Hundreds trim space", 100, "Seratus"},
		{"Thousands trim space", 1000, "Seribu"},
		{"Millions trim space", 1500000, "Satu Juta Lima Ratus Ribu"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatTerbilang(tc.number)
			if result != tc.expected {
				t.Errorf("FormatTerbilang(%d) expected %q, got %q", tc.number, tc.expected, result)
			}
		})
	}
}
