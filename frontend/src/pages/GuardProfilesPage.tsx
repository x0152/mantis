import { useState, useEffect, useCallback } from 'react'
import { Plus, Pencil, Trash2, Shield, Copy, ChevronDown, ChevronRight, X } from 'lucide-react'
import { toast } from 'sonner'
import { api } from '../api'
import type { GuardProfile, GuardCapabilities, CommandRule } from '../types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/dialog'
import { EmptyState } from '@/components/EmptyState'
import { FormField } from '@/components/FormField'
import { ConfirmDelete } from '@/components/ConfirmDelete'

const defaultCaps: GuardCapabilities = {
  pipes: false, redirects: false, cmdSubst: false, background: false,
  sudo: false, codeExec: false, download: false, install: false,
  writeFs: false, networkOut: false, cron: false, unrestricted: false,
}

const capLabels: Record<keyof GuardCapabilities, string> = {
  pipes: 'Pipes (|)', redirects: 'Redirects (>, >>)', cmdSubst: 'Command substitution $()',
  background: 'Background (&)', sudo: 'Sudo', codeExec: 'Code execution (bash -c)',
  download: 'Download (curl, wget)', install: 'Package install (apt, pip)',
  writeFs: 'Filesystem writes (cp, mv)', networkOut: 'Outbound network',
  cron: 'Cron / scheduling', unrestricted: 'Unrestricted (allow everything)',
}

