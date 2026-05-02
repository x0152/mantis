import { Cloud, Plus } from '@/lib/icons'
import { Button } from '@/components/ui/button'
import { Section, EmptyHint } from '@/components/StepperSection'
import type { Connection } from '@/types'
import { ServerRow } from './ServerRow'

interface Props {
  connections: Connection[]
  expanded: string | null
  setExpanded: (id: string | null) => void
  memoryInput: string
  setMemoryInput: (v: string) => void
  presetName: (id: string) => string
  profileName: (id: string) => string
  onCreate: () => void
  onEdit: (c: Connection) => void
  onDelete: (id: string) => void
  addMemory: (connectionId: string) => void
  deleteMemory: (connectionId: string, memoryId: string) => void
}

export function RemoteServersSection({
  connections,
  expanded,
  setExpanded,
  memoryInput,
  setMemoryInput,
  presetName,
  profileName,
  onCreate,
  onEdit,
  onDelete,
  addMemory,
  deleteMemory,
}: Props) {
  return (
    <Section
      n={1}
      icon={Cloud}
      title="Remote servers"
      subtitle="SSH into machines you already own — laptops, VPS, production boxes."
      action={
        <Button size="sm" onClick={onCreate}>
          <Plus size={13} /> Add server
        </Button>
      }
    >
      {connections.length === 0 ? (
        <EmptyHint>
          Add a server to give the assistant SSH access to a machine you control. The assistant will only see what you
          allow via security profiles.
        </EmptyHint>
      ) : (
        <div className="rounded-lg border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 divide-y divide-zinc-100 dark:divide-zinc-800 overflow-hidden">
          {connections.map(conn => (
            <ServerRow
              key={conn.id}
              conn={conn}
              isOpen={expanded === conn.id}
              memoryInput={memoryInput}
              setMemoryInput={setMemoryInput}
              onToggle={() => setExpanded(expanded === conn.id ? null : conn.id)}
              onEdit={() => onEdit(conn)}
              onDelete={() => onDelete(conn.id)}
              presetName={presetName}
              profileName={profileName}
              addMemory={addMemory}
              deleteMemory={deleteMemory}
            />
          ))}
        </div>
      )}
    </Section>
  )
}
