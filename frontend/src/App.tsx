import { useState, useEffect } from 'react'
import { Settings, Plug, Clock, MessageSquare, ScrollText, Link2, Radio, ShieldAlert } from 'lucide-react'
import ConfigPage from './pages/ConfigPage'
import LlmPage from './pages/LlmPage'
import ConnectionsPage from './pages/ConnectionsPage'
import ChannelsPage from './pages/ChannelsPage'
import CronJobsPage from './pages/CronJobsPage'
import ChatPage from './pages/ChatPage'
import LogsPage from './pages/LogsPage'
import GuardProfilesPage from './pages/GuardProfilesPage'
import SetupWizard from './pages/SetupWizard'
import { api } from './api'

type Page = 'chat' | 'channels' | 'connections' | 'logs' | 'llm' | 'cron-jobs' | 'guard-profiles' | 'config'

type NavItem = { id: Page; label: string; icon: typeof Settings }

type NavSection = { title: string; items: NavItem[] }

const nav: NavSection[] = [
  { title: 'Chat', items: [{ id: 'chat', label: 'Chat', icon: MessageSquare }] },
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
      { id: 'channels', label: 'Channels', icon: Radio },
      { id: 'connections', label: 'Servers', icon: Plug },
    ],
  },
  {
    title: 'System',
    items: [
      { id: 'guard-profiles', label: 'Guard Profiles', icon: ShieldAlert },
      { id: 'config', label: 'Configuration', icon: Settings },
    ],
  },
]

export default function App() {
  const [page, setPage] = useState<Page>('chat')
  const [needsSetup, setNeedsSetup] = useState<boolean | null>(null)

  useEffect(() => {
    api.llmConnections.list()
      .then(list => setNeedsSetup(list.length === 0))
      .catch(() => setNeedsSetup(false))
  }, [])

  if (needsSetup === null) return null
  if (needsSetup) return <SetupWizard onDone={() => setNeedsSetup(false)} />

  const renderNav = (items: NavItem[]) => items.map(item => (
    <button
      key={item.id}
      onClick={() => setPage(item.id)}
      className={`w-full flex items-center gap-2.5 px-3 py-2 rounded-lg text-[13px] font-medium min-w-0 ${
        page === item.id
          ? 'bg-teal-500/10 text-teal-400'
          : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/50'
      }`}
    >
      <item.icon size={16} strokeWidth={1.8} />
      <span className="truncate" title={item.label}>{item.label}</span>
    </button>
  ))

  return (
    <div className="flex h-screen bg-zinc-950 text-zinc-100">
      <aside className="w-56 bg-zinc-900/60 border-r border-zinc-800/80 flex flex-col shrink-0">
        <div className="px-5 py-5 border-b border-zinc-800/80">
          <div className="flex items-center gap-2.5">
            <div className="w-2 h-2 rounded-full bg-teal-400" />
            <h1 className="text-[15px] font-semibold text-zinc-100 tracking-tight">Mantis</h1>
          </div>
          <p className="text-[11px] text-zinc-600 mt-0.5 ml-[18px]">Control Plane</p>
        </div>
        <nav className="flex-1 p-2 overflow-auto">
          <div className="space-y-3">
            {nav.map(s => (
              <div key={s.title}>
                <div className="px-3 py-1 text-[10px] font-semibold uppercase tracking-wider text-zinc-600">{s.title}</div>
                <div className="space-y-0.5">
                  {renderNav(s.items)}
                </div>
              </div>
            ))}
          </div>
        </nav>
      </aside>
      <main className="flex-1 overflow-auto">
        {page === 'chat' && <ChatPage />}
        {page === 'channels' && <ChannelsPage />}
        {page === 'config' && <ConfigPage />}
        {page === 'llm' && <LlmPage />}
        {page === 'connections' && <ConnectionsPage />}
        {page === 'logs' && <LogsPage />}
        {page === 'cron-jobs' && <CronJobsPage />}
        {page === 'guard-profiles' && <GuardProfilesPage />}
      </main>
    </div>
  )
}
