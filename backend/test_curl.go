package main
import (
	"fmt"
	"io/ioutil"
	"net/http"
)
func main() {
	resp, _ := http.Get("http://localhost:8081/api/v1/auth/demo-users")
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(b))
}
