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

export interface Skill {
  id: string
  connectionId: string
  name: string
  description: string
  parameters: Record<string, unknown>
  script: string
}

export interface PlanNodePosition {
  x: number
  y: number
}

export interface PlanNode {
  id: string
  type: 'action' | 'decision'
  label: string
  prompt: string
  position: PlanNodePosition
  clearContext?: boolean
  maxRetries?: number
}

export interface PlanEdge {
  id: string
  source: string
  target: string
  label: string
}

export interface PlanGraph {
  nodes: PlanNode[]
  edges: PlanEdge[]
}

export interface Plan {
  id: string
  name: string
  description: string
  schedule: string
  enabled: boolean
  parameters: Record<string, unknown>
  graph: PlanGraph
}

export type PlanRunStatus = 'running' | 'completed' | 'failed' | 'cancelled' | 'paused'
export type PlanStepStatus = 'pending' | 'running' | 'completed' | 'failed' | 'skipped'

export interface PlanStepRun {
  nodeId: string
  status: PlanStepStatus
  result?: string
  messageId?: string
  startedAt?: string
  finishedAt?: string
}

export interface PlanRun {
  id: string
  planId: string
  status: PlanRunStatus
  trigger: 'manual' | 'schedule'
  input: Record<string, unknown>
  steps: PlanStepRun[]
  startedAt: string
  finishedAt?: string
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
  source?: string
  active?: boolean
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

export interface Attachment {
  id: string
  fileName: string
  mimeType: string
  size: number
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
  attachments?: Attachment[]
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
