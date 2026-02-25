import { useState } from 'react'
import { api } from '../api'

const SANDBOXES = [
  { name: 'base', host: 'sandbox', port: 2222, description: 'General-purpose Linux sandbox — shell, files, networking utilities' },
  { name: 'browser', host: 'browser-sandbox', port: 22, description: 'Headless Chromium + Playwright — web navigation, screenshots, PDF, parsing' },
  { name: 'ffmpeg', host: 'ffmpeg-sandbox', port: 22, description: 'FFmpeg + MediaInfo + ImageMagick — video, audio, image processing' },
  { name: 'python', host: 'python-sandbox', port: 22, description: 'Python 3 sandbox — scripts, data analysis, pip packages' },
  { name: 'db', host: 'db-sandbox', port: 22, description: 'PostgreSQL client — psql, pg_dump, pg_restore' },
]

export default function SetupWizard({ onDone }: { onDone: () => void }) {
  const [baseUrl, setBaseUrl] = useState(import.meta.env.VITE_LLM_BASE_URL ?? '')
  const [apiKey, setApiKey] = useState(import.meta.env.VITE_LLM_API_KEY ?? '')
  const [modelName, setModelName] = useState(import.meta.env.VITE_LLM_MODEL ?? '')
  const [tgToken, setTgToken] = useState(import.meta.env.VITE_TG_BOT_TOKEN ?? '')
  const [tgUserIds, setTgUserIds] = useState(import.meta.env.VITE_TG_USER_IDS ?? '')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const submit = async () => {
    if (!baseUrl.trim() || !modelName.trim()) {
      setError('Base URL and model name are required')
      return
    }
    setLoading(true)
    setError('')
    try {
      await api.llmConnections.create({ id: 'default', provider: 'openai', baseUrl: baseUrl.trim(), apiKey: apiKey.trim() })

      const model = await api.models.create('default', modelName.trim(), '')

      for (const sb of SANDBOXES) {
        await api.connections.create({
          type: 'ssh',
          name: sb.name,
          description: sb.description,
          modelId: model.id,
          config: { host: sb.host, port: sb.port, username: 'mantis', password: 'mantis' },
        })
      }

      await api.channels.update('chat', { name: 'Chat', token: '', modelId: model.id, allowedUserIds: [] })

      const cronConfig: Record<string, string> = { model_id: model.id }
      const token = tgToken.trim()
      if (token) {
        const ids = tgUserIds.trim()
          ? tgUserIds.split(',').map(s => parseInt(s.trim(), 10)).filter(n => !isNaN(n))
          : []
        await api.channels.create({ type: 'telegram', name: 'Telegram', token, modelId: model.id, allowedUserIds: ids })
        cronConfig.channel = 'telegram'
        if (ids.length) cronConfig.sender = ids[0].toString()
      }
      await api.config.update({ cron: cronConfig })

      onDone()
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Setup failed')
    } finally {
      setLoading(false)
    }
  }

  const inputCls = 'w-full px-3 py-2 border border-zinc-700 rounded-lg text-sm bg-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-teal-500/50'
  const labelCls = 'block text-xs font-medium text-zinc-400 mb-1.5'

  return (
    <div className="min-h-screen bg-zinc-950 flex items-center justify-center p-6">
      <div className="w-full max-w-lg">
        <div className="mb-8 text-center">
          <div className="flex items-center justify-center gap-2.5 mb-2">
            <div className="w-2.5 h-2.5 rounded-full bg-teal-400" />
            <h1 className="text-2xl font-semibold text-zinc-100 tracking-tight">Mantis</h1>
          </div>
          <p className="text-sm text-zinc-500">Connect your LLM provider to get started</p>
        </div>

        <div className="space-y-6 bg-zinc-900/60 border border-zinc-800/80 rounded-xl p-6">
          <div>
            <h2 className="text-sm font-semibold text-zinc-200 mb-3">LLM Connection</h2>
            <div className="space-y-3">
              <div>
                <label className={labelCls}>Base URL</label>
                <input value={baseUrl} onChange={e => setBaseUrl(e.target.value)} className={inputCls} placeholder="https://api.openai.com/v1" />
              </div>
              <div>
                <label className={labelCls}>API Key</label>
                <input type="password" value={apiKey} onChange={e => setApiKey(e.target.value)} className={inputCls} placeholder="sk-..." />
              </div>
              <div>
                <label className={labelCls}>Model name</label>
                <input value={modelName} onChange={e => setModelName(e.target.value)} className={inputCls} placeholder="gpt-4o-mini" />
              </div>
            </div>
          </div>

          <div>
            <h2 className="text-sm font-semibold text-zinc-200 mb-1">Telegram <span className="text-zinc-600 font-normal">(optional)</span></h2>
            <p className="text-xs text-zinc-600 mb-3">Connect a Telegram bot to chat from your phone</p>
            <div className="space-y-3">
              <div>
                <label className={labelCls}>Bot token</label>
                <input value={tgToken} onChange={e => setTgToken(e.target.value)} className={inputCls} placeholder="123456:ABC-DEF..." />
              </div>
              <div>
                <label className={labelCls}>Allowed user IDs <span className="text-zinc-600">(comma-separated, empty = allow all)</span></label>
                <input value={tgUserIds} onChange={e => setTgUserIds(e.target.value)} className={inputCls} placeholder="123456789, 987654321" />
              </div>
            </div>
          </div>

          {error && <p className="text-red-400 text-sm">{error}</p>}

          <button
            onClick={submit}
            disabled={loading}
            className="w-full py-2.5 bg-teal-500 hover:bg-teal-400 disabled:opacity-50 text-zinc-950 font-semibold text-sm rounded-lg transition-colors"
          >
            {loading ? 'Setting up...' : 'Complete Setup'}
          </button>
        </div>
      </div>
    </div>
  )
}
