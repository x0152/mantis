export function DropOverlay() {
  return (
    <div className="absolute inset-0 z-20 pointer-events-none flex items-center justify-center bg-teal-500/10 border border-dashed border-teal-500/60">
      <div className="bg-white/90 dark:bg-zinc-900/90 border border-teal-500/40 rounded-md px-4 py-2 font-mono text-[12px] lowercase tracking-tight text-teal-600 dark:text-teal-400 shadow-lg">
        drop files to attach
      </div>
    </div>
  )
}
