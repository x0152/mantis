import { useState, useEffect, useRef, useCallback } from 'react'
import { Terminal, Calculator, Download, Mic, Eye, Wrench, Wand2, Play, GitBranch, Bell, X, CheckCircle2, AlertCircle, Loader2, ScrollText, Maximize2 } from 'lucide-react'
import { Markdown } from './Markdown'
import { EntryLine, PromptBanner } from './LogEntries'
import { api } from '../api'
import { navigate } from '../router'
import type { ChatMessage, Step, Attachment, LogEntry, SessionLog } from '../types'

export const STEP_ICONS: Record<string, typeof Terminal> = {
  terminal: Terminal,
  calculator: Calculator,
  download: Download,
  mic: Mic,
  eye: Eye,
  skill: Wand2,
  play: Play,
  'git-branch': GitBranch,
  bell: Bell,
}

export function MessageBubble({ msg, onStepClick }: { msg: ChatMessage; onStepClick: (s: Step) => void }) {
  const isUser = msg.role === 'user'
  const steps = msg.steps ?? []

  if (isUser) {
    return (
      <div className="flex justify-end">
        <div className="max-w-[70%] rounded-xl rounded-br-sm text-sm bg-teal-600 text-white px-3.5 py-2.5">
          <div className="whitespace-pre-wrap">{msg.content}</div>
        </div>
      </div>
    )
  }

  if (msg.status === 'error') {
    return (
      <div className="flex justify-start">
        <div className="max-w-[70%] rounded-xl rounded-bl-sm text-sm bg-red-500/10 text-red-400 border border-red-500/20 px-3.5 py-2.5">
          <Markdown content={msg.content} />
        </div>
      </div>
    )
  }

  const isPending = msg.status === 'pending'
  const parts = buildInterleavedParts(msg.content, steps)
  const isImage = (a: Attachment) =>
    a.mimeType.startsWith('image/') || /\.(png|jpe?g|gif|webp|svg|bmp)$/i.test(a.fileName)
  const images = (msg.attachments ?? []).filter(isImage)
  const otherFiles = (msg.attachments ?? []).filter(a => !isImage(a))

  return (
    <div className="flex justify-start">
      <div className="max-w-[70%] rounded-xl rounded-bl-sm text-sm bg-zinc-100 dark:bg-zinc-900 text-zinc-700 dark:text-zinc-300 border border-zinc-200 dark:border-zinc-800 overflow-hidden">
        {(msg.modelName || msg.presetName || msg.modelRole === 'fallback') && (
          <div className="px-3.5 pt-2 pb-0">
            <div className="flex items-center gap-1.5 flex-wrap">
              {msg.presetName && (
                <span className="px-1.5 py-0.5 text-[10px] rounded-full bg-zinc-200 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-500">
                  {msg.presetName}
                </span>
              )}
              {msg.modelName && (
                <span className="text-[10px] font-medium text-zinc-500 dark:text-zinc-600">
                  {msg.modelName}
                </span>
              )}
              {msg.modelRole === 'fallback' && (
                <span className="px-1.5 py-0.5 text-[10px] rounded-full bg-amber-500/15 text-amber-400">
                  fallback
                </span>
              )}
            </div>
          </div>
        )}
        {parts.map((part, i) => {
          if (part.type === 'step') {
            return (
              <div key={part.step!.id} className={`px-3.5 ${i === 0 ? 'pt-2' : 'pt-1'} pb-1`}>
                <StepBadge step={part.step!} onClick={() => onStepClick(part.step!)} />
              </div>
            )
          }
          return (
            <div key={`text-${i}`} className={`px-3.5 ${i === 0 ? 'py-2.5' : 'pt-1 pb-2.5'}`}>
              <Markdown content={part.text!} />
            </div>
          )
        })}
        {images.length > 0 && (
          <div className={`px-2 pb-2 ${parts.length > 0 ? 'pt-1' : 'pt-2'}`}>
            <div className={`grid gap-1.5 ${images.length === 1 ? 'grid-cols-1' : 'grid-cols-2'}`}>
              {images.map(a => (
                <AttachmentImage key={a.id} attachment={a} sessionId={msg.sessionId} />
              ))}
            </div>
          </div>
        )}
        {otherFiles.length > 0 && (
          <div className="px-3.5 pb-2.5 pt-1 space-y-1">
            {otherFiles.map(a => (
              <a
                key={a.id}
                href={`/api/artifacts/${msg.sessionId}/${a.id}`}
                download={a.fileName}
                className="flex items-center gap-1.5 px-2 py-1 rounded-md text-xs bg-zinc-200 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-400 hover:text-zinc-800 dark:hover:text-zinc-200"
              >
                <Download size={11} />
                <span className="truncate">{a.fileName}</span>
                <span className="text-[10px] text-zinc-500 shrink-0">{formatBytes(a.size)}</span>
              </a>
            ))}
          </div>
        )}
        {isPending && (
          <div className="px-3.5 pb-2.5 pt-1"><PendingIndicator /></div>
        )}
      </div>
    </div>
  )
}

