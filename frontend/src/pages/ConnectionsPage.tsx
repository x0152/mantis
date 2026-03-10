import { useState, useEffect, useCallback } from 'react'
import { Plus, Pencil, Trash2, Plug, ChevronDown, ChevronRight, MessageSquare, Send, Shield } from 'lucide-react'
import { toast } from 'sonner'
import { api } from '../api'
import type { Connection, Model, GuardProfile } from '../types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/dialog'
import { EmptyState } from '@/components/EmptyState'
import { FormField } from '@/components/FormField'
import { ConfirmDelete } from '@/components/ConfirmDelete'

export default function ConnectionsPage() {
  const [connections, setConnections] = useState<Connection[]>([])
  const [models, setModels] = useState<Model[]>([])
  const [profiles, setProfiles] = useState<GuardProfile[]>([])
  const [loading, setLoading] = useState(true)
  const [expanded, setExpanded] = useState<string | null>(null)
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<Connection | null>(null)
  const emptySsh = { host: '', port: '22', username: '', password: '', privateKey: '' }
  const [form, setForm] = useState({ type: 'ssh', name: '', description: '', modelId: '', profileIds: [] as string[], memoryEnabled: true })
  const [ssh, setSsh] = useState(emptySsh)
  const [profileDropdownOpen, setProfileDropdownOpen] = useState(false)
  const [memoryInput, setMemoryInput] = useState('')
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null)

  const load = useCallback(async () => {
    try {
      setLoading(true)
      const [c, m, p] = await Promise.all([api.connections.list(), api.models.list(), api.guardProfiles.list()])
      setConnections(c)
      setModels(m)
      setProfiles(p)
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Failed to load')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])

  const toggle = (id: string) => setExpanded(prev => prev === id ? null : id)

  const parseSshConfig = (cfg: Record<string, unknown>) => ({
    host: String(cfg.host ?? ''),
    port: String(cfg.port ?? '22'),
    username: String(cfg.username ?? ''),
    password: String(cfg.password ?? ''),
    privateKey: String(cfg.privateKey ?? ''),
  })

  const buildConfig = () => {
    const cfg: Record<string, unknown> = {
      host: ssh.host,
      port: parseInt(ssh.port) || 22,
      username: ssh.username,
    }
    if (ssh.password) cfg.password = ssh.password
    if (ssh.privateKey) cfg.privateKey = ssh.privateKey
    return cfg
  }

  const openCreate = () => {
    setEditing(null)
    setForm({ type: 'ssh', name: '', description: '', modelId: '', profileIds: [], memoryEnabled: true })
    setSsh(emptySsh)
    setProfileDropdownOpen(false)
    setModalOpen(true)
  }

  const openEdit = (c: Connection) => {
    setEditing(c)
    setForm({ type: c.type, name: c.name, description: c.description, modelId: c.modelId, profileIds: c.profileIds || [], memoryEnabled: c.memoryEnabled })
    setSsh(parseSshConfig(c.config))
    setProfileDropdownOpen(false)
    setModalOpen(true)
  }

  const submit = async () => {
    try {
      const config = buildConfig()
      const data = { type: form.type, name: form.name, description: form.description, modelId: form.modelId, config, profileIds: form.profileIds, memoryEnabled: form.memoryEnabled }
      if (editing) {
        await api.connections.update(editing.id, data)
        toast.success('Server updated')
      } else {
        await api.connections.create(data)
        toast.success('Server created')
      }
      setModalOpen(false)
      load()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Operation failed')
    }
  }

  const remove = async (id: string) => {
    try {
      await api.connections.delete(id)
      toast.success('Server deleted')
      setDeleteTarget(null)
      if (expanded === id) setExpanded(null)
      load()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Delete failed')
    }
  }

  const addMemory = async (connectionId: string) => {
    if (!memoryInput.trim()) return
    try {
      await api.connections.addMemory(connectionId, memoryInput.trim())
      setMemoryInput('')
      toast.success('Memory added')
      load()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Failed to add memory')
    }
  }

  const deleteMemory = async (connectionId: string, memoryId: string) => {
    try {
      await api.connections.deleteMemory(connectionId, memoryId)
      toast.success('Memory deleted')
      load()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Failed to delete memory')
    }
  }

  const profileName = (id: string) => profiles.find(p => p.id === id)?.name ?? id.slice(0, 8)

  const modelName = (id: string) => {
    const model = models.find(m => m.id === id)
    if (model) return `${model.name} (${model.connectionId})`
    return id.slice(0, 8) + '...'
  }

  return (
    <div className="p-6 max-w-5xl">
      <div className="flex items-center justify-between mb-5">
        <div>
          <h1 className="text-lg font-semibold text-zinc-100">Servers</h1>
          <p className="text-xs text-zinc-600 mt-0.5">Manage SSH servers with configs, memories, and security profiles</p>
        </div>
        <Button size="sm" onClick={openCreate}>
          <Plus size={14} /> Add Server
        </Button>
      </div>

      {loading ? (
        <div className="text-center py-12 text-zinc-600 text-sm">Loading...</div>
      ) : connections.length === 0 ? (
        <EmptyState icon={Plug} title="No servers yet" description="Create your first server to get started" />
      ) : (
        <div className="space-y-2.5">
          {connections.map(conn => (
            <div key={conn.id} className="bg-zinc-900 rounded-lg border border-zinc-800 overflow-hidden">
              <div className="flex items-center px-4 py-3.5 cursor-pointer hover:bg-zinc-800/40" onClick={() => toggle(conn.id)}>
                <div className="mr-2.5 text-zinc-600">
                  {expanded === conn.id ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="font-medium text-zinc-200 text-sm">{conn.name}</span>
                    <Badge className="uppercase">{conn.type}</Badge>
                    {conn.memories.length > 0 && (
                      <Badge variant="secondary">
                        <MessageSquare size={10} /> {conn.memories.length}
                      </Badge>
                    )}
                    {conn.profileIds?.length > 0 && (
                      <Badge variant="warning">
                        <Shield size={10} /> {conn.profileIds.map(id => profileName(id)).join(', ')}
                      </Badge>
                    )}
                  </div>
                  {conn.description && <p className="text-xs text-zinc-500 mt-0.5">{conn.description}</p>}
                  <p className="text-[11px] text-zinc-600 mt-0.5">LLM: {modelName(conn.modelId)}</p>
                </div>
                <div className="flex gap-0.5 ml-3" onClick={e => e.stopPropagation()}>
                  <Button variant="ghost" size="icon" onClick={() => openEdit(conn)}>
                    <Pencil size={14} />
                  </Button>
                  <Button variant="destructive" size="icon" onClick={() => setDeleteTarget(conn.id)}>
                    <Trash2 size={14} />
                  </Button>
                </div>
              </div>

              {expanded === conn.id && (
                <div className="border-t border-zinc-800 px-4 py-4 bg-zinc-950/50">
                  <div className="grid grid-cols-2 gap-5">
                    <div>
                      <h3 className="text-[11px] font-semibold text-zinc-500 uppercase tracking-wider mb-2">SSH Config</h3>
                      <pre className="bg-zinc-900 border border-zinc-800 rounded-lg p-3 text-xs font-mono text-zinc-400 overflow-auto max-h-48">
                        {JSON.stringify(conn.config, null, 2)}
                      </pre>
                    </div>
                    <div>
                      <h3 className="text-[11px] font-semibold text-zinc-500 uppercase tracking-wider mb-2">Details</h3>
                      <dl className="space-y-2 text-sm">
                        <div className="flex justify-between">
                          <dt className="text-zinc-500">ID</dt>
                          <dd className="font-mono text-[11px] text-zinc-600">{conn.id}</dd>
                        </div>
                        <div className="flex justify-between">
                          <dt className="text-zinc-500">Type</dt>
                          <dd className="text-zinc-300">{conn.type}</dd>
                        </div>
                        <div className="flex justify-between">
                          <dt className="text-zinc-500">Model</dt>
                          <dd className="text-zinc-300">{modelName(conn.modelId)}</dd>
                        </div>
                        <div className="flex justify-between">
                          <dt className="text-zinc-500">Profiles</dt>
                          <dd className="text-zinc-300">{conn.profileIds?.length ? conn.profileIds.map(id => profileName(id)).join(', ') : 'None (unrestricted)'}</dd>
                        </div>
                      </dl>
                    </div>
                  </div>

                  <div className="mt-5">
                    <h3 className="text-[11px] font-semibold text-zinc-500 uppercase tracking-wider mb-2.5">
                      Memories ({conn.memories.length})
                    </h3>
                    {conn.memories.length > 0 && (
                      <div className="space-y-1.5 mb-3">
                        {conn.memories.map(mem => (
                          <div key={mem.id} className="flex items-start gap-2.5 bg-zinc-900 border border-zinc-800 rounded-lg px-3 py-2.5">
                            <MessageSquare size={13} className="text-violet-400 mt-0.5 shrink-0" />
                            <div className="flex-1 min-w-0">
                              <p className="text-sm text-zinc-300">{mem.content}</p>
                              <p className="text-[11px] text-zinc-600 mt-1">{new Date(mem.createdAt).toLocaleString()}</p>
                            </div>
                            <Button variant="ghost" size="icon" onClick={() => deleteMemory(conn.id, mem.id)} className="shrink-0">
                              <Trash2 size={12} />
                            </Button>
                          </div>
                        ))}
                      </div>
                    )}
                    <div className="flex gap-2">
                      <Input
                        value={memoryInput}
                        onChange={e => setMemoryInput(e.target.value)}
                        onKeyDown={e => e.key === 'Enter' && addMemory(conn.id)}
                        placeholder="Add a memory..."
                      />
                      <Button size="sm" onClick={() => addMemory(conn.id)} disabled={!memoryInput.trim()}>
                        <Send size={14} />
                      </Button>
                    </div>
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
            <DialogTitle>{editing ? 'Edit Server' : 'New Server'}</DialogTitle>
            <DialogDescription />
          </DialogHeader>
          <div className="space-y-4">
            <FormField label="Name">
              <Input value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} placeholder="My SSH Server" />
            </FormField>
            <FormField label="Description">
              <Textarea value={form.description} onChange={e => setForm(f => ({ ...f, description: e.target.value }))} className="h-20" placeholder="What this connection is used for..." />
            </FormField>
            <FormField label="Type">
              <select
                value={form.type}
                onChange={e => setForm(f => ({ ...f, type: e.target.value }))}
                className="flex w-full rounded-lg border border-zinc-700 bg-zinc-800 px-3 py-2 text-sm text-zinc-100 focus:outline-none focus:border-teal-500/50"
              >
                <option value="ssh">SSH</option>
              </select>
            </FormField>
            {form.type === 'ssh' && (
              <div className="space-y-3 p-3.5 bg-zinc-950 rounded-lg border border-zinc-800">
                <p className="text-[11px] font-semibold text-zinc-500 uppercase tracking-wider">SSH Configuration</p>
                <div className="grid grid-cols-3 gap-2.5">
                  <div className="col-span-2">
                    <label className="block text-[11px] font-medium text-zinc-500 mb-1">host</label>
                    <Input value={ssh.host} onChange={e => setSsh(s => ({ ...s, host: e.target.value }))} placeholder="192.168.1.100" />
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
                  <Input type="password" value={ssh.password} onChange={e => setSsh(s => ({ ...s, password: e.target.value }))} placeholder="optional" />
                </div>
                <div>
                  <label className="block text-[11px] font-medium text-zinc-500 mb-1">privateKey</label>
                  <Textarea value={ssh.privateKey} onChange={e => setSsh(s => ({ ...s, privateKey: e.target.value }))} className="h-20 font-mono" placeholder="-----BEGIN OPENSSH PRIVATE KEY-----" />
                </div>
              </div>
            )}
            <FormField label="Model">
              <select
                value={form.modelId}
                onChange={e => setForm(f => ({ ...f, modelId: e.target.value }))}
                className="flex w-full rounded-lg border border-zinc-700 bg-zinc-800 px-3 py-2 text-sm text-zinc-100 focus:outline-none focus:border-teal-500/50"
              >
                <option value="">Select model...</option>
                {models.map(m => (
                  <option key={m.id} value={m.id}>{m.name} ({m.connectionId})</option>
                ))}
              </select>
            </FormField>
            <div>
              <label className={`flex items-center gap-2.5 p-2.5 rounded-lg border cursor-pointer ${
                form.memoryEnabled ? 'border-teal-500/40 bg-teal-500/5' : 'border-zinc-800 bg-zinc-900'
              }`}>
                <Checkbox checked={form.memoryEnabled} onCheckedChange={v => setForm(f => ({ ...f, memoryEnabled: !!v }))} />
                <div>
                  <span className="text-sm font-medium text-zinc-200">Memory</span>
                  <p className="text-[11px] text-zinc-500 mt-0.5">Auto-extract and remember facts about this server</p>
                </div>
              </label>
            </div>
            <div>
              <label className="block text-xs font-medium text-zinc-400 mb-1.5">Security Profiles</label>
              <div className="relative">
                <button
                  type="button"
                  onClick={() => setProfileDropdownOpen(o => !o)}
                  className="w-full px-3 py-2 border border-zinc-700 rounded-lg text-sm bg-zinc-800 text-zinc-100 focus:outline-none focus:border-teal-500/50 text-left flex items-center justify-between"
                >
                  <span className={form.profileIds.length === 0 ? 'text-zinc-600' : ''}>
                    {form.profileIds.length === 0 ? 'No profiles (unrestricted)' : form.profileIds.map(id => profileName(id)).join(', ')}
                  </span>
                  <ChevronDown size={14} className={`text-zinc-600 transition-transform ${profileDropdownOpen ? 'rotate-180' : ''}`} />
                </button>
                {profileDropdownOpen && (
                  <>
                    <div className="fixed inset-0 z-10" onClick={() => setProfileDropdownOpen(false)} />
                    <div className="absolute z-20 mt-1 w-full bg-zinc-800 border border-zinc-700 rounded-lg shadow-xl py-1 max-h-48 overflow-y-auto">
                      {profiles.map(p => (
                        <label key={p.id} className="flex items-center gap-2 px-3 py-1.5 hover:bg-zinc-700/50 cursor-pointer text-sm text-zinc-300">
                          <Checkbox
                            checked={form.profileIds.includes(p.id)}
                            onCheckedChange={() => setForm(f => ({
                              ...f,
                              profileIds: f.profileIds.includes(p.id)
                                ? f.profileIds.filter(x => x !== p.id)
                                : [...f.profileIds, p.id]
                            }))}
                          />
                          {p.name}{p.builtin ? ' (built-in)' : ''}
                        </label>
                      ))}
                    </div>
                  </>
                )}
              </div>
            </div>
            <DialogFooter>
              <Button variant="secondary" onClick={() => setModalOpen(false)}>Cancel</Button>
              <Button onClick={submit} disabled={!form.name || !form.modelId}>
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
        title="Delete server?"
        description="This will permanently remove the server and its memories."
      />
    </div>
  )
}
