import { Bell, Plug, Sparkles, Wallet, type IconComponent } from '@/lib/icons'
import type { Provider } from '../types'

interface FinishStepProps {
  provider: Provider
  chatModel?: string
  endpoint: string
  telegramLabel: string
}

export function FinishStep({ provider, chatModel, endpoint, telegramLabel }: FinishStepProps) {
  const rows: Array<{ icon: IconComponent; label: string; value: string }> = [
    {
      icon: provider === 'openai' ? Plug : Wallet,
      label: 'Powered by',
      value: provider === 'openai' ? 'OpenAI-compatible API' : 'Gonka wallet',
    },
    { icon: Plug, label: 'Server', value: endpoint || '—' },
    { icon: Sparkles, label: 'Chat model', value: chatModel || '—' },
    { icon: Bell, label: 'Telegram', value: telegramLabel },
  ]
  return (
    <div className="space-y-2">
      {rows.map((row, i) => {
        const Icon = row.icon
        return (
          <div
            key={i}
            className="flex items-center gap-3 rounded-md border border-zinc-200/70 dark:border-zinc-800/60 bg-zinc-50 dark:bg-zinc-900 px-3 py-2"
          >
            <Icon size={14} className="text-zinc-500 dark:text-zinc-400 shrink-0" />
            <span className="text-[11px] uppercase tracking-wide text-zinc-500 dark:text-zinc-500 w-24">{row.label}</span>
            <span className="text-[12.5px] font-mono text-zinc-800 dark:text-zinc-200 truncate">{row.value}</span>
          </div>
        )
      })}
    </div>
  )
}
