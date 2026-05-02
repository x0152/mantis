import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '@/components/ui/dialog'
import { FormField } from '@/components/FormField'
import type { LlmConnection } from '@/types'
import type { EndpointForm } from './types'
import { SELECT_CLASS } from './types'

interface Props {
  open: boolean
  onOpenChange: (open: boolean) => void
  editing: LlmConnection | null
  form: EndpointForm
  setForm: React.Dispatch<React.SetStateAction<EndpointForm>>
  onSubmit: () => void
}

export function EndpointDialog({ open, onOpenChange, editing, form, setForm, onSubmit }: Props) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{editing ? 'Edit endpoint' : 'New endpoint'}</DialogTitle>
          <DialogDescription>Any OpenAI-compatible URL or a Gonka Source URL.</DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          <FormField label="ID" hint="Short identifier, e.g. openai, local, anthropic">
            <Input
              value={form.id}
              disabled={!!editing}
              onChange={e => setForm(f => ({ ...f, id: e.target.value }))}
              placeholder="openai"
            />
          </FormField>
          <FormField label="Provider">
            <select
              value={form.provider}
              onChange={e => setForm(f => ({ ...f, provider: e.target.value }))}
              className={SELECT_CLASS}
            >
              <option value="openai">openai</option>
              <option value="gonka">gonka</option>
            </select>
          </FormField>
          <FormField
            label={form.provider === 'gonka' ? 'Node URL' : 'Base URL'}
            hint={form.provider === 'gonka' ? 'Gonka Source URL (NODE_URL)' : undefined}
          >
            <Input
              value={form.baseUrl}
              onChange={e => setForm(f => ({ ...f, baseUrl: e.target.value }))}
              placeholder={form.provider === 'gonka' ? 'https://api.gonka.testnet.example.com' : 'https://api.openai.com/v1'}
            />
          </FormField>
          <FormField
            label={form.provider === 'gonka' ? 'Private key' : 'API key'}
            hint={form.provider === 'gonka' ? 'GONKA_PRIVATE_KEY wallet key' : 'Leave empty for local providers'}
          >
            <Input
              type="password"
              value={form.apiKey}
              onChange={e => setForm(f => ({ ...f, apiKey: e.target.value }))}
              placeholder={form.provider === 'gonka' ? '0x…' : 'sk-…'}
            />
          </FormField>
          <DialogFooter>
            <Button variant="secondary" onClick={() => onOpenChange(false)}>Cancel</Button>
            <Button
              onClick={onSubmit}
              disabled={!form.id || !form.baseUrl || (form.provider === 'gonka' && !form.apiKey)}
            >
              {editing ? 'Update' : 'Create'}
            </Button>
          </DialogFooter>
        </div>
      </DialogContent>
    </Dialog>
  )
}
