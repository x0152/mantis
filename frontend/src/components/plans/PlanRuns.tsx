import { useState, useEffect, useCallback, useMemo } from 'react'
import { Play, Clock, CheckCircle2, XCircle, Loader2, PauseCircle, Circle, SkipForward, ChevronDown, ChevronRight, Ban } from '@/lib/icons'
import { toast } from 'sonner'
import { api } from '../../api'
import type { PlanNode, PlanRun, PlanRunStatus, PlanStepRun, PlanStepStatus, ChatMessage, Step } from '../../types'
import { StepBadge, StepPanel } from '../ChatMessages'
import { Markdown } from '../Markdown'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { EmptyState } from '@/components/EmptyState'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/dialog'
import { FormField } from '@/components/FormField'

const runStatusCfg: Record<PlanRunStatus, { icon: typeof CheckCircle2; color: string; variant: 'success' | 'warning' | 'destructive' | 'muted' }> = {
  running:   { icon: Loader2,     color: 'text-blue-400 animate-spin', variant: 'warning' },
  completed: { icon: CheckCircle2, color: 'text-emerald-400',           variant: 'success' },
  failed:    { icon: XCircle,      color: 'text-red-400',               variant: 'destructive' },
  cancelled: { icon: Ban,          color: 'text-zinc-400',              variant: 'muted' },
  paused:    { icon: PauseCircle,  color: 'text-amber-400',             variant: 'muted' },
}

const stepStatusCfg: Record<PlanStepStatus, { icon: typeof CheckCircle2; color: string; variant: 'success' | 'warning' | 'destructive' | 'muted' | 'default' }> = {
  pending:   { icon: Circle,       color: 'text-zinc-400',              variant: 'muted' },
  running:   { icon: Loader2,      color: 'text-blue-400 animate-spin', variant: 'warning' },
  completed: { icon: CheckCircle2, color: 'text-emerald-400',           variant: 'success' },
  failed:    { icon: XCircle,      color: 'text-red-400',               variant: 'destructive' },
  skipped:   { icon: SkipForward,  color: 'text-zinc-400',              variant: 'muted' },
}

function timeAgo(date: string) {
  const s = Math.floor((Date.now() - new Date(date).getTime()) / 1000)
  if (s < 60) return `${s}s ago`
  if (s < 3600) return `${Math.floor(s / 60)}m ago`
  if (s < 86400) return `${Math.floor(s / 3600)}h ago`
  return `${Math.floor(s / 86400)}d ago`
}

