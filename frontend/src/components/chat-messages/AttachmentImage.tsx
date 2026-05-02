import { useState } from 'react'
import { Download, Maximize2, X } from '@/lib/icons'
import type { Attachment } from '../../types'

export function AttachmentImage({ attachment, sessionId }: { attachment: Attachment; sessionId: string }) {
  const [expanded, setExpanded] = useState(false)
  const src = `/api/artifacts/${sessionId}/${attachment.id}`

  return (
    <>
      <div className="relative group cursor-pointer rounded-lg overflow-hidden" onClick={() => setExpanded(true)}>
        <img src={src} alt={attachment.fileName} className="w-full rounded-lg" loading="lazy" />
        <div className="absolute inset-0 bg-black/0 group-hover:bg-black/20 transition-colors flex items-center justify-center">
          <Maximize2 size={20} className="text-white opacity-0 group-hover:opacity-100 transition-opacity drop-shadow" />
        </div>
        <div className="absolute bottom-0 left-0 right-0 px-2 py-1 bg-gradient-to-t from-black/50 to-transparent">
          <span className="text-[10px] text-white/80 truncate block">{attachment.fileName}</span>
        </div>
      </div>

      {expanded && (
        <div className="fixed inset-0 z-[70] flex items-center justify-center bg-black/70" onClick={() => setExpanded(false)}>
          <div className="relative max-w-[90vw] max-h-[90vh]" onClick={e => e.stopPropagation()}>
            <img src={src} alt={attachment.fileName} className="max-w-full max-h-[90vh] rounded-lg shadow-2xl" />
            <button
              onClick={() => setExpanded(false)}
              className="absolute top-2 right-2 p-1.5 rounded-full bg-black/50 text-white hover:bg-black/70"
            >
              <X size={16} />
            </button>
            <div className="absolute bottom-2 left-2 right-2 flex items-center justify-between">
              <span className="text-xs text-white/80 bg-black/40 px-2 py-1 rounded">{attachment.fileName}</span>
              <a
                href={src}
                download={attachment.fileName}
                onClick={e => e.stopPropagation()}
                className="text-xs text-white/80 bg-black/40 px-2 py-1 rounded hover:bg-black/60"
              >
                <Download size={12} className="inline mr-1" />
                Download
              </a>
            </div>
          </div>
        </div>
      )}
    </>
  )
}
