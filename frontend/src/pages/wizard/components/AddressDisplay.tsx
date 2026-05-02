import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Copy } from '@/lib/icons'

export function AddressDisplay({ label, address }: { label: string; address: string }) {
  const copy = () => {
    navigator.clipboard.writeText(address)
    toast.success(`${label} copied`)
  }
  return (
    <div className="space-y-2">
      <Label>{label}</Label>
      <div className="flex items-center gap-2">
        <code className="flex-1 truncate rounded-md border border-zinc-200/80 dark:border-zinc-800/70 bg-zinc-50 dark:bg-zinc-900 px-3 py-2 text-[12px] font-mono text-zinc-700 dark:text-zinc-300">
          {address || '—'}
        </code>
        <Button variant="secondary" size="sm" onClick={copy} disabled={!address}>
          <Copy size={12} /> copy
        </Button>
      </div>
    </div>
  )
}
