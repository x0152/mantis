import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { FormField } from '@/components/FormField'
import { Loader2 } from '@/lib/icons'
import { isValidPrivateKey } from '../utils'

interface WalletImportStepProps {
  privateKey: string
  nodeUrl: string
  submitting: boolean
  onChangePrivateKey: (v: string) => void
  onChangeNodeUrl: (v: string) => void
  onConfirm: () => Promise<void>
}

export function WalletImportStep({
  privateKey,
  nodeUrl,
  submitting,
  onChangePrivateKey,
  onChangeNodeUrl,
  onConfirm,
}: WalletImportStepProps) {
  const [showServer, setShowServer] = useState(false)
  const canImport = isValidPrivateKey(privateKey)
  return (
    <div className="space-y-4">
      <FormField label="Private key" hint="64 letters and numbers from your wallet app. We accept it with or without “0x”.">
        <Input
          type="password"
          value={privateKey}
          onChange={e => onChangePrivateKey(e.target.value)}
          placeholder="0x..."
        />
      </FormField>

      <Button
        onClick={onConfirm}
        disabled={!canImport || !nodeUrl.trim() || submitting}
        className="w-full h-10"
      >
        {submitting ? (
          <>
            <Loader2 size={14} className="animate-spin" /> checking…
          </>
        ) : (
          'Use this wallet'
        )}
      </Button>

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
