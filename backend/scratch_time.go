package main
import (
	"encoding/json"
	"fmt"
	"time"
)
type T struct {
	D time.Time `json:"d"`
}
func main() {
	var t T
	err := json.Unmarshal([]byte(`{"d":"2026-04-24"}`), &t)
	fmt.Println(err)
	fmt.Println(t.D.Format(time.RFC3339))
}
