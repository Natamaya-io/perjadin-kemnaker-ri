package document

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"

	"github.com/nguyenthenguyen/docx"
)

type DocumentRequest struct {
	TemplateName string
	Variables    map[string]interface{}
}

type ImageData struct {
	Data     string // base64-encoded image data (may include data:image/...;base64, prefix)
	MimeType string // e.g. "image/jpeg", "image/png"
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
			Timeout: 30 * time.Second,
		},
	}
}

// removeEmptyTableRows removes entire <w:tr>...</w:tr> table rows that contain
// the __REMOVE_ROW__ marker text. This is used to hide unused petugas rows.
func removeEmptyTableRows(xmlContent string) string {
	marker := "__REMOVE_ROW__"
	for {
		idx := strings.Index(xmlContent, marker)
		if idx == -1 {
			break
		}
		// Find the enclosing <w:tr ...> before this marker
		trStart1 := strings.LastIndex(xmlContent[:idx], "<w:tr ")
		trStart2 := strings.LastIndex(xmlContent[:idx], "<w:tr>")
		
		trStart := trStart1
		if trStart2 > trStart1 {
			trStart = trStart2
		}

		if trStart == -1 {
			// Can't find table row, just remove the marker text
			xmlContent = strings.Replace(xmlContent, marker, "", 1)
			continue
		}
		// Find the closing </w:tr> after this marker
		trEnd := strings.Index(xmlContent[idx:], "</w:tr>")
		if trEnd == -1 {
			xmlContent = strings.Replace(xmlContent, marker, "", 1)
			continue
		}
		trEnd = idx + trEnd + len("</w:tr>")
		// Remove the entire row
		xmlContent = xmlContent[:trStart] + xmlContent[trEnd:]
	}
	return xmlContent
}

// Adjust petugas table width and documentation layout
func adjustPetugasTableWidths(xmlContent string) string {
	// Adjust the petugas table grid definition
	xmlContent = strings.Replace(xmlContent,
		`<w:tblGrid><w:gridCol w:w="1906"/><w:gridCol w:w="5391"/><w:gridCol w:w="3260"/></w:tblGrid>`,
		`<w:tblGrid><w:gridCol w:w="700"/><w:gridCol w:w="6597"/><w:gridCol w:w="3260"/></w:tblGrid>`,
		-1)
	// Adjust individual cell widths in petugas rows
	xmlContent = strings.ReplaceAll(xmlContent, `<w:tcW w:w="1906" w:type="dxa"/>`, `<w:tcW w:w="700" w:type="dxa"/>`)
	xmlContent = strings.ReplaceAll(xmlContent, `<w:tcW w:w="5391" w:type="dxa"/>`, `<w:tcW w:w="6597" w:type="dxa"/>`)

	// We look for any of the placeholders.
	fotoPlaceholders := []string{"satu", "dua", "tiga", "empat", "lima", "enam"}
	var fotoIdx int = -1

	for _, ordinal := range fotoPlaceholders {
		key := fmt.Sprintf("foto_dokumentasi_%s", ordinal)
		idx := strings.Index(xmlContent, key)
		if idx != -1 {
			fotoIdx = idx
			break
		}
	}

	if fotoIdx != -1 {
		// Find the enclosing <w:tbl>
		tblStart1 := strings.LastIndex(xmlContent[:fotoIdx], "<w:tbl>")
		tblStart2 := strings.LastIndex(xmlContent[:fotoIdx], "<w:tbl ")
		
		tblStart := tblStart1
		if tblStart2 > tblStart1 {
			tblStart = tblStart2
		}

		if tblStart != -1 {
			tblEnd := strings.Index(xmlContent[tblStart:], "</w:tbl>")
			if tblEnd != -1 {
				tblEnd = tblStart + tblEnd + len("</w:tbl>")
				// CRITICAL: Only proceed if the placeholder is actually INSIDE this table
				if fotoIdx < (tblStart + strings.Index(xmlContent[tblStart:], "</w:tbl>") + len("</w:tbl>")) {
					tableContent := xmlContent[tblStart:tblEnd]

					var replacement string
					foundCount := 0
					for _, ordinal := range fotoPlaceholders {
						key := fmt.Sprintf("foto_dokumentasi_%s", ordinal)
						if strings.Contains(tableContent, key) {
							if foundCount > 0 {
								replacement += `</w:p>`
							}
							replacement += `<w:p><w:pPr><w:jc w:val="center"/></w:pPr>`
							replacement += fmt.Sprintf(`<w:r><w:t>{{%s}}</w:t></w:r>`, key)
							foundCount++
						}
					}
					if foundCount > 0 {
						replacement += `</w:p>`
					}
					// Replace the entire table with our new grouped layout
					xmlContent = xmlContent[:tblStart] + replacement + xmlContent[tblEnd:]
				}
			}
		}
	}

	return xmlContent
}

