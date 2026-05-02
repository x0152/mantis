import { ChevronDown } from '@/lib/icons'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '@/components/ui/dialog'
import { FormField } from '@/components/FormField'
import type { Connection, GuardProfile, Preset } from '@/types'
import type { ServerForm, SshConfig } from './types'
import { SELECT_CLASS } from './types'

interface Props {
  open: boolean
  onOpenChange: (open: boolean) => void
  editing: Connection | null
  form: ServerForm
  setForm: React.Dispatch<React.SetStateAction<ServerForm>>
  ssh: SshConfig
  setSsh: React.Dispatch<React.SetStateAction<SshConfig>>
  presets: Preset[]
  profiles: GuardProfile[]
  profileDropdownOpen: boolean
  setProfileDropdownOpen: (v: boolean | ((prev: boolean) => boolean)) => void
  profileName: (id: string) => string
  toggleProfileSelection: (id: string) => void
  onSubmit: () => void
}

export function ServerDialog({
  open,
  onOpenChange,
  editing,
  form,
  setForm,
  ssh,
  setSsh,
  presets,
  profiles,
  profileDropdownOpen,
  setProfileDropdownOpen,
  profileName,
  toggleProfileSelection,
  onSubmit,
}: Props) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{editing ? 'Edit server' : 'New server'}</DialogTitle>
          <DialogDescription>
            SSH credentials for a machine you control. Sandboxes are managed separately and don’t need this form.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          <FormField label="Name">
            <Input
              value={form.name}
              onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
              placeholder="my-prod-box"
            />
          </FormField>
          <FormField label="Description" hint="What this server is used for. Helps the assistant pick the right one.">
            <Textarea
              value={form.description}
              onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
              className="h-20"
              placeholder="Production web server, backups every night…"
            />
          </FormField>
          <FormField label="Type">
            <select
              value={form.type}
              onChange={e => setForm(f => ({ ...f, type: e.target.value }))}
              className={SELECT_CLASS}
            >
              <option value="ssh">SSH</option>
            </select>
          </FormField>
          <SshFields ssh={ssh} setSsh={setSsh} />
          <FormField label="Preset" hint="Which AI profile this server uses. Inherits the global default if left empty.">
            <select
              value={form.presetId}
              onChange={e => setForm(f => ({ ...f, presetId: e.target.value }))}
              className={SELECT_CLASS}
            >
              <option value="">Inherit (global server preset)</option>
              {presets.map(p => (
                <option key={p.id} value={p.id}>{p.name}</option>
              ))}
            </select>
          </FormField>
          <label
            className={`flex items-center gap-2.5 p-2.5 rounded-lg border cursor-pointer ${
              form.memoryEnabled ? 'border-teal-500/40 bg-teal-500/5' : 'border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900'
            }`}
          >
            <Checkbox checked={form.memoryEnabled} onCheckedChange={v => setForm(f => ({ ...f, memoryEnabled: !!v }))} />
            <div>
              <span className="text-sm font-medium text-zinc-800 dark:text-zinc-200">Memory</span>
              <p className="text-[11px] text-zinc-500 mt-0.5">
                Auto-extract and remember facts about this server (paths, conventions, quirks).
              </p>
            </div>
          </label>
          <ProfileSelector
            form={form}
            profiles={profiles}
            isOpen={profileDropdownOpen}
            setOpen={setProfileDropdownOpen}
            profileName={profileName}
            toggleProfileSelection={toggleProfileSelection}
          />
          <DialogFooter>
            <Button variant="secondary" onClick={() => onOpenChange(false)}>Cancel</Button>
            <Button onClick={onSubmit} disabled={!form.name}>
              {editing ? 'Update' : 'Create'}
            </Button>
          </DialogFooter>
        </div>
      </DialogContent>
    </Dialog>
  )
}

