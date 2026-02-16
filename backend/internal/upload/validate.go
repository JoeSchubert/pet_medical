// Package upload provides file upload validation by magic bytes (file signatures).
// Validation is done by inspecting actual file content, not extension or Content-Type.
package upload

import (
	"bytes"
	"path/filepath"
	"strings"
)

const maxBasenameLen = 200

// SafeBasename returns a safe filename for storage: path separators and unsafe characters
// are removed or replaced to prevent path traversal. Result is truncated to maxBasenameLen.
func SafeBasename(filename string) string {
	base := filepath.Base(filename)
	var b strings.Builder
	for _, r := range base {
		if r == '/' || r == '\\' || r == 0 {
			continue
		}
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
		if b.Len() >= maxBasenameLen {
			break
		}
	}
	s := b.String()
	if s == "" {
		return "file"
	}
	return s
}

// MaxHeaderBytes is the number of bytes we read to detect file type (enough for all supported signatures).
const MaxHeaderBytes = 512

// Magic-byte signatures (hex). Order can matter for disambiguation (e.g. ZIP before PDF).
var (
	sigPDF = []byte("%PDF")                           // 25 50 44 46
	sigJPEG = []byte{0xFF, 0xD8, 0xFF}                // JPEG
	sigPNG  = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	sigGIF87 = []byte("GIF87a")                        // 47 49 46 38 37 61
	sigGIF89 = []byte("GIF89a")                        // 47 49 46 38 39 61
	sigZIP  = []byte{0x50, 0x4B, 0x03, 0x04}          // ZIP (also docx, xlsx, odt)
	sigZIP2 = []byte{0x50, 0x4B, 0x05, 0x06}
	sigZIP3 = []byte{0x50, 0x4B, 0x07, 0x08}
	sigOLE  = []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1} // OLE/CFB (.doc, .xls)
	sigRTF  = []byte("{\\rtf")                         // 7B 5C 72 74 66
	sigRIFF = []byte("RIFF")                           // 52 49 46 46
	sigWEBP = []byte("WEBP")                           // at offset 8 in RIFF file
)

func hasPrefix(b, prefix []byte) bool {
	return len(b) >= len(prefix) && bytes.Equal(b[:len(prefix)], prefix)
}

// DetectDocumentType returns a non-empty type if the header is an allowed document format.
// Allowed: PDF, Word (.doc OLE / .docx ZIP), RTF, ODT (ZIP), and PNG/JPEG for scanned docs.
func DetectDocumentType(header []byte) string {
	if hasPrefix(header, sigPDF) {
		return "pdf"
	}
	if hasPrefix(header, sigJPEG) {
		return "jpeg"
	}
	if hasPrefix(header, sigPNG) {
		return "png"
	}
	if hasPrefix(header, sigZIP) || hasPrefix(header, sigZIP2) || hasPrefix(header, sigZIP3) {
		return "zip" // docx, xlsx, odt, etc.
	}
	if hasPrefix(header, sigOLE) {
		return "ole" // .doc, .xls
	}
	if hasPrefix(header, sigRTF) {
		return "rtf"
	}
	return ""
}

// AllowedDocument returns true if the file header matches an allowed document type.
func AllowedDocument(header []byte) bool {
	return DetectDocumentType(header) != ""
}

// DetectImageType returns a non-empty type if the header is an allowed image format.
// Allowed: JPEG, PNG, GIF, WebP.
func DetectImageType(header []byte) string {
	if hasPrefix(header, sigJPEG) {
		return "jpeg"
	}
	if hasPrefix(header, sigPNG) {
		return "png"
	}
	if hasPrefix(header, sigGIF87) || hasPrefix(header, sigGIF89) {
		return "gif"
	}
	// WebP: RIFF....WEBP (bytes 0-3 RIFF, 8-11 WEBP)
	if len(header) >= 12 && hasPrefix(header, sigRIFF) && bytes.Equal(header[8:12], sigWEBP) {
		return "webp"
	}
	return ""
}

// AllowedImage returns true if the file header matches an allowed image type.
func AllowedImage(header []byte) bool {
	return DetectImageType(header) != ""
}
