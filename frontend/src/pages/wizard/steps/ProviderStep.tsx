import { Plug, Wallet } from '@/lib/icons'
import type { GonkaConfig } from '@/types'
import { BigChoiceCard } from '../components/BigChoiceCard'
import type { Provider } from '../types'

interface ProviderStepProps {
  provider: Provider | null
  gonkaConfig: GonkaConfig | null
  onSelect: (p: Provider) => void
}

export function ProviderStep({ provider, gonkaConfig, onSelect }: ProviderStepProps) {
  const gonkaUnavailable = gonkaConfig != null && !gonkaConfig.inferencedAvailable
  return (
    <div className="space-y-3">
      <BigChoiceCard
        icon={Plug}
        title="Use my own API key"
        description="Connect with an OpenAI key, OpenRouter, Together, Groq, or any other OpenAI-compatible provider."
        bullets={[
          'Bring an existing API key',
          'Pick any model your provider offers',
          'Pay your provider directly',
        ]}
        selected={provider === 'openai'}
        onClick={() => onSelect('openai')}
      />
      <BigChoiceCard
        icon={Wallet}
        title="Connect a Gonka wallet"
        description="Pay per request with crypto. No subscription, no monthly bill — just top up your wallet."
        bullets={[
          'Create a wallet in one click, or use your existing one',
          'Top up with GNK and start chatting',
          'Picks from a marketplace of providers automatically',
        ]}
        selected={provider === 'gonka'}
        disabled={gonkaUnavailable}
        badge={gonkaUnavailable ? 'unavailable here' : undefined}
        badgeTone="amber"
        onClick={() => onSelect('gonka')}
      />
    </div>
  )
}
