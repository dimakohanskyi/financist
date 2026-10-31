/** Financist's mark, same as the app icon (build/appicon.png). Kept in sync with public/favicon.svg. */
interface LogoProps {
  size?: number;
  className?: string;
}

export function Logo({ size = 24, className }: LogoProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 32 32" className={className} aria-hidden="true">
      <defs>
        <linearGradient id="financistMark" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" style={{ stopColor: "var(--series-1)" }} />
          <stop offset="1" style={{ stopColor: "var(--series-6)" }} />
        </linearGradient>
      </defs>
      <rect x="1" y="1" width="30" height="30" rx="7" fill="url(#financistMark)" />
      {/* Lucide "chart-no-axes-combined" (ISC) */}
      <g
        transform="translate(6 6) scale(0.833)"
        fill="none"
        stroke="#fff"
        strokeWidth="2.2"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <path d="M12 16v5M16 14v7M20 10v11M4 18v3M8 14v7" />
        <path d="m22 3-8.646 8.646a.5.5 0 0 1-.708 0L9.354 8.354a.5.5 0 0 0-.707 0L2 15" />
      </g>
    </svg>
  );
}
