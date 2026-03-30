package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"strings"
)

func main() {
	r, err := zip.OpenReader("templates/Berkas Luar Kota - Laporan.docx")
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				log.Fatal(err)
			}
			b, _ := io.ReadAll(rc)
			content := string(b)

			// Find the table grid containing petugas (look for no_urut_satu)
			idx := strings.Index(content, "no_urut_satu")
			if idx == -1 {
				fmt.Println("not found")
				return
			}
			// Find the <w:tbl that contains this
			tblStart := strings.LastIndex(content[:idx], "<w:tbl>")
			if tblStart == -1 {
				tblStart = strings.LastIndex(content[:idx], "<w:tbl ")
			}
			// Find the tblGrid
			gridStart := strings.Index(content[tblStart:], "<w:tblGrid>")
			if gridStart != -1 {
				gridEnd := strings.Index(content[tblStart+gridStart:], "</w:tblGrid>")
				if gridEnd != -1 {
					fmt.Println("=== TABLE GRID ===")
					fmt.Println(content[tblStart+gridStart : tblStart+gridStart+gridEnd+len("</w:tblGrid>")])
				}
			}
			// Show first row structure (header or first petugas)
			trStart := strings.Index(content[tblStart:], "<w:tr ")
			if trStart != -1 {
				trEnd := strings.Index(content[tblStart+trStart:], "</w:tr>")
				if trEnd != -1 {
					row := content[tblStart+trStart : tblStart+trStart+trEnd+len("</w:tr>")]
					fmt.Println("\n=== FIRST ROW tcW values ===")
					parts := strings.Split(row, "<w:tcW")
					for i, p := range parts {
						if i == 0 { continue }
						end := strings.Index(p, "/>")
						if end != -1 {
							fmt.Printf("Cell %d: <w:tcW%s/>\n", i, p[:end])
						}
					}
				}
			}
		}
	}
}
