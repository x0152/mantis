import { useState, useEffect, useRef } from 'react'
import { Plus, MessageSquare, Pencil, Trash2, Check, X } from 'lucide-react'
import { api } from '../api'
import type { ChatSession } from '../types'
import { Button } from '@/components/ui/button'
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

  useEffect(() => { loadSessions() }, [refreshKey])

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
        const remaining = sessions.filter(s => s.id !== id)
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

  return (
    <div className="flex flex-col h-full">
      <Button
        variant="secondary"
        size="sm"
        onClick={onNew}
        className="mx-2 mt-2 mb-1"
      >
        <Plus size={14} />
        New Chat
      </Button>

      <div className="flex-1 overflow-auto px-2 py-1 space-y-0.5">
        {sessions.map(session => (
          <div
            key={session.id}
            onClick={() => { if (editingId !== session.id) onSelect(session) }}
            className={`group flex items-center gap-2 px-3 py-2 rounded-lg cursor-pointer text-[13px] min-w-0 ${
              activeSessionId === session.id
                ? 'bg-teal-500/10 text-teal-400'
                : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/50'
            }`}
          >
            <MessageSquare size={14} className="shrink-0" strokeWidth={1.8} />

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
                  className="flex-1 min-w-0 bg-zinc-800 border border-zinc-700 rounded px-1.5 py-0.5 text-xs text-zinc-100 focus:outline-none focus:border-teal-500/50"
                />
                <button onClick={() => confirmRename(session.id)} className="text-teal-400 hover:text-teal-300"><Check size={12} /></button>
                <button onClick={cancelRename} className="text-zinc-500 hover:text-zinc-300"><X size={12} /></button>
              </div>
            ) : (
              <>
                <div className="flex-1 min-w-0">
                  <div className="truncate font-medium">{displayTitle(session)}</div>
                  <div className="text-[10px] text-zinc-600 mt-0.5">{formatDate(session.createdAt)}</div>
                </div>
                <div className="hidden group-hover:flex items-center gap-0.5 shrink-0">
                  <button
                    onClick={e => startRename(e, session)}
                    className="p-1 text-zinc-600 hover:text-zinc-300 rounded"
                  >
                    <Pencil size={11} />
                  </button>
                  <button
                    onClick={e => { e.stopPropagation(); setDeleteTarget(session.id) }}
                    className="p-1 text-zinc-600 hover:text-red-400 rounded"
                  >
                    <Trash2 size={11} />
                  </button>
                </div>
              </>
            )}
          </div>
        ))}

        {sessions.length === 0 && (
          <div className="text-center text-zinc-600 text-xs py-6">
            No chats yet
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