// fixSplitRuns mencoba menyatukan placeholder {{...}} atau <<...>> yang terpecah oleh tag XML
func fixSplitRuns(xmlContent string) string {
	// Fix {{ and }} split by XML tags
	re := regexp.MustCompile(`(\{)(<[^>]+>)+(\{)`)
	xmlContent = re.ReplaceAllString(xmlContent, "$1$3")

	re = regexp.MustCompile(`(\})(<[^>]+>)+(\})`)
	xmlContent = re.ReplaceAllString(xmlContent, "$1$3")

	reInside := regexp.MustCompile(`(\{\{[^{}]*?)(<[^>]+>)+([^{}]*?\}\})`)
	for i := 0; i < 5; i++ {
		xmlContent = reInside.ReplaceAllString(xmlContent, "$1$3")
	}

	// Fix << and >> split by XML tags
	re2 := regexp.MustCompile(`(&lt;)(<[^>]+>)+(&lt;)`)
	xmlContent = re2.ReplaceAllString(xmlContent, "$1$3")

	re3 := regexp.MustCompile(`(&gt;)(<[^>]+>)+(&gt;)`)
	xmlContent = re3.ReplaceAllString(xmlContent, "$1$3")

	// Clean XML tags inside &lt;&lt; ... &gt;&gt;
	reAngle := regexp.MustCompile(`(&lt;&lt;[^&]*?)(<[^>]+>)+([^&]*?&gt;&gt;)`)
	for i := 0; i < 10; i++ {
		xmlContent = reAngle.ReplaceAllString(xmlContent, "$1$3")
	}

	// Also handle literal << >> (less common in DOCX but possible)
	reLit := regexp.MustCompile(`(<<[^<>]*?)(<[^>]+>)+([^<>]*?>>)`)
	for i := 0; i < 10; i++ {
		xmlContent = reLit.ReplaceAllString(xmlContent, "$1$3")
	}

	return xmlContent
}

// replaceAngleBrackets replaces <<key>> placeholders (stored as &lt;&lt;key&gt;&gt; in XML)
func replaceAngleBrackets(content string, key string, value string) string {
	// In DOCX XML, < and > are stored as &lt; and &gt;
	xmlPlaceholder := fmt.Sprintf("&lt;&lt;%s&gt;&gt;", key)
	content = strings.ReplaceAll(content, xmlPlaceholder, value)

	// Also handle literal << >> just in case
	literalPlaceholder := fmt.Sprintf("<<%s>>", key)
	content = strings.ReplaceAll(content, literalPlaceholder, value)

	return content
}

// Generate mengisi template DOCX dan mengonversinya ke PDF
func (g *Generator) Generate(ctx context.Context, req DocumentRequest) ([]byte, error) {
	templatePath := filepath.Join(g.templateDir, req.TemplateName)

	doc, err := docx.ReadDocxFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open template: %w", err)
	}
	defer doc.Close()

	d := doc.Editable()

	// Fix split runs for both placeholder formats
	d.SetContent(fixSplitRuns(d.GetContent()))

	// Replace variables in both {{key}} and <<key>> formats
	for k, v := range req.Variables {
		val := fmt.Sprintf("%v", v)

		// {{key}} format
		placeholder := fmt.Sprintf("{{%s}}", k)
		d.Replace(placeholder, val, -1)
		d.ReplaceFooter(placeholder, val)
		d.ReplaceHeader(placeholder, val)

		// <<key>> format - need to handle in raw XML since docx library only does {{}}
		content := d.GetContent()
		content = replaceAngleBrackets(content, k, val)
		d.SetContent(content)
	}

	// Remove table rows marked for deletion (empty petugas slots)
	d.SetContent(removeEmptyTableRows(d.GetContent()))

	// Adjust petugas table column widths (narrow number col, widen name col)
	d.SetContent(adjustPetugasTableWidths(d.GetContent()))

	var docxBuf bytes.Buffer
	if err := d.Write(&docxBuf); err != nil {
		return nil, fmt.Errorf("failed to write docx: %w", err)
	}

	return g.convertToPDF(ctx, docxBuf.Bytes())
}

