import { useState, useEffect, useRef, useCallback } from 'react'
import { Send } from 'lucide-react'
import { api } from '../api'
import { MessageBubble, StepPanel } from '../components/ChatMessages'
import type { ChatMessage, Step } from '../types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

const PAGE_SIZE = 10
const POLL_INTERVAL = 300
const BG_POLL_INTERVAL = 7000

interface Props {
  sessionId: string
  onFirstMessage?: () => void
}

export default function ChatPage({ sessionId, onFirstMessage }: Props) {
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [input, setInput] = useState('')
  const [sending, setSending] = useState(false)
  const [hasMore, setHasMore] = useState(false)
  const [loadingMore, setLoadingMore] = useState(false)
  const [activeStep, setActiveStep] = useState<Step | null>(null)
  const bottomRef = useRef<HTMLDivElement>(null)
  const scrollRef = useRef<HTMLDivElement>(null)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const bgPollRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const prependingRef = useRef(false)
  const userScrolledUp = useRef(false)

  const hasPending = messages.some(m => m.status === 'pending')

  const poll = useCallback(async () => {
    try {
      if (!sessionId) return
      const latest = await api.chat.listMessages({ sessionId, limit: 1, offset: 0 })
      const m = latest[0]
      if (!m) return
      setMessages(prev => {
        const idx = prev.findIndex(x => x.id === m.id)
        if (idx === -1) return [...prev, m]
        const next = prev.slice()
        next[idx] = m
        return next
      })
      setActiveStep(prev => {
        if (!prev || !m.steps) return prev
        const fresh = m.steps.find(s => s.id === prev.id)
        return fresh ?? prev
      })
    } catch {}
  }, [sessionId])

  useEffect(() => {
    if (hasPending) {
      pollRef.current = setInterval(poll, POLL_INTERVAL)
    } else if (pollRef.current) {
      clearInterval(pollRef.current)
      pollRef.current = null
    }
    return () => { if (pollRef.current) clearInterval(pollRef.current) }
  }, [hasPending, poll])

  const bgPoll = useCallback(async () => {
    if (!sessionId || hasPending) return
    try {
      const latest = await api.chat.listMessages({ sessionId, limit: PAGE_SIZE, offset: 0 })
      setMessages(prev => {
        const byId = new Map<string, ChatMessage>()
        for (const m of prev) byId.set(m.id, m)
        let changed = false
        for (const m of latest) {
          if (!byId.has(m.id)) changed = true
          byId.set(m.id, m)
        }
        if (!changed && latest.length === prev.length) return prev
        return Array.from(byId.values()).sort((a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime())
      })
    } catch {}
  }, [sessionId, hasPending])

  useEffect(() => {
    bgPollRef.current = setInterval(bgPoll, BG_POLL_INTERVAL)
    return () => { if (bgPollRef.current) clearInterval(bgPollRef.current) }
  }, [bgPoll])

  useEffect(() => {
    setMessages([])
    setHasMore(false)
    setActiveStep(null)
    userScrolledUp.current = false
    if (sessionId) load()
  }, [sessionId])

  useEffect(() => {
    if (prependingRef.current) {
      prependingRef.current = false
      return
    }
    if (!userScrolledUp.current) {
      bottomRef.current?.scrollIntoView({ behavior: 'instant' })
    }
  }, [messages])

  function handleScroll() {
    const el = scrollRef.current
    if (!el) return
    const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 80
    userScrolledUp.current = !atBottom
  }

  async function load() {
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

  async function send() {
    if (!input.trim() || sending || !sessionId) return
    const text = input.trim()
    setInput('')
    setSending(true)
    userScrolledUp.current = false
    try {
      const res = await api.chat.sendMessage(sessionId, text)
      setMessages(prev => [...prev, res.userMessage, res.assistantMessage])
      onFirstMessage?.()
    } catch (e) {
      console.error(e)
    } finally {
      setSending(false)
    }
  }

  if (!sessionId) {
    return (
      <div className="flex items-center justify-center h-full text-zinc-600 text-sm">
        Select a chat or create a new one
      </div>
    )
  }

  return (
    <div className="flex flex-col h-full">
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
            <MessageBubble key={msg.id} msg={msg} onStepClick={setActiveStep} />
          ))}
        </div>
        <div ref={bottomRef} />
      </div>

      <div className="px-6 py-3 border-t border-zinc-800/80 bg-zinc-900/40 shrink-0">
        <div className="flex gap-2">
          <Input
            value={input}
            onChange={e => setInput(e.target.value)}
            onKeyDown={e => e.key === 'Enter' && !e.shiftKey && send()}
            placeholder="Type a message..."
            className="flex-1 bg-zinc-900 border-zinc-800"
            disabled={sending || hasPending}
          />
          <Button
            onClick={send}
            disabled={sending || hasPending || !input.trim()}
          >
            <Send size={16} />
          </Button>
        </div>
      </div>

      {activeStep && <StepPanel step={activeStep} onClose={() => setActiveStep(null)} />}
    </div>
  )
}
