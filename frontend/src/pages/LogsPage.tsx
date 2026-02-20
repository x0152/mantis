import { useState, useEffect, useCallback } from 'react'
import { ScrollText, Terminal, ChevronRight, ChevronDown, Clock, CheckCircle2, Loader2, Filter, Trash2 } from 'lucide-react'
import { api } from '../api'
import { EntryLine, PromptBanner } from '../components/LogEntries'
import type { SessionLog, Connection } from '../types'

const PAGE_SIZE = 10

function timeAgo(date: string) {
  const s = Math.floor((Date.now() - new Date(date).getTime()) / 1000)
  if (s < 60) return `${s}s ago`
  if (s < 3600) return `${Math.floor(s / 60)}m ago`
  if (s < 86400) return `${Math.floor(s / 3600)}h ago`
  return `${Math.floor(s / 86400)}d ago`
}

function duration(start: string, end?: string) {
  const ms = (end ? new Date(end).getTime() : Date.now()) - new Date(start).getTime()
  if (ms < 1000) return `${ms}ms`
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`
  return `${Math.floor(ms / 60000)}m ${Math.floor((ms % 60000) / 1000)}s`
}

function SessionCard({ log, expanded, onToggle }: { log: SessionLog; expanded: boolean; onToggle: () => void }) {
  const isRunning = log.status === 'running'

  return (
    <div className="bg-zinc-900 rounded-lg border border-zinc-800 overflow-hidden">
      <div
        className="flex items-center gap-3 px-4 py-3.5 cursor-pointer hover:bg-zinc-800/40"
        onClick={onToggle}
      >
        <div className="text-zinc-600">
          {expanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
        </div>
        <div className={`p-1.5 rounded-md ${isRunning ? 'bg-amber-500/10' : 'bg-emerald-500/10'}`}>
          {isRunning
            ? <Loader2 size={14} className="text-amber-400 animate-spin" />
            : <CheckCircle2 size={14} className="text-emerald-400" />
          }
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-medium text-zinc-200 text-sm">{log.agentName}</span>
            <span className={`px-1.5 py-0.5 text-[11px] font-medium rounded-full ${
              isRunning ? 'bg-amber-500/15 text-amber-400' : 'bg-emerald-500/15 text-emerald-400'
            }`}>
              {log.status}
            </span>
          </div>
          {log.prompt && (
            <p className="text-xs text-zinc-500 mt-0.5 truncate">{log.prompt}</p>
          )}
          <div className="flex items-center gap-3 mt-0.5 text-[11px] text-zinc-600">
            <span className="flex items-center gap-1"><Clock size={10} />{timeAgo(log.startedAt)}</span>
            <span>{duration(log.startedAt, log.finishedAt)}</span>
            <span>{log.entries.length} entries</span>
          </div>
        </div>
        <span className="font-mono text-[11px] text-zinc-700">{log.id.slice(0, 8)}</span>
      </div>

      {expanded && (
        <div className="border-t border-zinc-800 bg-zinc-950 px-4 py-3.5 space-y-1 max-h-[500px] overflow-y-auto">
          {log.prompt && <PromptBanner prompt={log.prompt} />}
          {log.entries.length === 0 && !log.prompt ? (
            <p className="text-zinc-600 text-xs font-mono">No entries yet</p>
          ) : (
            log.entries.map((entry, i) => <EntryLine key={i} entry={entry} />)
          )}
        </div>
      )}
    </div>
  )
}

export default function LogsPage() {
  const [logs, setLogs] = useState<SessionLog[]>([])
  const [connections, setConnections] = useState<Connection[]>([])
  const [selectedConn, setSelectedConn] = useState('')
  const [loading, setLoading] = useState(true)
  const [hasMore, setHasMore] = useState(false)
  const [loadingMore, setLoadingMore] = useState(false)
  const [expanded, setExpanded] = useState<string | null>(null)

  const load = useCallback(async () => {
    try {
      setLoading(true)
      const [l, c] = await Promise.all([
        api.sessionLogs.list({ connectionId: selectedConn || undefined, limit: PAGE_SIZE, offset: 0 }),
        api.connections.list(),
      ])
      setLogs(l)
      setHasMore(l.length === PAGE_SIZE)
      setConnections(c)
    } catch {
      setLogs([])
      setHasMore(false)
    } finally {
      setLoading(false)
    }
  }, [selectedConn])

  useEffect(() => { load() }, [load])

  async function loadMoreLogs() {
    if (loadingMore || !hasMore) return
    setLoadingMore(true)
    try {
      const offset = logs.length
      const more = await api.sessionLogs.list({ connectionId: selectedConn || undefined, limit: PAGE_SIZE, offset })
      if (more.length < PAGE_SIZE) setHasMore(false)
      setLogs(prev => {
        const combined = [...prev, ...more]
        const seen = new Set<string>()
        const out: SessionLog[] = []
        for (const item of combined) {
          if (seen.has(item.id)) continue
          seen.add(item.id)
          out.push(item)
        }
        return out
      })
    } catch {
      setHasMore(false)
    } finally {
      setLoadingMore(false)
    }
  }

  useEffect(() => {
    const interval = setInterval(async () => {
      try {
        const l = await api.sessionLogs.list({ connectionId: selectedConn || undefined, limit: PAGE_SIZE, offset: 0 })
        setLogs(prev => {
          const seen = new Set(l.map(x => x.id))
          const rest = prev.filter(x => !seen.has(x.id))
          return [...l, ...rest]
        })
      } catch { /* noop */ }
    }, 5000)
    return () => clearInterval(interval)
  }, [selectedConn])

  const connName = (id: string) => connections.find(c => c.id === id)?.name ?? id.slice(0, 8)

  return (
    <div className="p-6 max-w-5xl">
      <div className="flex items-center justify-between mb-5">
        <div>
          <h1 className="text-lg font-semibold text-zinc-100">Session Logs</h1>
          <p className="text-xs text-zinc-600 mt-0.5">Execution history per connection</p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={async () => {
              if (!confirm('Clear all session logs? This cannot be undone.')) return
              await api.sessionLogs.clear()
              setLogs([])
              setHasMore(false)
            }}
            className="flex items-center gap-1.5 px-2.5 py-1.5 text-xs font-medium text-red-400/70 bg-zinc-800 rounded-lg hover:text-red-300 hover:bg-zinc-700"
          >
            <Trash2 size={13} />
            Clear all
          </button>
          <Filter size={14} className="text-zinc-600" />
          <select
            value={selectedConn}
            onChange={e => { setSelectedConn(e.target.value); setExpanded(null) }}
            className="px-2.5 py-1.5 border border-zinc-800 rounded-lg text-xs bg-zinc-900 text-zinc-400 focus:outline-none focus:border-teal-500/50"
          >
            <option value="">All connections</option>
            {connections.map(c => (
              <option key={c.id} value={c.id}>{c.name}</option>
            ))}
          </select>
        </div>
      </div>

      {loading ? (
        <div className="text-center py-12 text-zinc-600 text-sm">Loading...</div>
      ) : logs.length === 0 ? (
        <div className="text-center py-20">
          <ScrollText size={40} className="mx-auto text-zinc-700 mb-3" />
          <p className="text-zinc-400 text-sm font-medium">No session logs yet</p>
          <p className="text-xs text-zinc-600 mt-1">Logs will appear here as agents execute tasks</p>
        </div>
      ) : (
        <div className="space-y-2.5">
          {logs.map(log => (
            <div key={log.id}>
              {selectedConn === '' && (
                <p className="text-[11px] text-zinc-600 mb-1 ml-1 flex items-center gap-1.5">
                  <Terminal size={10} />
                  {connName(log.connectionId)}
                </p>
              )}
              <SessionCard
                log={log}
                expanded={expanded === log.id}
                onToggle={() => setExpanded(prev => prev === log.id ? null : log.id)}
              />
            </div>
          ))}
          {hasMore && (
            <div className="flex justify-center pt-2">
              <button
                onClick={loadMoreLogs}
                disabled={loadingMore}
                className="px-3 py-1.5 text-xs font-medium text-zinc-500 bg-zinc-800 rounded-lg hover:text-zinc-300 hover:bg-zinc-700 disabled:opacity-30 disabled:cursor-not-allowed"
              >
                {loadingMore ? 'Loading...' : `Load more ${PAGE_SIZE}`}
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
