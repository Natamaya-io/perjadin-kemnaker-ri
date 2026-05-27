package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/kemnaker/perjadin-backend/internal/config"
)

type FonntePayload struct {
	Target      string `json:"target"`
	Message     string `json:"message"`
	Delay       string `json:"delay"`
	CountryCode string `json:"countryCode"`
	Typing      bool   `json:"typing"`
}

var (
	fonnteMutex sync.Mutex
	lastSent    time.Time
)

// cleanPhoneNumber membersihkan karakter non-angka agar sesuai format yang aman
func cleanPhoneNumber(phone string) string {
	reg := regexp.MustCompile("[^0-9]+")
	cleaned := reg.ReplaceAllString(phone, "")

	// Fonnte menggunakan parameter CountryCode: "62", jadi "08" adalah format standar yang baik.
	if strings.HasPrefix(cleaned, "62") {
		cleaned = "0" + strings.TrimPrefix(cleaned, "62")
	}
	return cleaned
}

func SendWhatsAppMessage(cfg *config.Config, target string, message string) error {
	if cfg.Fonnte.Token == "" {
		return fmt.Errorf("Fonnte token is not configured")
	}

	target = cleanPhoneNumber(target)
	if target == "" || len(target) < 9 {
		return fmt.Errorf("invalid or empty target phone number")
	}

	// 1. Thread-safe Global Rate Limiting (Anti-Ban Utama)
	// Mencegah aplikasi mengirim pesan secara konkuren (bersamaan) yang akan di-flag spam oleh WA/Fonnte.
	fonnteMutex.Lock()
	defer fonnteMutex.Unlock()

	// Menghitung waktu sejak pengiriman pesan terakhir
	elapsed := time.Since(lastSent)

	// Minimal jeda 3 detik antara SEMUA pesan dari server ini
	minDelay := 3 * time.Second
	if elapsed < minDelay {
		time.Sleep(minDelay - elapsed)
	}

	// 2. Jitter / Jeda Acak (Human-like behavior)
	// Menambahkan delay acak 1 sampai 3 detik agar terlihat seperti manusia
	jitter := time.Duration(rand.Intn(3)+1) * time.Second
	time.Sleep(jitter)

	url := "https://api.fonnte.com/send"

	// 3. Konfigurasi Khusus Fonnte
	payload := FonntePayload{
		Target:      target,
		Message:     message,
		Delay:       "2-5", // Fonnte delay internal (2 hingga 5 detik)
		Typing:      true,  // Menampilkan status "sedang mengetik..."
		CountryCode: "62",  // Default ke nomor Indonesia
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", cfg.Fonnte.Token)
	req.Header.Set("Content-Type", "application/json")

	// 4. Strict Timeout
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Do(req)

	// Catat waktu terakhir setelah request dilakukan
	lastSent = time.Now()

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message, status code: %d", resp.StatusCode)
	}

	return nil
}
