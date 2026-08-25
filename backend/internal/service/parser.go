// Package service holds the concrete business logic invoked by workflows.
package service

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"

	"lingxi-claw/internal/model"
	"lingxi-claw/internal/router"
)

// Parser turns an uploaded file into plain text. It owns the three branches of
// the routing tree in API.md §8.2: DOCX parser, text-PDF parser, and OCR.
type Parser struct {
	ocr OCR
}

// OCR extracts text from images and scanned PDFs.
type OCR interface {
	Recognize(name string, content []byte) (string, error)
}

// NewParser wires a parser to an OCR backend.
func NewParser(ocr OCR) *Parser {
	return &Parser{ocr: ocr}
}

// Parse routes one file and returns its extracted text. The returned
// UploadedFile records which branch handled the file so the API can report
// file_route (API.md §12).
func (p *Parser) Parse(name string, content []byte) (model.UploadedFile, error) {
	decision := router.RouteFile(name, content)
	if !decision.Supported {
		return model.UploadedFile{}, fmt.Errorf("不支持的文件格式: %s", decision.Ext)
	}

	file := model.UploadedFile{
		Name:      name,
		Size:      int64(len(content)),
		Ext:       decision.Ext,
		FileRoute: decision.FileRoute,
	}

	var (
		text string
		err  error
	)
	switch decision.FileRoute {
	case model.FileRouteDocxParser:
		text, err = ParseDOCX(content)
	case model.FileRouteTextParser:
		text, err = ParsePDFText(content)
	case model.FileRouteOCR:
		text, err = p.ocr.Recognize(name, content)
	default:
		err = fmt.Errorf("无法确定解析方式")
	}
	if err != nil {
		return model.UploadedFile{}, err
	}

	text = normalizeText(text)
	if strings.TrimSpace(text) == "" {
		return model.UploadedFile{}, fmt.Errorf("文件无法解析")
	}
	file.Text = text
	return file, nil
}

// ParseDOCX extracts the text runs from word/document.xml inside a .docx zip.
func ParseDOCX(content []byte) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return "", fmt.Errorf("DOCX 文件无法读取: %w", err)
	}

	var doc *zip.File
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			doc = f
			break
		}
	}
	if doc == nil {
		return "", fmt.Errorf("DOCX 缺少 word/document.xml")
	}

	rc, err := doc.Open()
	if err != nil {
		return "", fmt.Errorf("DOCX 内容无法打开: %w", err)
	}
	defer rc.Close()

	// Cap the decompressed size so a zip bomb cannot exhaust memory.
	raw, err := io.ReadAll(io.LimitReader(rc, 32<<20))
	if err != nil {
		return "", fmt.Errorf("DOCX 内容无法读取: %w", err)
	}
	return docxXMLToText(raw)
}

// docxXMLToText walks the WordprocessingML stream, joining w:t runs and turning
// paragraph boundaries into newlines.
func docxXMLToText(raw []byte) (string, error) {
	dec := xml.NewDecoder(bytes.NewReader(raw))
	var (
		sb     strings.Builder
		inText bool
	)
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("DOCX XML 解析失败: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "t":
				inText = true
			case "tab":
				sb.WriteString("\t")
			case "br":
				sb.WriteString("\n")
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "t":
				inText = false
			case "p":
				sb.WriteString("\n")
			}
		case xml.CharData:
			if inText {
				sb.Write(t)
			}
		}
	}
	return sb.String(), nil
}

var (
	pdfStreamRe = regexp.MustCompile(`(?s)stream\r?\n(.*?)endstream`)
	pdfTextRe   = regexp.MustCompile(`\(((?:\\.|[^\\()])*)\)\s*(?:Tj|TJ)`)
	pdfArrayRe  = regexp.MustCompile(`\[(.*?)\]\s*TJ`)
)

// ParsePDFText extracts text from the uncompressed content streams of a PDF.
// Flate-compressed streams are skipped, so a PDF whose text is entirely
// compressed yields no text and the caller reports FILE_PARSE_ERROR.
func ParsePDFText(content []byte) (string, error) {
	if !bytes.HasPrefix(bytes.TrimSpace(content), []byte("%PDF-")) {
		return "", fmt.Errorf("不是合法的 PDF 文件")
	}

	var sb strings.Builder
	for _, m := range pdfStreamRe.FindAllSubmatch(content, -1) {
		body := m[1]
		for _, tm := range pdfTextRe.FindAllSubmatch(body, -1) {
			sb.WriteString(decodePDFString(string(tm[1])))
		}
		for _, am := range pdfArrayRe.FindAllSubmatch(body, -1) {
			for _, tm := range pdfTextRe.FindAllSubmatch(append(am[1], []byte(" Tj")...), -1) {
				sb.WriteString(decodePDFString(string(tm[1])))
			}
		}
		sb.WriteString("\n")
	}

	if strings.TrimSpace(sb.String()) == "" {
		return "", fmt.Errorf("PDF 文本层为空或已压缩，无法提取文本")
	}
	return sb.String(), nil
}

// decodePDFString resolves the escape sequences of a PDF literal string.
func decodePDFString(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) {
			sb.WriteByte(s[i])
			continue
		}
		i++
		switch s[i] {
		case 'n':
			sb.WriteByte('\n')
		case 'r':
			sb.WriteByte('\r')
		case 't':
			sb.WriteByte('\t')
		case 'b':
			sb.WriteByte('\b')
		case 'f':
			sb.WriteByte('\f')
		default:
			sb.WriteByte(s[i])
		}
	}
	return sb.String()
}

// normalizeText collapses carriage returns and trims trailing spaces so that
// downstream question splitting sees consistent line breaks.
func normalizeText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, strings.TrimRight(line, " \t"))
	}
	s = strings.Join(out, "\n")

	// Squeeze runs of blank lines down to one.
	for strings.Contains(s, "\n\n\n") {
		s = strings.ReplaceAll(s, "\n\n\n", "\n\n")
	}
	return strings.TrimFunc(s, unicode.IsSpace)
}
