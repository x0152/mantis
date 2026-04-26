import { useState } from 'react'
import { api } from '../api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card } from '@/components/ui/card'
import { FormField } from '@/components/FormField'
import { MantisLogo } from '@/components/MantisLogo'
import type { Skill, Plan } from '../types'

function normalizeBaseUrl(raw: string): string {
  const trimmed = raw.trim()
  if (!trimmed) return ''
  try {
    const u = new URL(trimmed)
    return `${u.protocol}//${u.host}${u.pathname}`.replace(/\/+$/, '').toLowerCase()
  } catch {
    return trimmed.replace(/\/+$/, '').toLowerCase()
  }
}

function slugify(input: string): string {
  const s = input
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+/, '')
    .replace(/-+$/, '')
  return s || 'llm-primary'
}

function suggestEndpointId(baseUrl: string, provider: string): string {
  if (provider === 'gonka') return 'gonka-primary'
  try {
    const u = new URL(baseUrl.trim())
    const hostSlug = slugify(`${u.hostname}${u.port ? `-${u.port}` : ''}`)
    if (hostSlug.includes('openai')) return 'openai-primary'
    if (hostSlug.includes('localhost') || hostSlug.startsWith('127-')) return 'local-llm'
    return `llm-${hostSlug}`.slice(0, 48)
  } catch {
    return 'llm-primary'
  }
}

async function pollConnections(names: string[], timeoutMs: number) {
  const deadline = Date.now() + timeoutMs
  while (Date.now() < deadline) {
    const list = await api.connections.list()
    const have = new Set(list.map(c => c.name))
    if (names.every(n => have.has(n))) return list
    await new Promise(r => setTimeout(r, 500))
  }
  return api.connections.list()
}

function makeUniqueID(base: string, existing: Set<string>): string {
  if (!existing.has(base)) return base
  for (let i = 2; i < 1000; i++) {
    const candidate = `${base}-${i}`.slice(0, 48)
    if (!existing.has(candidate)) return candidate
  }
  return `${base}-${Date.now()}`.slice(0, 48)
}

type SeedSkill = Omit<Skill, 'id' | 'connectionId'>

const SEED_SKILLS: Record<string, SeedSkill[]> = {
  base: [
    {
      name: 'http_health_check',
      description: 'Check if a URL is reachable. Returns HTTP status code and response time.',
      parameters: {
        type: 'object',
        properties: {
          url: { type: 'string', description: 'URL to check' },
          expected_status: { type: 'integer', description: 'Expected HTTP status code (default 200)' },
        },
        required: ['url'],
      },
      script: 'curl -sS -o /dev/null -w "status=%{http_code} time=%{time_total}s" --connect-timeout 10 --max-time 30 "{{.url}}"',
    },
    {
      name: 'system_health',
      description: 'Server health snapshot: CPU load, memory, disk usage, top processes.',
      parameters: { type: 'object', properties: {} },
      script: 'echo "=== Uptime ===" && uptime && echo "\\n=== Memory ===" && free -h && echo "\\n=== Disk ===" && df -h / && echo "\\n=== Top processes ===" && ps aux --sort=-%mem | head -6',
    },
    {
      name: 'find_large_files',
      description: 'Find files larger than a given size in a directory.',
      parameters: {
        type: 'object',
        properties: {
          path: { type: 'string', description: 'Directory to search in' },
          min_size_mb: { type: 'integer', description: 'Minimum file size in MB (default 100)' },
        },
        required: ['path'],
      },
      script: 'find "{{.path}}" -type f -size +{{if .min_size_mb}}{{.min_size_mb}}{{else}}100{{end}}M -exec ls -lh {} \\; 2>/dev/null | sort -k5 -hr | head -20',
    },
  ],
  browser: [
    {
      name: 'web_search',
      description: 'Search the web via DuckDuckGo. Returns titles, URLs, and snippets.',
      parameters: {
        type: 'object',
        properties: {
          query: { type: 'string', description: 'Search query' },
        },
        required: ['query'],
      },
      script: 'web-search \'{{.query}}\'',
    },
    {
      name: 'screenshot',
      description: 'Take a screenshot of a web page using Playwright.',
      parameters: {
        type: 'object',
        properties: {
          url: { type: 'string', description: 'Page URL to screenshot' },
          full_page: { type: 'boolean', description: 'Capture full scrollable page' },
        },
        required: ['url'],
      },
      script: 'pw-screenshot {{if .full_page}}--full-page {{end}}"{{.url}}" /tmp/screenshot.png && echo "Screenshot saved to /tmp/screenshot.png"',
    },
    {
      name: 'read_webpage',
      description: 'Extract clean text from a URL as Markdown. Use ONLY for reading a specific known URL, NOT for general searching — let the agent handle research tasks.',
      parameters: {
        type: 'object',
        properties: {
          url: { type: 'string', description: 'Page URL to read' },
        },
        required: ['url'],
      },
      script: 'jina-read "{{.url}}"',
    },
  ],
}

