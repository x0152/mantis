import { useState, useEffect, useCallback, useMemo } from 'react'
import { Save, RefreshCw, Check, AlertTriangle } from 'lucide-react'
import CodeMirror from '@uiw/react-codemirror'
import { json, jsonParseLinter } from '@codemirror/lang-json'
import { linter } from '@codemirror/lint'
import { api } from '../api'

export default function ConfigPage() {
  const [value, setValue] = useState('')
  const [saving, setSaving] = useState(false)
  const [toast, setToast] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  const isValid = useMemo(() => {
    try { JSON.parse(value); return true } catch { return false }
  }, [value])

  const showToast = (type: 'success' | 'error', text: string) => {
    setToast({ type, text })
    setTimeout(() => setToast(null), 3000)
  }

  const load = useCallback(async () => {
    try {
      const cfg = await api.config.get()
      setValue(JSON.stringify(cfg.data, null, 2))
    } catch (e: unknown) {
      showToast('error', e instanceof Error ? e.message : 'Failed to load')
    }
  }, [])

  useEffect(() => { load() }, [load])

  const save = async () => {
    if (!isValid) { showToast('error', 'Fix JSON errors before saving'); return }
    try {
      setSaving(true)
      await api.config.update(JSON.parse(value))
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

  const extensions = useMemo(() => [json(), linter(jsonParseLinter())], [])

  return (
    <div className="p-6 max-w-5xl">
      <div className="flex items-center justify-between mb-5">
        <div>
          <h1 className="text-lg font-semibold text-zinc-100">Configuration</h1>
          <p className="text-xs text-zinc-600 mt-0.5">Global application configuration stored as JSON</p>
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
          <button onClick={format} className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-zinc-400 bg-zinc-800 border border-zinc-700 rounded-lg hover:text-zinc-200 hover:bg-zinc-700">
            Format
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

      <div className="bg-zinc-900 rounded-lg border border-zinc-800 overflow-hidden">
        <div className="px-4 py-2.5 border-b border-zinc-800 flex items-center justify-between">
          <span className="text-[11px] font-medium text-zinc-500 uppercase tracking-wider">JSON Editor</span>
          <span className="text-[11px] text-zinc-600">{value.split('\n').length} lines</span>
        </div>
        <CodeMirror
          value={value}
          onChange={setValue}
          extensions={extensions}
          height="600px"
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
