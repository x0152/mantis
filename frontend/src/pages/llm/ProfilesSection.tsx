import { ChevronDown, ChevronRight, Layers, Pencil, Plus, Trash2 } from '@/lib/icons'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Section, EmptyHint } from '@/components/StepperSection'
import type { Model, Preset } from '@/types'
import { ROLE_DEFS } from './types'
import { RoutingPanel } from './RoutingPanel'

interface Props {
  profiles: Preset[]
  models: Model[]
  routing: { chatPresetId: string; serverPresetId: string }
  expandedProfiles: Set<string>
  onToggleProfile: (id: string) => void
  modelLabel: (modelId: string) => string
  onCreate: () => void
  onEdit: (p: Preset) => void
  onDelete: (id: string) => void
  onRoutingChange: (key: 'chatPresetId' | 'serverPresetId', value: string) => void
}

export function ProfilesSection({
  profiles,
  models,
  routing,
  expandedProfiles,
  onToggleProfile,
  modelLabel,
  onCreate,
  onEdit,
  onDelete,
  onRoutingChange,
}: Props) {
  return (
    <Section
      n={3}
      icon={Layers}
      title="Profiles & routing"
      subtitle="Recipes that map models to roles — and where each runs."
      disabled={models.length === 0}
      disabledHint="Register a model first — profiles pick from your models."
      action={
        <Button size="sm" onClick={onCreate} disabled={models.length === 0}>
          <Plus size={13} /> Add profile
        </Button>
      }
    >
      {profiles.length === 0 ? (
        <EmptyHint>
          A profile assigns a model to each role: chat, summary, vision, fallback.
        </EmptyHint>
      ) : (
        <>
          <RoutingPanel routing={routing} profiles={profiles} onChange={onRoutingChange} />
          <div className="grid grid-cols-1 md:grid-cols-1 gap-3">
            {profiles.map(p => (
              <ProfileCard
                key={p.id}
                profile={p}
                used={{
                  chat: routing.chatPresetId === p.id,
                  server: routing.serverPresetId === p.id,
                }}
                isExpanded={expandedProfiles.has(p.id)}
                modelLabel={modelLabel}
                onToggle={() => onToggleProfile(p.id)}
                onEdit={() => onEdit(p)}
                onDelete={() => onDelete(p.id)}
              />
            ))}
          </div>
        </>
      )}
    </Section>
  )
}

interface CardProps {
  profile: Preset
  used: { chat: boolean; server: boolean }
  isExpanded: boolean
  modelLabel: (id: string) => string
  onToggle: () => void
  onEdit: () => void
  onDelete: () => void
}

function ProfileCard({ profile, used, isExpanded, modelLabel, onToggle, onEdit, onDelete }: CardProps) {
  return (
    <div className="rounded-lg border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 overflow-hidden">
      <div className="flex items-center px-4 py-3 gap-2">
        <button onClick={onToggle} className="flex items-center gap-2 flex-1 min-w-0 text-left">
          {isExpanded
            ? <ChevronDown size={14} className="text-zinc-500 shrink-0" />
            : <ChevronRight size={14} className="text-zinc-500 shrink-0" />}
          <span className="font-medium text-sm text-zinc-800 dark:text-zinc-200 truncate">{profile.name}</span>
          {used.chat && <Badge variant="success">chat</Badge>}
          {used.server && <Badge>server</Badge>}
        </button>
        <div className="flex gap-0.5 shrink-0">
          <Button variant="ghost" size="icon" onClick={onEdit}><Pencil size={13} /></Button>
          <Button variant="destructive" size="icon" onClick={onDelete}><Trash2 size={13} /></Button>
        </div>
      </div>
      <div className="border-t border-zinc-200/60 dark:border-zinc-800/60 px-4 py-3 space-y-1.5">
        {ROLE_DEFS.map(({ key, field, icon: Icon, label }) => {
          const resolved = modelLabel(profile[field])
          const placeholder = key === 'summary' ? 'same as chat' : '—'
          return (
            <div key={key} className="flex items-center gap-2 text-[12px]">
              <Icon size={12} className="text-zinc-400 shrink-0" />
              <span className="text-zinc-500 w-16 shrink-0">{label}</span>
              <span className={resolved ? 'text-zinc-700 dark:text-zinc-300 truncate' : 'text-zinc-400 dark:text-zinc-600'}>
                {resolved || placeholder}
              </span>
            </div>
          )
        })}
        {isExpanded && (
          <div className="mt-2 pt-2 border-t border-zinc-200/60 dark:border-zinc-800/60 space-y-1.5 text-[11px]">
            <div className="flex justify-between">
              <span className="text-zinc-500">Temperature</span>
              <span className="text-zinc-700 dark:text-zinc-300">
                {profile.temperature != null ? profile.temperature : 'model default'}
              </span>
            </div>
            {profile.systemPrompt && (
              <div>
                <div className="text-zinc-500 mb-1">System prompt</div>
                <div className="text-zinc-600 dark:text-zinc-400 bg-zinc-50 dark:bg-zinc-950 rounded p-2 font-mono text-[10.5px] leading-relaxed whitespace-pre-wrap">
                  {profile.systemPrompt}
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
