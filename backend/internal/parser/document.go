package parser

import (
	"bytes"
	"strings"

	"github.com/dslipak/pdf"
)

// ExtractDocumentText extracts plain text from a raw document byte slice.
// Supports: "application/pdf", "text/plain", "text/csv".
func ExtractDocumentText(data []byte, mimeType string) (string, error) {
	if strings.HasPrefix(mimeType, "text/") {
		// Treat as plain text (txt, csv)
		return string(data), nil
	}

	if mimeType == "application/pdf" {
		reader := bytes.NewReader(data)
		pdfReader, err := pdf.NewReader(reader, int64(len(data)))
		if err != nil {
			return "", err
		}

		var textBuilder strings.Builder
		numPages := pdfReader.NumPage()
		if numPages > 10 {
			numPages = 10 // Prevent OOM by capping massive PDFs
		}
		for i := 1; i <= numPages; i++ {
			p := pdfReader.Page(i)
			if p.V.IsNull() {
				continue
			}
			text, err := p.GetPlainText(nil)
			if err != nil {
				// skip unreadable pages or return error
				continue
			}
			textBuilder.WriteString(text)
			textBuilder.WriteString("\n")
		}
		return textBuilder.String(), nil
	}

	return "", nil // Unsupported mime type
}
