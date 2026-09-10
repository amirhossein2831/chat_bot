package files

import (
	"context"
	"fmt"
	"strings"

	"github.com/amirhossein2831/chat_bot/gateway/internal/constants"
)

const (
	// The DPI the OCR service was tuned against. Lower loses Persian diacritics;
	// higher costs render time and bytes for no accuracy.
	ocrDPI = 250

	ocrPageLimit = 200
)

var ocrLanguages = []string{"fa", "en"}

// read turns an uploaded file into text.
//
// A PDF takes one of two routes, and choosing between them is the whole point:
// almost every PDF carries a text layer that is read in milliseconds, while a
// scan has to be rendered and recognised, which is slow and needs a GPU. The
// trap is the third case, a text layer that extracts successfully and means
// nothing, which UsableTextLayer exists to catch.
func (s *Service) read(ctx context.Context, mime string, body []byte) (string, error) {
	switch {
	case mime == mimePDF:
		return s.readPDF(ctx, body)
	case strings.HasPrefix(mime, "image/"):
		return s.recognize(ctx, [][]byte{body})
	case Extractable(mime):
		return extract(mime, body)
	}
	return "", fmt.Errorf("%w: %s cannot be read", constants.ErrUnsupportedMedia, mime)
}

func (s *Service) readPDF(ctx context.Context, body []byte) (string, error) {
	if s.pages == nil {
		return "", fmt.Errorf("%w: pdf reading is not configured", constants.ErrUnsupportedMedia)
	}

	pages, err := s.pages.Pages(body)
	if err != nil {
		return "", err
	}

	if UsableTextLayer(pages) {
		return strings.Join(pages, "\n\n"), nil
	}

	rendered, err := s.pages.Render(body, ocrDPI, ocrPageLimit)
	if err != nil {
		return "", err
	}
	return s.recognize(ctx, rendered)
}

func (s *Service) recognize(ctx context.Context, pages [][]byte) (string, error) {
	if s.recognizer == nil {
		return "", fmt.Errorf("%w: this file needs OCR, which is not deployed",
			constants.ErrUnsupportedMedia)
	}
	return s.recognizer.Recognize(ctx, pages, ocrLanguages)
}
