import type { LucideIcon } from "@/lib/icons"

interface EmptyStateProps {
  icon: LucideIcon
  title: string
  description?: string
}

export function EmptyState({ icon: Icon, title, description }: EmptyStateProps) {
  return (
    <div className="text-center py-20">
      <Icon size={40} className="mx-auto text-zinc-400 dark:text-zinc-700 mb-3" />
      <p className="text-zinc-500 dark:text-zinc-400 text-sm font-medium">{title}</p>
      {description && (
        <p className="text-xs text-zinc-600 mt-1">{description}</p>
      )}
    </div>
  )
}
