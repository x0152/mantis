import { useState, useEffect, useRef, useCallback } from 'react'
import { Terminal, Calculator, Download, Mic, Eye, Wrench, Wand2, Play, GitBranch, Bell, Layers, X, CheckCircle2, AlertCircle, Loader2, ScrollText, Maximize2, Square, RotateCcw, Copy, Check } from 'lucide-react'
import { Markdown } from './Markdown'
import { EntryLine, PromptBanner } from './LogEntries'
import { api } from '../api'
import { navigate } from '../router'
import type { ChatMessage, Step, Attachment, LogEntry, SessionLog } from '../types'

function useTicker(active: boolean, intervalMs = 1000) {
  const [, setTick] = useState(0)
  useEffect(() => {
    if (!active) return
    const id = setInterval(() => setTick(t => t + 1), intervalMs)
    return () => clearInterval(id)
  }, [active, intervalMs])
}

function fmtElapsed(ms: number): string {
  if (ms < 0) ms = 0
  if (ms < 1000) return `${ms}ms`
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`
  const m = Math.floor(ms / 60000)
  const s = Math.floor((ms % 60000) / 1000)
  return `${m}m ${s}s`
}

function estimateTokens(text: string | undefined | null): number {
  if (!text) return 0
  return Math.ceil(text.length / 4)
}

function fmtTokens(n: number): string {
  if (n < 1000) return String(n)
  if (n < 10000) return (n / 1000).toFixed(1) + 'k'
  return Math.round(n / 1000) + 'k'
}

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
  layers: Layers,
}

function stripThinking(content: string): string {
  if (!content) return ''
  return content
    .replace(/<think>[\s\S]*?<\/think>/g, '')
    .replace(/<think>[\s\S]*$/, '')
    .trim()
}

function CopyMessageButton({
  text,
  variant = 'inline',
  tone = 'neutral',
}: {
  text: string
  variant?: 'inline' | 'hover'
  tone?: 'neutral' | 'danger'
}) {
  const [copied, setCopied] = useState(false)
  const resetRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => () => {
    if (resetRef.current) clearTimeout(resetRef.current)
  }, [])

  const handleCopy = async (e: React.MouseEvent<HTMLButtonElement>) => {
    e.stopPropagation()
    if (!text) return
    try {
      await navigator.clipboard.writeText(text)
    } catch {
      try {
        const ta = document.createElement('textarea')
        ta.value = text
        ta.style.position = 'fixed'
        ta.style.left = '-9999px'
        document.body.appendChild(ta)
        ta.select()
        document.execCommand('copy')
        document.body.removeChild(ta)
      } catch { return }
    }
    setCopied(true)
    if (resetRef.current) clearTimeout(resetRef.current)
    resetRef.current = setTimeout(() => setCopied(false), 1500)
  }

  if (variant === 'hover') {
    return (
      <button
        onClick={handleCopy}
        type="button"
        className="opacity-0 group-hover:opacity-100 focus-visible:opacity-100 transition-opacity p-1.5 rounded-md text-zinc-500 dark:text-zinc-500 hover:text-teal-500 dark:hover:text-teal-400 hover:bg-zinc-200 dark:hover:bg-zinc-800 shrink-0"
        title={copied ? 'Copied' : 'Copy message'}
        aria-label={copied ? 'Copied' : 'Copy message'}
      >
        {copied
          ? <Check size={13} className="text-emerald-500" />
          : <Copy size={13} />
        }
      </button>
    )
  }

  const toneClasses = tone === 'danger'
    ? 'text-red-400 hover:text-red-300 hover:bg-red-500/10'
    : 'text-zinc-500 dark:text-zinc-500 hover:text-teal-500 dark:hover:text-teal-400 hover:bg-teal-500/5'

  return (
    <button
      onClick={handleCopy}
      type="button"
      className={`inline-flex items-center gap-1.5 px-2 py-1 text-[11px] font-medium rounded-md transition-colors ${toneClasses}`}
      title={copied ? 'Copied' : 'Copy message'}
      aria-label={copied ? 'Copied' : 'Copy message'}
    >
      {copied
        ? <Check size={11} className="text-emerald-500" />
        : <Copy size={11} />
      }
      {copied ? 'Copied' : 'Copy'}
    </button>
  )
}

export function MessageBubble({
  msg,
  onStepClick,
  canRegenerate,
  onRegenerate,
  regenerating,
}: {
  msg: ChatMessage
  onStepClick: (s: Step) => void
  canRegenerate?: boolean
  onRegenerate?: () => void
  regenerating?: boolean
}) {
  const isUser = msg.role === 'user'
  const steps = msg.steps ?? []

  if (isUser) {
    const userTokens = msg.tokens ?? estimateTokens(msg.content)
    return (
      <div className="group flex justify-end items-center gap-2">
        {msg.content && <CopyMessageButton text={msg.content} variant="hover" />}
        <div className="max-w-[80%] rounded-xl rounded-br-sm text-[15px] bg-teal-600 text-white px-4 py-3">
          <div className="whitespace-pre-wrap">{msg.content}</div>
          {userTokens > 0 && (
            <div className="mt-1 text-[10px] font-mono tabular-nums text-white/60 text-right">
              ~{fmtTokens(userTokens)} tok
            </div>
          )}
        </div>
      </div>
    )
  }

  if (msg.status === 'error') {
    const errorCopyText = stripThinking(msg.content)
    const showErrorActions = !!errorCopyText || (canRegenerate && !!onRegenerate)
    return (
      <div className="flex justify-start">
        <div className="max-w-[80%] rounded-xl rounded-bl-sm text-[15px] bg-red-500/10 text-red-400 border border-red-500/20 px-4 py-3 space-y-2">
          <Markdown content={msg.content} />
          {showErrorActions && (
            <div className="flex items-center gap-1">
              {errorCopyText && <CopyMessageButton text={errorCopyText} tone="danger" />}
              {canRegenerate && onRegenerate && (
                <button
                  onClick={onRegenerate}
                  disabled={regenerating}
                  className="inline-flex items-center gap-1.5 px-2 py-1 text-[11px] font-medium rounded-md text-red-400 hover:text-red-300 hover:bg-red-500/10 disabled:opacity-50 transition-colors"
                  title="Regenerate response"
                >
                  {regenerating
                    ? <Loader2 size={11} className="animate-spin" />
                    : <RotateCcw size={11} />
                  }
                  Regenerate
                </button>
              )}
            </div>
          )}
        </div>
      </div>
    )
  }

  const isPending = msg.status === 'pending'
  const isCancelled = msg.status === 'cancelled'
  const hasRunningStep = steps.some(s => s.status === 'running')
  const parts = buildInterleavedParts(msg.content, steps)
  useTicker(isPending)

  const contentLen = msg.content?.length ?? 0
  const lastContentChangeRef = useRef<number>(Date.now())
  const prevContentLenRef = useRef<number>(contentLen)
  useEffect(() => {
    if (contentLen !== prevContentLenRef.current) {
      prevContentLenRef.current = contentLen
      lastContentChangeRef.current = Date.now()
    }
  }, [contentLen])
  const sinceContentChange = Date.now() - lastContentChangeRef.current
  const lastOpen = msg.content ? msg.content.lastIndexOf('<think>') : -1
  const lastClose = msg.content ? msg.content.lastIndexOf('</think>') : -1
  const inThinkTag = isPending && lastOpen >= 0 && lastOpen > lastClose
  const isTyping = isPending && contentLen > 0 && sinceContentChange < 900 && !inThinkTag
  const showThinking = isPending && !hasRunningStep && !isTyping && !inThinkTag
  const showTyping = isTyping && !hasRunningStep
  const msgStart = new Date(msg.createdAt).getTime()
  const msgEnd = msg.finishedAt ? new Date(msg.finishedAt).getTime() : Date.now()
  const msgElapsed = msgEnd - msgStart
  const showMsgDuration = msgElapsed >= 300 && (isPending || !!msg.finishedAt)
  const isImage = (a: Attachment) =>
    a.mimeType.startsWith('image/') || /\.(png|jpe?g|gif|webp|svg|bmp)$/i.test(a.fileName)
  const images = (msg.attachments ?? []).filter(isImage)
  const otherFiles = (msg.attachments ?? []).filter(a => !isImage(a))

  const estimatedTokens = msg.tokens ?? estimateTokens(msg.content)
  const showTokens = estimatedTokens > 0
  const showHeader = !!(msg.modelName || msg.presetName || msg.modelRole === 'fallback' || showMsgDuration || showTokens)
  const hasText = parts.some(p => p.type === 'text')
  const hasAnyContent =
    parts.length > 0 ||
    images.length > 0 ||
    otherFiles.length > 0 ||
    showThinking ||
    showTyping ||
    isCancelled
  const showEmpty = !hasAnyContent && !isPending

  return (
    <div className="flex justify-start">
      <div className="max-w-[80%] rounded-xl rounded-bl-sm text-[15px] bg-zinc-100 dark:bg-zinc-900 text-zinc-700 dark:text-zinc-300 border border-zinc-200 dark:border-zinc-800 overflow-hidden px-4 py-2.5 space-y-2">
        {showHeader && (
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
            {showMsgDuration && (
              <span className="text-[10px] font-mono text-zinc-500 dark:text-zinc-600 tabular-nums">
                · {fmtElapsed(msgElapsed)}
              </span>
            )}
            {showTokens && (
              <span
                className="text-[10px] font-mono text-zinc-500 dark:text-zinc-600 tabular-nums"
                title="Estimated tokens in this message"
              >
                · ~{fmtTokens(estimatedTokens)} tok
              </span>
            )}
          </div>
        )}
        {parts.map((part, i) => {
          if (part.type === 'step') {
            return (
              <div key={part.step!.id}>
                <StepBadge step={part.step!} onClick={() => onStepClick(part.step!)} />
              </div>
            )
          }
          return (
            <div key={`text-${i}`} className={hasText ? 'py-0.5' : ''}>
              <Markdown content={part.text!} />
            </div>
          )
        })}
        {images.length > 0 && (
          <div className={`grid gap-1.5 ${images.length === 1 ? 'grid-cols-1' : 'grid-cols-2'} -mx-2`}>
            {images.map(a => (
              <AttachmentImage key={a.id} attachment={a} sessionId={msg.sessionId} />
            ))}
          </div>
        )}
        {otherFiles.length > 0 && (
          <div className="space-y-1">
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
        {showThinking && <PendingIndicator mode="thinking" />}
        {showTyping && <PendingIndicator mode="typing" />}
        {isCancelled && (
          <div>
            <span className="inline-flex items-center gap-1 px-1.5 py-0.5 text-[10px] rounded-full bg-zinc-200 dark:bg-zinc-800 text-zinc-500 dark:text-zinc-500">
              <Square size={9} className="fill-current" />
              stopped
            </span>
          </div>
        )}
        {showEmpty && (
          <div className="text-xs italic text-zinc-500 dark:text-zinc-600">
            (no response)
          </div>
        )}
        {!isPending && (() => {
          const copyText = stripThinking(msg.content)
          const canShowCopy = copyText.length > 0
          const canShowRegenerate = !!canRegenerate && !!onRegenerate
          if (!canShowCopy && !canShowRegenerate) return null
          return (
            <div className="flex items-center gap-1">
              {canShowCopy && <CopyMessageButton text={copyText} />}
              {canShowRegenerate && (
                <button
                  onClick={onRegenerate}
                  disabled={regenerating}
                  className="inline-flex items-center gap-1.5 px-2 py-1 text-[11px] font-medium rounded-md text-zinc-500 dark:text-zinc-500 hover:text-teal-500 dark:hover:text-teal-400 hover:bg-teal-500/5 disabled:opacity-50 transition-colors"
                  title="Regenerate response"
                >
                  {regenerating
                    ? <Loader2 size={11} className="animate-spin" />
                    : <RotateCcw size={11} />
                  }
                  Regenerate
                </button>
              )}
            </div>
          )
        })()}
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
  const isCancelled = step.status === 'cancelled'
  const argSummary = stepArgsSummary(step)

  useTicker(isRunning)
  const startMs = step.startedAt ? new Date(step.startedAt).getTime() : 0
  const endMs = step.finishedAt ? new Date(step.finishedAt).getTime() : Date.now()
  const elapsed = startMs > 0 ? endMs - startMs : 0
  const showElapsed = elapsed >= 300 || isRunning

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
      className={`inline-flex items-start gap-1.5 px-2 py-1 rounded-md text-xs font-medium cursor-pointer transition-colors max-w-full text-left step-enter ${
        isRunning
          ? 'text-teal-400 border border-teal-500/40 bg-teal-500/5 step-running'
          : isError
            ? 'bg-red-500/10 text-red-400 border border-red-500/20'
            : isCancelled
              ? 'bg-zinc-200/60 dark:bg-zinc-800/60 text-zinc-400 dark:text-zinc-500 border border-zinc-300/60 dark:border-zinc-700/60 line-through decoration-zinc-400/50'
              : 'bg-zinc-200 dark:bg-zinc-800 text-zinc-500 border border-zinc-300 dark:border-zinc-700 hover:text-zinc-700 dark:hover:text-zinc-300'
      }`}
    >
      {isRunning
        ? <Loader2 size={11} className="animate-spin shrink-0 mt-0.5" />
        : isError
          ? <AlertCircle size={11} className="shrink-0 mt-0.5" />
          : isCancelled
            ? <Square size={10} className="fill-current opacity-70 shrink-0 mt-1" />
            : <Icon size={11} className="shrink-0 mt-0.5" />
      }
      <span className="break-words whitespace-normal leading-snug min-w-0 flex-1">
        {step.label}
        {argSummary && (
          <span className="ml-1.5 text-[10px] opacity-60 font-normal">{argSummary}</span>
        )}
      </span>
      {showElapsed && (
        <span className="text-[10px] font-mono opacity-70 tabular-nums shrink-0 mt-0.5">{fmtElapsed(elapsed)}</span>
      )}
      {!isRunning && !isError && !isCancelled && <CheckCircle2 size={10} className="text-emerald-500 shrink-0 mt-1" />}
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
      <div className="w-[420px] max-w-[48vw] min-w-[340px] h-full bg-white dark:bg-zinc-900 border-l border-zinc-200 dark:border-zinc-800 flex flex-col shadow-2xl">
        <div className="flex items-center justify-between px-5 py-4 border-b border-zinc-200 dark:border-zinc-800 shrink-0">
          <div className="flex items-center gap-2 min-w-0">
            {isRunning
              ? <Loader2 size={14} className="text-teal-400 animate-spin shrink-0" />
              : isError
                ? <AlertCircle size={14} className="text-red-400 shrink-0" />
                : <CheckCircle2 size={14} className="text-emerald-400 shrink-0" />
            }
            <span className="font-medium text-sm text-zinc-900 dark:text-zinc-100 whitespace-normal break-words leading-snug">{step.label}</span>
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

export function PendingIndicator({ mode = 'thinking' }: { mode?: 'thinking' | 'typing' }) {
  const label = mode === 'typing' ? 'Typing' : 'Thinking'
  return (
    <div className="flex items-center gap-1 py-1">
      <span className="text-[13px] font-medium shimmer-text">{label}</span>
      {mode === 'typing' && (
        <span className="typing-caret text-zinc-500 dark:text-zinc-400" aria-hidden />
      )}
    </div>
  )
}
