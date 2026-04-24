import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from 'react'
import type { MantisLogoState } from './MantisLogo'

interface LogoStateContextValue {
  state: MantisLogoState
  setState: (state: MantisLogoState) => void
}

const LogoStateContext = createContext<LogoStateContextValue>({
  state: 'idle',
  setState: () => {},
})

export function LogoStateProvider({ children }: { children: ReactNode }) {
  const [state, setLogoState] = useState<MantisLogoState>('idle')
  const setState = useCallback((next: MantisLogoState) => {
    setLogoState(prev => (prev === next ? prev : next))
  }, [])
  const value = useMemo(() => ({ state, setState }), [state, setState])
  return <LogoStateContext.Provider value={value}>{children}</LogoStateContext.Provider>
}

export function useLogoState() {
  return useContext(LogoStateContext)
}
