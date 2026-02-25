import { useState, useEffect, useCallback } from 'react'
import { Plus, Pencil, Trash2, Shield, AlertTriangle, Check, Copy, ChevronDown, ChevronRight, X } from 'lucide-react'
import { api } from '../api'
import type { GuardProfile, GuardCapabilities, CommandRule } from '../types'
import Modal from '../components/Modal'

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
  const [toast, setToast] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  const showToast = (type: 'success' | 'error', text: string) => {
    setToast({ type, text })
    setTimeout(() => setToast(null), 3000)
  }

  const load = useCallback(async () => {
    try {
      setLoading(true)
      setProfiles(await api.guardProfiles.list())
    } catch (e: unknown) {
      showToast('error', e instanceof Error ? e.message : 'Failed to load')
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
        showToast('success', 'Profile updated')
      } else {
        await api.guardProfiles.create(data)
        showToast('success', 'Profile created')
      }
      setModalOpen(false)
      load()
    } catch (e: unknown) {
      showToast('error', e instanceof Error ? e.message : 'Operation failed')
    }
  }

  const remove = async (id: string) => {
    if (!confirm('Delete this profile?')) return
    try {
      await api.guardProfiles.delete(id)
      showToast('success', 'Profile deleted')
      load()
    } catch (e: unknown) {
      showToast('error', e instanceof Error ? e.message : 'Delete failed')
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
    <div className="p-6 max-w-5xl">
      <div className="flex items-center justify-between mb-5">
        <div>
          <h1 className="text-lg font-semibold text-zinc-100">Guard Profiles</h1>
          <p className="text-xs text-zinc-600 mt-0.5">Security profiles control which commands SSH agents can execute</p>
        </div>
        <button onClick={openCreate} className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-white bg-teal-600 rounded-lg hover:bg-teal-500">
          <Plus size={14} /> New Profile
        </button>
      </div>

      {toast && (
        <div className={`mb-4 flex items-center gap-2 px-3.5 py-2.5 rounded-lg text-xs font-medium ${
          toast.type === 'success' ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-red-500/10 text-red-400 border border-red-500/20'
        }`}>
          {toast.type === 'success' ? <Check size={14} /> : <AlertTriangle size={14} />}
          {toast.text}
        </div>
      )}

      {loading ? (
        <div className="text-center py-12 text-zinc-600 text-sm">Loading...</div>
      ) : profiles.length === 0 ? (
        <div className="text-center py-20">
          <Shield size={40} className="mx-auto text-zinc-700 mb-3" />
          <p className="text-zinc-400 text-sm font-medium">No guard profiles yet</p>
          <p className="text-xs text-zinc-600 mt-1">Create profiles to control SSH command execution</p>
        </div>
      ) : (
        <div className="space-y-2.5">
          {profiles.map(p => (
            <div key={p.id} className="bg-zinc-900 rounded-lg border border-zinc-800 overflow-hidden">
              <div className="px-4 py-3.5">
                <div className="flex items-center justify-between">
                  <button onClick={() => toggle(p.id)} className="flex items-center gap-2 min-w-0 text-left">
                    {expanded.has(p.id) ? <ChevronDown size={14} className="text-zinc-600" /> : <ChevronRight size={14} className="text-zinc-600" />}
                    <span className="font-medium text-zinc-200 text-sm">{p.name}</span>
                    {p.builtin && <span className="px-1.5 py-0.5 text-[11px] font-medium rounded-full bg-violet-500/15 text-violet-400">Built-in</span>}
                    {p.capabilities.unrestricted && <span className="px-1.5 py-0.5 text-[11px] font-medium rounded-full bg-amber-500/15 text-amber-400">Unrestricted</span>}
                  </button>
                  <div className="flex gap-0.5 ml-3">
                    <button onClick={() => openClone(p)} className="p-1.5 rounded-md text-zinc-600 hover:text-blue-400 hover:bg-blue-500/10" title="Clone">
                      <Copy size={14} />
                    </button>
                    <button onClick={() => openEdit(p)} className="p-1.5 rounded-md text-zinc-600 hover:text-teal-400 hover:bg-teal-500/10">
                      <Pencil size={14} />
                    </button>
                    {!p.builtin && (
                      <button onClick={() => remove(p.id)} className="p-1.5 rounded-md text-zinc-600 hover:text-red-400 hover:bg-red-500/10">
                        <Trash2 size={14} />
                      </button>
                    )}
                  </div>
                </div>
                {p.description && <p className="text-xs text-zinc-500 mt-1 ml-[22px]">{p.description}</p>}
              </div>
              {expanded.has(p.id) && (
                <div className="px-4 pb-4 space-y-3">
                  <div className="bg-zinc-950 rounded-lg border border-zinc-800 p-3">
                    <p className="text-[11px] font-semibold text-zinc-600 uppercase tracking-wider mb-2">Capabilities</p>
                    <p className="text-xs text-zinc-400">{capList(p.capabilities)}</p>
                  </div>
                  <div className="bg-zinc-950 rounded-lg border border-zinc-800 p-3">
                    <p className="text-[11px] font-semibold text-zinc-600 uppercase tracking-wider mb-2">Commands ({p.commands.length})</p>
                    {p.commands.length === 0 ? (
                      <p className="text-xs text-zinc-600 italic">{p.capabilities.unrestricted ? 'All commands allowed' : 'No commands configured'}</p>
                    ) : (
                      <div className="flex flex-wrap gap-1.5">
                        {p.commands.map((c, i) => (
                          <span key={i} className="px-2 py-0.5 text-xs font-mono bg-zinc-800 text-zinc-300 rounded" title={
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

      <Modal open={modalOpen} onClose={() => setModalOpen(false)} title={editing ? 'Edit Profile' : 'New Profile'}>
        <div className="space-y-4 max-h-[70vh] overflow-y-auto pr-1">
          <div>
            <label className="block text-xs font-medium text-zinc-400 mb-1.5">Name</label>
            <input value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
              className="w-full px-3 py-2 border border-zinc-700 rounded-lg text-sm bg-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-teal-500/50"
              placeholder="My Custom Profile" />
          </div>
          <div>
            <label className="block text-xs font-medium text-zinc-400 mb-1.5">Description</label>
            <input value={form.description} onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
              className="w-full px-3 py-2 border border-zinc-700 rounded-lg text-sm bg-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-teal-500/50"
              placeholder="Read-only monitoring for production" />
          </div>
          <div>
            <label className={`flex items-center gap-2.5 p-2.5 rounded-lg border cursor-pointer mb-3 ${
              form.capabilities.unrestricted ? 'border-amber-500/40 bg-amber-500/5' : 'border-zinc-800 bg-zinc-950'
            }`}>
              <input type="checkbox" checked={form.capabilities.unrestricted} onChange={e => setCap('unrestricted', e.target.checked)}
                className="rounded border-zinc-600 bg-zinc-800 text-amber-500 focus:ring-amber-500/30 focus:ring-offset-0" />
              <div>
                <span className="text-sm font-medium text-zinc-200">Unrestricted</span>
                <p className="text-[11px] text-zinc-500 mt-0.5">Allow all commands and capabilities</p>
              </div>
            </label>
            {!form.capabilities.unrestricted && (
              <>
                <label className="block text-xs font-medium text-zinc-400 mb-2">Capabilities</label>
                <div className="grid grid-cols-2 gap-2">
                  {(Object.keys(capLabels) as (keyof GuardCapabilities)[]).filter(k => k !== 'unrestricted').map(key => (
                    <label key={key} className="flex items-center gap-2 text-xs text-zinc-300 cursor-pointer">
                      <input type="checkbox" checked={form.capabilities[key]} onChange={e => setCap(key, e.target.checked)}
                        className="rounded border-zinc-600 bg-zinc-800 text-teal-500 focus:ring-teal-500/30 focus:ring-offset-0" />
                      {capLabels[key]}
                    </label>
                  ))}
                </div>
              </>
            )}
          </div>
          {!form.capabilities.unrestricted && (
            <div>
              <label className="block text-xs font-medium text-zinc-400 mb-2">Allowed Commands</label>
              <div className="flex gap-2 mb-2">
                <input value={newCmd} onChange={e => setNewCmd(e.target.value)}
                  onKeyDown={e => e.key === 'Enter' && (e.preventDefault(), addCommand())}
                  className="flex-1 px-3 py-1.5 border border-zinc-700 rounded-lg text-sm font-mono bg-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-teal-500/50"
                  placeholder="ls, redis-cli[GET,KEYS], psql(SELECT,SHOW)" />
                <button onClick={addCommand} className="px-3 py-1.5 text-xs font-medium text-white bg-teal-600 rounded-lg hover:bg-teal-500">
                  Add
                </button>
              </div>
              {form.commands.length > 0 && (
                <div className="flex flex-wrap gap-1.5">
                  {form.commands.map((c, i) => (
                    <span key={i} className="inline-flex items-center gap-1 px-2 py-0.5 text-xs font-mono bg-zinc-800 text-zinc-300 rounded border border-zinc-700">
                      {formatCommandRule(c)}
                      <button onClick={() => removeCommand(i)} className="text-zinc-600 hover:text-red-400"><X size={12} /></button>
                    </span>
                  ))}
                </div>
              )}
            </div>
          )}
          <div className="flex justify-end gap-2 pt-2">
            <button onClick={() => setModalOpen(false)} className="px-3.5 py-2 text-xs font-medium text-zinc-400 bg-zinc-800 border border-zinc-700 rounded-lg hover:text-zinc-200 hover:bg-zinc-700">
              Cancel
            </button>
            <button onClick={submit} disabled={!form.name} className="px-3.5 py-2 text-xs font-medium text-white bg-teal-600 rounded-lg hover:bg-teal-500 disabled:opacity-40">
              {editing ? 'Update' : 'Create'}
            </button>
          </div>
        </div>
      </Modal>
    </div>
  )
}
