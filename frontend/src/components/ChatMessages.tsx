import { useState, useEffect, useRef, useCallback } from 'react'
import { Terminal, Calculator, Download, Mic, Eye, Wrench, X, CheckCircle2, AlertCircle, Loader2, ScrollText } from 'lucide-react'
import { Markdown } from './Markdown'
import { EntryLine, PromptBanner } from './LogEntries'
import { api } from '../api'
import type { ChatMessage, Step, LogEntry, SessionLog } from '../types'

export const STEP_ICONS: Record<string, typeof Terminal> = {
  terminal: Terminal,
  calculator: Calculator,
  download: Download,
  mic: Mic,
  eye: Eye,
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

  const hasContent = !!msg.content
  const isPending = msg.status === 'pending' && !hasContent && steps.length === 0
  const parts = buildInterleavedParts(msg.content, steps)

  return (
    <div className="flex justify-start">
      <div className="max-w-[70%] rounded-xl rounded-bl-sm text-sm bg-zinc-900 text-zinc-300 border border-zinc-800 overflow-hidden">
        {msg.modelName && (
          <div className="px-3.5 pt-2 pb-0">
            <span className="text-[10px] font-medium text-zinc-600">{msg.modelName}</span>
          </div>
        )}
        {isPending && (
          <div className="px-3.5 py-2.5"><PendingIndicator /></div>
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

export function StepBadge({ step, onClick }: { step: Step; onClick: () => void }) {
  const Icon = STEP_ICONS[step.icon] ?? Wrench
  const isRunning = step.status === 'running'
  const isError = step.status === 'error'

  return (
    <button
      onClick={onClick}
      className={`inline-flex items-center gap-1.5 px-2 py-1 rounded-md text-xs font-medium cursor-pointer ${
        isRunning
          ? 'bg-teal-500/10 text-teal-400 border border-teal-500/20'
          : isError
            ? 'bg-red-500/10 text-red-400 border border-red-500/20'
            : 'bg-zinc-800 text-zinc-500 border border-zinc-700 hover:text-zinc-300'
      }`}
    >
      {isRunning
        ? <Loader2 size={11} className="animate-spin" />
        : isError
          ? <AlertCircle size={11} />
          : <Icon size={11} />
      }
      <span className="max-w-[200px] truncate">{step.label}</span>
      {!isRunning && !isError && <CheckCircle2 size={10} className="text-emerald-500" />}
    </button>
  )
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
      <div className="fixed right-0 top-0 bottom-0 z-50 w-full max-w-md bg-zinc-900 border-l border-zinc-800 flex flex-col shadow-2xl">
        <div className="flex items-center justify-between px-5 py-4 border-b border-zinc-800 shrink-0">
          <div className="flex items-center gap-2 min-w-0">
            {isRunning
              ? <Loader2 size={14} className="text-teal-400 animate-spin shrink-0" />
              : isError
                ? <AlertCircle size={14} className="text-red-400 shrink-0" />
                : <CheckCircle2 size={14} className="text-emerald-400 shrink-0" />
            }
            <span className="font-medium text-sm text-zinc-100 truncate">{step.label}</span>
          </div>
          <button onClick={onClose} className="p-1 rounded-md text-zinc-600 hover:text-zinc-300 hover:bg-zinc-800 shrink-0 ml-2">
            <X size={16} />
          </button>
        </div>

        <div className="px-5 py-3 border-b border-zinc-800 space-y-2 shrink-0">
          <div className="flex items-center gap-2 flex-wrap">
            <span className="text-[11px] text-zinc-600 font-mono">{step.tool}</span>
            {step.modelName && (
              <span className="px-1.5 py-0.5 text-[10px] font-medium rounded-full bg-zinc-800 text-zinc-500">{step.modelName}</span>
            )}
            {step.finishedAt && step.startedAt && (
              <span className="text-[11px] text-zinc-600">{fmtDuration(step.startedAt, step.finishedAt)}</span>
            )}
          </div>
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
          <div className="bg-zinc-950 px-4 py-3.5 space-y-1 min-h-[120px]">
            {prompt && <PromptBanner prompt={prompt} />}
            {entries.map((entry, i) => <EntryLine key={i} entry={entry} />)}
            {isRunning && !step.result && (
              <div className="font-mono text-xs flex items-center gap-2 text-zinc-600 py-2">
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
            className="bg-zinc-900 border border-zinc-800 rounded-xl w-full max-w-3xl mx-4 overflow-hidden flex flex-col max-h-[85vh]"
            onClick={e => e.stopPropagation()}
          >
            <div className="flex items-center justify-between px-5 py-4 border-b border-zinc-800 shrink-0">
              <div className="flex items-center gap-2.5">
                <ScrollText size={14} className="text-teal-400" />
                <span className="font-medium text-sm text-zinc-100">{log?.agentName ?? 'Agent'}</span>
                {log && (
                  <span className={`px-1.5 py-0.5 text-[10px] font-medium rounded-full ${
                    log.status === 'running' ? 'bg-amber-500/15 text-amber-400' : 'bg-emerald-500/15 text-emerald-400'
                  }`}>{log.status}</span>
                )}
                {!log && <Loader2 size={12} className="text-zinc-600 animate-spin" />}
              </div>
              <button onClick={() => setShowLog(false)} className="p-1 rounded-md text-zinc-600 hover:text-zinc-300 hover:bg-zinc-800">
                <X size={16} />
              </button>
            </div>
            <div ref={logScrollRef} className="flex-1 overflow-auto min-h-0">
              <div className="bg-zinc-950 px-4 py-3.5 space-y-1 min-h-[120px]">
                {log?.prompt && <PromptBanner prompt={log.prompt} />}
                {log && log.entries.length === 0 && log.status === 'running' ? (
                  <div className="font-mono text-xs flex items-center gap-2 text-zinc-600 py-2">
                    <Loader2 size={11} className="animate-spin" />
                    <span>Waiting for output...</span>
                  </div>
                ) : log && log.entries.length === 0 ? (
                  <p className="text-zinc-600 text-xs font-mono">No entries</p>
                ) : log ? (
                  <>
                    {log.entries.map((entry, i) => <EntryLine key={i} entry={entry} />)}
                    {log.status === 'running' && (
                      <div className="font-mono text-xs flex items-center gap-2 text-zinc-600 py-2">
                        <Loader2 size={11} className="animate-spin" />
                        <span>Running...</span>
                      </div>
                    )}
                  </>
                ) : (
                  <div className="font-mono text-xs flex items-center gap-2 text-zinc-600 py-2">
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

export function PendingIndicator() {
  return (
    <div className="flex items-center gap-1.5 py-1 text-zinc-600">
      <Loader2 size={13} className="animate-spin" />
      <span className="text-xs">Thinking...</span>
    </div>
  )
}