function SshFields({ ssh, setSsh }: { ssh: SshConfig; setSsh: React.Dispatch<React.SetStateAction<SshConfig>> }) {
  return (
    <div className="space-y-3 p-3.5 bg-zinc-50 dark:bg-zinc-950 rounded-lg border border-zinc-200 dark:border-zinc-800">
      <p className="text-[11px] font-semibold text-zinc-500 uppercase tracking-wider">SSH configuration</p>
      <div className="grid grid-cols-3 gap-2.5">
        <div className="col-span-2">
          <label className="block text-[11px] font-medium text-zinc-500 mb-1">host</label>
          <Input
            value={ssh.host}
            onChange={e => setSsh(s => ({ ...s, host: e.target.value }))}
            placeholder="192.168.1.100 or example.com"
          />
        </div>
        <div>
          <label className="block text-[11px] font-medium text-zinc-500 mb-1">port</label>
          <Input value={ssh.port} onChange={e => setSsh(s => ({ ...s, port: e.target.value }))} placeholder="22" />
        </div>
      </div>
      <div>
        <label className="block text-[11px] font-medium text-zinc-500 mb-1">username</label>
        <Input value={ssh.username} onChange={e => setSsh(s => ({ ...s, username: e.target.value }))} placeholder="root" />
      </div>
      <div>
        <label className="block text-[11px] font-medium text-zinc-500 mb-1">password</label>
        <Input
          type="password"
          value={ssh.password}
          onChange={e => setSsh(s => ({ ...s, password: e.target.value }))}
          placeholder="optional"
        />
      </div>
      <div>
        <label className="block text-[11px] font-medium text-zinc-500 mb-1">privateKey</label>
        <Textarea
          value={ssh.privateKey}
          onChange={e => setSsh(s => ({ ...s, privateKey: e.target.value }))}
          className="h-20 font-mono"
          placeholder="-----BEGIN OPENSSH PRIVATE KEY-----"
        />
      </div>
    </div>
  )
}

interface ProfileSelectorProps {
  form: ServerForm
  profiles: GuardProfile[]
  isOpen: boolean
  setOpen: (v: boolean | ((prev: boolean) => boolean)) => void
  profileName: (id: string) => string
  toggleProfileSelection: (id: string) => void
}

function ProfileSelector({ form, profiles, isOpen, setOpen, profileName, toggleProfileSelection }: ProfileSelectorProps) {
  return (
    <div>
      <label className="block text-xs font-medium text-zinc-500 dark:text-zinc-400 mb-1.5">Security profiles</label>
      <div className="relative">
        <button
          type="button"
          onClick={() => setOpen(o => !o)}
          className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded-lg text-sm bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 focus:outline-none focus:border-teal-500/50 text-left flex items-center justify-between"
        >
          <span className={form.profileIds.length === 0 ? 'text-zinc-500' : ''}>
            {form.profileIds.length === 0
              ? 'No profiles (unrestricted)'
              : form.profileIds.map(id => profileName(id)).join(', ')}
          </span>
          <ChevronDown
            size={14}
            className={`text-zinc-500 transition-transform ${isOpen ? 'rotate-180' : ''}`}
          />
        </button>
        {isOpen && (
          <>
            <div className="fixed inset-0 z-10" onClick={() => setOpen(false)} />
            <div className="absolute z-20 mt-1 w-full bg-white dark:bg-zinc-800 border border-zinc-200 dark:border-zinc-700 rounded-lg shadow-xl py-1 max-h-48 overflow-y-auto">
              {profiles.map(p => (
                <label
                  key={p.id}
                  className="flex items-center gap-2 px-3 py-1.5 hover:bg-zinc-100 dark:hover:bg-zinc-700/50 cursor-pointer text-sm text-zinc-700 dark:text-zinc-300"
                >
                  <Checkbox
                    checked={form.profileIds.includes(p.id)}
                    onCheckedChange={() => toggleProfileSelection(p.id)}
                  />
                  {p.name}{p.builtin ? ' (built-in)' : ''}
                </label>
              ))}
            </div>
          </>
        )}
      </div>
    </div>
  )
}
