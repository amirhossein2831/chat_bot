package pdf

import (
	"bytes"
	"fmt"
	"image/png"
	"sync"
	"time"

	"github.com/klippa-app/go-pdfium"
	"github.com/klippa-app/go-pdfium/references"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/webassembly"
)

// Reader wraps pdfium. It is the WebAssembly build rather than the cgo one, so
// the gateway stays a single static binary with nothing to install alongside
// it, and pdfium rather than MuPDF because MuPDF is AGPL and this is a
// commercial network service.
const instanceTimeout = 30 * time.Second

type Reader struct {
	pool pdfium.Pool

	// pdfium instances are not safe for concurrent use.
	mu sync.Mutex
}

func New() (*Reader, error) {
	pool, err := webassembly.Init(webassembly.Config{MinIdle: 1, MaxIdle: 1, MaxTotal: 1})
	if err != nil {
		return nil, fmt.Errorf("pdfium init: %w", err)
	}
	return &Reader{pool: pool}, nil
}

func (r *Reader) Close() error { return r.pool.Close() }

// Pages returns the text pdfium can read from each page. A page with no text
// layer comes back empty rather than as an error: whether that means the
// document needs OCR is the caller's decision.
func (r *Reader) Pages(document []byte) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	instance, doc, err := r.open(document)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = instance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: doc})
		_ = instance.Close()
	}()

	count, err := instance.FPDF_GetPageCount(&requests.FPDF_GetPageCount{Document: doc})
	if err != nil {
		return nil, fmt.Errorf("page count: %w", err)
	}

	pages := make([]string, 0, count.PageCount)
	for index := range count.PageCount {
		text, err := instance.GetPageText(&requests.GetPageText{
			Page: requests.Page{ByIndex: &requests.PageByIndex{Document: doc, Index: index}},
		})
		if err != nil {
			return nil, fmt.Errorf("page %d text: %w", index, err)
		}
		pages = append(pages, text.Text)
	}
	return pages, nil
}

// Render rasterises every page as PNG, which is what the OCR service reads.
func (r *Reader) Render(document []byte, dpi int, limit int) ([][]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	instance, doc, err := r.open(document)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = instance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: doc})
		_ = instance.Close()
	}()

	count, err := instance.FPDF_GetPageCount(&requests.FPDF_GetPageCount{Document: doc})
	if err != nil {
		return nil, fmt.Errorf("page count: %w", err)
	}

	total := count.PageCount
	if limit > 0 && total > limit {
		total = limit
	}

	rendered := make([][]byte, 0, total)
	for index := range total {
		page, err := instance.RenderPageInDPI(&requests.RenderPageInDPI{
			Page: requests.Page{ByIndex: &requests.PageByIndex{Document: doc, Index: index}},
			DPI:  dpi,
		})
		if err != nil {
			return nil, fmt.Errorf("render page %d: %w", index, err)
		}

		var buffer bytes.Buffer
		if err := png.Encode(&buffer, page.Result.Image); err != nil {
			page.Cleanup()
			return nil, fmt.Errorf("encode page %d: %w", index, err)
		}
		page.Cleanup()

		rendered = append(rendered, buffer.Bytes())
	}
	return rendered, nil
}

func (r *Reader) open(document []byte) (pdfium.Pdfium, references.FPDF_DOCUMENT, error) {
	instance, err := r.pool.GetInstance(instanceTimeout)
	if err != nil {
		return nil, "", fmt.Errorf("pdfium instance: %w", err)
	}

	opened, err := instance.FPDF_LoadMemDocument(&requests.FPDF_LoadMemDocument{Data: &document})
	if err != nil {
		_ = instance.Close()
		return nil, "", fmt.Errorf("load document: %w", err)
	}
	return instance, opened.Document, nil
}
