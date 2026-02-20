import { Sparkles, MessageSquareText } from 'lucide-react'
import { Markdown } from './Markdown'
import type { LogEntry } from '../types'

export function parseCommand(content: string): string {
  try {
    const step = JSON.parse(content)
    if (step.args) {
      const args = JSON.parse(step.args)
      if (args.command) return args.command
      if (args.task) return args.task
    }
    return step.label ?? step.tool ?? content
  } catch {
    return content
  }
}

export function EntryLine({ entry }: { entry: LogEntry }) {
  const time = new Date(entry.timestamp).toLocaleTimeString('en-US', { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' })

  if (entry.type === 'thought') {
    return (
      <div className="my-1.5 rounded-md bg-violet-500/[0.06] border border-violet-500/10 px-3 py-2.5">
        <div className="flex items-center gap-1.5 mb-1.5">
          <Sparkles size={11} className="text-violet-400" />
          <span className="text-[11px] font-medium text-violet-400/80">Agent</span>
          <span className="text-[11px] text-zinc-700 ml-auto select-none">{time}</span>
        </div>
        <div className="pl-[17px] text-zinc-300">
          <Markdown content={entry.content} />
        </div>
      </div>
    )
  }

  if (entry.type === 'command') {
    const cmd = parseCommand(entry.content)
    return (
      <div className="font-mono text-xs leading-relaxed flex gap-3 text-teal-400">
        <span className="text-zinc-600 shrink-0 select-none">{time}</span>
        <span className="shrink-0 select-none">$</span>
        <span>{cmd}</span>
      </div>
    )
  }

  if (entry.type === 'output') {
    return (
      <div className="font-mono text-xs leading-relaxed flex gap-3 text-zinc-400">
        <span className="text-zinc-600 shrink-0 select-none">{time}</span>
        <span className="shrink-0 w-3" />
        <pre className="whitespace-pre-wrap break-all m-0">{entry.content}</pre>
      </div>
    )
  }

  if (entry.type === 'error') {
    return (
      <div className="font-mono text-xs leading-relaxed flex gap-3 text-red-400">
        <span className="text-zinc-600 shrink-0 select-none">{time}</span>
        <span className="shrink-0 select-none text-red-500">x</span>
        <span className="whitespace-pre-wrap break-all">{entry.content}</span>
      </div>
    )
  }

  return (
    <div className="font-mono text-xs leading-relaxed flex gap-3 text-zinc-500">
      <span className="text-zinc-600 shrink-0 select-none">{time}</span>
      <span className="shrink-0 w-3" />
      <span className="whitespace-pre-wrap break-all">{entry.content}</span>
    </div>
  )
}

export function PromptBanner({ prompt }: { prompt: string }) {
  if (!prompt) return null
  return (
    <div className="mb-2 rounded-md bg-teal-500/[0.06] border border-teal-500/10 px-3 py-2.5">
      <div className="flex items-center gap-1.5 mb-1.5">
        <MessageSquareText size={11} className="text-teal-400" />
        <span className="text-[11px] font-medium text-teal-400/80">Prompt</span>
      </div>
      <p className="pl-[17px] text-sm text-zinc-300 leading-relaxed whitespace-pre-wrap">{prompt}</p>
    </div>
  )
}
