package document

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"regexp"
	"time"

	"github.com/nguyenthenguyen/docx"
)

type DocumentRequest struct {
	TemplateName string
	Variables    map[string]interface{}
}

type Generator struct {
	gotenbergURL string
	templateDir  string
	httpClient   *http.Client
}

func NewGenerator(gotenbergURL string, templateDir string) *Generator {
	return &Generator{
		gotenbergURL: gotenbergURL,
		templateDir:  templateDir,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

// fixSplitRuns mencoba menyatukan placeholder {{...}} yang terpecah oleh tag XML (seperti <w:t>)
func fixSplitRuns(xmlContent string) string {
	// Menghapus semua tag XML di antara { dan }
	// Ini adalah trik umum untuk DOCX karena Microsoft Word sering memecah string di dalam variabel.
	re := regexp.MustCompile(`(\{)(<[^>]+>)+(\{)`)
	xmlContent = re.ReplaceAllString(xmlContent, "$1$3")

	re = regexp.MustCompile(`(\})(<[^>]+>)+(\})`)
	xmlContent = re.ReplaceAllString(xmlContent, "$1$3")

	// Membersihkan tag XML di dalam kurung kurawal ganda {{ ... }}
	// Membutuhkan regex yang lebih kompleks, tapi pendekatan sederhana:
	reInside := regexp.MustCompile(`(\{\{[^{}]*?)(<[^>]+>)+([^{}]*?\}\})`)
	// Ulangi beberapa kali karena bisa ada banyak tag bertumpuk di dalam satu variabel
	for i := 0; i < 5; i++ {
		xmlContent = reInside.ReplaceAllString(xmlContent, "$1$3")
	}

	return xmlContent
}

// Generate mengisi template DOCX dan mengonversinya ke PDF
func (g *Generator) Generate(ctx context.Context, req DocumentRequest) ([]byte, error) {
	templatePath := filepath.Join(g.templateDir, req.TemplateName)

	// 1. Load & Fill DOCX
	doc, err := docx.ReadDocxFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open template: %w", err)
	}
	defer doc.Close()

	d := doc.Editable()

	// Coba satukan placeholder yang terpecah oleh Word formatting
	d.SetContent(fixSplitRuns(d.GetContent()))

	// Replace variables in document body, headers, and footers
	for k, v := range req.Variables {
		placeholder := fmt.Sprintf("{{%s}}", k)
		val := fmt.Sprintf("%v", v)
		d.Replace(placeholder, val, -1)
		d.ReplaceFooter(placeholder, val)
		d.ReplaceHeader(placeholder, val)
	}

	var docxBuf bytes.Buffer
	if err := d.Write(&docxBuf); err != nil {
		return nil, fmt.Errorf("failed to write docx: %w", err)
	}

	// 2. Convert to PDF via Gotenberg
	return g.convertToPDF(ctx, docxBuf.Bytes())
}

func (g *Generator) convertToPDF(ctx context.Context, docxBytes []byte) ([]byte, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Ensure native page size is used (e.g., for F4/Legal/Folio size templates)
	// This prevents the bottom of the document from being cut off in Gotenberg.
	if err := writer.WriteField("nativePageSize", "true"); err != nil {
		return nil, err
	}

	part, err := writer.CreateFormFile("files", "document.docx")
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, bytes.NewReader(docxBytes)); err != nil {
		return nil, err
	}
	writer.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/forms/libreoffice/convert", g.gotenbergURL), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gotenberg request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gotenberg error (%d): %s", resp.StatusCode, string(respBody))
	}

	return io.ReadAll(resp.Body)
}
