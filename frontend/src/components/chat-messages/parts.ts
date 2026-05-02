import type { Step } from '../../types'

export type Part = { type: 'text'; text: string } | { type: 'step'; step: Step }
export type InterleavedPart = Part & { text?: string; step?: Step }

const encoder = new TextEncoder()
const decoder = new TextDecoder()

export function buildInterleavedParts(content: string, steps: Step[]): InterleavedPart[] {
  if (steps.length === 0) {
    return content ? [{ type: 'text', text: content }] : []
  }

  const hasOffsets = steps.some(s => (s.contentOffset ?? 0) > 0)
  if (!hasOffsets) {
    const parts: InterleavedPart[] = steps.map(s => ({ type: 'step' as const, step: s }))
    if (content) parts.push({ type: 'text', text: content })
    return parts
  }

  const sorted = [...steps].sort((a, b) => (a.contentOffset ?? 0) - (b.contentOffset ?? 0))
  const bytes = encoder.encode(content)
  const parts: InterleavedPart[] = []
  let pos = 0

  for (const step of sorted) {
    const offset = step.contentOffset ?? 0
    if (offset > pos) {
      const text = decoder.decode(bytes.slice(pos, offset)).trim()
      if (text) parts.push({ type: 'text', text })
    }
    parts.push({ type: 'step', step })
    pos = Math.max(pos, offset)
  }

  if (pos < bytes.length) {
    const text = decoder.decode(bytes.slice(pos)).trim()
    if (text) parts.push({ type: 'text', text })
  }

  return parts
}
