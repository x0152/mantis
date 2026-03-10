import { useState, useEffect, useCallback, useMemo } from 'react'
import { Save, RefreshCw, AlertTriangle } from 'lucide-react'
import { toast } from 'sonner'
import CodeMirror from '@uiw/react-codemirror'
import { json, jsonParseLinter } from '@codemirror/lang-json'
import { linter } from '@codemirror/lint'
import { api } from '../api'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'

export default function ConfigPage() {
  const [value, setValue] = useState('')
  const [saving, setSaving] = useState(false)

  const isValid = useMemo(() => {
    try { JSON.parse(value); return true } catch { return false }
  }, [value])

  const load = useCallback(async () => {
    try {
      const cfg = await api.config.get()
      setValue(JSON.stringify(cfg.data, null, 2))
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Failed to load')
    }
  }, [])

  useEffect(() => { load() }, [load])

  const save = async () => {
    if (!isValid) { toast.error('Fix JSON errors before saving'); return }
    try {
      setSaving(true)
      await api.config.update(JSON.parse(value))
      toast.success('Configuration saved')
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Save failed')
    } finally {
      setSaving(false)
    }
  }

  const format = () => {
    try {
      setValue(JSON.stringify(JSON.parse(value), null, 2))
    } catch {
      toast.error('Cannot format — fix JSON errors first')
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
            <Badge variant="destructive" className="gap-1.5 px-2.5 py-1.5 text-xs rounded-lg border border-red-500/20">
              <AlertTriangle size={12} /> Invalid JSON
            </Badge>
          )}
          <Button variant="secondary" size="sm" onClick={load}>
            <RefreshCw size={13} /> Reload
          </Button>
          <Button variant="secondary" size="sm" onClick={format}>
            Format
          </Button>
          <Button size="sm" onClick={save} disabled={saving || !isValid}>
            <Save size={13} /> {saving ? 'Saving...' : 'Save'}
          </Button>
        </div>
      </div>

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
