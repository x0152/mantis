import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { FormField } from '@/components/FormField'
import { Loader2, Sparkles } from '@/lib/icons'
import type { GonkaConfig } from '@/types'

interface WalletCreateStepProps {
  gonkaConfig: GonkaConfig | null
  nodeUrl: string
  submitting: boolean
  onChangeNodeUrl: (v: string) => void
  onCreate: () => Promise<void>
}

export function WalletCreateStep({
  gonkaConfig,
  nodeUrl,
  submitting,
  onChangeNodeUrl,
  onCreate,
}: WalletCreateStepProps) {
  const [showServer, setShowServer] = useState(false)
  const inferencedReady = gonkaConfig?.inferencedAvailable ?? false
  return (
    <div className="space-y-4">
      <div className="rounded-xl border border-teal-500/20 bg-teal-500/5 px-5 py-4 flex items-start gap-3">
        <div className="size-9 rounded-lg bg-teal-500/15 text-teal-600 dark:text-teal-400 flex items-center justify-center shrink-0">
          <Sparkles size={18} />
        </div>
        <div className="text-[12.5px] text-zinc-700 dark:text-zinc-300 leading-relaxed">
          <p className="font-medium text-zinc-900 dark:text-zinc-50">Here’s what happens next</p>
          <ol className="mt-1.5 space-y-0.5 list-decimal list-inside text-zinc-600 dark:text-zinc-400">
            <li>We generate a fresh wallet on your server.</li>
            <li>You see your 24 secret words — save them somewhere safe.</li>
            <li>You top up the wallet with a small amount of GNK.</li>
          </ol>
        </div>
      </div>

      <Button
        onClick={onCreate}
        disabled={!inferencedReady || !nodeUrl.trim() || submitting}
        className="w-full h-10"
      >
        {submitting ? (
          <>
            <Loader2 size={14} className="animate-spin" /> creating…
          </>
        ) : (
          'Create my wallet'
        )}
      </Button>

      {!inferencedReady && (
        <p className="text-[11.5px] text-amber-600 dark:text-amber-400 text-center">
          Wallet creation isn’t available on this server. Use the “I already have a wallet” option instead.
        </p>
      )}

      <div className="pt-1">
        {!showServer ? (
          <button
            type="button"
            onClick={() => setShowServer(true)}
            className="text-[11.5px] text-zinc-500 dark:text-zinc-500 hover:text-teal-600 dark:hover:text-teal-400"
          >
            Server: <span className="font-mono">{nodeUrl || 'default'}</span> · change
          </button>
        ) : (
          <FormField label="Gonka server" hint="Where Mantis sends your AI requests. The default works.">
            <Input
              value={nodeUrl}
              onChange={e => onChangeNodeUrl(e.target.value)}
              placeholder="http://node1.gonka.ai:8000"
            />
          </FormField>
        )}
      </div>
    </div>
  )
}
