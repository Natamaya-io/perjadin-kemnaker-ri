package main

import (
	"encoding/json"
	"fmt"
	"github.com/kemnaker/perjadin-backend/internal/models"
)

func main() {
	payload := []byte(`{"reportData": {"tanggalMerah": ["2026-04-22"]}}`)
	var r models.TravelRecord
	err := json.Unmarshal(payload, &r)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	
	if r.Report != nil {
		fmt.Printf("TanggalMerah bytes: %s\n", string(r.Report.TanggalMerah))
	} else {
		fmt.Println("Report is nil")
	}
}
