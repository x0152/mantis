import { useState, useEffect, useCallback } from 'react'
import { Plus, Pencil, Trash2, Clock, AlertTriangle, Check, Play, Pause, ChevronDown, ChevronRight, MessageSquare } from 'lucide-react'
import { api } from '../api'
import { MessageBubble, StepPanel } from '../components/ChatMessages'
import type { CronJob, ChatMessage, Step } from '../types'
import Modal from '../components/Modal'

export default function CronJobsPage() {
  const [jobs, setJobs] = useState<CronJob[]>([])
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<CronJob | null>(null)
  const [form, setForm] = useState({ name: '', schedule: '', prompt: '', enabled: true })
  const [toast, setToast] = useState<{ type: 'success' | 'error'; text: string } | null>(null)
  const [expandedJobId, setExpandedJobId] = useState<string | null>(null)

  const showToast = (type: 'success' | 'error', text: string) => {
    setToast({ type, text })
    setTimeout(() => setToast(null), 3000)
  }

  const load = useCallback(async () => {
    try {
      setLoading(true)
      const j = await api.cronJobs.list()
      setJobs(j)
    } catch (e: unknown) {
      showToast('error', e instanceof Error ? e.message : 'Failed to load')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])

  const openCreate = () => {
    setEditing(null)
    setForm({ name: '', schedule: '', prompt: '', enabled: true })
    setModalOpen(true)
  }

  const openEdit = (j: CronJob) => {
    setEditing(j)
    setForm({ name: j.name, schedule: j.schedule, prompt: j.prompt, enabled: j.enabled })
    setModalOpen(true)
  }

  const submit = async () => {
    try {
      if (editing) {
        await api.cronJobs.update(editing.id, form)
        showToast('success', 'Cron job updated')
      } else {
        await api.cronJobs.create(form)
        showToast('success', 'Cron job created')
      }
      setModalOpen(false)
      load()
    } catch (e: unknown) {
      showToast('error', e instanceof Error ? e.message : 'Operation failed')
    }
  }

  const toggleEnabled = async (j: CronJob) => {
    try {
      await api.cronJobs.update(j.id, { name: j.name, schedule: j.schedule, prompt: j.prompt, enabled: !j.enabled })
      showToast('success', j.enabled ? 'Job paused' : 'Job enabled')
      load()
    } catch (e: unknown) {
      showToast('error', e instanceof Error ? e.message : 'Toggle failed')
    }
  }

  const remove = async (id: string) => {
    if (!confirm('Delete this cron job?')) return
    try {
      await api.cronJobs.delete(id)
      showToast('success', 'Cron job deleted')
      load()
    } catch (e: unknown) {
      showToast('error', e instanceof Error ? e.message : 'Delete failed')
    }
  }

  return (
    <div className="p-6 max-w-5xl">
      <div className="flex items-center justify-between mb-5">
        <div>
          <h1 className="text-lg font-semibold text-zinc-100">Cron Jobs</h1>
          <p className="text-xs text-zinc-600 mt-0.5">Scheduled tasks with prompt payloads</p>
        </div>
        <button onClick={openCreate} className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-white bg-teal-600 rounded-lg hover:bg-teal-500">
          <Plus size={14} /> Add Cron Job
        </button>
      </div>

      {toast && (
        <div className={`mb-4 flex items-center gap-2 px-3.5 py-2.5 rounded-lg text-xs font-medium ${
          toast.type === 'success' ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-red-500/10 text-red-400 border border-red-500/20'
        }`}>
          {toast.type === 'success' ? <Check size={14} /> : <AlertTriangle size={14} />}
          {toast.text}
        </div>
      )}

      {loading ? (
        <div className="text-center py-12 text-zinc-600 text-sm">Loading...</div>
      ) : jobs.length === 0 ? (
        <div className="text-center py-20">
          <Clock size={40} className="mx-auto text-zinc-700 mb-3" />
          <p className="text-zinc-400 text-sm font-medium">No cron jobs yet</p>
          <p className="text-xs text-zinc-600 mt-1">Create your first scheduled task</p>
        </div>
      ) : (
        <div className="space-y-2.5">
          {jobs.map(j => (
            <div key={j.id} className={`bg-zinc-900 rounded-lg border border-zinc-800 overflow-hidden ${!j.enabled ? 'opacity-50' : ''}`}>
              <div className="px-4 py-3.5">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2 min-w-0">
                    <span className="font-medium text-zinc-200 text-sm">{j.name}</span>
                    <code className="px-1.5 py-0.5 text-[11px] font-mono rounded-md bg-amber-500/10 text-amber-400 border border-amber-500/20">{j.schedule}</code>
                    <span className={`px-1.5 py-0.5 text-[11px] font-medium rounded-full ${j.enabled ? 'bg-emerald-500/15 text-emerald-400' : 'bg-zinc-800 text-zinc-600'}`}>
                      {j.enabled ? 'Active' : 'Paused'}
                    </span>
                  </div>
                  <div className="flex gap-0.5 ml-3">
                    <button
                      onClick={() => setExpandedJobId(expandedJobId === j.id ? null : j.id)}
                      className="p-1.5 rounded-md text-zinc-600 hover:text-zinc-300 hover:bg-zinc-800"
                      title="View messages"
                    >
                      <MessageSquare size={14} />
                    </button>
                    <button onClick={() => toggleEnabled(j)} className={`p-1.5 rounded-md ${j.enabled ? 'text-zinc-600 hover:text-amber-400 hover:bg-amber-500/10' : 'text-zinc-600 hover:text-emerald-400 hover:bg-emerald-500/10'}`}>
                      {j.enabled ? <Pause size={14} /> : <Play size={14} />}
                    </button>
                    <button onClick={() => openEdit(j)} className="p-1.5 rounded-md text-zinc-600 hover:text-teal-400 hover:bg-teal-500/10">
                      <Pencil size={14} />
                    </button>
                    <button onClick={() => remove(j.id)} className="p-1.5 rounded-md text-zinc-600 hover:text-red-400 hover:bg-red-500/10">
                      <Trash2 size={14} />
                    </button>
                  </div>
                </div>
                <div className="mt-2.5 bg-zinc-950 rounded-lg border border-zinc-800 p-3">
                  <div
                    className="flex items-center gap-1 cursor-pointer select-none"
                    onClick={() => setExpandedJobId(expandedJobId === j.id ? null : j.id)}
                  >
                    {expandedJobId === j.id ? <ChevronDown size={12} className="text-zinc-600" /> : <ChevronRight size={12} className="text-zinc-600" />}
                    <p className="text-[11px] font-semibold text-zinc-600 uppercase tracking-wider">Prompt</p>
                  </div>
                  <p className="text-sm text-zinc-400 whitespace-pre-wrap mt-1">{j.prompt}</p>
                </div>
              </div>
              {expandedJobId === j.id && (
                <CronMessages jobId={j.id} />
              )}
            </div>
          ))}
        </div>
      )}

      <Modal open={modalOpen} onClose={() => setModalOpen(false)} title={editing ? 'Edit Cron Job' : 'New Cron Job'}>
        <div className="space-y-4">
          <div>
            <label className="block text-xs font-medium text-zinc-400 mb-1.5">Name</label>
            <input
              value={form.name}
              onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
              className="w-full px-3 py-2 border border-zinc-700 rounded-lg text-sm bg-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-teal-500/50"
              placeholder="Daily health check"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-zinc-400 mb-1.5">Schedule (cron)</label>
            <input
              value={form.schedule}
              onChange={e => setForm(f => ({ ...f, schedule: e.target.value }))}
              className="w-full px-3 py-2 border border-zinc-700 rounded-lg text-sm font-mono bg-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-teal-500/50"
              placeholder="0 * * * *"
            />
            <p className="text-[11px] text-zinc-600 mt-1">min hour day month weekday</p>
          </div>
          <div>
            <label className="block text-xs font-medium text-zinc-400 mb-1.5">Prompt</label>
            <textarea
              value={form.prompt}
              onChange={e => setForm(f => ({ ...f, prompt: e.target.value }))}
              className="w-full h-32 px-3 py-2 border border-zinc-700 rounded-lg text-sm bg-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-teal-500/50 resize-none"
              placeholder="Check disk usage and report if any partition exceeds 90%..."
            />
          </div>
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => setForm(f => ({ ...f, enabled: !f.enabled }))}
              className={`relative inline-flex h-5 w-9 items-center rounded-full ${form.enabled ? 'bg-teal-600' : 'bg-zinc-700'}`}
            >
              <span className={`inline-block h-3.5 w-3.5 rounded-full bg-white ${form.enabled ? 'translate-x-[18px]' : 'translate-x-[3px]'}`} />
            </button>
            <span className="text-xs text-zinc-400">{form.enabled ? 'Enabled' : 'Disabled'}</span>
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <button onClick={() => setModalOpen(false)} className="px-3.5 py-2 text-xs font-medium text-zinc-400 bg-zinc-800 border border-zinc-700 rounded-lg hover:text-zinc-200 hover:bg-zinc-700">
              Cancel
            </button>
            <button onClick={submit} disabled={!form.name || !form.schedule || !form.prompt} className="px-3.5 py-2 text-xs font-medium text-white bg-teal-600 rounded-lg hover:bg-teal-500 disabled:opacity-40">
              {editing ? 'Update' : 'Create'}
            </button>
          </div>
        </div>
      </Modal>
    </div>
  )
}

