import { useState, useEffect, useCallback, useMemo, type ReactNode } from 'react'
import {
  Plus, Pencil, Trash2, Link2, Box, ChevronDown, ChevronRight, Copy,
  Layers, Route as RouteIcon, MessageSquare, FileText, Eye, RotateCcw,
  Wallet, Infinity as InfinityIcon,
} from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import { toast } from 'sonner'
import { api } from '../api'
import type { LlmConnection, Model, Preset, ProviderModel, InferenceLimit } from '../types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Badge } from '@/components/ui/badge'
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter,
} from '@/components/ui/dialog'
import { FormField } from '@/components/FormField'
import { ConfirmDelete } from '@/components/ConfirmDelete'

type EndpointForm = { id: string; provider: string; baseUrl: string; apiKey: string }
type ModelForm = { name: string; connectionId: string; thinkingMode: string; contextWindow: string; reserveTokens: string; compactTokens: string }

const defaultContextWindow = 128000
const defaultReserveTokens = 20000
type ProfileForm = {
  name: string
  chatModelId: string
  summaryModelId: string
  imageModelId: string
  fallbackModelId: string
  temperature: string
  systemPrompt: string
}

const emptyProfileForm: ProfileForm = {
  name: '', chatModelId: '', summaryModelId: '', imageModelId: '',
  fallbackModelId: '', temperature: '', systemPrompt: '',
}

const selectClass =
  'flex w-full rounded-lg border border-zinc-300 dark:border-zinc-700 bg-white dark:bg-zinc-800 px-3 py-2 text-sm text-zinc-900 dark:text-zinc-100 focus:outline-none focus:border-teal-500/50'

type RoleKey = 'chat' | 'summary' | 'image' | 'fallback'
type RoleDef = {
  key: RoleKey
  field: keyof Pick<ProfileForm, 'chatModelId' | 'summaryModelId' | 'imageModelId' | 'fallbackModelId'>
  icon: LucideIcon
  label: string
  hint: string
  required?: boolean
}

const roleDefs: RoleDef[] = [
  { key: 'chat',     field: 'chatModelId',     icon: MessageSquare, label: 'Chat',     hint: 'Main conversations and tool use', required: true },
  { key: 'summary',  field: 'summaryModelId',  icon: FileText,      label: 'Summary',  hint: 'Memory extraction and session titles' },
  { key: 'image',    field: 'imageModelId',    icon: Eye,           label: 'Vision',   hint: 'Used when a user sends an image' },
  { key: 'fallback', field: 'fallbackModelId', icon: RotateCcw,     label: 'Fallback', hint: 'Used when the chat model errors out' },
]

const routingDefs: { key: 'chatPresetId' | 'serverPresetId'; label: string; hint: string }[] = [
  { key: 'chatPresetId',   label: 'Chat & Telegram', hint: 'Web chat, Telegram channels, plan execution' },
  { key: 'serverPresetId', label: 'Servers & SSH',   hint: 'SSH connections and remote tool execution' },
]

