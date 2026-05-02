import { useCallback, useEffect, useRef, useState } from 'react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { FormField } from '@/components/FormField'
import { CheckCircle2, Copy, Link2, Loader2, Send } from '@/lib/icons'
import { api } from '@/api'
import type { TelegramWizardBot, TelegramWizardUser } from '@/types'

interface TelegramStepProps {
  token: string
  linkedUser: TelegramWizardUser | null
  skip: boolean
  onChangeToken: (v: string) => void
  onChangeLinkedUser: (user: TelegramWizardUser | null) => void
  onChangeSkip: (v: boolean) => void
}

export function TelegramStep({
  token,
  linkedUser,
  skip,
  onChangeToken,
  onChangeLinkedUser,
  onChangeSkip,
}: TelegramStepProps) {
  const [bot, setBot] = useState<TelegramWizardBot | null>(null)
  const [verifying, setVerifying] = useState(false)
  const [verifyError, setVerifyError] = useState('')
  const [waiting, setWaiting] = useState(false)
  const initialTokenRef = useRef(token)

  const reset = useCallback(() => {
    setBot(null)
    setVerifyError('')
    setWaiting(false)
    onChangeLinkedUser(null)
  }, [onChangeLinkedUser])

  const verify = useCallback(
    async (value: string): Promise<TelegramWizardBot | null> => {
      const trimmed = value.trim()
      if (!trimmed) {
        reset()
        return null
      }
      setVerifying(true)
      setVerifyError('')
      try {
        const result = await api.telegram.verify(trimmed)
        setBot(result)
        return result
      } catch (e) {
        setBot(null)
        setVerifyError(e instanceof Error ? e.message : 'Could not reach Telegram')
        return null
      } finally {
        setVerifying(false)
      }
    },
    [reset],
  )

  useEffect(() => {
    if (initialTokenRef.current.trim()) void verify(initialTokenRef.current)
  }, [verify])

  useEffect(() => {
    if (skip || !bot || !token.trim() || linkedUser) {
      setWaiting(false)
      return
    }
    setWaiting(true)
    let cancelled = false
    const tick = async () => {
      try {
        const res = await api.telegram.status(token.trim())
        if (cancelled) return
        if (res.user) {
          onChangeLinkedUser(res.user)
          setWaiting(false)
        }
      } catch {}
    }
    void tick()
    const id = setInterval(tick, 2500)
    return () => {
      cancelled = true
      clearInterval(id)
    }
  }, [skip, bot, token, linkedUser, onChangeLinkedUser])

  const copyCode = async (code: string) => {
    try {
      await navigator.clipboard.writeText(code)
      toast.success('Code copied')
    } catch {
      toast.error('Copy failed')
    }
  }

  const inputsDisabled = skip

  return (
    <div className="space-y-3">
      <FormField
        label="Bot token"
        hint={
          <>
            Get one from{' '}
            <a
              href="https://t.me/BotFather"
              target="_blank"
              rel="noreferrer"
              className="text-teal-600 dark:text-teal-400 hover:underline"
            >
              @BotFather
            </a>
            : open the chat, send <code className="font-mono">/newbot</code>, follow the steps. You can connect later via the <code className="font-mono">TG_BOT_TOKEN</code> env variable.
          </>
        }
      >
        <div className="flex gap-2">
          <Input
            value={token}
            onChange={e => {
              onChangeToken(e.target.value)
              reset()
            }}
            onBlur={e => {
              if (e.target.value.trim() && !bot && !skip) void verify(e.target.value)
            }}
            placeholder="123456:ABC-DEF..."
            className="flex-1"
            disabled={inputsDisabled}
          />
          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={!token.trim() || verifying || inputsDisabled}
            onClick={() => void verify(token)}
            className="h-9 text-[12px] shrink-0"
          >
            {verifying ? (
              <>
                <Loader2 size={12} className="animate-spin" /> Checking…
              </>
            ) : bot ? (
              'Re-check'
            ) : (
              'Verify'
            )}
          </Button>
        </div>
      </FormField>

      {verifyError && !skip && (
        <div className="rounded-md border border-rose-500/30 bg-rose-500/5 px-3 py-2 text-[12px] text-rose-500">
          {verifyError}
        </div>
      )}

      {bot && !linkedUser && !skip && (
        <div className="space-y-2">
          <div className="rounded-md border border-zinc-200/80 dark:border-zinc-800/70 bg-zinc-50 dark:bg-zinc-900 px-3 py-2.5 text-[12.5px] text-zinc-700 dark:text-zinc-300 space-y-2.5">
            <div className="flex items-center gap-1.5">
              <CheckCircle2 size={13} className="text-teal-500" />
              <span>
                Connected to <span className="font-mono">{bot.name || bot.username}</span>
                {bot.username && (
                  <>
                    {' '}—{' '}
                    <a
                      href={bot.link}
                      target="_blank"
                      rel="noreferrer"
                      className="font-mono text-teal-600 dark:text-teal-400 hover:underline"
                    >
                      @{bot.username}
                    </a>
                  </>
                )}
              </span>
            </div>

            <div className="space-y-1.5">
              <div className="text-[11.5px] uppercase tracking-wide text-zinc-500">
                Send this code to the bot
              </div>
              <div className="flex items-center gap-2">
                <code className="flex-1 rounded-md border border-zinc-200/70 dark:border-zinc-800/70 bg-white dark:bg-zinc-950 px-3 py-2 font-mono text-[18px] tracking-[0.3em] text-zinc-900 dark:text-zinc-50 text-center">
                  {bot.code}
                </code>
                <Button type="button" variant="outline" size="sm" onClick={() => void copyCode(bot.code)} className="h-9 shrink-0">
                  <Copy size={12} /> copy
                </Button>
                {bot.deepLink && (
                  <a
                    href={bot.deepLink}
                    target="_blank"
                    rel="noreferrer"
                    className="inline-flex h-9 shrink-0 items-center gap-1.5 rounded-md border border-teal-500/40 bg-teal-500/10 px-3 text-[12px] text-teal-600 dark:text-teal-400 hover:bg-teal-500/15"
                  >
                    <Link2 size={12} /> open bot
                  </a>
                )}
              </div>
              <p className="text-[11.5px] text-zinc-500 dark:text-zinc-500">
                Open the bot, paste the code as a message, and send it. Mantis will detect you automatically.
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2 rounded-md border border-zinc-200/70 dark:border-zinc-800/60 px-3 py-2 text-[12px] text-zinc-500 dark:text-zinc-400">
            {waiting ? (
              <>
                <Loader2 size={13} className="animate-spin text-teal-500" />
                Waiting for the code from your Telegram…
              </>
            ) : (
              <>
                <Send size={13} className="text-zinc-400" />
                Verify the bot to start listening for the code.
              </>
            )}
          </div>
        </div>
      )}

      {bot && linkedUser && !skip && (
        <div className="rounded-md border border-teal-500/30 bg-teal-500/5 px-3 py-2.5 text-[12.5px] text-zinc-700 dark:text-zinc-300 space-y-1.5">
          <div className="flex items-center gap-1.5">
            <CheckCircle2 size={14} className="text-teal-500" />
            <span>
              Linked as <span className="font-mono">{linkedUser.name || linkedUser.username || linkedUser.id}</span>
              {linkedUser.username && (
                <span className="font-mono text-zinc-500"> · @{linkedUser.username}</span>
              )}
              <span className="font-mono text-zinc-500"> · id {linkedUser.id}</span>
            </span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-[11.5px] text-zinc-500 dark:text-zinc-500">
              Only this account can talk to the bot. You can add more later in Hosts → Channels.
            </span>
            <Button type="button" variant="ghost" size="sm" onClick={() => onChangeLinkedUser(null)} className="h-7 text-[11.5px]">
              re-link
            </Button>
          </div>
        </div>
      )}

      <label
        className={`flex items-start gap-2 rounded-md border px-3 py-2 text-[12px] cursor-pointer transition-colors ${
          skip
            ? 'border-zinc-300 dark:border-zinc-700 bg-zinc-100/60 dark:bg-zinc-900 text-zinc-700 dark:text-zinc-300'
            : 'border-zinc-200/70 dark:border-zinc-800/60 text-zinc-500 dark:text-zinc-400 hover:border-zinc-300 dark:hover:border-zinc-700'
        }`}
      >
        <input
          type="checkbox"
          checked={skip}
          onChange={e => onChangeSkip(e.target.checked)}
          className="mt-0.5 size-4 accent-teal-600"
        />
        <span>I don’t want to connect Telegram — chat is enough.</span>
      </label>
    </div>
  )
}
