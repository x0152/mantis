import { useEffect, useRef } from 'react'
import { Download, Loader2, RotateCcw, Square } from '@/lib/icons'
import { Markdown } from '../Markdown'
import type { Attachment, ChatMessage, Step } from '../../types'
import { AttachmentImage } from './AttachmentImage'
import { CopyMessageButton } from './CopyMessageButton'
import { PendingIndicator } from './PendingIndicator'
import { StepBadge } from './StepBadge'
import { useTicker } from './useTicker'
import { buildInterleavedParts } from './parts'
import { estimateTokens, fmtElapsed, fmtTokens, formatBytes, stripThinking } from './utils'

interface AssistantBubbleProps {
  msg: ChatMessage
  onStepClick: (s: Step) => void
  canRegenerate?: boolean
  onRegenerate?: () => void
  regenerating?: boolean
}

export function AssistantBubble({ msg, onStepClick, canRegenerate, onRegenerate, regenerating }: AssistantBubbleProps) {
  const steps = msg.steps ?? []
  const isPending = msg.status === 'pending'
  const isCancelled = msg.status === 'cancelled'
  const hasRunningStep = steps.some(s => s.status === 'running')
  const parts = buildInterleavedParts(msg.content, steps)
  useTicker(isPending)

  const contentLen = msg.content?.length ?? 0
  const lastContentChangeRef = useRef<number>(Date.now())
  const prevContentLenRef = useRef<number>(contentLen)
  useEffect(() => {
    if (contentLen !== prevContentLenRef.current) {
      prevContentLenRef.current = contentLen
      lastContentChangeRef.current = Date.now()
    }
  }, [contentLen])

  const sinceContentChange = Date.now() - lastContentChangeRef.current
  const lastOpen = msg.content ? msg.content.lastIndexOf('<think>') : -1
  const lastClose = msg.content ? msg.content.lastIndexOf('</think>') : -1
  const inThinkTag = isPending && lastOpen >= 0 && lastOpen > lastClose
  const isTyping = isPending && contentLen > 0 && sinceContentChange < 900 && !inThinkTag
  const showThinking = isPending && !hasRunningStep && !isTyping && !inThinkTag
  const showTyping = isTyping && !hasRunningStep

  const msgStart = new Date(msg.createdAt).getTime()
  const msgEnd = msg.finishedAt ? new Date(msg.finishedAt).getTime() : Date.now()
  const msgElapsed = msgEnd - msgStart
  const showMsgDuration = msgElapsed >= 300 && (isPending || !!msg.finishedAt)

  const isImage = (a: Attachment) =>
    a.mimeType.startsWith('image/') || /\.(png|jpe?g|gif|webp|svg|bmp)$/i.test(a.fileName)
  const images = (msg.attachments ?? []).filter(isImage)
  const otherFiles = (msg.attachments ?? []).filter(a => !isImage(a))

  const estimatedTokens = msg.tokens ?? estimateTokens(msg.content)
  const showTokens = estimatedTokens > 0
  const showHeader = !!(msg.modelName || msg.presetName || msg.modelRole === 'fallback' || showMsgDuration || showTokens)
  const hasText = parts.some(p => p.type === 'text')
  const hasAnyContent =
    parts.length > 0 ||
    images.length > 0 ||
    otherFiles.length > 0 ||
    showThinking ||
    showTyping ||
    isCancelled
  const showEmpty = !hasAnyContent && !isPending

  return (
    <div className="flex justify-start">
      <div className="max-w-[80%] rounded-md text-[14.5px] bg-zinc-50/80 dark:bg-zinc-900/60 text-zinc-700 dark:text-zinc-300 border border-zinc-200/80 dark:border-zinc-800/60 border-l-2 border-l-zinc-300 dark:border-l-zinc-700 overflow-hidden px-4 py-2.5 space-y-2 leading-relaxed">
        {showHeader && (
          <Header
            presetName={msg.presetName}
            modelName={msg.modelName}
            modelRole={msg.modelRole}
            showMsgDuration={showMsgDuration}
            msgElapsed={msgElapsed}
            showTokens={showTokens}
            estimatedTokens={estimatedTokens}
          />
        )}
        {parts.map((part, i) => {
          if (part.type === 'step') {
            return (
              <div key={part.step!.id}>
                <StepBadge step={part.step!} onClick={() => onStepClick(part.step!)} />
              </div>
            )
          }
          return (
            <div key={`text-${i}`} className={hasText ? 'py-0.5' : ''}>
              <Markdown content={part.text!} sessionId={msg.sessionId} />
            </div>
          )
        })}
        {images.length > 0 && (
          <div className={`grid gap-1.5 ${images.length === 1 ? 'grid-cols-1' : 'grid-cols-2'} -mx-2`}>
            {images.map(a => (
              <AttachmentImage key={a.id} attachment={a} sessionId={msg.sessionId} />
            ))}
          </div>
        )}
        {otherFiles.length > 0 && (
          <div className="space-y-1">
            {otherFiles.map(a => (
              <a
                key={a.id}
                href={`/api/artifacts/${msg.sessionId}/${a.id}`}
                download={a.fileName}
                className="flex items-center gap-1.5 px-2 py-1 rounded-md text-xs bg-zinc-200 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-400 hover:text-zinc-800 dark:hover:text-zinc-200"
              >
                <Download size={11} />
                <span className="truncate">{a.fileName}</span>
                <span className="text-[10px] text-zinc-500 shrink-0">{formatBytes(a.size)}</span>
              </a>
            ))}
          </div>
        )}
        {showThinking && <PendingIndicator mode="thinking" />}
        {showTyping && <PendingIndicator mode="typing" />}
        {isCancelled && (
          <div>
            <span className="inline-flex items-center gap-1 px-1.5 py-0.5 font-mono text-[10px] lowercase rounded-sm bg-zinc-200/70 dark:bg-zinc-800/70 text-zinc-500 dark:text-zinc-500">
              <Square size={9} className="fill-current" />
              stopped
            </span>
          </div>
        )}
        {showEmpty && (
          <div className="font-mono text-[11px] lowercase text-zinc-500 dark:text-zinc-600">
            — no response —
          </div>
        )}
        {!isPending && (
          <Footer
            content={msg.content}
            canRegenerate={canRegenerate}
            onRegenerate={onRegenerate}
            regenerating={regenerating}
          />
        )}
      </div>
    </div>
  )
}

