import { useEffect, useLayoutEffect, useRef, useState } from 'react'

// ---- Icons: small inline SVGs sharing one 2px stroke ---------------------------

function Svg({ className = '', children }) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      className={`size-4 shrink-0 ${className}`}
    >
      {children}
    </svg>
  )
}

export const Icon = {
  Logo: (props) => (
    <Svg {...props}>
      <rect x="3" y="3" width="12" height="12" rx="3" />
      <rect x="9" y="9" width="12" height="12" rx="3" />
    </Svg>
  ),
  Plus: (props) => (
    <Svg {...props}>
      <path d="M12 5v14M5 12h14" />
    </Svg>
  ),
  Search: (props) => (
    <Svg {...props}>
      <circle cx="11" cy="11" r="7" />
      <path d="m20 20-3.5-3.5" />
    </Svg>
  ),
  ArrowRight: (props) => (
    <Svg {...props}>
      <path d="M5 12h14M13 6l6 6-6 6" />
    </Svg>
  ),
  Close: (props) => (
    <Svg {...props}>
      <path d="M6 6l12 12M18 6 6 18" />
    </Svg>
  ),
  Copy: (props) => (
    <Svg {...props}>
      <rect x="9" y="9" width="12" height="12" rx="2" />
      <path d="M5 15V5a2 2 0 0 1 2-2h10" />
    </Svg>
  ),
  Check: (props) => (
    <Svg {...props}>
      <path d="m5 12 5 5L20 7" />
    </Svg>
  ),
  External: (props) => (
    <Svg {...props}>
      <path d="M14 4h6v6M20 4l-9 9M18 14v5a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1V7a1 1 0 0 1 1-1h5" />
    </Svg>
  ),
  Trash: (props) => (
    <Svg {...props}>
      <path d="M4 7h16M10 11v6M14 11v6M6 7l1 13h10l1-13M9 7V4h6v3" />
    </Svg>
  ),
  Terminal: (props) => (
    <Svg {...props}>
      <rect x="3" y="4" width="18" height="16" rx="2" />
      <path d="m7 9 3 3-3 3M13 15h4" />
    </Svg>
  ),
  Server: (props) => (
    <Svg {...props}>
      <rect x="3" y="4" width="18" height="7" rx="1.5" />
      <rect x="3" y="13" width="18" height="7" rx="1.5" />
      <path d="M7 7.5h.01M7 16.5h.01" />
    </Svg>
  ),
  Bolt: (props) => (
    <Svg {...props}>
      <path d="M13 3 4 14h7l-1 7 9-11h-7l1-7z" />
    </Svg>
  ),
  Database: (props) => (
    <Svg {...props}>
      <ellipse cx="12" cy="5" rx="8" ry="3" />
      <path d="M4 5v14c0 1.7 3.6 3 8 3s8-1.3 8-3V5M4 12c0 1.7 3.6 3 8 3s8-1.3 8-3" />
    </Svg>
  ),
  Browser: (props) => (
    <Svg {...props}>
      <rect x="3" y="4" width="18" height="16" rx="2" />
      <path d="M3 9h18M7 6.5h.01M10 6.5h.01" />
    </Svg>
  ),
  GitHub: ({ className = '' }) => (
    <svg viewBox="0 0 16 16" fill="currentColor" aria-hidden="true" className={`size-4 shrink-0 ${className}`}>
      <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z" />
    </svg>
  ),
}

// ---- Buttons, badges and form fields --------------------------------------------

const BUTTON_STYLES = {
  primary: 'bg-ink text-white hover:bg-[#333]',
  secondary: 'border border-line bg-white text-ink hover:bg-bone',
  danger: 'border border-red-ink/20 bg-white text-red-ink hover:bg-red-pale',
}

export function Button({ variant = 'primary', className = '', ...props }) {
  return (
    <button
      type="button"
      className={`inline-flex items-center justify-center gap-2 rounded-md px-3.5 py-2 text-sm font-medium transition focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ink active:scale-[0.98] disabled:pointer-events-none disabled:opacity-40 ${BUTTON_STYLES[variant]} ${className}`}
      {...props}
    />
  )
}

export const TONES = {
  gray: 'bg-bone text-body',
  red: 'bg-red-pale text-red-ink',
  blue: 'bg-blue-pale text-blue-ink',
  green: 'bg-green-pale text-green-ink',
  yellow: 'bg-yellow-pale text-yellow-ink',
}

export function Badge({ tone = 'gray', children }) {
  return (
    <span
      className={`inline-flex items-center rounded-full px-2 py-0.5 text-[11px] font-medium uppercase tracking-wider ${TONES[tone]}`}
    >
      {children}
    </span>
  )
}

// Where a looked-up user came from, colored like the store's own branding
const SOURCES = {
  cache: { tone: 'red', label: 'Redis cache' },
  database: { tone: 'blue', label: 'PostgreSQL' },
  github: { tone: 'gray', label: 'GitHub' },
}

export function SourceBadge({ source }) {
  const { tone, label } = SOURCES[source] ?? SOURCES.github
  return <Badge tone={tone}>{label}</Badge>
}

export function Notice({ tone = 'gray', children }) {
  return (
    <p role="status" className={`rounded-lg px-4 py-3 text-sm ${TONES[tone]}`}>
      {children}
    </p>
  )
}

export function Code({ children }) {
  return <code className="rounded bg-bone px-1.5 py-0.5 font-mono text-[0.85em] text-ink">{children}</code>
}

