import { useState } from 'react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Copy, Eye, ShieldAlert } from '@/lib/icons'
import { AddressDisplay } from '../components/AddressDisplay'

interface WalletRevealStepProps {
  address: string
  words: string[]
  acknowledged: boolean
  onAcknowledge: (v: boolean) => void
}

export function WalletRevealStep({ address, words, acknowledged, onAcknowledge }: WalletRevealStepProps) {
  const [revealed, setRevealed] = useState(false)
  const phrase = words.join(' ')
  const copy = async () => {
    try {
      await navigator.clipboard.writeText(phrase)
      toast.success('Recovery phrase copied to clipboard')
    } catch {
      toast.error('Copy failed — please write the phrase down manually')
    }
  }
  return (
    <div className="space-y-4">
      <div className="rounded-md border border-amber-500/40 bg-amber-500/5 px-3 py-2.5 text-[12px] text-amber-700 dark:text-amber-400 leading-relaxed flex items-start gap-2">
        <ShieldAlert size={16} className="mt-0.5 shrink-0" />
        <div>Anyone with these words can take your money. Save them now — we won’t show them again.</div>
      </div>

      <AddressDisplay label="Your wallet address" address={address} />

      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <Label className="mb-0">Your 24 secret words</Label>
          <div className="flex items-center gap-2">
            <Button variant="ghost" size="sm" onClick={() => setRevealed(v => !v)}>
              <Eye size={12} /> {revealed ? 'hide' : 'show'}
            </Button>
            <Button variant="secondary" size="sm" onClick={copy}>
              <Copy size={12} /> copy
            </Button>
          </div>
        </div>
        <div
          className={`grid grid-cols-3 gap-1.5 rounded-md border border-zinc-200/80 dark:border-zinc-800/70 bg-zinc-50 dark:bg-zinc-900 p-3 ${
            revealed ? '' : 'select-none'
          }`}
        >
          {words.map((word, i) => (
            <div
              key={i}
              className="flex items-center gap-1.5 rounded border border-zinc-200/60 dark:border-zinc-800/60 bg-white dark:bg-zinc-950 px-2 py-1"
            >
              <span className="font-mono text-[10px] tabular-nums text-zinc-400 dark:text-zinc-600 w-5">{i + 1}.</span>
              <span
                className={`text-[12px] font-mono ${
                  revealed
                    ? 'text-zinc-900 dark:text-zinc-50'
                    : 'tracking-widest text-zinc-300 dark:text-zinc-700 blur-[5px]'
                }`}
              >
                {revealed ? word : '••••••'}
              </span>
            </div>
          ))}
        </div>
      </div>

      <label className="flex items-center gap-2 text-[12.5px] text-zinc-700 dark:text-zinc-300 cursor-pointer">
        <input
          type="checkbox"
          checked={acknowledged}
          onChange={e => onAcknowledge(e.target.checked)}
          className="size-4 accent-teal-600"
        />
        <span>I saved my 24 words. I get that I can’t see them again.</span>
      </label>
    </div>
  )
}