interface HeaderProps {
  presetName?: string
  modelName?: string
  modelRole?: string
  showMsgDuration: boolean
  msgElapsed: number
  showTokens: boolean
  estimatedTokens: number
}

function Header({ presetName, modelName, modelRole, showMsgDuration, msgElapsed, showTokens, estimatedTokens }: HeaderProps) {
  return (
    <div className="flex items-center gap-x-2 gap-y-1 flex-wrap font-mono text-[10px] lowercase tracking-tight text-zinc-500 dark:text-zinc-500">
      {presetName && <span className="text-zinc-600 dark:text-zinc-400">{presetName.toLowerCase()}</span>}
      {modelName && (
        <>
          {presetName && <span className="text-zinc-300 dark:text-zinc-700">/</span>}
          <span className="text-zinc-600 dark:text-zinc-400">{modelName.toLowerCase()}</span>
        </>
      )}
      {modelRole === 'fallback' && (
        <span className="px-1 py-px rounded-sm bg-amber-500/15 text-amber-600 dark:text-amber-400">fallback</span>
      )}
      {showMsgDuration && (
        <>
          <span className="text-zinc-300 dark:text-zinc-700">·</span>
          <span className="tabular-nums">{fmtElapsed(msgElapsed)}</span>
        </>
      )}
      {showTokens && (
        <>
          <span className="text-zinc-300 dark:text-zinc-700">·</span>
          <span className="tabular-nums" title="Estimated tokens in this message">
            ~{fmtTokens(estimatedTokens)} tok
          </span>
        </>
      )}
    </div>
  )
}

interface FooterProps {
  content: string
  canRegenerate?: boolean
  onRegenerate?: () => void
  regenerating?: boolean
}

function Footer({ content, canRegenerate, onRegenerate, regenerating }: FooterProps) {
  const copyText = stripThinking(content)
  const canShowCopy = copyText.length > 0
  const canShowRegenerate = !!canRegenerate && !!onRegenerate
  if (!canShowCopy && !canShowRegenerate) return null
  return (
    <div className="flex items-center gap-1">
      {canShowCopy && <CopyMessageButton text={copyText} />}
      {canShowRegenerate && (
        <button
          onClick={onRegenerate}
          disabled={regenerating}
          className="inline-flex items-center gap-1.5 px-2 py-1 text-[11px] font-medium rounded-md text-zinc-500 dark:text-zinc-500 hover:text-teal-500 dark:hover:text-teal-400 hover:bg-teal-500/5 disabled:opacity-50 transition-colors"
          title="Regenerate response"
        >
          {regenerating ? <Loader2 size={11} className="animate-spin" /> : <RotateCcw size={11} />}
          Regenerate
        </button>
      )}
    </div>
  )
}
