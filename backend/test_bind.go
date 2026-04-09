package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
)

type TravelRecord struct {
	SPDNumber   string      `json:"spd"`
	EmployeeID  uuid.UUID   `json:"employeeId"`
	EmployeeIDs []uuid.UUID `json:"employeeIds,omitempty"`
}

func main() {
	payload := `{"spd":"ID-SPJ-001","employeeId":"550e8400-e29b-41d4-a716-446655440000"}`
	var r TravelRecord
	err := json.Unmarshal([]byte(payload), &r)
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Printf("Parsed EmployeeID: %s\n", r.EmployeeID)
}