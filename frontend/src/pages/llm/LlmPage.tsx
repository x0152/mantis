import { useCallback, useEffect, useMemo, useState } from 'react'
import { Box, Layers, Link2 } from '@/lib/icons'
import { toast } from 'sonner'
import { Stepper } from '@/components/StepperSection'
import { ConfirmDelete } from '@/components/ConfirmDelete'
import { api } from '@/api'
import type { InferenceLimit, LlmConnection, Model, Preset, ProviderModel } from '@/types'
import { EndpointsSection } from './EndpointsSection'
import { ModelsSection } from './ModelsSection'
import { ProfilesSection } from './ProfilesSection'
import { EndpointDialog } from './EndpointDialog'
import { ModelDialog } from './ModelDialog'
import { ProfileDialog } from './ProfileDialog'
import {
  DEFAULT_CONTEXT_WINDOW,
  DEFAULT_RESERVE_TOKENS,
  EMPTY_PROFILE_FORM,
  type EndpointForm,
  type ModelForm,
  type ProfileForm,
} from './types'

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
  const [modelForm, setModelForm] = useState<ModelForm>({
    name: '',
    connectionId: '',
    thinkingMode: '',
    contextWindow: String(DEFAULT_CONTEXT_WINDOW),
    reserveTokens: String(DEFAULT_RESERVE_TOKENS),
    compactTokens: '',
  })
  const [loadingAvailableModels, setLoadingAvailableModels] = useState(false)
  const [availableModelsByEndpoint, setAvailableModelsByEndpoint] = useState<Record<string, ProviderModel[]>>({})
  const [endpointLimits, setEndpointLimits] = useState<Record<string, InferenceLimit>>({})

  const [profileModalOpen, setProfileModalOpen] = useState(false)
  const [editingProfile, setEditingProfile] = useState<Preset | null>(null)
  const [profileForm, setProfileForm] = useState<ProfileForm>(EMPTY_PROFILE_FORM)

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
    setModelForm({
      name: '',
      connectionId: selectedConnection,
      thinkingMode: '',
      contextWindow: String(DEFAULT_CONTEXT_WINDOW),
      reserveTokens: String(DEFAULT_RESERVE_TOKENS),
      compactTokens: '',
    })
    setModelModalOpen(true)
  }
  const openEditModel = (m: Model) => {
    setEditingModel(m)
    setModelForm({
      name: m.name,
      connectionId: m.connectionId,
      thinkingMode: m.thinkingMode,
      contextWindow: String(m.contextWindow || DEFAULT_CONTEXT_WINDOW),
      reserveTokens: String(m.reserveTokens || DEFAULT_RESERVE_TOKENS),
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
    setProfileForm(EMPTY_PROFILE_FORM)
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
    if (next.has(id)) next.delete(id)
    else next.add(id)
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
        <EndpointsSection
          endpoints={endpoints}
          modelsByEndpoint={modelsByEndpoint}
          endpointLimits={endpointLimits}
          onCreate={openCreateEndpoint}
          onEdit={openEditEndpoint}
          onDelete={setDeleteEndpointId}
        />
        <ModelsSection
          endpoints={endpoints}
          models={models}
          modelsByEndpoint={modelsByEndpoint}
          onCreate={openCreateModel}
          onEdit={openEditModel}
          onDelete={setDeleteModelId}
        />
        <ProfilesSection
          profiles={profiles}
          models={models}
          routing={routing}
          expandedProfiles={expandedProfiles}
          onToggleProfile={toggleProfile}
          modelLabel={modelLabel}
          onCreate={openCreateProfile}
          onEdit={openEditProfile}
          onDelete={setDeleteProfileId}
          onRoutingChange={updateRouting}
        />
      </div>

      <EndpointDialog
        open={endpointModalOpen}
        onOpenChange={setEndpointModalOpen}
        editing={editingEndpoint}
        form={endpointForm}
        setForm={setEndpointForm}
        onSubmit={submitEndpoint}
      />
      <ModelDialog
        open={modelModalOpen}
        onOpenChange={setModelModalOpen}
        editing={editingModel}
        form={modelForm}
        setForm={setModelForm}
        onSubmit={submitModel}
        endpoints={endpoints}
        availableModelsByEndpoint={availableModelsByEndpoint}
        loadAvailableModels={loadAvailableModels}
        loadingAvailableModels={loadingAvailableModels}
      />
      <ProfileDialog
        open={profileModalOpen}
        onOpenChange={setProfileModalOpen}
        editing={editingProfile}
        form={profileForm}
        setForm={setProfileForm}
        onSubmit={submitProfile}
        models={models}
        endpointById={endpointById}
      />

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
