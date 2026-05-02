import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '@/components/ui/dialog'
import { FormField } from '@/components/FormField'
import type { LlmConnection, Model, ProviderModel } from '@/types'
import type { ModelForm } from './types'
import { DEFAULT_CONTEXT_WINDOW, DEFAULT_RESERVE_TOKENS, SELECT_CLASS } from './types'

interface Props {
  open: boolean
  onOpenChange: (open: boolean) => void
  editing: Model | null
  form: ModelForm
  setForm: React.Dispatch<React.SetStateAction<ModelForm>>
  onSubmit: () => void
  endpoints: LlmConnection[]
  availableModelsByEndpoint: Record<string, ProviderModel[]>
  loadAvailableModels: (connectionId: string, force?: boolean) => void
  loadingAvailableModels: boolean
}

export function ModelDialog({
  open,
  onOpenChange,
  editing,
  form,
  setForm,
  onSubmit,
  endpoints,
  availableModelsByEndpoint,
  loadAvailableModels,
  loadingAvailableModels,
}: Props) {
  const items = availableModelsByEndpoint[form.connectionId] ?? []
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{editing ? 'Edit model' : 'Add model'}</DialogTitle>
          <DialogDescription>The model name exactly as your endpoint expects it.</DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          <FormField label="Endpoint">
            <select
              value={form.connectionId}
              onChange={e => setForm(f => ({ ...f, connectionId: e.target.value }))}
              className={SELECT_CLASS}
            >
              <option value="">Select…</option>
              {endpoints.map(ep => (
                <option key={ep.id} value={ep.id}>{ep.id} · {ep.provider}</option>
              ))}
            </select>
          </FormField>
          <FormField label="Model name" hint="Type manually or pick from loaded endpoint models">
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Input
                  list="available-models"
                  value={form.name}
                  onFocus={() => form.connectionId && loadAvailableModels(form.connectionId, true)}
                  onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                  placeholder="gpt-4o"
                  disabled={!form.connectionId}
                />
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={() => form.connectionId && loadAvailableModels(form.connectionId, true)}
                  disabled={!form.connectionId || loadingAvailableModels}
                >
                  {loadingAvailableModels ? 'Loading...' : 'Refresh'}
                </Button>
              </div>
              <datalist id="available-models">
                {items.map(m => <option key={m.id} value={m.id} />)}
              </datalist>
              <p className="text-[11px] text-zinc-500 dark:text-zinc-600">
                {form.connectionId
                  ? `${items.length} models loaded from ${form.connectionId}. You can still enter any custom model name.`
                  : 'Select endpoint first'}
              </p>
            </div>
          </FormField>
          <FormField label="Reasoning output" hint="How to handle <think> blocks in model responses">
            <select
              value={form.thinkingMode}
              onChange={e => setForm(f => ({ ...f, thinkingMode: e.target.value }))}
              className={SELECT_CLASS}
            >
              <option value="">Keep as-is (default)</option>
              <option value="skip">Remove reasoning blocks</option>
              <option value="inline">Strip tags, keep content</option>
            </select>
          </FormField>
          <FormField label="Context window (tokens)" hint="Model's maximum context length">
            <Input
              type="number"
              min={1000}
              step={1000}
              value={form.contextWindow}
              onChange={e => setForm(f => ({ ...f, contextWindow: e.target.value }))}
              placeholder={String(DEFAULT_CONTEXT_WINDOW)}
            />
          </FormField>
          <FormField label="Reserve for output (tokens)" hint="Headroom left for the reply so prompts never fill the window">
            <Input
              type="number"
              min={0}
              step={1000}
              value={form.reserveTokens}
              onChange={e => setForm(f => ({ ...f, reserveTokens: e.target.value }))}
              placeholder={String(DEFAULT_RESERVE_TOKENS)}
            />
          </FormField>
          <FormField label="Compact at (tokens, optional override)" hint="If empty, uses context window − reserve">
            <Input
              type="number"
              min={0}
              step={1000}
              value={form.compactTokens}
              onChange={e => setForm(f => ({ ...f, compactTokens: e.target.value }))}
              placeholder="auto"
            />
          </FormField>
          <DialogFooter>
            <Button variant="secondary" onClick={() => onOpenChange(false)}>Cancel</Button>
            <Button onClick={onSubmit} disabled={!form.name || !form.connectionId}>
              {editing ? 'Update' : 'Create'}
            </Button>
          </DialogFooter>
        </div>
      </DialogContent>
    </Dialog>
  )
}
