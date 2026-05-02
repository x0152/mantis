import { api } from '@/api'
import type { Provider, State } from './types'

export function normalizeBaseUrl(raw: string): string {
  const trimmed = raw.trim()
  if (!trimmed) return ''
  try {
    const u = new URL(trimmed)
    return `${u.protocol}//${u.host}${u.pathname}`.replace(/\/+$/, '').toLowerCase()
  } catch {
    return trimmed.replace(/\/+$/, '').toLowerCase()
  }
}

export function slugify(input: string): string {
  const s = input
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+/, '')
    .replace(/-+$/, '')
  return s || 'llm-primary'
}

export function suggestEndpointId(baseUrl: string, provider: Provider): string {
  if (provider === 'gonka') return 'gonka-primary'
  try {
    const u = new URL(baseUrl.trim())
    const hostSlug = slugify(`${u.hostname}${u.port ? `-${u.port}` : ''}`)
    if (hostSlug.includes('openai')) return 'openai-primary'
    if (hostSlug.includes('localhost') || hostSlug.startsWith('127-')) return 'local-llm'
    return `llm-${hostSlug}`.slice(0, 48)
  } catch {
    return 'llm-primary'
  }
}

export function makeUniqueID(base: string, existing: Set<string>): string {
  if (!existing.has(base)) return base
  for (let i = 2; i < 1000; i++) {
    const candidate = `${base}-${i}`.slice(0, 48)
    if (!existing.has(candidate)) return candidate
  }
  return `${base}-${Date.now()}`.slice(0, 48)
}

export async function pollConnections(names: string[], timeoutMs: number) {
  const deadline = Date.now() + timeoutMs
  while (Date.now() < deadline) {
    const list = await api.connections.list()
    const have = new Set(list.map(c => c.name))
    if (names.every(n => have.has(n))) return list
    await new Promise(r => setTimeout(r, 500))
  }
  return api.connections.list()
}

export function formatGnk(v: number): string {
  if (v >= 1000) return v.toFixed(0)
  if (v >= 1) return v.toFixed(2)
  if (v > 0) return v.toFixed(4)
  return '0'
}

export const isValidPrivateKey = (raw: string): boolean => {
  const trimmed = raw.trim().replace(/^0x/i, '')
  return /^[0-9a-fA-F]{64}$/.test(trimmed)
}

export function telegramSummary(state: State): string {
  if (state.tgSkip) return 'off — connect later via env'
  if (!state.tgToken.trim()) return 'off — connect later via env'
  if (state.tgLinkedUser) {
    const name = state.tgLinkedUser.name || state.tgLinkedUser.username || String(state.tgLinkedUser.id)
    return `linked · ${name}`
  }
  return 'token saved · pending link'
}
