import { useEffect, useId, useRef, useState } from 'react'

interface Props {
  size?: number
  className?: string
  title?: string
  active?: boolean
}

interface CellState {
  char: string
  lockedChar: string | null
}

const GLYPHS = (
  '01ｱｲｳｴｵｶｷｸｹｺｻｼｽｾｿﾀﾁﾂﾃﾄﾅﾆﾇﾈﾉ#%&*+=<>?/{}@$~^'
).split('')

const CELL_COUNT = 4
const CELL_X = [11, 23, 35, 47]
const CELL_Y = 35

const LOCK_TIMES = [480, 760, 1080, 1380]
const CYCLE_MS = 2400
const SCRAMBLE_INTERVAL = 70

const CARET_X = 54
const CARET_Y = 23
const CARET_H = 18

function rnd(): string {
  return GLYPHS[Math.floor(Math.random() * GLYPHS.length)]
}

function initialCells(): CellState[] {
  return Array.from({ length: CELL_COUNT }, () => ({
    char: GLYPHS[0],
    lockedChar: null,
  }))
}

export function TypingLines({
  size = 22,
  className,
  title = 'Typing',
  active = true,
}: Props) {
  const rawId = useId()
  const uid = rawId.replace(/[^a-zA-Z0-9_-]/g, '')
  const haloId = `typing-halo-${uid}`

  const [cells, setCells] = useState<CellState[]>(initialCells)
  const startRef = useRef(0)
  const lastCycleRef = useRef(-1)

  useEffect(() => {
    if (!active) return undefined

    startRef.current = performance.now()
    lastCycleRef.current = -1
    setCells(initialCells())

    const tick = () => {
      const elapsed = performance.now() - startRef.current
      const cycle = Math.floor(elapsed / CYCLE_MS)
      const phase = elapsed % CYCLE_MS
      const cycleChanged = cycle !== lastCycleRef.current
      lastCycleRef.current = cycle

      setCells(prev =>
        prev.map((cell, i) => {
          const lockAt = LOCK_TIMES[i]
          const carryLocked = cycleChanged ? null : cell.lockedChar

          if (phase < lockAt) {
            return { char: rnd(), lockedChar: null }
          }
          if (carryLocked === null) {
            const final = rnd()
            return { char: final, lockedChar: final }
          }
          return { char: carryLocked, lockedChar: carryLocked }
        }),
      )
    }

    tick()
    const id = window.setInterval(tick, SCRAMBLE_INTERVAL)
    return () => window.clearInterval(id)
  }, [active])

  const rootClass = [
    'typing-logo',
    active ? 'typing-active' : undefined,
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
          <stop offset="0%" stopColor="currentColor" stopOpacity="0.28" />
          <stop offset="55%" stopColor="currentColor" stopOpacity="0.06" />
          <stop offset="100%" stopColor="currentColor" stopOpacity="0" />
        </radialGradient>
      </defs>

      <circle
        cx="32"
        cy="32"
        r="26"
        fill={`url(#${haloId})`}
        className="typing-halo"
      />

      <line
        x1="6"
        y1="46"
        x2="58"
        y2="46"
        stroke="currentColor"
        strokeWidth="1"
        strokeLinecap="round"
        opacity="0.18"
        className="typing-baseline"
      />

      <g className="typing-stream">
        {cells.map((cell, i) => {
          const locked = cell.lockedChar !== null
          return (
            <text
              key={i}
              x={CELL_X[i]}
              y={CELL_Y}
              fill="currentColor"
              fontFamily="ui-monospace, 'SF Mono', SFMono-Regular, Menlo, Consolas, monospace"
              fontSize="14"
              fontWeight="600"
              textAnchor="middle"
              dominantBaseline="middle"
              className={`typing-glyph${locked ? ' is-locked' : ''}`}
            >
              {cell.char}
            </text>
          )
        })}
      </g>

      <rect
        x={CARET_X}
        y={CARET_Y}
        width="2"
        height={CARET_H}
        rx="1"
        ry="1"
        fill="currentColor"
        className="typing-caret-bar"
      />
    </svg>
  )
}
