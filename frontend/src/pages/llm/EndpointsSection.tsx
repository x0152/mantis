import { Pencil, Plus, Trash2, Link2, Wallet, Infinity as InfinityIcon } from '@/lib/icons'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Section, EmptyHint } from '@/components/StepperSection'
import type { InferenceLimit, LlmConnection, Model } from '@/types'

interface Props {
  endpoints: LlmConnection[]
  modelsByEndpoint: Map<string, Model[]>
  endpointLimits: Record<string, InferenceLimit>
  onCreate: () => void
  onEdit: (e: LlmConnection) => void
  onDelete: (id: string) => void
}

export function EndpointsSection({ endpoints, modelsByEndpoint, endpointLimits, onCreate, onEdit, onDelete }: Props) {
  return (
    <Section
      n={1}
      icon={Link2}
      title="Endpoints"
      subtitle="Where your LLMs live. Any OpenAI-compatible HTTP API."
      action={<Button size="sm" onClick={onCreate}><Plus size={13} /> Add endpoint</Button>}
    >
      {endpoints.length === 0 ? (
        <EmptyHint>
          Add OpenAI, a local Ollama, LM Studio, or any OpenAI-compatible URL to get started.
        </EmptyHint>
      ) : (
        <div className="rounded-lg border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 divide-y divide-zinc-100 dark:divide-zinc-800">
          {endpoints.map(ep => (
            <EndpointRow
              key={ep.id}
              ep={ep}
              count={modelsByEndpoint.get(ep.id)?.length ?? 0}
              limit={endpointLimits[ep.id] ?? { type: 'unlimited', label: 'No inference limit reported' }}
              onEdit={() => onEdit(ep)}
              onDelete={() => onDelete(ep.id)}
            />
          ))}
        </div>
      )}
    </Section>
  )
}

interface RowProps {
  ep: LlmConnection
  count: number
  limit: InferenceLimit
  onEdit: () => void
  onDelete: () => void
}

function EndpointRow({ ep, count, limit, onEdit, onDelete }: RowProps) {
  const limitType = (limit.type || 'unlimited').toLowerCase()
  const isQuota = limitType === 'quota'
  const isBalance = limitType === 'balance'
  const percent = isQuota ? Math.max(0, Math.min(100, Math.round(limit.percentage ?? 0))) : 100
  const limitLabel = limit.label?.trim() || (isQuota ? 'Quota usage is not available' : 'No inference limit reported')

  return (
    <div className="px-4 py-3">
      <div className="flex items-center gap-3">
        <Link2 size={13} className="text-zinc-400 shrink-0" />
        <span className="font-medium text-sm text-zinc-800 dark:text-zinc-200 shrink-0">{ep.id}</span>
        <Badge variant="muted">{ep.provider}</Badge>
        <span className="text-[11px] text-zinc-500 dark:text-zinc-600 font-mono truncate flex-1 min-w-0">{ep.baseUrl}</span>
        <span className="text-[11px] text-zinc-500 whitespace-nowrap shrink-0">
          {count} {count === 1 ? 'model' : 'models'}
        </span>
        <div className="flex gap-0.5 shrink-0">
          <Button variant="ghost" size="icon" onClick={onEdit}><Pencil size={13} /></Button>
          <Button variant="destructive" size="icon" onClick={onDelete}><Trash2 size={13} /></Button>
        </div>
      </div>
      <div className="mt-2.5">
        {isQuota ? (
          <QuotaMeter percent={percent} label={limitLabel} />
        ) : isBalance ? (
          <div className="flex items-center gap-1.5 text-[11px] text-zinc-600 dark:text-zinc-400">
            <Wallet size={12} className="text-teal-500 shrink-0" />
            <span className="font-medium">{limitLabel}</span>
          </div>
        ) : (
          <div className="flex items-center gap-1.5 text-[11px] text-zinc-500 dark:text-zinc-500">
            <InfinityIcon size={12} className="shrink-0" />
            <span>{limitLabel}</span>
          </div>
        )}
      </div>
    </div>
  )
}

function QuotaMeter({ percent, label }: { percent: number; label: string }) {
  const tone =
    percent >= 85 ? 'bg-rose-500' : percent >= 60 ? 'bg-amber-500' : 'bg-teal-500'
  return (
    <div className="grid grid-cols-[1fr_auto] items-center gap-x-2 gap-y-1">
      <div className="h-2.5 w-full rounded-full bg-zinc-200 dark:bg-zinc-800 border border-zinc-300/70 dark:border-zinc-700/70 overflow-hidden">
        <div className={`h-full ${tone}`} style={{ width: `${percent}%` }} />
      </div>
      <span className="text-[10px] font-mono text-zinc-500 dark:text-zinc-600 tabular-nums">{percent}%</span>
      <p className="col-span-2 text-[10px] text-zinc-500 dark:text-zinc-600">{label}</p>
    </div>
  )
}
