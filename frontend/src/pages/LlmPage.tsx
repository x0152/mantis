import { useState, useEffect, useCallback } from 'react'
import { Plus, Pencil, Trash2, Link2, Box, ChevronDown, ChevronRight, Copy } from 'lucide-react'
import { toast } from 'sonner'
import { api } from '../api'
import type { LlmConnection, Model } from '../types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/dialog'
import { EmptyState } from '@/components/EmptyState'
import { FormField } from '@/components/FormField'
import { ConfirmDelete } from '@/components/ConfirmDelete'

type ConnForm = { id: string; provider: string; baseUrl: string; apiKey: string }
type ModelForm = { name: string; connectionId: string; thinkingMode: string }
type Roles = { summaryModelId: string | null; visionModelId: string | null }

export default function LlmPage() {
  const [connections, setConnections] = useState<LlmConnection[]>([])
  const [models, setModels] = useState<Model[]>([])
  const [roles, setRoles] = useState<Roles>({ summaryModelId: null, visionModelId: null })
  const [loading, setLoading] = useState(true)
  const [expanded, setExpanded] = useState<Set<string>>(new Set())
  const [deleteConnTarget, setDeleteConnTarget] = useState<string | null>(null)
  const [deleteModelTarget, setDeleteModelTarget] = useState<string | null>(null)

  const [connModalOpen, setConnModalOpen] = useState(false)
  const [editingConn, setEditingConn] = useState<LlmConnection | null>(null)
  const [connForm, setConnForm] = useState<ConnForm>({ id: '', provider: 'openai', baseUrl: '', apiKey: '' })

  const [modelModalOpen, setModelModalOpen] = useState(false)
  const [editingModel, setEditingModel] = useState<Model | null>(null)
  const [modelForm, setModelForm] = useState<ModelForm>({ name: '', connectionId: '', thinkingMode: '' })

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
        toast.success('Connection updated')
      } else {
        await api.llmConnections.create(connForm)
        toast.success('Connection created')
      }
      setConnModalOpen(false)
      load()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Operation failed')
    }
  }
  const removeConn = async (id: string) => {
    try {
      const connModels = models.filter(m => m.connectionId === id)
      for (const m of connModels) await api.models.delete(m.id)
      await api.llmConnections.delete(id)
      toast.success('Connection deleted')
      setDeleteConnTarget(null)
      load()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Delete failed')
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
        toast.success('Model updated')
      } else {
        await api.models.create(modelForm.connectionId, modelForm.name, modelForm.thinkingMode)
        toast.success('Model created')
      }
      setModelModalOpen(false)
      load()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Operation failed')
    }
  }
  const removeModel = async (id: string) => {
    try {
      await api.models.delete(id)
      toast.success('Model deleted')
      setDeleteModelTarget(null)
      load()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Delete failed')
    }
  }

  const updateRole = async (key: keyof Roles, value: string | null) => {
    try {
      const next = { ...roles, [key]: value }
      setRoles(next)
      const cfg = await api.config.get()
      const data = (cfg.data ?? {}) as Record<string, unknown>
      await api.config.update({ ...data, ...next })
      toast.success('Role updated')
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Failed to update role')
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
          <h1 className="text-lg font-semibold text-zinc-900 dark:text-zinc-100">LLM</h1>
          <p className="text-xs text-zinc-500 dark:text-zinc-600 mt-0.5">Connections and models</p>
        </div>
        <Button size="sm" onClick={openCreateConn}>
          <Plus size={14} /> Add Connection
        </Button>
      </div>

      {loading ? (
        <div className="text-center py-12 text-zinc-500 dark:text-zinc-600 text-sm">Loading...</div>
      ) : connections.length === 0 ? (
        <EmptyState icon={Link2} title="No LLM connections yet" description="Add a connection to start configuring models" />
      ) : (
        <div className="space-y-3">
          {connections.map(conn => {
            const connModels = models.filter(m => m.connectionId === conn.id)
            const isExpanded = expanded.has(conn.id)
            return (
              <div key={conn.id} className="bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800 overflow-hidden">
                <div className="px-4 py-3.5">
                  <div className="flex items-center justify-between">
                    <div
                      className="flex items-center gap-2 min-w-0 cursor-pointer select-none"
                      onClick={() => toggle(conn.id)}
                    >
                      {isExpanded ? <ChevronDown size={14} className="text-zinc-600 shrink-0" /> : <ChevronRight size={14} className="text-zinc-600 shrink-0" />}
                      <Link2 size={14} className="text-teal-500 shrink-0" />
                      <span className="font-medium text-zinc-800 dark:text-zinc-200 text-sm">{conn.id}</span>
                      <Badge variant="muted">{conn.provider}</Badge>
                      <span className="text-xs text-zinc-500 dark:text-zinc-600 truncate">{conn.baseUrl}</span>
                    </div>
                    <div className="flex items-center gap-1 ml-3">
                      <span className="text-[11px] text-zinc-500 dark:text-zinc-600 mr-1">{connModels.length} model{connModels.length !== 1 ? 's' : ''}</span>
                      <Button variant="ghost" size="icon" onClick={() => openEditConn(conn)}>
                        <Pencil size={14} />
                      </Button>
                      <Button variant="destructive" size="icon" onClick={() => setDeleteConnTarget(conn.id)}>
                        <Trash2 size={14} />
                      </Button>
                    </div>
                  </div>
                </div>

                {isExpanded && (
                  <div className="border-t border-zinc-200 dark:border-zinc-800">
                    {connModels.length === 0 ? (
                      <div className="px-4 py-4 text-center">
                        <p className="text-xs text-zinc-600 mb-2">No models in this connection</p>
                        <Button size="sm" variant="ghost" onClick={() => openCreateModel(conn.id)}>
                          <Plus size={12} /> Add Model
                        </Button>
                      </div>
                    ) : (
                      <div>
                          <div className="divide-y divide-zinc-200/50 dark:divide-zinc-800/50">
                          {connModels.map(m => (
                            <div key={m.id} className="px-4 py-2.5 flex items-center justify-between hover:bg-zinc-100/50 dark:hover:bg-zinc-800/30">
                              <div className="flex items-center gap-2.5 min-w-0">
                                <Box size={13} className="text-zinc-600 shrink-0" />
                                <span className="text-sm text-zinc-700 dark:text-zinc-300 font-medium">{m.name}</span>
                                {m.thinkingMode && (
                                  <Badge variant="secondary">{m.thinkingMode}</Badge>
                                )}
                                <button
                                  onClick={() => { navigator.clipboard.writeText(m.id); toast.success('Model ID copied') }}
                                  className="inline-flex items-center gap-1 text-[11px] text-zinc-400 dark:text-zinc-700 font-mono hover:text-zinc-600 dark:hover:text-zinc-400 transition-colors"
                                  title={m.id}
                                >
                                  {m.id.slice(0, 12)}... <Copy size={11} />
                                </button>
                              </div>
                              <div className="flex gap-0.5">
                                <Button variant="ghost" size="icon" onClick={() => openEditModel(m)}>
                                  <Pencil size={13} />
                                </Button>
                                <Button variant="destructive" size="icon" onClick={() => setDeleteModelTarget(m.id)}>
                                  <Trash2 size={13} />
                                </Button>
                              </div>
                            </div>
                          ))}
                        </div>
                        <div className="px-4 py-2 border-t border-zinc-200/50 dark:border-zinc-800/50">
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
        <div className="mt-6 bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800 p-4">
          <h2 className="text-sm font-semibold text-zinc-800 dark:text-zinc-200 mb-3">Model Roles</h2>
          <div className="grid grid-cols-2 gap-4">
            {([
              { key: 'summaryModelId' as const, label: 'Summary Model', hint: 'Used for conversation summaries and memory extraction' },
              { key: 'visionModelId' as const, label: 'Vision Model', hint: 'Used for image description (Vision LLM)' },
            ]).map(({ key, label, hint }) => {
              const stale = !!roles[key] && !models.find(m => m.id === roles[key])
              return (
                <div key={key}>
                  <label className="block text-xs font-medium text-zinc-500 dark:text-zinc-400 mb-1.5">{label}</label>
                  <select
                    value={roles[key] ?? ''}
                    onChange={e => updateRole(key, e.target.value || null)}
                    className={`w-full px-3 py-2 border rounded-lg text-sm bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 focus:outline-none focus:border-teal-500/50 ${stale ? 'border-amber-500/50' : 'border-zinc-300 dark:border-zinc-700'}`}
                  >
                    <option value="">None</option>
                    {stale && <option value={roles[key]!}>(deleted) {modelLabel(roles[key])}</option>}
                    {models.map(m => (
                      <option key={m.id} value={m.id}>{m.name} ({connections.find(c => c.id === m.connectionId)?.id ?? '?'})</option>
                    ))}
                  </select>
                  {stale && <p className="text-[11px] text-amber-400 mt-1">Selected model was deleted — please choose another</p>}
                  <p className="text-[11px] text-zinc-500 dark:text-zinc-600 mt-1">{hint}</p>
                </div>
              )
            })}
          </div>
        </div>
      )}

      <Dialog open={connModalOpen} onOpenChange={setConnModalOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editingConn ? 'Edit Connection' : 'New Connection'}</DialogTitle>
            <DialogDescription />
          </DialogHeader>
          <div className="space-y-4">
            <FormField label="ID">
              <Input
                value={connForm.id}
                disabled={!!editingConn}
                onChange={e => setConnForm(f => ({ ...f, id: e.target.value }))}
                placeholder="openai-prod"
              />
            </FormField>
            <FormField label="Provider">
              <select
                value={connForm.provider}
                onChange={e => setConnForm(f => ({ ...f, provider: e.target.value }))}
                className="flex w-full rounded-lg border border-zinc-300 dark:border-zinc-700 bg-white dark:bg-zinc-800 px-3 py-2 text-sm text-zinc-900 dark:text-zinc-100 focus:outline-none focus:border-teal-500/50"
              >
                <option value="openai">openai</option>
              </select>
            </FormField>
            <FormField label="Base URL">
              <Input
                value={connForm.baseUrl}
                onChange={e => setConnForm(f => ({ ...f, baseUrl: e.target.value }))}
                placeholder="http://localhost:1234/v1"
              />
            </FormField>
            <FormField label="API Key">
              <Input
                type="password"
                value={connForm.apiKey}
                onChange={e => setConnForm(f => ({ ...f, apiKey: e.target.value }))}
                placeholder="optional"
              />
            </FormField>
            <DialogFooter>
              <Button variant="secondary" onClick={() => setConnModalOpen(false)}>Cancel</Button>
              <Button onClick={submitConn} disabled={!connForm.id || !connForm.provider || !connForm.baseUrl}>
                {editingConn ? 'Update' : 'Create'}
              </Button>
            </DialogFooter>
          </div>
        </DialogContent>
      </Dialog>

      <Dialog open={modelModalOpen} onOpenChange={setModelModalOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editingModel ? 'Edit Model' : 'New Model'}</DialogTitle>
            <DialogDescription />
          </DialogHeader>
          <div className="space-y-4">
            <FormField label="Model name">
              <Input
                value={modelForm.name}
                onChange={e => setModelForm(f => ({ ...f, name: e.target.value }))}
                placeholder="gpt-4o"
              />
            </FormField>
            <FormField label="Connection">
              <select
                value={modelForm.connectionId}
                onChange={e => setModelForm(f => ({ ...f, connectionId: e.target.value }))}
                className="flex w-full rounded-lg border border-zinc-300 dark:border-zinc-700 bg-white dark:bg-zinc-800 px-3 py-2 text-sm text-zinc-900 dark:text-zinc-100 focus:outline-none focus:border-teal-500/50"
              >
                <option value="">Select connection...</option>
                {connections.map(c => (
                  <option key={c.id} value={c.id}>{c.id} ({c.provider})</option>
                ))}
              </select>
            </FormField>
            <FormField label="Thinking mode">
              <select
                value={modelForm.thinkingMode}
                onChange={e => setModelForm(f => ({ ...f, thinkingMode: e.target.value }))}
                className="flex w-full rounded-lg border border-zinc-300 dark:border-zinc-700 bg-white dark:bg-zinc-800 px-3 py-2 text-sm text-zinc-900 dark:text-zinc-100 focus:outline-none focus:border-teal-500/50"
              >
                <option value="">None (default)</option>
                <option value="skip">Skip — remove thinking blocks</option>
                <option value="inline">Inline — strip tags, keep content</option>
              </select>
            </FormField>
            <DialogFooter>
              <Button variant="secondary" onClick={() => setModelModalOpen(false)}>Cancel</Button>
              <Button onClick={submitModel} disabled={!modelForm.name || !modelForm.connectionId}>
                {editingModel ? 'Update' : 'Create'}
              </Button>
            </DialogFooter>
          </div>
        </DialogContent>
      </Dialog>

      <ConfirmDelete
        open={!!deleteConnTarget}
        onCancel={() => setDeleteConnTarget(null)}
        onConfirm={() => deleteConnTarget && removeConn(deleteConnTarget)}
        title="Delete connection?"
        description="This will delete the connection and all its models."
      />
      <ConfirmDelete
        open={!!deleteModelTarget}
        onCancel={() => setDeleteModelTarget(null)}
        onConfirm={() => deleteModelTarget && removeModel(deleteModelTarget)}
        title="Delete model?"
        description="This will permanently remove the model."
      />
    </div>
  )
}
