package files

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/xuri/excelize/v2"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

const (
	mimePDF   = "application/pdf"
	mimeDOCX  = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	mimeXLSX  = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	mimePlain = "text/plain"
	mimeCSV   = "text/csv"
	mimeMD    = "text/markdown"
)

var extractable = map[string]bool{
	mimeDOCX: true, mimeXLSX: true, mimePlain: true, mimeCSV: true, mimeMD: true,
}

// Extractable reports whether this gateway can read the format itself. PDFs
// and images need the OCR service, which is deployed separately.
func Extractable(mime string) bool { return extractable[mime] }

func extract(mime string, body []byte) (string, error) {
	switch mime {
	case mimeDOCX:
		return extractDOCX(body)
	case mimeXLSX:
		return extractXLSX(body)
	case mimePlain, mimeCSV, mimeMD:
		return decodeText(body), nil
	}
	return "", fmt.Errorf("unsupported mime %q", mime)
}

// decodeText handles Windows-1256, which Iranian text and CSV files use far
// more often than UTF-8. Assuming UTF-8 turns them into mojibake.
func decodeText(body []byte) string {
	if isProbablyUTF8(body) {
		return string(body)
	}

	decoded, _, err := transform.Bytes(charmap.Windows1256.NewDecoder(), body)
	if err != nil {
		return string(body)
	}
	return string(decoded)
}

func isProbablyUTF8(body []byte) bool {
	return utf8.Valid(body)
}

func extractDOCX(body []byte) (string, error) {
	reader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return "", err
	}

	for _, file := range reader.File {
		if file.Name != "word/document.xml" {
			continue
		}
		content, err := file.Open()
		if err != nil {
			return "", err
		}
		defer func() { _ = content.Close() }()
		return textFromOOXML(content)
	}
	return "", fmt.Errorf("no document.xml in docx")
}

// textFromOOXML walks the XML rather than parsing the full schema: every
// paragraph is a w:p and every run of text is a w:t, which is all that matters
// for feeding a model.
func textFromOOXML(r io.Reader) (string, error) {
	decoder := xml.NewDecoder(r)

	var out strings.Builder
	inText := false

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		switch element := token.(type) {
		case xml.StartElement:
			if element.Name.Local == "t" {
				inText = true
			}
		case xml.EndElement:
			if element.Name.Local == "t" {
				inText = false
			}
			if element.Name.Local == "p" {
				out.WriteString("\n")
			}
		case xml.CharData:
			if inText {
				out.Write(element)
			}
		}
	}

	return strings.TrimSpace(out.String()), nil
}

func extractXLSX(body []byte) (string, error) {
	workbook, err := excelize.OpenReader(bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer func() { _ = workbook.Close() }()

	var out strings.Builder
	for _, sheet := range workbook.GetSheetList() {
		rows, err := workbook.GetRows(sheet)
		if err != nil {
			return "", err
		}

		out.WriteString("# " + sheet + "\n")
		for _, row := range rows {
			out.WriteString(strings.Join(row, "\t"))
			out.WriteString("\n")
		}
	}
	return strings.TrimSpace(out.String()), nil
}
