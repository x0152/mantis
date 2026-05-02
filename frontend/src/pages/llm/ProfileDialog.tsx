import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '@/components/ui/dialog'
import { FormField } from '@/components/FormField'
import type { LlmConnection, Model, Preset } from '@/types'
import type { ProfileForm } from './types'
import { ROLE_DEFS, SELECT_CLASS } from './types'

interface Props {
  open: boolean
  onOpenChange: (open: boolean) => void
  editing: Preset | null
  form: ProfileForm
  setForm: React.Dispatch<React.SetStateAction<ProfileForm>>
  onSubmit: () => void
  models: Model[]
  endpointById: Map<string, LlmConnection>
}

export function ProfileDialog({ open, onOpenChange, editing, form, setForm, onSubmit, models, endpointById }: Props) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-xl">
        <DialogHeader>
          <DialogTitle>{editing ? 'Edit profile' : 'New profile'}</DialogTitle>
          <DialogDescription>Assign a model to each role, then tune behavior.</DialogDescription>
        </DialogHeader>
        <div className="space-y-5">
          <FormField label="Name">
            <Input
              value={form.name}
              onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
              placeholder="e.g. Fast, Smart, Creative"
            />
          </FormField>

          <div>
            <div className="text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-2">Roles</div>
            <div className="rounded-lg border border-zinc-200 dark:border-zinc-800 divide-y divide-zinc-200 dark:divide-zinc-800 bg-white dark:bg-zinc-900">
              {ROLE_DEFS.map(({ key, field, icon: Icon, label, hint, required }) => {
                const emptyLabel = key === 'summary' ? 'Same as chat' : 'None'
                return (
                  <div key={key} className="px-3 py-3 grid grid-cols-[7.5rem_1fr] gap-x-3 gap-y-1 items-center">
                    <div className="flex items-center gap-2">
                      <Icon size={14} className="text-teal-500" />
                      <span className="text-sm font-medium text-zinc-800 dark:text-zinc-200">{label}</span>
                      {required && <span className="text-[10px] text-rose-500">*</span>}
                    </div>
                    <select
                      value={form[field]}
                      onChange={e => setForm(f => ({ ...f, [field]: e.target.value }))}
                      className={SELECT_CLASS}
                    >
                      <option value="">{required ? 'Select a model…' : emptyLabel}</option>
                      {models.map(m => {
                        const ep = endpointById.get(m.connectionId)
                        return (
                          <option key={m.id} value={m.id}>
                            {m.name}{ep ? ` · ${ep.id}` : ''}
                          </option>
                        )
                      })}
                    </select>
                    <div />
                    <p className="text-[11px] text-zinc-500 dark:text-zinc-600">{hint}</p>
                  </div>
                )
              })}
            </div>
          </div>

          <div>
            <div className="text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-2">Behavior</div>
            <div className="space-y-3">
              <FormField label="Temperature" hint="0 = deterministic · 1 = creative · empty = model default">
                <Input
                  type="number" step="0.1" min="0" max="2"
                  value={form.temperature}
                  onChange={e => setForm(f => ({ ...f, temperature: e.target.value }))}
                  placeholder="e.g. 0.7"
                />
              </FormField>
              <FormField label="System prompt" hint="Custom instructions added to every conversation. Optional.">
                <Textarea
                  value={form.systemPrompt}
                  onChange={e => setForm(f => ({ ...f, systemPrompt: e.target.value }))}
                  className="h-24"
                  placeholder="You are a helpful assistant…"
                />
              </FormField>
            </div>
          </div>

          <DialogFooter>
            <Button variant="secondary" onClick={() => onOpenChange(false)}>Cancel</Button>
            <Button onClick={onSubmit} disabled={!form.name || !form.chatModelId}>
              {editing ? 'Update' : 'Create'}
            </Button>
          </DialogFooter>
        </div>
      </DialogContent>
    </Dialog>
  )
}
