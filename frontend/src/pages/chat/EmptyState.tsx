import { SUGGESTIONS } from './suggestions'

interface EmptyStateProps {
  disabled?: boolean
  onInsert: (prompt: string) => void
  onSend: (prompt: string) => void
}

export function EmptyState({ disabled, onInsert, onSend }: EmptyStateProps) {
  if (disabled) {
    return (
      <div className="flex items-center justify-center h-full font-mono text-[11.5px] lowercase tracking-tight text-zinc-500 dark:text-zinc-600">
        — plan execution chat (read-only) —
      </div>
    )
  }
  return (
    <div className="flex flex-col items-center justify-center h-full px-4 py-10 max-w-2xl mx-auto w-full">
      <div className="self-stretch mb-3">
        <div className="kicker mb-1.5">
          <span className="kicker-num">00</span>
          <span className="kicker-sep">/</span>
          <span>what should we try?</span>
        </div>
        <p className="font-mono text-[11px] lowercase tracking-tight text-zinc-500 dark:text-zinc-600">
          a few ideas tailored to this setup · click to insert · double-click to send
        </p>
      </div>
      <div className="self-stretch grid grid-cols-1 sm:grid-cols-2 gap-1.5">
        {SUGGESTIONS.map(s => (
          <SuggestionCard
            key={s.title}
            icon={s.icon}
            title={s.title}
            prompt={s.prompt}
            disabled={disabled}
            onInsert={() => onInsert(s.prompt)}
            onSend={() => onSend(s.prompt)}
          />
        ))}
      </div>
    </div>
  )
}

interface SuggestionCardProps {
  icon: import('@/lib/icons').IconComponent
  title: string
  prompt: string
  disabled?: boolean
  onInsert: () => void
  onSend: () => void
}

function SuggestionCard({ icon: Icon, title, prompt, disabled, onInsert, onSend }: SuggestionCardProps) {
  return (
    <button
      type="button"
      onClick={onInsert}
      onDoubleClick={onSend}
      disabled={disabled}
      title="Click to insert into the prompt, double-click to send"
      className="group flex items-start gap-2.5 text-left px-3 py-2.5 min-w-0 rounded-md
                 border border-zinc-200/80 dark:border-zinc-800/70 bg-white/60 dark:bg-zinc-900/40
                 hover:border-teal-500/60 hover:bg-teal-500/5
                 text-zinc-700 dark:text-zinc-300
                 hover:text-zinc-900 dark:hover:text-zinc-100
                 transition-colors disabled:opacity-50 disabled:cursor-not-allowed select-none"
    >
      <div className="shrink-0 mt-0.5 text-zinc-500 dark:text-zinc-500 group-hover:text-teal-600 dark:group-hover:text-teal-400">
        <Icon size={14} strokeWidth={1.5} />
      </div>
      <div className="min-w-0">
        <div className="text-[13px] truncate">{title}</div>
        <div className="font-mono text-[10.5px] lowercase tracking-tight text-zinc-500 dark:text-zinc-600 mt-0.5 line-clamp-2 leading-snug">
          {prompt}
        </div>
      </div>
    </button>
  )
}
