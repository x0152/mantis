import { forwardRef, useEffect, useRef } from 'react'
import { Paperclip, Send, Square } from '@/lib/icons'
import { Textarea } from '@/components/ui/textarea'
import { ContextMeter } from '../../components/ContextMeter'
import type { ChatMessage, ContextStatus } from '../../types'
import { FilePreview } from './FilePreview'
import type { PendingFile } from './types'

interface ComposerProps {
  input: string
  onChangeInput: (v: string) => void
  onSend: () => void
  onStop: () => void
  sending: boolean
  hasPending: boolean
  stopping: boolean
  files: PendingFile[]
  onAddFiles: (files: FileList | File[]) => void
  onRemoveFile: (id: string) => void
  attachError: string | null
  messages: ChatMessage[]
  partial: boolean
  contextStatus: ContextStatus | null
}

export const Composer = forwardRef<HTMLTextAreaElement, ComposerProps>(function Composer(
  {
    input,
    onChangeInput,
    onSend,
    onStop,
    sending,
    hasPending,
    stopping,
    files,
    onAddFiles,
    onRemoveFile,
    attachError,
    messages,
    partial,
    contextStatus,
  },
  textareaRef,
) {
  const fileInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    const el = (textareaRef as React.RefObject<HTMLTextAreaElement>).current
    if (!el) return
    el.style.height = 'auto'
    const max = 220
    el.style.height = `${Math.min(el.scrollHeight, max)}px`
  }, [input, textareaRef])

  return (
    <div className="px-6 py-3 border-t border-zinc-200/80 dark:border-zinc-800/60 bg-zinc-50/70 dark:bg-zinc-900/40 shrink-0 space-y-2">
      {messages.length > 0 && (
        <ContextMeter
          messages={messages}
          partial={partial}
          threshold={contextStatus?.compactThreshold}
          serverTokens={contextStatus?.lastContextTokens}
          compactionCount={contextStatus?.summaryVersion ?? 0}
        />
      )}
      {files.length > 0 && (
        <div className="flex flex-wrap gap-1.5">
          {files.map(pf => (
            <FilePreview key={pf.id} file={pf} onRemove={() => onRemoveFile(pf.id)} />
          ))}
        </div>
      )}
      {attachError && (
        <div className="font-mono text-[11px] lowercase text-rose-500 dark:text-rose-400">{attachError}</div>
      )}
      <div className="flex gap-1.5 items-end">
        <input
          ref={fileInputRef}
          type="file"
          multiple
          className="hidden"
          onChange={e => {
            if (e.target.files) onAddFiles(e.target.files)
            e.target.value = ''
          }}
        />
        <AttachButton
          onClick={() => fileInputRef.current?.click()}
          disabled={sending || hasPending}
        />
        <Textarea
          ref={textareaRef}
          value={input}
          onChange={e => onChangeInput(e.target.value)}
          onKeyDown={e => {
            if (e.key === 'Enter' && !e.shiftKey && !e.nativeEvent.isComposing) {
              e.preventDefault()
              onSend()
            }
          }}
          placeholder="type a message — shift+enter for newline"
          rows={1}
          className="flex-1 min-h-[36px] max-h-[220px] py-2 px-3 text-[14px] bg-white dark:bg-zinc-900
                     border border-zinc-200/80 dark:border-zinc-800/70 rounded-md
                     leading-relaxed overflow-y-auto placeholder:text-zinc-400 dark:placeholder:text-zinc-600
                     focus-visible:border-teal-500/50 focus-visible:ring-0"
          disabled={sending || hasPending}
        />
        {hasPending ? (
          <StopButton onClick={onStop} disabled={stopping} />
        ) : (
          <SendButton onClick={onSend} disabled={sending || (!input.trim() && files.length === 0)} />
        )}
      </div>
    </div>
  )
})

function AttachButton({ onClick, disabled }: { onClick: () => void; disabled: boolean }) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      title="Attach files"
      className="shrink-0 h-9 w-9 inline-flex items-center justify-center rounded-md
                 border border-zinc-200/80 dark:border-zinc-800/70 bg-white dark:bg-zinc-900
                 text-zinc-500 dark:text-zinc-500 hover:text-teal-600 dark:hover:text-teal-400
                 hover:border-teal-500/40 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
    >
      <Paperclip size={14} strokeWidth={1.5} />
    </button>
  )
}

function StopButton({ onClick, disabled }: { onClick: () => void; disabled: boolean }) {
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      title="Stop generation"
      className="shrink-0 h-9 w-9 inline-flex items-center justify-center rounded-md
                 border border-zinc-200/80 dark:border-zinc-800/70 bg-white dark:bg-zinc-900
                 text-zinc-500 dark:text-zinc-500 hover:text-rose-600 dark:hover:text-rose-400
                 hover:border-rose-500/40 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
    >
      <Square size={12} className="fill-current" />
    </button>
  )
}

function SendButton({ onClick, disabled }: { onClick: () => void; disabled: boolean }) {
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      title="Send"
      className="shrink-0 h-9 w-9 inline-flex items-center justify-center rounded-md
                 border border-teal-500/40 bg-teal-500/10 dark:bg-teal-500/15
                 text-teal-600 dark:text-teal-400 hover:bg-teal-500/20 dark:hover:bg-teal-500/25
                 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
    >
      <Send size={14} strokeWidth={1.6} />
    </button>
  )
}
