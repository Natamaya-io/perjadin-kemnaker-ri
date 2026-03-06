package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kemnaker/perjadin-backend/internal/config"
)

type FonntePayload struct {
        Target      string `json:"target"`
        Message     string `json:"message"`
        Delay       string `json:"delay"`
        Typing      bool   `json:"typing"`
        CountryCode string `json:"countryCode"`
}

func SendWhatsAppMessage(cfg *config.Config, target string, message string) error {
	if cfg.Fonnte.Token == "" {
		return fmt.Errorf("Fonnte token is not configured")
	}

	if target == "" {
		return fmt.Errorf("target phone number is empty")
	}

	url := "https://api.fonnte.com/send"

	        payload := FonntePayload{
                Target:      target,
                Message:     message,
                Delay:       "2-5",
                Typing:      true,
                CountryCode: "62",
        }

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", cfg.Fonnte.Token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message, status code: %d", resp.StatusCode)
	}

	return nil
}


