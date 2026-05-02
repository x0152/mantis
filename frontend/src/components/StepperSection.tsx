import { type ReactNode } from 'react'
import type { LucideIcon } from '@/lib/icons'

export type StepItem = {
  n: number
  label: string
  icon: LucideIcon
  count: number
  total?: number
  done: boolean
}

export function scrollToSection(n: number) {
  document.getElementById(`section-${n}`)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

export function Stepper({ steps }: { steps: StepItem[] }) {
  return (
    <nav className="rounded-lg border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 px-2 py-2 flex items-center gap-1">
      {steps.map((s, i) => {
        const Icon = s.icon
        return (
          <div key={s.n} className="flex items-center gap-1 flex-1 min-w-0">
            <button
              onClick={() => scrollToSection(s.n)}
              className="flex items-center gap-2 px-2 py-1.5 rounded-md hover:bg-zinc-100 dark:hover:bg-zinc-800/70 flex-1 min-w-0 text-left"
            >
              <span
                className={`w-5 h-5 rounded-full flex items-center justify-center text-[10px] font-bold shrink-0 ${
                  s.done
                    ? 'bg-teal-500 text-white'
                    : 'bg-zinc-200 dark:bg-zinc-800 text-zinc-500 dark:text-zinc-400'
                }`}
              >
                {s.n}
              </span>
              <Icon size={13} className={s.done ? 'text-teal-500' : 'text-zinc-400'} />
              <span className={`text-xs font-medium truncate ${s.done ? 'text-zinc-800 dark:text-zinc-200' : 'text-zinc-500'}`}>
                {s.label}
              </span>
              <span className="text-[11px] text-zinc-400 ml-auto shrink-0">
                {s.total != null ? `${s.count}/${s.total}` : s.count}
              </span>
            </button>
            {i < steps.length - 1 && <div className="w-3 h-px bg-zinc-200 dark:bg-zinc-800 shrink-0" />}
          </div>
        )
      })}
    </nav>
  )
}

export type SectionProps = {
  n: number
  icon: LucideIcon
  title: string
  subtitle: string
  action?: ReactNode
  disabled?: boolean
  disabledHint?: string
  children: ReactNode
}

export function Section({ n, icon: Icon, title, subtitle, action, disabled, disabledHint, children }: SectionProps) {
  return (
    <section
      id={`section-${n}`}
      className={`rounded-xl border border-zinc-200 dark:border-zinc-800 bg-zinc-50/50 dark:bg-zinc-900/40 p-4 scroll-mt-4 transition-opacity ${disabled ? 'opacity-60' : ''}`}
    >
      <header className="flex items-start justify-between gap-3 mb-3">
        <div className="flex items-start gap-3 min-w-0">
          <div className="w-6 h-6 rounded-full bg-teal-500/15 text-teal-600 dark:text-teal-400 flex items-center justify-center text-[11px] font-bold shrink-0 mt-0.5">
            {n}
          </div>
          <div className="min-w-0">
            <div className="flex items-center gap-1.5">
              <Icon size={14} className="text-teal-500" />
              <h2 className="text-sm font-semibold text-zinc-900 dark:text-zinc-100">{title}</h2>
            </div>
            <p className="text-[11px] text-zinc-500 dark:text-zinc-600 mt-0.5">
              {disabled && disabledHint ? disabledHint : subtitle}
            </p>
          </div>
        </div>
        {action && <div className="shrink-0">{action}</div>}
      </header>
      {!disabled && children}
    </section>
  )
}

export function EmptyHint({ children }: { children: ReactNode }) {
  return (
    <div className="rounded-lg border border-dashed border-zinc-300 dark:border-zinc-700 bg-white/60 dark:bg-zinc-900/40 px-4 py-6 text-center text-xs text-zinc-500 dark:text-zinc-500">
      {children}
    </div>
  )
}
