import { useState, useEffect, useCallback } from 'react'
import { Plug, ScrollText, Sparkles, Radio, ShieldAlert, Wrench, GitBranch, LogOut } from 'lucide-react'
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
      className={`w-full flex items-center gap-2.5 px-3 py-2 rounded-lg text-[13px] font-medium min-w-0 ${
        route.page === item.id
          ? 'bg-teal-500/10 text-teal-600 dark:text-teal-400'
          : 'text-zinc-600 dark:text-zinc-400 hover:text-zinc-900 dark:hover:text-zinc-200 hover:bg-zinc-100 dark:hover:bg-zinc-800/50'
      }`}
    >
      <item.icon size={16} strokeWidth={1.8} />
      <span className="truncate" title={item.label}>{item.label}</span>
    </button>
  ))

  const planId = route.page === 'plans' && 'planId' in route ? route.planId : undefined

  return (
    <LogoStateProvider>
    <div className="flex h-screen bg-white dark:bg-zinc-950 text-zinc-900 dark:text-zinc-100 transition-colors">
      <aside className="w-56 bg-zinc-50 dark:bg-zinc-900/60 border-r border-zinc-200 dark:border-zinc-800/80 flex flex-col shrink-0 transition-colors">
        <div className="px-5 py-5 border-b border-zinc-200 dark:border-zinc-800/80 flex justify-between items-center transition-colors">
          <div className="flex items-center gap-2.5">
            <SidebarLogo />
            <div className="leading-tight">
              <h1 className="text-[15px] font-semibold text-zinc-900 dark:text-zinc-100 tracking-tight transition-colors">Mantis</h1>
              <p className="text-[11px] text-zinc-500 dark:text-zinc-600 mt-0.5 transition-colors">Control Plane</p>
            </div>
          </div>
          <div className="flex items-center gap-1">
            <ModeToggle />
            <Button
              variant="ghost"
              size="icon"
              onClick={handleLogout}
              className="h-8 w-8 hover:bg-zinc-200 dark:hover:bg-zinc-800"
              title={user ? `Sign out ${user.name}` : 'Sign out'}
            >
              <LogOut className="h-4 w-4" />
              <span className="sr-only">Sign out</span>
            </Button>
          </div>
        </div>

        <div className="flex flex-col flex-1 overflow-hidden">
          <div className="px-2 pt-2">
            <div className="px-3 py-1 text-[10px] font-semibold uppercase tracking-wider text-zinc-500 dark:text-zinc-600 transition-colors">Chat</div>
          </div>
          <div className="flex-1 overflow-auto min-h-0">
            <ChatSidebar
              activeSessionId={activeSessionId}
              onSelect={handleSelectSession}
              onNew={handleNewChat}
              refreshKey={sidebarRefreshKey}
            />
          </div>

          <div className="border-t border-zinc-200 dark:border-zinc-800/80 p-2 space-y-3 overflow-auto transition-colors">
            {nav.map(s => (
              <div key={s.title}>
                <div className="px-3 py-1 text-[10px] font-semibold uppercase tracking-wider text-zinc-500 dark:text-zinc-600 transition-colors">{s.title}</div>
                <div className="space-y-0.5">
                  {renderNav(s.items)}
                </div>
              </div>
            ))}
          </div>
        </div>
      </aside>
      <main className="flex-1 overflow-auto bg-white dark:bg-zinc-950 transition-colors">
        {route.page === 'chat' && <ChatPage sessionId={activeSessionId ?? ''} onFirstMessage={handleFirstMessage} />}
        {route.page === 'channels' && <ChannelsPage />}
        {route.page === 'llm' && <LlmPage />}
        {route.page === 'connections' && <ConnectionsPage />}
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
