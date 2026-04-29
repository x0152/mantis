import { useState, useEffect, useCallback } from 'react'
import { Plug, ScrollText, Sparkles, Radio, ShieldAlert, Wrench, GitBranch, LogOut, Container } from '@/lib/icons'
import { Toaster } from '@/components/ui/sonner'
import { Button } from '@/components/ui/button'
import LlmPage from './pages/LlmPage'
import ConnectionsPage from './pages/ConnectionsPage'
import ChannelsPage from './pages/ChannelsPage'
import ChatPage from './pages/ChatPage'
import LogsPage from './pages/LogsPage'
import GuardProfilesPage from './pages/GuardProfilesPage'
import SkillsPage from './pages/SkillsPage'
import PlansPage from './pages/PlansPage'
import RuntimesPage from './pages/RuntimesPage'
import SetupWizard from './pages/SetupWizard'
import LoginPage from './pages/LoginPage'
import ChatSidebar from './components/ChatSidebar'
import { MantisLogo } from './components/MantisLogo'
import { LogoStateProvider, useLogoState } from './components/LogoState'
import { ModeToggle } from './components/mode-toggle'
import { api, setUnauthorizedHandler, UnauthorizedError } from './api'
import { parseRoute, navigate, type PageId, type Route } from './router'
import type { ChatSession, User } from './types'

type NavItem = { id: PageId; label: string; icon: typeof Sparkles }
type NavSection = { title: string; items: NavItem[] }

const nav: NavSection[] = [
  {
    title: 'Activity',
    items: [
      { id: 'plans', label: 'Plans', icon: GitBranch },
      { id: 'logs', label: 'Logs', icon: ScrollText },
    ],
  },
  {
    title: 'Configuration',
    items: [
      { id: 'llm', label: 'AI Engine', icon: Sparkles },
      { id: 'connections', label: 'Servers', icon: Plug },
      { id: 'runtimes', label: 'Runtimes', icon: Container },
      { id: 'skills', label: 'Skills', icon: Wrench },
      { id: 'channels', label: 'Channels', icon: Radio },
    ],
  },
  {
    title: 'System',
    items: [
      { id: 'guard-profiles', label: 'Guard Profiles', icon: ShieldAlert },
    ],
  },
]

