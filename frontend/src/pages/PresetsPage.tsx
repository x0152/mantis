import { useState, useEffect, useCallback } from 'react'
import { Plus, Pencil, Trash2, Layers, ChevronDown, ChevronRight } from 'lucide-react'
import { toast } from 'sonner'
import { api } from '../api'
import type { Preset, Model, LlmConnection } from '../types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/dialog'
import { EmptyState } from '@/components/EmptyState'
import { FormField } from '@/components/FormField'
import { ConfirmDelete } from '@/components/ConfirmDelete'

type GlobalSlots = { chatPresetId: string | null; serverPresetId: string | null }

type PresetForm = {
  name: string
  chatModelId: string
  summaryModelId: string
  imageModelId: string
  fallbackModelId: string
  temperature: string
  systemPrompt: string
}

const emptyForm: PresetForm = {
  name: '', chatModelId: '', summaryModelId: '', imageModelId: '',
  fallbackModelId: '', temperature: '', systemPrompt: '',
}

function formFromPreset(p: Preset): PresetForm {
  return {
    name: p.name,
    chatModelId: p.chatModelId,
    summaryModelId: p.summaryModelId,
    imageModelId: p.imageModelId,
    fallbackModelId: p.fallbackModelId,
    temperature: p.temperature != null ? String(p.temperature) : '',
    systemPrompt: p.systemPrompt,
  }
}

function formToPayload(f: PresetForm): Omit<Preset, 'id'> {
  return {
    name: f.name,
    chatModelId: f.chatModelId,
    summaryModelId: f.summaryModelId,
    imageModelId: f.imageModelId,
    fallbackModelId: f.fallbackModelId,
    temperature: f.temperature ? parseFloat(f.temperature) : null,
    systemPrompt: f.systemPrompt,
  }
}

const selectClass = 'flex w-full rounded-lg border border-zinc-300 dark:border-zinc-700 bg-white dark:bg-zinc-800 px-3 py-2 text-sm text-zinc-900 dark:text-zinc-100 focus:outline-none focus:border-teal-500/50'