type Part = { type: 'text'; text: string } | { type: 'step'; step: Step }

const encoder = new TextEncoder()
const decoder = new TextDecoder()

function buildInterleavedParts(content: string, steps: Step[]): (Part & { text?: string; step?: Step })[] {
  if (steps.length === 0) {
    return content ? [{ type: 'text', text: content }] : []
  }

  const hasOffsets = steps.some(s => (s.contentOffset ?? 0) > 0)
  if (!hasOffsets) {
    const parts: (Part & { text?: string; step?: Step })[] = steps.map(s => ({ type: 'step' as const, step: s }))
    if (content) parts.push({ type: 'text', text: content })
    return parts
  }

  const sorted = [...steps].sort((a, b) => (a.contentOffset ?? 0) - (b.contentOffset ?? 0))
  const bytes = encoder.encode(content)
  const parts: (Part & { text?: string; step?: Step })[] = []
  let pos = 0

  for (const step of sorted) {
    const offset = step.contentOffset ?? 0
    if (offset > pos) {
      const text = decoder.decode(bytes.slice(pos, offset)).trim()
      if (text) parts.push({ type: 'text', text })
    }
    parts.push({ type: 'step', step })
    pos = Math.max(pos, offset)
  }

  if (pos < bytes.length) {
    const text = decoder.decode(bytes.slice(pos)).trim()
    if (text) parts.push({ type: 'text', text })
  }

  return parts
}

function stepArgsSummary(step: Step): string {
  try {
    const parsed = JSON.parse(step.args)
    const keys = Object.keys(parsed).filter(k => k !== 'task' && k !== 'prompt')
    if (keys.length === 0) return ''
    const parts = keys.slice(0, 3).map(k => {
      const v = parsed[k]
      const s = typeof v === 'string' ? v : JSON.stringify(v)
      return s.length > 40 ? s.slice(0, 37) + '...' : s
    })
    return parts.join(', ')
  } catch {
    return ''
  }
}

function planIdFromStepArgs(step: Step): string | undefined {
  if (!['plan_run', 'plan_create', 'plan_update', 'plan_get'].includes(step.tool)) return undefined
  try {
    const parsed = JSON.parse(step.result || '{}')
    return parsed.id || parsed.planId || parsed.plan_id
  } catch {
    return undefined
  }
}

export function StepBadge({ step, onClick }: { step: Step; onClick: () => void }) {
  const Icon = STEP_ICONS[step.icon] ?? Wrench
  const isRunning = step.status === 'running'
  const isError = step.status === 'error'
  const argSummary = stepArgsSummary(step)

  const planId = !isRunning ? planIdFromStepArgs(step) : undefined
  const handleClick = () => {
    if (planId) {
      navigate({ page: 'plans', planId })
    }
    onClick()
  }

  return (
    <button
      onClick={handleClick}
      className={`inline-flex items-center gap-1.5 px-2 py-1 rounded-md text-xs font-medium cursor-pointer ${
        isRunning
          ? 'bg-teal-500/10 text-teal-400 border border-teal-500/20'
          : isError
            ? 'bg-red-500/10 text-red-400 border border-red-500/20'
            : 'bg-zinc-200 dark:bg-zinc-800 text-zinc-500 border border-zinc-300 dark:border-zinc-700 hover:text-zinc-700 dark:hover:text-zinc-300'
      }`}
    >
      {isRunning
        ? <Loader2 size={11} className="animate-spin" />
        : isError
          ? <AlertCircle size={11} />
          : <Icon size={11} />
      }
      <span className="max-w-[200px] truncate">{step.label}</span>
      {argSummary && (
        <span className="max-w-[150px] truncate text-[10px] opacity-60 font-normal">{argSummary}</span>
      )}
      {!isRunning && !isError && <CheckCircle2 size={10} className="text-emerald-500" />}
    </button>
  )
}

