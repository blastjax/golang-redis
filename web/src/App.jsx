import { useEffect, useState } from 'react'
import AboutPage from './AboutPage.jsx'
import { api } from './api.js'
import { Icon } from './ui.jsx'
import UsersPage from './UsersPage.jsx'

// Two pages switched by the URL hash, so links and the back button work without a router
function useHashRoute() {
  const [route, setRoute] = useState(() => window.location.hash.slice(1) || '/')

  useEffect(() => {
    const update = () => {
      setRoute(window.location.hash.slice(1) || '/')
      window.scrollTo(0, 0)
    }
    window.addEventListener('hashchange', update)
    return () => window.removeEventListener('hashchange', update)
  }, [])

  return route
}

// Checks GET /health now and every 30 seconds, so the header can say whether the API is up
function useApiOnline() {
  const [online, setOnline] = useState(null)

  useEffect(() => {
    const check = () => api.health().then((res) => setOnline(res.ok))
    check()
    const timer = setInterval(check, 30_000)
    return () => clearInterval(timer)
  }, [])

  return online
}

export default function App() {
  const route = useHashRoute()
  const online = useApiOnline()
  const onAbout = route === '/about'

  return (
    <>
      {/* A faint warm light behind the page */}
      <div
        aria-hidden="true"
        className="pointer-events-none fixed inset-0 -z-10 bg-[radial-gradient(48rem_32rem_at_85%_-10%,rgba(222,203,170,0.25),transparent)]"
      />

      <header className="sticky top-0 z-20 border-b border-line bg-canvas/85 backdrop-blur">
        <div className="mx-auto flex h-14 max-w-5xl items-center gap-3 px-4 sm:gap-6 sm:px-6">
          <a href="#/" className="flex items-center gap-2 whitespace-nowrap text-sm font-semibold text-ink">
            <Icon.Logo />
            golang-redis
          </a>
          <nav className="flex gap-1 text-sm">
            <NavLink href="#/" active={!onAbout}>
              Users
            </NavLink>
            <NavLink href="#/about" active={onAbout}>
              About
            </NavLink>
          </nav>
          <ApiStatus online={online} />
        </div>
      </header>

      <main className="mx-auto max-w-5xl px-4 sm:px-6">{onAbout ? <AboutPage /> : <UsersPage />}</main>

      <footer className="mx-auto max-w-5xl border-t border-line px-4 py-8 text-xs text-muted sm:px-6">
        Built on the golang-redis REST API: Go, Gin, GORM, PostgreSQL and Redis.
      </footer>
    </>
  )
}

function NavLink({ href, active, children }) {
  return (
    <a
      href={href}
      aria-current={active ? 'page' : undefined}
      className={`rounded-md px-3 py-1.5 transition ${active ? 'bg-bone font-medium text-ink' : 'text-muted hover:text-ink'}`}
    >
      {children}
    </a>
  )
}

function ApiStatus({ online }) {
  const [label, dot] =
    online === null ? ['Checking API', 'bg-line'] : online ? ['API online', 'bg-green-ink'] : ['API offline', 'bg-red-ink']

  return (
    <span className="ml-auto flex items-center gap-2 whitespace-nowrap text-xs text-muted" title="Result of GET /health">
      <span className={`size-1.5 rounded-full ${dot}`} />
      {/* Phones only get the colored dot; screen readers still hear the label */}
      <span className="sr-only sm:not-sr-only">{label}</span>
    </span>
  )
}
