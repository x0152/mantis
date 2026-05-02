import { GitBranch } from '@/lib/icons'
import { navigate } from '../../router'

export function PlanBanner({ planId }: { planId?: string }) {
  return (
    <div className="px-6 py-3 border-t border-zinc-200/80 dark:border-zinc-800/60 bg-zinc-50/70 dark:bg-zinc-900/40 shrink-0">
      <div className="flex items-center gap-2 font-mono text-[11px] lowercase tracking-tight text-zinc-500 dark:text-zinc-500">
        <GitBranch size={12} className="text-amber-500/70" />
        <span>plan execution chat · read-only</span>
        {planId && (
          <button
            onClick={() => navigate({ page: 'plans', planId })}
            className="ml-auto text-amber-500/80 hover:text-amber-400 hover:underline"
          >
            open plan →
          </button>
        )}
      </div>
    </div>
  )
}
