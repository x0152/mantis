import { Sparkles, Wallet } from '@/lib/icons'
import type { GonkaConfig } from '@/types'
import { BigChoiceCard } from '../components/BigChoiceCard'
import type { WalletMode } from '../types'

interface WalletChoiceStepProps {
  walletMode: WalletMode | null
  gonkaConfig: GonkaConfig | null
  onSelect: (mode: WalletMode) => void
}

export function WalletChoiceStep({ walletMode, gonkaConfig, onSelect }: WalletChoiceStepProps) {
  const inferencedReady = gonkaConfig?.inferencedAvailable ?? false
  return (
    <div className="space-y-3">
      <BigChoiceCard
        icon={Sparkles}
        title="I don’t have a wallet"
        description="We’ll create a brand-new Gonka wallet for you. Takes one click."
        bullets={[
          'Generates a 24-word recovery phrase',
          'You write it down once and you’re set',
          'No crypto experience needed',
        ]}
        selected={walletMode === 'create'}
        disabled={!inferencedReady}
        badge={!inferencedReady ? 'unavailable here' : undefined}
        badgeTone="amber"
        onClick={() => onSelect('create')}
      />
      <BigChoiceCard
        icon={Wallet}
        title="I already have a wallet"
        description="Paste a private key from your wallet app — Mantis will use it to sign requests."
        bullets={[
          '64-character private key (with or without 0x)',
          'Stays on your server, never leaves',
          'Switch wallets anytime in settings',
        ]}
        selected={walletMode === 'import'}
        onClick={() => onSelect('import')}
      />
    </div>
  )
}
