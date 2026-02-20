import { useState, useEffect, useCallback } from 'react'
import { Plus, Pencil, Trash2, Radio, AlertTriangle, Check } from 'lucide-react'
import { api } from '../api'
import type { Channel, Model } from '../types'
import Modal from '../components/Modal'

function parseAllowedUserIds(s: string): number[] {
  return s
    .split(',')
    .map(x => x.trim())
    .filter(Boolean)
    .map(x => Number.parseInt(x, 10))
    .filter(n => Number.isFinite(n))
}

export default function ChannelsPage() {
  const [channels, setChannels] = useState<Channel[]>([])
  const [models, setModels] = useState<Model[]>([])
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<Channel | null>(null)
  const [form, setForm] = useState({ name: '', token: '', modelId: '', allowedUserIds: '' })
  const [toast, setToast] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  const showToast = (type: 'success' | 'error', text: string) => {
    setToast({ type, text })
    setTimeout(() => setToast(null), 3000)
  }

  const load = useCallback(async () => {
    try {
      setLoading(true)
      const [chs, mdls] = await Promise.all([api.channels.list(), api.models.list()])
      setChannels(chs)
      setModels(mdls)
    } catch (e: unknown) {
      showToast('error', e instanceof Error ? e.message : 'Failed to load')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])

  const openCreate = () => {
    setEditing(null)
    setForm({ name: '', token: '', modelId: '', allowedUserIds: '' })
    setModalOpen(true)
  }

  const openEdit = (c: Channel) => {
    setEditing(c)
    setForm({
      name: c.name,
      token: c.token,
      modelId: c.modelId,
      allowedUserIds: (c.allowedUserIds ?? []).join(','),
    })
    setModalOpen(true)
  }

  const modelName = (id: string) => {
    const m = models.find(x => x.id === id)
    if (m) return `${m.name} (${m.connectionId})`
    return id ? (id.slice(0, 8) + '...') : '—'
  }

  const submit = async () => {
    try {
      const allowed = parseAllowedUserIds(form.allowedUserIds)
      if (editing) {
        const token = editing.type === 'chat' ? '' : form.token
        const allowedUserIds = editing.type === 'chat' ? [] : allowed
        await api.channels.update(editing.id, { name: form.name, token, modelId: form.modelId, allowedUserIds })
        showToast('success', 'Channel updated')
      } else {
        await api.channels.create({ type: 'telegram', name: form.name, token: form.token, modelId: form.modelId, allowedUserIds: allowed })
        showToast('success', 'Channel created')
      }
      setModalOpen(false)
      load()
    } catch (e: unknown) {
      showToast('error', e instanceof Error ? e.message : 'Operation failed')
    }
  }

  const remove = async (id: string) => {
    if (!confirm('Delete this channel?')) return
    try {
      await api.channels.delete(id)
      showToast('success', 'Channel deleted')
      load()
    } catch (e: unknown) {
      showToast('error', e instanceof Error ? e.message : 'Delete failed')
    }
  }

  return (
    <div className="p-6 max-w-5xl">
      <div className="flex items-center justify-between mb-5">
        <div>
          <h1 className="text-lg font-semibold text-zinc-100">Channels</h1>
          <p className="text-xs text-zinc-600 mt-0.5">Chat channel is default; Telegram channels can be added</p>
        </div>
        <button onClick={openCreate} className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-white bg-teal-600 rounded-lg hover:bg-teal-500">
          <Plus size={14} /> Add Telegram Channel
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
      ) : channels.length === 0 ? (
        <div className="text-center py-20">
          <Radio size={40} className="mx-auto text-zinc-700 mb-3" />
          <p className="text-zinc-400 text-sm font-medium">No channels yet</p>
        </div>
      ) : (
        <div className="bg-zinc-900 rounded-lg border border-zinc-800 overflow-hidden">
          <table className="w-full">
            <thead>
              <tr className="border-b border-zinc-800">
                <th className="text-left px-5 py-3 text-[11px] font-semibold text-zinc-500 uppercase tracking-wider">Name</th>
                <th className="text-left px-5 py-3 text-[11px] font-semibold text-zinc-500 uppercase tracking-wider">Type</th>
                <th className="text-left px-5 py-3 text-[11px] font-semibold text-zinc-500 uppercase tracking-wider">Model</th>
                <th className="text-left px-5 py-3 text-[11px] font-semibold text-zinc-500 uppercase tracking-wider">Token</th>
                <th className="text-left px-5 py-3 text-[11px] font-semibold text-zinc-500 uppercase tracking-wider">Allowed</th>
                <th className="px-5 py-3 w-24" />
              </tr>
            </thead>
            <tbody className="divide-y divide-zinc-800/50">
              {channels.slice().sort((a, b) => {
                if (a.type !== b.type) return a.type === 'chat' ? -1 : 1
                return (a.name || a.id).localeCompare(b.name || b.id)
              }).map(ch => (
                <tr key={ch.id} className="hover:bg-zinc-800/30">
                  <td className="px-5 py-3.5 font-medium text-zinc-200 text-sm">{ch.name || (ch.type === 'chat' ? 'Chat' : 'Telegram')}</td>
                  <td className="px-5 py-3.5">
                    <span className="px-1.5 py-0.5 text-[11px] font-medium rounded-full bg-blue-500/15 text-blue-400 uppercase">{ch.type}</span>
                  </td>
                  <td className="px-5 py-3.5 text-zinc-400 text-sm">{modelName(ch.modelId)}</td>
                  <td className="px-5 py-3.5 text-zinc-500 text-sm">{ch.token ? 'set' : '—'}</td>
                  <td className="px-5 py-3.5 text-zinc-500 text-sm">{(ch.allowedUserIds ?? []).length || '—'}</td>
                  <td className="px-5 py-3.5">
                    <div className="flex gap-1 justify-end">
                      <button onClick={() => openEdit(ch)} className="p-1.5 rounded-md text-zinc-600 hover:text-teal-400 hover:bg-teal-500/10">
                        <Pencil size={14} />
                      </button>
                      {ch.id !== 'chat' && (
                        <button onClick={() => remove(ch.id)} className="p-1.5 rounded-md text-zinc-600 hover:text-red-400 hover:bg-red-500/10">
                          <Trash2 size={14} />
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <Modal open={modalOpen} onClose={() => setModalOpen(false)} title={editing ? 'Edit Channel' : 'New Telegram Channel'}>
        <div className="space-y-4">
          <div>
            <label className="block text-xs font-medium text-zinc-400 mb-1.5">Name</label>
            <input
              value={form.name}
              onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
              className="w-full px-3 py-2 border border-zinc-700 rounded-lg text-sm bg-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-teal-500/50"
              placeholder={editing?.type === 'chat' ? 'Chat' : 'Telegram'}
            />
          </div>

          {editing?.type !== 'chat' && (
            <div>
              <label className="block text-xs font-medium text-zinc-400 mb-1.5">Token</label>
              <input
                value={form.token}
                onChange={e => setForm(f => ({ ...f, token: e.target.value }))}
                className="w-full px-3 py-2 border border-zinc-700 rounded-lg text-sm bg-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-teal-500/50"
                placeholder="123:ABC..."
              />
            </div>
          )}

          <div>
            <label className="block text-xs font-medium text-zinc-400 mb-1.5">Model</label>
            <select
              value={form.modelId}
              onChange={e => setForm(f => ({ ...f, modelId: e.target.value }))}
              className="w-full px-3 py-2 border border-zinc-700 rounded-lg text-sm bg-zinc-800 text-zinc-100 focus:outline-none focus:border-teal-500/50"
            >
              <option value="">Select model...</option>
              {models.map(m => (
                <option key={m.id} value={m.id}>{m.name} ({m.connectionId})</option>
              ))}
            </select>
          </div>

          {editing?.type !== 'chat' && (
            <div>
              <label className="block text-xs font-medium text-zinc-400 mb-1.5">Allowed user IDs</label>
              <input
                value={form.allowedUserIds}
                onChange={e => setForm(f => ({ ...f, allowedUserIds: e.target.value }))}
                className="w-full px-3 py-2 border border-zinc-700 rounded-lg text-sm font-mono bg-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-teal-500/50"
                placeholder="12345,67890"
              />
            </div>
          )}

          <div className="flex justify-end gap-2 pt-2">
            <button onClick={() => setModalOpen(false)} className="px-3.5 py-2 text-xs font-medium text-zinc-400 bg-zinc-800 border border-zinc-700 rounded-lg hover:text-zinc-200 hover:bg-zinc-700">
              Cancel
            </button>
            <button
              onClick={submit}
              disabled={!form.modelId || (!editing && !form.token)}
              className="px-3.5 py-2 text-xs font-medium text-white bg-teal-600 rounded-lg hover:bg-teal-500 disabled:opacity-40"
            >
              {editing ? 'Update' : 'Create'}
            </button>
          </div>
        </div>
      </Modal>
    </div>
  )
}

