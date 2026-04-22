package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	req, _ := http.NewRequest("GET", "http://localhost:8081/api/v1/records", nil)
	req.Header.Set("User-ID", "1d728fc8-a21c-4f95-93b2-6d75a42a50be") // ID for super_admin
	req.Header.Set("Role", "super_admin")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Err", err)
		return
	}
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)
	
	var records []map[string]interface{}
	json.Unmarshal(b, &records)
	if len(records) == 0 {
		fmt.Println("No records found")
		return
	}
	
	for _, r := range records {
		reportData, ok := r["reportData"].(map[string]interface{})
		if !ok {
			fmt.Println("No reportData in record", r["id"])
			continue
		}
		tm := reportData["tanggalMerah"]
		fmt.Printf("Record %s tanggalMerah: type=%T value=%v\n", r["id"], tm, tm)
	}
}
