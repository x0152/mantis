import { useState, useEffect, useCallback } from 'react'
import { Plus, Pencil, Trash2, Shield, AlertTriangle, Check, Play, Pause } from 'lucide-react'
import { api } from '../api'
import type { GuardRule, Connection } from '../types'
import Modal from '../components/Modal'

export default function GuardRulesPage() {
  const [rules, setRules] = useState<GuardRule[]>([])
  const [connections, setConnections] = useState<Connection[]>([])
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<GuardRule | null>(null)
  const [form, setForm] = useState({ name: '', description: '', pattern: '', connectionId: '', enabled: true })
  const [toast, setToast] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  const showToast = (type: 'success' | 'error', text: string) => {
    setToast({ type, text })
    setTimeout(() => setToast(null), 3000)
  }

  const load = useCallback(async () => {
    try {
      setLoading(true)
      const [r, c] = await Promise.all([api.guardRules.list(), api.connections.list()])
      setRules(r)
      setConnections(c)
    } catch (e: unknown) {
      showToast('error', e instanceof Error ? e.message : 'Failed to load')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])

  const openCreate = () => {
    setEditing(null)
    setForm({ name: '', description: '', pattern: '', connectionId: '', enabled: true })
    setModalOpen(true)
  }

  const openEdit = (r: GuardRule) => {
    setEditing(r)
    setForm({ name: r.name, description: r.description, pattern: r.pattern, connectionId: r.connectionId, enabled: r.enabled })
    setModalOpen(true)
  }

  const submit = async () => {
    try {
      if (editing) {
        await api.guardRules.update(editing.id, form)
        showToast('success', 'Rule updated')
      } else {
        await api.guardRules.create(form)
        showToast('success', 'Rule created')
      }
      setModalOpen(false)
      load()
    } catch (e: unknown) {
      showToast('error', e instanceof Error ? e.message : 'Operation failed')
    }
  }

  const toggleEnabled = async (r: GuardRule) => {
    try {
      await api.guardRules.update(r.id, { ...r, enabled: !r.enabled })
      showToast('success', r.enabled ? 'Rule disabled' : 'Rule enabled')
      load()
    } catch (e: unknown) {
      showToast('error', e instanceof Error ? e.message : 'Toggle failed')
    }
  }

  const remove = async (id: string) => {
    if (!confirm('Delete this guard rule?')) return
    try {
      await api.guardRules.delete(id)
      showToast('success', 'Rule deleted')
      load()
    } catch (e: unknown) {
      showToast('error', e instanceof Error ? e.message : 'Delete failed')
    }
  }

  const connName = (id: string) => connections.find(c => c.id === id)?.name ?? ''

  return (
    <div className="p-6 max-w-5xl">
      <div className="flex items-center justify-between mb-5">
        <div>
          <h1 className="text-lg font-semibold text-zinc-100">Guard Rules</h1>
          <p className="text-xs text-zinc-600 mt-0.5">Command validation rules for SSH agents</p>
        </div>
        <button onClick={openCreate} className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-white bg-teal-600 rounded-lg hover:bg-teal-500">
          <Plus size={14} /> Add Rule
        </button>
      </div>

      {toast && (
        <div className={`mb-4 flex items-center gap-2 px-3.5 py-2.5 rounded-lg text-xs font-medium ${
          toast.type === 'success' ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-red-500/10 text-red-400 border border-red-500/20'
        }`}>
          {toast.type === 'success' ? <Check size={14} /> : <AlertTriangle size={14} />}
          {toast.text}
        </div>
      )}

      {loading ? (
        <div className="text-center py-12 text-zinc-600 text-sm">Loading...</div>
      ) : rules.length === 0 ? (
        <div className="text-center py-20">
          <Shield size={40} className="mx-auto text-zinc-700 mb-3" />
          <p className="text-zinc-400 text-sm font-medium">No guard rules yet</p>
          <p className="text-xs text-zinc-600 mt-1">Create rules to validate SSH commands</p>
        </div>
      ) : (
        <div className="space-y-2.5">
          {rules.map(r => (
            <div key={r.id} className={`bg-zinc-900 rounded-lg border border-zinc-800 overflow-hidden ${!r.enabled ? 'opacity-50' : ''}`}>
              <div className="px-4 py-3.5">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2 min-w-0">
                    <span className="font-medium text-zinc-200 text-sm">{r.name}</span>
                    {r.connectionId ? (
                      <span className="px-1.5 py-0.5 text-[11px] font-medium rounded-full bg-blue-500/15 text-blue-400">{connName(r.connectionId) || r.connectionId}</span>
                    ) : (
                      <span className="px-1.5 py-0.5 text-[11px] font-medium rounded-full bg-violet-500/15 text-violet-400">Global</span>
                    )}
                    <span className={`px-1.5 py-0.5 text-[11px] font-medium rounded-full ${r.enabled ? 'bg-emerald-500/15 text-emerald-400' : 'bg-zinc-800 text-zinc-600'}`}>
                      {r.enabled ? 'Active' : 'Disabled'}
                    </span>
                  </div>
                  <div className="flex gap-0.5 ml-3">
                    <button onClick={() => toggleEnabled(r)} className={`p-1.5 rounded-md ${r.enabled ? 'text-zinc-600 hover:text-amber-400 hover:bg-amber-500/10' : 'text-zinc-600 hover:text-emerald-400 hover:bg-emerald-500/10'}`}>
                      {r.enabled ? <Pause size={14} /> : <Play size={14} />}
                    </button>
                    <button onClick={() => openEdit(r)} className="p-1.5 rounded-md text-zinc-600 hover:text-teal-400 hover:bg-teal-500/10">
                      <Pencil size={14} />
                    </button>
                    <button onClick={() => remove(r.id)} className="p-1.5 rounded-md text-zinc-600 hover:text-red-400 hover:bg-red-500/10">
                      <Trash2 size={14} />
                    </button>
                  </div>
                </div>
                {r.description && <p className="text-xs text-zinc-500 mt-1">{r.description}</p>}
                <div className="mt-2.5 bg-zinc-950 rounded-lg border border-zinc-800 p-3">
                  <p className="text-[11px] font-semibold text-zinc-600 uppercase tracking-wider mb-1">Pattern (regex)</p>
                  <code className="text-sm text-zinc-400 font-mono">{r.pattern}</code>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      <Modal open={modalOpen} onClose={() => setModalOpen(false)} title={editing ? 'Edit Guard Rule' : 'New Guard Rule'}>
        <div className="space-y-4">
          <div>
            <label className="block text-xs font-medium text-zinc-400 mb-1.5">Name</label>
            <input
              value={form.name}
              onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
              className="w-full px-3 py-2 border border-zinc-700 rounded-lg text-sm bg-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-teal-500/50"
              placeholder="Block destructive commands"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-zinc-400 mb-1.5">Description</label>
            <input
              value={form.description}
              onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
              className="w-full px-3 py-2 border border-zinc-700 rounded-lg text-sm bg-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-teal-500/50"
              placeholder="Prevents rm -rf on critical paths"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-zinc-400 mb-1.5">Pattern (regex)</label>
            <input
              value={form.pattern}
              onChange={e => setForm(f => ({ ...f, pattern: e.target.value }))}
              className="w-full px-3 py-2 border border-zinc-700 rounded-lg text-sm font-mono bg-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-teal-500/50"
              placeholder="rm\s+-rf\s+/"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-zinc-400 mb-1.5">Scope</label>
            <select
              value={form.connectionId}
              onChange={e => setForm(f => ({ ...f, connectionId: e.target.value }))}
              className="w-full px-3 py-2 border border-zinc-700 rounded-lg text-sm bg-zinc-800 text-zinc-100 focus:outline-none focus:border-teal-500/50"
            >
              <option value="">Global (all connections)</option>
              {connections.map(c => (
                <option key={c.id} value={c.id}>{c.name}</option>
              ))}
            </select>
          </div>
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => setForm(f => ({ ...f, enabled: !f.enabled }))}
              className={`relative inline-flex h-5 w-9 items-center rounded-full ${form.enabled ? 'bg-teal-600' : 'bg-zinc-700'}`}
            >
              <span className={`inline-block h-3.5 w-3.5 rounded-full bg-white ${form.enabled ? 'translate-x-[18px]' : 'translate-x-[3px]'}`} />
            </button>
            <span className="text-xs text-zinc-400">{form.enabled ? 'Enabled' : 'Disabled'}</span>
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <button onClick={() => setModalOpen(false)} className="px-3.5 py-2 text-xs font-medium text-zinc-400 bg-zinc-800 border border-zinc-700 rounded-lg hover:text-zinc-200 hover:bg-zinc-700">
              Cancel
            </button>
            <button onClick={submit} disabled={!form.name || !form.pattern} className="px-3.5 py-2 text-xs font-medium text-white bg-teal-600 rounded-lg hover:bg-teal-500 disabled:opacity-40">
              {editing ? 'Update' : 'Create'}
            </button>
          </div>
        </div>
      </Modal>
    </div>
  )
}
