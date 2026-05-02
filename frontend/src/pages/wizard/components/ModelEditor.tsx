import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type { ProviderModel } from '@/types'
import type { ModelRow } from '../types'

interface ModelEditorProps {
  rows: ModelRow[]
  onChange: (rows: ModelRow[]) => void
  available?: ProviderModel[] | null
  listId?: string
}

export function ModelEditor({ rows, onChange, available, listId }: ModelEditorProps) {
  const datalistId = listId ?? 'wizard-available-models'
  const hasAvailable = !!available && available.length > 0
  return (
    <div>
      <Label>Models</Label>
      <div className="space-y-2 mt-1.5">
        {rows.map((row, i) => (
          <div key={i} className="flex items-center gap-2">
            <Input
              list={hasAvailable ? datalistId : undefined}
              value={row.name}
              onChange={e => onChange(rows.map((r, j) => (j === i ? { ...r, name: e.target.value } : r)))}
              placeholder={hasAvailable ? 'pick or type a model' : 'model name'}
              className="flex-1"
            />
            <select
              value={row.role}
              onChange={e => {
                const newRole = e.target.value as ModelRow['role']
                onChange(
                  rows.map((r, j) => {
                    if (j === i) return { ...r, role: newRole }
                    if (newRole && r.role === newRole) return { ...r, role: '' }
                    return r
                  }),
                )
              }}
              className="w-28 shrink-0 rounded-lg border border-zinc-300 dark:border-zinc-700 bg-white dark:bg-zinc-800 px-2 py-2 text-xs text-zinc-700 dark:text-zinc-300 focus:outline-none focus:border-teal-500/50"
            >
              <option value="">—</option>
              <option value="chat">chat</option>
              <option value="summary">summary</option>
              <option value="vision">vision</option>
            </select>
            {rows.length > 1 && (
              <button
                type="button"
                onClick={() => onChange(rows.filter((_, j) => j !== i))}
                className="text-zinc-400 hover:text-rose-500 text-sm px-1"
                aria-label="Remove model"
              >
                ×
              </button>
            )}
          </div>
        ))}
        <button
          type="button"
          onClick={() => onChange([...rows, { name: '', role: '' }])}
          className="text-[12px] text-teal-600 dark:text-teal-400 hover:underline"
        >
          + Add model
        </button>
      </div>
      {hasAvailable && (
        <datalist id={datalistId}>
          {available!.map(m => (
            <option key={m.id} value={m.id} />
          ))}
        </datalist>
      )}
      <p className="text-[11px] text-zinc-500 dark:text-zinc-600 mt-1.5">
        <span className="font-medium">chat</span> — for conversations (required).{' '}
        <span className="font-medium">summary</span> — for chat titles.{' '}
        <span className="font-medium">vision</span> — for images.
      </p>
    </div>
  )
}
