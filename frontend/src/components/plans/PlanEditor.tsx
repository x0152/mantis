import { useState, useCallback, useMemo, useRef } from 'react'
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  addEdge,
  useNodesState,
  useEdgesState,
  MarkerType,
  type Connection,
  type Node,
  type Edge,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import { ArrowLeft, Pencil, Zap, GitFork, Trash2, Save, Pause, Play } from '@/lib/icons'
import { toast } from 'sonner'
import { api } from '../../api'
import type { Plan, PlanGraph, PlanStepRun, PlanStepStatus } from '../../types'
import { describeCron } from '../../lib/cron'
import { planNodeTypes, toFlowNodes, toFlowEdges, fromFlowNodes, fromFlowEdges } from './PlanFlowNodes'
import PlanRuns from './PlanRuns'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/dialog'
import { FormField } from '@/components/FormField'
import { ParameterEditor } from '@/components/ParameterEditor'
import { ParamButtons, insertAtCursor } from '@/components/ParamButtons'

let nodeIdCounter = 0

interface Props {
  plan: Plan
  onBack: () => void
  onSaved: () => void
}

export default function PlanEditor({ plan: initialPlan, onBack, onSaved }: Props) {
  const [plan, setPlan] = useState(initialPlan)
  const [metaOpen, setMetaOpen] = useState(!initialPlan.id)
  const [metaForm, setMetaForm] = useState({
    name: initialPlan.name,
    description: initialPlan.description,
    schedule: initialPlan.schedule,
    enabled: initialPlan.enabled,
    parameters: initialPlan.parameters ?? { type: 'object', properties: {} },
  })

  const [nodes, setNodes, onNodesChange] = useNodesState<Node>(toFlowNodes(initialPlan.graph.nodes))
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>(toFlowEdges(initialPlan.graph.edges))
  const [selectedNode, setSelectedNode] = useState<Node | null>(null)
  const [nodeForm, setNodeForm] = useState({ label: '', prompt: '', clearContext: false, maxRetries: 0 })
  const [selectedEdge, setSelectedEdge] = useState<Edge | null>(null)
  const [edgeLabel, setEdgeLabel] = useState('')
  const [activeRunSteps, setActiveRunSteps] = useState<PlanStepRun[] | null>(null)

  useState(() => {
    const maxId = initialPlan.graph.nodes.reduce((mx, n) => {
      const num = parseInt(n.id.replace(/\D/g, ''), 10)
      return isNaN(num) ? mx : Math.max(mx, num)
    }, 0)
    nodeIdCounter = maxId
  })

  const flowNodeTypes = useMemo(() => planNodeTypes, [])

  const displayNodes = useMemo(() => {
    if (!activeRunSteps) return nodes
    const statusMap = new Map<string, PlanStepStatus>(
      activeRunSteps.map(s => [s.nodeId, s.status])
    )
    return nodes.map(n => ({
      ...n,
      data: { ...n.data, status: statusMap.get(n.id) },
    }))
  }, [nodes, activeRunSteps])

  const addNode = (type: 'action' | 'decision') => {
    nodeIdCounter++
    const id = `n${nodeIdCounter}`
    const newNode: Node = {
      id,
      type,
      position: { x: 250 + Math.random() * 100 - 50, y: 100 + nodes.length * 120 },
      data: { label: type === 'action' ? 'New Action' : 'New Decision', prompt: '' },
    }
    setNodes(nds => [...nds, newNode])
  }

  const onConnect = useCallback((params: Connection) => {
    const sourceNode = nodes.find(n => n.id === params.source)
    const existingOut = edges.filter(e => e.source === params.source)
    if (sourceNode?.type === 'action' && existingOut.length >= 1) {
      toast.error('Action nodes can only have one outgoing connection')
      return
    }
    if (sourceNode?.type === 'decision' && existingOut.length >= 2) {
      toast.error('Decision nodes can have at most two connections (yes/no)')
      return
    }
    const label = sourceNode?.type === 'decision' ? (params.sourceHandle || '') : ''
    const color = label === 'no' ? '#f87171' : label === 'yes' ? '#34d399' : '#71717a'
    setEdges(eds => addEdge({
      ...params,
      id: `e${params.source}-${params.target}-${label}`,
      label: label || undefined,
      animated: true,
      style: { stroke: color },
      markerEnd: { type: MarkerType.ArrowClosed, color },
    }, eds))
  }, [nodes, edges, setEdges])

  const onNodeClick = useCallback((_: React.MouseEvent, node: Node) => {
    setSelectedNode(node)
    setNodeForm({
      label: (node.data.label as string) || '',
      prompt: (node.data.prompt as string) || '',
      clearContext: (node.data.clearContext as boolean) || false,
      maxRetries: (node.data.maxRetries as number) || 0,
    })
    setSelectedEdge(null)
  }, [])

  const onEdgeClick = useCallback((_: React.MouseEvent, edge: Edge) => {
    setSelectedEdge(edge)
    setEdgeLabel((edge.label as string) || '')
    setSelectedNode(null)
  }, [])

  const onPaneClick = useCallback(() => {
    setSelectedNode(null)
    setSelectedEdge(null)
  }, [])

  const updateSelectedNode = () => {
    if (!selectedNode) return
    setNodes(nds => nds.map(n =>
      n.id === selectedNode.id ? { ...n, data: { ...n.data, label: nodeForm.label, prompt: nodeForm.prompt, clearContext: nodeForm.clearContext, maxRetries: nodeForm.maxRetries } } : n
    ))
    setSelectedNode(null)
  }

  const deleteSelectedNode = () => {
    if (!selectedNode) return
    setNodes(nds => nds.filter(n => n.id !== selectedNode.id))
    setEdges(eds => eds.filter(e => e.source !== selectedNode.id && e.target !== selectedNode.id))
    setSelectedNode(null)
  }

  const updateSelectedEdge = () => {
    if (!selectedEdge) return
    const color = edgeLabel === 'no' ? '#f87171' : edgeLabel === 'yes' ? '#34d399' : '#71717a'
    setEdges(eds => eds.map(e =>
      e.id === selectedEdge.id ? {
        ...e,
        label: edgeLabel || undefined,
        sourceHandle: edgeLabel || undefined,
        style: { stroke: color },
        markerEnd: { type: MarkerType.ArrowClosed, color },
      } : e
    ))
    setSelectedEdge(null)
  }

  const deleteSelectedEdge = () => {
    if (!selectedEdge) return
    setEdges(eds => eds.filter(e => e.id !== selectedEdge.id))
    setSelectedEdge(null)
  }

  const submitMeta = () => {
    if (!metaForm.name.trim()) {
      toast.error('Name is required')
      return
    }
    setPlan(p => ({ ...p, ...metaForm }))
    setMetaOpen(false)
  }

  const closeMeta = () => {
    setMetaOpen(false)
    if (!plan.id && !plan.name) onBack()
  }

  const savePlan = async () => {
    const graph: PlanGraph = { nodes: fromFlowNodes(nodes), edges: fromFlowEdges(edges) }
    const payload = { name: plan.name, description: plan.description, schedule: plan.schedule, enabled: plan.enabled, parameters: plan.parameters ?? {}, graph }
    try {
      if (plan.id) {
        await api.plans.update(plan.id, payload)
        toast.success('Plan saved')
      } else {
        const created = await api.plans.create(payload)
        setPlan(created)
        toast.success('Plan created')
      }
      onSaved()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Save failed')
    }
  }

  const currentPlanNodes = useMemo(() =>
    fromFlowNodes(nodes),
  [nodes])

  const toggleEnabled = async () => {
    const nextEnabled = !plan.enabled
    setPlan(prev => ({ ...prev, enabled: nextEnabled }))
    setMetaForm(prev => ({ ...prev, enabled: nextEnabled }))
    if (!plan.id) return
    const graph: PlanGraph = { nodes: fromFlowNodes(nodes), edges: fromFlowEdges(edges) }
    const payload = {
      name: plan.name,
      description: plan.description,
      schedule: plan.schedule,
      enabled: nextEnabled,
      parameters: plan.parameters ?? {},
      graph,
    }
    try {
      await api.plans.update(plan.id, payload)
      toast.success(nextEnabled ? 'Plan enabled' : 'Plan disabled')
      onSaved()
    } catch (e: unknown) {
      setPlan(prev => ({ ...prev, enabled: !nextEnabled }))
      setMetaForm(prev => ({ ...prev, enabled: !nextEnabled }))
      toast.error(e instanceof Error ? e.message : 'Toggle failed')
    }
  }

  return (
    <div className="flex flex-col h-full">
      <div className="px-4 py-3 border-b border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-950 shrink-0">
        <div className="flex items-start justify-between gap-3">
          <div className="flex items-start gap-3 min-w-0 flex-1">
            <Button variant="ghost" size="icon" onClick={onBack} title="Back to plans" className="mt-0.5">
              <ArrowLeft size={16} />
            </Button>
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2 flex-wrap">
                <span
                  className={`inline-block w-2 h-2 rounded-full shrink-0 ${
                    plan.enabled
                      ? (plan.schedule ? 'bg-emerald-500 animate-pulse' : 'bg-emerald-500')
                      : 'bg-zinc-400 dark:bg-zinc-600'
                  }`}
                  title={plan.enabled ? 'Enabled' : 'Disabled'}
                />
                <h1 className="text-base font-semibold text-zinc-900 dark:text-zinc-100 truncate">
                  {plan.name || 'New Plan'}
                </h1>
                <span className={`px-1.5 py-0.5 text-[10px] rounded border font-medium uppercase tracking-wide ${
                  plan.enabled
                    ? 'border-emerald-500/30 text-emerald-600 dark:text-emerald-400 bg-emerald-500/10'
                    : 'border-zinc-300 dark:border-zinc-700 text-zinc-500 dark:text-zinc-500'
                }`}>
                  {plan.enabled ? 'Enabled' : 'Disabled'}
                </span>
                {plan.schedule && (
                  <span
                    className="px-1.5 py-0.5 text-[10px] rounded border border-zinc-300 dark:border-zinc-700 font-mono text-zinc-500 dark:text-zinc-500"
                    title="Cron expression"
                  >
                    {plan.schedule}
                  </span>
                )}
              </div>
              <p className={`text-[12px] mt-1 ${
                plan.enabled
                  ? 'text-zinc-700 dark:text-zinc-300'
                  : 'text-zinc-500 dark:text-zinc-500'
              }`}>
                {plan.schedule
                  ? (plan.enabled
                    ? describeCron(plan.schedule)
                    : `Schedule "${describeCron(plan.schedule)}" — currently disabled, will not run`)
                  : (plan.enabled
                    ? 'Manual only — runs when you trigger it'
                    : 'Disabled — runs only when manually triggered')}
              </p>
              {plan.description && (
                <p className="text-[11px] text-zinc-500 dark:text-zinc-600 mt-0.5 truncate">
                  {plan.description}
                </p>
              )}
            </div>
          </div>
          <div className="flex items-center gap-2 shrink-0">
            <Button
              variant={plan.enabled ? 'secondary' : 'default'}
              size="sm"
              onClick={toggleEnabled}
              disabled={!plan.name}
              title={plan.enabled ? 'Disable plan (stop scheduled runs)' : 'Enable plan'}
            >
              {plan.enabled ? <Pause size={12} /> : <Play size={12} />}
              {plan.enabled ? 'Disable' : 'Enable'}
            </Button>
            <Button variant="ghost" size="sm" onClick={() => setMetaOpen(true)} className="text-xs">
              <Pencil size={12} /> Settings
            </Button>
            <div className="w-px h-6 bg-zinc-200 dark:bg-zinc-800 mx-1" />
            <Button variant="secondary" size="sm" onClick={() => addNode('action')}>
              <Zap size={12} /> Action
            </Button>
            <Button variant="secondary" size="sm" onClick={() => addNode('decision')}>
              <GitFork size={12} /> Decision
            </Button>
            <Button size="sm" onClick={savePlan} disabled={!plan.name}>
              <Save size={14} /> Save
            </Button>
          </div>
        </div>
      </div>

      <div className="flex min-h-0" style={{ flex: plan.id ? '3 1 0' : '1 1 0' }}>
        <div className="flex-1 relative">
          <ReactFlow
            nodes={displayNodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            onNodeClick={onNodeClick}
            onEdgeClick={onEdgeClick}
            onPaneClick={onPaneClick}
            nodeTypes={flowNodeTypes}
            fitView
            className="bg-zinc-50 dark:bg-zinc-950"
            defaultEdgeOptions={{ animated: true }}
          >
            <Background gap={20} size={1} />
            <Controls className="!bg-white dark:!bg-zinc-900 !border-zinc-200 dark:!border-zinc-800 !rounded-lg !shadow-sm [&>button]:!bg-white [&>button]:dark:!bg-zinc-900 [&>button]:!border-zinc-200 [&>button]:dark:!border-zinc-800 [&>button]:!text-zinc-600 [&>button]:dark:!text-zinc-400" />
            <MiniMap
              nodeColor={n => n.type === 'decision' ? '#f59e0b' : '#14b8a6'}
              className="!bg-white dark:!bg-zinc-900 !border-zinc-200 dark:!border-zinc-800 !rounded-lg !shadow-sm"
            />
          </ReactFlow>
        </div>

        {selectedNode && (
          <NodePropertiesPanel
            node={selectedNode}
            form={nodeForm}
            onFormChange={setNodeForm}
            onApply={updateSelectedNode}
            onDelete={deleteSelectedNode}
            planParameters={plan.parameters as Record<string, unknown> ?? {}}
          />
        )}

        {selectedEdge && (
          <EdgePropertiesPanel
            label={edgeLabel}
            onLabelChange={setEdgeLabel}
            onApply={updateSelectedEdge}
            onDelete={deleteSelectedEdge}
          />
        )}
      </div>

      {plan.id && (
        <div className="border-t border-zinc-200 dark:border-zinc-800 overflow-auto" style={{ flex: '2 1 0', minHeight: '120px' }}>
          <PlanRuns
            planId={plan.id}
            planNodes={currentPlanNodes}
            planParameters={plan.parameters ?? {}}
            onActiveSteps={setActiveRunSteps}
          />
        </div>
      )}

      <Dialog open={metaOpen} onOpenChange={setMetaOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{plan.id ? 'Plan Settings' : 'New Plan'}</DialogTitle>
            <DialogDescription />
          </DialogHeader>
          <div className="space-y-4">
            <FormField label="Name">
              <Input value={metaForm.name} onChange={e => setMetaForm(f => ({ ...f, name: e.target.value }))} placeholder="Deploy pipeline" />
            </FormField>
            <FormField label="Description">
              <Input value={metaForm.description} onChange={e => setMetaForm(f => ({ ...f, description: e.target.value }))} placeholder="Full deployment workflow with rollback" />
            </FormField>
            <FormField label="Schedule (cron)" hint="Leave empty for manual trigger only">
              <Input value={metaForm.schedule} onChange={e => setMetaForm(f => ({ ...f, schedule: e.target.value }))} className="font-mono" placeholder="0 3 * * *" />
            </FormField>
            <div className="flex items-center gap-2">
              <Switch checked={metaForm.enabled} onCheckedChange={v => setMetaForm(f => ({ ...f, enabled: v }))} />
              <span className="text-xs text-zinc-600 dark:text-zinc-400">{metaForm.enabled ? 'Enabled' : 'Disabled'}</span>
            </div>
            <FormField label="Parameters" hint="Define input parameters for this plan. Use {{.param_name}} in node prompts.">
              <ParameterEditor
                value={metaForm.parameters as Record<string, unknown>}
                onChange={v => setMetaForm(f => ({ ...f, parameters: v }))}
              />
            </FormField>
            <DialogFooter>
              <Button variant="secondary" onClick={closeMeta}>Cancel</Button>
              <Button onClick={submitMeta} disabled={!metaForm.name.trim()}>
                {plan.id ? 'Update' : 'Continue'}
              </Button>
            </DialogFooter>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}

function NodePropertiesPanel({ node, form, onFormChange, onApply, onDelete, planParameters }: {
  node: Node
  form: { label: string; prompt: string; clearContext: boolean; maxRetries: number }
  onFormChange: (f: { label: string; prompt: string; clearContext: boolean; maxRetries: number }) => void
  onApply: () => void
  onDelete: () => void
  planParameters: Record<string, unknown>
}) {
  const promptRef = useRef<HTMLTextAreaElement>(null)

  return (
    <div className="w-72 border-l border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 p-4 overflow-auto shrink-0">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          {node.type === 'decision' ? (
            <GitFork size={14} className="text-amber-500" />
          ) : (
            <Zap size={14} className="text-teal-500" />
          )}
          <span className="text-xs font-semibold uppercase tracking-wider text-zinc-500">
            {node.type === 'decision' ? 'Decision' : 'Action'}
          </span>
        </div>
        <Button variant="destructive" size="icon" className="h-7 w-7" onClick={onDelete}>
          <Trash2 size={12} />
        </Button>
      </div>
      <div className="space-y-3">
        <FormField label="Label">
          <Input value={form.label} onChange={e => onFormChange({ ...form, label: e.target.value })} placeholder="Step name" />
        </FormField>
        <FormField label={node.type === 'decision' ? 'Question / Condition' : 'Prompt'}>
          <Textarea
            ref={promptRef}
            value={form.prompt}
            onChange={e => onFormChange({ ...form, prompt: e.target.value })}
            className="h-32 text-xs"
            placeholder={node.type === 'decision' ? 'Was the deployment successful?' : 'Deploy the latest version to production...'}
          />
        </FormField>
        <ParamButtons
          parameters={planParameters}
          onInsert={snippet => insertAtCursor(promptRef.current, snippet, form.prompt, v => onFormChange({ ...form, prompt: v }))}
        />
        {node.type === 'decision' && (
          <p className="text-[11px] text-zinc-500 dark:text-zinc-600">
            Connect the green handle (left) for &quot;yes&quot; and red handle (right) for &quot;no&quot;
          </p>
        )}
        <div className="flex items-center gap-2">
          <Switch checked={form.clearContext} onCheckedChange={v => onFormChange({ ...form, clearContext: v })} />
          <span className="text-xs text-zinc-600 dark:text-zinc-400">Clear context</span>
        </div>
        <FormField label="Max retries" hint="0 = no retry on failure">
          <Input
            type="number"
            min={0}
            max={10}
            value={form.maxRetries}
            onChange={e => onFormChange({ ...form, maxRetries: Math.max(0, parseInt(e.target.value) || 0) })}
            className="w-20"
          />
        </FormField>
        <Button size="sm" className="w-full" onClick={onApply}>Apply</Button>
      </div>
    </div>
  )
}

function EdgePropertiesPanel({ label, onLabelChange, onApply, onDelete }: {
  label: string
  onLabelChange: (v: string) => void
  onApply: () => void
  onDelete: () => void
}) {
  return (
    <div className="w-72 border-l border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 p-4 overflow-auto shrink-0">
      <div className="flex items-center justify-between mb-4">
        <span className="text-xs font-semibold uppercase tracking-wider text-zinc-500">Edge</span>
        <Button variant="destructive" size="icon" className="h-7 w-7" onClick={onDelete}>
          <Trash2 size={12} />
        </Button>
      </div>
      <div className="space-y-3">
        <FormField label="Label" hint="Use 'yes' or 'no' for decision branches">
          <Input value={label} onChange={e => onLabelChange(e.target.value)} placeholder="yes / no" />
        </FormField>
        <Button size="sm" className="w-full" onClick={onApply}>Apply</Button>
      </div>
    </div>
  )
}
