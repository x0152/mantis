import type { StepId } from '../types'
import { stepMeta } from '../stepMeta'

export function StepHeader({ path, stepId }: { path: StepId[]; stepId: StepId }) {
  const meta = stepMeta(stepId)
  const idx = Math.max(0, path.indexOf(stepId))
  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        {path.map((_, i) => (
          <span
            key={i}
            className={`h-1 flex-1 rounded-full transition-colors ${
              i <= idx ? 'bg-teal-500/80' : 'bg-zinc-200 dark:bg-zinc-800'
            }`}
          />
        ))}
      </div>
      <div className="kicker">
        <span>step {idx + 1} of {path.length}</span>
      </div>
      <div>
        <h2 className="text-[17px] font-semibold tracking-tight text-zinc-900 dark:text-zinc-50">{meta.title}</h2>
        <p className="text-[12.5px] text-zinc-500 dark:text-zinc-400 mt-1 leading-relaxed">{meta.subtitle}</p>
      </div>
    </div>
  )
}
