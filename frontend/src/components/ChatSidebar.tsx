import { useState, useEffect, useRef, useMemo } from 'react'
import { Plus, MessageSquare, GitBranch, Pencil, Trash2, Check, X, ChevronDown, ChevronRight, Loader2 } from '@/lib/icons'
import { api } from '../api'
import { navigate } from '../router'
import type { ChatSession } from '../types'
import { ConfirmDelete } from '@/components/ConfirmDelete'

interface Props {
  activeSessionId: string | null
  onSelect: (session: ChatSession) => void
  onNew: () => void
  refreshKey: number
}

export default function ChatSidebar({ activeSessionId, onSelect, onNew, refreshKey }: Props) {
  const [sessions, setSessions] = useState<ChatSession[]>([])
  const [editingId, setEditingId] = useState<string | null>(null)
  const [editTitle, setEditTitle] = useState('')
  const editRef = useRef<HTMLInputElement>(null)
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null)
  const [planCollapsed, setPlanCollapsed] = useState(false)

  const regularSessions = useMemo(() => sessions.filter(s => s.source !== 'plan'), [sessions])
  const planSessions = useMemo(() => sessions.filter(s => s.source === 'plan'), [sessions])

  useEffect(() => { loadSessions() }, [refreshKey])

  useEffect(() => {
    const iv = setInterval(loadSessions, 5000)
    return () => clearInterval(iv)
  }, [])

  async function loadSessions() {
    try {
      const list = await api.chat.listSessions({ limit: 100 })
      setSessions(list)
    } catch {}
  }

  async function handleDelete(id: string) {
    try {
      await api.chat.deleteSession(id)
      setSessions(prev => prev.filter(s => s.id !== id))
      setDeleteTarget(null)
      if (activeSessionId === id) {
        const remaining = sessions.filter(s => s.id !== id && s.source !== 'plan')
        if (remaining.length > 0) {
          onSelect(remaining[0])
        } else {
          onNew()
        }
      }
    } catch {}
  }

  function startRename(e: React.MouseEvent, session: ChatSession) {
    e.stopPropagation()
    setEditingId(session.id)
    setEditTitle(session.title || '')
    setTimeout(() => editRef.current?.focus(), 0)
  }

  async function confirmRename(id: string) {
    try {
      const updated = await api.chat.updateSession(id, editTitle.trim())
      setSessions(prev => prev.map(s => s.id === id ? updated : s))
    } catch {}
    setEditingId(null)
  }

  function cancelRename() {
    setEditingId(null)
  }

  function displayTitle(s: ChatSession) {
    return s.title || 'New Chat'
  }

  function formatDate(dateStr: string) {
    const d = new Date(dateStr)
    const now = new Date()
    const diff = now.getTime() - d.getTime()
    const days = Math.floor(diff / (1000 * 60 * 60 * 24))
    if (days === 0) return 'Today'
    if (days === 1) return 'Yesterday'
    if (days < 7) return `${days}d ago`
    return d.toLocaleDateString()
  }

  function planIdFromSession(s: ChatSession): string | undefined {
    const m = s.id.match(/^plan:([^:]+):/)
    return m?.[1]
  }

  function renderSession(session: ChatSession, isPlan = false) {
    const Icon = isPlan ? GitBranch : MessageSquare
    const active = activeSessionId === session.id
    return (
      <div
        key={session.id}
        onClick={() => { if (editingId !== session.id) onSelect(session) }}
        data-active={active}
        className={`group relative flex items-center gap-2 pl-3.5 pr-2 py-1.5 cursor-pointer text-[13px] min-w-0 border-l ${
          active
            ? 'border-l-teal-500 text-zinc-900 dark:text-zinc-50'
            : 'border-l-transparent text-zinc-600 dark:text-zinc-400 hover:text-zinc-900 dark:hover:text-zinc-100'
        }`}
      >
        {session.active
          ? <Loader2 size={12} className="shrink-0 text-teal-500 animate-spin" />
          : <Icon size={12} className={`shrink-0 opacity-80 ${isPlan ? 'text-amber-500/70' : ''}`} strokeWidth={1.5} />
        }

        {editingId === session.id ? (
          <div className="flex items-center gap-1 flex-1 min-w-0" onClick={e => e.stopPropagation()}>
            <input
              ref={editRef}
              value={editTitle}
              onChange={e => setEditTitle(e.target.value)}
              onKeyDown={e => {
                if (e.key === 'Enter') confirmRename(session.id)
                if (e.key === 'Escape') cancelRename()
              }}
              className="flex-1 min-w-0 bg-white dark:bg-zinc-800 border border-zinc-300 dark:border-zinc-700 rounded px-1.5 py-0.5 text-xs text-zinc-900 dark:text-zinc-100 focus:outline-none focus:border-teal-500/50"
            />
            <button onClick={() => confirmRename(session.id)} className="text-teal-400 hover:text-teal-300"><Check size={12} /></button>
            <button onClick={cancelRename} className="text-zinc-500 hover:text-zinc-300"><X size={12} /></button>
          </div>
        ) : (
          <>
            <div className="flex-1 min-w-0">
              <div className={`truncate ${active ? 'font-medium' : ''}`}>{displayTitle(session)}</div>
              <div className="flex items-center gap-1.5 mt-0.5">
                <span className="font-mono text-[10px] tabular-nums text-zinc-400 dark:text-zinc-600">{formatDate(session.createdAt).toLowerCase()}</span>
                {isPlan && planIdFromSession(session) && (
                  <button
                    onClick={e => { e.stopPropagation(); navigate({ page: 'plans', planId: planIdFromSession(session)! }) }}
                    className="font-mono text-[10px] text-amber-500/70 hover:text-amber-400 hover:underline"
                  >
                    plan
                  </button>
                )}
              </div>
            </div>
            <div className="hidden group-hover:flex items-center gap-0.5 shrink-0">
              {!isPlan && (
                <button
                  onClick={e => startRename(e, session)}
                  className="p-1 text-zinc-500 dark:text-zinc-600 hover:text-zinc-700 dark:hover:text-zinc-300 rounded"
                >
                  <Pencil size={11} />
                </button>
              )}
              <button
                onClick={e => { e.stopPropagation(); setDeleteTarget(session.id) }}
                className="p-1 text-zinc-500 dark:text-zinc-600 hover:text-red-400 rounded"
              >
                <Trash2 size={11} />
              </button>
            </div>
          </>
        )}
      </div>
    )
  }

  return (
    <div className="flex flex-col h-full">
      <button
        onClick={onNew}
        className="mx-2 mt-1 mb-1.5 flex items-center justify-center gap-1.5 px-2 py-1.5 rounded-[5px] border border-dashed border-zinc-300 dark:border-zinc-700 text-[11.5px] font-mono lowercase text-zinc-500 dark:text-zinc-500 hover:text-teal-600 dark:hover:text-teal-400 hover:border-teal-500/60 transition-colors"
      >
        <Plus size={12} strokeWidth={1.6} />
        new chat
      </button>

      <div className="flex-1 overflow-auto py-0.5">
        <div>
          {regularSessions.map(s => renderSession(s))}
        </div>

        {regularSessions.length === 0 && planSessions.length === 0 && (
          <div className="text-center font-mono lowercase text-zinc-500 dark:text-zinc-600 text-[11px] py-6">
            — no chats yet —
          </div>
        )}

        {planSessions.length > 0 && (
          <div className="mt-2">
            <button
              onClick={() => setPlanCollapsed(v => !v)}
              className="kicker px-3.5 pt-2 pb-1 w-full text-left hover:text-zinc-700 dark:hover:text-zinc-400 transition-colors"
            >
              {planCollapsed ? <ChevronRight size={9} /> : <ChevronDown size={9} />}
              <span className="kicker-num">{String(planSessions.length).padStart(2, '0')}</span>
              <span className="kicker-sep">/</span>
              <span>plans</span>
            </button>
            {!planCollapsed && (
              <div>
                {planSessions.map(s => renderSession(s, true))}
              </div>
            )}
          </div>
        )}
      </div>

      <ConfirmDelete
        open={!!deleteTarget}
        onCancel={() => setDeleteTarget(null)}
        onConfirm={() => deleteTarget && handleDelete(deleteTarget)}
        title="Delete chat?"
        description="This will delete the chat and all its messages."
      />
    </div>
  )
}
