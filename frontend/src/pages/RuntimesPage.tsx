import { useEffect, useState, useCallback } from 'react'
import { Container, Play, Square, RotateCw, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { api } from '../api'
import type { SandboxStatus } from '../types'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { EmptyState } from '@/components/EmptyState'
import { ConfirmDelete } from '@/components/ConfirmDelete'

const STATE_STYLES: Record<string, string> = {
  running: 'bg-teal-500/15 text-teal-700 dark:text-teal-300 border-teal-500/30',
  exited: 'bg-zinc-500/15 text-zinc-600 dark:text-zinc-400 border-zinc-500/30',
  stopped: 'bg-zinc-500/15 text-zinc-600 dark:text-zinc-400 border-zinc-500/30',
  'not-built': 'bg-amber-500/15 text-amber-700 dark:text-amber-300 border-amber-500/30',
  created: 'bg-sky-500/15 text-sky-700 dark:text-sky-300 border-sky-500/30',
  paused: 'bg-violet-500/15 text-violet-700 dark:text-violet-300 border-violet-500/30',
  restarting: 'bg-amber-500/15 text-amber-700 dark:text-amber-300 border-amber-500/30',
  dead: 'bg-red-500/15 text-red-700 dark:text-red-300 border-red-500/30',
}

const stateClass = (state: string) => STATE_STYLES[state] ?? 'bg-zinc-500/15 text-zinc-700 dark:text-zinc-300 border-zinc-500/30'

function sandboxName(connName: string) {
  return connName.startsWith('sb-') ? connName.slice(3) : connName
}

function relativeTime(iso: string) {
  if (!iso || iso.startsWith('0001')) return '—'
  const d = new Date(iso).getTime()
  if (!d) return '—'
  const diff = Date.now() - d
  const sec = Math.max(0, Math.floor(diff / 1000))
  if (sec < 60) return `${sec}s ago`
  const min = Math.floor(sec / 60)
  if (min < 60) return `${min}m ago`
  const hrs = Math.floor(min / 60)
  if (hrs < 48) return `${hrs}h ago`
  return `${Math.floor(hrs / 24)}d ago`
}

export default function RuntimesPage() {
  const [items, setItems] = useState<SandboxStatus[]>([])
  const [loading, setLoading] = useState(true)
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null)
  const [busy, setBusy] = useState<string | null>(null)

  const load = useCallback(async () => {
    try {
      setLoading(true)
      const list = await api.sandboxes.list()
      setItems(list.sort((a, b) => a.connection.name.localeCompare(b.connection.name)))
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Failed to load sandboxes')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
    const interval = setInterval(load, 5000)
    return () => clearInterval(interval)
  }, [load])

  const rebuild = async (name: string) => {
    setBusy(name)
    try {
      const res = await api.sandboxes.rebuild(name)
      if (!res.ok) throw new Error(`Rebuild failed: ${res.status}`)
      toast.success(`Rebuilding ${name} — follow logs in Docker logs`)
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Rebuild failed')
    } finally {
      setBusy(null)
      load()
    }
  }
  const start = async (name: string) => {
    setBusy(name)
    try { await api.sandboxes.start(name); toast.success(`Starting ${name}`) } catch (e: unknown) { toast.error(e instanceof Error ? e.message : 'Start failed') }
    setBusy(null); load()
  }
  const stop = async (name: string) => {
    setBusy(name)
    try { await api.sandboxes.stop(name); toast.success(`Stopped ${name}`) } catch (e: unknown) { toast.error(e instanceof Error ? e.message : 'Stop failed') }
    setBusy(null); load()
  }
  const remove = async (name: string) => {
    setBusy(name)
    try { await api.sandboxes.delete(name); toast.success(`Removed ${name}`) } catch (e: unknown) { toast.error(e instanceof Error ? e.message : 'Remove failed') }
    setDeleteTarget(null); setBusy(null); load()
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-5">
        <div>
          <h1 className="text-lg font-semibold text-zinc-900 dark:text-zinc-100">Runtimes</h1>
          <p className="text-xs text-zinc-500 dark:text-zinc-600 mt-0.5">Sandboxes built from Dockerfile records. Auto-refreshes every 5s.</p>
        </div>
        <Button size="sm" variant="secondary" onClick={load} disabled={loading}>
          <RotateCw size={14} className={loading ? 'animate-spin' : ''} /> Refresh
        </Button>
      </div>

      {loading && items.length === 0 ? (
        <div className="text-center py-12 text-zinc-500 dark:text-zinc-600 text-sm">Loading...</div>
      ) : items.length === 0 ? (
        <EmptyState icon={Container} title="No sandboxes yet" description="Sandboxes appear here once a Dockerfile is registered." />
      ) : (
        <div className="space-y-2">
          {items.map(item => {
            const name = sandboxName(item.connection.name)
            const isBusy = busy === name
            const isRunning = item.state === 'running'
            const isBuilt = item.state !== 'not-built'
            return (
              <div key={item.connection.id} className="bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800 px-4 py-3 flex items-center gap-4">
                <Container size={18} className="text-zinc-500 dark:text-zinc-400 shrink-0" />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="font-medium text-zinc-800 dark:text-zinc-200 text-sm">{name}</span>
                    <Badge className={`border ${stateClass(item.state)}`}>{item.state}</Badge>
                    {item.connection.profileIds?.[0] && (
                      <Badge variant="secondary">{item.connection.profileIds[0]}</Badge>
                    )}
                  </div>
                  <div className="mt-1 flex flex-wrap items-center gap-x-4 gap-y-0.5 text-[11px] text-zinc-500 dark:text-zinc-500">
                    <span>ssh: <span className="font-mono text-zinc-600 dark:text-zinc-400">{item.connection.name}</span></span>
                    {item.container.host && <span>host: <span className="font-mono">{item.container.host}</span></span>}
                    {item.container.image && <span>image: <span className="font-mono">{item.container.image}</span></span>}
                    {isBuilt && <span>started: {relativeTime(item.container.createdAt)}</span>}
                  </div>
                  {item.connection.description && (
                    <p className="mt-1 text-[11px] text-zinc-500 dark:text-zinc-500 line-clamp-1">{item.connection.description}</p>
                  )}
                </div>
                <div className="flex gap-1 shrink-0">
                  {isRunning ? (
                    <Button variant="ghost" size="icon" disabled={isBusy} onClick={() => stop(name)} title="Stop">
                      <Square size={14} />
                    </Button>
                  ) : (
                    <Button variant="ghost" size="icon" disabled={isBusy || !isBuilt} onClick={() => start(name)} title="Start">
                      <Play size={14} />
                    </Button>
                  )}
                  <Button variant="ghost" size="icon" disabled={isBusy} onClick={() => rebuild(name)} title="Rebuild">
                    <RotateCw size={14} className={isBusy ? 'animate-spin' : ''} />
                  </Button>
                  <Button variant="destructive" size="icon" disabled={isBusy} onClick={() => setDeleteTarget(name)} title="Remove">
                    <Trash2 size={14} />
                  </Button>
                </div>
              </div>
            )
          })}
        </div>
      )}

      <ConfirmDelete
        open={!!deleteTarget}
        onCancel={() => setDeleteTarget(null)}
        onConfirm={() => deleteTarget && remove(deleteTarget)}
        title="Remove sandbox?"
        description="Stops and removes the container and deletes the SSH connection. The Dockerfile record is deleted too."
      />
    </div>
  )
}