function formatArgs(raw: string): string {
  try {
    return JSON.stringify(JSON.parse(raw), null, 2)
  } catch {
    return raw
  }
}

function extractStepPrompt(step: Step): string {
  try {
    const parsed = JSON.parse(step.args)
    return (parsed.task ?? parsed.prompt ?? '') as string
  } catch {
    return ''
  }
}

function stepToEntries(step: Step): LogEntry[] {
  const entries: LogEntry[] = []
  const ts = step.startedAt || new Date().toISOString()
  const endTs = step.finishedAt || ts

  entries.push({ type: 'command', content: JSON.stringify(step), timestamp: ts })

  if (step.result) {
    entries.push({
      type: step.status === 'error' ? 'error' : 'output',
      content: step.result,
      timestamp: endTs,
    })
  }

  return entries
}

function fmtDuration(start: string, end: string) {
  const ms = new Date(end).getTime() - new Date(start).getTime()
  if (ms < 1000) return `${ms}ms`
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`
  return `${Math.floor(ms / 60000)}m ${Math.floor((ms % 60000) / 1000)}s`
}

export function StepPanel({ step, onClose }: { step: Step; onClose: () => void }) {
  const [log, setLog] = useState<SessionLog | null>(null)
  const [logLoading, setLogLoading] = useState(false)
  const [showLog, setShowLog] = useState(false)
  const logPollRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const logScrollRef = useRef<HTMLDivElement>(null)
  const prevLogId = useRef(step.logId)

  const isRunning = step.status === 'running'
  const isError = step.status === 'error'
  const prompt = extractStepPrompt(step)
  const entries = stepToEntries(step)

  const fetchLog = useCallback(async () => {
    if (!step.logId) return
    try {
      const data = await api.sessionLogs.get(step.logId)
      setLog(data)
    } catch { /* noop */ }
  }, [step.logId])

  useEffect(() => {
    if (step.logId !== prevLogId.current) {
      prevLogId.current = step.logId
      setLog(null)
      setShowLog(false)
    }
  }, [step.logId])

  useEffect(() => {
    if (!showLog) return
    fetchLog()
    if (isRunning || log?.status === 'running') {
      logPollRef.current = setInterval(fetchLog, 500)
    }
    return () => { if (logPollRef.current) { clearInterval(logPollRef.current); logPollRef.current = null } }
  }, [showLog, isRunning, fetchLog])

  useEffect(() => {
    if (!isRunning && log?.status !== 'running' && logPollRef.current) {
      clearInterval(logPollRef.current)
      logPollRef.current = null
      fetchLog()
    }
  }, [isRunning, log?.status, fetchLog])

  useEffect(() => {
    if (showLog && logScrollRef.current) {
      const el = logScrollRef.current
      const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 80
      if (atBottom) el.scrollTop = el.scrollHeight
    }
  }, [log?.entries.length, showLog])

  async function openLog() {
    if (!step.logId) return
    setShowLog(true)
    if (!log) {
      setLogLoading(true)
      await fetchLog()
      setLogLoading(false)
    }
  }

  return (
    <>
      <div className="fixed inset-0 z-40" onClick={onClose} />
      <div className="fixed right-0 top-0 bottom-0 z-50 w-full max-w-md bg-white dark:bg-zinc-900 border-l border-zinc-200 dark:border-zinc-800 flex flex-col shadow-2xl">
        <div className="flex items-center justify-between px-5 py-4 border-b border-zinc-200 dark:border-zinc-800 shrink-0">
          <div className="flex items-center gap-2 min-w-0">
            {isRunning
              ? <Loader2 size={14} className="text-teal-400 animate-spin shrink-0" />
              : isError
                ? <AlertCircle size={14} className="text-red-400 shrink-0" />
                : <CheckCircle2 size={14} className="text-emerald-400 shrink-0" />
            }
            <span className="font-medium text-sm text-zinc-900 dark:text-zinc-100 truncate">{step.label}</span>
          </div>
          <button onClick={onClose} className="p-1 rounded-md text-zinc-500 dark:text-zinc-600 hover:text-zinc-700 dark:hover:text-zinc-300 hover:bg-zinc-100 dark:hover:bg-zinc-800 shrink-0 ml-2">
            <X size={16} />
          </button>
        </div>

        <div className="px-5 py-3 border-b border-zinc-200 dark:border-zinc-800 space-y-2 shrink-0">
          <div className="flex items-center gap-2 flex-wrap">
            <span className="text-[11px] text-zinc-500 dark:text-zinc-600 font-mono">{step.tool}</span>
            {step.presetName && (
              <span className="px-1.5 py-0.5 text-[10px] font-medium rounded-full bg-zinc-200 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-500">{step.presetName}</span>
            )}
            {step.modelName && (
              <span className="px-1.5 py-0.5 text-[10px] font-medium rounded-full bg-zinc-200 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-500">{step.modelName}</span>
            )}
            {step.modelRole === 'fallback' && (
              <span className="px-1.5 py-0.5 text-[10px] font-medium rounded-full bg-amber-500/15 text-amber-400">fallback</span>
            )}
            {step.finishedAt && step.startedAt && (
              <span className="text-[11px] text-zinc-500 dark:text-zinc-600">{fmtDuration(step.startedAt, step.finishedAt)}</span>
            )}
          </div>
          {step.args && step.args !== '{}' && (
            <pre className="text-[11px] font-mono text-zinc-500 dark:text-zinc-600 bg-zinc-50 dark:bg-zinc-950 rounded-md px-2.5 py-1.5 overflow-x-auto max-h-24 whitespace-pre-wrap break-all">{formatArgs(step.args)}</pre>
          )}
          {step.logId && (
            <button
              onClick={openLog}
              disabled={logLoading}
              className="flex items-center gap-1.5 px-2.5 py-1.5 text-xs font-medium text-teal-400 bg-teal-500/10 border border-teal-500/20 rounded-lg hover:bg-teal-500/20 disabled:opacity-50"
            >
              {logLoading ? <Loader2 size={12} className="animate-spin" /> : <ScrollText size={12} />}
              {isRunning ? 'View Live Log' : 'View Agent Log'}
            </button>
          )}
        </div>

        <div className="flex-1 overflow-auto min-h-0">
          <div className="bg-zinc-50 dark:bg-zinc-950 px-4 py-3.5 space-y-1 min-h-[120px]">
            {prompt && <PromptBanner prompt={prompt} />}
            {entries.map((entry, i) => <EntryLine key={i} entry={entry} />)}
            {isRunning && !step.result && (
              <div className="font-mono text-xs flex items-center gap-2 text-zinc-500 dark:text-zinc-600 py-2">
                <Loader2 size={11} className="animate-spin" />
                <span>Running...</span>
              </div>
            )}
          </div>
        </div>
      </div>

      {showLog && (
        <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/60" onClick={() => setShowLog(false)}>
          <div
            className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-xl w-full max-w-3xl mx-4 overflow-hidden flex flex-col max-h-[85vh]"
            onClick={e => e.stopPropagation()}
          >
            <div className="flex items-center justify-between px-5 py-4 border-b border-zinc-200 dark:border-zinc-800 shrink-0">
              <div className="flex items-center gap-2.5">
                <ScrollText size={14} className="text-teal-400" />
                <span className="font-medium text-sm text-zinc-900 dark:text-zinc-100">{log?.agentName ?? 'Agent'}</span>
                {log && (
                  <span className={`px-1.5 py-0.5 text-[10px] font-medium rounded-full ${
                    log.status === 'running' ? 'bg-amber-500/15 text-amber-400' : 'bg-emerald-500/15 text-emerald-400'
                  }`}>{log.status}</span>
                )}
                {log && (log.presetName || log.modelName || log.modelRole === 'fallback') && (
                  <div className="flex items-center gap-1.5">
                    {log.presetName && (
                      <span className="px-1.5 py-0.5 text-[10px] font-medium rounded-full bg-zinc-200 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-500">{log.presetName}</span>
                    )}
                    {log.modelName && (
                      <span className="px-1.5 py-0.5 text-[10px] font-medium rounded-full bg-zinc-200 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-500">{log.modelName}</span>
                    )}
                    {log.modelRole === 'fallback' && (
                      <span className="px-1.5 py-0.5 text-[10px] font-medium rounded-full bg-amber-500/15 text-amber-400">fallback</span>
                    )}
                  </div>
                )}
                {!log && <Loader2 size={12} className="text-zinc-500 dark:text-zinc-600 animate-spin" />}
              </div>
              <button onClick={() => setShowLog(false)} className="p-1 rounded-md text-zinc-500 dark:text-zinc-600 hover:text-zinc-700 dark:hover:text-zinc-300 hover:bg-zinc-100 dark:hover:bg-zinc-800">
                <X size={16} />
              </button>
            </div>
            <div ref={logScrollRef} className="flex-1 overflow-auto min-h-0">
              <div className="bg-zinc-50 dark:bg-zinc-950 px-4 py-3.5 space-y-1 min-h-[120px]">
                {log?.prompt && <PromptBanner prompt={log.prompt} />}
                {log && log.entries.length === 0 && log.status === 'running' ? (
                  <div className="font-mono text-xs flex items-center gap-2 text-zinc-500 dark:text-zinc-600 py-2">
                    <Loader2 size={11} className="animate-spin" />
                    <span>Waiting for output...</span>
                  </div>
                ) : log && log.entries.length === 0 ? (
                  <p className="text-zinc-500 dark:text-zinc-600 text-xs font-mono">No entries</p>
                ) : log ? (
                  <>
                    {log.entries.map((entry, i) => <EntryLine key={i} entry={entry} />)}
                    {log.status === 'running' && (
                      <div className="font-mono text-xs flex items-center gap-2 text-zinc-500 dark:text-zinc-600 py-2">
                        <Loader2 size={11} className="animate-spin" />
                        <span>Running...</span>
                      </div>
                    )}
                  </>
                ) : (
                  <div className="font-mono text-xs flex items-center gap-2 text-zinc-500 dark:text-zinc-600 py-2">
                    <Loader2 size={11} className="animate-spin" />
                    <span>Loading...</span>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  )
}

function AttachmentImage({ attachment, sessionId }: { attachment: Attachment; sessionId: string }) {
  const [expanded, setExpanded] = useState(false)
  const src = `/api/artifacts/${sessionId}/${attachment.id}`

  return (
    <>
      <div className="relative group cursor-pointer rounded-lg overflow-hidden" onClick={() => setExpanded(true)}>
        <img
          src={src}
          alt={attachment.fileName}
          className="w-full rounded-lg"
          loading="lazy"
        />
        <div className="absolute inset-0 bg-black/0 group-hover:bg-black/20 transition-colors flex items-center justify-center">
          <Maximize2 size={20} className="text-white opacity-0 group-hover:opacity-100 transition-opacity drop-shadow" />
        </div>
        <div className="absolute bottom-0 left-0 right-0 px-2 py-1 bg-gradient-to-t from-black/50 to-transparent">
          <span className="text-[10px] text-white/80 truncate block">{attachment.fileName}</span>
        </div>
      </div>

      {expanded && (
        <div className="fixed inset-0 z-[70] flex items-center justify-center bg-black/70" onClick={() => setExpanded(false)}>
          <div className="relative max-w-[90vw] max-h-[90vh]" onClick={e => e.stopPropagation()}>
            <img src={src} alt={attachment.fileName} className="max-w-full max-h-[90vh] rounded-lg shadow-2xl" />
            <button
              onClick={() => setExpanded(false)}
              className="absolute top-2 right-2 p-1.5 rounded-full bg-black/50 text-white hover:bg-black/70"
            >
              <X size={16} />
            </button>
            <div className="absolute bottom-2 left-2 right-2 flex items-center justify-between">
              <span className="text-xs text-white/80 bg-black/40 px-2 py-1 rounded">{attachment.fileName}</span>
              <a
                href={src}
                download={attachment.fileName}
                onClick={e => e.stopPropagation()}
                className="text-xs text-white/80 bg-black/40 px-2 py-1 rounded hover:bg-black/60"
              >
                <Download size={12} className="inline mr-1" />
                Download
              </a>
            </div>
          </div>
        </div>
      )}
    </>
  )
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export function PendingIndicator() {
  return (
    <div className="flex items-center gap-2 py-1">
      <Loader2 size={12} className="animate-spin text-teal-500" />
      <span className="text-xs text-zinc-400 dark:text-zinc-600">Thinking…</span>
    </div>
  )
}
