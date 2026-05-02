import { useEffect, useState } from 'react'
import { Clock, Download, FileText, Image as ImageIcon } from '@/lib/icons'

export interface ParsedAttachment {
  id: string
  name: string
  mime: string
  size: number
  createdAt?: Date
  expiresAt?: Date
}

export type ContentBlock =
  | { type: 'text'; text: string }
  | { type: 'attachments'; items: ParsedAttachment[] }

const ATTACHMENT_TAG_RE = /<attachment\s+([^>]*?)\/>/g
const ATTR_RE = /([a-zA-Z_][\w-]*)\s*=\s*"([^"]*)"/g

function decodeEntities(value: string): string {
  return value
    .replace(/&quot;/g, '"')
    .replace(/&apos;/g, "'")
    .replace(/&#39;/g, "'")
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&amp;/g, '&')
}

function parseAttrs(attrs: string): Record<string, string> {
  const out: Record<string, string> = {}
  let m: RegExpExecArray | null
  ATTR_RE.lastIndex = 0
  while ((m = ATTR_RE.exec(attrs)) !== null) {
    out[m[1]] = decodeEntities(m[2])
  }
  return out
}

function parseDate(value: string | undefined): Date | undefined {
  if (!value) return undefined
  const d = new Date(value)
  return Number.isNaN(d.getTime()) ? undefined : d
}

export function parseAttachmentBlocks(content: string): ContentBlock[] {
  if (!content) return []

  type Part = { type: 'text'; text: string } | ({ type: 'attachment' } & ParsedAttachment)
  const parts: Part[] = []
  let lastIndex = 0
  let match: RegExpExecArray | null
  ATTACHMENT_TAG_RE.lastIndex = 0
  while ((match = ATTACHMENT_TAG_RE.exec(content)) !== null) {
    if (match.index > lastIndex) {
      parts.push({ type: 'text', text: content.slice(lastIndex, match.index) })
    }
    const attrs = parseAttrs(match[1])
    parts.push({
      type: 'attachment',
      id: attrs.id ?? '',
      name: attrs.name || 'attachment',
      mime: attrs.mime ?? '',
      size: Number(attrs.size) || 0,
      createdAt: parseDate(attrs.created),
      expiresAt: parseDate(attrs.expires),
    })
    lastIndex = ATTACHMENT_TAG_RE.lastIndex
  }
  if (lastIndex < content.length) {
    parts.push({ type: 'text', text: content.slice(lastIndex) })
  }

  const blocks: ContentBlock[] = []
  let textBuffer = ''
  let attachmentBuffer: ParsedAttachment[] = []

  const flushText = () => {
    const trimmed = textBuffer.replace(/^[ \t]*\n+|\n+[ \t]*$/g, '')
    if (trimmed.length > 0) {
      blocks.push({ type: 'text', text: trimmed })
    }
    textBuffer = ''
  }

  const flushAttachments = () => {
    if (attachmentBuffer.length > 0) {
      blocks.push({ type: 'attachments', items: attachmentBuffer })
      attachmentBuffer = []
    }
  }

  for (const part of parts) {
    if (part.type === 'text') {
      flushAttachments()
      textBuffer += part.text
    } else {
      flushText()
      attachmentBuffer.push({
        id: part.id,
        name: part.name,
        mime: part.mime,
        size: part.size,
        createdAt: part.createdAt,
        expiresAt: part.expiresAt,
      })
    }
  }
  flushText()
  flushAttachments()

  return blocks
}

export function hasAttachmentTags(content: string): boolean {
  if (!content) return false
  ATTACHMENT_TAG_RE.lastIndex = 0
  return ATTACHMENT_TAG_RE.test(content)
}

function formatBytes(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes <= 0) return ''
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function isImageAttachment(a: ParsedAttachment): boolean {
  if (a.mime && a.mime.startsWith('image/')) return true
  return /\.(png|jpe?g|gif|webp|svg|bmp|avif)$/i.test(a.name)
}

function formatRelative(target: Date, now: number = Date.now()): string {
  const diffSec = Math.round((target.getTime() - now) / 1000)
  const past = diffSec < 0
  const abs = Math.abs(diffSec)
  if (abs < 45) return past ? 'just now' : 'in <1 min'
  const min = Math.round(abs / 60)
  if (min < 60) return past ? `${min} min ago` : `in ${min} min`
  const hr = Math.round(min / 60)
  if (hr < 24) return past ? `${hr} h ago` : `in ${hr} h`
  const day = Math.round(hr / 24)
  return past ? `${day} d ago` : `in ${day} d`
}