export default function App() {
  const [route, setRoute] = useState<Route>(parseRoute)
  const [user, setUser] = useState<User | null>(null)
  const [authChecked, setAuthChecked] = useState(false)
  const [needsSetup, setNeedsSetup] = useState<boolean | null>(null)
  const [activeSessionId, setActiveSessionId] = useState<string | null>(null)
  const [sidebarRefreshKey, setSidebarRefreshKey] = useState(0)

  useEffect(() => {
    const sync = () => setRoute(parseRoute())
    window.addEventListener('popstate', sync)
    return () => window.removeEventListener('popstate', sync)
  }, [])

  useEffect(() => {
    setUnauthorizedHandler(() => {
      setUser(null)
      setNeedsSetup(null)
    })
    return () => setUnauthorizedHandler(null)
  }, [])

  useEffect(() => {
    api.auth.me()
      .then(u => setUser(u))
      .catch(err => {
        if (!(err instanceof UnauthorizedError)) console.error(err)
        setUser(null)
      })
      .finally(() => setAuthChecked(true))
  }, [])

  useEffect(() => {
    if (!user) return
    api.llmConnections.list()
      .then(list => setNeedsSetup(list.length === 0))
      .catch(() => setNeedsSetup(false))
  }, [user])

  useEffect(() => {
    if (route.page === 'chat' && 'sessionId' in route && route.sessionId) {
      setActiveSessionId(route.sessionId)
    }
  }, [route])

  useEffect(() => {
    if (needsSetup === false && !activeSessionId) {
      api.chat.getSession().then(s => {
        setActiveSessionId(s.id)
        if (route.page === 'chat') navigate({ page: 'chat', sessionId: s.id })
      }).catch(() => {})
    }
  }, [needsSetup])

  const handleSelectSession = useCallback((session: ChatSession) => {
    setActiveSessionId(session.id)
    navigate({ page: 'chat', sessionId: session.id })
  }, [])

  const handleNewChat = useCallback(async () => {
    try {
      const session = await api.chat.createSession()
      setActiveSessionId(session.id)
      navigate({ page: 'chat', sessionId: session.id })
      setSidebarRefreshKey(k => k + 1)
    } catch {}
  }, [])

  const handleFirstMessage = useCallback(() => {
    setSidebarRefreshKey(k => k + 1)
  }, [])

  const goTo = useCallback((page: PageId) => {
    navigate({ page } as Route)
  }, [])

  const handleLogout = useCallback(async () => {
    try {
      await api.auth.logout()
    } catch {}
    setUser(null)
    setNeedsSetup(null)
  }, [])

  if (!authChecked) return null
  if (!user) return <LoginPage onLogin={setUser} />
  if (needsSetup === null) return null
  if (needsSetup) return <SetupWizard onDone={() => setNeedsSetup(false)} />

  const renderNav = (items: NavItem[]) => items.map(item => (
    <button
      key={item.id}
      onClick={() => goTo(item.id)}
      data-active={route.page === item.id}
      className="nav-row"
    >
      <item.icon size={14} strokeWidth={1.5} className="opacity-80" />
      <span className="truncate" title={item.label}>{item.label}</span>
    </button>
  ))

  const renderSectionLabel = (title: string, idx: number) => (
    <div className="kicker px-3.5 pt-2 pb-1">
      <span className="kicker-num">{String(idx).padStart(2, '0')}</span>
      <span className="kicker-sep">/</span>
      <span>{title.toLowerCase()}</span>
    </div>
  )

  const planId = route.page === 'plans' && 'planId' in route ? route.planId : undefined

  return (
    <LogoStateProvider>
    <div className="flex h-screen bg-white dark:bg-zinc-950 text-zinc-900 dark:text-zinc-100 transition-colors">
      <aside className="w-56 bg-zinc-50/70 dark:bg-zinc-900/40 border-r border-zinc-200/80 dark:border-zinc-800/60 flex flex-col shrink-0 min-w-0 overflow-hidden transition-colors">
        <div className="px-3 py-3.5 border-b border-zinc-200/80 dark:border-zinc-800/60 flex justify-between items-center gap-2 transition-colors min-w-0">
          <div className="flex items-center gap-2 min-w-0">
            <SidebarLogo />
            <div className="leading-tight min-w-0">
              <h1 className="text-[14px] font-medium lowercase text-zinc-900 dark:text-zinc-50 tracking-tight truncate">mantis</h1>
              <p className="font-mono text-[10px] lowercase tracking-tight text-zinc-500 dark:text-zinc-500 mt-0.5 truncate">v0 · control</p>
            </div>
          </div>
          <div className="flex items-center gap-0 shrink-0">
            <ModeToggle />
            <Button
              variant="ghost"
              size="icon"
              onClick={handleLogout}
              title={user ? `Sign out ${user.name}` : 'Sign out'}
            >
              <LogOut className="h-3.5 w-3.5" strokeWidth={1.5} />
              <span className="sr-only">Sign out</span>
            </Button>
          </div>
        </div>

        <div className="flex flex-col flex-1 overflow-hidden">
          <div className="pt-1">
            {renderSectionLabel('chat', 1)}
          </div>
          <div className="flex-1 overflow-auto min-h-0">
            <ChatSidebar
              activeSessionId={activeSessionId}
              onSelect={handleSelectSession}
              onNew={handleNewChat}
              refreshKey={sidebarRefreshKey}
            />
          </div>

          <div className="border-t border-zinc-200/80 dark:border-zinc-800/60 py-1 overflow-auto transition-colors">
            {nav.map((s, i) => (
              <div key={s.title}>
                {renderSectionLabel(s.title, i + 2)}
                <div>
                  {renderNav(s.items)}
                </div>
              </div>
            ))}
          </div>
        </div>
      </aside>
      <main className="flex-1 min-w-0 overflow-auto bg-white dark:bg-zinc-950 transition-colors">
        {route.page === 'chat' && <ChatPage sessionId={activeSessionId ?? ''} onFirstMessage={handleFirstMessage} />}
        {route.page === 'channels' && <ChannelsPage />}
        {route.page === 'llm' && <LlmPage />}
        {route.page === 'connections' && <ConnectionsPage />}
        {route.page === 'runtimes' && <RuntimesPage />}
        {route.page === 'skills' && <SkillsPage />}
        {route.page === 'plans' && <PlansPage deepPlanId={planId} key={planId ?? '_'} />}
        {route.page === 'logs' && <LogsPage />}
        {route.page === 'guard-profiles' && <GuardProfilesPage />}
      </main>
      <Toaster />
    </div>
    </LogoStateProvider>
  )
}

function SidebarLogo() {
  const { state } = useLogoState()
  return (
    <MantisLogo
      size={34}
      state={state}
      className="text-teal-500 dark:text-teal-400 shrink-0"
    />
  )
}
