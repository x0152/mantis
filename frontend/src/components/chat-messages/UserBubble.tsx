import type { ChatMessage } from '../../types'
import { CopyMessageButton } from './CopyMessageButton'
import { UserMessageContent } from './UserMessageContent'
import { estimateTokens, fmtTokens } from './utils'

export function UserBubble({ msg }: { msg: ChatMessage }) {
  const userTokens = msg.tokens ?? estimateTokens(msg.content)
  return (
    <div className="group flex justify-end items-center gap-2">
      {msg.content && <CopyMessageButton text={msg.content} variant="hover" />}
      <div className="max-w-[80%] rounded-md text-[14.5px] bg-teal-500/10 dark:bg-teal-500/[0.08] text-zinc-900 dark:text-zinc-50 border border-teal-500/30 dark:border-teal-500/30 border-r-2 border-r-teal-500 px-4 py-2.5 leading-relaxed">
        <UserMessageContent content={msg.content} sessionId={msg.sessionId} />
        {userTokens > 0 && (
          <div className="mt-1 font-mono text-[10px] lowercase tabular-nums text-zinc-500 dark:text-zinc-500 text-right">
            ~{fmtTokens(userTokens)} tok
          </div>
        )}
      </div>
    </div>
  )
}
