import { useState } from 'react'
import { ChevronRight } from '@/lib/icons'
import { MantisLogo } from './MantisLogo'

export function ThinkingBlock({ content, streaming }: { content: string; streaming: boolean }) {
  const [open, setOpen] = useState(false)
  const trimmed = content.trim()
  const chars = trimmed.length

  return (
    <div
      className={`not-prose my-2 last:mb-0 rounded-[5px] border border-teal-500/20 bg-teal-500/[0.04] overflow-hidden ${
        open ? 'w-full' : 'w-fit max-w-full'
      }`}
    >
      <button
        type="button"
        onClick={() => setOpen(v => !v)}
        className="w-full flex items-center gap-2 px-2.5 py-1 text-left font-mono text-[11px] lowercase tracking-tight text-teal-700 dark:text-teal-300 hover:bg-teal-500/10 transition-colors"
      >
        <MantisLogo
          size={12}
          state={streaming ? 'thinking' : 'idle'}
          className="shrink-0"
        />
        <span>{streaming ? 'thinking' : 'thoughts'}</span>
        <span className="text-[10px] tabular-nums text-teal-700/60 dark:text-teal-300/60">
          {chars}
        </span>
        <ChevronRight
          size={11}
          strokeWidth={1.6}
          className={`ml-1 shrink-0 text-teal-700/60 dark:text-teal-300/60 transition-transform ${open ? 'rotate-90' : ''}`}
        />
      </button>
      {open && trimmed && (
        <div className="px-3 py-2 border-t border-teal-500/20 text-[13px] leading-relaxed whitespace-pre-wrap break-words text-zinc-600 dark:text-zinc-400 max-h-64 overflow-auto">
          {trimmed}
        </div>
      )}
    </div>
  )
}
