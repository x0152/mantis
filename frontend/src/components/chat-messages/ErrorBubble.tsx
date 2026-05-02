import { Loader2, RotateCcw } from '@/lib/icons'
import { Markdown } from '../Markdown'
import type { ChatMessage } from '../../types'
import { CopyMessageButton } from './CopyMessageButton'
import { stripThinking } from './utils'

interface ErrorBubbleProps {
  msg: ChatMessage
  canRegenerate?: boolean
  onRegenerate?: () => void
  regenerating?: boolean
}

export function ErrorBubble({ msg, canRegenerate, onRegenerate, regenerating }: ErrorBubbleProps) {
  const errorCopyText = stripThinking(msg.content)
  const showActions = !!errorCopyText || (canRegenerate && !!onRegenerate)
  return (
    <div className="flex justify-start">
      <div className="max-w-[80%] rounded-md text-[14.5px] bg-red-500/10 text-red-400 border border-red-500/20 border-l-2 border-l-red-500 px-4 py-2.5 space-y-2 leading-relaxed">
        <Markdown content={msg.content} />
        {showActions && (
          <div className="flex items-center gap-1">
            {errorCopyText && <CopyMessageButton text={errorCopyText} tone="danger" />}
            {canRegenerate && onRegenerate && (
              <button
                onClick={onRegenerate}
                disabled={regenerating}
                className="inline-flex items-center gap-1.5 px-2 py-1 text-[11px] font-medium rounded-md text-red-400 hover:text-red-300 hover:bg-red-500/10 disabled:opacity-50 transition-colors"
                title="Regenerate response"
              >
                {regenerating ? <Loader2 size={11} className="animate-spin" /> : <RotateCcw size={11} />}
                Regenerate
              </button>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
