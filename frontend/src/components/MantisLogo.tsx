import { useId } from 'react'

export type MantisLogoState = 'idle' | 'thinking' | 'typing' | 'working'

interface Props {
  size?: number
  className?: string
  title?: string
  state?: MantisLogoState
  animated?: boolean
}

export function MantisLogo({
  size = 20,
  className,
  title = 'Mantis',
  state = 'idle',
  animated = true,
}: Props) {
  const rawId = useId()
  const uid = rawId.replace(/[^a-zA-Z0-9_-]/g, '')
  const auraId = `mantis-aura-${uid}`
  const headId = `mantis-head-${uid}`
  const eyeClipId = `mantis-eye-clip-${uid}`

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
      style={{ overflow: 'visible' }}
    >
      <title>{title}</title>
      <defs>
        <radialGradient id={auraId} cx="32" cy="32" r="28" gradientUnits="userSpaceOnUse">
          <stop offset="0%" stopColor="currentColor" stopOpacity="0.32" />
          <stop offset="55%" stopColor="currentColor" stopOpacity="0.06" />
          <stop offset="100%" stopColor="currentColor" stopOpacity="0" />
        </radialGradient>
        <linearGradient id={headId} x1="32" y1="10" x2="32" y2="44" gradientUnits="userSpaceOnUse">
          <stop offset="0%" stopColor="currentColor" stopOpacity="0.34" />
          <stop offset="100%" stopColor="currentColor" stopOpacity="0.04" />
        </linearGradient>
        <clipPath id={eyeClipId}>
          <path d="M22 18 L28 24 L22 30 L17 24 Z M42 18 L47 24 L42 30 L36 24 Z" />
        </clipPath>
      </defs>

      <g className="mantis-aura-wrap">
        <circle cx="32" cy="32" r="26" fill={`url(#${auraId})`} className="mantis-aura" />
      </g>

      <path
        d="M27 13 Q 20 5 12 4"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
        fill="none"
        className="mantis-antenna"
      />
      <path
        d="M37 13 Q 44 5 52 4"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
        fill="none"
        className="mantis-antenna"
      />
      <circle cx="12" cy="4" r="1.9" fill="currentColor" className="mantis-tip tip-l" />
      <circle cx="52" cy="4" r="1.9" fill="currentColor" className="mantis-tip tip-r" />

      <g className="mantis-head-group">
        <path
          d="M32 10 L52 22 L46 38 L32 44 L18 38 L12 22 Z"
          fill={`url(#${headId})`}
          stroke="currentColor"
          strokeWidth="2.4"
          strokeLinejoin="round"
          className="mantis-head"
        />

        <path
          d="M22 18 L28 24 L22 30 L17 24 Z"
          fill="currentColor"
          fillOpacity="0.18"
          stroke="currentColor"
          strokeWidth="1.4"
          strokeLinejoin="round"
          className="mantis-eye eye-left"
        />
        <path
          d="M42 18 L47 24 L42 30 L36 24 Z"
          fill="currentColor"
          fillOpacity="0.18"
          stroke="currentColor"
          strokeWidth="1.4"
          strokeLinejoin="round"
          className="mantis-eye eye-right"
        />

        <circle cx="22" cy="24" r="1.6" fill="currentColor" className="mantis-pupil pupil-left" />
        <circle cx="42" cy="24" r="1.6" fill="currentColor" className="mantis-pupil pupil-right" />

        <g clipPath={`url(#${eyeClipId})`}>
          <rect x="14" y="-2" width="36" height="2.2" fill="currentColor" className="mantis-scan" />
        </g>
      </g>

      <path
        d="M27 42 L18 51 L28 57"
        fill="none"
        stroke="currentColor"
        strokeWidth="2.1"
        strokeLinecap="round"
        strokeLinejoin="round"
        className="mantis-arm arm-l"
      />
      <path
        d="M37 42 L46 51 L36 57"
        fill="none"
        stroke="currentColor"
        strokeWidth="2.1"
        strokeLinecap="round"
        strokeLinejoin="round"
        className="mantis-arm arm-r"
      />

      <circle cx="28" cy="57" r="1.6" fill="currentColor" className="mantis-claw claw-l" />
      <circle cx="36" cy="57" r="1.6" fill="currentColor" className="mantis-claw claw-r" />
    </svg>
  )
}
