import { useState, useEffect, useMemo, useRef, useCallback } from 'react'
import { api } from '../../api'
import { MessageBubble, StepPanel } from '../../components/ChatMessages'
import { ResourcesPanel } from '../../components/ResourcesPanel'
import { useLogoState } from '../../components/LogoState'
import type { ChatMessage, ContextStatus, Step } from '../../types'
import { EmptyState } from './EmptyState'
import { Composer } from './Composer'
import { PlanBanner } from './PlanBanner'
import { DropOverlay } from './DropOverlay'
import { LoadMoreButton } from './LoadMoreButton'
import type { PendingFile } from './types'
import { MAX_FILE_BYTES, buildPendingFile, fileToBase64, revokePreview } from './files'

const PAGE_SIZE = 10
const ACTIVE_POLL = 400
const IDLE_POLL = 8000

interface Props {
  sessionId: string
  onFirstMessage?: () => void
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
  const [contextStatus, setContextStatus] = useState<ContextStatus | null>(null)
  const [files, setFiles] = useState<PendingFile[]>([])
  const [dragOver, setDragOver] = useState(false)
  const [attachError, setAttachError] = useState<string | null>(null)

  const bottomRef = useRef<HTMLDivElement>(null)
  const scrollRef = useRef<HTMLDivElement>(null)
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const prependingRef = useRef(false)
  const userScrolledUp = useRef(false)

  const hasPending = messages.some(m => m.status === 'pending')
  const activeSteps = useMemo<Step[]>(() => {
    const out: Step[] = []
    for (const m of messages) {
      if (!m.steps) continue
      for (const s of m.steps) {
        if (s.status === 'running') out.push(s)
      }
    }
    return out
  }, [messages])

  const { setState: setLogoState } = useLogoState()

  useEffect(() => {
    const pending = messages.find(m => m.status === 'pending' && m.role === 'assistant')
    if (!pending) {
      setLogoState('idle')
      return
    }
    const runningStep = pending.steps?.find(s => s.status === 'running')
    if (runningStep) setLogoState('working')
    else if (pending.content && pending.content.length > 0) setLogoState('typing')
    else setLogoState('thinking')
  }, [messages, setLogoState])

  useEffect(() => () => setLogoState('idle'), [setLogoState])

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
        const next = Array.from(byId.values()).sort(
          (a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime(),
        )
        if (prev.length === next.length && prev.every((m, i) => JSON.stringify(m) === JSON.stringify(next[i]))) {
          return prev
        }
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
    return () => {
      if (pollRef.current) clearInterval(pollRef.current)
    }
  }, [sessionId, isPlanSession, hasPending, pollLatest])

  const loadFull = useCallback(async () => {
    try {
      const msgs = await api.chat.listMessages({ sessionId, limit: PAGE_SIZE, offset: 0 })
      setMessages(msgs)
      setHasMore(msgs.length === PAGE_SIZE)
    } catch {}
  }, [sessionId])

  useEffect(() => {
    setMessages([])
    setHasMore(false)
    setActiveStep(null)
    setContextStatus(null)
    setFiles(prev => {
      prev.forEach(revokePreview)
      return []
    })
    setAttachError(null)
    userScrolledUp.current = false
    if (sessionId) loadFull()
  }, [sessionId, loadFull])

  useEffect(() => {
    if (!sessionId || isPlanSession) return
    let cancelled = false
    const refresh = async () => {
      try {
        const status = await api.chat.getContextStatus(sessionId)
        if (!cancelled) setContextStatus(status)
      } catch {}
    }
    refresh()
    const interval = setInterval(refresh, hasPending ? 2000 : 15000)
    return () => {
      cancelled = true
      clearInterval(interval)
    }
  }, [sessionId, isPlanSession, hasPending])

  useEffect(() => () => {
    files.forEach(revokePreview)
  }, [files])

  useEffect(() => {
    if (prependingRef.current) {
      prependingRef.current = false
      return
    }
    if (!userScrolledUp.current) bottomRef.current?.scrollIntoView({ behavior: 'instant' })
  }, [messages])

  function handleScroll() {
    const el = scrollRef.current
    if (!el) return
    userScrolledUp.current = el.scrollHeight - el.scrollTop - el.clientHeight > 80
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
          return Array.from(byId.values()).sort(
            (a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime(),
          )
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
        setAttachError(`${f.name}: too large (max 10 GB)`)
        continue
      }
      additions.push(buildPendingFile(f))
    }
    if (additions.length) setFiles(prev => [...prev, ...additions])
  }

  function removeFile(id: string) {
    setFiles(prev => {
      const removed = prev.find(p => p.id === id)
      if (removed) revokePreview(removed)
      return prev.filter(p => p.id !== id)
    })
  }

  async function send(override?: string | null) {
    if (sending || !sessionId) return
    const text = (typeof override === 'string' ? override : input).trim()
    if (!text && files.length === 0) return
    if (override === undefined) setInput('')
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
      files.forEach(revokePreview)
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
      <div className="flex items-center justify-center h-full font-mono text-[11.5px] lowercase tracking-tight text-zinc-500 dark:text-zinc-600">
        — select a chat or create a new one —
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
            <LoadMoreButton pageSize={PAGE_SIZE} loading={loadingMore} onClick={loadMoreMessages} />
          )}
          {messages.length === 0 && (
            <EmptyState
              disabled={isPlanSession || sending}
              onInsert={prompt => {
                setInput(prompt)
                textareaRef.current?.focus()
              }}
              onSend={prompt => {
                setInput('')
                void send(prompt)
              }}
            />
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

        {dragOver && !isPlanSession && <DropOverlay />}

        {isPlanSession ? (
          <PlanBanner planId={planId} />
        ) : (
          <Composer
            ref={textareaRef}
            input={input}
            onChangeInput={setInput}
            onSend={() => void send()}
            onStop={stop}
            sending={sending}
            hasPending={hasPending}
            stopping={stopping}
            files={files}
            onAddFiles={addFiles}
            onRemoveFile={removeFile}
            attachError={attachError}
            messages={messages}
            partial={hasMore}
            contextStatus={contextStatus}
          />
        )}
      </div>
      {activeStep && <StepPanel step={activeStep} onClose={() => setActiveStep(null)} />}
      {!isPlanSession && <ResourcesPanel activeSteps={activeSteps} />}
    </div>
  )
}