const SEED_PLANS: Array<Omit<Plan, 'id'>> = [
  {
    name: 'Screenshot',
    description: 'Take a screenshot of a web page and send it to the chat.',
    schedule: '',
    enabled: true,
    parameters: {
      type: 'object',
      properties: {
        url: { type: 'string', description: 'Full URL of the page to screenshot (e.g. https://example.com)' },
      },
    },
    graph: {
      nodes: [
        { id: 'n1', type: 'action', label: 'Screenshot', prompt: 'Use the screenshot skill on the browser connection to take a screenshot of "{{.url}}".', position: { x: 250, y: 0 } },
        { id: 'n2', type: 'action', label: 'Download', prompt: 'Download the screenshot file from the browser server. The file was saved to /tmp/screenshot.png — use ssh_download_browser with remotePath "/tmp/screenshot.png".', position: { x: 250, y: 150 } },
        { id: 'n3', type: 'action', label: 'Send', prompt: 'Send the downloaded screenshot artifact to the chat using artifact_send_to_chat. Use the artifact ID from the previous step.', position: { x: 250, y: 300 } },
      ],
      edges: [
        { id: 'e1', source: 'n1', target: 'n2', label: '' },
        { id: 'e2', source: 'n2', target: 'n3', label: '' },
      ],
    },
  },
  {
    name: 'Morning Server Report',
    description: 'Check server health and send a notification with the status.',
    schedule: '',
    enabled: false,
    parameters: {},
    graph: {
      nodes: [
        { id: 'n1', type: 'action', label: 'Check health', prompt: 'Run the system_health skill on the base connection. Analyze the output: note CPU load average, memory usage percentage, and disk usage percentage.', position: { x: 250, y: 0 } },
        { id: 'n2', type: 'decision', label: 'Any issues?', prompt: 'Based on the health check results, are there any problems? Answer YES if: load average > 2.0, memory usage > 80%, or disk usage > 85%. Answer NO otherwise.', position: { x: 250, y: 150 } },
        { id: 'n3', type: 'action', label: 'Alert', prompt: 'Send a notification via send_notification describing the problems found: which metrics are above thresholds and their current values.', position: { x: 50, y: 300 } },
        { id: 'n4', type: 'action', label: 'All OK', prompt: 'Send a short notification via send_notification saying all systems are running normally.', position: { x: 450, y: 300 } },
      ],
      edges: [
        { id: 'e1', source: 'n1', target: 'n2', label: '' },
        { id: 'e2', source: 'n2', target: 'n3', label: 'yes' },
        { id: 'e3', source: 'n2', target: 'n4', label: 'no' },
      ],
    },
  },
  {
    name: 'Research Assistant',
    description: 'Search the web for a given topic, read top articles, and send a summary digest.',
    schedule: '',
    enabled: false,
    parameters: {
      type: 'object',
      properties: {
        topic: { type: 'string', description: 'The topic to research (e.g. "latest Kubernetes news")' },
      },
    },
    graph: {
      nodes: [
        { id: 'n1', type: 'action', label: 'Search', prompt: 'Use the web_search skill on the browser connection to search for "{{.topic}}". Return the top 5 results with titles and URLs.', position: { x: 250, y: 0 } },
        { id: 'n2', type: 'action', label: 'Read articles', prompt: 'Take the first 2 URLs from the search results. Use the read_webpage skill to get the content. Summarize the key points about "{{.topic}}".', clearContext: true, position: { x: 250, y: 150 } },
        { id: 'n3', type: 'action', label: 'Send digest', prompt: 'Compile a brief digest from the article summaries about "{{.topic}}". Send it via send_notification.', position: { x: 250, y: 300 } },
      ],
      edges: [
        { id: 'e1', source: 'n1', target: 'n2', label: '' },
        { id: 'e2', source: 'n2', target: 'n3', label: '' },
      ],
    },
  },
  {
    name: 'Restart Service',
    description: 'Restart a system service, verify it is running, and send an alert with the result.',
    schedule: '',
    enabled: false,
    parameters: {
      type: 'object',
      properties: {
        service_name: { type: 'string', description: 'Name of the systemd service (e.g. nginx, docker)' },
      },
    },
    graph: {
      nodes: [
        { id: 'n1', type: 'action', label: 'Restart', prompt: 'On the base server, restart the "{{.service_name}}" service and then check if it is active.', position: { x: 250, y: 0 } },
        { id: 'n2', type: 'decision', label: 'Is active?', prompt: 'Based on the output, is the {{.service_name}} service active and running? Answer YES or NO.', position: { x: 250, y: 150 } },
        { id: 'n3', type: 'action', label: 'Success', prompt: 'Send a notification via send_notification saying that {{.service_name}} was successfully restarted and is running.', position: { x: 50, y: 300 } },
        { id: 'n4', type: 'action', label: 'Failure', prompt: 'Send an URGENT notification via send_notification saying that {{.service_name}} failed to restart and needs manual intervention.', position: { x: 450, y: 300 } },
      ],
      edges: [
        { id: 'e1', source: 'n1', target: 'n2', label: '' },
        { id: 'e2', source: 'n2', target: 'n3', label: 'yes' },
        { id: 'e3', source: 'n2', target: 'n4', label: 'no' },
      ],
    },
  },
]

