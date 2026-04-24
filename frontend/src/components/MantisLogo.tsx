import { useId } from 'react'

export type MantisLogoState = 'idle' | 'thinking' | 'typing' | 'working'

interface Props {
  size?: number
  className?: string
  title?: string
  accent?: boolean
  state?: MantisLogoState
  animated?: boolean
}

export function MantisLogo({
  size = 20,
  className,
  title = 'Mantis',
  accent = true,
  state = 'idle',
  animated = true,
}: Props) {
  const shieldId = useId()
  const glowId = useId()

  const rootClass = [
    animated ? 'mantis-logo' : undefined,
    className,
  ].filter(Boolean).join(' ')

  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 64 64"
      fill="none"
      role="img"
      aria-label={title}
      className={rootClass}
      data-state={animated ? state : undefined}
      xmlns="http://www.w3.org/2000/svg"
    >
      <title>{title}</title>
      <defs>
        <linearGradient id={shieldId} x1="32" y1="6" x2="32" y2="46" gradientUnits="userSpaceOnUse">
          <stop offset="0%" stopColor="currentColor" stopOpacity="0.38" />
          <stop offset="55%" stopColor="currentColor" stopOpacity="0.14" />
          <stop offset="100%" stopColor="currentColor" stopOpacity="0.04" />
        </linearGradient>
        <radialGradient id={glowId} cx="32" cy="26" r="9" gradientUnits="userSpaceOnUse">
          <stop offset="0%" stopColor="currentColor" stopOpacity="0.9" />
          <stop offset="100%" stopColor="currentColor" stopOpacity="0" />
        </radialGradient>
      </defs>

      <g className="mantis-antenna antenna-left">
        <path
          d="M28 12 L12 4"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
        />
        <circle cx="12" cy="4" r="1.9" fill="currentColor" className="mantis-node n1" />
      </g>
      <g className="mantis-antenna antenna-right">
        <path
          d="M36 12 L52 4"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
        />
        <circle cx="52" cy="4" r="1.9" fill="currentColor" className="mantis-node n2" />
      </g>

      <path
        d="M32 8 L46 20 L42 40 L32 46 L22 40 L18 20 Z"
        fill={`url(#${shieldId})`}
        stroke="currentColor"
        strokeWidth="2.4"
        strokeLinejoin="round"
      />

      {accent && (
        <circle cx="32" cy="26" r="6.2" fill={`url(#${glowId})`} className="mantis-core" />
      )}

      <circle
        cx="32"
        cy="26"
        r="6"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.6"
        className="mantis-ring"
      />

      <circle cx="27" cy="22" r="2.1" fill="currentColor" className="mantis-eye eye-left" />
      <circle cx="37" cy="22" r="2.1" fill="currentColor" className="mantis-eye eye-right" />

      <path
        d="M32 31 L32 38"
        stroke="currentColor"
        strokeWidth="2.4"
        strokeLinecap="round"
      />

      <path
        d="M22 40 C 11 46 5 54 9 62"
        stroke="currentColor"
        strokeWidth="2.4"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M42 40 C 53 46 59 54 55 62"
        stroke="currentColor"
        strokeWidth="2.4"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <circle cx="9" cy="62" r="1.9" fill="currentColor" className="mantis-node n3" />
      <circle cx="55" cy="62" r="1.9" fill="currentColor" className="mantis-node n4" />

      <path
        d="M32 46 L32 54"
        stroke="currentColor"
        strokeWidth="2.2"
        strokeLinecap="round"
      />
      <circle cx="32" cy="57" r="1.7" fill="currentColor" className="mantis-node n5" />
    </svg>
  )
}
