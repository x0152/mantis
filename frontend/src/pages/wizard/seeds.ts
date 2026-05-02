import type { Plan, Skill } from '@/types'

export const MIN_BALANCE_GNK = 0.1
export const DEFAULT_GONKA_NODE = 'http://node1.gonka.ai:8000'

type SeedSkill = Omit<Skill, 'id' | 'connectionId'>

export const SEED_SKILLS: Record<string, SeedSkill[]> = {
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
        properties: { query: { type: 'string', description: 'Search query' } },
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
      script: 'pw-screenshot {{if .full_page}}--full-page {{end}}"{{.url}}" /tmp/screenshot.png && echo "Screenshot saved to /tmp/screenshot.png" && echo "NEXT: call ssh_download_browser with remotePath /tmp/screenshot.png, then call send_file with the artifactId from the download result."',
    },
    {
      name: 'read_webpage',
      description: 'Extract clean text from a URL as Markdown. Use ONLY for reading a specific known URL, NOT for general searching — let the agent handle research tasks.',
      parameters: {
        type: 'object',
        properties: { url: { type: 'string', description: 'Page URL to read' } },
        required: ['url'],
      },
      script: 'jina-read "{{.url}}"',
    },
  ],
}

export const SEED_PLANS: Array<Omit<Plan, 'id'>> = [
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
        { id: 'n3', type: 'action', label: 'Send', prompt: 'Send the downloaded screenshot artifact to the chat using send_file. Use the artifact ID from the previous step.', position: { x: 250, y: 300 } },
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
