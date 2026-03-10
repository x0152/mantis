import { useState, useEffect, useCallback } from 'react'
import { Plus, Pencil, Trash2, Clock, Play, Pause, ChevronDown, ChevronRight, MessageSquare } from 'lucide-react'
import { toast } from 'sonner'
import { api } from '../api'
import { MessageBubble, StepPanel } from '../components/ChatMessages'
import type { CronJob, ChatMessage, Step } from '../types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/dialog'
import { EmptyState } from '@/components/EmptyState'
import { FormField } from '@/components/FormField'
import { ConfirmDelete } from '@/components/ConfirmDelete'

export default function CronJobsPage() {
  const [jobs, setJobs] = useState<CronJob[]>([])
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<CronJob | null>(null)
  const [form, setForm] = useState({ name: '', schedule: '', prompt: '', enabled: true })
  const [expandedJobId, setExpandedJobId] = useState<string | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null)

  const load = useCallback(async () => {
    try {
      setLoading(true)
      const j = await api.cronJobs.list()
      setJobs(j)
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Failed to load')
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
        toast.success('Cron job updated')
      } else {
        await api.cronJobs.create(form)
        toast.success('Cron job created')
      }
      setModalOpen(false)
      load()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Operation failed')
    }
  }

  const toggleEnabled = async (j: CronJob) => {
    try {
      await api.cronJobs.update(j.id, { name: j.name, schedule: j.schedule, prompt: j.prompt, enabled: !j.enabled })
      toast.success(j.enabled ? 'Job paused' : 'Job enabled')
      load()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Toggle failed')
    }
  }

  const remove = async (id: string) => {
    try {
      await api.cronJobs.delete(id)
      toast.success('Cron job deleted')
      setDeleteTarget(null)
      load()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Delete failed')
    }
  }

  return (
    <div className="p-6 max-w-5xl">
      <div className="flex items-center justify-between mb-5">
        <div>
          <h1 className="text-lg font-semibold text-zinc-100">Cron Jobs</h1>
          <p className="text-xs text-zinc-600 mt-0.5">Scheduled tasks with prompt payloads</p>
        </div>
        <Button size="sm" onClick={openCreate}>
          <Plus size={14} /> Add Cron Job
        </Button>
      </div>

      {loading ? (
        <div className="text-center py-12 text-zinc-600 text-sm">Loading...</div>
      ) : jobs.length === 0 ? (
        <EmptyState icon={Clock} title="No cron jobs yet" description="Create your first scheduled task" />
      ) : (
        <div className="space-y-2.5">
          {jobs.map(j => (
            <div key={j.id} className={`bg-zinc-900 rounded-lg border border-zinc-800 overflow-hidden ${!j.enabled ? 'opacity-50' : ''}`}>
              <div className="px-4 py-3.5">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2 min-w-0">
                    <span className="font-medium text-zinc-200 text-sm">{j.name}</span>
                    <Badge variant="outline">{j.schedule}</Badge>
                    <Badge variant={j.enabled ? 'success' : 'muted'}>
                      {j.enabled ? 'Active' : 'Paused'}
                    </Badge>
                  </div>
                  <div className="flex gap-0.5 ml-3">
                    <Button
                      variant="ghost" size="icon"
                      onClick={() => setExpandedJobId(expandedJobId === j.id ? null : j.id)}
                      title="View messages"
                    >
                      <MessageSquare size={14} />
                    </Button>
                    <Button variant="ghost" size="icon" onClick={() => toggleEnabled(j)}>
                      {j.enabled ? <Pause size={14} /> : <Play size={14} />}
                    </Button>
                    <Button variant="ghost" size="icon" onClick={() => openEdit(j)}>
                      <Pencil size={14} />
                    </Button>
                    <Button variant="destructive" size="icon" onClick={() => setDeleteTarget(j.id)}>
                      <Trash2 size={14} />
                    </Button>
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

      <Dialog open={modalOpen} onOpenChange={setModalOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing ? 'Edit Cron Job' : 'New Cron Job'}</DialogTitle>
            <DialogDescription />
          </DialogHeader>
          <div className="space-y-4">
            <FormField label="Name">
              <Input
                value={form.name}
                onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                placeholder="Daily health check"
              />
            </FormField>
            <FormField label="Schedule (cron)">
              <Input
                value={form.schedule}
                onChange={e => setForm(f => ({ ...f, schedule: e.target.value }))}
                className="font-mono"
                placeholder="0 * * * *"
              />
              <p className="text-[11px] text-zinc-600 mt-1">min hour day month weekday</p>
            </FormField>
            <FormField label="Prompt">
              <Textarea
                value={form.prompt}
                onChange={e => setForm(f => ({ ...f, prompt: e.target.value }))}
                className="h-32"
                placeholder="Check disk usage and report if any partition exceeds 90%..."
              />
            </FormField>
            <div className="flex items-center gap-2">
              <Switch checked={form.enabled} onCheckedChange={v => setForm(f => ({ ...f, enabled: v }))} />
              <span className="text-xs text-zinc-400">{form.enabled ? 'Enabled' : 'Disabled'}</span>
            </div>
            <DialogFooter>
              <Button variant="secondary" onClick={() => setModalOpen(false)}>Cancel</Button>
              <Button onClick={submit} disabled={!form.name || !form.schedule || !form.prompt}>
                {editing ? 'Update' : 'Create'}
              </Button>
            </DialogFooter>
          </div>
        </DialogContent>
      </Dialog>

      <ConfirmDelete
        open={!!deleteTarget}
        onCancel={() => setDeleteTarget(null)}
        onConfirm={() => deleteTarget && remove(deleteTarget)}
        title="Delete cron job?"
        description="This will permanently remove the scheduled task."
      />
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
                <Button variant="secondary" size="sm" onClick={loadMore} disabled={loadingMore}>
                  {loadingMore ? 'Loading...' : 'Load older'}
                </Button>
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
