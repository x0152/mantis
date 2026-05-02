import type { ReactNode } from 'react'
import { CheckCircle2, type IconComponent } from '@/lib/icons'

export interface BigChoiceCardProps {
  icon: IconComponent
  title: string
  description: string
  selected?: boolean
  disabled?: boolean
  badge?: string
  badgeTone?: 'amber' | 'teal' | 'zinc'
  bullets?: ReactNode[]
  onClick: () => void
}

const badgeColors: Record<NonNullable<BigChoiceCardProps['badgeTone']>, string> = {
  amber: 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/20',
  teal: 'bg-teal-500/10 text-teal-600 dark:text-teal-400 border-teal-500/20',
  zinc: 'bg-zinc-500/10 text-zinc-600 dark:text-zinc-400 border-zinc-500/20',
}

export function BigChoiceCard({
  icon: Icon,
  title,
  description,
  selected,
  disabled,
  badge,
  badgeTone = 'amber',
  bullets,
  onClick,
}: BigChoiceCardProps) {
  return (
    <button
      type="button"
      disabled={disabled}
      onClick={onClick}
      className={`group relative w-full text-left rounded-xl border-2 px-5 py-4 transition-all ${
        disabled
          ? 'border-zinc-200/60 dark:border-zinc-800/40 opacity-50 cursor-not-allowed bg-transparent'
          : selected
            ? 'border-teal-500/70 bg-teal-500/5 shadow-sm'
            : 'border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900/40 hover:border-zinc-300 dark:hover:border-zinc-700 hover:bg-zinc-50/70 dark:hover:bg-zinc-900/60'
      }`}
    >
      {selected && (
        <div className="absolute right-3 top-3 text-teal-500 dark:text-teal-400">
          <CheckCircle2 size={18} weight="fill" />
        </div>
      )}
      <div className="flex items-start gap-3.5">
        <div
          className={`size-11 rounded-xl flex items-center justify-center shrink-0 transition-colors ${
            selected
              ? 'bg-teal-500/15 text-teal-600 dark:text-teal-400'
              : 'bg-zinc-100 dark:bg-zinc-800/80 text-zinc-600 dark:text-zinc-400 group-hover:text-zinc-900 dark:group-hover:text-zinc-200'
          }`}
        >
          <Icon size={22} />
        </div>
        <div className="flex-1 min-w-0 pr-6">
          <div className="flex items-center gap-2 flex-wrap">
            <span className="text-[14.5px] font-semibold tracking-tight text-zinc-900 dark:text-zinc-50">
              {title}
            </span>
            {badge && (
              <span
                className={`text-[10px] uppercase tracking-wide rounded-md border px-1.5 py-0.5 ${badgeColors[badgeTone]}`}
              >
                {badge}
              </span>
            )}
          </div>
          <p className="text-[12.5px] text-zinc-600 dark:text-zinc-400 mt-1 leading-relaxed">
            {description}
          </p>
          {bullets && bullets.length > 0 && (
            <ul className="mt-2.5 space-y-1">
              {bullets.map((b, i) => (
                <li
                  key={i}
                  className="flex items-start gap-1.5 text-[11.5px] text-zinc-500 dark:text-zinc-500 leading-relaxed"
                >
                  <span className="mt-1 size-1 shrink-0 rounded-full bg-zinc-400 dark:bg-zinc-600" />
                  <span>{b}</span>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>
    </button>
  )
}
