import type { Config, Model, Connection, CronJob, GuardRule, ChatSession, ChatMessage, SessionLog, LlmConnection, Channel } from './types'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`/api${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  if (!res.ok) {
    const body = await res.json().catch(() => null)
    throw new Error(body?.title ?? `${res.status} ${res.statusText}`)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

export const api = {
  config: {
    get: () => request<Config>('/config'),
    update: (data: unknown) =>
      request<Config>('/config', { method: 'PUT', body: JSON.stringify({ data }) }),
  },
  llmConnections: {
    list: () => request<LlmConnection[]>('/llm-connections'),
    get: (id: string) => request<LlmConnection>(`/llm-connections/${id}`),
    create: (data: { id: string; provider: string; baseUrl: string; apiKey: string }) =>
      request<LlmConnection>('/llm-connections', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: { provider: string; baseUrl: string; apiKey: string }) =>
      request<LlmConnection>(`/llm-connections/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id: string) => request<void>(`/llm-connections/${id}`, { method: 'DELETE' }),
  },
  models: {
    list: () => request<Model[]>('/models'),
    get: (id: string) => request<Model>(`/models/${id}`),
    create: (connectionId: string, name: string, thinkingMode: string) =>
      request<Model>('/models', { method: 'POST', body: JSON.stringify({ connectionId, name, thinkingMode }) }),
    update: (id: string, connectionId: string, name: string, thinkingMode: string) =>
      request<Model>(`/models/${id}`, { method: 'PUT', body: JSON.stringify({ connectionId, name, thinkingMode }) }),
    delete: (id: string) => request<void>(`/models/${id}`, { method: 'DELETE' }),
  },
  connections: {
    list: () => request<Connection[]>('/connections'),
    get: (id: string) => request<Connection>(`/connections/${id}`),
    create: (data: { type: string; name: string; description: string; modelId: string; config: unknown }) =>
      request<Connection>('/connections', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: { type: string; name: string; description: string; modelId: string; config: unknown }) =>
      request<Connection>(`/connections/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id: string) => request<void>(`/connections/${id}`, { method: 'DELETE' }),
    addMemory: (id: string, content: string) =>
      request<Connection>(`/connections/${id}/memories`, { method: 'POST', body: JSON.stringify({ content }) }),
    deleteMemory: (id: string, memoryId: string) =>
      request<void>(`/connections/${id}/memories/${memoryId}`, { method: 'DELETE' }),
  },
  cronJobs: {
    list: () => request<CronJob[]>('/cron-jobs'),
    get: (id: string) => request<CronJob>(`/cron-jobs/${id}`),
    create: (data: Omit<CronJob, 'id'>) =>
      request<CronJob>('/cron-jobs', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: Omit<CronJob, 'id'>) =>
      request<CronJob>(`/cron-jobs/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id: string) => request<void>(`/cron-jobs/${id}`, { method: 'DELETE' }),
  },
  guardRules: {
    list: () => request<GuardRule[]>('/guard-rules'),
    create: (data: Omit<GuardRule, 'id'>) =>
      request<GuardRule>('/guard-rules', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: Omit<GuardRule, 'id'>) =>
      request<GuardRule>(`/guard-rules/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id: string) => request<void>(`/guard-rules/${id}`, { method: 'DELETE' }),
  },
  channels: {
    list: () => request<Channel[]>('/channels'),
    get: (id: string) => request<Channel>(`/channels/${id}`),
    create: (data: Omit<Channel, 'id'>) =>
      request<Channel>('/channels', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: { name: string; token: string; modelId: string; allowedUserIds: number[] }) =>
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
    listMessages: (opts?: { limit?: number; offset?: number; sessionId?: string; source?: string }) => {
      const qs = new URLSearchParams()
      if (opts?.limit != null) qs.set('limit', String(opts.limit))
      if (opts?.offset != null) qs.set('offset', String(opts.offset))
      if (opts?.sessionId) qs.set('sessionId', opts.sessionId)
      if (opts?.source) qs.set('source', opts.source)
      const q = qs.toString()
      return request<ChatMessage[]>(`/chat/messages${q ? `?${q}` : ''}`)
    },
    sendMessage: (sessionId: string, content: string) =>
      request<{ userMessage: ChatMessage; assistantMessage: ChatMessage }>('/chat/messages', {
        method: 'POST', body: JSON.stringify({ sessionId, content }),
      }),
    clearHistory: () => request<void>('/chat/history', { method: 'DELETE' }),
  },
}
