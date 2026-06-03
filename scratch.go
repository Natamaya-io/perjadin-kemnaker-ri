package main
import (
	"fmt"
	"os"
	"strings"
)

func main() {
	content, _ := os.ReadFile("backend/internal/domain/record/repository.go")
	lines := strings.Split(string(content), "\n")
	
	openBraces := 0
	inFunc := false
	for i, line := range lines {
		if i == 1213 { // Line 1214
			inFunc = true
			fmt.Printf("Started parsing GetPaginatedRecords at line %d\n", i+1)
		}
		if inFunc {
			for _, char := range line {
				if char == '{' {
					openBraces++
				} else if char == '}' {
					openBraces--
				}
			}
			if openBraces == 0 && i >= 1213 {
				fmt.Printf("GetPaginatedRecords closed at line %d\n", i+1)
				inFunc = false
				break
			}
		}
	}
	fmt.Printf("Remaining open braces: %d\n", openBraces)
}
