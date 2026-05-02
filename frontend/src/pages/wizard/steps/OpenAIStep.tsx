import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { FormField } from '@/components/FormField'
import { Loader2 } from '@/lib/icons'
import type { ProviderModel } from '@/types'
import { ModelEditor } from '../components/ModelEditor'
import type { ModelRow } from '../types'

interface OpenAIStepProps {
  baseUrl: string
  apiKey: string
  modelRows: ModelRow[]
  available: ProviderModel[] | null
  loadingModels: boolean
  modelsError: string
  onChangeBaseUrl: (v: string) => void
  onChangeApiKey: (v: string) => void
  onChangeModelRows: (rows: ModelRow[]) => void
  onLoadModels: () => void
}

export function OpenAIStep({
  baseUrl,
  apiKey,
  modelRows,
  available,
  loadingModels,
  modelsError,
  onChangeBaseUrl,
  onChangeApiKey,
  onChangeModelRows,
  onLoadModels,
}: OpenAIStepProps) {
  return (
    <div className="space-y-3.5">
      <FormField label="Server URL" hint="Your provider’s OpenAI-compatible endpoint.">
        <Input value={baseUrl} onChange={e => onChangeBaseUrl(e.target.value)} placeholder="https://api.openai.com/v1" />
      </FormField>
      <FormField label="API key" hint="Anyone with this key can use your account — keep it private.">
        <Input
          type="password"
          value={apiKey}
          onChange={e => onChangeApiKey(e.target.value)}
          placeholder="sk-..."
        />
      </FormField>
      <div className="flex items-center justify-between gap-2">
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={onLoadModels}
          disabled={loadingModels || !baseUrl.trim()}
          className="h-8 text-[12px]"
        >
          {loadingModels ? (
            <>
              <Loader2 size={12} className="animate-spin" /> Loading…
            </>
          ) : (
            'Load models'
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
      <ModelEditor rows={modelRows} onChange={onChangeModelRows} available={available} listId="wizard-openai-models" />
    </div>
  )
}