function CronMessages({ jobId }: { jobId: string }) {
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [loading, setLoading] = useState(true)
  const [activeStep, setActiveStep] = useState<Step | null>(null)
  const [hasMore, setHasMore] = useState(false)
  const [loadingMore, setLoadingMore] = useState(false)

  const PAGE_SIZE = 20

  useEffect(() => {
    loadMessages()
  }, [jobId])

  async function loadMessages() {
    setLoading(true)
    try {
      const msgs = await api.chat.listMessages({ source: 'cron', sessionId: `cron:${jobId}`, limit: PAGE_SIZE, offset: 0 })
      setMessages(msgs)
      setHasMore(msgs.length === PAGE_SIZE)
    } catch {}
    setLoading(false)
  }

  async function loadMore() {
    if (loadingMore || !hasMore) return
    setLoadingMore(true)
    try {
      const older = await api.chat.listMessages({ source: 'cron', sessionId: `cron:${jobId}`, limit: PAGE_SIZE, offset: messages.length })
      if (older.length < PAGE_SIZE) setHasMore(false)
      if (older.length > 0) {
        setMessages(prev => {
          const byId = new Map<string, ChatMessage>()
          for (const m of older) byId.set(m.id, m)
          for (const m of prev) byId.set(m.id, m)
          return Array.from(byId.values()).sort((a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime())
        })
      }
    } catch {
      setHasMore(false)
    } finally {
      setLoadingMore(false)
    }
  }

  return (
    <div className="border-t border-zinc-800">
      <div className="px-4 py-3 bg-zinc-950/50">
        <p className="text-[11px] font-semibold text-zinc-600 uppercase tracking-wider mb-3">Execution History</p>
        {loading ? (
          <div className="text-xs text-zinc-600 py-4 text-center">Loading messages...</div>
        ) : messages.length === 0 ? (
          <div className="text-xs text-zinc-600 py-4 text-center">No executions yet</div>
        ) : (
          <div className="space-y-3">
            {hasMore && (
              <div className="flex justify-center mb-2">
                <button
                  onClick={loadMore}
                  disabled={loadingMore}
                  className="px-3 py-1 text-[11px] font-medium text-zinc-500 bg-zinc-800 rounded-md hover:text-zinc-300 hover:bg-zinc-700 disabled:opacity-30"
                >
                  {loadingMore ? 'Loading...' : 'Load older'}
                </button>
              </div>
            )}
            {messages.map(msg => (
              <MessageBubble key={msg.id} msg={msg} onStepClick={setActiveStep} />
            ))}
          </div>
        )}
      </div>
      {activeStep && <StepPanel step={activeStep} onClose={() => setActiveStep(null)} />}
    </div>
  )
}
