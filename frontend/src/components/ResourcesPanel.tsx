import { useEffect, useMemo, useRef, useState } from 'react'
import { Container, Cloud, Send, MessageSquare, Plug, ChevronRight, Loader2 } from '@/lib/icons'
import { api } from '../api'
import type { Connection, SandboxStatus, Step } from '../types'

interface Props {
  activeSteps?: Step[]
}

type SandboxByConnId = Record<string, SandboxStatus>

const ACTIVE_POLL = 3500
const IDLE_POLL = 9000

export function ResourcesPanel({ activeSteps }: Props) {
  const [connections, setConnections] = useState<Connection[]>([])
  const [sandboxes, setSandboxes] = useState<SandboxByConnId>({})
  const [openConn, setOpenConn] = useState<string | null>(null)
  const [loaded, setLoaded] = useState(false)
  const prevIdsRef = useRef<Set<string>>(new Set())
  const [recentlyAdded, setRecentlyAdded] = useState<Set<string>>(new Set())

  const hasActive = (activeSteps?.length ?? 0) > 0

  const activeConnIdSet = useMemo(() => {
    const s = new Set<string>()
    if (!activeSteps || activeSteps.length === 0) return s
    const byName = new Map(connections.map(c => [c.name.toLowerCase(), c.id]))
    const ids = new Set(connections.map(c => c.id))
    for (const st of activeSteps) {
      if (st.status !== 'running') continue
      let parsed: Record<string, unknown> | null = null
      try { parsed = JSON.parse(st.args || '{}') } catch { parsed = null }
      const candidates: string[] = []
      if (parsed) {
        for (const key of ['connection', 'connectionId', 'connection_id', 'connectionID', 'server', 'host', 'name']) {
          const v = parsed[key]
          if (typeof v === 'string') candidates.push(v)
        }
      }
      for (const c of candidates) {
        const lower = c.toLowerCase()
        if (ids.has(c)) s.add(c)
        else if (byName.has(lower)) s.add(byName.get(lower)!)
      }
      if (candidates.length === 0) {
        const tool = st.tool.toLowerCase()
        for (const conn of connections) {
          if (tool.includes(conn.name.toLowerCase())) s.add(conn.id)
        }
      }
    }
    return s
  }, [activeSteps, connections])

  useEffect(() => {
    let cancelled = false
    const refresh = async () => {
      try {
        const [conns, sbList] = await Promise.all([
          api.connections.list(),
          api.sandboxes.list().catch(() => [] as SandboxStatus[]),
        ])
        if (cancelled) return
        const sbMap: SandboxByConnId = {}
        for (const sb of sbList) sbMap[sb.connection.id] = sb
        if (loaded && prevIdsRef.current.size > 0) {
          const newOnes = new Set<string>()
          for (const c of conns) {
            if (!prevIdsRef.current.has(c.id)) newOnes.add(c.id)
          }
          if (newOnes.size > 0) {
            setRecentlyAdded(newOnes)
            setTimeout(() => setRecentlyAdded(new Set()), 4000)
          }
        }
        prevIdsRef.current = new Set(conns.map(c => c.id))
        setConnections(conns)
        setSandboxes(sbMap)
        setLoaded(true)
      } catch {}
    }
    refresh()
    const interval = hasActive ? ACTIVE_POLL : IDLE_POLL
    const id = setInterval(refresh, interval)
    return () => { cancelled = true; clearInterval(id) }
  }, [hasActive, loaded])

  if (!loaded || connections.length === 0) {
    return null
  }

  const newCount = recentlyAdded.size

  return (
    <aside className="hidden md:flex w-60 shrink-0 flex-col h-full border-l border-zinc-200/80 dark:border-zinc-800/60 bg-zinc-50/50 dark:bg-zinc-900/30">
      <div className="flex items-center justify-between px-3 py-2.5 border-b border-zinc-200/70 dark:border-zinc-800/60 shrink-0">
        <span className="kicker">
          <span className="kicker-num tabular-nums">{String(connections.length).padStart(2, '0')}</span>
          <span className="kicker-sep">/</span>
          <span>resources</span>
        </span>
        {newCount > 0 && (
          <span className="font-mono text-[10px] lowercase text-teal-600 dark:text-teal-400">
            +{newCount} new
          </span>
        )}
      </div>
      <div className="flex-1 overflow-auto px-1.5 py-1.5 space-y-0.5 min-h-0">
        {connections.map(c => (
          <ResourceRow
            key={c.id}
            conn={c}
            sandbox={sandboxes[c.id]}
            active={activeConnIdSet.has(c.id)}
            isNew={recentlyAdded.has(c.id)}
            isOpen={openConn === c.id}
            onToggle={() => setOpenConn(prev => prev === c.id ? null : c.id)}
          />
        ))}
      </div>
    </aside>
  )
}

