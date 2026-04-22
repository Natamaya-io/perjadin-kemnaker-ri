package main

import (
	"encoding/json"
	"fmt"
)

type TravelReport struct {
	Text         string          `json:"text"`
	TanggalMerah json.RawMessage `json:"tanggalMerah"`
}

type TravelRecord struct {
	Report *TravelReport `json:"reportData,omitempty"`
}

func main() {
	r := &TravelRecord{}
	
	jsonPayload := []byte(`{"reportData": {"text": "new text", "tanggalMerah": ["2026-04-24"]}}`)
	err := json.Unmarshal(jsonPayload, r)
	if err != nil {
		fmt.Println("Error:", err)
	}
	
	fmt.Printf("TanggalMerah: %s\n", string(r.Report.TanggalMerah))
}
