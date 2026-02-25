import { useState, useEffect, useCallback, useMemo } from 'react'
import { Save, RefreshCw, Check, AlertTriangle, Trash2, Plus, Brain } from 'lucide-react'
import CodeMirror from '@uiw/react-codemirror'
import { json, jsonParseLinter } from '@codemirror/lang-json'
import { linter } from '@codemirror/lint'
import { api } from '../api'
import type { Model } from '../types'

export default function ConfigPage() {
  const [value, setValue] = useState('')
  const [saving, setSaving] = useState(false)
  const [toast, setToast] = useState<{ type: 'success' | 'error'; text: string } | null>(null)
  const [models, setModels] = useState<Model[]>([])

  const [memoryEnabled, setMemoryEnabled] = useState(false)
  const [summaryModelId, setSummaryModelId] = useState('')
  const [userMemories, setUserMemories] = useState<string[]>([])
  const [newFact, setNewFact] = useState('')

  const isValid = useMemo(() => {
    try { JSON.parse(value); return true } catch { return false }
  }, [value])

  const showToast = (type: 'success' | 'error', text: string) => {
    setToast({ type, text })
    setTimeout(() => setToast(null), 3000)
  }

  const load = useCallback(async () => {
    try {
      const [cfg, m] = await Promise.all([api.config.get(), api.models.list()])
      setModels(m)
      const data = cfg.data as Record<string, unknown>
      setMemoryEnabled(!!data.memoryEnabled)
      setSummaryModelId((data.summaryModelId as string) || '')
      setUserMemories(Array.isArray(data.userMemories) ? (data.userMemories as string[]) : [])

      const { memoryEnabled: _me, summaryModelId: _sm, userMemories: _um, ...rest } = data
      setValue(JSON.stringify(rest, null, 2))
    } catch (e: unknown) {
      showToast('error', e instanceof Error ? e.message : 'Failed to load')
    }
  }, [])

  useEffect(() => { load() }, [load])

  const buildFullData = () => {
    let rest: Record<string, unknown> = {}
    try { rest = JSON.parse(value) } catch { /* */ }
    return {
      ...rest,
      memoryEnabled,
      summaryModelId,
      userMemories,
    }
  }

  const save = async () => {
    if (!isValid) { showToast('error', 'Fix JSON errors before saving'); return }
    try {
      setSaving(true)
      await api.config.update(buildFullData())
      showToast('success', 'Configuration saved')
    } catch (e: unknown) {
      showToast('error', e instanceof Error ? e.message : 'Save failed')
    } finally {
      setSaving(false)
    }
  }

  const format = () => {
    try {
      setValue(JSON.stringify(JSON.parse(value), null, 2))
    } catch {
      showToast('error', 'Cannot format — fix JSON errors first')
    }
  }

  const addFact = () => {
    const f = newFact.trim()
    if (!f) return
    setUserMemories(prev => [...prev, f])
    setNewFact('')
  }

  const removeFact = (idx: number) => {
    setUserMemories(prev => prev.filter((_, i) => i !== idx))
  }

  const extensions = useMemo(() => [json(), linter(jsonParseLinter())], [])

  return (
    <div className="p-6 max-w-5xl">
      <div className="flex items-center justify-between mb-5">
        <div>
          <h1 className="text-lg font-semibold text-zinc-100">Configuration</h1>
          <p className="text-xs text-zinc-600 mt-0.5">Global settings, memory, and raw configuration</p>
        </div>
        <div className="flex items-center gap-2">
          {!isValid && value.length > 0 && (
            <span className="inline-flex items-center gap-1.5 px-2.5 py-1.5 text-xs font-medium text-red-400 bg-red-500/10 border border-red-500/20 rounded-lg">
              <AlertTriangle size={12} /> Invalid JSON
            </span>
          )}
          <button onClick={load} className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-zinc-400 bg-zinc-800 border border-zinc-700 rounded-lg hover:text-zinc-200 hover:bg-zinc-700">
            <RefreshCw size={13} /> Reload
          </button>
          <button onClick={save} disabled={saving || !isValid} className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-white bg-teal-600 rounded-lg hover:bg-teal-500 disabled:opacity-40">
            <Save size={13} /> {saving ? 'Saving...' : 'Save'}
          </button>
        </div>
      </div>

      {toast && (
        <div className={`mb-4 flex items-center gap-2 px-3.5 py-2.5 rounded-lg text-xs font-medium ${
          toast.type === 'success' ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-red-500/10 text-red-400 border border-red-500/20'
        }`}>
          {toast.type === 'success' ? <Check size={14} /> : <AlertTriangle size={14} />}
          {toast.text}
        </div>
      )}

      <div className="bg-zinc-900 rounded-lg border border-zinc-800 p-4 mb-4">
        <div className="flex items-center gap-2.5 mb-4">
          <Brain size={16} className="text-violet-400" />
          <h2 className="text-sm font-semibold text-zinc-200">Memory</h2>
        </div>

        <div className="space-y-4">
          <label className={`flex items-center gap-2.5 p-2.5 rounded-lg border cursor-pointer ${
            memoryEnabled ? 'border-teal-500/40 bg-teal-500/5' : 'border-zinc-800 bg-zinc-950'
          }`}>
            <input type="checkbox" checked={memoryEnabled} onChange={e => setMemoryEnabled(e.target.checked)}
              className="rounded border-zinc-600 bg-zinc-800 text-teal-500 focus:ring-teal-500/30 focus:ring-offset-0" />
            <div>
              <span className="text-sm font-medium text-zinc-200">Enable Memory</span>
              <p className="text-[11px] text-zinc-500 mt-0.5">Auto-extract facts from conversations and inject them into context</p>
            </div>
          </label>

          {memoryEnabled && (
            <>
              <div>
                <label className="block text-xs font-medium text-zinc-400 mb-1.5">Summary Model</label>
                <p className="text-[11px] text-zinc-600 mb-1.5">Model used for extracting facts from conversations (can be a cheap/fast model)</p>
                <select
                  value={summaryModelId}
                  onChange={e => setSummaryModelId(e.target.value)}
                  className="w-full px-3 py-2 border border-zinc-700 rounded-lg text-sm bg-zinc-800 text-zinc-100 focus:outline-none focus:border-teal-500/50"
                >
                  <option value="">Not set (memory extraction disabled)</option>
                  {models.map(m => (
                    <option key={m.id} value={m.id}>{m.name} ({m.connectionId})</option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-xs font-medium text-zinc-400 mb-2">User Facts ({userMemories.length})</label>
                {userMemories.length > 0 && (
                  <div className="space-y-1.5 mb-3">
                    {userMemories.map((fact, i) => (
                      <div key={i} className="flex items-start gap-2.5 bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2.5">
                        <span className="text-violet-400 mt-0.5 text-xs">•</span>
                        <span className="flex-1 text-sm text-zinc-300">{fact}</span>
                        <button onClick={() => removeFact(i)} className="p-1 rounded-md text-zinc-700 hover:text-red-400 hover:bg-red-500/10 shrink-0">
                          <Trash2 size={12} />
                        </button>
                      </div>
                    ))}
                  </div>
                )}
                <div className="flex gap-2">
                  <input
                    value={newFact}
                    onChange={e => setNewFact(e.target.value)}
                    onKeyDown={e => e.key === 'Enter' && addFact()}
                    className="flex-1 px-3 py-2 border border-zinc-700 rounded-lg text-sm bg-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-teal-500/50"
                    placeholder="Add a fact about yourself..."
                  />
                  <button onClick={addFact} disabled={!newFact.trim()} className="inline-flex items-center gap-1.5 px-3 py-2 text-xs font-medium text-white bg-teal-600 rounded-lg hover:bg-teal-500 disabled:opacity-40">
                    <Plus size={14} />
                  </button>
                </div>
              </div>
            </>
          )}
        </div>
      </div>

      <div className="bg-zinc-900 rounded-lg border border-zinc-800 overflow-hidden">
        <div className="px-4 py-2.5 border-b border-zinc-800 flex items-center justify-between">
          <span className="text-[11px] font-medium text-zinc-500 uppercase tracking-wider">Advanced (JSON)</span>
          <div className="flex items-center gap-2">
            <button onClick={format} className="text-[11px] text-zinc-600 hover:text-zinc-400">Format</button>
            <span className="text-[11px] text-zinc-600">{value.split('\n').length} lines</span>
          </div>
        </div>
        <CodeMirror
          value={value}
          onChange={setValue}
          extensions={extensions}
          height="400px"
          theme="dark"
          basicSetup={{
            lineNumbers: true,
            foldGutter: true,
            bracketMatching: true,
            closeBrackets: true,
            highlightActiveLine: true,
            indentOnInput: true,
          }}
          className="text-sm"
        />
      </div>
    </div>
  )
}
