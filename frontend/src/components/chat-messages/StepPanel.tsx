import { useCallback, useEffect, useRef, useState } from 'react'
import { AlertCircle, CheckCircle2, Loader2, ScrollText, X } from '@/lib/icons'
import { api } from '../../api'
import type { SessionLog, Step } from '../../types'
import { EntryLine, PromptBanner } from '../LogEntries'
import { extractStepPrompt, stepToEntries } from './stepHelpers'
import { fmtDuration, formatArgs } from './utils'

export function StepPanel({ step, onClose }: { step: Step; onClose: () => void }) {
  const [log, setLog] = useState<SessionLog | null>(null)
  const logPollRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const logScrollRef = useRef<HTMLDivElement>(null)
  const prevLogId = useRef(step.logId)
  const stickToBottomRef = useRef(true)
  const autoScrollingRef = useRef(false)

  const isRunning = step.status === 'running'
  const isError = step.status === 'error'
  const prompt = extractStepPrompt(step)
  const stepEntries = stepToEntries(step)
  const hasLog = !!step.logId
  const logPrompt = log?.prompt || prompt

  const fetchLog = useCallback(async () => {
    if (!step.logId) return
    try {
      const data = await api.sessionLogs.get(step.logId)
      setLog(data)
    } catch {}
  }, [step.logId])

  useEffect(() => {
    if (step.logId !== prevLogId.current) {
      prevLogId.current = step.logId
      setLog(null)
      stickToBottomRef.current = true
    }
  }, [step.logId])

  useEffect(() => {
    if (!step.logId) return
    fetchLog()
  }, [step.logId, fetchLog])

  useEffect(() => {
    if (!step.logId) return
    const shouldPoll = isRunning || log?.status === 'running' || !log
    if (shouldPoll) logPollRef.current = setInterval(fetchLog, 500)
    return () => {
      if (logPollRef.current) {
        clearInterval(logPollRef.current)
        logPollRef.current = null
      }
    }
  }, [step.logId, isRunning, log?.status, log, fetchLog])

  const handleScroll = useCallback(() => {
    if (autoScrollingRef.current) return
    const el = logScrollRef.current
    if (!el) return
    const distance = el.scrollHeight - el.scrollTop - el.clientHeight
    stickToBottomRef.current = distance < 40
  }, [])

  useEffect(() => {
    if (!stickToBottomRef.current) return
    const el = logScrollRef.current
    if (!el) return
    autoScrollingRef.current = true
    el.scrollTop = el.scrollHeight
    requestAnimationFrame(() => {
      autoScrollingRef.current = false
    })
  }, [log?.entries.length, stepEntries.length])

  return (
    <div className="w-[560px] max-w-[60vw] min-w-[400px] h-full bg-white dark:bg-zinc-900 border-l border-zinc-200 dark:border-zinc-800 flex flex-col shadow-2xl">
      <PanelHeader step={step} isRunning={isRunning} isError={isError} onClose={onClose} />
      <PanelMeta step={step} log={log} hasLog={hasLog} />
      <div ref={logScrollRef} onScroll={handleScroll} className="flex-1 overflow-auto min-h-0">
        <div className="bg-zinc-50 dark:bg-zinc-950 px-4 py-3.5 space-y-1 min-h-[120px]">
          {logPrompt && <PromptBanner prompt={logPrompt} />}
          <PanelBody
            log={log}
            hasLog={hasLog}
            isRunning={isRunning}
            stepEntries={stepEntries}
            stepResultPresent={!!step.result}
          />
        </div>
      </div>
    </div>
  )
}

function PanelHeader({
  step,
  isRunning,
  isError,
  onClose,
}: {
  step: Step
  isRunning: boolean
  isError: boolean
  onClose: () => void
}) {
  return (
    <div className="flex items-center justify-between px-5 py-4 border-b border-zinc-200 dark:border-zinc-800 shrink-0">
      <div className="flex items-center gap-2 min-w-0">
        {isRunning ? (
          <Loader2 size={14} className="text-teal-400 animate-spin shrink-0" />
        ) : isError ? (
          <AlertCircle size={14} className="text-red-400 shrink-0" />
        ) : (
          <CheckCircle2 size={14} className="text-emerald-400 shrink-0" />
        )}
        <span className="font-medium text-sm text-zinc-900 dark:text-zinc-100 whitespace-normal break-words leading-snug">
          {step.label}
        </span>
      </div>
      <button
        onClick={onClose}
        className="p-1 rounded-md text-zinc-500 dark:text-zinc-600 hover:text-zinc-700 dark:hover:text-zinc-300 hover:bg-zinc-100 dark:hover:bg-zinc-800 shrink-0 ml-2"
      >
        <X size={16} />
      </button>
    </div>
  )
}

