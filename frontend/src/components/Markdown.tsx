import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import remarkBreaks from 'remark-breaks'
import { ThinkingBlock } from './ThinkingBlock'
import {
  AttachmentBadgeRow,
  hasAttachmentTags,
  parseAttachmentBlocks,
} from './AttachmentTag'

type Segment =
  | { type: 'text'; content: string }
  | { type: 'thinking'; content: string; streaming: boolean }

function splitThinking(input: string): Segment[] {
  const segments: Segment[] = []
  let rest = input
  while (rest.length > 0) {
    const open = rest.indexOf('<think>')
    if (open < 0) {
      segments.push({ type: 'text', content: rest })
      break
    }
    if (open > 0) {
      segments.push({ type: 'text', content: rest.slice(0, open) })
    }
    const afterOpen = rest.slice(open + 7)
    const close = afterOpen.indexOf('</think>')
    if (close < 0) {
      segments.push({ type: 'thinking', content: afterOpen, streaming: true })
      break
    }
    segments.push({ type: 'thinking', content: afterOpen.slice(0, close), streaming: false })
    rest = afterOpen.slice(close + 8)
  }
  return segments
}

export function Markdown({
  content,
  sessionId,
}: {
  content: string
  sessionId?: string
}) {
  if (!content) return null

  const segments = splitThinking(content)

  return (
    <div className="markdown text-[15px] leading-relaxed">
      {segments.map((seg, i) => {
        if (seg.type === 'thinking') {
          return <ThinkingBlock key={i} content={seg.content} streaming={seg.streaming} />
        }
        if (!seg.content) return null
        return <MarkdownTextSegment key={i} content={seg.content} sessionId={sessionId} />
      })}
    </div>
  )
}

function MarkdownTextSegment({
  content,
  sessionId,
}: {
  content: string
  sessionId?: string
}) {
  if (!sessionId || !hasAttachmentTags(content)) {
    return <MarkdownSegment content={content} />
  }
  const blocks = parseAttachmentBlocks(content)
  return (
    <>
      {blocks.map((block, i) => {
        if (block.type === 'text') {
          return <MarkdownSegment key={i} content={block.text} />
        }
        return (
          <div key={i} className="my-1.5">
            <AttachmentBadgeRow items={block.items} sessionId={sessionId} variant="message" />
          </div>
        )
      })}
    </>
  )
}

function MarkdownSegment({ content }: { content: string }) {
  return (
    <ReactMarkdown
        skipHtml
        remarkPlugins={[remarkGfm, remarkBreaks]}
        components={{
          p: ({ children }) => <p className="m-0 mb-2 last:mb-0 whitespace-pre-wrap">{children}</p>,

          a: ({ href, children }) => (
            <a
              href={href ?? '#'}
              target="_blank"
              rel="noreferrer"
              className="text-teal-400 underline underline-offset-2 hover:text-teal-300"
            >
              {children}
            </a>
          ),

          strong: ({ children }) => <strong className="font-semibold text-zinc-900 dark:text-zinc-100">{children}</strong>,
          em: ({ children }) => <em className="italic">{children}</em>,
          del: ({ children }) => <del className="line-through">{children}</del>,

          h1: ({ children }) => <h1 className="m-0 mt-3 mb-2 text-lg font-bold text-zinc-900 dark:text-zinc-100">{children}</h1>,
          h2: ({ children }) => <h2 className="m-0 mt-3 mb-2 text-base font-bold text-zinc-900 dark:text-zinc-100">{children}</h2>,
          h3: ({ children }) => <h3 className="m-0 mt-3 mb-2 text-sm font-semibold text-zinc-800 dark:text-zinc-200">{children}</h3>,
          h4: ({ children }) => <h4 className="m-0 mt-3 mb-2 text-sm font-semibold text-zinc-800 dark:text-zinc-200">{children}</h4>,
          h5: ({ children }) => <h5 className="m-0 mt-3 mb-2 text-sm font-semibold text-zinc-700 dark:text-zinc-300">{children}</h5>,
          h6: ({ children }) => <h6 className="m-0 mt-3 mb-2 text-sm font-semibold text-zinc-700 dark:text-zinc-300">{children}</h6>,

          ul: ({ children }) => <ul className="m-0 mb-2 last:mb-0 pl-5 list-disc space-y-1">{children}</ul>,
          ol: ({ children }) => <ol className="m-0 mb-2 last:mb-0 pl-5 list-decimal space-y-1">{children}</ol>,
          li: ({ children }) => <li className="m-0">{children}</li>,

          blockquote: ({ children }) => (
            <blockquote className="m-0 mb-2 last:mb-0 pl-3 border-l-2 border-zinc-300 dark:border-zinc-700 text-zinc-600 dark:text-zinc-400 italic">
              {children}
            </blockquote>
          ),

          hr: () => <hr className="my-3 border-zinc-200 dark:border-zinc-800" />,

          code: ({ className, children, ...props }) => {
            const code = String(children).replace(/\n$/, '')
            return (
              <code className={`font-mono ${className ?? ''}`} {...props}>
                {code}
              </code>
            )
          },

          pre: ({ children }) => (
            <pre className="m-0 mb-2 last:mb-0 bg-zinc-100 dark:bg-zinc-950 text-zinc-700 dark:text-zinc-300 rounded-lg p-3 overflow-auto border border-zinc-200 dark:border-zinc-800">
              {children}
            </pre>
          ),

          table: ({ children }) => (
            <div className="mb-2 last:mb-0 overflow-auto rounded-lg border border-zinc-200 dark:border-zinc-800">
              <table className="w-full text-left text-sm">{children}</table>
            </div>
          ),
          thead: ({ children }) => <thead className="bg-zinc-100 dark:bg-zinc-800/50">{children}</thead>,
          th: ({ children }) => <th className="px-3 py-2 border-b border-zinc-200 dark:border-zinc-800 font-semibold text-zinc-700 dark:text-zinc-300">{children}</th>,
          td: ({ children }) => <td className="px-3 py-2 border-b border-zinc-200/50 dark:border-zinc-800/50 align-top text-zinc-600 dark:text-zinc-400">{children}</td>,

          img: ({ src, alt }) => (
            <img src={src ?? ''} alt={alt ?? ''} className="max-w-full rounded-lg border border-zinc-200 dark:border-zinc-800" />
          ),
        }}
    >
      {content}
    </ReactMarkdown>
  )
}
