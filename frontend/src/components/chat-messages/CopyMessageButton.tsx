import { useEffect, useRef, useState } from 'react'
import { Check, Copy } from '@/lib/icons'

export interface CopyMessageButtonProps {
  text: string
  variant?: 'inline' | 'hover'
  tone?: 'neutral' | 'danger'
}

export function CopyMessageButton({ text, variant = 'inline', tone = 'neutral' }: CopyMessageButtonProps) {
  const [copied, setCopied] = useState(false)
  const resetRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => () => {
    if (resetRef.current) clearTimeout(resetRef.current)
  }, [])

  const handleCopy = async (e: React.MouseEvent<HTMLButtonElement>) => {
    e.stopPropagation()
    if (!text) return
    try {
      await navigator.clipboard.writeText(text)
    } catch {
      try {
        const ta = document.createElement('textarea')
        ta.value = text
        ta.style.position = 'fixed'
        ta.style.left = '-9999px'
        document.body.appendChild(ta)
        ta.select()
        document.execCommand('copy')
        document.body.removeChild(ta)
      } catch {
        return
      }
    }
    setCopied(true)
    if (resetRef.current) clearTimeout(resetRef.current)
    resetRef.current = setTimeout(() => setCopied(false), 1500)
  }

  if (variant === 'hover') {
    return (
      <button
        onClick={handleCopy}
        type="button"
        className="opacity-0 group-hover:opacity-100 focus-visible:opacity-100 transition-opacity p-1.5 rounded-md text-zinc-500 dark:text-zinc-500 hover:text-teal-500 dark:hover:text-teal-400 hover:bg-zinc-200 dark:hover:bg-zinc-800 shrink-0"
        title={copied ? 'Copied' : 'Copy message'}
        aria-label={copied ? 'Copied' : 'Copy message'}
      >
        {copied ? <Check size={13} className="text-emerald-500" /> : <Copy size={13} />}
      </button>
    )
  }

  const toneClasses = tone === 'danger'
    ? 'text-red-400 hover:text-red-300 hover:bg-red-500/10'
    : 'text-zinc-500 dark:text-zinc-500 hover:text-teal-500 dark:hover:text-teal-400 hover:bg-teal-500/5'

  return (
    <button
      onClick={handleCopy}
      type="button"
      className={`inline-flex items-center gap-1.5 px-2 py-1 text-[11px] font-medium rounded-md transition-colors ${toneClasses}`}
      title={copied ? 'Copied' : 'Copy message'}
      aria-label={copied ? 'Copied' : 'Copy message'}
    >
      {copied ? <Check size={11} className="text-emerald-500" /> : <Copy size={11} />}
      {copied ? 'Copied' : 'Copy'}
    </button>
  )
}
