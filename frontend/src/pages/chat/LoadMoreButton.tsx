interface Props {
  pageSize: number
  loading: boolean
  onClick: () => void
}

export function LoadMoreButton({ pageSize, loading, onClick }: Props) {
  return (
    <div className="flex justify-center mb-4">
      <button
        type="button"
        onClick={onClick}
        disabled={loading}
        className="font-mono text-[11px] lowercase tracking-tight px-2.5 py-1 rounded-sm
                   border border-zinc-200/80 dark:border-zinc-800/70 bg-white/60 dark:bg-zinc-900/60
                   text-zinc-500 dark:text-zinc-500 hover:text-teal-600 dark:hover:text-teal-400
                   hover:border-teal-500/40 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {loading ? 'loading…' : `↑ load previous ${pageSize}`}
      </button>
    </div>
  )
}
