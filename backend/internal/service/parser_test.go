package service_test

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"

	"lingxi-claw/internal/model"
	"lingxi-claw/internal/service"
)

// buildDOCX builds a minimal valid .docx containing the given paragraphs.
func buildDOCX(t *testing.T, paragraphs ...string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	var body strings.Builder
	body.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	body.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>`)
	for _, p := range paragraphs {
		body.WriteString(`<w:p><w:r><w:t>` + p + `</w:t></w:r></w:p>`)
	}
	body.WriteString(`</w:body></w:document>`)

	w, err := zw.Create("word/document.xml")
	if err != nil {
		t.Fatalf("create document.xml: %v", err)
	}
	if _, err := w.Write([]byte(body.String())); err != nil {
		t.Fatalf("write document.xml: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}

// buildTextPDF builds a PDF with an uncompressed text layer.
func buildTextPDF(lines ...string) []byte {
	var sb strings.Builder
	sb.WriteString("%PDF-1.4\n")
	sb.WriteString("1 0 obj<</Type/Font/Subtype/Type1/BaseFont/Helvetica>>endobj\n")
	sb.WriteString("2 0 obj<</Length 100>>\nstream\nBT /F1 12 Tf\n")
	for _, l := range lines {
		sb.WriteString("(" + l + ") Tj\n")
	}
	sb.WriteString("ET\nendstream\nendobj\n%%EOF\n")
	return []byte(sb.String())
}

func TestParseDOCX(t *testing.T) {
	content := buildDOCX(t,
		"1. 计算二重积分 ∬_D x dxdy。",
		"2. 证明连续函数在闭区间上有界。",
	)

	text, err := service.ParseDOCX(content)
	if err != nil {
		t.Fatalf("ParseDOCX: %v", err)
	}
	if !strings.Contains(text, "二重积分") {
		t.Errorf("text missing first paragraph: %q", text)
	}
	if !strings.Contains(text, "证明连续函数") {
		t.Errorf("text missing second paragraph: %q", text)
	}
	// Paragraphs must land on separate lines so question splitting works.
	if !strings.Contains(text, "\n") {
		t.Errorf("paragraphs were not separated by newlines: %q", text)
	}
}

func TestParseDOCXErrors(t *testing.T) {
	t.Run("not a zip", func(t *testing.T) {
		if _, err := service.ParseDOCX([]byte("this is not a docx")); err == nil {
			t.Fatal("want error for non-zip content")
		}
	})

	t.Run("zip without document.xml", func(t *testing.T) {
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)
		w, _ := zw.Create("other.xml")
		_, _ = w.Write([]byte("<x/>"))
		_ = zw.Close()

		if _, err := service.ParseDOCX(buf.Bytes()); err == nil {
			t.Fatal("want error when word/document.xml is missing")
		}
	})
}

func TestParsePDFText(t *testing.T) {
	content := buildTextPDF("1. 求解微分方程 y' + 2y = 0。", "2. 判断级数收敛性。")

	text, err := service.ParsePDFText(content)
	if err != nil {
		t.Fatalf("ParsePDFText: %v", err)
	}
	if !strings.Contains(text, "微分方程") {
		t.Errorf("text missing first line: %q", text)
	}
	if !strings.Contains(text, "级数") {
		t.Errorf("text missing second line: %q", text)
	}
}

func TestParsePDFTextErrors(t *testing.T) {
	t.Run("not a pdf", func(t *testing.T) {
		if _, err := service.ParsePDFText([]byte("hello world")); err == nil {
			t.Fatal("want error for non-PDF content")
		}
	})

	t.Run("no extractable text", func(t *testing.T) {
		content := []byte("%PDF-1.4\n1 0 obj<</XObject<</Im0 2 0 R>>>>\nstream\n\xff\xd8binary\nendstream\n")
		if _, err := service.ParsePDFText(content); err == nil {
			t.Fatal("want error when the PDF has no text layer")
		}
	})
}

func TestParserRoutesByFormat(t *testing.T) {
	p := service.NewParser(service.NewMockOCR())

	cases := []struct {
		name      string
		file      string
		content   []byte
		wantRoute string
	}{
		{"docx via docx parser", "期末.docx", buildDOCX(t, "1. 计算二重积分。"), model.FileRouteDocxParser},
		{"text pdf via text parser", "期末.pdf", buildTextPDF("1. 计算二重积分 ∬ x dxdy。"), model.FileRouteTextParser},
		{"image via ocr", "作业.jpg", []byte("\xff\xd8\xff photo"), model.FileRouteOCR},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			file, err := p.Parse(tc.file, tc.content)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if file.FileRoute != tc.wantRoute {
				t.Errorf("FileRoute = %q, want %q", file.FileRoute, tc.wantRoute)
			}
			if strings.TrimSpace(file.Text) == "" {
				t.Error("Text is empty")
			}
			if file.Size != int64(len(tc.content)) {
				t.Errorf("Size = %d, want %d", file.Size, len(tc.content))
			}
		})
	}
}

func TestParserRejectsUnsupportedFormat(t *testing.T) {
	p := service.NewParser(service.NewMockOCR())

	if _, err := p.Parse("notes.txt", []byte("plain text")); err == nil {
		t.Fatal("want error for unsupported format")
	}
}

func TestParserPropagatesOCRFailure(t *testing.T) {
	p := service.NewParser(service.NewMockOCR())

	// The mock OCR fails on the "损坏" fixture (API.md §8.3 partial success).
	if _, err := p.Parse("损坏文件.png", []byte("bad")); err == nil {
		t.Fatal("want error when OCR fails")
	}
}

func TestMockOCRIsDeterministic(t *testing.T) {
	ocr := service.NewMockOCR()

	first, err := ocr.Recognize("2024期末.png", []byte("bytes"))
	if err != nil {
		t.Fatalf("Recognize: %v", err)
	}
	second, err := ocr.Recognize("2024期末.png", []byte("bytes"))
	if err != nil {
		t.Fatalf("Recognize: %v", err)
	}
	if first != second {
		t.Error("mock OCR is not deterministic for the same file name")
	}
	if strings.TrimSpace(first) == "" {
		t.Error("mock OCR returned empty text")
	}
}

func TestMockOCRRejectsEmptyContent(t *testing.T) {
	if _, err := service.NewMockOCR().Recognize("a.png", nil); err == nil {
		t.Fatal("want error for empty content")
	}
}