function typeIcon(type: string, isSandbox: boolean) {
  if (isSandbox) return Container
  if (type === 'ssh') return Cloud
  if (type === 'telegram') return Send
  if (type === 'chat') return MessageSquare
  return Plug
}

function connectionTone(conn: Connection, sandbox: SandboxStatus | undefined): SandboxStateMeta {
  const isSandbox = !!conn.dockerfile || !!sandbox
  if (isSandbox) return sandboxMeta(sandbox)
  if (conn.type === 'ssh') return { label: 'remote', tone: 'ok' }
  if (conn.type === 'telegram') return { label: 'telegram', tone: 'ok' }
  if (conn.type === 'chat') return { label: 'chat', tone: 'idle' }
  return { label: conn.type, tone: 'idle' }
}

interface SandboxStateMeta {
  label: string
  tone: 'ok' | 'warn' | 'err' | 'idle'
}

function sandboxMeta(sb: SandboxStatus | undefined): SandboxStateMeta {
  if (!sb) return { label: '—', tone: 'idle' }
  const state = (sb.state || '').toLowerCase()
  const containerStatus = (sb.container?.status || '').toLowerCase()
  if (state === 'running' || containerStatus.includes('up') || containerStatus.includes('running')) {
    return { label: 'running', tone: 'ok' }
  }
  if (state === 'starting' || state === 'building' || containerStatus.includes('start')) {
    return { label: state || 'starting', tone: 'warn' }
  }
  if (state === 'error' || state === 'failed' || containerStatus.includes('exit') || containerStatus.includes('dead')) {
    return { label: state || 'failed', tone: 'err' }
  }
  if (state === 'stopped' || containerStatus.includes('stop') || containerStatus.includes('paused')) {
    return { label: 'stopped', tone: 'idle' }
  }
  return { label: state || containerStatus || 'unknown', tone: 'idle' }
}

function iconToneClass(tone: SandboxStateMeta['tone']) {
  return `icon-status tone-${tone}`
}

function toneText(tone: SandboxStateMeta['tone']) {
  switch (tone) {
    case 'ok': return 'text-teal-600 dark:text-teal-400'
    case 'warn': return 'text-amber-600 dark:text-amber-400'
    case 'err': return 'text-rose-500 dark:text-rose-400'
    case 'idle':
    default: return 'text-zinc-500 dark:text-zinc-500'
  }
}

interface RowProps {
  conn: Connection
  sandbox?: SandboxStatus
  active: boolean
  isNew: boolean
  isOpen: boolean
  onToggle: () => void
}

