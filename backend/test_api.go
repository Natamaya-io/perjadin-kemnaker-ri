package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	// Let's find an ID. We can just hit GET /api/v1/records and get an ID.
	req1, _ := http.NewRequest("GET", "http://localhost:8081/api/v1/records", nil)
	req1.Header.Set("User-Agent", "Test")
	req1.Header.Set("Authorization", "Bearer MOCK") // We need valid auth or we can use the test server directly?
	// Oh, we can just use the backend API if we bypass auth. The logs show Auth failed.
}
