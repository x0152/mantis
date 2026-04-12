export interface LlmConnection {
  id: string
  provider: string
  baseUrl: string
  apiKey: string
}

export interface Settings {
  id: string
  chatPresetId: string
  serverPresetId: string
  memoryEnabled: boolean
  userMemories: string[]
}

export interface Model {
  id: string
  connectionId: string
  name: string
  thinkingMode: '' | 'skip' | 'inline'
}

export interface Preset {
  id: string
  name: string
  chatModelId: string
  summaryModelId: string
  imageModelId: string
  fallbackModelId: string
  temperature: number | null
  systemPrompt: string
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
  presetId: string
  config: Record<string, unknown>
  memories: Memory[]
  profileIds: string[]
  memoryEnabled: boolean
}

export interface GuardCapabilities {
  pipes: boolean
  redirects: boolean
  cmdSubst: boolean
  background: boolean
  sudo: boolean
  codeExec: boolean
  download: boolean
  install: boolean
  writeFs: boolean
  networkOut: boolean
  cron: boolean
  unrestricted: boolean
}

export interface CommandRule {
  command: string
  allowedArgs?: string[]
  allowedSql?: string[]
}

export interface GuardProfile {
  id: string
  name: string
  description: string
  builtin: boolean
  capabilities: GuardCapabilities
  commands: CommandRule[]
}

export interface Channel {
  id: string
  type: string
  name: string
  token: string
  modelId: string
  presetId: string
  allowedUserIds: number[]
}

export interface ChatSession {
  id: string
  title: string
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
  modelId?: string
  modelName?: string
  presetId?: string
  presetName?: string
  modelRole?: string
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
  modelId?: string
  modelName?: string
  presetId?: string
  presetName?: string
  modelRole?: string
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
  modelId?: string
  modelName?: string
  presetId?: string
  presetName?: string
  modelRole?: string
  entries: LogEntry[]
  startedAt: string
  finishedAt?: string
}
