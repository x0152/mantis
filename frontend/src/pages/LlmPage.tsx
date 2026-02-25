import { useState, useEffect, useCallback } from 'react'
import { Plus, Pencil, Trash2, Link2, Box, AlertTriangle, Check, ChevronDown, ChevronRight, Copy } from 'lucide-react'
import { api } from '../api'
import type { LlmConnection, Model } from '../types'
import Modal from '../components/Modal'

type ConnForm = { id: string; provider: string; baseUrl: string; apiKey: string }
type ModelForm = { name: string; connectionId: string; thinkingMode: string }
type Roles = { summaryModelId: string | null; visionModelId: string | null }

export default function LlmPage() {
  const [connections, setConnections] = useState<LlmConnection[]>([])
  const [models, setModels] = useState<Model[]>([])
  const [roles, setRoles] = useState<Roles>({ summaryModelId: null, visionModelId: null })
  const [loading, setLoading] = useState(true)
  const [expanded, setExpanded] = useState<Set<string>>(new Set())
  const [toast, setToast] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  const [connModalOpen, setConnModalOpen] = useState(false)
  const [editingConn, setEditingConn] = useState<LlmConnection | null>(null)
  const [connForm, setConnForm] = useState<ConnForm>({ id: '', provider: 'openai', baseUrl: '', apiKey: '' })

  const [modelModalOpen, setModelModalOpen] = useState(false)
  const [editingModel, setEditingModel] = useState<Model | null>(null)
  const [modelForm, setModelForm] = useState<ModelForm>({ name: '', connectionId: '', thinkingMode: '' })

  const showToast = (type: 'success' | 'error', text: string) => {
    setToast({ type, text })
    setTimeout(() => setToast(null), 3000)
  }

  const load = useCallback(async () => {
    try {
      setLoading(true)
      const [conns, mdls, cfg] = await Promise.all([api.llmConnections.list(), api.models.list(), api.config.get()])
      setConnections(conns)
      setModels(mdls)
      const data = (cfg.data ?? {}) as Record<string, unknown>
      setRoles({
        summaryModelId: (data.summaryModelId as string) ?? null,
        visionModelId: (data.visionModelId as string) ?? null,
      })
      if (conns.length > 0 && expanded.size === 0) {
        setExpanded(new Set(conns.map(c => c.id)))
      }
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

  const openCreateConn = () => {
    setEditingConn(null)
    setConnForm({ id: '', provider: 'openai', baseUrl: '', apiKey: '' })
    setConnModalOpen(true)
  }
  const openEditConn = (c: LlmConnection) => {
    setEditingConn(c)
    setConnForm({ id: c.id, provider: c.provider, baseUrl: c.baseUrl, apiKey: c.apiKey })
    setConnModalOpen(true)
  }
  const submitConn = async () => {
    try {
      if (editingConn) {
        await api.llmConnections.update(editingConn.id, { provider: connForm.provider, baseUrl: connForm.baseUrl, apiKey: connForm.apiKey })
        showToast('success', 'Connection updated')
      } else {
        await api.llmConnections.create(connForm)
        showToast('success', 'Connection created')
      }
      setConnModalOpen(false)
      load()
    } catch (e: unknown) {
      showToast('error', e instanceof Error ? e.message : 'Operation failed')
    }
  }
  const removeConn = async (id: string) => {
    if (!confirm('Delete this connection and all its models?')) return
    try {
      const connModels = models.filter(m => m.connectionId === id)
      for (const m of connModels) await api.models.delete(m.id)
      await api.llmConnections.delete(id)
      showToast('success', 'Connection deleted')
      load()
    } catch (e: unknown) {
      showToast('error', e instanceof Error ? e.message : 'Delete failed')
    }
  }

  const openCreateModel = (connectionId: string) => {
    setEditingModel(null)
    setModelForm({ name: '', connectionId, thinkingMode: '' })
    setModelModalOpen(true)
  }
  const openEditModel = (m: Model) => {
    setEditingModel(m)
    setModelForm({ name: m.name, connectionId: m.connectionId, thinkingMode: m.thinkingMode })
    setModelModalOpen(true)
  }
  const submitModel = async () => {
    try {
      if (editingModel) {
        await api.models.update(editingModel.id, modelForm.connectionId, modelForm.name, modelForm.thinkingMode)
        showToast('success', 'Model updated')
      } else {
        await api.models.create(modelForm.connectionId, modelForm.name, modelForm.thinkingMode)
        showToast('success', 'Model created')
      }
      setModelModalOpen(false)
      load()
    } catch (e: unknown) {
      showToast('error', e instanceof Error ? e.message : 'Operation failed')
    }
  }
  const removeModel = async (id: string) => {
    if (!confirm('Delete this model?')) return
    try {
      await api.models.delete(id)
      showToast('success', 'Model deleted')
      load()
    } catch (e: unknown) {
      showToast('error', e instanceof Error ? e.message : 'Delete failed')
    }
  }

  const updateRole = async (key: keyof Roles, value: string | null) => {
    try {
      const next = { ...roles, [key]: value }
      setRoles(next)
      const cfg = await api.config.get()
      const data = (cfg.data ?? {}) as Record<string, unknown>
      await api.config.update({ ...data, ...next })
      showToast('success', 'Role updated')
    } catch (e: unknown) {
      showToast('error', e instanceof Error ? e.message : 'Failed to update role')
    }
  }

  const modelLabel = (id: string | null) => {
    if (!id) return null
    const m = models.find(m => m.id === id)
    return m ? null : id.slice(0, 12) + '...'
  }

  return (
    <div className="p-6 max-w-5xl">
      <div className="flex items-center justify-between mb-5">
        <div>
          <h1 className="text-lg font-semibold text-zinc-100">LLM</h1>
          <p className="text-xs text-zinc-600 mt-0.5">Connections and models</p>
        </div>
        <button onClick={openCreateConn} className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-white bg-teal-600 rounded-lg hover:bg-teal-500">
          <Plus size={14} /> Add Connection
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
          <Link2 size={40} className="mx-auto text-zinc-700 mb-3" />
          <p className="text-zinc-400 text-sm font-medium">No LLM connections yet</p>
          <p className="text-xs text-zinc-600 mt-1">Add a connection to start configuring models</p>
        </div>
      ) : (
        <div className="space-y-3">
          {connections.map(conn => {
            const connModels = models.filter(m => m.connectionId === conn.id)
            const isExpanded = expanded.has(conn.id)
            return (
              <div key={conn.id} className="bg-zinc-900 rounded-lg border border-zinc-800 overflow-hidden">
                <div className="px-4 py-3.5">
                  <div className="flex items-center justify-between">
                    <div
                      className="flex items-center gap-2 min-w-0 cursor-pointer select-none"
                      onClick={() => toggle(conn.id)}
                    >
                      {isExpanded ? <ChevronDown size={14} className="text-zinc-600 shrink-0" /> : <ChevronRight size={14} className="text-zinc-600 shrink-0" />}
                      <Link2 size={14} className="text-teal-500 shrink-0" />
                      <span className="font-medium text-zinc-200 text-sm">{conn.id}</span>
                      <span className="px-1.5 py-0.5 text-[11px] font-medium rounded-full bg-zinc-800 text-zinc-500">{conn.provider}</span>
                      <span className="text-xs text-zinc-600 truncate">{conn.baseUrl}</span>
                    </div>
                    <div className="flex items-center gap-1 ml-3">
                      <span className="text-[11px] text-zinc-600 mr-1">{connModels.length} model{connModels.length !== 1 ? 's' : ''}</span>
                      <button onClick={() => openEditConn(conn)} className="p-1.5 rounded-md text-zinc-600 hover:text-teal-400 hover:bg-teal-500/10">
                        <Pencil size={14} />
                      </button>
                      <button onClick={() => removeConn(conn.id)} className="p-1.5 rounded-md text-zinc-600 hover:text-red-400 hover:bg-red-500/10">
                        <Trash2 size={14} />
                      </button>
                    </div>
                  </div>
                </div>

                {isExpanded && (
                  <div className="border-t border-zinc-800">
                    {connModels.length === 0 ? (
                      <div className="px-4 py-4 text-center">
                        <p className="text-xs text-zinc-600 mb-2">No models in this connection</p>
                        <button
                          onClick={() => openCreateModel(conn.id)}
                          className="inline-flex items-center gap-1.5 px-2.5 py-1 text-[11px] font-medium text-teal-400 bg-teal-500/10 rounded-md hover:bg-teal-500/20"
                        >
                          <Plus size={12} /> Add Model
                        </button>
                      </div>
                    ) : (
                      <div>
                        <div className="divide-y divide-zinc-800/50">
                          {connModels.map(m => (
                            <div key={m.id} className="px-4 py-2.5 flex items-center justify-between hover:bg-zinc-800/30">
                              <div className="flex items-center gap-2.5 min-w-0">
                                <Box size={13} className="text-zinc-600 shrink-0" />
                                <span className="text-sm text-zinc-300 font-medium">{m.name}</span>
                                {m.thinkingMode && (
                                  <span className="px-1.5 py-0.5 text-[11px] font-medium rounded-full bg-violet-500/15 text-violet-400">{m.thinkingMode}</span>
                                )}
                                <button
                                  onClick={() => { navigator.clipboard.writeText(m.id); showToast('success', 'Model ID copied') }}
                                  className="inline-flex items-center gap-1 text-[11px] text-zinc-700 font-mono hover:text-zinc-400 transition-colors"
                                  title={m.id}
                                >
                                  {m.id.slice(0, 12)}... <Copy size={11} />
                                </button>
                              </div>
                              <div className="flex gap-0.5">
                                <button onClick={() => openEditModel(m)} className="p-1 rounded-md text-zinc-600 hover:text-teal-400 hover:bg-teal-500/10">
                                  <Pencil size={13} />
                                </button>
                                <button onClick={() => removeModel(m.id)} className="p-1 rounded-md text-zinc-600 hover:text-red-400 hover:bg-red-500/10">
                                  <Trash2 size={13} />
                                </button>
                              </div>
                            </div>
                          ))}
                        </div>
                        <div className="px-4 py-2 border-t border-zinc-800/50">
                          <button
                            onClick={() => openCreateModel(conn.id)}
                            className="inline-flex items-center gap-1 text-[11px] font-medium text-zinc-600 hover:text-teal-400"
                          >
                            <Plus size={12} /> Add model
                          </button>
                        </div>
                      </div>
                    )}
                  </div>
                )}
              </div>
            )
          })}
        </div>
      )}

      {!loading && models.length > 0 && (
        <div className="mt-6 bg-zinc-900 rounded-lg border border-zinc-800 p-4">
          <h2 className="text-sm font-semibold text-zinc-200 mb-3">Model Roles</h2>
          <div className="grid grid-cols-2 gap-4">
            {([
              { key: 'summaryModelId' as const, label: 'Summary Model', hint: 'Used for conversation summaries and memory extraction' },
              { key: 'visionModelId' as const, label: 'Vision Model', hint: 'Used for image description (Vision LLM)' },
            ]).map(({ key, label, hint }) => {
              const stale = !!roles[key] && !models.find(m => m.id === roles[key])
              return (
                <div key={key}>
                  <label className="block text-xs font-medium text-zinc-400 mb-1.5">{label}</label>
                  <select
                    value={roles[key] ?? ''}
                    onChange={e => updateRole(key, e.target.value || null)}
                    className={`w-full px-3 py-2 border rounded-lg text-sm bg-zinc-800 text-zinc-100 focus:outline-none focus:border-teal-500/50 ${stale ? 'border-amber-500/50' : 'border-zinc-700'}`}
                  >
                    <option value="">None</option>
                    {stale && <option value={roles[key]!}>(deleted) {modelLabel(roles[key])}</option>}
                    {models.map(m => (
                      <option key={m.id} value={m.id}>{m.name} ({connections.find(c => c.id === m.connectionId)?.id ?? '?'})</option>
                    ))}
                  </select>
                  {stale && <p className="text-[11px] text-amber-400 mt-1">Selected model was deleted — please choose another</p>}
                  <p className="text-[11px] text-zinc-600 mt-1">{hint}</p>
                </div>
              )
            })}
          </div>
        </div>
      )}

      <Modal open={connModalOpen} onClose={() => setConnModalOpen(false)} title={editingConn ? 'Edit Connection' : 'New Connection'}>
        <div className="space-y-4">
          <div>
            <label className="block text-xs font-medium text-zinc-400 mb-1.5">ID</label>
            <input
              value={connForm.id}
              disabled={!!editingConn}
              onChange={e => setConnForm(f => ({ ...f, id: e.target.value }))}
              className="w-full px-3 py-2 border border-zinc-700 rounded-lg text-sm bg-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-teal-500/50 disabled:opacity-50"
              placeholder="openai-prod"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-zinc-400 mb-1.5">Provider</label>
            <select
              value={connForm.provider}
              onChange={e => setConnForm(f => ({ ...f, provider: e.target.value }))}
              className="w-full px-3 py-2 border border-zinc-700 rounded-lg text-sm bg-zinc-800 text-zinc-100 focus:outline-none focus:border-teal-500/50"
            >
              <option value="openai">openai</option>
            </select>
          </div>
          <div>
            <label className="block text-xs font-medium text-zinc-400 mb-1.5">Base URL</label>
            <input
              value={connForm.baseUrl}
              onChange={e => setConnForm(f => ({ ...f, baseUrl: e.target.value }))}
              className="w-full px-3 py-2 border border-zinc-700 rounded-lg text-sm bg-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-teal-500/50"
              placeholder="http://localhost:1234/v1"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-zinc-400 mb-1.5">API Key</label>
            <input
              type="password"
              value={connForm.apiKey}
              onChange={e => setConnForm(f => ({ ...f, apiKey: e.target.value }))}
              className="w-full px-3 py-2 border border-zinc-700 rounded-lg text-sm bg-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-teal-500/50"
              placeholder="optional"
            />
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <button onClick={() => setConnModalOpen(false)} className="px-3.5 py-2 text-xs font-medium text-zinc-400 bg-zinc-800 border border-zinc-700 rounded-lg hover:text-zinc-200 hover:bg-zinc-700">
              Cancel
            </button>
            <button onClick={submitConn} disabled={!connForm.id || !connForm.provider || !connForm.baseUrl} className="px-3.5 py-2 text-xs font-medium text-white bg-teal-600 rounded-lg hover:bg-teal-500 disabled:opacity-40">
              {editingConn ? 'Update' : 'Create'}
            </button>
          </div>
        </div>
      </Modal>

      <Modal open={modelModalOpen} onClose={() => setModelModalOpen(false)} title={editingModel ? 'Edit Model' : 'New Model'}>
        <div className="space-y-4">
          <div>
            <label className="block text-xs font-medium text-zinc-400 mb-1.5">Model name</label>
            <input
              value={modelForm.name}
              onChange={e => setModelForm(f => ({ ...f, name: e.target.value }))}
              className="w-full px-3 py-2 border border-zinc-700 rounded-lg text-sm bg-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-teal-500/50"
              placeholder="gpt-4o"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-zinc-400 mb-1.5">Connection</label>
            <select
              value={modelForm.connectionId}
              onChange={e => setModelForm(f => ({ ...f, connectionId: e.target.value }))}
              className="w-full px-3 py-2 border border-zinc-700 rounded-lg text-sm bg-zinc-800 text-zinc-100 focus:outline-none focus:border-teal-500/50"
            >
              <option value="">Select connection...</option>
              {connections.map(c => (
                <option key={c.id} value={c.id}>{c.id} ({c.provider})</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-xs font-medium text-zinc-400 mb-1.5">Thinking mode</label>
            <select
              value={modelForm.thinkingMode}
              onChange={e => setModelForm(f => ({ ...f, thinkingMode: e.target.value }))}
              className="w-full px-3 py-2 border border-zinc-700 rounded-lg text-sm bg-zinc-800 text-zinc-100 focus:outline-none focus:border-teal-500/50"
            >
              <option value="">None (default)</option>
              <option value="skip">Skip — remove thinking blocks</option>
              <option value="inline">Inline — strip tags, keep content</option>
            </select>
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <button onClick={() => setModelModalOpen(false)} className="px-3.5 py-2 text-xs font-medium text-zinc-400 bg-zinc-800 border border-zinc-700 rounded-lg hover:text-zinc-200 hover:bg-zinc-700">
              Cancel
            </button>
            <button onClick={submitModel} disabled={!modelForm.name || !modelForm.connectionId} className="px-3.5 py-2 text-xs font-medium text-white bg-teal-600 rounded-lg hover:bg-teal-500 disabled:opacity-40">
              {editingModel ? 'Update' : 'Create'}
            </button>
          </div>
        </div>
      </Modal>
    </div>
  )
}
