import { useCallback, useEffect, useRef, useState } from 'react'
import { Button } from '@/components/ui/button'
import { Loader2, RotateCw } from '@/lib/icons'
import { api } from '@/api'
import type { GonkaBalance } from '@/types'
import { AddressDisplay } from '../components/AddressDisplay'
import { formatGnk } from '../utils'

interface WalletBalanceStepProps {
  address: string
  nodeUrl: string
  minBalance: number
  balance: GonkaBalance | null
  onBalanceChange: (b: GonkaBalance | null) => void
  bypass: boolean
  onBypassChange: (v: boolean) => void
}

export function WalletBalanceStep({
  address,
  nodeUrl,
  minBalance,
  balance,
  onBalanceChange,
  bypass,
  onBypassChange,
}: WalletBalanceStepProps) {
  const [loading, setLoading] = useState(false)
  const [refreshError, setRefreshError] = useState('')

  const onBalanceChangeRef = useRef(onBalanceChange)
  useEffect(() => {
    onBalanceChangeRef.current = onBalanceChange
  })

  const refresh = useCallback(async () => {
    if (!address || !nodeUrl) return
    setLoading(true)
    setRefreshError('')
    try {
      const b = await api.gonka.balance(address, nodeUrl)
      onBalanceChangeRef.current(b)
    } catch (e) {
      setRefreshError(e instanceof Error ? e.message : 'Balance check failed')
    } finally {
      setLoading(false)
    }
  }, [address, nodeUrl])

  const sufficient = !!balance && balance.gnk >= minBalance

  useEffect(() => {
    if (sufficient) return
    void refresh()
    const id = window.setInterval(refresh, 15_000)
    return () => window.clearInterval(id)
  }, [refresh, sufficient])

  return (
    <div className="space-y-4">
      <AddressDisplay label="Send GNK to this address" address={address} />

      <div
        className={`rounded-md border px-4 py-3 transition-colors ${
          sufficient
            ? 'border-teal-500/40 bg-teal-500/5'
            : 'border-zinc-200/80 dark:border-zinc-800/70 bg-zinc-50 dark:bg-zinc-900'
        }`}
      >
        <div className="flex items-center justify-between gap-3">
          <div>
            <div className="text-[11px] uppercase tracking-wide text-zinc-500 dark:text-zinc-500">Balance</div>
            <div className="mt-1 text-[20px] font-medium tabular-nums text-zinc-900 dark:text-zinc-50">
              {balance ? `${formatGnk(balance.gnk)} GNK` : '—'}
            </div>
          </div>
          <Button variant="secondary" size="sm" onClick={refresh} disabled={loading}>
            {loading ? <Loader2 size={12} className="animate-spin" /> : <RotateCw size={12} />} refresh
          </Button>
        </div>
        {refreshError && <p className="mt-2 text-[11.5px] text-rose-500">{refreshError}</p>}
        {!sufficient && (
          <p className="mt-2 text-[11.5px] text-zinc-500 dark:text-zinc-500">
            We check every few seconds. You need at least {minBalance} GNK.
          </p>
        )}
      </div>

      <label className="flex items-start gap-2 text-[12px] text-zinc-500 dark:text-zinc-500 cursor-pointer">
        <input
          type="checkbox"
          checked={bypass}
          onChange={e => onBypassChange(e.target.checked)}
          className="mt-0.5 size-4 accent-amber-500"
        />
        <span>I’ll add money later. Skip this for now.</span>
      </label>
    </div>
  )
}
