import { useState, useEffect, useCallback } from 'react'
import { Plus, Pencil, Trash2, Plug, ChevronDown, ChevronRight, MessageSquare, AlertTriangle, Check, Send, Shield, Play, Pause } from 'lucide-react'
import { api } from '../api'
import type { Connection, Model, GuardRule } from '../types'
import Modal from '../components/Modal'

export default function ConnectionsPage() {
  const [connections, setConnections] = useState<Connection[]>([])
  const [models, setModels] = useState<Model[]>([])
  const [guardRules, setGuardRules] = useState<GuardRule[]>([])
  const [loading, setLoading] = useState(true)
  const [expanded, setExpanded] = useState<string | null>(null)
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<Connection | null>(null)
  const emptySsh = { host: '', port: '22', username: '', password: '', privateKey: '' }
  const [form, setForm] = useState({ type: 'ssh', name: '', description: '', modelId: '' })
  const [ssh, setSsh] = useState(emptySsh)
  const [memoryInput, setMemoryInput] = useState('')
  const [ruleModalOpen, setRuleModalOpen] = useState(false)
  const [editingRule, setEditingRule] = useState<GuardRule | null>(null)
  const [ruleConnId, setRuleConnId] = useState('')
  const [ruleForm, setRuleForm] = useState({ name: '', description: '', pattern: '', enabled: true })
  const [toast, setToast] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  const showToast = (type: 'success' | 'error', text: string) => {
    setToast({ type, text })
    setTimeout(() => setToast(null), 3000)
  }

  const load = useCallback(async () => {
    try {
      setLoading(true)
      const [c, m, gr] = await Promise.all([api.connections.list(), api.models.list(), api.guardRules.list()])
      setConnections(c)
      setModels(m)
      setGuardRules(gr)
    } catch (e: unknown) {
      showToast('error', e instanceof Error ? e.message : 'Failed to load')
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
    setForm({ type: 'ssh', name: '', description: '', modelId: '' })
    setSsh(emptySsh)
    setModalOpen(true)
  }

  const openEdit = (c: Connection) => {
    setEditing(c)
    setForm({ type: c.type, name: c.name, description: c.description, modelId: c.modelId })
    setSsh(parseSshConfig(c.config))
    setModalOpen(true)
  }

  const submit = async () => {
    try {
      const config = buildConfig()
      const data = { type: form.type, name: form.name, description: form.description, modelId: form.modelId, config }
      if (editing) {
        await api.connections.update(editing.id, data)
        showToast('success', 'Server updated')
      } else {
        await api.connections.create(data)
        showToast('success', 'Server created')
      }
      setModalOpen(false)
      load()
    } catch (e: unknown) {
      showToast('error', e instanceof Error ? e.message : 'Operation failed')
    }
  }

  const remove = async (id: string) => {
    if (!confirm('Delete this server?')) return
    try {
      await api.connections.delete(id)
      showToast('success', 'Server deleted')
      if (expanded === id) setExpanded(null)
      load()
    } catch (e: unknown) {
      showToast('error', e instanceof Error ? e.message : 'Delete failed')
    }
  }

  const addMemory = async (connectionId: string) => {
    if (!memoryInput.trim()) return
    try {
      await api.connections.addMemory(connectionId, memoryInput.trim())
      setMemoryInput('')
      showToast('success', 'Memory added')
      load()
    } catch (e: unknown) {
      showToast('error', e instanceof Error ? e.message : 'Failed to add memory')
    }
  }

  const deleteMemory = async (connectionId: string, memoryId: string) => {
    try {
      await api.connections.deleteMemory(connectionId, memoryId)
      showToast('success', 'Memory deleted')
      load()
    } catch (e: unknown) {
      showToast('error', e instanceof Error ? e.message : 'Failed to delete memory')
    }
  }

  const openCreateRule = (connectionId: string) => {
    setEditingRule(null)
    setRuleConnId(connectionId)
    setRuleForm({ name: '', description: '', pattern: '', enabled: true })
    setRuleModalOpen(true)
  }

  const openEditRule = (r: GuardRule) => {
    setEditingRule(r)
    setRuleConnId(r.connectionId)
    setRuleForm({ name: r.name, description: r.description, pattern: r.pattern, enabled: r.enabled })
    setRuleModalOpen(true)
  }

  const submitRule = async () => {
    try {
      const data = { ...ruleForm, connectionId: ruleConnId }
      if (editingRule) {
        await api.guardRules.update(editingRule.id, data)
        showToast('success', 'Rule updated')
      } else {
        await api.guardRules.create(data)
        showToast('success', 'Rule created')
      }
      setRuleModalOpen(false)
      load()
    } catch (e: unknown) {
      showToast('error', e instanceof Error ? e.message : 'Operation failed')
    }
  }

  const toggleRule = async (r: GuardRule) => {
    try {
      await api.guardRules.update(r.id, { ...r, enabled: !r.enabled })
      load()
    } catch (e: unknown) {
      showToast('error', e instanceof Error ? e.message : 'Toggle failed')
    }
  }

  const removeRule = async (id: string) => {
    if (!confirm('Delete this guard rule?')) return
    try {
      await api.guardRules.delete(id)
      showToast('success', 'Rule deleted')
      load()
    } catch (e: unknown) {
      showToast('error', e instanceof Error ? e.message : 'Delete failed')
    }
  }

  const rulesForConnection = (connectionId: string) => guardRules.filter(r => r.connectionId === connectionId)

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
          <p className="text-xs text-zinc-600 mt-0.5">Manage SSH servers with configs and memories</p>
        </div>
        <button onClick={openCreate} className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-white bg-teal-600 rounded-lg hover:bg-teal-500">
          <Plus size={14} /> Add Server
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
      ) : connections.length === 0 ? (
        <div className="text-center py-20">
          <Plug size={40} className="mx-auto text-zinc-700 mb-3" />
          <p className="text-zinc-400 text-sm font-medium">No servers yet</p>
          <p className="text-xs text-zinc-600 mt-1">Create your first server to get started</p>
        </div>
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
                    <span className="px-1.5 py-0.5 text-[11px] font-medium rounded-full bg-blue-500/15 text-blue-400 uppercase">{conn.type}</span>
                    {conn.memories.length > 0 && (
                      <span className="inline-flex items-center gap-1 px-1.5 py-0.5 text-[11px] font-medium rounded-full bg-violet-500/15 text-violet-400">
                        <MessageSquare size={10} /> {conn.memories.length}
                      </span>
                    )}
                    {rulesForConnection(conn.id).length > 0 && (
                      <span className="inline-flex items-center gap-1 px-1.5 py-0.5 text-[11px] font-medium rounded-full bg-amber-500/15 text-amber-400">
                        <Shield size={10} /> {rulesForConnection(conn.id).length}
                      </span>
                    )}
                  </div>
                  {conn.description && <p className="text-xs text-zinc-500 mt-0.5">{conn.description}</p>}
                  <p className="text-[11px] text-zinc-600 mt-0.5">LLM: {modelName(conn.modelId)}</p>
                </div>
                <div className="flex gap-0.5 ml-3" onClick={e => e.stopPropagation()}>
                  <button onClick={() => openEdit(conn)} className="p-1.5 rounded-md text-zinc-600 hover:text-teal-400 hover:bg-teal-500/10">
                    <Pencil size={14} />
                  </button>
                  <button onClick={() => remove(conn.id)} className="p-1.5 rounded-md text-zinc-600 hover:text-red-400 hover:bg-red-500/10">
                    <Trash2 size={14} />
                  </button>
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
                            <button onClick={() => deleteMemory(conn.id, mem.id)} className="p-1 rounded-md text-zinc-700 hover:text-red-400 hover:bg-red-500/10 shrink-0">
                              <Trash2 size={12} />
                            </button>
                          </div>
                        ))}
                      </div>
                    )}
                    <div className="flex gap-2">
                      <input
                        value={memoryInput}
                        onChange={e => setMemoryInput(e.target.value)}
                        onKeyDown={e => e.key === 'Enter' && addMemory(conn.id)}
                        className="flex-1 px-3 py-2 border border-zinc-700 rounded-lg text-sm bg-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-teal-500/50"
                        placeholder="Add a memory..."
                      />
                      <button onClick={() => addMemory(conn.id)} disabled={!memoryInput.trim()} className="px-3 py-2 text-xs font-medium text-white bg-teal-600 rounded-lg hover:bg-teal-500 disabled:opacity-40">
                        <Send size={14} />
                      </button>
                    </div>
                  </div>

                  <div className="mt-5">
                    <div className="flex items-center justify-between mb-2.5">
                      <h3 className="text-[11px] font-semibold text-zinc-500 uppercase tracking-wider">
                        Guard Rules ({rulesForConnection(conn.id).length})
                      </h3>
                      <button onClick={() => openCreateRule(conn.id)} className="inline-flex items-center gap-1 px-2 py-1 text-[11px] font-medium text-teal-400 hover:bg-teal-500/10 rounded-md">
                        <Plus size={12} /> Add Rule
                      </button>
                    </div>
                    {rulesForConnection(conn.id).length > 0 && (
                      <div className="space-y-1.5">
                        {rulesForConnection(conn.id).map(rule => (
                          <div key={rule.id} className={`flex items-center gap-2.5 bg-zinc-900 border border-zinc-800 rounded-lg px-3 py-2.5 ${!rule.enabled ? 'opacity-50' : ''}`}>
                            <Shield size={13} className="text-amber-400 shrink-0" />
                            <div className="flex-1 min-w-0">
                              <div className="flex items-center gap-2">
                                <span className="text-sm font-medium text-zinc-300">{rule.name}</span>
                                <span className={`px-1.5 py-0.5 text-[11px] rounded-full ${rule.enabled ? 'bg-emerald-500/15 text-emerald-400' : 'bg-zinc-800 text-zinc-600'}`}>
                                  {rule.enabled ? 'on' : 'off'}
                                </span>
                              </div>
                              <code className="text-[11px] text-zinc-600 font-mono">{rule.pattern}</code>
                            </div>
                            <div className="flex gap-0.5 shrink-0">
                              <button onClick={() => toggleRule(rule)} className="p-1 rounded-md text-zinc-600 hover:text-amber-400 hover:bg-amber-500/10">
                                {rule.enabled ? <Pause size={12} /> : <Play size={12} />}
                              </button>
                              <button onClick={() => openEditRule(rule)} className="p-1 rounded-md text-zinc-600 hover:text-teal-400 hover:bg-teal-500/10">
                                <Pencil size={12} />
                              </button>
                              <button onClick={() => removeRule(rule.id)} className="p-1 rounded-md text-zinc-600 hover:text-red-400 hover:bg-red-500/10">
                                <Trash2 size={12} />
                              </button>
                            </div>
                          </div>
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

      <Modal open={modalOpen} onClose={() => setModalOpen(false)} title={editing ? 'Edit Server' : 'New Server'}>
        <div className="space-y-4">
          <div>
            <label className="block text-xs font-medium text-zinc-400 mb-1.5">Name</label>
            <input
              value={form.name}
              onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
              className="w-full px-3 py-2 border border-zinc-700 rounded-lg text-sm bg-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-teal-500/50"
              placeholder="My SSH Server"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-zinc-400 mb-1.5">Description</label>
            <textarea
              value={form.description}
              onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
              className="w-full px-3 py-2 border border-zinc-700 rounded-lg text-sm bg-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-teal-500/50 resize-none h-20"
              placeholder="What this connection is used for..."
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-zinc-400 mb-1.5">Type</label>
            <select
              value={form.type}
              onChange={e => setForm(f => ({ ...f, type: e.target.value }))}
              className="w-full px-3 py-2 border border-zinc-700 rounded-lg text-sm bg-zinc-800 text-zinc-100 focus:outline-none focus:border-teal-500/50"
            >
              <option value="ssh">SSH</option>
            </select>
          </div>
          {form.type === 'ssh' && (
            <div className="space-y-3 p-3.5 bg-zinc-950 rounded-lg border border-zinc-800">
              <p className="text-[11px] font-semibold text-zinc-500 uppercase tracking-wider">SSH Configuration</p>
              <div className="grid grid-cols-3 gap-2.5">
                <div className="col-span-2">
                  <label className="block text-[11px] font-medium text-zinc-500 mb-1">host</label>
                  <input value={ssh.host} onChange={e => setSsh(s => ({ ...s, host: e.target.value }))} className="w-full px-2.5 py-1.5 border border-zinc-700 rounded-md text-sm bg-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-teal-500/50" placeholder="192.168.1.100" />
                </div>
                <div>
                  <label className="block text-[11px] font-medium text-zinc-500 mb-1">port</label>
                  <input value={ssh.port} onChange={e => setSsh(s => ({ ...s, port: e.target.value }))} className="w-full px-2.5 py-1.5 border border-zinc-700 rounded-md text-sm bg-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-teal-500/50" placeholder="22" />
                </div>
              </div>
              <div>
                <label className="block text-[11px] font-medium text-zinc-500 mb-1">username</label>
                <input value={ssh.username} onChange={e => setSsh(s => ({ ...s, username: e.target.value }))} className="w-full px-2.5 py-1.5 border border-zinc-700 rounded-md text-sm bg-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-teal-500/50" placeholder="root" />
              </div>
              <div>
                <label className="block text-[11px] font-medium text-zinc-500 mb-1">password</label>
                <input type="password" value={ssh.password} onChange={e => setSsh(s => ({ ...s, password: e.target.value }))} className="w-full px-2.5 py-1.5 border border-zinc-700 rounded-md text-sm bg-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-teal-500/50" placeholder="optional" />
              </div>
              <div>
                <label className="block text-[11px] font-medium text-zinc-500 mb-1">privateKey</label>
                <textarea value={ssh.privateKey} onChange={e => setSsh(s => ({ ...s, privateKey: e.target.value }))} className="w-full h-20 px-2.5 py-1.5 border border-zinc-700 rounded-md text-sm font-mono bg-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-teal-500/50 resize-none" placeholder="-----BEGIN OPENSSH PRIVATE KEY-----" />
              </div>
            </div>
          )}
          <div>
            <label className="block text-xs font-medium text-zinc-400 mb-1.5">Model</label>
            <select
              value={form.modelId}
              onChange={e => setForm(f => ({ ...f, modelId: e.target.value }))}
              className="w-full px-3 py-2 border border-zinc-700 rounded-lg text-sm bg-zinc-800 text-zinc-100 focus:outline-none focus:border-teal-500/50"
            >
              <option value="">Select model...</option>
              {models.map(m => (
                <option key={m.id} value={m.id}>{m.name} ({m.connectionId})</option>
              ))}
            </select>
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <button onClick={() => setModalOpen(false)} className="px-3.5 py-2 text-xs font-medium text-zinc-400 bg-zinc-800 border border-zinc-700 rounded-lg hover:text-zinc-200 hover:bg-zinc-700">
              Cancel
            </button>
            <button onClick={submit} disabled={!form.name || !form.modelId} className="px-3.5 py-2 text-xs font-medium text-white bg-teal-600 rounded-lg hover:bg-teal-500 disabled:opacity-40">
              {editing ? 'Update' : 'Create'}
            </button>
          </div>
        </div>
      </Modal>

      <Modal open={ruleModalOpen} onClose={() => setRuleModalOpen(false)} title={editingRule ? 'Edit Guard Rule' : 'New Guard Rule'}>
        <div className="space-y-4">
          <div>
            <label className="block text-xs font-medium text-zinc-400 mb-1.5">Name</label>
            <input
              value={ruleForm.name}
              onChange={e => setRuleForm(f => ({ ...f, name: e.target.value }))}
              className="w-full px-3 py-2 border border-zinc-700 rounded-lg text-sm bg-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-teal-500/50"
              placeholder="Block destructive commands"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-zinc-400 mb-1.5">Description</label>
            <input
              value={ruleForm.description}
              onChange={e => setRuleForm(f => ({ ...f, description: e.target.value }))}
              className="w-full px-3 py-2 border border-zinc-700 rounded-lg text-sm bg-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-teal-500/50"
              placeholder="Prevents rm -rf on critical paths"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-zinc-400 mb-1.5">Pattern (regex)</label>
            <input
              value={ruleForm.pattern}
              onChange={e => setRuleForm(f => ({ ...f, pattern: e.target.value }))}
              className="w-full px-3 py-2 border border-zinc-700 rounded-lg text-sm font-mono bg-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-teal-500/50"
              placeholder="rm\s+-rf\s+/"
            />
          </div>
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => setRuleForm(f => ({ ...f, enabled: !f.enabled }))}
              className={`relative inline-flex h-5 w-9 items-center rounded-full ${ruleForm.enabled ? 'bg-teal-600' : 'bg-zinc-700'}`}
            >
              <span className={`inline-block h-3.5 w-3.5 rounded-full bg-white ${ruleForm.enabled ? 'translate-x-[18px]' : 'translate-x-[3px]'}`} />
            </button>
            <span className="text-xs text-zinc-400">{ruleForm.enabled ? 'Enabled' : 'Disabled'}</span>
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <button onClick={() => setRuleModalOpen(false)} className="px-3.5 py-2 text-xs font-medium text-zinc-400 bg-zinc-800 border border-zinc-700 rounded-lg hover:text-zinc-200 hover:bg-zinc-700">
              Cancel
            </button>
            <button onClick={submitRule} disabled={!ruleForm.name || !ruleForm.pattern} className="px-3.5 py-2 text-xs font-medium text-white bg-teal-600 rounded-lg hover:bg-teal-500 disabled:opacity-40">
              {editingRule ? 'Update' : 'Create'}
            </button>
          </div>
        </div>
      </Modal>
    </div>
  )
}
