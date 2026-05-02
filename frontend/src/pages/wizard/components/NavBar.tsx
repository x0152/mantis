import { Button } from '@/components/ui/button'
import { ArrowLeft, CheckCircle2, Loader2 } from '@/lib/icons'
import type { State, StepId } from '../types'
import { MIN_BALANCE_GNK } from '../seeds'

interface NavBarProps {
  stepId: StepId
  canBack: boolean
  onBack: () => void
  onNext: () => void
  onComplete: () => void
  submitting: boolean
  state: State
}

export function NavBar({ stepId, canBack, onBack, onNext, onComplete, submitting, state }: NavBarProps) {
  const isFinal = stepId === 'finish'
  const nextDisabled = isNextDisabled(stepId, state)

  if (stepId === 'wallet-choice' || stepId === 'wallet-create' || stepId === 'wallet-import') {
    return canBack ? (
      <div className="flex items-center justify-between pt-2">
        <Button variant="ghost" size="sm" onClick={onBack}>
          <ArrowLeft size={12} /> Back
        </Button>
        <span />
      </div>
    ) : null
  }

  return (
    <div className="flex items-center justify-between pt-2">
      <div>
        {canBack && (
          <Button variant="ghost" size="sm" onClick={onBack}>
            <ArrowLeft size={12} /> Back
          </Button>
        )}
      </div>
      {isFinal ? (
        <Button onClick={onComplete} disabled={submitting} className="h-9 px-4">
          {submitting ? (
            <>
              <Loader2 size={14} className="animate-spin" /> saving…
            </>
          ) : (
            <>
              <CheckCircle2 size={14} /> Finish
            </>
          )}
        </Button>
      ) : (
        <Button onClick={onNext} disabled={nextDisabled} className="h-9 px-4">
          Next
        </Button>
      )}
    </div>
  )
}

function isNextDisabled(stepId: StepId, state: State): boolean {
  switch (stepId) {
    case 'provider':
      return false
    case 'openai': {
      const hasChat = state.modelRows.some(r => r.role === 'chat' && r.name.trim())
      return !state.openaiBaseUrl.trim() || !hasChat
    }
    case 'wallet-choice':
    case 'wallet-create':
    case 'wallet-import':
      return true
    case 'wallet-reveal':
      return !state.mnemonicAcknowledged
    case 'wallet-balance':
      return !state.bypassBalance && !hasEnoughBalance(state)
    case 'gonka-models': {
      const hasChat = state.modelRows.some(r => r.role === 'chat' && r.name.trim())
      return !hasChat
    }
    case 'telegram':
      return !state.tgSkip && !state.tgLinkedUser
    case 'finish':
      return false
  }
}

function hasEnoughBalance(state: State): boolean {
  return !!state.gonkaBalance && state.gonkaBalance.gnk >= MIN_BALANCE_GNK
}
