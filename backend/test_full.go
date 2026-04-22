package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	// 1. Get Records
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
		fmt.Println("No records found, creating one...")
		createPayload := []byte(`{"employeeId":"1d728fc8-a21c-4f95-93b2-6d75a42a50be", "purpose":"Testing", "location":"Jakarta", "province":"DKI Jakarta"}`)
		reqCreate, _ := http.NewRequest("POST", "http://localhost:8081/api/v1/records", bytes.NewBuffer(createPayload))
		reqCreate.Header.Set("Content-Type", "application/json")
		reqCreate.Header.Set("User-ID", "1d728fc8-a21c-4f95-93b2-6d75a42a50be")
		reqCreate.Header.Set("Role", "super_admin")
		http.DefaultClient.Do(reqCreate)

		// Fetch again
		resp, _ = http.DefaultClient.Do(req)
		b, _ = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		json.Unmarshal(b, &records)
	}
	id := records[0]["id"].(string)
	fmt.Println("Testing on ID:", id)
	
	// 2. PUT Update
	payload := []byte(`{"reportStatus":"Completed","reportData":{"text":"testing","tanggalMerah":["2026-04-24"]}}`)
	reqPut, _ := http.NewRequest("PUT", "http://localhost:8081/api/v1/records/"+id, bytes.NewBuffer(payload))
	reqPut.Header.Set("Content-Type", "application/json")
	reqPut.Header.Set("User-ID", "1d728fc8-a21c-4f95-93b2-6d75a42a50be")
	reqPut.Header.Set("Role", "super_admin")
	
	respPut, _ := http.DefaultClient.Do(reqPut)
	bPut, _ := ioutil.ReadAll(respPut.Body)
	respPut.Body.Close()
	fmt.Println("PUT response:", string(bPut))
	
	// 3. GET Again
	reqGet, _ := http.NewRequest("GET", "http://localhost:8081/api/v1/records/"+id, nil)
	reqGet.Header.Set("User-ID", "1d728fc8-a21c-4f95-93b2-6d75a42a50be")
	reqGet.Header.Set("Role", "super_admin")
	respGet, _ := http.DefaultClient.Do(reqGet)
	bGet, _ := ioutil.ReadAll(respGet.Body)
	respGet.Body.Close()
	fmt.Println("GET response:", string(bGet))
}
