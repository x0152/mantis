import { Gauge, Layers } from 'lucide-react'
import type { ChatMessage } from '../types'

const DEFAULT_THRESHOLD = 100_000

function estimateTokens(text: string | undefined | null): number {
  if (!text) return 0
  return Math.ceil(text.length / 4)
}

function fmtTokens(n: number): string {
  if (n < 1000) return String(n)
  if (n < 10000) return (n / 1000).toFixed(1) + 'k'
  return Math.round(n / 1000) + 'k'
}

export function ContextMeter({
  messages,
  threshold = DEFAULT_THRESHOLD,
  serverTokens,
  compactionCount = 0,
  partial = false,
}: {
  messages: ChatMessage[]
  threshold?: number
  serverTokens?: number
  compactionCount?: number
  partial?: boolean
}) {
  const clientTotal = messages.reduce(
    (sum, m) => sum + (m.tokens ?? estimateTokens(m.content)),
    0,
  )
  const total = serverTokens && serverTokens > 0 ? serverTokens : clientTotal
  const isServer = serverTokens !== undefined && serverTokens > 0
  const pct = Math.min(100, Math.round((total / threshold) * 100))

  const color =
    pct >= 90
      ? 'bg-red-500'
      : pct >= 70
        ? 'bg-amber-500'
        : 'bg-teal-500'

  const tooltip = [
    isServer
      ? 'Last prompt size reported by the provider.'
      : partial
        ? 'Sum of loaded messages (older ones may not be counted).'
        : 'Approximate context usage.',
    `Threshold = compact limit (${fmtTokens(threshold)}).`,
    compactionCount > 0 ? `Compacted ${compactionCount} time${compactionCount === 1 ? '' : 's'}.` : '',
  ]
    .filter(Boolean)
    .join(' ')

  return (
    <div className="flex items-center gap-2 text-[11px] text-zinc-500 dark:text-zinc-500" title={tooltip}>
      <Gauge size={12} className="shrink-0" />
      <span className="font-mono tabular-nums shrink-0">
        {!isServer && partial && '≥'}
        {isServer ? '' : '~'}
        {fmtTokens(total)} / {fmtTokens(threshold)}
      </span>
      <div className="flex-1 min-w-16 max-w-40 h-1 rounded-full bg-zinc-200 dark:bg-zinc-800 overflow-hidden">
        <div
          className={`h-full ${color} transition-all`}
          style={{ width: `${Math.max(2, pct)}%` }}
        />
      </div>
      <span className="font-mono tabular-nums shrink-0 w-8 text-right">{pct}%</span>
      {compactionCount > 0 && (
        <span
          className="flex items-center gap-0.5 shrink-0 text-teal-600 dark:text-teal-400"
          title={`${compactionCount} compaction${compactionCount === 1 ? '' : 's'}`}
        >
          <Layers size={10} />
          <span className="font-mono tabular-nums">×{compactionCount}</span>
        </span>
      )}
    </div>
  )
}