function buildTitle(a: ParsedAttachment, expired: boolean): string {
  const parts: string[] = [a.name]
  const now = Date.now()
  if (expired && a.expiresAt) {
    parts.push(`expired ${formatRelative(a.expiresAt, now)}`)
  } else {
    if (a.createdAt) parts.push(`uploaded ${formatRelative(a.createdAt, now)}`)
    if (a.expiresAt) parts.push(`expires ${formatRelative(a.expiresAt, now)}`)
  }
  return parts.join(' · ')
}

function useExpiryStatus(expiresAt: Date | undefined): boolean {
  const expiresMs = expiresAt ? expiresAt.getTime() : null
  const [expired, setExpired] = useState(() =>
    expiresMs !== null ? Date.now() >= expiresMs : false,
  )
  useEffect(() => {
    if (expiresMs === null) {
      setExpired(false)
      return undefined
    }
    const remaining = expiresMs - Date.now()
    if (remaining <= 0) {
      setExpired(true)
      return undefined
    }
    setExpired(false)
    const id = window.setTimeout(() => setExpired(true), remaining + 250)
    return () => window.clearTimeout(id)
  }, [expiresMs])
  return expired
}

export type AttachmentBadgeVariant = 'user' | 'message'

export function AttachmentBadge({
  attachment,
  sessionId,
  variant = 'message',
}: {
  attachment: ParsedAttachment
  sessionId: string
  variant?: AttachmentBadgeVariant
}) {
  const expired = useExpiryStatus(attachment.expiresAt)
  const isImage = isImageAttachment(attachment)
  const Icon = isImage ? ImageIcon : FileText
  const sizeLabel = formatBytes(attachment.size)
  const href = !expired && attachment.id ? `/api/artifacts/${sessionId}/${attachment.id}` : undefined

  const isUser = variant === 'user'

  let containerClass: string
  if (expired) {
    containerClass = isUser
      ? 'inline-flex items-center gap-1.5 px-2 py-1 rounded-md text-[12px] bg-white/10 text-white/55 max-w-full cursor-default'
      : 'inline-flex items-center gap-1.5 px-2 py-1 rounded-md text-[12px] bg-zinc-200/60 dark:bg-zinc-800/60 text-zinc-500 dark:text-zinc-500 max-w-full cursor-default'
  } else if (isUser) {
    containerClass =
      'inline-flex items-center gap-1.5 px-2 py-1 rounded-md text-[12px] bg-white/15 text-white hover:bg-white/25 transition-colors max-w-full'
  } else {
    containerClass =
      'inline-flex items-center gap-1.5 px-2 py-1 rounded-md text-[12px] bg-zinc-200 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-400 hover:text-zinc-800 dark:hover:text-zinc-200 transition-colors max-w-full'
  }

  const sizeClass = isUser
    ? `text-[10px] tabular-nums shrink-0 ${expired ? 'text-white/45' : 'text-white/70'}`
    : `text-[10px] tabular-nums shrink-0 ${expired ? 'text-zinc-500/70 dark:text-zinc-500/70' : 'text-zinc-500 dark:text-zinc-500'}`
  const trailingIconClass = isUser
    ? `shrink-0 ${expired ? 'text-white/55' : 'text-white/70'}`
    : `shrink-0 ${expired ? 'text-zinc-500 dark:text-zinc-500' : 'text-zinc-500 dark:text-zinc-500'}`

  const nameClass = `truncate min-w-0 ${expired ? 'line-through' : ''}`

  const inner = (
    <>
      <Icon size={13} className="shrink-0" />
      <span className={nameClass}>{attachment.name}</span>
      {expired ? (
        <span className={`${sizeClass} uppercase tracking-wide`}>expired</span>
      ) : (
        sizeLabel && <span className={sizeClass}>{sizeLabel}</span>
      )}
      {expired ? (
        <Clock size={11} className={trailingIconClass} />
      ) : (
        href && <Download size={11} className={trailingIconClass} />
      )}
    </>
  )

  const title = buildTitle(attachment, expired)

  if (!href) {
    return (
      <span className={containerClass} title={title}>
        {inner}
      </span>
    )
  }

  return (
    <a
      href={href}
      download={attachment.name}
      className={containerClass}
      title={title}
      target="_blank"
      rel="noreferrer"
    >
      {inner}
    </a>
  )
}

export function AttachmentBadgeRow({
  items,
  sessionId,
  variant = 'message',
}: {
  items: ParsedAttachment[]
  sessionId: string
  variant?: AttachmentBadgeVariant
}) {
  if (!items.length) return null
  return (
    <div className="flex flex-wrap gap-1.5">
      {items.map(a => (
        <AttachmentBadge
          key={`${a.id}-${a.name}`}
          attachment={a}
          sessionId={sessionId}
          variant={variant}
        />
      ))}
    </div>
  )
}