function PanelMeta({ step, log, hasLog }: { step: Step; log: SessionLog | null; hasLog: boolean }) {
  return (
    <div className="px-5 py-3 border-b border-zinc-200 dark:border-zinc-800 space-y-2 shrink-0">
      <div className="flex items-center gap-x-2 gap-y-1 flex-wrap font-mono text-[11px] lowercase tracking-tight text-zinc-500 dark:text-zinc-500">
        <span>{step.tool}</span>
        {hasLog && (log?.agentName || log?.status) && (
          <>
            <span className="text-zinc-300 dark:text-zinc-700">/</span>
            <span className="flex items-center gap-1">
              <ScrollText size={11} className="text-teal-500 dark:text-teal-400" />
              {log?.agentName ?? 'agent'}
              {log?.status === 'running' && (
                <span className="px-1 py-px rounded-sm bg-amber-500/15 text-amber-600 dark:text-amber-400">running</span>
              )}
            </span>
          </>
        )}
        {(step.presetName || log?.presetName) && (
          <>
            <span className="text-zinc-300 dark:text-zinc-700">·</span>
            <span>{(step.presetName || log?.presetName || '').toLowerCase()}</span>
          </>
        )}
        {(step.modelName || log?.modelName) && (
          <>
            <span className="text-zinc-300 dark:text-zinc-700">/</span>
            <span>{(step.modelName || log?.modelName || '').toLowerCase()}</span>
          </>
        )}
        {(step.modelRole === 'fallback' || log?.modelRole === 'fallback') && (
          <span className="px-1 py-px rounded-sm bg-amber-500/15 text-amber-600 dark:text-amber-400">fallback</span>
        )}
        {step.finishedAt && step.startedAt && (
          <>
            <span className="text-zinc-300 dark:text-zinc-700">·</span>
            <span className="tabular-nums">{fmtDuration(step.startedAt, step.finishedAt)}</span>
          </>
        )}
      </div>
      {step.args && step.args !== '{}' && (
        <pre className="text-[11px] font-mono text-zinc-500 dark:text-zinc-600 bg-zinc-50 dark:bg-zinc-950 rounded-md px-2.5 py-1.5 overflow-x-auto max-h-24 whitespace-pre-wrap break-all">
          {formatArgs(step.args)}
        </pre>
      )}
    </div>
  )
}

interface PanelBodyProps {
  log: SessionLog | null
  hasLog: boolean
  isRunning: boolean
  stepEntries: ReturnType<typeof stepToEntries>
  stepResultPresent: boolean
}

function PanelBody({ log, hasLog, isRunning, stepEntries, stepResultPresent }: PanelBodyProps) {
  if (hasLog) {
    if (!log) {
      return (
        <div className="font-mono text-xs flex items-center gap-2 text-zinc-500 dark:text-zinc-600 py-2">
          <Loader2 size={11} className="animate-spin" />
          <span>Loading log…</span>
        </div>
      )
    }
    return (
      <>
        {log.entries.map((entry, i) => <EntryLine key={i} entry={entry} />)}
        {log.entries.length === 0 && log.status !== 'running' && (
          <p className="text-zinc-500 dark:text-zinc-600 text-xs font-mono">No entries</p>
        )}
        {(log.status === 'running' || isRunning) && (
          <div className="font-mono text-xs flex items-center gap-2 text-zinc-500 dark:text-zinc-600 py-2">
            <Loader2 size={11} className="animate-spin" />
            <span>{log.entries.length === 0 ? 'Waiting for output…' : 'Running…'}</span>
          </div>
        )}
      </>
    )
  }
  return (
    <>
      {stepEntries.map((entry, i) => <EntryLine key={i} entry={entry} />)}
      {isRunning && !stepResultPresent && (
        <div className="font-mono text-xs flex items-center gap-2 text-zinc-500 dark:text-zinc-600 py-2">
          <Loader2 size={11} className="animate-spin" />
          <span>Running…</span>
        </div>
      )}
    </>
  )
}
