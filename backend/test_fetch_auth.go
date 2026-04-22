package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	// Login
	loginPayload := []byte(`{"email":"superadmin", "password":"12345678"}`)
	req, _ := http.NewRequest("POST", "http://localhost:8081/api/v1/auth/login", bytes.NewBuffer(loginPayload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Login Err", err)
		return
	}
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)
	
	var loginResp map[string]interface{}
	json.Unmarshal(b, &loginResp)
	
	token, ok := loginResp["token"].(string)
	if !ok {
		fmt.Println("No token:", string(b))
		return
	}
	
	// Fetch
	req2, _ := http.NewRequest("GET", "http://localhost:8081/api/v1/records", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	resp2, _ := http.DefaultClient.Do(req2)
	defer resp2.Body.Close()
	b2, _ := ioutil.ReadAll(resp2.Body)
	fmt.Println("Raw JSON:", string(b2))
	
	var records []map[string]interface{}
	json.Unmarshal(b2, &records)
	
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