export default function SetupWizard({ onDone }: { onDone: () => void }) {
  const [provider, setProvider] = useState<'openai' | 'gonka'>('openai')
  const [baseUrl, setBaseUrl] = useState(import.meta.env.VITE_LLM_BASE_URL ?? '')
  const [apiKey, setApiKey] = useState(import.meta.env.VITE_LLM_API_KEY ?? '')
  const envModels = (import.meta.env.VITE_LLM_MODEL ?? '').split(',').map((s: string) => s.trim()).filter(Boolean)
  const [modelRows, setModelRows] = useState<{ name: string; role: 'chat' | 'summary' | 'vision' | '' }[]>(() => {
    if (envModels.length === 0) return [{ name: '', role: 'chat' }]
    return envModels.map((name: string, i: number) => ({
      name,
      role: i === 0 ? 'chat' as const : i === envModels.length - 1 && envModels.length > 1 ? 'summary' as const : '' as const,
    }))
  })
  const [tgToken, setTgToken] = useState(import.meta.env.VITE_TG_BOT_TOKEN ?? '')
  const [tgUserIds, setTgUserIds] = useState(import.meta.env.VITE_TG_USER_IDS ?? '')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const chatRow = modelRows.find(r => r.role === 'chat')
  const hasChatModel = chatRow && chatRow.name.trim()

  const submit = async () => {
    if (!baseUrl.trim() || !hasChatModel) {
      setError('Base URL and at least one model marked as "chat" are required')
      return
    }
    if (provider === 'gonka' && !apiKey.trim()) {
      setError('Private key is required for Gonka provider')
      return
    }
    setLoading(true)
    setError('')
    try {
      const cleanBase = baseUrl.trim()
      const existingEndpoints = await api.llmConnections.list()
      const existingEndpointByURL = existingEndpoints.find(e => normalizeBaseUrl(e.baseUrl) === normalizeBaseUrl(cleanBase))
      const endpointID = existingEndpointByURL
        ? existingEndpointByURL.id
        : makeUniqueID(suggestEndpointId(cleanBase, provider), new Set(existingEndpoints.map(e => e.id)))

      if (existingEndpointByURL) {
        await api.llmConnections.update(endpointID, { provider, baseUrl: cleanBase, apiKey: apiKey.trim() })
      } else {
        await api.llmConnections.create({ id: endpointID, provider, baseUrl: cleanBase, apiKey: apiKey.trim() })
      }

      const existing = await api.models.list()
      const findOrCreate = async (name: string) => {
        const found = existing.find(m => m.connectionId === endpointID && m.name === name)
        return found ?? await api.models.create({
          connectionId: endpointID,
          name,
          thinkingMode: '',
          contextWindow: 0,
          reserveTokens: 0,
          compactTokens: 0,
        })
      }

      const validRows = modelRows.filter(r => r.name.trim())
      const created: Record<string, { id: string }> = {}
      for (const row of validRows) {
        created[row.name.trim()] = await findOrCreate(row.name.trim())
      }

      const chatModelRow = validRows.find(r => r.role === 'chat')!
      const summaryModelRow = validRows.find(r => r.role === 'summary')
      const visionModelRow = validRows.find(r => r.role === 'vision')
      const primaryModelId = created[chatModelRow.name.trim()].id
      const summaryModelId = summaryModelRow ? created[summaryModelRow.name.trim()].id : ''
      const imageModelId = visionModelRow ? created[visionModelRow.name.trim()].id : ''

      const profileName = `Main profile (${endpointID})`
      const existingPresets = await api.presets.list()
      const existingPreset = existingPresets.find(p => p.name === profileName) ?? existingPresets.find(p => p.name === 'Default')
      const preset = existingPreset
        ? await api.presets.update(existingPreset.id, {
            name: profileName,
            chatModelId: primaryModelId,
            summaryModelId,
            imageModelId,
            fallbackModelId: '',
            temperature: null,
            systemPrompt: '',
          })
        : await api.presets.create({
            name: profileName,
            chatModelId: primaryModelId,
            summaryModelId,
            imageModelId,
            fallbackModelId: '',
            temperature: null,
            systemPrompt: '',
          })

      const connList = await pollConnections(['base', 'browser'], 15_000).catch(() => [] as Awaited<ReturnType<typeof api.connections.list>>)
      const connByName = new Map(connList.map(c => [c.name, c]))
      for (const profileId of Object.keys(SEED_SKILLS)) {
        const conn = connByName.get(profileId)
        if (!conn) continue
        for (const sk of SEED_SKILLS[profileId]) {
          await api.skills.create({ ...sk, connectionId: conn.id }).catch(() => {})
        }
      }

      for (const plan of SEED_PLANS) {
        await api.plans.create(plan).catch(() => {})
      }

      await api.channels.update('chat', { name: 'Chat', token: '', allowedUserIds: [] })

      const token = tgToken.trim()
      if (token) {
        const ids = tgUserIds.trim()
          ? tgUserIds.split(',').map((s: string) => parseInt(s.trim(), 10)).filter((n: number) => !isNaN(n))
          : []
        await api.channels.create({ type: 'telegram', name: 'Telegram', token, modelId: '', presetId: '', allowedUserIds: ids })
          .catch(() => {})
      }
      await api.settings.update({
        chatPresetId: preset.id,
        serverPresetId: preset.id,
        memoryEnabled: true,
        userMemories: [],
      })

      onDone()
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Setup failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-zinc-50 dark:bg-zinc-950 flex items-center justify-center p-6">
      <div className="w-full max-w-lg">
        <div className="mb-8 flex items-center justify-center gap-3">
          <MantisLogo size={48} className="text-teal-500 dark:text-teal-400" />
          <div className="leading-tight">
            <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-100 tracking-tight">Mantis</h1>
            <p className="text-sm text-zinc-500 mt-0.5">Connect your LLM provider to get started</p>
          </div>
        </div>

        <Card className="space-y-6 bg-white dark:bg-zinc-900/60 border-zinc-200 dark:border-zinc-800/80 p-6">
          <div>
            <h2 className="text-sm font-semibold text-zinc-800 dark:text-zinc-200 mb-3">LLM Connection</h2>
            <div className="space-y-3">
              <FormField label="Provider">
                <select
                  value={provider}
                  onChange={e => setProvider(e.target.value as 'openai' | 'gonka')}
                  className="flex w-full rounded-lg border border-zinc-300 dark:border-zinc-700 bg-white dark:bg-zinc-800 px-3 py-2 text-sm text-zinc-900 dark:text-zinc-100 focus:outline-none focus:border-teal-500/50"
                >
                  <option value="openai">openai</option>
                  <option value="gonka">gonka</option>
                </select>
              </FormField>
              <FormField
                label={provider === 'gonka' ? 'Node URL' : 'Base URL'}
                hint={provider === 'gonka' ? 'Gonka Source URL (NODE_URL)' : undefined}
              >
                <Input
                  value={baseUrl}
                  onChange={e => setBaseUrl(e.target.value)}
                  placeholder={provider === 'gonka' ? 'https://api.gonka.testnet.example.com' : 'https://api.openai.com/v1'}
                />
              </FormField>
              <FormField
                label={provider === 'gonka' ? 'Private key' : 'API Key'}
                hint={provider === 'gonka' ? 'GONKA_PRIVATE_KEY wallet key' : undefined}
              >
                <Input
                  type="password"
                  value={apiKey}
                  onChange={e => setApiKey(e.target.value)}
                  placeholder={provider === 'gonka' ? '0x...' : 'sk-...'}
                />
              </FormField>
              <div>
                <Label>Models</Label>
                <div className="space-y-2 mt-1.5">
                  {modelRows.map((row, i) => (
                    <div key={i} className="flex items-center gap-2">
                      <Input
                        value={row.name}
                        onChange={e => setModelRows(prev => prev.map((r, j) => j === i ? { ...r, name: e.target.value } : r))}
                        placeholder="model name"
                        className="flex-1"
                      />
                      <select
                        value={row.role}
                        onChange={e => {
                          const newRole = e.target.value as 'chat' | 'summary' | 'vision' | ''
                          setModelRows(prev => prev.map((r, j) => {
                            if (j === i) return { ...r, role: newRole }
                            if (newRole && r.role === newRole) return { ...r, role: '' }
                            return r
                          }))
                        }}
                        className="w-28 shrink-0 rounded-lg border border-zinc-300 dark:border-zinc-700 bg-white dark:bg-zinc-800 px-2 py-2 text-xs text-zinc-700 dark:text-zinc-300 focus:outline-none focus:border-teal-500/50"
                      >
                        <option value="">—</option>
                        <option value="chat">chat</option>
                        <option value="summary">summary</option>
                        <option value="vision">vision</option>
                      </select>
                      {modelRows.length > 1 && (
                        <button
                          type="button"
                          onClick={() => setModelRows(prev => prev.filter((_, j) => j !== i))}
                          className="text-zinc-400 hover:text-red-400 text-sm px-1"
                        >×</button>
                      )}
                    </div>
                  ))}
                  <button
                    type="button"
                    onClick={() => setModelRows(prev => [...prev, { name: '', role: '' }])}
                    className="text-[12px] text-teal-600 dark:text-teal-400 hover:underline"
                  >+ Add model</button>
                </div>
                <p className="text-[11px] text-zinc-500 dark:text-zinc-600 mt-1.5">Mark one as <span className="font-medium">chat</span> (required). Optionally: <span className="font-medium">summary</span> for titles & memory, <span className="font-medium">vision</span> for image understanding.</p>
              </div>
            </div>
          </div>

          <div>
            <h2 className="text-sm font-semibold text-zinc-800 dark:text-zinc-200 mb-1">Telegram <span className="text-zinc-500 dark:text-zinc-600 font-normal">(optional)</span></h2>
            <p className="text-xs text-zinc-500 dark:text-zinc-600 mb-3">Connect a Telegram bot to chat from your phone</p>
            <div className="space-y-3">
              <FormField label="Bot token">
                <Input value={tgToken} onChange={e => setTgToken(e.target.value)} placeholder="123456:ABC-DEF..." />
              </FormField>
              <div>
                <Label>Allowed user IDs <span className="text-zinc-500 dark:text-zinc-600">(comma-separated, empty = allow all)</span></Label>
                <Input value={tgUserIds} onChange={e => setTgUserIds(e.target.value)} placeholder="123456789, 987654321" />
              </div>
            </div>
          </div>

          {error && <p className="text-red-400 text-sm">{error}</p>}

          <Button
            onClick={submit}
            disabled={loading}
            className="w-full py-2.5 bg-teal-500 hover:bg-teal-400 text-zinc-950 font-semibold text-sm"
          >
            {loading ? 'Setting up...' : 'Complete Setup'}
          </Button>
        </Card>
      </div>
    </div>
  )
}
