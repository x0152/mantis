import { useState } from 'react'
import { Plus, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'

const paramTypes = ['string', 'number', 'integer', 'boolean'] as const
type ParamType = (typeof paramTypes)[number]

interface Param {
  name: string
  type: ParamType
  description: string
  required: boolean
}

interface Props {
  value: Record<string, unknown>
  onChange: (schema: Record<string, unknown>) => void
}

function fromSchema(schema: Record<string, unknown>): Param[] {
  const props = (schema?.properties ?? {}) as Record<string, { type?: string; description?: string }>
  const req = Array.isArray(schema?.required) ? (schema.required as string[]) : []
  return Object.entries(props).map(([name, p]) => ({
    name,
    type: paramTypes.includes(p.type as ParamType) ? (p.type as ParamType) : 'string',
    description: p.description ?? '',
    required: req.includes(name),
  }))
}

function toSchema(params: Param[]): Record<string, unknown> {
  const properties: Record<string, unknown> = {}
  const required: string[] = []
  for (const p of params) {
    const key = p.name.trim()
    if (!key) continue
    properties[key] = { type: p.type, description: p.description }
    if (p.required) required.push(key)
  }
  const schema: Record<string, unknown> = { type: 'object', properties }
  if (required.length > 0) schema.required = required
  return schema
}

export function ParameterEditor({ value, onChange }: Props) {
  const [params, setParams] = useState<Param[]>(() => fromSchema(value))

  const emit = (next: Param[]) => {
    setParams(next)
    onChange(toSchema(next))
  }

  const add = () => {
    emit([...params, { name: '', type: 'string', description: '', required: false }])
  }

  const update = (idx: number, patch: Partial<Param>) => {
    emit(params.map((p, i) => (i === idx ? { ...p, ...patch } : p)))
  }

  const remove = (idx: number) => {
    emit(params.filter((_, i) => i !== idx))
  }

  return (
    <div className="space-y-2">
      {params.length > 0 && (
        <div className="grid grid-cols-[1fr_100px_1.5fr_auto_auto] gap-x-2 items-center px-1 text-[10px] font-semibold uppercase tracking-wider text-zinc-500 dark:text-zinc-600">
          <span>Name</span>
          <span>Type</span>
          <span>Description</span>
          <span className="text-center w-14">Required</span>
          <span className="w-8" />
        </div>
      )}

      {params.map((p, idx) => (
        <div
          key={idx}
          className="grid grid-cols-[1fr_100px_1.5fr_auto_auto] gap-x-2 items-center rounded-md bg-zinc-50 dark:bg-zinc-800/50 px-1 py-1.5"
        >
          <Input
            value={p.name}
            onChange={e => update(idx, { name: e.target.value.replace(/[^a-zA-Z0-9_]/g, '') })}
            placeholder="param_name"
            className="h-8 text-xs font-mono"
          />
          <select
            value={p.type}
            onChange={e => update(idx, { type: e.target.value as ParamType })}
            className="h-8 rounded-md border border-zinc-300 dark:border-zinc-700 bg-white dark:bg-zinc-800 px-2 text-xs text-zinc-900 dark:text-zinc-100 focus:outline-none focus:border-teal-500/50"
          >
            {paramTypes.map(t => (
              <option key={t} value={t}>{t}</option>
            ))}
          </select>
          <Input
            value={p.description}
            onChange={e => update(idx, { description: e.target.value })}
            placeholder="What this parameter does"
            className="h-8 text-xs"
          />
          <div className="flex justify-center w-14">
            <Switch checked={p.required} onCheckedChange={v => update(idx, { required: v })} />
          </div>
          <Button variant="ghost" size="icon" className="h-8 w-8 text-zinc-400 hover:text-red-500" onClick={() => remove(idx)}>
            <Trash2 size={13} />
          </Button>
        </div>
      ))}

      <Button variant="secondary" size="sm" onClick={add} className="w-full">
        <Plus size={12} /> Add Parameter
      </Button>
    </div>
  )
}
