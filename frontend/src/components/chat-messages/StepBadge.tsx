import { AlertCircle, CheckCircle2, Loader2, Square, Wrench } from '@/lib/icons'
import { navigate } from '../../router'
import type { Step } from '../../types'
import { STEP_ICONS, planIdFromStepArgs, stepArgsSummary } from './stepHelpers'
import { useTicker } from './useTicker'
import { fmtElapsed } from './utils'

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
    if (planId) navigate({ page: 'plans', planId })
    onClick()
  }

  return (
    <button
      onClick={handleClick}
      className={`inline-flex items-start gap-1.5 px-2 py-1 rounded-sm font-mono text-[11px] lowercase tracking-tight cursor-pointer transition-colors max-w-full text-left step-enter ${
        isRunning
          ? 'text-teal-600 dark:text-teal-400 border border-teal-500/40 bg-teal-500/5 step-running'
          : isError
            ? 'bg-red-500/10 text-red-400 border border-red-500/20'
            : isCancelled
              ? 'bg-zinc-200/40 dark:bg-zinc-800/40 text-zinc-400 dark:text-zinc-600 border border-zinc-300/40 dark:border-zinc-700/40 line-through decoration-zinc-400/50'
              : 'bg-zinc-100/70 dark:bg-zinc-800/50 text-zinc-500 dark:text-zinc-500 border border-zinc-200 dark:border-zinc-700/70 hover:text-zinc-700 dark:hover:text-zinc-300'
      }`}
    >
      {isRunning ? (
        <Loader2 size={11} className="animate-spin shrink-0 mt-0.5" />
      ) : isError ? (
        <AlertCircle size={11} className="shrink-0 mt-0.5" />
      ) : isCancelled ? (
        <Square size={10} className="fill-current opacity-70 shrink-0 mt-1" />
      ) : (
        <Icon size={11} className="shrink-0 mt-0.5" />
      )}
      <span className="break-words whitespace-normal leading-snug min-w-0 flex-1">
        {step.label.toLowerCase()}
        {argSummary && <span className="ml-1.5 opacity-60 normal-case">{argSummary}</span>}
      </span>
      {showElapsed && (
        <span className="opacity-70 tabular-nums shrink-0 mt-0.5">{fmtElapsed(elapsed)}</span>
      )}
      {!isRunning && !isError && !isCancelled && <CheckCircle2 size={10} className="text-emerald-500 shrink-0 mt-1" />}
    </button>
  )
}
