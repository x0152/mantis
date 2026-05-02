import { AttachmentBadgeRow, hasAttachmentTags, parseAttachmentBlocks } from '../AttachmentTag'

export function UserMessageContent({ content, sessionId }: { content: string; sessionId: string }) {
  if (!content) return null
  if (!hasAttachmentTags(content)) {
    return <div className="whitespace-pre-wrap">{content}</div>
  }
  const blocks = parseAttachmentBlocks(content)
  return (
    <div className="space-y-2">
      {blocks.map((block, i) => {
        if (block.type === 'text') {
          return (
            <div key={i} className="whitespace-pre-wrap">
              {block.text}
            </div>
          )
        }
        return <AttachmentBadgeRow key={i} items={block.items} sessionId={sessionId} variant="user" />
      })}
    </div>
  )
}