function duration(start?: string, end?: string) {
  if (!start) return '—'
  const ms = (end ? new Date(end).getTime() : Date.now()) - new Date(start).getTime()
  if (ms < 1000) return `${ms}ms`
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`
  return `${Math.floor(ms / 60000)}m ${Math.floor((ms % 60000) / 1000)}s`
}

interface ParamDef {
  name: string
  type: string
  description: string
}

function extractParams(schema: Record<string, unknown>): ParamDef[] {
  const props = (schema?.properties ?? {}) as Record<string, { type?: string; description?: string }>
  return Object.entries(props).map(([name, p]) => ({
    name,
    type: p.type ?? 'string',
    description: p.description ?? '',
  }))
}

interface Props {
  planId: string
  planNodes: PlanNode[]
  planParameters: Record<string, unknown>
  onActiveSteps: (steps: PlanStepRun[] | null) => void
}

export default function PlanRuns({ planId, planNodes, planParameters, onActiveSteps }: Props) {
  const [runs, setRuns] = useState<PlanRun[]>([])
  const [loading, setLoading] = useState(true)
  const [expandedId, setExpandedId] = useState<string | null>(null)
  const [runMessages, setRunMessages] = useState<Record<string, ChatMessage[]>>({})
  const [openStep, setOpenStep] = useState<Step | null>(null)
  const [inputOpen, setInputOpen] = useState(false)
  const [inputValues, setInputValues] = useState<Record<string, string>>({})

  const paramDefs = useMemo(() => extractParams(planParameters), [planParameters])

  const nodeMap = useMemo(() => new Map(planNodes.map(n => [n.id, n])), [planNodes])

  const loadRuns = useCallback(async () => {
    try {
      const data = await api.planRuns.list(planId)
      setRuns(data)
      if (expandedId) {
        const expanded = data.find(r => r.id === expandedId)
        if (expanded) onActiveSteps(expanded.steps)
      }
    } catch {
      // backend may not have the migration yet
    } finally {
      setLoading(false)
    }
  }, [planId, expandedId, onActiveSteps])

  useEffect(() => { loadRuns() }, [loadRuns])

  useEffect(() => {
    const hasRunning = runs.some(r => r.status === 'running')
    const interval = hasRunning ? 2000 : 10000
    const iv = setInterval(loadRuns, interval)
    return () => clearInterval(iv)
  }, [runs, loadRuns])

  const fetchMessages = useCallback(async (runId: string) => {
    try {
      const msgs = await api.chat.listMessages({ sessionId: `plan:${planId}:${runId}` })
      setRunMessages(prev => ({ ...prev, [runId]: msgs }))
    } catch {
      // no messages yet
    }
  }, [planId])

  const toggle = (run: PlanRun) => {
    if (expandedId === run.id) {
      setExpandedId(null)
      onActiveSteps(null)
    } else {
      setExpandedId(run.id)
      onActiveSteps(run.steps)
      if (!runMessages[run.id]) {
        fetchMessages(run.id)
      }
    }
  }

  useEffect(() => {
    if (!expandedId) return
    const run = runs.find(r => r.id === expandedId)
    if (run?.status !== 'running') return
    const iv = setInterval(() => fetchMessages(expandedId), 3000)
    return () => clearInterval(iv)
  }, [expandedId, runs, fetchMessages])

  const cancelRun = async (runId: string) => {
    try {
      await api.planRuns.cancel(runId)
      toast.success('Run cancelled')
      loadRuns()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Failed to cancel run')
    }
  }

  const handleRunClick = () => {
    if (paramDefs.length > 0) {
      const defaults: Record<string, string> = {}
      for (const p of paramDefs) defaults[p.name] = ''
      setInputValues(defaults)
      setInputOpen(true)
    } else {
      doTriggerRun({})
    }
  }

  const doTriggerRun = async (input: Record<string, unknown>) => {
    try {
      const newRun = await api.planRuns.trigger(planId, input)
      toast.success('Plan execution started')
      setExpandedId(newRun.id)
      onActiveSteps(newRun.steps)
      loadRuns()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Failed to trigger run')
    }
  }

  const submitInput = () => {
    const input: Record<string, unknown> = {}
    for (const p of paramDefs) {
      const raw = inputValues[p.name] ?? ''
      if (p.type === 'number' || p.type === 'integer') {
        input[p.name] = Number(raw) || 0
      } else if (p.type === 'boolean') {
        input[p.name] = raw === 'true'
      } else {
        input[p.name] = raw
      }
    }
    setInputOpen(false)
    doTriggerRun(input)
  }

  if (loading) {
    return <div className="p-4 text-center text-sm text-zinc-500">Loading runs...</div>
  }

  return (
    <div className="p-4">
      <div className="flex items-center justify-between mb-3">
        <span className="text-xs font-semibold uppercase tracking-wider text-zinc-500 dark:text-zinc-600">
          Runs{runs.length > 0 && ` (${runs.length})`}
        </span>
        <Button size="sm" onClick={handleRunClick}>
          <Play size={12} /> Run Now
        </Button>
      </div>

      {runs.length === 0 ? (
        <EmptyState icon={Clock} title="No runs yet" description="Run this plan manually or wait for the schedule" />
      ) : (
        <div className="space-y-2">
          {runs.map(run => {
            const cfg = runStatusCfg[run.status] ?? runStatusCfg.running
            const StatusIcon = cfg.icon
            const completed = run.steps.filter(s => s.status === 'completed').length
            const expanded = expandedId === run.id
            const messages = runMessages[run.id] ?? []
            const msgById = new Map(messages.filter(m => m.role === 'assistant').map(m => [m.id, m]))

            return (
              <div key={run.id}>
                <div
                  className={`bg-white dark:bg-zinc-900 rounded-lg border cursor-pointer transition-colors ${
                    expanded
                      ? 'border-teal-500/40 dark:border-teal-500/30'
                      : 'border-zinc-200 dark:border-zinc-800 hover:border-zinc-300 dark:hover:border-zinc-700'
                  }`}
                  onClick={() => toggle(run)}
                >
                  <div className="flex items-center justify-between px-4 py-3">
                    <div className="flex items-center gap-2.5 min-w-0">
                      {expanded ? <ChevronDown size={14} className="text-zinc-400 shrink-0" /> : <ChevronRight size={14} className="text-zinc-400 shrink-0" />}
                      <StatusIcon size={14} className={cfg.color} />
                      <span className="font-mono text-xs text-zinc-500">{run.id.slice(0, 8)}</span>
                      <Badge variant={cfg.variant}>{run.status}</Badge>
                      <Badge variant="muted">{run.trigger}</Badge>
                      {run.input && Object.keys(run.input).length > 0 && (
                        <span className="text-[10px] text-zinc-400 font-mono truncate max-w-[200px]" title={JSON.stringify(run.input)}>
                          {Object.entries(run.input).map(([k, v]) => `${k}=${v}`).join(', ')}
                        </span>
                      )}
                    </div>
                    <div className="flex items-center gap-3 text-[11px] text-zinc-500 dark:text-zinc-600 shrink-0 ml-3">
                      <span className="flex items-center gap-1"><Clock size={10} />{timeAgo(run.startedAt)}</span>
                      <span>{duration(run.startedAt, run.finishedAt)}</span>
                      <span>{completed}/{run.steps.length}</span>
                      {run.status === 'running' && (
                        <Button
                          variant="destructive"
                          size="sm"
                          className="h-6 px-2 text-[10px]"
                          onClick={e => { e.stopPropagation(); cancelRun(run.id) }}
                        >
                          <Ban size={10} /> Cancel
                        </Button>
                      )}
                    </div>
                  </div>

                  <div className="flex gap-1 px-4 pb-3">
                    {run.steps.map(step => (
                      <div
                        key={step.nodeId}
                        className={`h-1.5 flex-1 rounded-full ${
                          step.status === 'completed' ? 'bg-emerald-500' :
                          step.status === 'running' ? 'bg-blue-500 animate-pulse' :
                          step.status === 'failed' ? 'bg-red-500' :
                          step.status === 'skipped' ? 'bg-zinc-400 dark:bg-zinc-600' :
                          'bg-zinc-200 dark:bg-zinc-800'
                        }`}
                      />
                    ))}
                  </div>
                </div>

                {expanded && (
                  <div className="mt-1 ml-4 pl-4 border-l-2 border-zinc-200 dark:border-zinc-800 space-y-1 py-2">
                    {run.steps.map(step => (
                      <StepRunRow
                        key={step.nodeId}
                        step={step}
                        node={nodeMap.get(step.nodeId)}
                        message={step.messageId ? msgById.get(step.messageId) : undefined}
                        onStepClick={setOpenStep}
                      />
                    ))}
                  </div>
                )}
              </div>
            )
          })}
        </div>
      )}

      {openStep && <StepPanel step={openStep} onClose={() => setOpenStep(null)} />}

      <Dialog open={inputOpen} onOpenChange={setInputOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Run with parameters</DialogTitle>
            <DialogDescription>Fill in the required inputs for this plan.</DialogDescription>
          </DialogHeader>
          <div className="space-y-3">
            {paramDefs.map(p => (
              <FormField key={p.name} label={p.name} hint={p.description}>
                <Input
                  value={inputValues[p.name] ?? ''}
                  onChange={e => setInputValues(v => ({ ...v, [p.name]: e.target.value }))}
                  placeholder={p.type}
                  className="font-mono text-xs"
                />
              </FormField>
            ))}
          </div>
          <DialogFooter>
            <Button variant="secondary" onClick={() => setInputOpen(false)}>Cancel</Button>
            <Button onClick={submitInput}>
              <Play size={12} /> Run
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

function StepRunRow({ step, node, message, onStepClick }: {
  step: PlanStepRun
  node?: PlanNode
  message?: ChatMessage
  onStepClick: (s: Step) => void
}) {
  const sc = stepStatusCfg[step.status] ?? stepStatusCfg.pending
  const StepIcon = sc.icon
  const toolSteps = message?.steps ?? []

  return (
    <div className="py-2">
      <div className="flex items-start gap-3">
        <StepIcon size={14} className={`${sc.color} mt-0.5 shrink-0`} />
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <span className="text-sm font-medium text-zinc-800 dark:text-zinc-200">{node?.label || step.nodeId}</span>
            <span className="text-[10px] text-zinc-500 capitalize">{node?.type}</span>
            <Badge variant={sc.variant} className="text-[10px]">{step.status}</Badge>
            {step.startedAt && (
              <span className="text-[11px] text-zinc-500">{duration(step.startedAt, step.finishedAt)}</span>
            )}
          </div>

          {toolSteps.length > 0 && (
            <div className="flex flex-wrap gap-1.5 mt-2">
              {toolSteps.map(ts => (
                <StepBadge key={ts.id} step={ts} onClick={() => onStepClick(ts)} />
              ))}
            </div>
          )}

          {message?.content && (
            <div className="mt-2 px-3 py-2 rounded-md bg-zinc-100 dark:bg-zinc-800/50 text-xs text-zinc-700 dark:text-zinc-300">
              <Markdown content={message.content} />
            </div>
          )}

          {step.result && step.status === 'failed' && (
            <div className="mt-2 px-3 py-2 rounded-md text-xs bg-red-500/5 border border-red-500/10 text-red-400">
              {step.result}
            </div>
          )}

          {step.status === 'running' && !message && (
            <div className="flex items-center gap-1.5 mt-2 text-xs text-zinc-500">
              <Loader2 size={11} className="animate-spin" />
              <span>Running...</span>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
