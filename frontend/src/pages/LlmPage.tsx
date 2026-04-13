import { useState, useEffect, useCallback, useMemo } from 'react'
import { Plus, Pencil, Trash2, Link2, Box, ChevronDown, ChevronRight, Copy, Layers, Sparkles } from 'lucide-react'
import { toast } from 'sonner'
import { api } from '../api'
import type { LlmConnection, Model, Preset } from '../types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Badge } from '@/components/ui/badge'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/dialog'
import { EmptyState } from '@/components/EmptyState'
import { FormField } from '@/components/FormField'
import { ConfirmDelete } from '@/components/ConfirmDelete'

type ConnForm = { id: string; provider: string; baseUrl: string; apiKey: string }
type ModelForm = { name: string; connectionId: string; thinkingMode: string }
type PresetForm = {
  name: string; chatModelId: string; summaryModelId: string
  imageModelId: string; fallbackModelId: string; temperature: string; systemPrompt: string
}

const emptyPresetForm: PresetForm = {
  name: '', chatModelId: '', summaryModelId: '', imageModelId: '',
  fallbackModelId: '', temperature: '', systemPrompt: '',
}

const selectClass = 'flex w-full rounded-lg border border-zinc-300 dark:border-zinc-700 bg-white dark:bg-zinc-800 px-3 py-2 text-sm text-zinc-900 dark:text-zinc-100 focus:outline-none focus:border-teal-500/50'

