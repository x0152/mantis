import type { GonkaConfig } from '@/types'
import type { StepId, State } from './types'

export function buildPath(state: State, gonkaConfig: GonkaConfig | null): StepId[] {
  if (state.provider === 'openai') {
    return ['provider', 'openai', 'telegram', 'finish']
  }
  const path: StepId[] = ['provider', 'wallet-choice']
  if (state.walletMode === 'create') {
    path.push('wallet-create')
    if (gonkaConfig?.inferencedAvailable) {
      path.push('wallet-reveal')
    }
  } else {
    path.push('wallet-import')
  }
  path.push('wallet-balance', 'gonka-models', 'telegram', 'finish')
  return path
}
