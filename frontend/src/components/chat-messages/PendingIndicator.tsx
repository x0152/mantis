import { BrainThinking } from '../BrainThinking'
import { TypingLines } from '../TypingLines'

export function PendingIndicator({ mode = 'thinking' }: { mode?: 'thinking' | 'typing' }) {
  return (
    <div className="flex items-center gap-2 py-1 font-mono text-[12px] lowercase tracking-tight text-zinc-500 dark:text-zinc-400">
      {mode === 'thinking' ? (
        <BrainThinking size={26} active className="shrink-0 text-teal-500 dark:text-teal-400" />
      ) : (
        <TypingLines size={26} active className="shrink-0 text-teal-500 dark:text-teal-400" />
      )}
      {mode === 'typing' ? (
        <>
          <span>typing</span>
          <span className="typing-caret text-teal-500" aria-hidden />
        </>
      ) : (
        <span>thinking</span>
      )}
    </div>
  )
}
