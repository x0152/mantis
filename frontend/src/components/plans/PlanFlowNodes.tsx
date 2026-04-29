import { Handle, Position, MarkerType, type NodeProps, type Node, type Edge } from '@xyflow/react'
import { Zap, GitFork } from '@/lib/icons'
import type { PlanNode, PlanEdge, PlanStepStatus } from '../../types'

const statusBorder: Record<PlanStepStatus, string> = {
  pending: 'border-zinc-300 dark:border-zinc-700',
  running: 'border-blue-500 animate-pulse',
  completed: 'border-emerald-500',
  failed: 'border-red-500',
  skipped: 'border-zinc-400 dark:border-zinc-600 opacity-50',
}

const statusDot: Record<PlanStepStatus, string> = {
  pending: 'bg-zinc-400',
  running: 'bg-blue-500 animate-pulse',
  completed: 'bg-emerald-500',
  failed: 'bg-red-500',
  skipped: 'bg-zinc-400',
}

export function ActionNode({ data, selected }: NodeProps) {
  const status = data.status as PlanStepStatus | undefined
  const borderClass = status ? statusBorder[status] : (selected ? 'border-teal-500' : 'border-zinc-300 dark:border-zinc-700')

  return (
    <div className={`px-4 py-3 rounded-lg border-2 bg-white dark:bg-zinc-900 min-w-[180px] max-w-[240px] shadow-sm ${borderClass}`}>
      <Handle type="target" position={Position.Top} className="!w-3 !h-3 !bg-teal-500 !border-2 !border-white dark:!border-zinc-900" />
      <div className="flex items-center gap-2 mb-1">
        <Zap size={12} className="text-teal-500 shrink-0" />
        <span className="text-[10px] font-semibold uppercase tracking-wider text-teal-600 dark:text-teal-400">Action</span>
        {status && <div className={`w-2 h-2 rounded-full ml-auto ${statusDot[status]}`} />}
      </div>
      <p className="text-sm font-medium text-zinc-800 dark:text-zinc-200 truncate">{String(data.label || 'Untitled')}</p>
      {data.prompt ? (
        <p className="text-[11px] text-zinc-500 mt-1 line-clamp-2">{String(data.prompt)}</p>
      ) : null}
      <NodeBadges data={data} />
      <Handle type="source" position={Position.Bottom} className="!w-3 !h-3 !bg-teal-500 !border-2 !border-white dark:!border-zinc-900" />
    </div>
  )
}

export function DecisionNode({ data, selected }: NodeProps) {
  const status = data.status as PlanStepStatus | undefined
  const borderClass = status ? statusBorder[status] : (selected ? 'border-amber-500' : 'border-zinc-300 dark:border-zinc-700')

  return (
    <div className={`px-4 py-3 rounded-lg border-2 bg-white dark:bg-zinc-900 min-w-[180px] max-w-[240px] shadow-sm ${borderClass}`}>
      <Handle type="target" position={Position.Top} className="!w-3 !h-3 !bg-amber-500 !border-2 !border-white dark:!border-zinc-900" />
      <div className="flex items-center gap-2 mb-1">
        <GitFork size={12} className="text-amber-500 shrink-0" />
        <span className="text-[10px] font-semibold uppercase tracking-wider text-amber-600 dark:text-amber-400">Decision</span>
        {status && <div className={`w-2 h-2 rounded-full ml-auto ${statusDot[status]}`} />}
      </div>
      <p className="text-sm font-medium text-zinc-800 dark:text-zinc-200 truncate">{String(data.label || 'Untitled')}</p>
      {data.prompt ? (
        <p className="text-[11px] text-zinc-500 mt-1 line-clamp-2">{String(data.prompt)}</p>
      ) : null}
      <NodeBadges data={data} />
      <Handle type="source" position={Position.Bottom} id="yes" style={{ left: '30%' }} className="!w-3 !h-3 !bg-emerald-500 !border-2 !border-white dark:!border-zinc-900" />
      <Handle type="source" position={Position.Bottom} id="no" style={{ left: '70%' }} className="!w-3 !h-3 !bg-red-400 !border-2 !border-white dark:!border-zinc-900" />
    </div>
  )
}

function NodeBadges({ data }: { data: Record<string, unknown> }) {
  const cc = data.clearContext as boolean | undefined
  const retries = data.maxRetries as number | undefined
  if (!cc && !retries) return null
  return (
    <div className="flex gap-1 mt-1.5">
      {cc && <span className="px-1 py-0.5 text-[9px] rounded bg-violet-500/10 text-violet-500 font-medium">clean ctx</span>}
      {!!retries && retries > 0 && <span className="px-1 py-0.5 text-[9px] rounded bg-amber-500/10 text-amber-500 font-medium">retry ×{retries}</span>}
    </div>
  )
}

export const planNodeTypes = { action: ActionNode, decision: DecisionNode }

function edgeColor(label: string) {
  if (label === 'no') return '#f87171'
  if (label === 'yes') return '#34d399'
  return '#71717a'
}

export function toFlowNodes(planNodes: PlanNode[], stepStatuses?: Map<string, PlanStepStatus>): Node[] {
  return planNodes.map(n => ({
    id: n.id,
    type: n.type,
    position: n.position,
    data: {
      label: n.label,
      prompt: n.prompt,
      clearContext: n.clearContext,
      maxRetries: n.maxRetries,
      status: stepStatuses?.get(n.id),
    },
    selected: false,
  }))
}

export function toFlowEdges(planEdges: PlanEdge[]): Edge[] {
  return planEdges.map(e => ({
    id: e.id,
    source: e.source,
    target: e.target,
    sourceHandle: e.label || undefined,
    label: e.label || undefined,
    animated: true,
    style: { stroke: edgeColor(e.label) },
    markerEnd: { type: MarkerType.ArrowClosed, color: edgeColor(e.label) },
  }))
}

export function fromFlowNodes(nodes: Node[]): PlanNode[] {
  return nodes.map(n => ({
    id: n.id,
    type: (n.type || 'action') as 'action' | 'decision',
    label: (n.data.label as string) || '',
    prompt: (n.data.prompt as string) || '',
    position: { x: n.position.x, y: n.position.y },
    clearContext: (n.data.clearContext as boolean) || false,
    maxRetries: (n.data.maxRetries as number) || 0,
  }))
}

export function fromFlowEdges(edges: Edge[]): PlanEdge[] {
  return edges.map(e => ({
    id: e.id,
    source: e.source,
    target: e.target,
    label: (e.sourceHandle as string) || '',
  }))
}
