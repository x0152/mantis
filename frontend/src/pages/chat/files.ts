import type { PendingFile } from './types'

export const MAX_FILE_BYTES = 10 * 1024 * 1024 * 1024

export async function fileToBase64(file: File): Promise<string> {
  const buf = await file.arrayBuffer()
  const bytes = new Uint8Array(buf)
  let binary = ''
  const chunkSize = 0x8000
  for (let i = 0; i < bytes.length; i += chunkSize) {
    binary += String.fromCharCode.apply(null, Array.from(bytes.subarray(i, i + chunkSize)))
  }
  return btoa(binary)
}

export function buildPendingFile(f: File): PendingFile {
  const id = `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`
  const previewUrl = f.type.startsWith('image/') ? URL.createObjectURL(f) : undefined
  return { id, file: f, previewUrl }
}

export function revokePreview(p: PendingFile): void {
  if (p.previewUrl) URL.revokeObjectURL(p.previewUrl)
}
