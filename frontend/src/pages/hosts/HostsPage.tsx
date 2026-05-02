import { useCallback, useEffect, useState } from 'react'
import { Cloud, Container } from '@/lib/icons'
import { toast } from 'sonner'
import { Stepper } from '@/components/StepperSection'
import { ConfirmDelete } from '@/components/ConfirmDelete'
import { api } from '@/api'
import type { Connection, GuardProfile, Preset, SandboxStatus } from '@/types'
import { RemoteServersSection } from './RemoteServersSection'
import { SandboxesSection } from './SandboxesSection'
import { ServerDialog } from './ServerDialog'
import {
  EMPTY_SSH,
  parseSshConfig,
  type ServerForm,
  type SshConfig,
} from './types'

const INITIAL_FORM: ServerForm = {
  type: 'ssh',
  name: '',
  description: '',
  presetId: '',
  profileIds: [],
  memoryEnabled: true,
}

export default function HostsPage() {
  const [connections, setConnections] = useState<Connection[]>([])
  const [sandboxes, setSandboxes] = useState<SandboxStatus[]>([])
  const [presets, setPresets] = useState<Preset[]>([])
  const [profiles, setProfiles] = useState<GuardProfile[]>([])
  const [loading, setLoading] = useState(true)

  const [expanded, setExpanded] = useState<string | null>(null)
  const [memoryInput, setMemoryInput] = useState('')

  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<Connection | null>(null)
  const [form, setForm] = useState<ServerForm>(INITIAL_FORM)
  const [ssh, setSsh] = useState<SshConfig>(EMPTY_SSH)
  const [profileDropdownOpen, setProfileDropdownOpen] = useState(false)

  const [deleteServerId, setDeleteServerId] = useState<string | null>(null)
  const [deleteSandboxName, setDeleteSandboxName] = useState<string | null>(null)
  const [busy, setBusy] = useState<string | null>(null)

  const loadAll = useCallback(async () => {
    try {
      setLoading(true)
      const [conns, sbx, ps, gps] = await Promise.all([
        api.connections.list(),
        api.sandboxes.list(),
        api.presets.list(),
        api.guardProfiles.list(),
      ])
      setConnections(conns)
      setSandboxes(sbx.sort((a, b) => a.connection.name.localeCompare(b.connection.name)))
      setPresets(ps)
      setProfiles(gps)
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Failed to load hosts')
    } finally {
      setLoading(false)
    }
  }, [])

  const refreshSandboxes = useCallback(async () => {
    try {
      const sbx = await api.sandboxes.list()
      setSandboxes(sbx.sort((a, b) => a.connection.name.localeCompare(b.connection.name)))
    } catch {}
  }, [])

  useEffect(() => {
    loadAll()
  }, [loadAll])

  useEffect(() => {
    const interval = setInterval(refreshSandboxes, 5000)
    return () => clearInterval(interval)
  }, [refreshSandboxes])

  const remoteConnections = connections.filter(c => !c.dockerfile)

  const profileName = (id: string) => profiles.find(p => p.id === id)?.name ?? id.slice(0, 8)
  const presetName = (id: string) => {
    const p = presets.find(x => x.id === id)
    if (p) return p.name
    return id ? id.slice(0, 8) + '…' : '—'
  }

  const buildConfig = () => {
    const cfg: Record<string, unknown> = {
      host: ssh.host,
      port: parseInt(ssh.port) || 22,
      username: ssh.username,
    }
    if (ssh.password) cfg.password = ssh.password
    if (ssh.privateKey) cfg.privateKey = ssh.privateKey
    return cfg
  }

  const openCreateServer = () => {
    setEditing(null)
    setForm(INITIAL_FORM)
    setSsh(EMPTY_SSH)
    setProfileDropdownOpen(false)
    setModalOpen(true)
  }

  const openEditServer = (c: Connection) => {
    setEditing(c)
    setForm({
      type: c.type,
      name: c.name,
      description: c.description,
      presetId: c.presetId ?? '',
      profileIds: c.profileIds || [],
      memoryEnabled: c.memoryEnabled,
    })
    setSsh(parseSshConfig(c.config))
    setProfileDropdownOpen(false)
    setModalOpen(true)
  }

  const submitServer = async () => {
    try {
      const data = {
        type: form.type,
        name: form.name,
        description: form.description,
        presetId: form.presetId,
        config: buildConfig(),
        profileIds: form.profileIds,
        memoryEnabled: form.memoryEnabled,
      }
      if (editing) {
        await api.connections.update(editing.id, data)
        toast.success('Server updated')
      } else {
        await api.connections.create(data)
        toast.success('Server created')
      }
      setModalOpen(false)
      loadAll()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Operation failed')
    }
  }

  const removeServer = async (id: string) => {
    try {
      await api.connections.delete(id)
      toast.success('Server deleted')
      setDeleteServerId(null)
      if (expanded === id) setExpanded(null)
      loadAll()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Delete failed')
    }
  }

  const addMemory = async (connectionId: string) => {
    if (!memoryInput.trim()) return
    try {
      await api.connections.addMemory(connectionId, memoryInput.trim())
      setMemoryInput('')
      toast.success('Memory added')
      loadAll()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Failed to add memory')
    }
  }

  const deleteMemory = async (connectionId: string, memoryId: string) => {
    try {
      await api.connections.deleteMemory(connectionId, memoryId)
      toast.success('Memory deleted')
      loadAll()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Failed to delete memory')
    }
  }

  const startSandbox = async (name: string) => {
    setBusy(name)
    try {
      await api.sandboxes.start(name)
      toast.success(`Starting ${name}`)
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Start failed')
    }
    setBusy(null)
    refreshSandboxes()
  }

  const stopSandbox = async (name: string) => {
    setBusy(name)
    try {
      await api.sandboxes.stop(name)
      toast.success(`Stopped ${name}`)
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Stop failed')
    }
    setBusy(null)
    refreshSandboxes()
  }

  const rebuildSandbox = async (name: string) => {
    setBusy(name)
    try {
      const res = await api.sandboxes.rebuild(name)
      if (!res.ok) throw new Error(`Rebuild failed: ${res.status}`)
      toast.success(`Rebuilding ${name} — follow logs in Docker logs`)
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Rebuild failed')
    }
    setBusy(null)
    refreshSandboxes()
  }

  const removeSandbox = async (name: string) => {
    setBusy(name)
    try {
      await api.sandboxes.delete(name)
      toast.success(`Removed ${name}`)
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Remove failed')
    }
    setDeleteSandboxName(null)
    setBusy(null)
    loadAll()
  }

  const toggleProfileSelection = (id: string) =>
    setForm(f => ({
      ...f,
      profileIds: f.profileIds.includes(id) ? f.profileIds.filter(x => x !== id) : [...f.profileIds, id],
    }))

  if (loading && connections.length === 0 && sandboxes.length === 0) {
    return <div className="p-6 text-center py-16 text-sm text-zinc-500 dark:text-zinc-600">Loading…</div>
  }

  return (
    <div className="p-6 max-w-5xl mx-auto">
      <header className="mb-6">
        <h1 className="text-lg font-semibold text-zinc-900 dark:text-zinc-100">Hosts</h1>
        <p className="text-xs text-zinc-500 dark:text-zinc-600 mt-0.5">
          Where the assistant runs commands. Connect to your own machines or spin up isolated sandboxes.
        </p>
      </header>

      <Stepper
        steps={[
          { n: 1, label: 'Remote',    icon: Cloud,     count: remoteConnections.length, done: remoteConnections.length > 0 },
          { n: 2, label: 'Sandboxes', icon: Container, count: sandboxes.length,         done: sandboxes.length > 0 },
        ]}
      />

      <div className="mt-6 space-y-4">
        <RemoteServersSection
          connections={remoteConnections}
          expanded={expanded}
          setExpanded={setExpanded}
          memoryInput={memoryInput}
          setMemoryInput={setMemoryInput}
          presetName={presetName}
          profileName={profileName}
          onCreate={openCreateServer}
          onEdit={openEditServer}
          onDelete={setDeleteServerId}
          addMemory={addMemory}
          deleteMemory={deleteMemory}
        />
        <SandboxesSection
          sandboxes={sandboxes}
          busy={busy}
          profileName={profileName}
          onRefresh={refreshSandboxes}
          onStart={startSandbox}
          onStop={stopSandbox}
          onRebuild={rebuildSandbox}
          onDelete={setDeleteSandboxName}
        />
      </div>

      <ServerDialog
        open={modalOpen}
        onOpenChange={setModalOpen}
        editing={editing}
        form={form}
        setForm={setForm}
        ssh={ssh}
        setSsh={setSsh}
        presets={presets}
        profiles={profiles}
        profileDropdownOpen={profileDropdownOpen}
        setProfileDropdownOpen={setProfileDropdownOpen}
        profileName={profileName}
        toggleProfileSelection={toggleProfileSelection}
        onSubmit={submitServer}
      />

      <ConfirmDelete
        open={!!deleteServerId}
        onCancel={() => setDeleteServerId(null)}
        onConfirm={() => deleteServerId && removeServer(deleteServerId)}
        title="Delete server?"
        description="This permanently removes the SSH server and its memories."
      />
      <ConfirmDelete
        open={!!deleteSandboxName}
        onCancel={() => setDeleteSandboxName(null)}
        onConfirm={() => deleteSandboxName && removeSandbox(deleteSandboxName)}
        title="Remove sandbox?"
        description="Stops and deletes the container, the SSH wrapper, and the Dockerfile record."
      />
    </div>
  )
}
