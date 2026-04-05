import type { CSSProperties } from 'react'

export function CloudIcon({ size = 28, style }: { size?: number; style?: CSSProperties }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      style={style}
      aria-hidden
    >
      <path
        d="M7.5 18.5h8.75a4.25 4.25 0 0 0 .43-8.48A5.76 5.76 0 0 0 6.19 8.8 3.97 3.97 0 0 0 7.5 18.5Z"
        stroke="currentColor"
        strokeWidth="1.7"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <circle cx="10" cy="8.3" r="1" fill="currentColor" />
    </svg>
  )
}
