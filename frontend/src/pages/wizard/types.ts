import type { GonkaBalance, TelegramWizardUser } from '@/types'

export type Provider = 'openai' | 'gonka'
export type WalletMode = 'create' | 'import'
export type ModelRow = { name: string; role: 'chat' | 'summary' | 'vision' | '' }

export type StepId =
  | 'provider'
  | 'openai'
  | 'wallet-choice'
  | 'wallet-import'
  | 'wallet-create'
  | 'wallet-reveal'
  | 'wallet-balance'
  | 'gonka-models'
  | 'telegram'
  | 'finish'

export interface State {
  provider: Provider
  openaiBaseUrl: string
  openaiApiKey: string
  modelRows: ModelRow[]
  walletMode: WalletMode
  gonkaNodeUrl: string
  gonkaPrivateKey: string
  gonkaAddress: string
  gonkaMnemonicWords: string[]
  mnemonicAcknowledged: boolean
  bypassBalance: boolean
  gonkaBalance: GonkaBalance | null
  tgToken: string
  tgLinkedUser: TelegramWizardUser | null
  tgEnvUserIds: string
  tgSkip: boolean
}

export interface StepMeta {
  title: string
  subtitle: string
}