export default function LlmPage() {
  const [connections, setConnections] = useState<LlmConnection[]>([])
  const [models, setModels] = useState<Model[]>([])
  const [presets, setPresets] = useState<Preset[]>([])
  const [slots, setSlots] = useState<{ chatPresetId: string; serverPresetId: string }>({ chatPresetId: '', serverPresetId: '' })
  const [loading, setLoading] = useState(true)
  const [expandedConns, setExpandedConns] = useState<Set<string>>(new Set())

  const [connModalOpen, setConnModalOpen] = useState(false)
  const [editingConn, setEditingConn] = useState<LlmConnection | null>(null)
  const [connForm, setConnForm] = useState<ConnForm>({ id: '', provider: 'openai', baseUrl: '', apiKey: '' })

  const [modelModalOpen, setModelModalOpen] = useState(false)
  const [editingModel, setEditingModel] = useState<Model | null>(null)
  const [modelForm, setModelForm] = useState<ModelForm>({ name: '', connectionId: '', thinkingMode: '' })

  const [presetModalOpen, setPresetModalOpen] = useState(false)
  const [editingPreset, setEditingPreset] = useState<Preset | null>(null)
  const [presetForm, setPresetForm] = useState<PresetForm>(emptyPresetForm)

  const [deleteConnTarget, setDeleteConnTarget] = useState<string | null>(null)
  const [deleteModelTarget, setDeleteModelTarget] = useState<string | null>(null)
  const [deletePresetTarget, setDeletePresetTarget] = useState<string | null>(null)

  const load = useCallback(async () => {
    try {
      setLoading(true)
      const [conns, mdls, prs, settings] = await Promise.all([
        api.llmConnections.list(), api.models.list(), api.presets.list(), api.settings.get(),
      ])
      setConnections(conns)
      setModels(mdls)
      setPresets(prs)
      setSlots({ chatPresetId: settings.chatPresetId || '', serverPresetId: settings.serverPresetId || '' })
      if (conns.length > 0 && expandedConns.size === 0) {
        setExpandedConns(new Set(conns.map(c => c.id)))
      }
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Failed to load')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])

  const toggleConn = (id: string) => setExpandedConns(prev => {
    const next = new Set(prev)
    next.has(id) ? next.delete(id) : next.add(id)
    return next
  })

  const modelLabel = useCallback((modelId: string) => {
    if (!modelId) return '—'
    const m = models.find(x => x.id === modelId)
    if (!m) return modelId.slice(0, 12) + '...'
    const c = connections.find(x => x.id === m.connectionId)
    return `${m.name}${c ? ` (${c.id})` : ''}`
  }, [models, connections])

  // --- Connection CRUD ---
  const openCreateConn = () => { setEditingConn(null); setConnForm({ id: '', provider: 'openai', baseUrl: '', apiKey: '' }); setConnModalOpen(true) }
  const openEditConn = (c: LlmConnection) => { setEditingConn(c); setConnForm({ id: c.id, provider: c.provider, baseUrl: c.baseUrl, apiKey: c.apiKey }); setConnModalOpen(true) }
  const submitConn = async () => {
    try {
      if (editingConn) { await api.llmConnections.update(editingConn.id, { provider: connForm.provider, baseUrl: connForm.baseUrl, apiKey: connForm.apiKey }); toast.success('Provider updated') }
      else { await api.llmConnections.create(connForm); toast.success('Provider created') }
      setConnModalOpen(false); load()
    } catch (e: unknown) { toast.error(e instanceof Error ? e.message : 'Operation failed') }
  }
  const removeConn = async (id: string) => {
    try {
      for (const m of models.filter(m => m.connectionId === id)) await api.models.delete(m.id)
      await api.llmConnections.delete(id); toast.success('Provider deleted'); setDeleteConnTarget(null); load()
    } catch (e: unknown) { toast.error(e instanceof Error ? e.message : 'Delete failed') }
  }

  // --- Model CRUD ---
  const openCreateModel = (connectionId: string) => { setEditingModel(null); setModelForm({ name: '', connectionId, thinkingMode: '' }); setModelModalOpen(true) }
  const openEditModel = (m: Model) => { setEditingModel(m); setModelForm({ name: m.name, connectionId: m.connectionId, thinkingMode: m.thinkingMode }); setModelModalOpen(true) }
  const submitModel = async () => {
    try {
      if (editingModel) { await api.models.update(editingModel.id, modelForm.connectionId, modelForm.name, modelForm.thinkingMode); toast.success('Model updated') }
      else { await api.models.create(modelForm.connectionId, modelForm.name, modelForm.thinkingMode); toast.success('Model created') }
      setModelModalOpen(false); load()
    } catch (e: unknown) { toast.error(e instanceof Error ? e.message : 'Operation failed') }
  }
  const removeModel = async (id: string) => {
    try { await api.models.delete(id); toast.success('Model deleted'); setDeleteModelTarget(null); load() }
    catch (e: unknown) { toast.error(e instanceof Error ? e.message : 'Delete failed') }
  }

  // --- Preset CRUD ---
  const openCreatePreset = () => { setEditingPreset(null); setPresetForm(emptyPresetForm); setPresetModalOpen(true) }
  const openEditPreset = (p: Preset) => {
    setEditingPreset(p)
    setPresetForm({
      name: p.name, chatModelId: p.chatModelId, summaryModelId: p.summaryModelId,
      imageModelId: p.imageModelId, fallbackModelId: p.fallbackModelId,
      temperature: p.temperature != null ? String(p.temperature) : '', systemPrompt: p.systemPrompt,
    })
    setPresetModalOpen(true)
  }
  const submitPreset = async () => {
    try {
      const payload = {
        name: presetForm.name, chatModelId: presetForm.chatModelId, summaryModelId: presetForm.summaryModelId,
        imageModelId: presetForm.imageModelId, fallbackModelId: presetForm.fallbackModelId,
        temperature: presetForm.temperature ? parseFloat(presetForm.temperature) : null, systemPrompt: presetForm.systemPrompt,
      }
      if (editingPreset) { await api.presets.update(editingPreset.id, payload); toast.success('Preset updated') }
      else { await api.presets.create(payload); toast.success('Preset created') }
      setPresetModalOpen(false); load()
    } catch (e: unknown) { toast.error(e instanceof Error ? e.message : 'Operation failed') }
  }
  const removePreset = async (id: string) => {
    try { await api.presets.delete(id); toast.success('Preset deleted'); setDeletePresetTarget(null); load() }
    catch (e: unknown) { toast.error(e instanceof Error ? e.message : 'Delete failed') }
  }

  // --- Defaults ---
  const updateSlot = async (key: 'chatPresetId' | 'serverPresetId', value: string) => {
    try {
      const next = { ...slots, [key]: value }
      setSlots(next)
      const current = await api.settings.get()
      await api.settings.update({
        chatPresetId: key === 'chatPresetId' ? value : current.chatPresetId,
        serverPresetId: key === 'serverPresetId' ? value : current.serverPresetId,
        memoryEnabled: current.memoryEnabled,
        userMemories: current.userMemories ?? [],
      })
      toast.success('Default updated')
    } catch (e: unknown) { toast.error(e instanceof Error ? e.message : 'Failed to update') }
  }

  const ModelSelect = useMemo(() => {
    return function MS({ value, onChange, allowEmpty, emptyLabel }: { value: string; onChange: (v: string) => void; allowEmpty?: boolean; emptyLabel?: string }) {
      return (
        <select value={value} onChange={e => onChange(e.target.value)} className={selectClass}>
          <option value="">{allowEmpty ? (emptyLabel ?? 'None') : 'Select model...'}</option>
          {models.map(m => {
            const c = connections.find(x => x.id === m.connectionId)
            return <option key={m.id} value={m.id}>{m.name}{c ? ` (${c.id})` : ''}</option>
          })}
        </select>
      )
    }
  }, [models, connections])

  if (loading) return <div className="p-6 text-center py-12 text-zinc-500 dark:text-zinc-600 text-sm">Loading...</div>

  const hasProviders = connections.length > 0
  const hasModels = models.length > 0
  const hasPresets = presets.length > 0

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-lg font-semibold text-zinc-900 dark:text-zinc-100">AI Engine</h1>
        <p className="text-xs text-zinc-500 dark:text-zinc-600 mt-0.5">Configure providers, models, and presets — everything that powers the AI agent</p>
      </div>

      {/* Step indicator for empty state */}
      {!hasProviders && (
        <div className="mb-6 bg-teal-500/5 border border-teal-500/20 rounded-lg p-5">
          <h2 className="text-sm font-semibold text-teal-600 dark:text-teal-400 mb-1">Getting started</h2>
          <p className="text-xs text-zinc-600 dark:text-zinc-400 mb-3">Set up AI in 3 steps:</p>
          <div className="flex items-center gap-3 text-xs">
            <span className="flex items-center gap-1.5"><span className="w-5 h-5 rounded-full bg-teal-500 text-white flex items-center justify-center text-[10px] font-bold">1</span> Add a provider</span>
            <span className="text-zinc-400">→</span>
            <span className="flex items-center gap-1.5 text-zinc-400"><span className="w-5 h-5 rounded-full bg-zinc-200 dark:bg-zinc-700 flex items-center justify-center text-[10px] font-bold">2</span> Register models</span>
            <span className="text-zinc-400">→</span>
            <span className="flex items-center gap-1.5 text-zinc-400"><span className="w-5 h-5 rounded-full bg-zinc-200 dark:bg-zinc-700 flex items-center justify-center text-[10px] font-bold">3</span> Create a preset</span>
          </div>
        </div>
      )}

      {/* Defaults bar */}
      {hasPresets && (
        <div className="mb-6 bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800 p-4">
          <div className="flex items-center gap-2 mb-3">
            <Sparkles size={14} className="text-teal-500" />
            <h2 className="text-sm font-semibold text-zinc-800 dark:text-zinc-200">Active Presets</h2>
            <span className="text-[11px] text-zinc-500 dark:text-zinc-600">— applied everywhere unless overridden on a channel or server</span>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {([
              { key: 'chatPresetId' as const, label: 'Chat & Telegram', hint: 'Web chat, Telegram, plan execution' },
              { key: 'serverPresetId' as const, label: 'Servers & SSH', hint: 'All SSH connections and tool execution' },
            ]).map(({ key, label, hint }) => (
              <div key={key}>
                <label className="block text-xs font-medium text-zinc-500 dark:text-zinc-400 mb-1.5">{label}</label>
                <select
                  value={slots[key]}
                  onChange={e => updateSlot(key, e.target.value)}
                  className={`w-full px-3 py-2 border rounded-lg text-sm bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 focus:outline-none focus:border-teal-500/50 border-zinc-300 dark:border-zinc-700`}
                >
                  <option value="">None</option>
                  {presets.map(p => <option key={p.id} value={p.id}>{p.name}</option>)}
                </select>
                <p className="text-[11px] text-zinc-500 dark:text-zinc-600 mt-1">{hint}</p>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Two-column layout: Providers+Models | Presets */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* --- LEFT: Providers & Models --- */}
        <div>
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <Link2 size={14} className="text-teal-500" />
              <h2 className="text-sm font-semibold text-zinc-800 dark:text-zinc-200">Providers & Models</h2>
            </div>
            <Button size="sm" variant="secondary" onClick={openCreateConn}>
              <Plus size={13} /> Provider
            </Button>
          </div>

          {!hasProviders ? (
            <EmptyState icon={Link2} title="No AI providers" description="Add an OpenAI-compatible endpoint to get started" />
          ) : (
            <div className="space-y-2.5">
              {connections.map(conn => {
                const connModels = models.filter(m => m.connectionId === conn.id)
                const isExpanded = expandedConns.has(conn.id)
                return (
                  <div key={conn.id} className="bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800 overflow-hidden">
                    <div className="px-4 py-3">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2 min-w-0 cursor-pointer select-none" onClick={() => toggleConn(conn.id)}>
                          {isExpanded ? <ChevronDown size={14} className="text-zinc-500 shrink-0" /> : <ChevronRight size={14} className="text-zinc-500 shrink-0" />}
                          <span className="font-medium text-zinc-800 dark:text-zinc-200 text-sm">{conn.id}</span>
                          <Badge variant="muted">{conn.provider}</Badge>
                          <span className="text-[11px] text-zinc-500 dark:text-zinc-600 truncate hidden sm:inline">{conn.baseUrl}</span>
                        </div>
                        <div className="flex items-center gap-0.5 ml-2 shrink-0">
                          <span className="text-[11px] text-zinc-500 mr-1">{connModels.length}</span>
                          <Button variant="ghost" size="icon" onClick={() => openEditConn(conn)}><Pencil size={13} /></Button>
                          <Button variant="destructive" size="icon" onClick={() => setDeleteConnTarget(conn.id)}><Trash2 size={13} /></Button>
                        </div>
                      </div>
                    </div>
                    {isExpanded && (
                      <div className="border-t border-zinc-200 dark:border-zinc-800">
                        {connModels.length === 0 ? (
                          <div className="px-4 py-3 text-center">
                            <p className="text-xs text-zinc-500 mb-2">No models yet — add the model names your provider serves</p>
                            <Button size="sm" variant="ghost" onClick={() => openCreateModel(conn.id)}><Plus size={12} /> Add Model</Button>
                          </div>
                        ) : (
                          <div>
                            <div className="divide-y divide-zinc-200/50 dark:divide-zinc-800/50">
                              {connModels.map(m => (
                                <div key={m.id} className="px-4 py-2 flex items-center justify-between hover:bg-zinc-50 dark:hover:bg-zinc-800/30">
                                  <div className="flex items-center gap-2 min-w-0">
                                    <Box size={12} className="text-zinc-500 shrink-0" />
                                    <span className="text-sm text-zinc-700 dark:text-zinc-300 font-medium">{m.name}</span>
                                    {m.thinkingMode && <Badge variant="secondary">{m.thinkingMode}</Badge>}
                                    <button
                                      onClick={() => { navigator.clipboard.writeText(m.id); toast.success('ID copied') }}
                                      className="inline-flex items-center gap-1 text-[10px] text-zinc-400 font-mono hover:text-zinc-600 dark:hover:text-zinc-400"
                                      title={m.id}
                                    >
                                      {m.id.slice(0, 10)}… <Copy size={10} />
                                    </button>
                                  </div>
                                  <div className="flex gap-0.5">
                                    <Button variant="ghost" size="icon" onClick={() => openEditModel(m)}><Pencil size={12} /></Button>
                                    <Button variant="destructive" size="icon" onClick={() => setDeleteModelTarget(m.id)}><Trash2 size={12} /></Button>
                                  </div>
                                </div>
                              ))}
                            </div>
                            <div className="px-4 py-2 border-t border-zinc-100 dark:border-zinc-800/50">
                              <button onClick={() => openCreateModel(conn.id)} className="inline-flex items-center gap-1 text-[11px] font-medium text-zinc-500 hover:text-teal-500">
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
        </div>

        {/* --- RIGHT: Presets --- */}
        <div>
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <Layers size={14} className="text-teal-500" />
              <h2 className="text-sm font-semibold text-zinc-800 dark:text-zinc-200">Presets</h2>
              <span className="text-[11px] text-zinc-500 dark:text-zinc-600">— which model does what</span>
            </div>
            <Button size="sm" variant="secondary" onClick={openCreatePreset} disabled={!hasModels}>
              <Plus size={13} /> Preset
            </Button>
          </div>

          {!hasModels ? (
            <div className="bg-zinc-50 dark:bg-zinc-900/50 border border-dashed border-zinc-300 dark:border-zinc-700 rounded-lg p-6 text-center">
              <Layers size={24} className="mx-auto text-zinc-400 mb-2" />
              <p className="text-sm text-zinc-500">Add a provider and models first</p>
              <p className="text-xs text-zinc-400 mt-1">Presets assign models to roles: chat, summary, vision, fallback</p>
            </div>
          ) : !hasPresets ? (
            <div className="bg-zinc-50 dark:bg-zinc-900/50 border border-dashed border-zinc-300 dark:border-zinc-700 rounded-lg p-6 text-center">
              <Layers size={24} className="mx-auto text-zinc-400 mb-2" />
              <p className="text-sm text-zinc-500">No presets yet</p>
              <p className="text-xs text-zinc-400 mt-1">A preset maps roles (chat, summary, vision) to your models</p>
              <Button size="sm" className="mt-3" onClick={openCreatePreset}><Plus size={13} /> Create Preset</Button>
            </div>
          ) : (
            <div className="space-y-2.5">
              {presets.map(p => {
                const isChat = slots.chatPresetId === p.id
                const isServer = slots.serverPresetId === p.id
                return (
                  <div key={p.id} className="bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800 overflow-hidden">
                    <div className="px-4 py-3">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2 min-w-0">
                          <span className="font-medium text-zinc-800 dark:text-zinc-200 text-sm">{p.name}</span>
                          {isChat && <Badge variant="default" className="text-[10px] bg-teal-500/15 text-teal-600 dark:text-teal-400 border-teal-500/30">chat</Badge>}
                          {isServer && <Badge variant="default" className="text-[10px] bg-blue-500/15 text-blue-600 dark:text-blue-400 border-blue-500/30">server</Badge>}
                        </div>
                        <div className="flex gap-0.5 ml-2 shrink-0">
                          <Button variant="ghost" size="icon" onClick={() => openEditPreset(p)}><Pencil size={13} /></Button>
                          <Button variant="destructive" size="icon" onClick={() => setDeletePresetTarget(p.id)}><Trash2 size={13} /></Button>
                        </div>
                      </div>
                      <div className="mt-2 grid grid-cols-2 gap-x-4 gap-y-1 text-[12px]">
                        <div className="flex justify-between"><span className="text-zinc-500">Chat</span><span className="text-zinc-700 dark:text-zinc-300 text-right truncate ml-2">{modelLabel(p.chatModelId)}</span></div>
                        <div className="flex justify-between"><span className="text-zinc-500">Summary</span><span className="text-zinc-700 dark:text-zinc-300 text-right truncate ml-2">{modelLabel(p.summaryModelId)}</span></div>
                        {p.imageModelId && <div className="flex justify-between"><span className="text-zinc-500">Vision</span><span className="text-zinc-700 dark:text-zinc-300 text-right truncate ml-2">{modelLabel(p.imageModelId)}</span></div>}
                        {p.fallbackModelId && <div className="flex justify-between"><span className="text-zinc-500">Fallback</span><span className="text-zinc-700 dark:text-zinc-300 text-right truncate ml-2">{modelLabel(p.fallbackModelId)}</span></div>}
                        {p.temperature != null && <div className="flex justify-between"><span className="text-zinc-500">Temperature</span><span className="text-zinc-700 dark:text-zinc-300">{p.temperature}</span></div>}
                      </div>
                      {p.systemPrompt && <p className="text-[11px] text-zinc-400 mt-1.5 truncate">{p.systemPrompt}</p>}
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </div>
      </div>

      {/* --- Connection Modal --- */}
      <Dialog open={connModalOpen} onOpenChange={setConnModalOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editingConn ? 'Edit Provider' : 'New Provider'}</DialogTitle>
            <DialogDescription>An OpenAI-compatible API endpoint (works with OpenAI, Anthropic via proxy, LM Studio, Ollama, etc.)</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <FormField label="ID" hint="Short identifier, e.g. openai, local, anthropic">
              <Input value={connForm.id} disabled={!!editingConn} onChange={e => setConnForm(f => ({ ...f, id: e.target.value }))} placeholder="openai" />
            </FormField>
            <FormField label="Provider">
              <select value={connForm.provider} onChange={e => setConnForm(f => ({ ...f, provider: e.target.value }))} className={selectClass}>
                <option value="openai">openai</option>
              </select>
            </FormField>
            <FormField label="Base URL">
              <Input value={connForm.baseUrl} onChange={e => setConnForm(f => ({ ...f, baseUrl: e.target.value }))} placeholder="https://api.openai.com/v1" />
            </FormField>
            <FormField label="API Key" hint="Leave empty for local providers">
              <Input type="password" value={connForm.apiKey} onChange={e => setConnForm(f => ({ ...f, apiKey: e.target.value }))} placeholder="sk-..." />
            </FormField>
            <DialogFooter>
              <Button variant="secondary" onClick={() => setConnModalOpen(false)}>Cancel</Button>
              <Button onClick={submitConn} disabled={!connForm.id || !connForm.baseUrl}>{editingConn ? 'Update' : 'Create'}</Button>
            </DialogFooter>
          </div>
        </DialogContent>
      </Dialog>

      {/* --- Model Modal --- */}
      <Dialog open={modelModalOpen} onOpenChange={setModelModalOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editingModel ? 'Edit Model' : 'Add Model'}</DialogTitle>
            <DialogDescription>The exact model name as your provider expects it</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <FormField label="Model name">
              <Input value={modelForm.name} onChange={e => setModelForm(f => ({ ...f, name: e.target.value }))} placeholder="gpt-4o" />
            </FormField>
            <FormField label="Provider">
              <select value={modelForm.connectionId} onChange={e => setModelForm(f => ({ ...f, connectionId: e.target.value }))} className={selectClass}>
                <option value="">Select...</option>
                {connections.map(c => <option key={c.id} value={c.id}>{c.id} ({c.provider})</option>)}
              </select>
            </FormField>
            <FormField label="Thinking mode" hint="How to handle <think> blocks in model output">
              <select value={modelForm.thinkingMode} onChange={e => setModelForm(f => ({ ...f, thinkingMode: e.target.value }))} className={selectClass}>
                <option value="">None (default)</option>
                <option value="skip">Skip — remove thinking blocks</option>
                <option value="inline">Inline — strip tags, keep content</option>
              </select>
            </FormField>
            <DialogFooter>
              <Button variant="secondary" onClick={() => setModelModalOpen(false)}>Cancel</Button>
              <Button onClick={submitModel} disabled={!modelForm.name || !modelForm.connectionId}>{editingModel ? 'Update' : 'Create'}</Button>
            </DialogFooter>
          </div>
        </DialogContent>
      </Dialog>

      {/* --- Preset Modal --- */}
      <Dialog open={presetModalOpen} onOpenChange={setPresetModalOpen}>
        <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>{editingPreset ? 'Edit Preset' : 'New Preset'}</DialogTitle>
            <DialogDescription>A preset assigns models to roles. You can create multiple presets and switch between them.</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <FormField label="Name">
              <Input value={presetForm.name} onChange={e => setPresetForm(f => ({ ...f, name: e.target.value }))} placeholder="e.g. Fast, Smart, Creative" />
            </FormField>
            <FormField label="Chat model" hint="Primary model for conversations and tool use">
              <ModelSelect value={presetForm.chatModelId} onChange={v => setPresetForm(f => ({ ...f, chatModelId: v }))} />
            </FormField>
            <FormField label="Summary model" hint="For memory extraction and session titles. Falls back to chat model">
              <ModelSelect value={presetForm.summaryModelId} onChange={v => setPresetForm(f => ({ ...f, summaryModelId: v }))} allowEmpty emptyLabel="Same as chat model" />
            </FormField>
            <FormField label="Image / Vision model" hint="For describing images. Optional">
              <ModelSelect value={presetForm.imageModelId} onChange={v => setPresetForm(f => ({ ...f, imageModelId: v }))} allowEmpty emptyLabel="None" />
            </FormField>
            <FormField label="Fallback model" hint="Used when chat model fails. Optional">
              <ModelSelect value={presetForm.fallbackModelId} onChange={v => setPresetForm(f => ({ ...f, fallbackModelId: v }))} allowEmpty emptyLabel="None" />
            </FormField>
            <FormField label="Temperature" hint="0 = deterministic, 1 = creative. Empty = model default">
              <Input type="number" step="0.1" min="0" max="2" value={presetForm.temperature} onChange={e => setPresetForm(f => ({ ...f, temperature: e.target.value }))} placeholder="e.g. 0.7" />
            </FormField>
            <FormField label="System prompt" hint="Custom instructions for every conversation. Optional">
              <Textarea value={presetForm.systemPrompt} onChange={e => setPresetForm(f => ({ ...f, systemPrompt: e.target.value }))} className="h-24" placeholder="You are a helpful assistant..." />
            </FormField>
            <DialogFooter>
              <Button variant="secondary" onClick={() => setPresetModalOpen(false)}>Cancel</Button>
              <Button onClick={submitPreset} disabled={!presetForm.name || !presetForm.chatModelId}>{editingPreset ? 'Update' : 'Create'}</Button>
            </DialogFooter>
          </div>
        </DialogContent>
      </Dialog>

      {/* --- Delete confirmations --- */}
      <ConfirmDelete open={!!deleteConnTarget} onCancel={() => setDeleteConnTarget(null)} onConfirm={() => deleteConnTarget && removeConn(deleteConnTarget)} title="Delete provider?" description="This will delete the provider and all its models." />
      <ConfirmDelete open={!!deleteModelTarget} onCancel={() => setDeleteModelTarget(null)} onConfirm={() => deleteModelTarget && removeModel(deleteModelTarget)} title="Delete model?" description="Remove this model. Presets using it will need updating." />
      <ConfirmDelete open={!!deletePresetTarget} onCancel={() => setDeletePresetTarget(null)} onConfirm={() => deletePresetTarget && removePreset(deletePresetTarget)} title="Delete preset?" description="Channels and servers using this preset will fall back to defaults." />
    </div>
  )
}