// GenerateDocx generates a filled DOCX document without converting to PDF
func (g *Generator) GenerateDocx(ctx context.Context, req DocumentRequest) ([]byte, error) {
	templatePath := filepath.Join(g.templateDir, req.TemplateName)

	doc, err := docx.ReadDocxFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open template: %w", err)
	}
	defer doc.Close()

	d := doc.Editable()

	// Fix split runs for both placeholder formats
	d.SetContent(fixSplitRuns(d.GetContent()))

	// Replace variables in both {{key}} and <<key>> formats
	for k, v := range req.Variables {
		val := fmt.Sprintf("%v", v)

		// {{key}} format
		placeholder := fmt.Sprintf("{{%s}}", k)
		d.Replace(placeholder, val, -1)
		d.ReplaceFooter(placeholder, val)
		d.ReplaceHeader(placeholder, val)

		// <<key>> format - need to handle in raw XML since docx library only does {{}}
		content := d.GetContent()
		content = replaceAngleBrackets(content, k, val)
		d.SetContent(content)
	}

	// Remove table rows marked for deletion (empty petugas slots)
	d.SetContent(removeEmptyTableRows(d.GetContent()))

	// Adjust petugas table column widths (narrow number col, widen name col)
	d.SetContent(adjustPetugasTableWidths(d.GetContent()))

	var docxBuf bytes.Buffer
	if err := d.Write(&docxBuf); err != nil {
		return nil, fmt.Errorf("failed to write docx: %w", err)
	}

	return docxBuf.Bytes(), nil
}

// GenerateWithImages generates a document with text replacement AND embedded images
func (g *Generator) GenerateWithImages(ctx context.Context, req DocumentRequest, images map[string]ImageData) ([]byte, error) {
	templatePath := filepath.Join(g.templateDir, req.TemplateName)

	doc, err := docx.ReadDocxFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open template: %w", err)
	}
	defer doc.Close()

	d := doc.Editable()

	// Fix split runs
	d.SetContent(fixSplitRuns(d.GetContent()))

	// Replace text variables
	for k, v := range req.Variables {
		val := fmt.Sprintf("%v", v)

		placeholder := fmt.Sprintf("{{%s}}", k)
		d.Replace(placeholder, val, -1)
		d.ReplaceFooter(placeholder, val)
		d.ReplaceHeader(placeholder, val)

		content := d.GetContent()
		content = replaceAngleBrackets(content, k, val)
		d.SetContent(content)
	}

	// Remove table rows marked for deletion (empty petugas slots)
	d.SetContent(removeEmptyTableRows(d.GetContent()))

	// Adjust petugas table column widths (narrow number col, widen name col)
	d.SetContent(adjustPetugasTableWidths(d.GetContent()))

	var docxBuf bytes.Buffer
	if err := d.Write(&docxBuf); err != nil {
		return nil, fmt.Errorf("failed to write docx: %w", err)
	}

	// Now inject images into the DOCX
	docxBytes := docxBuf.Bytes()
	if len(images) > 0 {
		docxBytes, err = injectImages(docxBytes, images)
		if err != nil {
			return nil, fmt.Errorf("failed to inject images: %w", err)
		}
	}

	return g.convertToPDF(ctx, docxBytes)
}

// GenerateDocxWithImages generates a document with text replacement AND embedded images without converting to PDF
func (g *Generator) GenerateDocxWithImages(ctx context.Context, req DocumentRequest, images map[string]ImageData) ([]byte, error) {
	templatePath := filepath.Join(g.templateDir, req.TemplateName)

	doc, err := docx.ReadDocxFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open template: %w", err)
	}
	defer doc.Close()

	d := doc.Editable()

	// Fix split runs
	d.SetContent(fixSplitRuns(d.GetContent()))

	// Replace text variables
	for k, v := range req.Variables {
		val := fmt.Sprintf("%v", v)

		placeholder := fmt.Sprintf("{{%s}}", k)
		d.Replace(placeholder, val, -1)
		d.ReplaceFooter(placeholder, val)
		d.ReplaceHeader(placeholder, val)

		content := d.GetContent()
		content = replaceAngleBrackets(content, k, val)
		d.SetContent(content)
	}

	// Remove table rows marked for deletion (empty petugas slots)
	d.SetContent(removeEmptyTableRows(d.GetContent()))

	// Adjust petugas table column widths (narrow number col, widen name col)
	d.SetContent(adjustPetugasTableWidths(d.GetContent()))

	var docxBuf bytes.Buffer
	if err := d.Write(&docxBuf); err != nil {
		return nil, fmt.Errorf("failed to write docx: %w", err)
	}

	// Now inject images into the DOCX
	docxBytes := docxBuf.Bytes()
	if len(images) > 0 {
		var err error
		docxBytes, err = injectImages(docxBytes, images)
		if err != nil {
			return nil, fmt.Errorf("failed to inject images: %w", err)
		}
	}

	return docxBytes, nil
}

