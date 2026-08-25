// Package router implements the task routing layer from API.md §15: it decides
// which parser handles a file and which model tier handles a request.
package router

import (
	"path/filepath"
	"strings"

	"lingxi-claw/internal/model"
)

// SupportedExts are the upload formats listed in API.md §8.2.
var SupportedExts = map[string]bool{
	".pdf":  true,
	".docx": true,
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".webp": true,
	".bmp":  true,
}

// FileDecision is the outcome of routing one uploaded file.
type FileDecision struct {
	Ext       string
	FileRoute string // model.FileRoute* constant
	IsImage   bool
	Supported bool
}

// RouteFile picks a parser for a file based on its extension and, for PDFs,
// whether the bytes contain an extractable text layer. Scanned PDFs and images
// go to OCR; everything else goes to a text parser.
func RouteFile(name string, content []byte) FileDecision {
	ext := strings.ToLower(filepath.Ext(name))
	d := FileDecision{Ext: ext, Supported: SupportedExts[ext]}
	if !d.Supported {
		return d
	}

	switch ext {
	case ".docx":
		d.FileRoute = model.FileRouteDocxParser
	case ".pdf":
		if HasTextLayer(content) {
			d.FileRoute = model.FileRouteTextParser
		} else {
			d.FileRoute = model.FileRouteOCR
		}
	default: // images
		d.IsImage = true
		d.FileRoute = model.FileRouteOCR
	}
	return d
}

// HasTextLayer reports whether a PDF carries extractable text. It looks for the
// text-showing operators inside content streams; a scan-only PDF has image
// XObjects but no such operators.
func HasTextLayer(content []byte) bool {
	// A PDF small enough to hold no page content is treated as scanned so that
	// it is routed through OCR rather than silently yielding empty text.
	if len(content) == 0 {
		return false
	}
	markers := [][]byte{[]byte("/Font"), []byte("BT\n"), []byte("BT\r"), []byte("BT "), []byte(" Tj"), []byte(" TJ")}
	for _, m := range markers {
		if containsBytes(content, m) {
			return true
		}
	}
	return false
}

func containsBytes(haystack, needle []byte) bool {
	if len(needle) == 0 || len(haystack) < len(needle) {
		return false
	}
	limit := len(haystack) - len(needle)
	for i := 0; i <= limit; i++ {
		if haystack[i] == needle[0] && string(haystack[i:i+len(needle)]) == string(needle) {
			return true
		}
	}
	return false
}
