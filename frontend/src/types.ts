export interface LlmConnection {
  id: string
  provider: string
  baseUrl: string
  apiKey: string
}

export interface Config {
  id: string
  data: Record<string, unknown>
}

export interface Model {
  id: string
  connectionId: string
  name: string
  thinkingMode: '' | 'skip' | 'inline'
}

export interface Memory {
  id: string
  content: string
  createdAt: string
}

export interface CronJob {
  id: string
  name: string
  schedule: string
  prompt: string
  enabled: boolean
}

export interface Connection {
  id: string
  type: string
  name: string
  description: string
  modelId: string
  config: Record<string, unknown>
  memories: Memory[]
}

export interface GuardRule {
  id: string
  name: string
  description: string
  pattern: string
  connectionId: string
  enabled: boolean
}

export interface Channel {
  id: string
  type: string
  name: string
  token: string
  modelId: string
  allowedUserIds: number[]
}

export interface ChatSession {
  id: string
  createdAt: string
}

export interface Step {
  id: string
  tool: string
  label: string
  icon: string
  args: string
  status: 'running' | 'completed' | 'error'
  result?: string
  logId?: string
  modelName?: string
  contentOffset?: number
  startedAt: string
  finishedAt?: string
}

export interface ChatMessage {
  id: string
  sessionId: string
  role: 'user' | 'assistant'
  content: string
  status: string
  source?: string
  modelName?: string
  steps?: Step[]
  createdAt: string
}

export interface LogEntry {
  type: string
  content: string
  timestamp: string
}

export interface SessionLog {
  id: string
  connectionId: string
  agentName: string
  prompt: string
  status: 'running' | 'finished'
  entries: LogEntry[]
  startedAt: string
  finishedAt?: string
}
