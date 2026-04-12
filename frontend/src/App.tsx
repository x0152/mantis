import { useState, useEffect, useCallback } from 'react'
import { Settings, Plug, Clock, ScrollText, Link2, Radio, ShieldAlert, Layers } from 'lucide-react'
import { Toaster } from '@/components/ui/sonner'
import LlmPage from './pages/LlmPage'
import ConnectionsPage from './pages/ConnectionsPage'
import ChannelsPage from './pages/ChannelsPage'
import CronJobsPage from './pages/CronJobsPage'
import ChatPage from './pages/ChatPage'
import LogsPage from './pages/LogsPage'
import GuardProfilesPage from './pages/GuardProfilesPage'
import PresetsPage from './pages/PresetsPage'
import SetupWizard from './pages/SetupWizard'
import ChatSidebar from './components/ChatSidebar'
import { ModeToggle } from './components/mode-toggle'
import { api } from './api'
import type { ChatSession } from './types'

type Page = 'chat' | 'channels' | 'connections' | 'logs' | 'llm' | 'presets' | 'cron-jobs' | 'guard-profiles'

type NavItem = { id: Page; label: string; icon: typeof Settings }

type NavSection = { title: string; items: NavItem[] }

const nav: NavSection[] = [
  {
    title: 'Activity',
    items: [
      { id: 'cron-jobs', label: 'Cron Jobs', icon: Clock },
      { id: 'logs', label: 'Logs', icon: ScrollText },
    ],
  },
  {
    title: 'Connections',
    items: [
      { id: 'llm', label: 'LLMs & Models', icon: Link2 },
      { id: 'presets', label: 'Presets', icon: Layers },
      { id: 'channels', label: 'Channels', icon: Radio },
      { id: 'connections', label: 'Servers', icon: Plug },
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
  const [page, setPage] = useState<Page>('chat')
  const [needsSetup, setNeedsSetup] = useState<boolean | null>(null)
  const [activeSessionId, setActiveSessionId] = useState<string | null>(null)
  const [sidebarRefreshKey, setSidebarRefreshKey] = useState(0)

  useEffect(() => {
    api.llmConnections.list()
      .then(list => setNeedsSetup(list.length === 0))
      .catch(() => setNeedsSetup(false))
  }, [])

  useEffect(() => {
    if (needsSetup === false && !activeSessionId) {
      api.chat.getSession().then(s => setActiveSessionId(s.id)).catch(() => {})
    }
  }, [needsSetup])

  const handleSelectSession = useCallback((session: ChatSession) => {
    setActiveSessionId(session.id)
    setPage('chat')
  }, [])

  const handleNewChat = useCallback(async () => {
    try {
      const session = await api.chat.createSession()
      setActiveSessionId(session.id)
      setPage('chat')
      setSidebarRefreshKey(k => k + 1)
    } catch {}
  }, [])

  const handleFirstMessage = useCallback(() => {
    setSidebarRefreshKey(k => k + 1)
  }, [])

  if (needsSetup === null) return null
  if (needsSetup) return <SetupWizard onDone={() => setNeedsSetup(false)} />

  const renderNav = (items: NavItem[]) => items.map(item => (
    <button
      key={item.id}
      onClick={() => setPage(item.id)}
      className={`w-full flex items-center gap-2.5 px-3 py-2 rounded-lg text-[13px] font-medium min-w-0 ${
        page === item.id
          ? 'bg-teal-500/10 text-teal-600 dark:text-teal-400'
          : 'text-zinc-600 dark:text-zinc-400 hover:text-zinc-900 dark:hover:text-zinc-200 hover:bg-zinc-100 dark:hover:bg-zinc-800/50'
      }`}
    >
      <item.icon size={16} strokeWidth={1.8} />
      <span className="truncate" title={item.label}>{item.label}</span>
    </button>
  ))

  return (
    <div className="flex h-screen bg-white dark:bg-zinc-950 text-zinc-900 dark:text-zinc-100 transition-colors">
      <aside className="w-56 bg-zinc-50 dark:bg-zinc-900/60 border-r border-zinc-200 dark:border-zinc-800/80 flex flex-col shrink-0 transition-colors">
        <div className="px-5 py-5 border-b border-zinc-200 dark:border-zinc-800/80 flex justify-between items-center transition-colors">
          <div>
            <div className="flex items-center gap-2.5">
              <div className="w-2 h-2 rounded-full bg-teal-400" />
              <h1 className="text-[15px] font-semibold text-zinc-900 dark:text-zinc-100 tracking-tight transition-colors">Mantis</h1>
            </div>
            <p className="text-[11px] text-zinc-500 dark:text-zinc-600 mt-0.5 ml-4.5 transition-colors">Control Plane</p>
          </div>
          <ModeToggle />
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
        {page === 'chat' && <ChatPage sessionId={activeSessionId ?? ''} onFirstMessage={handleFirstMessage} />}
        {page === 'channels' && <ChannelsPage />}
        {page === 'llm' && <LlmPage />}
        {page === 'presets' && <PresetsPage />}
        {page === 'connections' && <ConnectionsPage />}
        {page === 'logs' && <LogsPage />}
        {page === 'cron-jobs' && <CronJobsPage />}
        {page === 'guard-profiles' && <GuardProfilesPage />}
      </main>
      <Toaster />
    </div>
  )
}
