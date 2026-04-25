import type { Settings, Model, Preset, Connection, Skill, Plan, PlanRun, GuardProfile, ChatSession, ChatMessage, SessionLog, LlmConnection, ProviderModel, InferenceLimit, Channel, User, ContextStatus } from './types'

export class UnauthorizedError extends Error {
  constructor(message = 'Unauthorized') {
    super(message)
    this.name = 'UnauthorizedError'
  }
}

type UnauthorizedHandler = () => void
let unauthorizedHandler: UnauthorizedHandler | null = null

export function setUnauthorizedHandler(handler: UnauthorizedHandler | null) {
  unauthorizedHandler = handler
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`/api${path}`, {
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  if (res.status === 401) {
    unauthorizedHandler?.()
    throw new UnauthorizedError()
  }
  if (!res.ok) {
    const body = await res.json().catch(() => null)
    throw new Error(body?.title ?? `${res.status} ${res.statusText}`)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

export const api = {
  auth: {
    me: () => request<User>('/auth/me'),
    login: (token: string) =>
      request<User>('/auth/login', { method: 'POST', body: JSON.stringify({ token }) }),
    logout: () => request<void>('/auth/logout', { method: 'POST' }),
  },
  settings: {
    get: () => request<Settings>('/settings'),
    update: (data: Omit<Settings, 'id'>) =>
      request<Settings>('/settings', { method: 'PUT', body: JSON.stringify(data) }),
  },
  llmConnections: {
    list: () => request<LlmConnection[]>('/llm-connections'),
    get: (id: string) => request<LlmConnection>(`/llm-connections/${id}`),
    create: (data: { id: string; provider: string; baseUrl: string; apiKey: string }) =>
      request<LlmConnection>('/llm-connections', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: { provider: string; baseUrl: string; apiKey: string }) =>
      request<LlmConnection>(`/llm-connections/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    listAvailableModels: (id: string) =>
      request<ProviderModel[]>(`/llm-connections/${id}/available-models`),
    getInferenceLimit: (id: string) =>
      request<InferenceLimit>(`/llm-connections/${id}/inference-limit`),
    delete: (id: string) => request<void>(`/llm-connections/${id}`, { method: 'DELETE' }),
  },
  models: {
    list: () => request<Model[]>('/models'),
    get: (id: string) => request<Model>(`/models/${id}`),
    create: (data: Omit<Model, 'id'>) =>
      request<Model>('/models', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: Omit<Model, 'id'>) =>
      request<Model>(`/models/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id: string) => request<void>(`/models/${id}`, { method: 'DELETE' }),
  },
  presets: {
    list: () => request<Preset[]>('/presets'),
    create: (data: Omit<Preset, 'id'>) =>
      request<Preset>('/presets', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: Omit<Preset, 'id'>) =>
      request<Preset>(`/presets/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id: string) => request<void>(`/presets/${id}`, { method: 'DELETE' }),
  },
  connections: {
    list: () => request<Connection[]>('/connections'),
    get: (id: string) => request<Connection>(`/connections/${id}`),
    create: (data: { type: string; name: string; description: string; modelId?: string; presetId?: string; config: unknown; profileIds?: string[]; memoryEnabled?: boolean }) =>
      request<Connection>('/connections', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: { type: string; name: string; description: string; modelId?: string; presetId?: string; config: unknown; profileIds?: string[]; memoryEnabled?: boolean }) =>
      request<Connection>(`/connections/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id: string) => request<void>(`/connections/${id}`, { method: 'DELETE' }),
    addMemory: (id: string, content: string) =>
      request<Connection>(`/connections/${id}/memories`, { method: 'POST', body: JSON.stringify({ content }) }),
    deleteMemory: (id: string, memoryId: string) =>
      request<void>(`/connections/${id}/memories/${memoryId}`, { method: 'DELETE' }),
  },
  skills: {
    list: (opts?: { connectionId?: string }) => {
      const qs = new URLSearchParams()
      if (opts?.connectionId) qs.set('connectionId', opts.connectionId)
      const q = qs.toString()
      return request<Skill[]>(`/skills${q ? `?${q}` : ''}`)
    },
    create: (data: Omit<Skill, 'id'>) =>
      request<Skill>('/skills', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: Omit<Skill, 'id'>) =>
      request<Skill>(`/skills/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id: string) => request<void>(`/skills/${id}`, { method: 'DELETE' }),
  },
  plans: {
    list: () => request<Plan[]>('/plans'),
    create: (data: Omit<Plan, 'id'>) =>
      request<Plan>('/plans', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: Omit<Plan, 'id'>) =>
      request<Plan>(`/plans/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id: string) => request<void>(`/plans/${id}`, { method: 'DELETE' }),
  },
  planRuns: {
    list: (planId: string) => request<PlanRun[]>(`/plans/${planId}/runs`),
    get: (id: string) => request<PlanRun>(`/plan-runs/${id}`),
    trigger: (planId: string, input?: Record<string, unknown>) =>
      request<PlanRun>(`/plans/${planId}/runs`, { method: 'POST', body: JSON.stringify({ input: input ?? {} }) }),
    cancel: (id: string) => request<PlanRun>(`/plan-runs/${id}/cancel`, { method: 'POST' }),
  },
  guardProfiles: {
    list: () => request<GuardProfile[]>('/guard-profiles'),
    create: (data: Omit<GuardProfile, 'id' | 'builtin'>) =>
      request<GuardProfile>('/guard-profiles', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: Omit<GuardProfile, 'id' | 'builtin'>) =>
      request<GuardProfile>(`/guard-profiles/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id: string) => request<void>(`/guard-profiles/${id}`, { method: 'DELETE' }),
  },
  channels: {
    list: () => request<Channel[]>('/channels'),
    get: (id: string) => request<Channel>(`/channels/${id}`),
    create: (data: Omit<Channel, 'id'>) =>
      request<Channel>('/channels', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: { name: string; token: string; modelId?: string; presetId?: string; allowedUserIds: number[] }) =>
      request<Channel>(`/channels/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id: string) => request<void>(`/channels/${id}`, { method: 'DELETE' }),
  },
  sessionLogs: {
    list: (opts?: { connectionId?: string; limit?: number; offset?: number }) => {
      const qs = new URLSearchParams()
      if (opts?.connectionId) qs.set('connectionId', opts.connectionId)
      if (opts?.limit != null) qs.set('limit', String(opts.limit))
      if (opts?.offset != null) qs.set('offset', String(opts.offset))
      const q = qs.toString()
      return request<SessionLog[]>(`/session-logs${q ? `?${q}` : ''}`)
    },
    get: (id: string) => request<SessionLog>(`/session-logs/${id}`),
    clear: () => request<void>('/session-logs', { method: 'DELETE' }),
  },
  chat: {
    getSession: () => request<ChatSession>('/chat/session'),
    resetContext: () => request<ChatSession>('/chat/reset', { method: 'POST' }),
    listSessions: (opts?: { limit?: number; offset?: number }) => {
      const qs = new URLSearchParams()
      if (opts?.limit != null) qs.set('limit', String(opts.limit))
      if (opts?.offset != null) qs.set('offset', String(opts.offset))
      const q = qs.toString()
      return request<ChatSession[]>(`/chat/sessions${q ? `?${q}` : ''}`)
    },
    createSession: (title?: string) => request<ChatSession>('/chat/sessions', { method: 'POST', body: JSON.stringify({ title: title ?? '' }) }),
    updateSession: (id: string, title: string) => request<ChatSession>(`/chat/sessions/${id}`, { method: 'PUT', body: JSON.stringify({ title }) }),
    deleteSession: (id: string) =>request<void>(`/chat/sessions/${id}`, { method: 'DELETE' }),
    listMessages: (opts?: { limit?: number; offset?: number; sessionId?: string; source?: string }) => {
      const qs = new URLSearchParams()
      if (opts?.limit != null) qs.set('limit', String(opts.limit))
      if (opts?.offset != null) qs.set('offset', String(opts.offset))
      if (opts?.sessionId) qs.set('sessionId', opts.sessionId)
      if (opts?.source) qs.set('source', opts.source)
      const q = qs.toString()
      return request<ChatMessage[]>(`/chat/messages${q ? `?${q}` : ''}`)
    },
    sendMessage: (
      sessionId: string,
      content: string,
      files?: { fileName: string; mimeType?: string; dataBase64: string; caption?: string }[],
    ) =>
      request<{ userMessage: ChatMessage; assistantMessage: ChatMessage }>('/chat/messages', {
        method: 'POST',
        body: JSON.stringify({ sessionId, content, files: files ?? [] }),
      }),
    regenerate: (sessionId: string) =>
      request<{ assistantMessage: ChatMessage }>(`/chat/sessions/${sessionId}/regenerate`, {
        method: 'POST',
      }),
    stopSession: (sessionId: string) =>
      request<{ stopped: boolean }>(`/chat/sessions/${sessionId}/stop`, { method: 'POST' }),
    getContextStatus: (sessionId: string) =>
      request<ContextStatus>(`/chat/sessions/${sessionId}/context`),
    clearHistory: () => request<void>('/chat/history', { method: 'DELETE' }),
  },
}
