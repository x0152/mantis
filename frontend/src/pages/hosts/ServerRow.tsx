import {
  ChevronDown, ChevronRight, Cloud, MessageSquare, Pencil, Send, Shield, Trash2,
} from '@/lib/icons'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import type { Connection } from '@/types'

interface Props {
  conn: Connection
  isOpen: boolean
  memoryInput: string
  setMemoryInput: (v: string) => void
  onToggle: () => void
  onEdit: () => void
  onDelete: () => void
  presetName: (id: string) => string
  profileName: (id: string) => string
  addMemory: (connectionId: string) => void
  deleteMemory: (connectionId: string, memoryId: string) => void
}

export function ServerRow({
  conn,
  isOpen,
  memoryInput,
  setMemoryInput,
  onToggle,
  onEdit,
  onDelete,
  presetName,
  profileName,
  addMemory,
  deleteMemory,
}: Props) {
  return (
    <div>
      <div
        className="flex items-center px-4 py-3 cursor-pointer hover:bg-zinc-50 dark:hover:bg-zinc-800/40"
        onClick={onToggle}
      >
        <div className="mr-2.5 text-zinc-500">
          {isOpen ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
        </div>
        <Cloud size={14} className="text-zinc-400 mr-2 shrink-0" />
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <span className="font-medium text-zinc-800 dark:text-zinc-200 text-sm">{conn.name}</span>
            <Badge variant="secondary" className="uppercase">{conn.type}</Badge>
            {conn.memories.length > 0 && (
              <Badge variant="secondary">
                <MessageSquare size={10} /> {conn.memories.length}
              </Badge>
            )}
            {conn.profileIds?.length > 0 && (
              <Badge variant="warning">
                <Shield size={10} /> {conn.profileIds.map(id => profileName(id)).join(', ')}
              </Badge>
            )}
          </div>
          {conn.description && (
            <p className="text-[11px] text-zinc-500 dark:text-zinc-500 mt-0.5 truncate">{conn.description}</p>
          )}
          <p className="text-[11px] text-zinc-500 dark:text-zinc-600 mt-0.5">
            Preset: {conn.presetId ? presetName(conn.presetId) : <span className="italic">Inherit (global)</span>}
            {conn.config?.host ? <> · <span className="font-mono">{String(conn.config.username || '')}@{String(conn.config.host)}</span></> : null}
          </p>
        </div>
        <div className="flex gap-0.5 ml-2 shrink-0" onClick={e => e.stopPropagation()}>
          <Button variant="ghost" size="icon" onClick={onEdit}>
            <Pencil size={13} />
          </Button>
          <Button variant="destructive" size="icon" onClick={onDelete}>
            <Trash2 size={13} />
          </Button>
        </div>
      </div>
      {isOpen && (
        <div className="border-t border-zinc-200 dark:border-zinc-800 px-4 py-4 bg-zinc-50 dark:bg-zinc-950/50">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
            <div>
              <h3 className="text-[11px] font-semibold text-zinc-500 uppercase tracking-wider mb-2">SSH config</h3>
              <pre className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-lg p-3 text-xs font-mono text-zinc-600 dark:text-zinc-400 overflow-auto max-h-48">
                {JSON.stringify(conn.config, null, 2)}
              </pre>
            </div>
            <div>
              <h3 className="text-[11px] font-semibold text-zinc-500 uppercase tracking-wider mb-2">Details</h3>
              <dl className="space-y-2 text-sm">
                <div className="flex justify-between gap-2">
                  <dt className="text-zinc-600 dark:text-zinc-500">ID</dt>
                  <dd className="font-mono text-[11px] text-zinc-500 dark:text-zinc-600 truncate">{conn.id}</dd>
                </div>
                <div className="flex justify-between gap-2">
                  <dt className="text-zinc-600 dark:text-zinc-500">Type</dt>
                  <dd className="text-zinc-700 dark:text-zinc-300">{conn.type}</dd>
                </div>
                <div className="flex justify-between gap-2">
                  <dt className="text-zinc-600 dark:text-zinc-500">Preset</dt>
                  <dd className="text-zinc-700 dark:text-zinc-300">
                    {conn.presetId ? presetName(conn.presetId) : <span className="italic text-zinc-500">Inherit (global)</span>}
                  </dd>
                </div>
                <div className="flex justify-between gap-2">
                  <dt className="text-zinc-600 dark:text-zinc-500">Profiles</dt>
                  <dd className="text-zinc-700 dark:text-zinc-300">
                    {conn.profileIds?.length ? conn.profileIds.map(id => profileName(id)).join(', ') : 'None (unrestricted)'}
                  </dd>
                </div>
              </dl>
            </div>
          </div>

          <div className="mt-5">
            <h3 className="text-[11px] font-semibold text-zinc-500 uppercase tracking-wider mb-2.5">
              Memories ({conn.memories.length})
            </h3>
            {conn.memories.length > 0 && (
              <div className="space-y-1.5 mb-3">
                {conn.memories.map(mem => (
                  <div
                    key={mem.id}
                    className="flex items-start gap-2.5 bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-lg px-3 py-2.5"
                  >
                    <MessageSquare size={13} className="text-violet-400 mt-0.5 shrink-0" />
                    <div className="flex-1 min-w-0">
                      <p className="text-sm text-zinc-700 dark:text-zinc-300">{mem.content}</p>
                      <p className="text-[11px] text-zinc-500 dark:text-zinc-600 mt-1">
                        {new Date(mem.createdAt).toLocaleString()}
                      </p>
                    </div>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => deleteMemory(conn.id, mem.id)}
                      className="shrink-0"
                    >
                      <Trash2 size={12} />
                    </Button>
                  </div>
                ))}
              </div>
            )}
            <div className="flex gap-2">
              <Input
                value={memoryInput}
                onChange={e => setMemoryInput(e.target.value)}
                onKeyDown={e => e.key === 'Enter' && addMemory(conn.id)}
                placeholder="Add a memory…"
              />
              <Button size="sm" onClick={() => addMemory(conn.id)} disabled={!memoryInput.trim()}>
                <Send size={14} />
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
