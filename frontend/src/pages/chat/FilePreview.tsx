import { FileText, Image as ImageIcon, X } from '@/lib/icons'
import type { PendingFile } from './types'

function formatSize(bytes: number): string {
  return bytes >= 1024 * 1024
    ? `${(bytes / (1024 * 1024)).toFixed(1)} mb`
    : `${Math.max(1, Math.round(bytes / 1024))} kb`
}

export function FilePreview({ file, onRemove }: { file: PendingFile; onRemove: () => void }) {
  const f = file.file
  const sizeLabel = formatSize(f.size)
  const isImage = !!file.previewUrl

  return (
    <div className="group relative flex items-center gap-1.5 pl-1 pr-1.5 py-1 rounded-sm border border-zinc-200/80 dark:border-zinc-800/70 bg-white dark:bg-zinc-900">
      {isImage ? (
        <img src={file.previewUrl} alt={f.name} className="w-7 h-7 object-cover rounded-sm" />
      ) : (
        <div className="w-7 h-7 rounded-sm bg-zinc-100 dark:bg-zinc-800 flex items-center justify-center text-zinc-500">
          {f.type.startsWith('image/') ? <ImageIcon size={12} /> : <FileText size={12} />}
        </div>
      )}
      <div className="leading-tight max-w-[160px]">
        <div className="truncate text-[11.5px] text-zinc-700 dark:text-zinc-300">{f.name}</div>
        <div className="font-mono text-[10px] lowercase tabular-nums text-zinc-500 dark:text-zinc-600">{sizeLabel}</div>
      </div>
      <button
        type="button"
        onClick={onRemove}
        className="ml-0.5 p-0.5 rounded-sm text-zinc-400 hover:text-rose-500 hover:bg-rose-500/10"
        title="Remove"
      >
        <X size={11} />
      </button>
    </div>
  )
}
