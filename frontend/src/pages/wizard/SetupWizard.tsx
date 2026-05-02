import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Card } from '@/components/ui/card'
import { Toaster } from '@/components/ui/sonner'
import { MantisLogo } from '@/components/MantisLogo'
import { ModeToggle } from '@/components/mode-toggle'
import { AlertCircle } from '@/lib/icons'
import { api } from '@/api'
import type { GonkaConfig, ProviderModel } from '@/types'
import type { ModelRow, Provider, State, StepId, WalletMode } from './types'
import { DEFAULT_GONKA_NODE, MIN_BALANCE_GNK } from './seeds'
import { buildPath } from './path'
import { completeSetup } from './completeSetup'
import { telegramSummary } from './utils'
import { StepHeader } from './components/StepHeader'
import { NavBar } from './components/NavBar'
import { ProviderStep } from './steps/ProviderStep'
import { OpenAIStep } from './steps/OpenAIStep'
import { WalletChoiceStep } from './steps/WalletChoiceStep'
import { WalletImportStep } from './steps/WalletImportStep'
import { WalletCreateStep } from './steps/WalletCreateStep'
import { WalletRevealStep } from './steps/WalletRevealStep'
import { WalletBalanceStep } from './steps/WalletBalanceStep'
import { GonkaModelsStep } from './steps/GonkaModelsStep'
import { TelegramStep } from './steps/TelegramStep'
import { FinishStep } from './steps/FinishStep'