function scrollToSection(n: number) {
  document.getElementById(`section-${n}`)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

export default function LlmPage() {
  const [endpoints, setEndpoints] = useState<LlmConnection[]>([])
  const [models, setModels] = useState<Model[]>([])
  const [profiles, setProfiles] = useState<Preset[]>([])
  const [routing, setRouting] = useState({ chatPresetId: '', serverPresetId: '' })
  const [loading, setLoading] = useState(true)
  const [expandedProfiles, setExpandedProfiles] = useState<Set<string>>(new Set())

  const [endpointModalOpen, setEndpointModalOpen] = useState(false)
  const [editingEndpoint, setEditingEndpoint] = useState<LlmConnection | null>(null)
  const [endpointForm, setEndpointForm] = useState<EndpointForm>({ id: '', provider: 'openai', baseUrl: '', apiKey: '' })

  const [modelModalOpen, setModelModalOpen] = useState(false)
  const [editingModel, setEditingModel] = useState<Model | null>(null)
  const [modelForm, setModelForm] = useState<ModelForm>({ name: '', connectionId: '', thinkingMode: '', contextWindow: String(defaultContextWindow), reserveTokens: String(defaultReserveTokens), compactTokens: '' })
  const [loadingAvailableModels, setLoadingAvailableModels] = useState(false)
  const [availableModelsByEndpoint, setAvailableModelsByEndpoint] = useState<Record<string, ProviderModel[]>>({})
  const [endpointLimits, setEndpointLimits] = useState<Record<string, InferenceLimit>>({})

  const [profileModalOpen, setProfileModalOpen] = useState(false)
  const [editingProfile, setEditingProfile] = useState<Preset | null>(null)
  const [profileForm, setProfileForm] = useState<ProfileForm>(emptyProfileForm)

  const [deleteEndpointId, setDeleteEndpointId] = useState<string | null>(null)
  const [deleteModelId, setDeleteModelId] = useState<string | null>(null)
  const [deleteProfileId, setDeleteProfileId] = useState<string | null>(null)

  const load = useCallback(async () => {
    try {
      setLoading(true)
      const [eps, mdls, prs, settings] = await Promise.all([
        api.llmConnections.list(),
        api.models.list(),
        api.presets.list(),
        api.settings.get(),
      ])
      setEndpoints(eps)
      setModels(mdls)
      setProfiles(prs)
      setRouting({
        chatPresetId: settings.chatPresetId || '',
        serverPresetId: settings.serverPresetId || '',
      })
      const limitEntries = await Promise.all(
        eps.map(async ep => {
          try {
            const limit = await api.llmConnections.getInferenceLimit(ep.id)
            return [ep.id, limit] as const
          } catch {
            return [ep.id, { type: 'unlimited', label: 'No inference limit reported' } satisfies InferenceLimit] as const
          }
        }),
      )
      setEndpointLimits(Object.fromEntries(limitEntries))
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Failed to load')
    } finally {
      setLoading(false)
    }
  }, [])

  const loadAvailableModels = useCallback(async (connectionId: string, force = false) => {
    if (!connectionId) return
    if (!force && availableModelsByEndpoint[connectionId]) return
    try {
      setLoadingAvailableModels(true)
      const items = await api.llmConnections.listAvailableModels(connectionId)
      setAvailableModelsByEndpoint(prev => ({ ...prev, [connectionId]: items }))
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Failed to load available models')
    } finally {
      setLoadingAvailableModels(false)
    }
  }, [availableModelsByEndpoint])

  useEffect(() => { load() }, [load])

  useEffect(() => {
    if (!modelModalOpen || !modelForm.connectionId) return
    void loadAvailableModels(modelForm.connectionId)
  }, [modelModalOpen, modelForm.connectionId, loadAvailableModels])

  const endpointById = useMemo(() => new Map(endpoints.map(e => [e.id, e])), [endpoints])
  const modelById = useMemo(() => new Map(models.map(m => [m.id, m])), [models])
  const modelsByEndpoint = useMemo(() => {
    const map = new Map<string, Model[]>()
    endpoints.forEach(e => map.set(e.id, []))
    models.forEach(m => {
      const list = map.get(m.connectionId)
      if (list) list.push(m)
    })
    return map
  }, [endpoints, models])

  const modelLabel = useCallback((modelId: string) => {
    if (!modelId) return ''
    const m = modelById.get(modelId)
    if (!m) return modelId.slice(0, 10) + '…'
    const ep = endpointById.get(m.connectionId)
    return ep ? `${m.name} · ${ep.id}` : m.name
  }, [modelById, endpointById])

  const profileUsage = useCallback((profileId: string) => ({
    chat: routing.chatPresetId === profileId,
    server: routing.serverPresetId === profileId,
  }), [routing])

  const openCreateEndpoint = () => {
    setEditingEndpoint(null)
    setEndpointForm({ id: '', provider: 'openai', baseUrl: '', apiKey: '' })
    setEndpointModalOpen(true)
  }
  const openEditEndpoint = (e: LlmConnection) => {
    setEditingEndpoint(e)
    setEndpointForm({ id: e.id, provider: e.provider, baseUrl: e.baseUrl, apiKey: e.apiKey })
    setEndpointModalOpen(true)
  }
  const submitEndpoint = async () => {
    try {
      if (editingEndpoint) {
        await api.llmConnections.update(editingEndpoint.id, {
          provider: endpointForm.provider,
          baseUrl: endpointForm.baseUrl,
          apiKey: endpointForm.apiKey,
        })
        toast.success('Endpoint updated')
      } else {
        await api.llmConnections.create(endpointForm)
        toast.success('Endpoint created')
      }
      setEndpointModalOpen(false)
      load()
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'Operation failed')
    }
  }
  const removeEndpoint = async (id: string) => {
    try {
      for (const m of models.filter(m => m.connectionId === id)) await api.models.delete(m.id)
      await api.llmConnections.delete(id)
      toast.success('Endpoint deleted')
      setDeleteEndpointId(null)
      load()
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'Delete failed')
    }
  }

  const openCreateModel = (connectionId?: string) => {
    setEditingModel(null)
    const selectedConnection = connectionId ?? endpoints[0]?.id ?? ''
    setModelForm({ name: '', connectionId: selectedConnection, thinkingMode: '', contextWindow: String(defaultContextWindow), reserveTokens: String(defaultReserveTokens), compactTokens: '' })
    setModelModalOpen(true)
  }
  const openEditModel = (m: Model) => {
    setEditingModel(m)
    setModelForm({
      name: m.name,
      connectionId: m.connectionId,
      thinkingMode: m.thinkingMode,
      contextWindow: String(m.contextWindow || defaultContextWindow),
      reserveTokens: String(m.reserveTokens || defaultReserveTokens),
      compactTokens: m.compactTokens ? String(m.compactTokens) : '',
    })
    setModelModalOpen(true)
  }
  const submitModel = async () => {
    try {
      const parseInt10 = (s: string) => Math.max(0, parseInt(s, 10) || 0)
      const payload = {
        connectionId: modelForm.connectionId,
        name: modelForm.name,
        thinkingMode: modelForm.thinkingMode as '' | 'skip' | 'inline',
        contextWindow: parseInt10(modelForm.contextWindow),
        reserveTokens: parseInt10(modelForm.reserveTokens),
        compactTokens: parseInt10(modelForm.compactTokens),
      }
      if (editingModel) {
        await api.models.update(editingModel.id, payload)
        toast.success('Model updated')
      } else {
        await api.models.create(payload)
        toast.success('Model created')
      }
      setModelModalOpen(false)
      load()
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'Operation failed')
    }
  }
  const removeModel = async (id: string) => {
    try {
      await api.models.delete(id)
      toast.success('Model deleted')
      setDeleteModelId(null)
      load()
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'Delete failed')
    }
  }

  const openCreateProfile = () => {
    setEditingProfile(null)
    setProfileForm(emptyProfileForm)
    setProfileModalOpen(true)
  }
  const openEditProfile = (p: Preset) => {
    setEditingProfile(p)
    setProfileForm({
      name: p.name,
      chatModelId: p.chatModelId,
      summaryModelId: p.summaryModelId,
      imageModelId: p.imageModelId,
      fallbackModelId: p.fallbackModelId,
      temperature: p.temperature != null ? String(p.temperature) : '',
      systemPrompt: p.systemPrompt,
    })
    setProfileModalOpen(true)
  }
  const submitProfile = async () => {
    try {
      const payload = {
        name: profileForm.name,
        chatModelId: profileForm.chatModelId,
        summaryModelId: profileForm.summaryModelId,
        imageModelId: profileForm.imageModelId,
        fallbackModelId: profileForm.fallbackModelId,
        temperature: profileForm.temperature ? parseFloat(profileForm.temperature) : null,
        systemPrompt: profileForm.systemPrompt,
      }
      if (editingProfile) {
        await api.presets.update(editingProfile.id, payload)
        toast.success('Profile updated')
      } else {
        await api.presets.create(payload)
        toast.success('Profile created')
      }
      setProfileModalOpen(false)
      load()
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'Operation failed')
    }
  }
  const removeProfile = async (id: string) => {
    try {
      await api.presets.delete(id)
      toast.success('Profile deleted')
      setDeleteProfileId(null)
      load()
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'Delete failed')
    }
  }

  const updateRouting = async (key: 'chatPresetId' | 'serverPresetId', value: string) => {
    try {
      const next = { ...routing, [key]: value }
      setRouting(next)
      const current = await api.settings.get()
      await api.settings.update({
        chatPresetId: key === 'chatPresetId' ? value : current.chatPresetId,
        serverPresetId: key === 'serverPresetId' ? value : current.serverPresetId,
        memoryEnabled: current.memoryEnabled,
        userMemories: current.userMemories ?? [],
      })
      toast.success('Routing updated')
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'Failed to update')
    }
  }

  const toggleProfile = (id: string) => setExpandedProfiles(prev => {
    const next = new Set(prev)
    if (next.has(id)) next.delete(id); else next.add(id)
    return next
  })

  if (loading) {
    return <div className="p-6 text-center py-16 text-sm text-zinc-500 dark:text-zinc-600">Loading…</div>
  }

  const routingFilled = (routing.chatPresetId ? 1 : 0) + (routing.serverPresetId ? 1 : 0)

  return (
    <div className="p-6 max-w-5xl mx-auto">
      <header className="mb-6">
        <h1 className="text-lg font-semibold text-zinc-900 dark:text-zinc-100">AI Engine</h1>
        <p className="text-xs text-zinc-500 dark:text-zinc-600 mt-0.5">
          Configure AI in three steps. Each step builds on the previous one.
        </p>
      </header>

      <Stepper
        steps={[
          { n: 1, label: 'Endpoints', icon: Link2,  count: endpoints.length, done: endpoints.length > 0 },
          { n: 2, label: 'Models',    icon: Box,    count: models.length,    done: models.length > 0 },
          { n: 3, label: 'Profiles',  icon: Layers, count: profiles.length,  done: profiles.length > 0 && routingFilled > 0 },
        ]}
      />

      <div className="mt-6 space-y-4">
        <Section
          n={1}
          icon={Link2}
          title="Endpoints"
          subtitle="Where your LLMs live. Any OpenAI-compatible HTTP API."
          action={<Button size="sm" onClick={openCreateEndpoint}><Plus size={13} /> Add endpoint</Button>}
        >
          {endpoints.length === 0 ? (
            <EmptyHint>
              Add OpenAI, a local Ollama, LM Studio, or any OpenAI-compatible URL to get started.
            </EmptyHint>
          ) : (
            <div className="rounded-lg border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 divide-y divide-zinc-100 dark:divide-zinc-800">
              {endpoints.map(ep => {
                const count = modelsByEndpoint.get(ep.id)?.length ?? 0
                const limit = endpointLimits[ep.id] ?? { type: 'unlimited', label: 'No inference limit reported' }
                const limitType = (limit.type || 'unlimited').toLowerCase()
                const isQuota = limitType === 'quota'
                const isBalance = limitType === 'balance'
                const percent = isQuota
                  ? Math.max(0, Math.min(100, Math.round(limit.percentage ?? 0)))
                  : 100
                const limitLabel = limit.label?.trim() || (isQuota ? 'Quota usage is not available' : 'No inference limit reported')
                return (
                  <div key={ep.id} className="px-4 py-3">
                    <div className="flex items-center gap-3">
                      <Link2 size={13} className="text-zinc-400 shrink-0" />
                      <span className="font-medium text-sm text-zinc-800 dark:text-zinc-200 shrink-0">{ep.id}</span>
                      <Badge variant="muted">{ep.provider}</Badge>
                      <span className="text-[11px] text-zinc-500 dark:text-zinc-600 font-mono truncate flex-1 min-w-0">{ep.baseUrl}</span>
                      <span className="text-[11px] text-zinc-500 whitespace-nowrap shrink-0">
                        {count} {count === 1 ? 'model' : 'models'}
                      </span>
                      <div className="flex gap-0.5 shrink-0">
                        <Button variant="ghost" size="icon" onClick={() => openEditEndpoint(ep)}><Pencil size={13} /></Button>
                        <Button variant="destructive" size="icon" onClick={() => setDeleteEndpointId(ep.id)}><Trash2 size={13} /></Button>
                      </div>
                    </div>
                    <div className="mt-2.5">
                      {isQuota ? (
                        <div className="grid grid-cols-[1fr_auto] items-center gap-x-2 gap-y-1">
                          <div className="h-2.5 w-full rounded-full bg-zinc-200 dark:bg-zinc-800 border border-zinc-300/70 dark:border-zinc-700/70 overflow-hidden">
                            <div
                              className={
                                percent >= 85
                                  ? 'h-full bg-rose-500'
                                  : percent >= 60
                                    ? 'h-full bg-amber-500'
                                    : 'h-full bg-teal-500'
                              }
                              style={{ width: `${percent}%` }}
                            />
                          </div>
                          <span className="text-[10px] font-mono text-zinc-500 dark:text-zinc-600 tabular-nums">{percent}%</span>
                          <p className="col-span-2 text-[10px] text-zinc-500 dark:text-zinc-600">{limitLabel}</p>
                        </div>
                      ) : isBalance ? (
                        <div className="flex items-center gap-1.5 text-[11px] text-zinc-600 dark:text-zinc-400">
                          <Wallet size={12} className="text-teal-500 shrink-0" />
                          <span className="font-medium">{limitLabel}</span>
                        </div>
                      ) : (
                        <div className="flex items-center gap-1.5 text-[11px] text-zinc-500 dark:text-zinc-500">
                          <InfinityIcon size={12} className="shrink-0" />
                          <span>{limitLabel}</span>
                        </div>
                      )}
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </Section>

        <Section
          n={2}
          icon={Box}
          title="Models"
          subtitle="The exact model names your endpoints serve."
          disabled={endpoints.length === 0}
          disabledHint="Add an endpoint first — models belong to an endpoint."
          action={
            <Button size="sm" onClick={() => openCreateModel()} disabled={endpoints.length === 0}>
              <Plus size={13} /> Add model
            </Button>
          }
        >
          {models.length === 0 ? (
            <EmptyHint>
              Register the models your endpoint serves, e.g. <code className="font-mono text-[11px] bg-zinc-100 dark:bg-zinc-800 rounded px-1 py-0.5">gpt-4o</code> or <code className="font-mono text-[11px] bg-zinc-100 dark:bg-zinc-800 rounded px-1 py-0.5">llama3</code>.
            </EmptyHint>
          ) : (
            <div className="space-y-3">
              {endpoints.map(ep => {
                const epModels = modelsByEndpoint.get(ep.id) ?? []
                return (
                  <div key={ep.id}>
                    <div className="flex items-center justify-between px-1 mb-1.5">
                      <div className="flex items-center gap-1.5 text-[11px] font-semibold uppercase tracking-wider text-zinc-500 dark:text-zinc-500">
                        <Link2 size={11} />
                        {ep.id}
                        <span className="text-zinc-400 dark:text-zinc-600 normal-case font-normal tracking-normal">· {epModels.length}</span>
                      </div>
                      <button
                        onClick={() => openCreateModel(ep.id)}
                        className="inline-flex items-center gap-1 text-[11px] font-medium text-zinc-500 hover:text-teal-500"
                      >
                        <Plus size={11} /> Add to {ep.id}
                      </button>
                    </div>
                    {epModels.length === 0 ? (
                      <div className="rounded-lg border border-dashed border-zinc-300 dark:border-zinc-700 px-4 py-3 text-[11px] text-zinc-500 dark:text-zinc-500">
                        No models yet on this endpoint.
                      </div>
                    ) : (
                      <div className="rounded-lg border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 divide-y divide-zinc-100 dark:divide-zinc-800">
                        {epModels.map(m => (
                          <div key={m.id} className="group px-4 py-2.5 flex items-center gap-2">
                            <Box size={12} className="text-zinc-400 shrink-0" />
                            <span className="text-sm text-zinc-800 dark:text-zinc-200 font-medium truncate">{m.name}</span>
                            {m.thinkingMode && <Badge variant="secondary">{m.thinkingMode}</Badge>}
                            <button
                              onClick={() => { navigator.clipboard.writeText(m.id); toast.success('ID copied') }}
                              className="inline-flex items-center gap-1 text-[10px] text-zinc-400 font-mono hover:text-zinc-600 dark:hover:text-zinc-400 opacity-0 group-hover:opacity-100 transition-opacity"
                              title={m.id}
                            >
                              {m.id.slice(0, 8)}… <Copy size={10} />
                            </button>
                            <div className="flex-1" />
                            <div className="flex gap-0.5 shrink-0">
                              <Button variant="ghost" size="icon" onClick={() => openEditModel(m)}><Pencil size={13} /></Button>
                              <Button variant="destructive" size="icon" onClick={() => setDeleteModelId(m.id)}><Trash2 size={13} /></Button>
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                )
              })}
            </div>
          )}
        </Section>

        <Section
          n={3}
          icon={Layers}
          title="Profiles & routing"
          subtitle="Recipes that map models to roles — and where each runs."
          disabled={models.length === 0}
          disabledHint="Register a model first — profiles pick from your models."
          action={
            <Button size="sm" onClick={openCreateProfile} disabled={models.length === 0}>
              <Plus size={13} /> Add profile
            </Button>
          }
        >
          {profiles.length === 0 ? (
            <EmptyHint>
              A profile assigns a model to each role: chat, summary, vision, fallback.
            </EmptyHint>
          ) : (
            <>
              <div className="rounded-lg border border-teal-500/20 bg-teal-500/5 dark:bg-teal-500/[0.06] px-4 py-3 mb-3">
                <div className="flex items-center gap-1.5 mb-2">
                  <RouteIcon size={12} className="text-teal-600 dark:text-teal-400" />
                  <span className="text-[11px] font-semibold uppercase tracking-wider text-teal-700 dark:text-teal-400">
                    Active routing
                  </span>
                </div>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-4 gap-y-2.5">
                  {routingDefs.map(({ key, label, hint }) => (
                    <div key={key} className="min-w-0">
                      <div className="flex items-center gap-2">
                        <label className="text-xs font-medium text-zinc-700 dark:text-zinc-300 shrink-0">{label}</label>
                        <span className="text-zinc-400 shrink-0">→</span>
                        <select
                          value={routing[key]}
                          onChange={e => updateRouting(key, e.target.value)}
                          className="flex-1 min-w-0 rounded-md border border-zinc-300 dark:border-zinc-700 bg-white dark:bg-zinc-900 px-2 py-1 text-xs text-zinc-900 dark:text-zinc-100 focus:outline-none focus:border-teal-500/50"
                        >
                          <option value="">None</option>
                          {profiles.map(p => <option key={p.id} value={p.id}>{p.name}</option>)}
                        </select>
                      </div>
                      <p className="text-[10.5px] text-zinc-500 dark:text-zinc-600 mt-1 ml-0.5">{hint}</p>
                    </div>
                  ))}
                </div>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-1 gap-3">
              {profiles.map(p => {
                const used = profileUsage(p.id)
                const isExpanded = expandedProfiles.has(p.id)
                return (
                  <div key={p.id} className="rounded-lg border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 overflow-hidden">
                    <div className="flex items-center px-4 py-3 gap-2">
                      <button onClick={() => toggleProfile(p.id)} className="flex items-center gap-2 flex-1 min-w-0 text-left">
                        {isExpanded
                          ? <ChevronDown size={14} className="text-zinc-500 shrink-0" />
                          : <ChevronRight size={14} className="text-zinc-500 shrink-0" />}
                        <span className="font-medium text-sm text-zinc-800 dark:text-zinc-200 truncate">{p.name}</span>
                        {used.chat && <Badge variant="success">chat</Badge>}
                        {used.server && <Badge>server</Badge>}
                      </button>
                      <div className="flex gap-0.5 shrink-0">
                        <Button variant="ghost" size="icon" onClick={() => openEditProfile(p)}><Pencil size={13} /></Button>
                        <Button variant="destructive" size="icon" onClick={() => setDeleteProfileId(p.id)}><Trash2 size={13} /></Button>
                      </div>
                    </div>
                    <div className="border-t border-zinc-200/60 dark:border-zinc-800/60 px-4 py-3 space-y-1.5">
                      {roleDefs.map(({ key, field, icon: Icon, label }) => {
                        const resolved = modelLabel(p[field])
                        const placeholder = key === 'summary' ? 'same as chat' : '—'
                        return (
                          <div key={key} className="flex items-center gap-2 text-[12px]">
                            <Icon size={12} className="text-zinc-400 shrink-0" />
                            <span className="text-zinc-500 w-16 shrink-0">{label}</span>
                            <span className={resolved ? 'text-zinc-700 dark:text-zinc-300 truncate' : 'text-zinc-400 dark:text-zinc-600'}>
                              {resolved || placeholder}
                            </span>
                          </div>
                        )
                      })}
                      {isExpanded && (
                        <div className="mt-2 pt-2 border-t border-zinc-200/60 dark:border-zinc-800/60 space-y-1.5 text-[11px]">
                          <div className="flex justify-between">
                            <span className="text-zinc-500">Temperature</span>
                            <span className="text-zinc-700 dark:text-zinc-300">{p.temperature != null ? p.temperature : 'model default'}</span>
                          </div>
                          {p.systemPrompt && (
                            <div>
                              <div className="text-zinc-500 mb-1">System prompt</div>
                              <div className="text-zinc-600 dark:text-zinc-400 bg-zinc-50 dark:bg-zinc-950 rounded p-2 font-mono text-[10.5px] leading-relaxed whitespace-pre-wrap">
                                {p.systemPrompt}
                              </div>
                            </div>
                          )}
                        </div>
                      )}
                    </div>
                  </div>
                )
              })}
              </div>
            </>
          )}
        </Section>
      </div>

      <Dialog open={endpointModalOpen} onOpenChange={setEndpointModalOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editingEndpoint ? 'Edit endpoint' : 'New endpoint'}</DialogTitle>
            <DialogDescription>
              Any OpenAI-compatible URL or a Gonka Source URL.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <FormField label="ID" hint="Short identifier, e.g. openai, local, anthropic">
              <Input
                value={endpointForm.id}
                disabled={!!editingEndpoint}
                onChange={e => setEndpointForm(f => ({ ...f, id: e.target.value }))}
                placeholder="openai"
              />
            </FormField>
            <FormField label="Provider">
              <select
                value={endpointForm.provider}
                onChange={e => setEndpointForm(f => ({ ...f, provider: e.target.value }))}
                className={selectClass}
              >
                <option value="openai">openai</option>
                <option value="gonka">gonka</option>
              </select>
            </FormField>
            <FormField
              label={endpointForm.provider === 'gonka' ? 'Node URL' : 'Base URL'}
              hint={endpointForm.provider === 'gonka' ? 'Gonka Source URL (NODE_URL)' : undefined}
            >
              <Input
                value={endpointForm.baseUrl}
                onChange={e => setEndpointForm(f => ({ ...f, baseUrl: e.target.value }))}
                placeholder={endpointForm.provider === 'gonka' ? 'https://api.gonka.testnet.example.com' : 'https://api.openai.com/v1'}
              />
            </FormField>
            <FormField
              label={endpointForm.provider === 'gonka' ? 'Private key' : 'API key'}
              hint={endpointForm.provider === 'gonka' ? 'GONKA_PRIVATE_KEY wallet key' : 'Leave empty for local providers'}
            >
              <Input
                type="password"
                value={endpointForm.apiKey}
                onChange={e => setEndpointForm(f => ({ ...f, apiKey: e.target.value }))}
                placeholder={endpointForm.provider === 'gonka' ? '0x…' : 'sk-…'}
              />
            </FormField>
            <DialogFooter>
              <Button variant="secondary" onClick={() => setEndpointModalOpen(false)}>Cancel</Button>
              <Button
                onClick={submitEndpoint}
                disabled={!endpointForm.id || !endpointForm.baseUrl || (endpointForm.provider === 'gonka' && !endpointForm.apiKey)}
              >
                {editingEndpoint ? 'Update' : 'Create'}
              </Button>
            </DialogFooter>
          </div>
        </DialogContent>
      </Dialog>

      <Dialog open={modelModalOpen} onOpenChange={setModelModalOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editingModel ? 'Edit model' : 'Add model'}</DialogTitle>
            <DialogDescription>The model name exactly as your endpoint expects it.</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <FormField label="Endpoint">
              <select
                value={modelForm.connectionId}
                onChange={e => setModelForm(f => ({ ...f, connectionId: e.target.value }))}
                className={selectClass}
              >
                <option value="">Select…</option>
                {endpoints.map(ep => <option key={ep.id} value={ep.id}>{ep.id} · {ep.provider}</option>)}
              </select>
            </FormField>
            <FormField label="Model name" hint="Type manually or pick from loaded endpoint models">
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <Input
                    list="available-models"
                    value={modelForm.name}
                    onFocus={() => modelForm.connectionId && loadAvailableModels(modelForm.connectionId, true)}
                    onChange={e => setModelForm(f => ({ ...f, name: e.target.value }))}
                    placeholder="gpt-4o"
                    disabled={!modelForm.connectionId}
                  />
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={() => modelForm.connectionId && loadAvailableModels(modelForm.connectionId, true)}
                    disabled={!modelForm.connectionId || loadingAvailableModels}
                  >
                    {loadingAvailableModels ? 'Loading...' : 'Refresh'}
                  </Button>
                </div>
                <datalist id="available-models">
                  {(availableModelsByEndpoint[modelForm.connectionId] ?? []).map(m => (
                    <option key={m.id} value={m.id} />
                  ))}
                </datalist>
                <p className="text-[11px] text-zinc-500 dark:text-zinc-600">
                  {modelForm.connectionId
                    ? `${(availableModelsByEndpoint[modelForm.connectionId] ?? []).length} models loaded from ${modelForm.connectionId}. You can still enter any custom model name.`
                    : 'Select endpoint first'}
                </p>
              </div>
            </FormField>
            <FormField label="Reasoning output" hint="How to handle <think> blocks in model responses">
              <select
                value={modelForm.thinkingMode}
                onChange={e => setModelForm(f => ({ ...f, thinkingMode: e.target.value }))}
                className={selectClass}
              >
                <option value="">Keep as-is (default)</option>
                <option value="skip">Remove reasoning blocks</option>
                <option value="inline">Strip tags, keep content</option>
              </select>
            </FormField>
            <FormField label="Context window (tokens)" hint="Model's maximum context length">
              <Input
                type="number"
                min={1000}
                step={1000}
                value={modelForm.contextWindow}
                onChange={e => setModelForm(f => ({ ...f, contextWindow: e.target.value }))}
                placeholder={String(defaultContextWindow)}
              />
            </FormField>
            <FormField label="Reserve for output (tokens)" hint="Headroom left for the reply so prompts never fill the window">
              <Input
                type="number"
                min={0}
                step={1000}
                value={modelForm.reserveTokens}
                onChange={e => setModelForm(f => ({ ...f, reserveTokens: e.target.value }))}
                placeholder={String(defaultReserveTokens)}
              />
            </FormField>
            <FormField label="Compact at (tokens, optional override)" hint="If empty, uses context window − reserve">
              <Input
                type="number"
                min={0}
                step={1000}
                value={modelForm.compactTokens}
                onChange={e => setModelForm(f => ({ ...f, compactTokens: e.target.value }))}
                placeholder="auto"
              />
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

      <Dialog open={profileModalOpen} onOpenChange={setProfileModalOpen}>
        <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-xl">
          <DialogHeader>
            <DialogTitle>{editingProfile ? 'Edit profile' : 'New profile'}</DialogTitle>
            <DialogDescription>Assign a model to each role, then tune behavior.</DialogDescription>
          </DialogHeader>
          <div className="space-y-5">
            <FormField label="Name">
              <Input
                value={profileForm.name}
                onChange={e => setProfileForm(f => ({ ...f, name: e.target.value }))}
                placeholder="e.g. Fast, Smart, Creative"
              />
            </FormField>

            <div>
              <div className="text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-2">Roles</div>
              <div className="rounded-lg border border-zinc-200 dark:border-zinc-800 divide-y divide-zinc-200 dark:divide-zinc-800 bg-white dark:bg-zinc-900">
                {roleDefs.map(({ key, field, icon: Icon, label, hint, required }) => {
                  const emptyLabel = key === 'summary' ? 'Same as chat' : 'None'
                  return (
                    <div key={key} className="px-3 py-3 grid grid-cols-[7.5rem_1fr] gap-x-3 gap-y-1 items-center">
                      <div className="flex items-center gap-2">
                        <Icon size={14} className="text-teal-500" />
                        <span className="text-sm font-medium text-zinc-800 dark:text-zinc-200">{label}</span>
                        {required && <span className="text-[10px] text-rose-500">*</span>}
                      </div>
                      <select
                        value={profileForm[field]}
                        onChange={e => setProfileForm(f => ({ ...f, [field]: e.target.value }))}
                        className={selectClass}
                      >
                        <option value="">{required ? 'Select a model…' : emptyLabel}</option>
                        {models.map(m => {
                          const ep = endpointById.get(m.connectionId)
                          return <option key={m.id} value={m.id}>{m.name}{ep ? ` · ${ep.id}` : ''}</option>
                        })}
                      </select>
                      <div />
                      <p className="text-[11px] text-zinc-500 dark:text-zinc-600">{hint}</p>
                    </div>
                  )
                })}
              </div>
            </div>

            <div>
              <div className="text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-2">Behavior</div>
              <div className="space-y-3">
                <FormField label="Temperature" hint="0 = deterministic · 1 = creative · empty = model default">
                  <Input
                    type="number" step="0.1" min="0" max="2"
                    value={profileForm.temperature}
                    onChange={e => setProfileForm(f => ({ ...f, temperature: e.target.value }))}
                    placeholder="e.g. 0.7"
                  />
                </FormField>
                <FormField label="System prompt" hint="Custom instructions added to every conversation. Optional.">
                  <Textarea
                    value={profileForm.systemPrompt}
                    onChange={e => setProfileForm(f => ({ ...f, systemPrompt: e.target.value }))}
                    className="h-24"
                    placeholder="You are a helpful assistant…"
                  />
                </FormField>
              </div>
            </div>

            <DialogFooter>
              <Button variant="secondary" onClick={() => setProfileModalOpen(false)}>Cancel</Button>
              <Button onClick={submitProfile} disabled={!profileForm.name || !profileForm.chatModelId}>
                {editingProfile ? 'Update' : 'Create'}
              </Button>
            </DialogFooter>
          </div>
        </DialogContent>
      </Dialog>

      <ConfirmDelete
        open={!!deleteEndpointId}
        onCancel={() => setDeleteEndpointId(null)}
        onConfirm={() => deleteEndpointId && removeEndpoint(deleteEndpointId)}
        title="Delete endpoint?"
        description="This removes the endpoint and all its models."
      />
      <ConfirmDelete
        open={!!deleteModelId}
        onCancel={() => setDeleteModelId(null)}
        onConfirm={() => deleteModelId && removeModel(deleteModelId)}
        title="Delete model?"
        description="Profiles using this model will need updating."
      />
      <ConfirmDelete
        open={!!deleteProfileId}
        onCancel={() => setDeleteProfileId(null)}
        onConfirm={() => deleteProfileId && removeProfile(deleteProfileId)}
        title="Delete profile?"
        description="Routing slots using this profile will reset to none."
      />
    </div>
  )
}

type StepItem = { n: number; label: string; icon: LucideIcon; count: number; total?: number; done: boolean }

function Stepper({ steps }: { steps: StepItem[] }) {
  return (
    <nav className="rounded-lg border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 px-2 py-2 flex items-center gap-1">
      {steps.map((s, i) => {
        const Icon = s.icon
        return (
          <div key={s.n} className="flex items-center gap-1 flex-1 min-w-0">
            <button
              onClick={() => scrollToSection(s.n)}
              className="flex items-center gap-2 px-2 py-1.5 rounded-md hover:bg-zinc-100 dark:hover:bg-zinc-800/70 flex-1 min-w-0 text-left"
            >
              <span
                className={`w-5 h-5 rounded-full flex items-center justify-center text-[10px] font-bold shrink-0 ${
                  s.done
                    ? 'bg-teal-500 text-white'
                    : 'bg-zinc-200 dark:bg-zinc-800 text-zinc-500 dark:text-zinc-400'
                }`}
              >
                {s.n}
              </span>
              <Icon size={13} className={s.done ? 'text-teal-500' : 'text-zinc-400'} />
              <span className={`text-xs font-medium truncate ${s.done ? 'text-zinc-800 dark:text-zinc-200' : 'text-zinc-500'}`}>
                {s.label}
              </span>
              <span className="text-[11px] text-zinc-400 ml-auto shrink-0">
                {s.total != null ? `${s.count}/${s.total}` : s.count}
              </span>
            </button>
            {i < steps.length - 1 && <div className="w-3 h-px bg-zinc-200 dark:bg-zinc-800 shrink-0" />}
          </div>
        )
      })}
    </nav>
  )
}

type SectionProps = {
  n: number
  icon: LucideIcon
  title: string
  subtitle: string
  action?: ReactNode
  disabled?: boolean
  disabledHint?: string
  children: ReactNode
}

function Section({ n, icon: Icon, title, subtitle, action, disabled, disabledHint, children }: SectionProps) {
  return (
    <section
      id={`section-${n}`}
      className={`rounded-xl border border-zinc-200 dark:border-zinc-800 bg-zinc-50/50 dark:bg-zinc-900/40 p-4 scroll-mt-4 transition-opacity ${disabled ? 'opacity-60' : ''}`}
    >
      <header className="flex items-start justify-between gap-3 mb-3">
        <div className="flex items-start gap-3 min-w-0">
          <div className="w-6 h-6 rounded-full bg-teal-500/15 text-teal-600 dark:text-teal-400 flex items-center justify-center text-[11px] font-bold shrink-0 mt-0.5">
            {n}
          </div>
          <div className="min-w-0">
            <div className="flex items-center gap-1.5">
              <Icon size={14} className="text-teal-500" />
              <h2 className="text-sm font-semibold text-zinc-900 dark:text-zinc-100">{title}</h2>
            </div>
            <p className="text-[11px] text-zinc-500 dark:text-zinc-600 mt-0.5">
              {disabled && disabledHint ? disabledHint : subtitle}
            </p>
          </div>
        </div>
        {action && <div className="shrink-0">{action}</div>}
      </header>
      {!disabled && children}
    </section>
  )
}

function EmptyHint({ children }: { children: ReactNode }) {
  return (
    <div className="rounded-lg border border-dashed border-zinc-300 dark:border-zinc-700 bg-white/60 dark:bg-zinc-900/40 px-4 py-6 text-center text-xs text-zinc-500 dark:text-zinc-500">
      {children}
    </div>
  )
}
