// Package extract provides text extraction from uploaded documents (PDF, images via OCR, DOCX, RTF)
// for full-text search. Extraction is best-effort; unsupported or failing files return empty text.
package extract

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/pet-medical/api/internal/upload"
)

const maxExtractedBytes = 2 * 1024 * 1024 // cap extracted text at 2MB for DB/store

// ExtractText reads the file at absPath, detects type from content, and returns extracted plain text.
// Returns empty string and nil error when type is unsupported or extraction yields nothing.
func ExtractText(absPath string) (string, error) {
	f, err := os.Open(absPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	header := make([]byte, upload.MaxHeaderBytes)
	n, _ := io.ReadFull(f, header)
	header = header[:n]
	docType := upload.DetectDocumentType(header)
	if docType == "" {
		return "", nil
	}

	switch docType {
	case "pdf":
		return extractPDF(absPath)
	case "jpeg", "png":
		return extractImageOCR(absPath)
	case "zip":
		return extractDocx(absPath)
	case "rtf":
		return extractRTF(absPath)
	case "ole":
		// .doc binary format; skip without adding a heavy dependency
		return "", nil
	default:
		return "", nil
	}
}

func extractPDF(absPath string) (string, error) {
	f, r, err := pdf.Open(absPath)
	if err != nil {
		return "", nil
	}
	defer f.Close()
	reader, err := r.GetPlainText()
	if err != nil {
		return "", nil
	}
	b, err := io.ReadAll(io.LimitReader(reader, maxExtractedBytes*2))
	if err != nil {
		return "", nil
	}
	return truncate(string(b), maxExtractedBytes), nil
}

func extractImageOCR(absPath string) (string, error) {
	// Tesseract must be installed (e.g. Windows: add to PATH). If not found, return empty.
	path, err := exec.LookPath("tesseract")
	if err != nil {
		return "", nil
	}
	// tesseract input outputbase -l eng => writes outputbase.txt
	tmpDir, err := os.MkdirTemp("", "pet-medical-ocr-*")
	if err != nil {
		return "", nil
	}
	defer os.RemoveAll(tmpDir)
	base := filepath.Join(tmpDir, "out")
	cmd := exec.Command(path, absPath, base, "-l", "eng")
	if err := cmd.Run(); err != nil {
		return "", nil
	}
	b, err := os.ReadFile(base + ".txt")
	if err != nil {
		return "", nil
	}
	return truncate(string(b), maxExtractedBytes), nil
}

// docxText matches <w:t ...>content</w:t> in word/document.xml (any namespace prefix).
var docxText = regexp.MustCompile(`<w:t[^>]*>([^<]*)</w:t>`)

func extractDocx(absPath string) (string, error) {
	r, err := zip.OpenReader(absPath)
	if err != nil {
		return "", nil
	}
	defer r.Close()
	var docXML *zip.File
	for _, f := range r.File {
		if f.Name == "word/document.xml" {
			docXML = f
			break
		}
	}
	if docXML == nil {
		return "", nil
	}
	rc, err := docXML.Open()
	if err != nil {
		return "", nil
	}
	defer rc.Close()
	raw, err := io.ReadAll(io.LimitReader(rc, maxExtractedBytes*4))
	if err != nil {
		return "", nil
	}
	// Decode XML entities in matches (e.g. &amp; &lt;) then concatenate
	matches := docxText.FindAllSubmatch(raw, -1)
	var buf strings.Builder
	for _, m := range matches {
		if len(m) > 1 {
			buf.Write(xmlEscapedToUTF8(m[1]))
			buf.WriteByte(' ')
		}
	}
	return truncate(strings.TrimSpace(buf.String()), maxExtractedBytes), nil
}

// xmlEscapedToUTF8 replaces common XML entities; returns bytes unchanged if no entity.
func xmlEscapedToUTF8(b []byte) []byte {
	s := string(b)
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&apos;", "'")
	return []byte(s)
}

// rtfStrip removes RTF control words and groups, leaving approximate plain text.
var rtfGroup = regexp.MustCompile(`\{[^{}]*\}`)
var rtfControl = regexp.MustCompile(`\\[a-z]+\d*\s?|\\[^a-z]|\n|\r`)

func extractRTF(absPath string) (string, error) {
	b, err := os.ReadFile(absPath)
	if err != nil {
		return "", nil
	}
	if len(b) > maxExtractedBytes*2 {
		b = b[:maxExtractedBytes*2]
	}
	s := string(b)
	// Remove nested groups (simple: remove {...} repeatedly)
	for i := 0; i < 20; i++ {
		next := rtfGroup.ReplaceAllString(s, " ")
		if next == s {
			break
		}
		s = next
	}
	s = rtfControl.ReplaceAllString(s, " ")
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	return truncate(strings.TrimSpace(s), maxExtractedBytes), nil
}

func truncate(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	b := []byte(s)
	if len(b) > maxBytes {
		b = b[:maxBytes]
	}
	return string(bytes.TrimRight(b, "\x00"))
}
