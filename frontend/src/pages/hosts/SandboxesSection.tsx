import { Container, Play, RotateCw, Shield, Square, Trash2 } from '@/lib/icons'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Section, EmptyHint } from '@/components/StepperSection'
import type { SandboxStatus } from '@/types'
import { relativeTime, sandboxName, stateClass } from './types'

interface Props {
  sandboxes: SandboxStatus[]
  busy: string | null
  profileName: (id: string) => string
  onRefresh: () => void
  onStart: (name: string) => void
  onStop: (name: string) => void
  onRebuild: (name: string) => void
  onDelete: (name: string) => void
}

export function SandboxesSection({
  sandboxes,
  busy,
  profileName,
  onRefresh,
  onStart,
  onStop,
  onRebuild,
  onDelete,
}: Props) {
  return (
    <Section
      n={2}
      icon={Container}
      title="Sandboxes"
      subtitle="Isolated local containers built from a Dockerfile. Safe to break — restart them any time."
      action={
        <Button size="sm" variant="secondary" onClick={onRefresh}>
          <RotateCw size={13} /> Refresh
        </Button>
      }
    >
      {sandboxes.length === 0 ? (
        <EmptyHint>
          Sandboxes appear here once a Dockerfile template is registered. Add one in the setup wizard or via the
          configuration API — they auto-build and stay in sync.
        </EmptyHint>
      ) : (
        <div className="rounded-lg border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 divide-y divide-zinc-100 dark:divide-zinc-800 overflow-hidden">
          {sandboxes.map(item => (
            <SandboxRow
              key={item.connection.id}
              item={item}
              busy={busy === sandboxName(item.connection.name)}
              profileName={profileName}
              onStart={onStart}
              onStop={onStop}
              onRebuild={onRebuild}
              onDelete={onDelete}
            />
          ))}
        </div>
      )}
    </Section>
  )
}

interface RowProps {
  item: SandboxStatus
  busy: boolean
  profileName: (id: string) => string
  onStart: (name: string) => void
  onStop: (name: string) => void
  onRebuild: (name: string) => void
  onDelete: (name: string) => void
}

function SandboxRow({ item, busy, profileName, onStart, onStop, onRebuild, onDelete }: RowProps) {
  const name = sandboxName(item.connection.name)
  const isRunning = item.state === 'running'
  const isBuilt = item.state !== 'not-built'
  return (
    <div className="px-4 py-3 flex items-center gap-3">
      <Container size={16} className="text-zinc-400 shrink-0" />
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 flex-wrap">
          <span className="font-medium text-zinc-800 dark:text-zinc-200 text-sm">{name}</span>
          <Badge className={`border ${stateClass(item.state)}`}>{item.state}</Badge>
          {item.connection.profileIds?.[0] && (
            <Badge variant="secondary">
              <Shield size={10} /> {profileName(item.connection.profileIds[0])}
            </Badge>
          )}
        </div>
        <div className="mt-0.5 flex flex-wrap items-center gap-x-3 gap-y-0.5 text-[11px] text-zinc-500 dark:text-zinc-500">
          <span>ssh: <span className="font-mono text-zinc-600 dark:text-zinc-400">{item.connection.name}</span></span>
          {item.container.host && <span>host: <span className="font-mono">{item.container.host}</span></span>}
          {item.container.image && <span>image: <span className="font-mono">{item.container.image}</span></span>}
          {isBuilt && <span>started: {relativeTime(item.container.createdAt)}</span>}
        </div>
        {item.connection.description && (
          <p className="mt-0.5 text-[11px] text-zinc-500 dark:text-zinc-500 line-clamp-1">
            {item.connection.description}
          </p>
        )}
      </div>
      <div className="flex gap-0.5 shrink-0">
        {isRunning ? (
          <Button variant="ghost" size="icon" disabled={busy} onClick={() => onStop(name)} title="Stop">
            <Square size={13} />
          </Button>
        ) : (
          <Button
            variant="ghost"
            size="icon"
            disabled={busy || !isBuilt}
            onClick={() => onStart(name)}
            title={isBuilt ? 'Start' : 'Not built yet — rebuild first'}
          >
            <Play size={13} />
          </Button>
        )}
        <Button variant="ghost" size="icon" disabled={busy} onClick={() => onRebuild(name)} title="Rebuild">
          <RotateCw size={13} className={busy ? 'animate-spin' : ''} />
        </Button>
        <Button variant="destructive" size="icon" disabled={busy} onClick={() => onDelete(name)} title="Remove">
          <Trash2 size={13} />
        </Button>
      </div>
    </div>
  )
}