export default function PresetsPage() {
  const [presets, setPresets] = useState<Preset[]>([])
  const [models, setModels] = useState<Model[]>([])
  const [connections, setConnections] = useState<LlmConnection[]>([])
  const [slots, setSlots] = useState<GlobalSlots>({ chatPresetId: null, serverPresetId: null })
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<Preset | null>(null)
  const [form, setForm] = useState<PresetForm>(emptyForm)
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null)
  const [expandedCards, setExpandedCards] = useState<Set<string>>(new Set())

  const load = useCallback(async () => {
    try {
      setLoading(true)
      const [p, m, c, settings] = await Promise.all([api.presets.list(), api.models.list(), api.llmConnections.list(), api.settings.get()])
      setPresets(p)
      setModels(m)
      setConnections(c)
      setSlots({
        chatPresetId: settings.chatPresetId || null,
        serverPresetId: settings.serverPresetId || null,
      })
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Failed to load')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])

  const toggleCard = (id: string) => setExpandedCards(prev => {
    const next = new Set(prev)
    next.has(id) ? next.delete(id) : next.add(id)
    return next
  })

  const openCreate = () => {
    setEditing(null)
    setForm(emptyForm)
    setModalOpen(true)
  }

  const openEdit = (p: Preset) => {
    setEditing(p)
    setForm(formFromPreset(p))
    setModalOpen(true)
  }

  const submit = async () => {
    try {
      const payload = formToPayload(form)
      if (editing) {
        await api.presets.update(editing.id, payload)
        toast.success('Preset updated')
      } else {
        await api.presets.create(payload)
        toast.success('Preset created')
      }
      setModalOpen(false)
      load()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Operation failed')
    }
  }

  const remove = async (id: string) => {
    try {
      await api.presets.delete(id)
      toast.success('Preset deleted')
      setDeleteTarget(null)
      load()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Delete failed')
    }
  }

  const updateSlot = async (key: keyof GlobalSlots, value: string | null) => {
    try {
      const next = { ...slots, [key]: value }
      setSlots(next)
      const current = await api.settings.get()
      await api.settings.update({
        chatPresetId: key === 'chatPresetId' ? (value ?? '') : current.chatPresetId,
        serverPresetId: key === 'serverPresetId' ? (value ?? '') : current.serverPresetId,
        memoryEnabled: current.memoryEnabled,
        userMemories: current.userMemories ?? [],
      })
      toast.success('Default updated')
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Failed to update')
    }
  }

  const modelLabel = (modelId: string) => {
    if (!modelId) return '—'
    const m = models.find(x => x.id === modelId)
    if (!m) return modelId.slice(0, 12) + '...'
    const c = connections.find(x => x.id === m.connectionId)
    return `${m.name}${c ? ` (${c.id})` : ''}`
  }

  const ModelSelect = ({ value, onChange, allowEmpty, emptyLabel }: { value: string; onChange: (v: string) => void; allowEmpty?: boolean; emptyLabel?: string }) => (
    <select value={value} onChange={e => onChange(e.target.value)} className={selectClass}>
      <option value="">{allowEmpty ? (emptyLabel ?? 'None') : 'Select model...'}</option>
      {models.map(m => {
        const c = connections.find(x => x.id === m.connectionId)
        return <option key={m.id} value={m.id}>{m.name}{c ? ` (${c.id})` : ''}</option>
      })}
    </select>
  )

  const slotDefs = [
    { key: 'chatPresetId' as const, label: 'Chat Preset', hint: 'Used for web chat, Telegram channels, cron jobs' },
    { key: 'serverPresetId' as const, label: 'Server Preset', hint: 'Used for all SSH server operations' },
  ]

  return (
    <div className="p-6 max-w-5xl">
      <div className="flex items-center justify-between mb-5">
        <div>
          <h1 className="text-lg font-semibold text-zinc-900 dark:text-zinc-100">Presets</h1>
          <p className="text-xs text-zinc-500 dark:text-zinc-600 mt-0.5">Model configurations for chat, summary, vision, and fallback. Assign globally or override per channel / server</p>
        </div>
        <Button size="sm" onClick={openCreate}>
          <Plus size={14} /> Add Preset
        </Button>
      </div>

      {!loading && presets.length > 0 && (
        <div className="mb-6 bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800 p-4">
          <h2 className="text-sm font-semibold text-zinc-800 dark:text-zinc-200 mb-1">Global Defaults</h2>
          <p className="text-xs text-zinc-500 dark:text-zinc-600 mb-4">Applied everywhere unless overridden on a specific channel or server</p>
          <div className="grid grid-cols-2 gap-4">
            {slotDefs.map(({ key, label, hint }) => {
              const stale = !!slots[key] && !presets.find(p => p.id === slots[key])
              return (
                <div key={key}>
                  <label className="block text-xs font-medium text-zinc-500 dark:text-zinc-400 mb-1.5">{label}</label>
                  <select
                    value={slots[key] ?? ''}
                    onChange={e => updateSlot(key, e.target.value || null)}
                    className={`w-full px-3 py-2 border rounded-lg text-sm bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 focus:outline-none focus:border-teal-500/50 ${stale ? 'border-amber-500/50' : 'border-zinc-300 dark:border-zinc-700'}`}
                  >
                    <option value="">None</option>
                    {presets.map(p => (
                      <option key={p.id} value={p.id}>{p.name}</option>
                    ))}
                  </select>
                  {stale && <p className="text-[11px] text-amber-400 mt-1">Selected preset was deleted</p>}
                  <p className="text-[11px] text-zinc-500 dark:text-zinc-600 mt-1">{hint}</p>
                </div>
              )
            })}
          </div>
        </div>
      )}

      {loading ? (
        <div className="text-center py-12 text-zinc-500 dark:text-zinc-600 text-sm">Loading...</div>
      ) : presets.length === 0 ? (
        <EmptyState icon={Layers} title="No presets yet" description="Create a preset to start managing models centrally" />
      ) : (
        <div className="space-y-2.5">
          {presets.map(p => {
            const isExpanded = expandedCards.has(p.id)
            return (
              <div key={p.id} className="bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800 overflow-hidden">
                <div className="flex items-center px-4 py-3.5 cursor-pointer hover:bg-zinc-100/50 dark:hover:bg-zinc-800/40" onClick={() => toggleCard(p.id)}>
                  <div className="mr-2.5 text-zinc-600">
                    {isExpanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-zinc-800 dark:text-zinc-200 text-sm">{p.name}</span>
                      <span className="text-xs text-zinc-500 dark:text-zinc-600">{modelLabel(p.chatModelId)}</span>
                    </div>
                    {p.systemPrompt && <p className="text-xs text-zinc-500 dark:text-zinc-600 mt-0.5 truncate max-w-md">{p.systemPrompt}</p>}
                  </div>
                  <div className="flex gap-0.5 ml-3" onClick={e => e.stopPropagation()}>
                    <Button variant="ghost" size="icon" onClick={() => openEdit(p)}>
                      <Pencil size={14} />
                    </Button>
                    <Button variant="destructive" size="icon" onClick={() => setDeleteTarget(p.id)}>
                      <Trash2 size={14} />
                    </Button>
                  </div>
                </div>

                {isExpanded && (
                  <div className="border-t border-zinc-200 dark:border-zinc-800 px-4 py-4 bg-zinc-50 dark:bg-zinc-950/50">
                    <div className="grid grid-cols-2 gap-x-6 gap-y-2.5 text-sm">
                      <div className="flex justify-between">
                        <span className="text-zinc-600 dark:text-zinc-500">Chat Model</span>
                        <span className="text-zinc-800 dark:text-zinc-200 text-right">{modelLabel(p.chatModelId)}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-zinc-600 dark:text-zinc-500">Summary Model</span>
                        <span className="text-zinc-800 dark:text-zinc-200 text-right">{modelLabel(p.summaryModelId)}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-zinc-600 dark:text-zinc-500">Image / Vision Model</span>
                        <span className="text-zinc-800 dark:text-zinc-200 text-right">{modelLabel(p.imageModelId)}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-zinc-600 dark:text-zinc-500">Fallback Model</span>
                        <span className="text-zinc-800 dark:text-zinc-200 text-right">{modelLabel(p.fallbackModelId)}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-zinc-600 dark:text-zinc-500">Temperature</span>
                        <span className="text-zinc-800 dark:text-zinc-200">{p.temperature != null ? p.temperature : '—'}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-zinc-600 dark:text-zinc-500">System Prompt</span>
                        <span className="text-zinc-800 dark:text-zinc-200 truncate max-w-[200px]">{p.systemPrompt || '—'}</span>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            )
          })}
        </div>
      )}

      <Dialog open={modalOpen} onOpenChange={setModalOpen}>
        <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>{editing ? 'Edit Preset' : 'New Preset'}</DialogTitle>
            <DialogDescription>Configure which models to use for each task</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <FormField label="Name">
              <Input
                value={form.name}
                onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                placeholder="e.g. Fast, Smart, Creative"
              />
            </FormField>

            <FormField label="Chat Model" hint="Primary model for conversations">
              <ModelSelect value={form.chatModelId} onChange={v => setForm(f => ({ ...f, chatModelId: v }))} />
            </FormField>

            <FormField label="Summary Model" hint="Used for memory extraction and summaries. Falls back to chat model if empty">
              <ModelSelect value={form.summaryModelId} onChange={v => setForm(f => ({ ...f, summaryModelId: v }))} allowEmpty emptyLabel="Same as chat model" />
            </FormField>

            <FormField label="Image / Vision Model" hint="Used for image description (Vision LLM). Optional">
              <ModelSelect value={form.imageModelId} onChange={v => setForm(f => ({ ...f, imageModelId: v }))} allowEmpty emptyLabel="None" />
            </FormField>

            <FormField label="Fallback Model" hint="Used when the primary chat model fails. Optional">
              <ModelSelect value={form.fallbackModelId} onChange={v => setForm(f => ({ ...f, fallbackModelId: v }))} allowEmpty emptyLabel="None" />
            </FormField>

            <FormField label="Temperature" hint="0.0 = deterministic, 1.0 = creative. Leave empty for model default">
              <Input
                type="number"
                step="0.1"
                min="0"
                max="2"
                value={form.temperature}
                onChange={e => setForm(f => ({ ...f, temperature: e.target.value }))}
                placeholder="e.g. 0.7"
              />
            </FormField>

            <FormField label="System Prompt" hint="Custom instructions prepended to every conversation. Optional">
              <Textarea
                value={form.systemPrompt}
                onChange={e => setForm(f => ({ ...f, systemPrompt: e.target.value }))}
                className="h-24"
                placeholder="You are a helpful assistant..."
              />
            </FormField>

            <DialogFooter>
              <Button variant="secondary" onClick={() => setModalOpen(false)}>Cancel</Button>
              <Button onClick={submit} disabled={!form.name || !form.chatModelId}>
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
        title="Delete preset?"
        description="Channels and servers using this preset will fall back to global defaults."
      />
    </div>
  )
}