function ResourceRow({ conn, sandbox, active, isNew, isOpen, onToggle }: RowProps) {
  const isSandbox = !!conn.dockerfile || !!sandbox
  const Icon = typeIcon(conn.type, isSandbox)
  const meta = connectionTone(conn, sandbox)
  const description = (conn.description || '').trim()

  return (
    <div
      data-new={isNew}
      data-active={active}
      className={`relative rounded-[5px] border transition-colors ${
        active
          ? 'border-violet-500/50 bg-violet-500/[0.07] shadow-[inset_2px_0_0_rgb(139_92_246)]'
          : isNew
            ? 'border-teal-500/30 bg-teal-500/[0.03]'
            : 'border-transparent hover:border-zinc-200 dark:hover:border-zinc-800/60 hover:bg-zinc-100/50 dark:hover:bg-zinc-800/30'
      }`}
    >
      <button
        type="button"
        onClick={onToggle}
        className="w-full flex flex-col items-stretch gap-0.5 px-2 py-1.5 text-left min-w-0"
      >
        <div className="flex items-center gap-1.5 min-w-0">
          <ChevronRight
            size={10}
            strokeWidth={1.6}
            className={`shrink-0 text-zinc-400 dark:text-zinc-600 transition-transform ${isOpen ? 'rotate-90' : ''}`}
          />
          {active ? (
            <Loader2 size={12} strokeWidth={1.7} className="animate-spin text-violet-500 dark:text-violet-400 shrink-0" />
          ) : (
            <Icon size={12} strokeWidth={1.7} className={`shrink-0 ${iconToneClass(meta.tone)}`} />
          )}
          <span className={`truncate text-[12.5px] tracking-tight ${active ? 'text-zinc-900 dark:text-zinc-50 font-medium' : 'text-zinc-700 dark:text-zinc-300'}`}>
            {conn.name}
          </span>
          <span className={`ml-auto font-mono text-[10px] lowercase shrink-0 ${active ? 'text-violet-600 dark:text-violet-400' : toneText(meta.tone)}`}>
            {active ? 'busy' : meta.label}
          </span>
        </div>
        {(description || isNew) && (
          <div className="flex items-center gap-1.5 pl-[26px] min-w-0">
            {description && (
              <span className="truncate text-[11px] text-zinc-500 dark:text-zinc-500 min-w-0">
                {description}
              </span>
            )}
            {isNew && !active && (
              <span className="font-mono text-[9.5px] lowercase text-teal-600 dark:text-teal-400 shrink-0 ml-auto">new</span>
            )}
          </div>
        )}
      </button>
      {isOpen && (
        <div className="px-2.5 pb-2 pt-1 space-y-1.5 border-t border-zinc-200/70 dark:border-zinc-800/60">
          <dl className="grid grid-cols-[max-content_1fr] gap-x-3 gap-y-0.5 font-mono text-[10.5px] lowercase">
            <dt className="text-zinc-400 dark:text-zinc-600">id</dt>
            <dd className="text-zinc-600 dark:text-zinc-400 truncate">{conn.id}</dd>
            <dt className="text-zinc-400 dark:text-zinc-600">type</dt>
            <dd className="text-zinc-600 dark:text-zinc-400">{isSandbox ? 'sandbox' : conn.type}</dd>
            {isSandbox && sandbox?.container?.image && (
              <>
                <dt className="text-zinc-400 dark:text-zinc-600">image</dt>
                <dd className="text-zinc-600 dark:text-zinc-400 truncate">{sandbox.container.image}</dd>
              </>
            )}
            {isSandbox && sandbox?.container?.host && (
              <>
                <dt className="text-zinc-400 dark:text-zinc-600">host</dt>
                <dd className="text-zinc-600 dark:text-zinc-400 truncate">
                  {sandbox.container.host}{sandbox.container.port ? `:${sandbox.container.port}` : ''}
                </dd>
              </>
            )}
            {isSandbox && sandbox?.container?.status && (
              <>
                <dt className="text-zinc-400 dark:text-zinc-600">container</dt>
                <dd className={toneText(meta.tone)}>{sandbox.container.status}</dd>
              </>
            )}
            {!isSandbox && conn.config && Object.keys(conn.config).length > 0 && (
              <>
                <dt className="text-zinc-400 dark:text-zinc-600">config</dt>
                <dd className="text-zinc-600 dark:text-zinc-400 truncate">
                  {previewConfig(conn.config)}
                </dd>
              </>
            )}
          </dl>
        </div>
      )}
    </div>
  )
}

function previewConfig(cfg: Record<string, unknown>): string {
  const host = typeof cfg.host === 'string' ? cfg.host : ''
  const port = cfg.port != null ? `:${cfg.port}` : ''
  const user = typeof cfg.username === 'string' ? `${cfg.username}@` : ''
  if (host) return `${user}${host}${port}`
  const keys = Object.keys(cfg).slice(0, 3)
  return keys.map(k => `${k}=${typeof cfg[k] === 'string' ? cfg[k] : '…'}`).join(' ')
}