export const inputClass =
  'w-full rounded-md border border-line bg-white px-3 py-2 text-sm text-ink outline-none transition placeholder:text-muted/70 focus:border-ink/30 focus:ring-4 focus:ring-ink/5'

export function Field({ label, hint, className = '', children }) {
  return (
    <label className={`block ${className}`}>
      <span className="mb-1.5 block text-xs font-medium text-body">{label}</span>
      {children}
      {hint && <span className="mt-1 block text-xs text-muted">{hint}</span>}
    </label>
  )
}

// ---- Bigger pieces ----------------------------------------------------------------

// GitHub avatars take a size parameter; ask for double size so they stay sharp
function avatarSrc(url, size) {
  try {
    const parsed = new URL(url)
    if (parsed.hostname === 'avatars.githubusercontent.com') parsed.searchParams.set('s', size * 2)
    return parsed.href
  } catch {
    return url
  }
}

export function Avatar({ user, size = 32 }) {
  if (!user.avatar_url) {
    return (
      <span
        style={{ width: size, height: size }}
        className="grid shrink-0 place-items-center rounded-full bg-bone font-mono text-[10px] uppercase text-muted"
      >
        {user.login.slice(0, 2)}
      </span>
    )
  }
  return (
    <img
      src={avatarSrc(user.avatar_url, size)}
      alt=""
      width={size}
      height={size}
      style={{ width: size, height: size }}
      loading="lazy"
      className="shrink-0 rounded-full border border-line bg-bone"
    />
  )
}

// A modal built on the native <dialog> element, which handles focus and the Esc key
export function Dialog({ open, onClose, label, children }) {
  const ref = useRef(null)

  // Layout effect, so the dialog closes before its emptied contents are painted
  useLayoutEffect(() => {
    const dialog = ref.current
    if (open && !dialog.open) dialog.showModal()
    if (!open && dialog.open) dialog.close()
  }, [open])

  return (
    <dialog
      ref={ref}
      aria-label={label}
      onClose={onClose}
      onClick={(event) => event.target === event.currentTarget && onClose()}
      className="m-auto max-h-[calc(100dvh-2rem)] w-[calc(100%-2rem)] max-w-xl overflow-y-auto rounded-xl border border-line bg-white p-0 text-body backdrop:bg-ink/20 backdrop:backdrop-blur-[2px]"
    >
      {open && children}
    </dialog>
  )
}

export function CopyButton({ text }) {
  const [copied, setCopied] = useState(false)

  function copy() {
    navigator.clipboard
      .writeText(text)
      .then(() => {
        setCopied(true)
        setTimeout(() => setCopied(false), 1500)
      })
      .catch(() => {})
  }

  return (
    <button
      type="button"
      onClick={copy}
      aria-label={copied ? 'Copied' : 'Copy command'}
      className="rounded p-1.5 text-muted transition hover:bg-bone hover:text-ink"
    >
      {copied ? <Icon.Check /> : <Icon.Copy />}
    </button>
  )
}

// Shell commands in a plain window frame, each with a copy button
export function Terminal({ commands, className = '' }) {
  return (
    <div className={`overflow-hidden rounded-lg border border-line bg-white ${className}`}>
      <div className="flex gap-1.5 border-b border-line px-4 py-3" aria-hidden="true">
        <span className="size-2.5 rounded-full bg-line" />
        <span className="size-2.5 rounded-full bg-line" />
        <span className="size-2.5 rounded-full bg-line" />
      </div>
      <ul className="space-y-1 p-3 font-mono text-[13px]">
        {commands.map((command) => (
          <li key={command} className="flex items-center justify-between gap-3 rounded px-2 py-1 hover:bg-canvas">
            <code className="overflow-x-auto whitespace-nowrap text-ink">
              <span className="select-none text-muted">$ </span>
              {command}
            </code>
            <CopyButton text={command} />
          </li>
        ))}
      </ul>
    </div>
  )
}

// Fades its children in the first time they scroll into view
export function Reveal({ children }) {
  const ref = useRef(null)
  const [shown, setShown] = useState(false)

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setShown(true)
          observer.disconnect()
        }
      },
      { threshold: 0.1 },
    )
    observer.observe(ref.current)
    return () => observer.disconnect()
  }, [])

  return (
    <div ref={ref} className={`reveal ${shown ? 'shown' : ''}`}>
      {children}
    </div>
  )
}

// ---- Formatting -----------------------------------------------------------------

// Go sends an unset time as year 1
const UNSET_TIME = '0001-01-01'

// Rows from PostgreSQL always have synced_at; users fetched straight from GitHub don't
export const isStored = (user) => Boolean(user?.synced_at) && !user.synced_at.startsWith(UNSET_TIME)

const relative = new Intl.RelativeTimeFormat('en', { numeric: 'auto' })
const UNITS = [
  ['year', 31536000],
  ['month', 2592000],
  ['day', 86400],
  ['hour', 3600],
  ['minute', 60],
]

export function timeAgo(iso) {
  if (!iso || iso.startsWith(UNSET_TIME)) return '—'
  const seconds = (Date.parse(iso) - Date.now()) / 1000
  const [unit, size] = UNITS.find(([, length]) => Math.abs(seconds) >= length) ?? []
  return unit ? relative.format(Math.round(seconds / size), unit) : 'just now'
}

export function formatDate(iso) {
  if (!iso || iso.startsWith(UNSET_TIME)) return '—'
  return new Date(iso).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
}
