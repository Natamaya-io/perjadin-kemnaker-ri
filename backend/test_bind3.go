package main
import (
	"encoding/json"
	"fmt"
)
func main() {
	var rawMap map[string]interface{}
	jsonPayload := []byte(`{"reportData": {"tanggalMerah": ["2026-04-24"]}}`)
	json.Unmarshal(jsonPayload, &rawMap)
	
	if rd, ok := rawMap["reportData"].(map[string]interface{}); ok {
		if tm, exists := rd["tanggalMerah"]; exists {
			fmt.Printf("exists! tm=%v\n", tm)
			if tm != nil {
				b, _ := json.Marshal(tm)
				fmt.Printf("marshaled: %s\n", string(b))
			}
		} else {
			fmt.Println("exists=false")
		}
	} else {
		fmt.Println("type assertion failed")
	}
}
