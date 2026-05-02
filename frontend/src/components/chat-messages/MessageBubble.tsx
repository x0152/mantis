import type { ChatMessage, Step } from '../../types'
import { AssistantBubble } from './AssistantBubble'
import { ErrorBubble } from './ErrorBubble'
import { UserBubble } from './UserBubble'

interface MessageBubbleProps {
  msg: ChatMessage
  onStepClick: (s: Step) => void
  canRegenerate?: boolean
  onRegenerate?: () => void
  regenerating?: boolean
}

export function MessageBubble({ msg, onStepClick, canRegenerate, onRegenerate, regenerating }: MessageBubbleProps) {
  if (msg.role === 'user') return <UserBubble msg={msg} />
  if (msg.status === 'error') {
    return (
      <ErrorBubble
        msg={msg}
        canRegenerate={canRegenerate}
        onRegenerate={onRegenerate}
        regenerating={regenerating}
      />
    )
  }
  return (
    <AssistantBubble
      msg={msg}
      onStepClick={onStepClick}
      canRegenerate={canRegenerate}
      onRegenerate={onRegenerate}
      regenerating={regenerating}
    />
  )
}
