import { useState, useEffect, useRef, useCallback } from 'react'
import { Send, RotateCcw, Trash2 } from 'lucide-react'
import { api } from '../api'
import { MessageBubble, StepPanel } from '../components/ChatMessages'
import type { ChatMessage, Step } from '../types'

const PAGE_SIZE = 10
const POLL_INTERVAL = 300

export default function ChatPage() {
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [sessionId, setSessionId] = useState('')
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

  const hasPending = messages.some(m => m.sessionId === sessionId && m.status === 'pending')

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

  useEffect(() => { load() }, [])
  useEffect(() => {
    if (prependingRef.current) {
      prependingRef.current = false
      return
    }
    if (!userScrolledUp.current) {
      bottomRef.current?.scrollIntoView({ behavior: 'instant' })
    }
  }, [messages, sessionId])

  function handleScroll() {
    const el = scrollRef.current
    if (!el) return
    const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 80
    userScrolledUp.current = !atBottom
  }

  async function load() {
    const [session, msgs] = await Promise.all([
      api.chat.getSession(),
      api.chat.listMessages({ limit: PAGE_SIZE, offset: 0 }),
    ])
    setSessionId(session.id)
    setMessages(msgs)
    setHasMore(msgs.length === PAGE_SIZE)
  }

  async function loadMoreMessages() {
    if (loadingMore || !hasMore) return
    setLoadingMore(true)
    try {
      const offset = messages.length
      const older = await api.chat.listMessages({ limit: PAGE_SIZE, offset })
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
    if (!input.trim() || sending) return
    const text = input.trim()
    setInput('')
    setSending(true)
    userScrolledUp.current = false
    try {
      const res = await api.chat.sendMessage(sessionId, text)
      setMessages(prev => [...prev, res.userMessage, res.assistantMessage])
    } catch (e) {
      console.error(e)
    } finally {
      setSending(false)
    }
  }

  async function resetContext() {
    const session = await api.chat.resetContext()
    setSessionId(session.id)
  }

  async function clearHistory() {
    if (!confirm('Clear all chat history? This cannot be undone.')) return
    await api.chat.clearHistory()
    setMessages([])
    setHasMore(false)
    const session = await api.chat.getSession()
    setSessionId(session.id)
  }

  const grouped = groupBySession(messages)
  const pendingReset = grouped.length > 0 && grouped[grouped.length - 1].sessionId !== sessionId

  return (
    <div className="flex flex-col h-full">
      <div className="px-6 py-4 border-b border-zinc-800/80 bg-zinc-900/40 flex items-center justify-between shrink-0">
        <div>
          <h2 className="text-sm font-semibold text-zinc-100">Chat</h2>
          <p className="text-[11px] text-zinc-600 mt-0.5">Interactive session</p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={clearHistory}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-red-400/70 bg-zinc-800 rounded-lg hover:text-red-300 hover:bg-zinc-700"
          >
            <Trash2 size={13} />
            Clear all
          </button>
          <button
            onClick={resetContext}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-zinc-500 bg-zinc-800 rounded-lg hover:text-zinc-300 hover:bg-zinc-700"
          >
            <RotateCcw size={13} />
            Reset
          </button>
        </div>
      </div>

      <div ref={scrollRef} onScroll={handleScroll} className="flex-1 overflow-auto px-6 py-5 space-y-1">
        {hasMore && messages.length > 0 && (
          <div className="flex justify-center mb-4">
            <button
              onClick={loadMoreMessages}
              disabled={loadingMore}
              className="px-3 py-1.5 text-xs font-medium text-zinc-500 bg-zinc-800 rounded-lg hover:text-zinc-300 hover:bg-zinc-700 disabled:opacity-30 disabled:cursor-not-allowed"
            >
              {loadingMore ? 'Loading...' : `Load previous ${PAGE_SIZE}`}
            </button>
          </div>
        )}
        {grouped.length === 0 && !pendingReset && (
          <div className="flex items-center justify-center h-full text-zinc-600 text-sm">
            Send a message to start the conversation
          </div>
        )}
        {grouped.map((group, gi) => (
          <div key={group.sessionId}>
            {gi > 0 && <ResetDivider />}
            <div className="space-y-3">
              {group.messages.map(msg => (
                <MessageBubble key={msg.id} msg={msg} onStepClick={setActiveStep} />
              ))}
            </div>
          </div>
        ))}
        {pendingReset && <ResetDivider />}
        <div ref={bottomRef} />
      </div>

      <div className="px-6 py-3 border-t border-zinc-800/80 bg-zinc-900/40 shrink-0">
        <div className="flex gap-2">
          <input
            value={input}
            onChange={e => setInput(e.target.value)}
            onKeyDown={e => e.key === 'Enter' && !e.shiftKey && send()}
            placeholder="Type a message..."
            className="flex-1 px-3.5 py-2.5 bg-zinc-900 border border-zinc-800 rounded-lg text-sm text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-teal-500/50"
            disabled={sending || hasPending}
          />
          <button
            onClick={send}
            disabled={sending || hasPending || !input.trim()}
            className="px-3.5 py-2.5 bg-teal-600 text-white rounded-lg hover:bg-teal-500 disabled:opacity-30 disabled:cursor-not-allowed"
          >
            <Send size={16} />
          </button>
        </div>
      </div>

      {activeStep && <StepPanel step={activeStep} onClose={() => setActiveStep(null)} />}
    </div>
  )
}

function ResetDivider() {
  return (
    <div className="flex items-center gap-3 my-5">
      <div className="flex-1 h-px bg-zinc-800" />
      <span className="text-[11px] text-zinc-600 font-medium px-2">Context Reset</span>
      <div className="flex-1 h-px bg-zinc-800" />
    </div>
  )
}

function groupBySession(messages: ChatMessage[]) {
  const groups: { sessionId: string; messages: ChatMessage[] }[] = []
  for (const msg of messages) {
    const last = groups[groups.length - 1]
    if (last && last.sessionId === msg.sessionId) {
      last.messages.push(msg)
    } else {
      groups.push({ sessionId: msg.sessionId, messages: [msg] })
    }
  }
  return groups
}
