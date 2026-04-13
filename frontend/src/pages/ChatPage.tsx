import { useState, useEffect, useRef, useCallback } from 'react'
import { Send, GitBranch } from 'lucide-react'
import { api } from '../api'
import { navigate } from '../router'
import { MessageBubble, StepPanel } from '../components/ChatMessages'
import type { ChatMessage, Step } from '../types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

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
  const [hasMore, setHasMore] = useState(false)
  const [loadingMore, setLoadingMore] = useState(false)
  const [activeStep, setActiveStep] = useState<Step | null>(null)
  const bottomRef = useRef<HTMLDivElement>(null)
  const scrollRef = useRef<HTMLDivElement>(null)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const prependingRef = useRef(false)
  const userScrolledUp = useRef(false)

  const hasPending = messages.some(m => m.status === 'pending')

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
    userScrolledUp.current = false
    if (sessionId) loadFull()
  }, [sessionId])

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
        <div className="px-6 py-3 border-t border-zinc-200/80 dark:border-zinc-800/80 bg-zinc-50/40 dark:bg-zinc-900/40 shrink-0">
          <div className="flex gap-2">
            <Input
              value={input}
              onChange={e => setInput(e.target.value)}
              onKeyDown={e => e.key === 'Enter' && !e.shiftKey && send()}
              placeholder="Type a message..."
              className="flex-1 bg-white dark:bg-zinc-900 border-zinc-200 dark:border-zinc-800"
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
      )}

      {activeStep && <StepPanel step={activeStep} onClose={() => setActiveStep(null)} />}
    </div>
  )
}
