import { useId } from 'react'

interface Props {
  size?: number
  className?: string
  title?: string
  active?: boolean
}

interface NodeDef {
  id: string
  x: number
  y: number
  r?: number
}

interface EdgeDef {
  id: string
  from: string
  to: string
  pulse?: { dur: string; begin: string }
}

const NODES: NodeDef[] = [
  { id: 'n1', x: 24, y: 19, r: 2.2 },
  { id: 'n2', x: 32, y: 16, r: 2.4 },
  { id: 'n3', x: 40, y: 19, r: 2.2 },
  { id: 'n4', x: 19, y: 30, r: 2.2 },
  { id: 'n5', x: 32, y: 28, r: 2.6 },
  { id: 'n6', x: 45, y: 30, r: 2.2 },
  { id: 'n7', x: 32, y: 42, r: 2.4 },
]

const EDGES: EdgeDef[] = [
  { id: 'e1', from: 'n1', to: 'n2' },
  { id: 'e2', from: 'n2', to: 'n3', pulse: { dur: '1.7s', begin: '0s' } },
  { id: 'e3', from: 'n1', to: 'n4' },
  { id: 'e4', from: 'n3', to: 'n6' },
  { id: 'e5', from: 'n4', to: 'n5', pulse: { dur: '1.5s', begin: '0.6s' } },
  { id: 'e6', from: 'n5', to: 'n6' },
  { id: 'e7', from: 'n2', to: 'n5', pulse: { dur: '1.9s', begin: '1.1s' } },
  { id: 'e8', from: 'n4', to: 'n7' },
  { id: 'e9', from: 'n6', to: 'n7', pulse: { dur: '1.6s', begin: '0.3s' } },
]

const BRAIN_OUTLINE = [
  'M 32 8',
  'C 38 5, 48 7, 50 16',
  'C 58 18, 60 28, 54 31',
  'C 60 35, 56 44, 49 44',
  'C 50 52, 39 55, 32 50',
  'C 25 55, 14 52, 15 44',
  'C 8 44, 4 35, 10 31',
  'C 4 28, 6 18, 14 16',
  'C 16 7, 26 5, 32 8',
  'Z',
].join(' ')

const FISSURE = 'M 32 9 C 31 22, 33 38, 32 51'

function nodeById(id: string): NodeDef {
  const n = NODES.find(node => node.id === id)
  if (!n) throw new Error(`unknown brain node: ${id}`)
  return n
}

export function BrainThinking({
  size = 20,
  className,
  title = 'Thinking',
  active = true,
}: Props) {
  const rawId = useId()
  const uid = rawId.replace(/[^a-zA-Z0-9_-]/g, '')
  const haloId = `brain-halo-${uid}`
  const fillId = `brain-fill-${uid}`

  const rootClass = [
    'brain-logo',
    active ? 'brain-active' : undefined,
    className,
  ]
    .filter(Boolean)
    .join(' ')

  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 64 64"
      fill="none"
      role="img"
      aria-label={title}
      className={rootClass}
      xmlns="http://www.w3.org/2000/svg"
      style={{ overflow: 'visible' }}
    >
      <title>{title}</title>
      <defs>
        <radialGradient
          id={haloId}
          cx="32"
          cy="32"
          r="28"
          gradientUnits="userSpaceOnUse"
        >
          <stop offset="0%" stopColor="currentColor" stopOpacity="0.32" />
          <stop offset="55%" stopColor="currentColor" stopOpacity="0.07" />
          <stop offset="100%" stopColor="currentColor" stopOpacity="0" />
        </radialGradient>
        <linearGradient
          id={fillId}
          x1="32"
          y1="8"
          x2="32"
          y2="55"
          gradientUnits="userSpaceOnUse"
        >
          <stop offset="0%" stopColor="currentColor" stopOpacity="0.18" />
          <stop offset="100%" stopColor="currentColor" stopOpacity="0.04" />
        </linearGradient>

        {EDGES.map(e => {
          const a = nodeById(e.from)
          const b = nodeById(e.to)
          return (
            <path
              key={`def-${e.id}`}
              id={`brain-edge-${e.id}-${uid}`}
              d={`M ${a.x} ${a.y} L ${b.x} ${b.y}`}
            />
          )
        })}
      </defs>

      <circle
        cx="32"
        cy="32"
        r="26"
        fill={`url(#${haloId})`}
        className="brain-halo"
      />

      <g className="brain-shell">
        <path
          d={BRAIN_OUTLINE}
          fill={`url(#${fillId})`}
          stroke="currentColor"
          strokeWidth="2"
          strokeLinejoin="round"
          className="brain-outline"
        />
        <path
          d={FISSURE}
          stroke="currentColor"
          strokeWidth="0.9"
          strokeLinecap="round"
          fill="none"
          opacity="0.32"
          className="brain-fissure"
        />
      </g>

      <g className="brain-graph">
        {EDGES.map(e => {
          const a = nodeById(e.from)
          const b = nodeById(e.to)
          return (
            <line
              key={`edge-${e.id}`}
              x1={a.x}
              y1={a.y}
              x2={b.x}
              y2={b.y}
              stroke="currentColor"
              strokeWidth="1.1"
              strokeLinecap="round"
              opacity="0.42"
              className="brain-edge"
            />
          )
        })}
        {NODES.map(n => (
          <circle
            key={`node-${n.id}`}
            cx={n.x}
            cy={n.y}
            r={n.r ?? 2.2}
            fill="currentColor"
            className={`brain-node brain-node-${n.id}`}
          />
        ))}
      </g>

      {active && (
        <g className="brain-pulses">
          {EDGES.filter(e => e.pulse).map(e => (
            <circle
              key={`pulse-${e.id}`}
              r="1.9"
              fill="currentColor"
              opacity="0"
              className="brain-pulse"
            >
              <animate
                attributeName="opacity"
                values="0;1;1;0"
                keyTimes="0;0.18;0.82;1"
                dur={e.pulse!.dur}
                begin={e.pulse!.begin}
                repeatCount="indefinite"
              />
              <animateMotion
                dur={e.pulse!.dur}
                begin={e.pulse!.begin}
                repeatCount="indefinite"
                rotate="auto"
              >
                <mpath href={`#brain-edge-${e.id}-${uid}`} />
              </animateMotion>
            </circle>
          ))}
        </g>
      )}
    </svg>
  )
}