export default function SetupWizard({ onDone }: { onDone: () => void }) {
  const envOpenAIBaseUrl = (import.meta.env.VITE_LLM_BASE_URL as string | undefined) ?? ''
  const envOpenAIApiKey = (import.meta.env.VITE_LLM_API_KEY as string | undefined) ?? ''
  const envOpenAIModels = ((import.meta.env.VITE_LLM_MODEL as string | undefined) ?? '')
    .split(',').map(s => s.trim()).filter(Boolean)
  const envGonkaPrivateKey = (import.meta.env.VITE_GONKA_PRIVATE_KEY as string | undefined) ?? ''
  const envGonkaNodeUrl = (import.meta.env.VITE_GONKA_NODE_URL as string | undefined) ?? ''
  const envTgToken = (import.meta.env.VITE_TG_BOT_TOKEN as string | undefined) ?? ''
  const envTgUserIds = (import.meta.env.VITE_TG_USER_IDS as string | undefined) ?? ''

  const defaultModelRowsFor = useCallback((p: Provider): ModelRow[] => {
    if (p === 'openai' && envOpenAIModels.length > 0) {
      return envOpenAIModels.map((name, i) => ({
        name,
        role: i === 0 ? 'chat' : (i === envOpenAIModels.length - 1 && envOpenAIModels.length > 1 ? 'summary' : ''),
      }))
    }
    return [{ name: '', role: 'chat' }]
  }, [envOpenAIModels])

  const [state, setState] = useState<State>(() => ({
    provider: 'openai',
    openaiBaseUrl: envOpenAIBaseUrl,
    openaiApiKey: envOpenAIApiKey,
    modelRows: defaultModelRowsFor('openai'),
    walletMode: 'create',
    gonkaNodeUrl: envGonkaNodeUrl || '',
    gonkaPrivateKey: envGonkaPrivateKey,
    gonkaAddress: '',
    gonkaMnemonicWords: [],
    mnemonicAcknowledged: false,
    bypassBalance: false,
    gonkaBalance: null,
    tgToken: envTgToken,
    tgLinkedUser: null,
    tgEnvUserIds: envTgUserIds,
    tgSkip: false,
  }))

  const [stepId, setStepId] = useState<StepId>('provider')
  const [gonkaConfig, setGonkaConfig] = useState<GonkaConfig | null>(null)
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const [availableModels, setAvailableModels] = useState<ProviderModel[] | null>(null)
  const [loadingModels, setLoadingModels] = useState(false)
  const [modelsError, setModelsError] = useState('')

  const loadModels = useCallback(
    async (provider: Provider, baseUrl: string, apiKey: string): Promise<ProviderModel[]> => {
      if (!baseUrl.trim()) {
        setModelsError('Fill in the server URL first.')
        return []
      }
      setLoadingModels(true)
      setModelsError('')
      const tempId = `wizard-${provider}-${Date.now()}`
      try {
        await api.llmConnections.create({ id: tempId, provider, baseUrl: baseUrl.trim(), apiKey: apiKey.trim() })
        try {
          const list = await api.llmConnections.listAvailableModels(tempId)
          setAvailableModels(list)
          return list
        } finally {
          await api.llmConnections.delete(tempId).catch(() => {})
        }
      } catch (e) {
        setModelsError(e instanceof Error ? e.message : 'Could not load models')
        return []
      } finally {
        setLoadingModels(false)
      }
    },
    [],
  )

  useEffect(() => {
    setAvailableModels(null)
    setModelsError('')
  }, [state.provider])

  const gonkaAutoLoadedRef = useRef(false)
  useEffect(() => {
    if (stepId !== 'gonka-models') return
    if (gonkaAutoLoadedRef.current) return
    if (availableModels !== null || loadingModels) return
    if (!state.gonkaNodeUrl.trim() || !state.gonkaPrivateKey.trim()) return
    gonkaAutoLoadedRef.current = true
    void (async () => {
      const list = await loadModels('gonka', state.gonkaNodeUrl, state.gonkaPrivateKey)
      setState(prev => {
        const rows = prev.modelRows
        const isEmpty = rows.length === 0 || (rows.length === 1 && !rows[0].name.trim())
        if (isEmpty && list.length > 0) {
          return { ...prev, modelRows: [{ name: list[0].id, role: 'chat' }] }
        }
        return prev
      })
    })()
  }, [stepId, availableModels, loadingModels, state.gonkaNodeUrl, state.gonkaPrivateKey, loadModels])

  useEffect(() => {
    gonkaAutoLoadedRef.current = false
  }, [state.provider, state.gonkaNodeUrl, state.gonkaPrivateKey])

  useEffect(() => {
    api.gonka
      .config()
      .then(cfg => {
        setGonkaConfig(cfg)
        setState(prev => ({
          ...prev,
          gonkaNodeUrl: prev.gonkaNodeUrl || cfg.defaultNodeUrl || DEFAULT_GONKA_NODE,
        }))
      })
      .catch(() =>
        setGonkaConfig({
          defaultNodeUrl: DEFAULT_GONKA_NODE,
          inferencedAvailable: false,
          hasPresetPrivateKey: false,
          hasPresetNodeUrl: false,
          minBalanceGnk: String(MIN_BALANCE_GNK),
        }),
      )
  }, [])

  const path = useMemo(() => buildPath(state, gonkaConfig), [state, gonkaConfig])
  const currentIdx = Math.max(0, path.indexOf(stepId))

  const update = <K extends keyof State>(key: K, value: State[K]) =>
    setState(prev => ({ ...prev, [key]: value }))

  const goNext = () => {
    setError('')
    const next = path[currentIdx + 1]
    if (next) setStepId(next)
  }
  const goBack = () => {
    setError('')
    const prev = path[currentIdx - 1]
    if (prev) setStepId(prev)
  }

  const onCompleteClick = async () => {
    setSubmitting(true)
    setError('')
    try {
      await completeSetup(state)
      onDone()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Setup failed')
    } finally {
      setSubmitting(false)
    }
  }

  const onProviderSelect = (p: Provider) => {
    setState(prev => (prev.provider === p ? prev : { ...prev, provider: p, modelRows: defaultModelRowsFor(p) }))
    if (p === 'openai') setStepId('openai')
    else setStepId('wallet-choice')
  }

  const onWalletModeSelect = (mode: WalletMode) => {
    setState(prev => ({ ...prev, walletMode: mode }))
    setStepId(mode === 'create' ? 'wallet-create' : 'wallet-import')
  }

  const onCreateWallet = async () => {
    setSubmitting(true)
    setError('')
    try {
      const wallet = await api.gonka.createWallet()
      setState(prev => ({
        ...prev,
        gonkaPrivateKey: wallet.privateKeyHex,
        gonkaAddress: wallet.address,
        gonkaMnemonicWords: wallet.words,
        mnemonicAcknowledged: false,
      }))
      setStepId('wallet-reveal')
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Wallet creation failed')
    } finally {
      setSubmitting(false)
    }
  }

  const onUseExisting = async () => {
    setSubmitting(true)
    setError('')
    try {
      const { address } = await api.gonka.deriveAddress(state.gonkaPrivateKey.trim())
      update('gonkaAddress', address)
      setStepId('wallet-balance')
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Invalid private key')
    } finally {
      setSubmitting(false)
    }
  }

  const onLoadOpenAIModels = async () => {
    const list = await loadModels('openai', state.openaiBaseUrl, state.openaiApiKey)
    setState(prev => {
      const rows = prev.modelRows
      const isEmpty = rows.length === 0 || (rows.length === 1 && !rows[0].name.trim())
      if (isEmpty && list.length > 0) return { ...prev, modelRows: [{ name: list[0].id, role: 'chat' }] }
      return prev
    })
  }

  const onReloadGonkaModels = async () => {
    const list = await loadModels('gonka', state.gonkaNodeUrl, state.gonkaPrivateKey)
    setState(prev => {
      const rows = prev.modelRows
      const isEmpty = rows.length === 0 || (rows.length === 1 && !rows[0].name.trim())
      if (isEmpty && list.length > 0) return { ...prev, modelRows: [{ name: list[0].id, role: 'chat' }] }
      return prev
    })
  }

  return (
    <div className="min-h-screen bg-zinc-50 dark:bg-zinc-950 flex items-center justify-center p-6">
      <Toaster />
      <div className="w-full max-w-xl">
        <div className="relative mb-6 flex items-center justify-center gap-3">
          <MantisLogo size={40} className="text-teal-500 dark:text-teal-400" />
          <div className="leading-tight">
            <h1 className="font-mono text-[20px] font-medium lowercase tracking-tight text-zinc-900 dark:text-zinc-50">
              mantis
            </h1>
            <p className="font-mono text-[11px] lowercase tracking-tight text-zinc-500 mt-0.5">let’s set you up</p>
          </div>
          <div className="absolute right-0 top-0">
            <ModeToggle />
          </div>
        </div>

        <Card className="bg-white dark:bg-zinc-900/60 p-6 space-y-6">
          <StepHeader path={path} stepId={stepId} />

          {stepId === 'provider' && (
            <ProviderStep provider={state.provider} gonkaConfig={gonkaConfig} onSelect={onProviderSelect} />
          )}

          {stepId === 'openai' && (
            <OpenAIStep
              baseUrl={state.openaiBaseUrl}
              apiKey={state.openaiApiKey}
              modelRows={state.modelRows}
              available={availableModels}
              loadingModels={loadingModels}
              modelsError={modelsError}
              onChangeBaseUrl={v => update('openaiBaseUrl', v)}
              onChangeApiKey={v => update('openaiApiKey', v)}
              onChangeModelRows={rows => update('modelRows', rows)}
              onLoadModels={onLoadOpenAIModels}
            />
          )}

          {stepId === 'wallet-choice' && (
            <WalletChoiceStep
              walletMode={state.walletMode}
              gonkaConfig={gonkaConfig}
              onSelect={onWalletModeSelect}
            />
          )}

          {stepId === 'wallet-import' && (
            <WalletImportStep
              privateKey={state.gonkaPrivateKey}
              nodeUrl={state.gonkaNodeUrl}
              submitting={submitting}
              onChangePrivateKey={v => update('gonkaPrivateKey', v)}
              onChangeNodeUrl={v => update('gonkaNodeUrl', v)}
              onConfirm={onUseExisting}
            />
          )}

          {stepId === 'wallet-create' && (
            <WalletCreateStep
              gonkaConfig={gonkaConfig}
              nodeUrl={state.gonkaNodeUrl}
              submitting={submitting}
              onChangeNodeUrl={v => update('gonkaNodeUrl', v)}
              onCreate={onCreateWallet}
            />
          )}

          {stepId === 'wallet-reveal' && (
            <WalletRevealStep
              address={state.gonkaAddress}
              words={state.gonkaMnemonicWords}
              acknowledged={state.mnemonicAcknowledged}
              onAcknowledge={v => update('mnemonicAcknowledged', v)}
            />
          )}

          {stepId === 'wallet-balance' && (
            <WalletBalanceStep
              address={state.gonkaAddress}
              nodeUrl={state.gonkaNodeUrl}
              minBalance={MIN_BALANCE_GNK}
              balance={state.gonkaBalance}
              onBalanceChange={b => update('gonkaBalance', b)}
              bypass={state.bypassBalance}
              onBypassChange={v => update('bypassBalance', v)}
            />
          )}

          {stepId === 'gonka-models' && (
            <GonkaModelsStep
              modelRows={state.modelRows}
              available={availableModels}
              loadingModels={loadingModels}
              modelsError={modelsError}
              onChange={rows => update('modelRows', rows)}
              onReload={onReloadGonkaModels}
            />
          )}

          {stepId === 'telegram' && (
            <TelegramStep
              token={state.tgToken}
              linkedUser={state.tgLinkedUser}
              skip={state.tgSkip}
              onChangeToken={v => update('tgToken', v)}
              onChangeLinkedUser={v => update('tgLinkedUser', v)}
              onChangeSkip={v => update('tgSkip', v)}
            />
          )}

          {stepId === 'finish' && (
            <FinishStep
              provider={state.provider}
              chatModel={state.modelRows.find(r => r.role === 'chat')?.name}
              endpoint={state.provider === 'openai' ? state.openaiBaseUrl : state.gonkaNodeUrl}
              telegramLabel={telegramSummary(state)}
            />
          )}

          {error && (
            <div className="flex items-start gap-2 rounded-md border border-rose-500/30 bg-rose-500/5 px-3 py-2 text-[12px] text-rose-500">
              <AlertCircle size={14} className="mt-0.5 shrink-0" />
              <span>{error}</span>
            </div>
          )}

          <NavBar
            stepId={stepId}
            canBack={currentIdx > 0 && stepId !== 'wallet-reveal'}
            onBack={goBack}
            onNext={goNext}
            onComplete={onCompleteClick}
            submitting={submitting}
            state={state}
          />
        </Card>
      </div>
    </div>
  )
}
