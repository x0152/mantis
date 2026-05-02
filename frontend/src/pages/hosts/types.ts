export const STATE_STYLES: Record<string, string> = {
  running: 'bg-teal-500/15 text-teal-700 dark:text-teal-300 border-teal-500/30',
  exited: 'bg-zinc-500/15 text-zinc-600 dark:text-zinc-400 border-zinc-500/30',
  stopped: 'bg-zinc-500/15 text-zinc-600 dark:text-zinc-400 border-zinc-500/30',
  'not-built': 'bg-amber-500/15 text-amber-700 dark:text-amber-300 border-amber-500/30',
  created: 'bg-sky-500/15 text-sky-700 dark:text-sky-300 border-sky-500/30',
  paused: 'bg-violet-500/15 text-violet-700 dark:text-violet-300 border-violet-500/30',
  restarting: 'bg-amber-500/15 text-amber-700 dark:text-amber-300 border-amber-500/30',
  dead: 'bg-red-500/15 text-red-700 dark:text-red-300 border-red-500/30',
}

export const SELECT_CLASS =
  'flex w-full rounded-lg border border-zinc-300 dark:border-zinc-700 bg-white dark:bg-zinc-800 px-3 py-2 text-sm text-zinc-900 dark:text-zinc-100 focus:outline-none focus:border-teal-500/50'

export const EMPTY_SSH = { host: '', port: '22', username: '', password: '', privateKey: '' }

export type SshConfig = typeof EMPTY_SSH

export type ServerForm = {
  type: string
  name: string
  description: string
  presetId: string
  profileIds: string[]
  memoryEnabled: boolean
}

export const stateClass = (state: string): string =>
  STATE_STYLES[state] ?? 'bg-zinc-500/15 text-zinc-700 dark:text-zinc-300 border-zinc-500/30'

export function sandboxName(connName: string): string {
  return connName.startsWith('sb-') ? connName.slice(3) : connName
}

export function relativeTime(iso: string): string {
  if (!iso || iso.startsWith('0001')) return '—'
  const d = new Date(iso).getTime()
  if (!d) return '—'
  const diff = Date.now() - d
  const sec = Math.max(0, Math.floor(diff / 1000))
  if (sec < 60) return `${sec}s ago`
  const min = Math.floor(sec / 60)
  if (min < 60) return `${min}m ago`
  const hrs = Math.floor(min / 60)
  if (hrs < 48) return `${hrs}h ago`
  return `${Math.floor(hrs / 24)}d ago`
}

export function parseSshConfig(cfg: Record<string, unknown>): SshConfig {
  return {
    host: String(cfg.host ?? ''),
    port: String(cfg.port ?? '22'),
    username: String(cfg.username ?? ''),
    password: String(cfg.password ?? ''),
    privateKey: String(cfg.privateKey ?? ''),
  }
}
