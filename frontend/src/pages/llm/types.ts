import { MessageSquare, FileText, Eye, RotateCcw, type LucideIcon } from '@/lib/icons'

export type EndpointForm = { id: string; provider: string; baseUrl: string; apiKey: string }

export type ModelForm = {
  name: string
  connectionId: string
  thinkingMode: string
  contextWindow: string
  reserveTokens: string
  compactTokens: string
}

export type ProfileForm = {
  name: string
  chatModelId: string
  summaryModelId: string
  imageModelId: string
  fallbackModelId: string
  temperature: string
  systemPrompt: string
}

export type RoleKey = 'chat' | 'summary' | 'image' | 'fallback'

export type RoleDef = {
  key: RoleKey
  field: keyof Pick<ProfileForm, 'chatModelId' | 'summaryModelId' | 'imageModelId' | 'fallbackModelId'>
  icon: LucideIcon
  label: string
  hint: string
  required?: boolean
}

export const DEFAULT_CONTEXT_WINDOW = 128000
export const DEFAULT_RESERVE_TOKENS = 20000

export const EMPTY_PROFILE_FORM: ProfileForm = {
  name: '',
  chatModelId: '',
  summaryModelId: '',
  imageModelId: '',
  fallbackModelId: '',
  temperature: '',
  systemPrompt: '',
}

export const SELECT_CLASS =
  'flex w-full rounded-lg border border-zinc-300 dark:border-zinc-700 bg-white dark:bg-zinc-800 px-3 py-2 text-sm text-zinc-900 dark:text-zinc-100 focus:outline-none focus:border-teal-500/50'

export const ROLE_DEFS: RoleDef[] = [
  { key: 'chat',     field: 'chatModelId',     icon: MessageSquare, label: 'Chat',     hint: 'Main conversations and tool use', required: true },
  { key: 'summary',  field: 'summaryModelId',  icon: FileText,      label: 'Summary',  hint: 'Memory extraction and session titles' },
  { key: 'image',    field: 'imageModelId',    icon: Eye,           label: 'Vision',   hint: 'Used when a user sends an image' },
  { key: 'fallback', field: 'fallbackModelId', icon: RotateCcw,     label: 'Fallback', hint: 'Used when the chat model errors out' },
]

export const ROUTING_DEFS: { key: 'chatPresetId' | 'serverPresetId'; label: string; hint: string }[] = [
  { key: 'chatPresetId',   label: 'Chat & Telegram', hint: 'Web chat, Telegram channels, plan execution' },
  { key: 'serverPresetId', label: 'Servers & SSH',   hint: 'SSH connections and remote tool execution' },
]
