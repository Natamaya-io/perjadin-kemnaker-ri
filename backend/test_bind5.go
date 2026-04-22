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
	
	// Simulate what frontend sends
	jsonPayload := []byte(`{"reportData": {"text": "hello", "tanggalMerah": ["2026-04-24"]}}`)
	
	err := json.Unmarshal(jsonPayload, r)
	if err != nil {
		fmt.Println("Bind Error:", err)
	}
	
	fmt.Printf("r.Report != nil? %v\n", r.Report != nil)
	if r.Report != nil {
		fmt.Printf("TanggalMerah: %s\n", string(r.Report.TanggalMerah))
	}
}
