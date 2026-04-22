package main
import (
	"bytes"
	"fmt"
	"net/http"
	"github.com/labstack/echo/v4"
)
type TravelReport struct {
	Text string `json:"text"`
}
type TravelRecord struct {
	Report *TravelReport `json:"reportData,omitempty"`
}
func main() {
	e := echo.New()
	req, _ := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(`{"reportData": {"text": "hello"}}`)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := e.AcquireContext()
	rec.Reset(req, nil)
	
	r := &TravelRecord{}
	rec.Bind(r)
	fmt.Printf("r.Report != nil ? %v\n", r.Report != nil)
	if r.Report != nil {
		fmt.Printf("text: %s\n", r.Report.Text)
	}
}
