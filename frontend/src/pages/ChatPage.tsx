import { useState, useEffect, useRef, useCallback } from 'react'
import { Send, GitBranch, Square, Paperclip, X, FileText, Image as ImageIcon } from 'lucide-react'
import { api } from '../api'
import { navigate } from '../router'
import { MessageBubble, StepPanel } from '../components/ChatMessages'
import { useLogoState } from '../components/LogoState'
import type { ChatMessage, Step } from '../types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

const PAGE_SIZE = 10
const ACTIVE_POLL = 400
const IDLE_POLL = 8000
const MAX_FILE_BYTES = 10 * 1024 * 1024

interface Props {
  sessionId: string
  onFirstMessage?: () => void
}

interface PendingFile {
  id: string
  file: File
  previewUrl?: string
}

export default function ChatPage({ sessionId, onFirstMessage }: Props) {
  const isPlanSession = sessionId.startsWith('plan:')
  const planId = isPlanSession ? sessionId.split(':')[1] : undefined

  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [input, setInput] = useState('')
  const [sending, setSending] = useState(false)
  const [stopping, setStopping] = useState(false)
  const [regenerating, setRegenerating] = useState(false)
  const [hasMore, setHasMore] = useState(false)
  const [loadingMore, setLoadingMore] = useState(false)
  const [activeStep, setActiveStep] = useState<Step | null>(null)
  const [files, setFiles] = useState<PendingFile[]>([])
  const [dragOver, setDragOver] = useState(false)
  const [attachError, setAttachError] = useState<string | null>(null)
  const bottomRef = useRef<HTMLDivElement>(null)
  const scrollRef = useRef<HTMLDivElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const prependingRef = useRef(false)
  const userScrolledUp = useRef(false)

  const hasPending = messages.some(m => m.status === 'pending')
  const { setState: setLogoState } = useLogoState()

  useEffect(() => {
    const pending = messages.find(m => m.status === 'pending' && m.role === 'assistant')
    if (!pending) {
      setLogoState('idle')
      return
    }
    const runningStep = pending.steps?.find(s => s.status === 'running')
    if (runningStep) {
      setLogoState('working')
    } else if (pending.content && pending.content.length > 0) {
      setLogoState('typing')
    } else {
      setLogoState('thinking')
    }
  }, [messages, setLogoState])

  useEffect(() => {
    return () => setLogoState('idle')
  }, [setLogoState])

  const lastAssistantId = (() => {
    for (let i = messages.length - 1; i >= 0; i--) {
      if (messages[i].role === 'assistant') return messages[i].id
    }
    return null
  })()

  const pollLatest = useCallback(async () => {
    if (!sessionId) return
    try {
      const latest = await api.chat.listMessages({ sessionId, limit: 4, offset: 0 })
      if (latest.length === 0) return
      setMessages(prev => {
        const byId = new Map<string, ChatMessage>()
        for (const m of prev) byId.set(m.id, m)
        for (const m of latest) byId.set(m.id, m)
        const next = Array.from(byId.values()).sort((a, b) =>
          new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
        )
        if (prev.length === next.length && prev.every((m, i) => JSON.stringify(m) === JSON.stringify(next[i]))) return prev
        return next
      })
      const last = latest[latest.length - 1]
      if (last) {
        setActiveStep(prev => {
          if (!prev || !last.steps) return prev
          return last.steps.find((s: Step) => s.id === prev.id) ?? prev
        })
      }
    } catch {}
  }, [sessionId])

  useEffect(() => {
    if (!sessionId) return
    const interval = (isPlanSession || hasPending) ? ACTIVE_POLL : IDLE_POLL
    pollRef.current = setInterval(pollLatest, interval)
    return () => { if (pollRef.current) clearInterval(pollRef.current) }
  }, [sessionId, isPlanSession, hasPending, pollLatest])

  useEffect(() => {
    setMessages([])
    setHasMore(false)
    setActiveStep(null)
    setFiles(prev => { prev.forEach(p => p.previewUrl && URL.revokeObjectURL(p.previewUrl)); return [] })
    setAttachError(null)
    userScrolledUp.current = false
    if (sessionId) loadFull()
  }, [sessionId])

  useEffect(() => {
    return () => {
      files.forEach(p => p.previewUrl && URL.revokeObjectURL(p.previewUrl))
    }
  }, [files])

  useEffect(() => {
    if (prependingRef.current) { prependingRef.current = false; return }
    if (!userScrolledUp.current) bottomRef.current?.scrollIntoView({ behavior: 'instant' })
  }, [messages])

  function handleScroll() {
    const el = scrollRef.current
    if (!el) return
    userScrolledUp.current = el.scrollHeight - el.scrollTop - el.clientHeight > 80
  }

  async function loadFull() {
    try {
      const msgs = await api.chat.listMessages({ sessionId, limit: PAGE_SIZE, offset: 0 })
      setMessages(msgs)
      setHasMore(msgs.length === PAGE_SIZE)
    } catch {}
  }

  async function loadMoreMessages() {
    if (loadingMore || !hasMore) return
    setLoadingMore(true)
    try {
      const offset = messages.length
      const older = await api.chat.listMessages({ sessionId, limit: PAGE_SIZE, offset })
      if (older.length < PAGE_SIZE) setHasMore(false)
      if (older.length > 0) {
        prependingRef.current = true
        setMessages(prev => {
          const byId = new Map<string, ChatMessage>()
          for (const m of older) byId.set(m.id, m)
          for (const m of prev) byId.set(m.id, m)
          return Array.from(byId.values()).sort((a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime())
        })
      }
    } catch {
      setHasMore(false)
    } finally {
      setLoadingMore(false)
    }
  }

  function addFiles(list: FileList | File[]) {
    setAttachError(null)
    const incoming = Array.from(list)
    const additions: PendingFile[] = []
    for (const f of incoming) {
      if (f.size > MAX_FILE_BYTES) {
        setAttachError(`${f.name}: too large (max 10 MB)`)
        continue
      }
      const id = `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`
      const previewUrl = f.type.startsWith('image/') ? URL.createObjectURL(f) : undefined
      additions.push({ id, file: f, previewUrl })
    }
    if (additions.length) setFiles(prev => [...prev, ...additions])
  }

  function removeFile(id: string) {
    setFiles(prev => {
      const next = prev.filter(p => p.id !== id)
      const removed = prev.find(p => p.id === id)
      if (removed?.previewUrl) URL.revokeObjectURL(removed.previewUrl)
      return next
    })
  }

  async function fileToBase64(file: File): Promise<string> {
    const buf = await file.arrayBuffer()
    const bytes = new Uint8Array(buf)
    let binary = ''
    const chunkSize = 0x8000
    for (let i = 0; i < bytes.length; i += chunkSize) {
      binary += String.fromCharCode.apply(null, Array.from(bytes.subarray(i, i + chunkSize)))
    }
    return btoa(binary)
  }

  async function send() {
    if (sending || !sessionId) return
    const text = input.trim()
    if (!text && files.length === 0) return
    setInput('')
    setSending(true)
    userScrolledUp.current = false
    try {
      const payloadFiles = await Promise.all(files.map(async pf => ({
        fileName: pf.file.name,
        mimeType: pf.file.type || undefined,
        dataBase64: await fileToBase64(pf.file),
      })))
      const res = await api.chat.sendMessage(sessionId, text, payloadFiles.length ? payloadFiles : undefined)
      setMessages(prev => [...prev, res.userMessage, res.assistantMessage])
      files.forEach(p => p.previewUrl && URL.revokeObjectURL(p.previewUrl))
      setFiles([])
      onFirstMessage?.()
    } catch (e) {
      console.error(e)
    } finally {
      setSending(false)
    }
  }

  async function stop() {
    if (!sessionId || stopping) return
    setStopping(true)
    try {
      await api.chat.stopSession(sessionId)
      await pollLatest()
    } catch (e) {
      console.error(e)
    } finally {
      setStopping(false)
    }
  }

  async function regenerate() {
    if (!sessionId || regenerating || hasPending) return
    setRegenerating(true)
    try {
      const res = await api.chat.regenerate(sessionId)
      setMessages(prev => {
        const next = prev.filter(m => !(m.role === 'assistant' && m.id === lastAssistantId))
        return [...next, res.assistantMessage]
      })
      userScrolledUp.current = false
    } catch (e) {
      console.error(e)
    } finally {
      setRegenerating(false)
    }
  }

  function onDragEnter(e: React.DragEvent) {
    if (isPlanSession) return
    if (!e.dataTransfer?.types?.includes('Files')) return
    e.preventDefault()
    setDragOver(true)
  }
  function onDragOver(e: React.DragEvent) {
    if (isPlanSession) return
    if (!e.dataTransfer?.types?.includes('Files')) return
    e.preventDefault()
  }
  function onDragLeave(e: React.DragEvent) {
    if (e.currentTarget === e.target) setDragOver(false)
  }
  function onDrop(e: React.DragEvent) {
    if (isPlanSession) return
    e.preventDefault()
    setDragOver(false)
    if (e.dataTransfer?.files?.length) addFiles(e.dataTransfer.files)
  }

  if (!sessionId) {
    return (
      <div className="flex items-center justify-center h-full text-zinc-600 text-sm">
        Select a chat or create a new one
      </div>
    )
  }

  return (
    <div className="flex h-full min-w-0">
      <div
        className="flex flex-col h-full min-w-0 flex-1 relative"
        onDragEnter={onDragEnter}
        onDragOver={onDragOver}
        onDragLeave={onDragLeave}
        onDrop={onDrop}
      >
        <div ref={scrollRef} onScroll={handleScroll} className="flex-1 overflow-auto px-6 py-5 space-y-1">
          {hasMore && messages.length > 0 && (
            <div className="flex justify-center mb-4">
              <Button variant="secondary" size="sm" onClick={loadMoreMessages} disabled={loadingMore}>
                {loadingMore ? 'Loading...' : `Load previous ${PAGE_SIZE}`}
              </Button>
            </div>
          )}
          {messages.length === 0 && (
            <div className="flex items-center justify-center h-full text-zinc-600 text-sm">
              Send a message to start the conversation
            </div>
          )}
          <div className="space-y-3">
            {messages.map(msg => (
              <MessageBubble
                key={msg.id}
                msg={msg}
                onStepClick={setActiveStep}
                canRegenerate={!isPlanSession && msg.id === lastAssistantId && !hasPending && msg.status !== 'pending'}
                onRegenerate={regenerate}
                regenerating={regenerating && msg.id === lastAssistantId}
              />
            ))}
          </div>
          <div ref={bottomRef} />
        </div>

        {dragOver && !isPlanSession && (
          <div className="absolute inset-0 z-20 pointer-events-none flex items-center justify-center bg-teal-500/10 border-2 border-dashed border-teal-500/60">
            <div className="bg-white/90 dark:bg-zinc-900/90 border border-teal-500/40 rounded-lg px-5 py-3 text-sm text-teal-600 dark:text-teal-400 font-medium shadow-lg">
              Drop files to attach
            </div>
          </div>
        )}

        {isPlanSession ? (
          <div className="px-6 py-3 border-t border-zinc-200/80 dark:border-zinc-800/80 bg-zinc-50/40 dark:bg-zinc-900/40 shrink-0">
            <div className="flex items-center gap-2 text-xs text-zinc-500 dark:text-zinc-600">
              <GitBranch size={13} className="text-amber-500/70" />
              <span>Plan execution chat (read-only)</span>
              {planId && (
                <button onClick={() => navigate({ page: 'plans', planId })} className="text-amber-500/70 hover:text-amber-400 hover:underline ml-auto">
                  Open plan
                </button>
              )}
            </div>
          </div>
        ) : (
          <div className="px-6 py-3 border-t border-zinc-200/80 dark:border-zinc-800/80 bg-zinc-50/40 dark:bg-zinc-900/40 shrink-0 space-y-2">
            {files.length > 0 && (
              <div className="flex flex-wrap gap-2">
                {files.map(pf => (
                  <FilePreview key={pf.id} file={pf} onRemove={() => removeFile(pf.id)} />
                ))}
              </div>
            )}
            {attachError && (
              <div className="text-xs text-rose-500 dark:text-rose-400">{attachError}</div>
            )}
            <div className="flex gap-2 items-start">
              <input
                ref={fileInputRef}
                type="file"
                multiple
                className="hidden"
                onChange={e => {
                  if (e.target.files) addFiles(e.target.files)
                  e.target.value = ''
                }}
              />
              <Button
                type="button"
                variant="secondary"
                onClick={() => fileInputRef.current?.click()}
                disabled={sending || hasPending}
                title="Attach files"
              >
                <Paperclip size={16} />
              </Button>
              <Input
                value={input}
                onChange={e => setInput(e.target.value)}
                onKeyDown={e => e.key === 'Enter' && !e.shiftKey && send()}
                placeholder="Type a message..."
                className="flex-1 bg-white dark:bg-zinc-900 border-zinc-200 dark:border-zinc-800"
                disabled={sending || hasPending}
              />
              {hasPending ? (
                <Button
                  onClick={stop}
                  disabled={stopping}
                  variant="secondary"
                  title="Stop generation"
                >
                  <Square size={14} className="fill-current" />
                </Button>
              ) : (
                <Button
                  onClick={send}
                  disabled={sending || (!input.trim() && files.length === 0)}
                >
                  <Send size={16} />
                </Button>
              )}
            </div>
          </div>
        )}

      </div>
      {activeStep && <StepPanel step={activeStep} onClose={() => setActiveStep(null)} />}
    </div>
  )
}

function FilePreview({ file, onRemove }: { file: PendingFile; onRemove: () => void }) {
  const f = file.file
  const sizeKb = f.size >= 1024 * 1024
    ? `${(f.size / (1024 * 1024)).toFixed(1)} MB`
    : `${Math.max(1, Math.round(f.size / 1024))} KB`
  const isImage = !!file.previewUrl

  return (
    <div className="group relative flex items-center gap-2 pl-1 pr-2 py-1 rounded-md border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900">
      {isImage ? (
        <img src={file.previewUrl} alt={f.name} className="w-8 h-8 object-cover rounded" />
      ) : (
        <div className="w-8 h-8 rounded bg-zinc-100 dark:bg-zinc-800 flex items-center justify-center text-zinc-500">
          {f.type.startsWith('image/') ? <ImageIcon size={14} /> : <FileText size={14} />}
        </div>
      )}
      <div className="text-[11px] leading-tight max-w-[160px]">
        <div className="truncate text-zinc-700 dark:text-zinc-300">{f.name}</div>
        <div className="text-zinc-500 dark:text-zinc-600">{sizeKb}</div>
      </div>
      <button
        type="button"
        onClick={onRemove}
        className="ml-1 p-0.5 rounded text-zinc-400 hover:text-rose-500 hover:bg-rose-500/10"
        title="Remove"
      >
        <X size={12} />
      </button>
    </div>
  )
}
