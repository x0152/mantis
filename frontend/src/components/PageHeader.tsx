import type { ReactNode } from "react"

interface PageHeaderProps {
  title: string
  kicker?: string
  action?: ReactNode
}

export function PageHeader({ title, kicker, action }: PageHeaderProps) {
  return (
    <div className="flex items-end justify-between mb-5 pb-3 border-b border-zinc-200/70 dark:border-zinc-800/60">
      <div className="flex flex-col gap-1">
        {kicker && (
          <span className="kicker">
            <span className="kicker-num">·</span>
            <span>{kicker}</span>
          </span>
        )}
        <h2 className="text-base font-medium tracking-tight text-zinc-900 dark:text-zinc-50">{title}</h2>
      </div>
      {action}
    </div>
  )
}
