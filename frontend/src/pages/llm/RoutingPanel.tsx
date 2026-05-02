import { Route as RouteIcon } from '@/lib/icons'
import type { Preset } from '@/types'
import { ROUTING_DEFS } from './types'

interface Props {
  routing: { chatPresetId: string; serverPresetId: string }
  profiles: Preset[]
  onChange: (key: 'chatPresetId' | 'serverPresetId', value: string) => void
}

export function RoutingPanel({ routing, profiles, onChange }: Props) {
  return (
    <div className="rounded-lg border border-teal-500/20 bg-teal-500/5 dark:bg-teal-500/[0.06] px-4 py-3 mb-3">
      <div className="flex items-center gap-1.5 mb-2">
        <RouteIcon size={12} className="text-teal-600 dark:text-teal-400" />
        <span className="text-[11px] font-semibold uppercase tracking-wider text-teal-700 dark:text-teal-400">
          Active routing
        </span>
      </div>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-4 gap-y-2.5">
        {ROUTING_DEFS.map(({ key, label, hint }) => (
          <div key={key} className="min-w-0">
            <div className="flex items-center gap-2">
              <label className="text-xs font-medium text-zinc-700 dark:text-zinc-300 shrink-0">{label}</label>
              <span className="text-zinc-400 shrink-0">→</span>
              <select
                value={routing[key]}
                onChange={e => onChange(key, e.target.value)}
                className="flex-1 min-w-0 rounded-md border border-zinc-300 dark:border-zinc-700 bg-white dark:bg-zinc-900 px-2 py-1 text-xs text-zinc-900 dark:text-zinc-100 focus:outline-none focus:border-teal-500/50"
              >
                <option value="">None</option>
                {profiles.map(p => (
                  <option key={p.id} value={p.id}>{p.name}</option>
                ))}
              </select>
            </div>
            <p className="text-[10.5px] text-zinc-500 dark:text-zinc-600 mt-1 ml-0.5">{hint}</p>
          </div>
        ))}
      </div>
    </div>
  )
}