// injectImages replaces image placeholders in the DOCX (zip) with actual embedded images
func injectImages(docxBytes []byte, images map[string]ImageData) ([]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(docxBytes), int64(len(docxBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to read docx as zip: %w", err)
	}

	var outBuf bytes.Buffer
	writer := zip.NewWriter(&outBuf)

	// Track what we need to add
	imageFiles := make(map[string][]byte) // filename -> decoded image bytes
	rIdCounter := 100                     // Start high to avoid conflicts

	// Build image info
	type imageInfo struct {
		rId      string
		filename string
		ext      string
		cx       int64
		cy       int64
	}
	imageInfoMap := make(map[string]imageInfo) // placeholder key -> info

	for key, img := range images {
		imgBytes, err := decodeBase64Image(img.Data)
		if err != nil {
			fmt.Printf("Warning: failed to decode image for %s: %v\n", key, err)
			continue
		}

		ext := getImageExtension(img.MimeType)
		rId := fmt.Sprintf("rId%d", rIdCounter)
		filename := fmt.Sprintf("image_doc_%d%s", rIdCounter, ext)
		rIdCounter++

		// Detect image dimensions for proper aspect ratio
		var cx, cy int64
		cx = 5400000 // 15cm default width
		cy = 4000000 // 11cm default height
		if cfg, _, err := image.DecodeConfig(bytes.NewReader(imgBytes)); err == nil && cfg.Width > 0 && cfg.Height > 0 {
			w := float64(cfg.Width)
			h := float64(cfg.Height)

			// Handle EXIF orientation
			if x, err := exif.Decode(bytes.NewReader(imgBytes)); err == nil {
				if tag, err := x.Get(exif.Orientation); err == nil {
					if orientation, err := tag.Int(0); err == nil {
						if orientation >= 5 && orientation <= 8 {
							w, h = h, w // Swap width and height for rotated images
						}
					}
				}
			}

			// Bounding box for 2x2 layout (each image is ~half page width)
			const maxWidth float64 = 2800000  // ~7.7cm width
			const maxHeight float64 = 4000000 // ~11.1cm height

			ratioW := maxWidth / w
			ratioH := maxHeight / h

			// Use the smaller ratio to ensure both fit the box without cropping
			ratio := ratioW
			if ratioH < ratioW {
				ratio = ratioH
			}

			cx = int64(w * ratio)
			cy = int64(h * ratio)
		}

		imageFiles[filename] = imgBytes
		imageInfoMap[key] = imageInfo{
			rId:      rId,
			filename: filename,
			ext:      ext,
			cx:       cx,
			cy:       cy,
		}
	}

	// Process each file in the zip
	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			return nil, err
		}

		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, err
		}

		if file.Name == "word/document.xml" {
			contentStr := string(content)
			// Fix split runs AGAIN on the written output (docx library may re-split)
			contentStr = fixSplitRuns(contentStr)
			// Replace image placeholders with inline drawing XML
			// We iterate through all placeholders and replace them one by one
			for key, info := range imageInfoMap {
				contentStr = replaceImagePlaceholder(contentStr, key, info.rId, info.cx, info.cy)
			}
			content = []byte(contentStr)
		}

		if file.Name == "word/_rels/document.xml.rels" {
			contentStr := string(content)
			// Add relationship entries for images before closing </Relationships>
			var relEntries []string
			for _, info := range imageInfoMap {
				relEntry := fmt.Sprintf(
					`<Relationship Id="%s" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="media/%s"/>`,
					info.rId, info.filename)
				relEntries = append(relEntries, relEntry)
			}
			if len(relEntries) > 0 {
				contentStr = strings.Replace(contentStr, "</Relationships>",
					strings.Join(relEntries, "\n")+"\n</Relationships>", 1)
			}
			content = []byte(contentStr)
		}

		// Add content types for images
		if file.Name == "[Content_Types].xml" {
			contentStr := string(content)
			// Ensure image content types are registered
			if !strings.Contains(contentStr, `Extension="jpeg"`) && !strings.Contains(contentStr, `Extension="jpg"`) {
				contentStr = strings.Replace(contentStr, "</Types>",
					`<Default Extension="jpeg" ContentType="image/jpeg"/>`+"\n</Types>", 1)
			}
			if !strings.Contains(contentStr, `Extension="png"`) {
				contentStr = strings.Replace(contentStr, "</Types>",
					`<Default Extension="png" ContentType="image/png"/>`+"\n</Types>", 1)
			}
			content = []byte(contentStr)
		}

		// Write the (possibly modified) file to the output zip
		w, err := writer.CreateHeader(&zip.FileHeader{
			Name:   file.Name,
			Method: file.Method,
		})
		if err != nil {
			return nil, err
		}
		if _, err := w.Write(content); err != nil {
			return nil, err
		}
	}

	// Add image files to word/media/
	for filename, imgBytes := range imageFiles {
		w, err := writer.Create("word/media/" + filename)
		if err != nil {
			return nil, err
		}
		if _, err := w.Write(imgBytes); err != nil {
			return nil, err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return outBuf.Bytes(), nil
}

// replaceImagePlaceholder replaces the run containing {{PLACEHOLDER}} with an image run
func replaceImagePlaceholder(xmlContent string, placeholderKey string, rId string, cx int64, cy int64) string {
	// The image run
	imgRun := fmt.Sprintf(
		`<w:r><w:drawing>`+
			`<wp:inline distT="0" distB="0" distL="0" distR="0">`+
			`<wp:extent cx="%d" cy="%d"/>`+
			`<wp:effectExtent l="0" t="0" r="0" b="0"/>`+
			`<wp:docPr id="%d" name="img_%s"/>`+
			`<wp:cNvGraphicFramePr>`+
			`<a:graphicFrameLocks xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" noChangeAspect="1"/>`+
			`</wp:cNvGraphicFramePr>`+
			`<a:graphic xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main">`+
			`<a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/picture">`+
			`<pic:pic xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture">`+
			`<pic:nvPicPr><pic:cNvPr id="0" name="img_%s"/><pic:cNvPicPr/></pic:nvPicPr>`+
			`<pic:blipFill>`+
			`<a:blip r:embed="%s" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"/>`+
			`<a:stretch><a:fillRect/></a:stretch>`+
			`</pic:blipFill>`+
			`<pic:spPr bwMode="auto">`+
			`<a:xfrm><a:off x="0" y="0"/><a:ext cx="%d" cy="%d"/></a:xfrm>`+
			`<a:prstGeom prst="rect"><a:avLst/></a:prstGeom>`+
			`<a:noFill/>`+
			`</pic:spPr>`+
			`</pic:pic></a:graphicData></a:graphic>`+
			`</wp:inline></w:drawing><w:br/></w:r>`,
		cx, cy,
		hash(placeholderKey), placeholderKey,
		placeholderKey,
		rId,
		cx, cy)

	// Try {{key}} format
	curlyPlaceholder := fmt.Sprintf("{{%s}}", placeholderKey)
	idx := strings.Index(xmlContent, curlyPlaceholder)
	if idx == -1 {
		// Try <<key>> format (encoded as &lt;&lt;key&gt;&gt;)
		curlyPlaceholder = fmt.Sprintf("&lt;&lt;%s&gt;&gt;", placeholderKey)
		idx = strings.Index(xmlContent, curlyPlaceholder)
	}

	if idx == -1 {
		return xmlContent
	}

	// Find the enclosing <w:r ...> ... </w:r> that contains this placeholder
	// We want to replace the whole run so we don't leave broken <w:t> tags
	rStart := strings.LastIndex(xmlContent[:idx], "<w:r")
	if rStart != -1 {
		rEnd := strings.Index(xmlContent[idx:], "</w:r>")
		if rEnd != -1 {
			rEnd = idx + rEnd + len("</w:r>")
			// Replace the entire run with the image run
			return xmlContent[:rStart] + imgRun + xmlContent[rEnd:]
		}
	}

	// Fallback: just replace the text (might lead to invalid XML if inside <w:t>)
	return strings.Replace(xmlContent, curlyPlaceholder, imgRun, 1)
}

func hash(s string) int {
	h := 0
	for _, c := range s {
		h = 31*h + int(c)
	}
	if h < 0 {
		h = -h
	}
	return h % 100000
}

func decodeBase64Image(data string) ([]byte, error) {
	// Strip data URI prefix if present (e.g., "data:image/jpeg;base64,")
	if idx := strings.Index(data, ";base64,"); idx != -1 {
		data = data[idx+8:]
	} else if idx := strings.Index(data, ","); idx != -1 && strings.HasPrefix(data, "data:") {
		data = data[idx+1:]
	}

	return base64.StdEncoding.DecodeString(data)
}

func getImageExtension(mimeType string) string {
	switch mimeType {
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ".jpeg"
	}
}

func (g *Generator) convertToPDF(ctx context.Context, docxBytes []byte) ([]byte, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

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
