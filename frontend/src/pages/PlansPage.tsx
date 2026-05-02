import { useState, useEffect, useCallback } from 'react'
import { Plus, GitBranch, Pencil, Trash2, Play, Pause } from '@/lib/icons'
import { toast } from 'sonner'
import { api } from '../api'
import type { Plan } from '../types'
import PlanEditor from '../components/plans/PlanEditor'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { EmptyState } from '@/components/EmptyState'
import { ConfirmDelete } from '@/components/ConfirmDelete'
import { navigate } from '../router'

const emptyPlan: Plan = {
  id: '',
  name: '',
  description: '',
  schedule: '',
  enabled: false,
  parameters: {},
  graph: { nodes: [], edges: [] },
}

export default function PlansPage({ deepPlanId }: { deepPlanId?: string }) {
  const [plans, setPlans] = useState<Plan[]>([])
  const [loading, setLoading] = useState(true)
  const [activePlan, setActivePlan] = useState<Plan | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null)
  const load = useCallback(async () => {
    try {
      setLoading(true)
      const list = await api.plans.list()
      setPlans(list)
      return list
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Failed to load')
      return []
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])

  useEffect(() => {
    if (!deepPlanId || loading) return
    const found = plans.find(p => p.id === deepPlanId)
    if (found) setActivePlan(found)
  }, [deepPlanId, loading, plans])

  const toggleEnabled = async (plan: Plan, e: React.MouseEvent) => {
    e.stopPropagation()
    try {
      await api.plans.update(plan.id, {
        name: plan.name,
        description: plan.description,
        schedule: plan.schedule,
        enabled: !plan.enabled,
        parameters: plan.parameters ?? {},
        graph: plan.graph,
      })
      toast.success(plan.enabled ? 'Plan disabled' : 'Plan enabled')
      load()
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'Toggle failed')
    }
  }

  const remove = async (id: string) => {
    try {
      await api.plans.delete(id)
      toast.success('Plan deleted')
      setDeleteTarget(null)
      load()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Delete failed')
    }
  }

  if (activePlan) {
    return (
      <PlanEditor
        plan={activePlan}
        onBack={() => {
          setActivePlan(null)
          if (deepPlanId) navigate({ page: 'plans' })
          load()
        }}
        onSaved={load}
      />
    )
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-5">
        <div>
          <h1 className="text-lg font-semibold text-zinc-900 dark:text-zinc-100">Plans</h1>
          <p className="text-xs text-zinc-500 dark:text-zinc-600 mt-0.5">Agentic workflows with actions and decisions</p>
        </div>
        <Button size="sm" onClick={() => setActivePlan(emptyPlan)}>
          <Plus size={14} /> New Plan
        </Button>
      </div>

      {loading ? (
        <div className="text-center py-12 text-zinc-500 dark:text-zinc-600 text-sm">Loading...</div>
      ) : plans.length === 0 ? (
        <EmptyState icon={GitBranch} title="No plans yet" description="Create your first agentic workflow" />
      ) : (
        <div className="space-y-2.5">
          {plans.map(plan => (
            <div
              key={plan.id}
              className={`bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800 px-4 py-3.5 cursor-pointer hover:border-zinc-300 dark:hover:border-zinc-700 transition-colors ${!plan.enabled ? 'opacity-60' : ''}`}
              onClick={() => setActivePlan(plan)}
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2 min-w-0">
                  <GitBranch size={16} className="text-teal-500 shrink-0" />
                  <span className="font-medium text-zinc-800 dark:text-zinc-200 text-sm">{plan.name}</span>
                  {plan.schedule && <Badge variant="outline" className="font-mono">{plan.schedule}</Badge>}
                  <Badge variant={plan.enabled ? 'success' : 'muted'}>
                    {plan.enabled ? 'Active' : 'Disabled'}
                  </Badge>
                  <Badge variant="secondary">
                    {plan.graph.nodes.length} nodes
                  </Badge>
                </div>
                <div className="flex gap-0.5 ml-3" onClick={e => e.stopPropagation()}>
                  <Button variant="ghost" size="icon" onClick={e => toggleEnabled(plan, e)}>
                    {plan.enabled ? <Pause size={14} /> : <Play size={14} />}
                  </Button>
                  <Button variant="ghost" size="icon" onClick={() => setActivePlan(plan)}>
                    <Pencil size={14} />
                  </Button>
                  <Button variant="destructive" size="icon" onClick={() => setDeleteTarget(plan.id)}>
                    <Trash2 size={14} />
                  </Button>
                </div>
              </div>
              {plan.description && (
                <p className="text-xs text-zinc-500 dark:text-zinc-500 mt-1.5 ml-6">{plan.description}</p>
              )}
            </div>
          ))}
        </div>
      )}

      <ConfirmDelete
        open={!!deleteTarget}
        onCancel={() => setDeleteTarget(null)}
        onConfirm={() => deleteTarget && remove(deleteTarget)}
        title="Delete plan?"
        description="This will permanently remove the plan and its graph."
      />
    </div>
  )
}
