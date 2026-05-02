import type { StepId, StepMeta } from './types'
import { MIN_BALANCE_GNK } from './seeds'

export function stepMeta(stepId: StepId): StepMeta {
  switch (stepId) {
    case 'provider':
      return {
        title: 'How do you want to power Mantis?',
        subtitle: 'Pick where Mantis sends your messages. You can change this later.',
      }
    case 'openai':
      return {
        title: 'Connect your AI provider',
        subtitle: 'Paste your provider’s URL and key, then list the models you want to use.',
      }
    case 'wallet-choice':
      return {
        title: 'Set up your Gonka wallet',
        subtitle: 'Already have a wallet? Use it. Brand new? We’ll create one in a click.',
      }
    case 'wallet-import':
      return {
        title: 'Use your existing wallet',
        subtitle: 'Paste your private key. We’ll use it to sign requests for you.',
      }
    case 'wallet-create':
      return {
        title: 'Create a new wallet',
        subtitle: 'One click and you’re done. We’ll show your recovery words right after.',
      }
    case 'wallet-reveal':
      return {
        title: 'Save your 24 secret words',
        subtitle: 'Write them on paper or save them in a password manager. We won’t show them again.',
      }
    case 'wallet-balance':
      return {
        title: 'Top up your wallet',
        subtitle: `Send at least ${MIN_BALANCE_GNK} GNK to the address below. Continue once it arrives.`,
      }
    case 'gonka-models':
      return {
        title: 'Choose your AI models',
        subtitle: 'Pick one for chat. The others are optional.',
      }
    case 'telegram':
      return {
        title: 'Talk to Mantis on Telegram?',
        subtitle: 'Paste a bot token and send the code to the bot — we’ll detect you. Or tick the box below to stay on chat-only.',
      }
    case 'finish':
      return {
        title: 'Ready to go',
        subtitle: 'Here’s what we’ll save. Click finish to wrap up.',
      }
  }
}
