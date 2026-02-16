/**
 * Validates file type by reading magic bytes (file signature), not extension or MIME.
 * Must match backend internal/upload/validate.go logic.
 */

const MAX_HEADER_BYTES = 512

const SIG = {
  PDF: [0x25, 0x50, 0x44, 0x46], // %PDF
  JPEG: [0xff, 0xd8, 0xff],
  PNG: [0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a],
  GIF87: [0x47, 0x49, 0x46, 0x38, 0x37, 0x61], // GIF87a
  GIF89: [0x47, 0x49, 0x46, 0x38, 0x39, 0x61], // GIF89a
  ZIP: [0x50, 0x4b, 0x03, 0x04],
  ZIP2: [0x50, 0x4b, 0x05, 0x06],
  ZIP3: [0x50, 0x4b, 0x07, 0x08],
  OLE: [0xd0, 0xcf, 0x11, 0xe0, 0xa1, 0xb1, 0x1a, 0xe1],
  RTF: [0x7b, 0x5c, 0x72, 0x74, 0x66], // {\rtf
  RIFF: [0x52, 0x49, 0x46, 0x46],
  WEBP: [0x57, 0x45, 0x42, 0x50], // at offset 8
}

function hasPrefix(buf: Uint8Array, sig: number[]): boolean {
  if (buf.length < sig.length) return false
  for (let i = 0; i < sig.length; i++) if (buf[i] !== sig[i]) return false
  return true
}

function readFileHeader(file: File, maxBytes: number): Promise<Uint8Array> {
  const blob = file.slice(0, maxBytes)
  return new Promise((resolve, reject) => {
    const r = new FileReader()
    r.onload = () => resolve(new Uint8Array(r.result as ArrayBuffer))
    r.onerror = () => reject(new Error('Failed to read file'))
    r.readAsArrayBuffer(blob)
  })
}

/** Allowed documents: PDF, Word (ZIP/OLE), RTF, PNG, JPEG */
function allowedDocument(header: Uint8Array): boolean {
  if (hasPrefix(header, SIG.PDF)) return true
  if (hasPrefix(header, SIG.JPEG)) return true
  if (hasPrefix(header, SIG.PNG)) return true
  if (hasPrefix(header, SIG.ZIP) || hasPrefix(header, SIG.ZIP2) || hasPrefix(header, SIG.ZIP3)) return true
  if (hasPrefix(header, SIG.OLE)) return true
  if (hasPrefix(header, SIG.RTF)) return true
  return false
}

/** Allowed images: JPEG, PNG, GIF, WebP */
function allowedImage(header: Uint8Array): boolean {
  if (hasPrefix(header, SIG.JPEG)) return true
  if (hasPrefix(header, SIG.PNG)) return true
  if (hasPrefix(header, SIG.GIF87) || hasPrefix(header, SIG.GIF89)) return true
  if (header.length >= 12 && hasPrefix(header, SIG.RIFF) && header[8] === 0x57 && header[9] === 0x45 && header[10] === 0x42 && header[11] === 0x50) return true
  return false
}

const DOCUMENT_ALLOWED = 'Documents must be PDF, Word (.doc/.docx), RTF, or PNG/JPEG images. File type was not recognized.'
const IMAGE_ALLOWED = 'Photos must be JPEG, PNG, GIF, or WebP. File type was not recognized.'

export async function validateDocumentFile(file: File): Promise<{ ok: true } | { ok: false; error: string }> {
  const header = await readFileHeader(file, MAX_HEADER_BYTES)
  if (allowedDocument(header)) return { ok: true }
  return { ok: false, error: DOCUMENT_ALLOWED }
}

export async function validateImageFile(file: File): Promise<{ ok: true } | { ok: false; error: string }> {
  const header = await readFileHeader(file, MAX_HEADER_BYTES)
  if (allowedImage(header)) return { ok: true }
  return { ok: false, error: IMAGE_ALLOWED }
}