export default function GuardProfilesPage() {
  const [profiles, setProfiles] = useState<GuardProfile[]>([])
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<GuardProfile | null>(null)
  const [expanded, setExpanded] = useState<Set<string>>(new Set())
  const [form, setForm] = useState<{ name: string; description: string; capabilities: GuardCapabilities; commands: CommandRule[] }>({
    name: '', description: '', capabilities: { ...defaultCaps }, commands: [],
  })
  const [newCmd, setNewCmd] = useState('')
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null)

  const load = useCallback(async () => {
    try {
      setLoading(true)
      setProfiles(await api.guardProfiles.list())
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Failed to load')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])

  const toggle = (id: string) => setExpanded(prev => {
    const next = new Set(prev)
    next.has(id) ? next.delete(id) : next.add(id)
    return next
  })

  const openCreate = () => {
    setEditing(null)
    setForm({ name: '', description: '', capabilities: { ...defaultCaps }, commands: [] })
    setNewCmd('')
    setModalOpen(true)
  }

  const openEdit = (p: GuardProfile) => {
    setEditing(p)
    setForm({ name: p.name, description: p.description, capabilities: { ...p.capabilities }, commands: [...p.commands] })
    setNewCmd('')
    setModalOpen(true)
  }

  const openClone = (p: GuardProfile) => {
    setEditing(null)
    setForm({ name: p.name + ' (copy)', description: p.description, capabilities: { ...p.capabilities }, commands: [...p.commands] })
    setNewCmd('')
    setModalOpen(true)
  }

  const submit = async () => {
    try {
      const data = { name: form.name, description: form.description, capabilities: form.capabilities, commands: form.commands }
      if (editing) {
        await api.guardProfiles.update(editing.id, data)
        toast.success('Profile updated')
      } else {
        await api.guardProfiles.create(data)
        toast.success('Profile created')
      }
      setModalOpen(false)
      load()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Operation failed')
    }
  }

  const remove = async (id: string) => {
    try {
      await api.guardProfiles.delete(id)
      toast.success('Profile deleted')
      setDeleteTarget(null)
      load()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Delete failed')
    }
  }

  const parseCommandRule = (input: string): CommandRule | null => {
    const s = input.trim()
    if (!s) return null
    const bracketMatch = s.match(/^([^\[(\s]+)\[([^\]]*)\]$/)
    if (bracketMatch) {
      const args = bracketMatch[2].split(',').map(a => a.trim()).filter(Boolean)
      return { command: bracketMatch[1], allowedArgs: args.length ? args : undefined }
    }
    const parenMatch = s.match(/^([^\[(\s]+)\(([^)]*)\)$/)
    if (parenMatch) {
      const sql = parenMatch[2].split(',').map(a => a.trim()).filter(Boolean)
      return { command: parenMatch[1], allowedSql: sql.length ? sql : undefined }
    }
    return { command: s }
  }

  const formatCommandRule = (c: CommandRule): string => {
    if (c.allowedArgs?.length) return `${c.command}[${c.allowedArgs.join(',')}]`
    if (c.allowedSql?.length) return `${c.command}(${c.allowedSql.join(',')})`
    return c.command
  }

  const addCommand = () => {
    const rule = parseCommandRule(newCmd)
    if (!rule) return
    if (form.commands.some(c => c.command === rule.command)) return
    setForm(f => ({ ...f, commands: [...f.commands, rule] }))
    setNewCmd('')
  }

  const removeCommand = (idx: number) => {
    setForm(f => ({ ...f, commands: f.commands.filter((_, i) => i !== idx) }))
  }

  const setCap = (key: keyof GuardCapabilities, val: boolean) => {
    setForm(f => ({ ...f, capabilities: { ...f.capabilities, [key]: val } }))
  }

  const capList = (caps: GuardCapabilities) => {
    if (caps.unrestricted) return 'Unrestricted'
    const on = (Object.keys(capLabels) as (keyof GuardCapabilities)[]).filter(k => k !== 'unrestricted' && caps[k])
    return on.length > 0 ? on.map(k => capLabels[k]).join(', ') : 'None'
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-5">
        <div>
          <h1 className="text-lg font-semibold text-zinc-900 dark:text-zinc-100">Guard Profiles</h1>
          <p className="text-xs text-zinc-500 dark:text-zinc-600 mt-0.5">Security profiles control which commands SSH agents can execute</p>
        </div>
        <Button size="sm" onClick={openCreate}>
          <Plus size={14} /> New Profile
        </Button>
      </div>

      {loading ? (
        <div className="text-center py-12 text-zinc-500 dark:text-zinc-600 text-sm">Loading...</div>
      ) : profiles.length === 0 ? (
        <EmptyState icon={Shield} title="No guard profiles yet" description="Create profiles to control SSH command execution" />
      ) : (
        <div className="space-y-2.5">
          {profiles.map(p => (
            <div key={p.id} className="bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800 overflow-hidden">
              <div className="px-4 py-3.5">
                <div className="flex items-center justify-between">
                  <button onClick={() => toggle(p.id)} className="flex items-center gap-2 min-w-0 text-left">
                    {expanded.has(p.id) ? <ChevronDown size={14} className="text-zinc-600" /> : <ChevronRight size={14} className="text-zinc-600" />}
                    <span className="font-medium text-zinc-800 dark:text-zinc-200 text-sm">{p.name}</span>
                    {p.builtin && <Badge variant="secondary">Built-in</Badge>}
                    {p.capabilities.unrestricted && <Badge variant="warning">Unrestricted</Badge>}
                  </button>
                  <div className="flex gap-0.5 ml-3">
                    <Button variant="ghost" size="icon" onClick={() => openClone(p)} title="Clone">
                      <Copy size={14} />
                    </Button>
                    <Button variant="ghost" size="icon" onClick={() => openEdit(p)}>
                      <Pencil size={14} />
                    </Button>
                    {!p.builtin && (
                      <Button variant="destructive" size="icon" onClick={() => setDeleteTarget(p.id)}>
                        <Trash2 size={14} />
                      </Button>
                    )}
                  </div>
                </div>
                  {p.description && <p className="text-xs text-zinc-600 dark:text-zinc-500 mt-1 ml-5.5">{p.description}</p>}
              </div>
              {expanded.has(p.id) && (
                <div className="px-4 pb-4 space-y-3">
                  <div className="bg-zinc-50 dark:bg-zinc-950 rounded-lg border border-zinc-200 dark:border-zinc-800 p-3">
                    <p className="text-[11px] font-semibold text-zinc-500 dark:text-zinc-600 uppercase tracking-wider mb-2">Capabilities</p>
                    <p className="text-xs text-zinc-600 dark:text-zinc-400">{capList(p.capabilities)}</p>
                  </div>
                  <div className="bg-zinc-50 dark:bg-zinc-950 rounded-lg border border-zinc-200 dark:border-zinc-800 p-3">
                    <p className="text-[11px] font-semibold text-zinc-500 dark:text-zinc-600 uppercase tracking-wider mb-2">Commands ({p.commands.length})</p>
                    {p.commands.length === 0 ? (
                      <p className="text-xs text-zinc-500 dark:text-zinc-600 italic">{p.capabilities.unrestricted ? 'All commands allowed' : 'No commands configured'}</p>
                    ) : (
                      <div className="flex flex-wrap gap-1.5">
                        {p.commands.map((c, i) => (
                          <span key={i} className="px-2 py-0.5 text-xs font-mono bg-zinc-200 dark:bg-zinc-800 text-zinc-700 dark:text-zinc-300 rounded" title={
                            (c.allowedArgs?.length ? `Args: ${c.allowedArgs.join(', ')}` : '') +
                            (c.allowedSql?.length ? `SQL: ${c.allowedSql.join(', ')}` : '')
                          }>
                            {c.command}
                            {c.allowedArgs?.length ? <span className="text-zinc-600 ml-1">[{c.allowedArgs.join(',')}]</span> : null}
                            {c.allowedSql?.length ? <span className="text-zinc-600 ml-1">(SQL:{c.allowedSql.length})</span> : null}
                          </span>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      <Dialog open={modalOpen} onOpenChange={setModalOpen}>
        <DialogContent className="max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>{editing ? 'Edit Profile' : 'New Profile'}</DialogTitle>
            <DialogDescription />
          </DialogHeader>
          <div className="space-y-4">
            <FormField label="Name">
              <Input value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                placeholder="My Custom Profile" />
            </FormField>
            <FormField label="Description">
              <Input value={form.description} onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
                placeholder="Read-only monitoring for production" />
            </FormField>
            <div>
              <label className={`flex items-center gap-2.5 p-2.5 rounded-lg border cursor-pointer mb-3 ${
                form.capabilities.unrestricted ? 'border-amber-500/40 bg-amber-500/5' : 'border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-950'
              }`}>
                <Checkbox checked={form.capabilities.unrestricted} onCheckedChange={v => setCap('unrestricted', !!v)} />
                <div>
                  <span className="text-sm font-medium text-zinc-200">Unrestricted</span>
                  <p className="text-[11px] text-zinc-500 mt-0.5">Allow all commands and capabilities</p>
                </div>
              </label>
              {!form.capabilities.unrestricted && (
                <>
                  <label className="block text-xs font-medium text-zinc-500 dark:text-zinc-400 mb-2">Capabilities</label>
                  <div className="grid grid-cols-2 gap-2">
                    {(Object.keys(capLabels) as (keyof GuardCapabilities)[]).filter(k => k !== 'unrestricted').map(key => (
                      <label key={key} className="flex items-center gap-2 text-xs text-zinc-700 dark:text-zinc-300 cursor-pointer">
                        <Checkbox checked={form.capabilities[key]} onCheckedChange={v => setCap(key, !!v)} />
                        {capLabels[key]}
                      </label>
                    ))}
                  </div>
                </>
              )}
            </div>
            {!form.capabilities.unrestricted && (
              <div>
                <label className="block text-xs font-medium text-zinc-500 dark:text-zinc-400 mb-2">Allowed Commands</label>
                <div className="flex gap-2 mb-2">
                  <Input value={newCmd} onChange={e => setNewCmd(e.target.value)}
                    onKeyDown={e => e.key === 'Enter' && (e.preventDefault(), addCommand())}
                    className="flex-1 font-mono"
                    placeholder="ls, redis-cli[GET,KEYS], psql(SELECT,SHOW)" />
                  <Button size="sm" onClick={addCommand}>Add</Button>
                </div>
                {form.commands.length > 0 && (
                  <div className="flex flex-wrap gap-1.5">
                    {form.commands.map((c, i) => (
                      <span key={i} className="inline-flex items-center gap-1 px-2 py-0.5 text-xs font-mono bg-zinc-200 dark:bg-zinc-800 text-zinc-700 dark:text-zinc-300 rounded border border-zinc-300 dark:border-zinc-700">
                        {formatCommandRule(c)}
                        <button onClick={() => removeCommand(i)} className="text-zinc-600 hover:text-red-400"><X size={12} /></button>
                      </span>
                    ))}
                  </div>
                )}
              </div>
            )}
            <DialogFooter>
              <Button variant="secondary" onClick={() => setModalOpen(false)}>Cancel</Button>
              <Button onClick={submit} disabled={!form.name}>
                {editing ? 'Update' : 'Create'}
              </Button>
            </DialogFooter>
          </div>
        </DialogContent>
      </Dialog>

      <ConfirmDelete
        open={!!deleteTarget}
        onCancel={() => setDeleteTarget(null)}
        onConfirm={() => deleteTarget && remove(deleteTarget)}
        title="Delete profile?"
        description="This will permanently remove the guard profile."
      />
    </div>
  )
}
