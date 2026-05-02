import { toast } from 'sonner'
import { Box, Copy, Link2, Pencil, Plus, Trash2 } from '@/lib/icons'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Section, EmptyHint } from '@/components/StepperSection'
import type { LlmConnection, Model } from '@/types'

interface Props {
  endpoints: LlmConnection[]
  models: Model[]
  modelsByEndpoint: Map<string, Model[]>
  onCreate: (endpointId?: string) => void
  onEdit: (m: Model) => void
  onDelete: (id: string) => void
}

export function ModelsSection({ endpoints, models, modelsByEndpoint, onCreate, onEdit, onDelete }: Props) {
  return (
    <Section
      n={2}
      icon={Box}
      title="Models"
      subtitle="The exact model names your endpoints serve."
      disabled={endpoints.length === 0}
      disabledHint="Add an endpoint first — models belong to an endpoint."
      action={
        <Button size="sm" onClick={() => onCreate()} disabled={endpoints.length === 0}>
          <Plus size={13} /> Add model
        </Button>
      }
    >
      {models.length === 0 ? (
        <EmptyHint>
          Register the models your endpoint serves, e.g.{' '}
          <code className="font-mono text-[11px] bg-zinc-100 dark:bg-zinc-800 rounded px-1 py-0.5">gpt-4o</code> or{' '}
          <code className="font-mono text-[11px] bg-zinc-100 dark:bg-zinc-800 rounded px-1 py-0.5">llama3</code>.
        </EmptyHint>
      ) : (
        <div className="space-y-3">
          {endpoints.map(ep => (
            <ModelGroup
              key={ep.id}
              ep={ep}
              epModels={modelsByEndpoint.get(ep.id) ?? []}
              onAdd={() => onCreate(ep.id)}
              onEdit={onEdit}
              onDelete={onDelete}
            />
          ))}
        </div>
      )}
    </Section>
  )
}

interface GroupProps {
  ep: LlmConnection
  epModels: Model[]
  onAdd: () => void
  onEdit: (m: Model) => void
  onDelete: (id: string) => void
}

function ModelGroup({ ep, epModels, onAdd, onEdit, onDelete }: GroupProps) {
  return (
    <div>
      <div className="flex items-center justify-between px-1 mb-1.5">
        <div className="flex items-center gap-1.5 text-[11px] font-semibold uppercase tracking-wider text-zinc-500 dark:text-zinc-500">
          <Link2 size={11} />
          {ep.id}
          <span className="text-zinc-400 dark:text-zinc-600 normal-case font-normal tracking-normal">
            · {epModels.length}
          </span>
        </div>
        <button
          onClick={onAdd}
          className="inline-flex items-center gap-1 text-[11px] font-medium text-zinc-500 hover:text-teal-500"
        >
          <Plus size={11} /> Add to {ep.id}
        </button>
      </div>
      {epModels.length === 0 ? (
        <div className="rounded-lg border border-dashed border-zinc-300 dark:border-zinc-700 px-4 py-3 text-[11px] text-zinc-500 dark:text-zinc-500">
          No models yet on this endpoint.
        </div>
      ) : (
        <div className="rounded-lg border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 divide-y divide-zinc-100 dark:divide-zinc-800">
          {epModels.map(m => (
            <ModelRow key={m.id} model={m} onEdit={() => onEdit(m)} onDelete={() => onDelete(m.id)} />
          ))}
        </div>
      )}
    </div>
  )
}

function ModelRow({ model, onEdit, onDelete }: { model: Model; onEdit: () => void; onDelete: () => void }) {
  return (
    <div className="group px-4 py-2.5 flex items-center gap-2">
      <Box size={12} className="text-zinc-400 shrink-0" />
      <span className="text-sm text-zinc-800 dark:text-zinc-200 font-medium truncate">{model.name}</span>
      {model.thinkingMode && <Badge variant="secondary">{model.thinkingMode}</Badge>}
      <button
        onClick={() => {
          navigator.clipboard.writeText(model.id)
          toast.success('ID copied')
        }}
        className="inline-flex items-center gap-1 text-[10px] text-zinc-400 font-mono hover:text-zinc-600 dark:hover:text-zinc-400 opacity-0 group-hover:opacity-100 transition-opacity"
        title={model.id}
      >
        {model.id.slice(0, 8)}… <Copy size={10} />
      </button>
      <div className="flex-1" />
      <div className="flex gap-0.5 shrink-0">
        <Button variant="ghost" size="icon" onClick={onEdit}><Pencil size={13} /></Button>
        <Button variant="destructive" size="icon" onClick={onDelete}><Trash2 size={13} /></Button>
      </div>
    </div>
  )
}
