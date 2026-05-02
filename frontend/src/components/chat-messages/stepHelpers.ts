import { Terminal, Calculator, Download, Mic, Eye, Wand2, Play, GitBranch, Bell, Layers } from '@/lib/icons'
import type { LogEntry, Step } from '../../types'

export const STEP_ICONS: Record<string, typeof Terminal> = {
  terminal: Terminal,
  calculator: Calculator,
  download: Download,
  mic: Mic,
  eye: Eye,
  skill: Wand2,
  play: Play,
  'git-branch': GitBranch,
  bell: Bell,
  layers: Layers,
}

export function stepArgsSummary(step: Step): string {
  try {
    const parsed = JSON.parse(step.args)
    const keys = Object.keys(parsed).filter(k => k !== 'task' && k !== 'prompt')
    if (keys.length === 0) return ''
    const parts = keys.slice(0, 3).map(k => {
      const v = parsed[k]
      const s = typeof v === 'string' ? v : JSON.stringify(v)
      return s.length > 40 ? s.slice(0, 37) + '...' : s
    })
    return parts.join(', ')
  } catch {
    return ''
  }
}

export function planIdFromStepArgs(step: Step): string | undefined {
  if (!['plan_run', 'plan_create', 'plan_update', 'plan_get'].includes(step.tool)) return undefined
  try {
    const parsed = JSON.parse(step.result || '{}')
    return parsed.id || parsed.planId || parsed.plan_id
  } catch {
    return undefined
  }
}

export function extractStepPrompt(step: Step): string {
  try {
    const parsed = JSON.parse(step.args)
    return (parsed.task ?? parsed.prompt ?? '') as string
  } catch {
    return ''
  }
}

export function stepToEntries(step: Step): LogEntry[] {
  const entries: LogEntry[] = []
  const ts = step.startedAt || new Date().toISOString()
  const endTs = step.finishedAt || ts

  entries.push({ type: 'command', content: JSON.stringify(step), timestamp: ts })

  if (step.result) {
    entries.push({
      type: step.status === 'error' ? 'error' : 'output',
      content: step.result,
      timestamp: endTs,
    })
  }

  return entries
}
