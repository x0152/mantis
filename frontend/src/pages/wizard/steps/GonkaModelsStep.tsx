import { Button } from '@/components/ui/button'
import { Loader2, RotateCw } from '@/lib/icons'
import type { ProviderModel } from '@/types'
import { ModelEditor } from '../components/ModelEditor'
import type { ModelRow } from '../types'

interface GonkaModelsStepProps {
  modelRows: ModelRow[]
  available: ProviderModel[] | null
  loadingModels: boolean
  modelsError: string
  onChange: (rows: ModelRow[]) => void
  onReload: () => void
}

export function GonkaModelsStep({
  modelRows,
  available,
  loadingModels,
  modelsError,
  onChange,
  onReload,
}: GonkaModelsStepProps) {
  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between gap-2">
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={onReload}
          disabled={loadingModels}
          className="h-8 text-[12px]"
        >
          {loadingModels ? (
            <>
              <Loader2 size={12} className="animate-spin" /> Loading…
            </>
          ) : (
            <>
              <RotateCw size={12} /> Reload
            </>
          )}
        </Button>
        {available && (
          <span className="text-[11px] text-zinc-500 dark:text-zinc-500">
            {available.length} model{available.length === 1 ? '' : 's'} found
          </span>
        )}
      </div>
      {modelsError && (
        <div className="rounded-md border border-rose-500/30 bg-rose-500/5 px-3 py-2 text-[12px] text-rose-500">
          {modelsError}
        </div>
      )}
      {available && available.length === 0 && !loadingModels && !modelsError && (
        <div className="rounded-md border border-zinc-200/70 dark:border-zinc-800/60 bg-zinc-50 dark:bg-zinc-900 px-3 py-2 text-[12px] text-zinc-600 dark:text-zinc-400">
          No models found on this Gonka server. Type the model name yourself, or change the server in the previous step.
        </div>
      )}
      <ModelEditor rows={modelRows} onChange={onChange} available={available} listId="wizard-gonka-models" />
    </div>
  )
}
