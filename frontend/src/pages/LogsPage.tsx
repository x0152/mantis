import { useState, useEffect, useCallback } from 'react'
import { ScrollText, Terminal, ChevronRight, ChevronDown, Clock, CheckCircle2, Loader2, Filter, Trash2 } from '@/lib/icons'
import { api } from '../api'
import { EntryLine, PromptBanner } from '../components/LogEntries'
import type { SessionLog, Connection } from '../types'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { EmptyState } from '@/components/EmptyState'
import { ConfirmDelete } from '@/components/ConfirmDelete'

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
    <div className="bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800 overflow-hidden">
      <div
        className="flex items-center gap-3 px-4 py-3.5 cursor-pointer hover:bg-zinc-50 dark:hover:bg-zinc-800/40"
        onClick={onToggle}
      >
        <div className="text-zinc-400 dark:text-zinc-600">
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
            <span className="font-medium text-zinc-800 dark:text-zinc-200 text-sm">{log.agentName}</span>
            <Badge variant={isRunning ? 'warning' : 'success'}>
              {log.status}
            </Badge>
          </div>
          {log.prompt && (
            <p className="text-xs text-zinc-500 mt-0.5 truncate">{log.prompt}</p>
          )}
          {(log.presetName || log.modelName || log.modelRole === 'fallback') && (
            <div className="flex items-center gap-1.5 mt-1 flex-wrap">
              {log.presetName && (
                <span className="px-1.5 py-0.5 text-[10px] rounded-full bg-zinc-200 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-500">
                  {log.presetName}
                </span>
              )}
              {log.modelName && (
                <span className="px-1.5 py-0.5 text-[10px] rounded-full bg-zinc-200 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-500">
                  {log.modelName}
                </span>
              )}
              {log.modelRole === 'fallback' && (
                <span className="px-1.5 py-0.5 text-[10px] rounded-full bg-amber-500/15 text-amber-400">
                  fallback
                </span>
              )}
            </div>
          )}
          <div className="flex items-center gap-3 mt-0.5 text-[11px] text-zinc-500 dark:text-zinc-600">
            <span className="flex items-center gap-1"><Clock size={10} />{timeAgo(log.startedAt)}</span>
            <span>{duration(log.startedAt, log.finishedAt)}</span>
            <span>{log.entries.length} entries</span>
          </div>
        </div>
        <span className="font-mono text-[11px] text-zinc-400 dark:text-zinc-700">{log.id.slice(0, 8)}</span>
      </div>

      {expanded && (
        <div className="border-t border-zinc-200 dark:border-zinc-800 bg-zinc-50 dark:bg-zinc-950 px-4 py-3.5 space-y-1 max-h-125 overflow-y-auto">
          {log.prompt && <PromptBanner prompt={log.prompt} />}
          {log.entries.length === 0 && !log.prompt ? (
            <p className="text-zinc-500 dark:text-zinc-600 text-xs font-mono">No entries yet</p>
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
  const [clearOpen, setClearOpen] = useState(false)

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
    <div className="p-6">
      <div className="flex items-center justify-between mb-5">
        <div>
          <h1 className="text-lg font-semibold text-zinc-900 dark:text-zinc-100">Session Logs</h1>
          <p className="text-xs text-zinc-500 dark:text-zinc-600 mt-0.5">Execution history per connection</p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="destructive"
            size="sm"
            onClick={() => setClearOpen(true)}
          >
            <Trash2 size={13} />
            Clear all
          </Button>
          <Filter size={14} className="text-zinc-600" />
          <select
            value={selectedConn}
            onChange={e => { setSelectedConn(e.target.value); setExpanded(null) }}
            className="px-2.5 py-1.5 border border-zinc-200 dark:border-zinc-800 rounded-lg text-xs bg-white dark:bg-zinc-900 text-zinc-600 dark:text-zinc-400 focus:outline-none focus:border-teal-500/50"
          >
            <option value="">All connections</option>
            {connections.map(c => (
              <option key={c.id} value={c.id}>{c.name}</option>
            ))}
          </select>
        </div>
      </div>

      {loading ? (
        <div className="text-center py-12 text-zinc-500 dark:text-zinc-600 text-sm">Loading...</div>
      ) : logs.length === 0 ? (
        <EmptyState icon={ScrollText} title="No session logs yet" description="Logs will appear here as agents execute tasks" />
      ) : (
        <div className="space-y-2.5">
          {logs.map(log => (
            <div key={log.id}>
              {selectedConn === '' && (
                <p className="text-[11px] text-zinc-500 dark:text-zinc-600 mb-1 ml-1 flex items-center gap-1.5">
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
              <Button variant="secondary" size="sm" onClick={loadMoreLogs} disabled={loadingMore}>
                {loadingMore ? 'Loading...' : `Load more ${PAGE_SIZE}`}
              </Button>
            </div>
          )}
        </div>
      )}

      <ConfirmDelete
        open={clearOpen}
        onCancel={() => setClearOpen(false)}
        onConfirm={async () => {
          await api.sessionLogs.clear()
          setLogs([])
          setHasMore(false)
          setClearOpen(false)
        }}
        title="Clear all session logs?"
        description="This cannot be undone."
      />
    </div>
  )
}
