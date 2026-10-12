import { Fragment, useEffect, useState } from 'react'
import { api, errorMessage } from './api.js'
import { lookupPath } from './flow.js'
import { Badge, Button, Code, Icon, Reveal, TONES, Terminal } from './ui.jsx'

export default function AboutPage() {
  return (
    <div className="pb-8">
      <Intro />
      <Parts />
      <LookupTracer />
      <SyncDemo />
      <Reference />
    </div>
  )
}

// A titled part of the page that fades in as it scrolls into view
function Section({ eyebrow, title, lead, children }) {
  return (
    <Reveal>
      <section className="border-t border-line py-16 sm:py-20">
        <p className="font-mono text-xs uppercase tracking-[0.2em] text-muted">{eyebrow}</p>
        <h2 className="mt-3 font-serif text-3xl tracking-tight text-ink sm:text-4xl">{title}</h2>
        <p className="mt-4 max-w-2xl">{lead}</p>
        <div className="mt-10">{children}</div>
      </section>
    </Reveal>
  )
}

// ---- What it is -------------------------------------------------------------------

const FACTS = [
  {
    icon: Icon.Terminal,
    title: 'Sync',
    text: "sync-users reads GitHub's user list 100 at a time and saves every page to PostgreSQL.",
  },
  {
    icon: Icon.Server,
    title: 'Serve',
    text: 'A REST API lists, looks up, adds, edits and deletes users. This site is built on it.',
  },
  {
    icon: Icon.Bolt,
    title: 'Cache',
    text: "Redis remembers each lookup for 30 seconds by default, including users that don't exist.",
  },
]

function Intro() {
  return (
    <section className="py-16 sm:py-24">
      <p className="font-mono text-xs uppercase tracking-[0.2em] text-muted">About golang-redis</p>
      <h1 className="mt-4 max-w-3xl font-serif text-5xl leading-[1.05] tracking-tight text-ink sm:text-6xl">
        A local copy of GitHub's users, with a cache in front.
      </h1>
      <p className="mt-6 max-w-2xl text-lg">
        golang-redis is a Go program in two parts. A command-line tool copies users from GitHub into PostgreSQL. A REST API
        serves them back, and Redis sits in front of it so repeat lookups don't touch the database or GitHub.
      </p>
      <div className="mt-12 grid gap-4 sm:grid-cols-3">
        {FACTS.map(({ icon: FactIcon, title, text }) => (
          <div key={title} className="rounded-xl border border-line bg-white p-6">
            <span className="grid size-9 place-items-center rounded-md bg-bone text-ink">
              <FactIcon />
            </span>
            <h3 className="mt-4 font-medium text-ink">{title}</h3>
            <p className="mt-1 text-sm text-muted">{text}</p>
          </div>
        ))}
      </div>
    </section>
  )
}

// ---- The parts ----------------------------------------------------------------------

const PARTS = [
  {
    id: 'cli',
    name: 'Command line',
    icon: Icon.Terminal,
    code: 'main.go',
    summary: 'One binary with four commands. It prepares the database, syncs users from GitHub and starts the API.',
    points: [
      <>
        <Code>sync-users -store</Code> continues after the highest user ID already stored, so each run picks up where the
        last one stopped.
      </>,
      <>
        <Code>list-users</Code> prints stored users as a table, JSON or a detailed list. Logs go to stderr, so the JSON can
        be piped into other tools.
      </>,
    ],
    commands: [
      './golang-redis setup-db',
      './golang-redis sync-users -count 100 -store',
      './golang-redis list-users -format json',
      './golang-redis serve',
    ],
  },
  {
    id: 'api',
    name: 'REST API',
    icon: Icon.Server,
    code: 'internal/api',
    summary: 'A Gin web server, and the only part your browser talks to.',
    points: [
      'Checks that a username could exist on GitHub before it goes anywhere near a URL or a cache key.',
      'Reads try Redis first, then PostgreSQL, then GitHub, and stop at the first one with an answer.',
      "Writes go to PostgreSQL and then clear the user's Redis entry, so the next read sees the change.",
    ],
  },
  {
    id: 'redis',
    name: 'Redis',
    icon: Icon.Bolt,
    code: 'internal/cache',
    summary: 'An in-memory store that answers repeat lookups in a millisecond or two.',
    points: [
      <>
        Each user is saved as JSON under <Code>{'user:<login>'}</Code>, in lower case, for <Code>REDIS_TTL_SECONDS</Code>{' '}
        (30 by default).
      </>,
      "Answers like \"this user doesn't exist\" are cached too, so a typo doesn't use up GitHub's rate limit.",
      'If Redis is down, lookups still work. They just skip the cache.',
    ],
  },
  {
    id: 'postgres',
    name: 'PostgreSQL',
    icon: Icon.Database,
    code: 'internal/database',
    summary: 'The permanent store for synced and hand-added users, accessed through GORM.',
    points: [
      <>
        Logins use the <Code>citext</Code> type, so <Code>mojombo</Code> and <Code>MOJOMBO</Code> are the same user, as on
        GitHub.
      </>,
      'A re-sync only refreshes the login, avatar, URL and type, so edits made here are kept.',
      'The database itself refuses duplicate IDs and logins, and the API reports that as 409 Conflict.',
    ],
  },
  {
    id: 'github',
    name: 'GitHub API',
    icon: Icon.GitHub,
    code: 'internal/github',
    summary: "The source for anyone the program hasn't stored yet.",
    points: [
      <>
        <Code>GET /users</Code> lists accounts in ID order, and <Code>{'GET /users/{login}'}</Code> fetches one.
      </>,
      <>
        Without a token GitHub allows 60 requests an hour. Set <Code>GITHUB_TOKEN</Code> to get 5,000.
      </>,
    ],
  },
]

function Parts() {
  const [selectedId, setSelectedId] = useState('api')
  const part = PARTS.find(({ id }) => id === selectedId)

  return (
    <Section
      eyebrow="The pieces"
      title="Five parts, one job each"
      lead="Choose a part to see what it does and where its code lives."
    >
      <div className="grid gap-4 md:grid-cols-[14rem_1fr]">
        <div className="flex gap-2 overflow-x-auto pb-1 md:flex-col md:pb-0">
          {PARTS.map(({ id, name, icon: PartIcon }) => (
            <button
              key={id}
              type="button"
              aria-pressed={id === selectedId}
              onClick={() => setSelectedId(id)}
              className={`flex shrink-0 items-center gap-3 rounded-lg border px-4 py-3 text-left text-sm transition ${
                id === selectedId
                  ? 'border-ink/15 bg-white font-medium text-ink'
                  : 'border-transparent text-muted hover:bg-white hover:text-ink'
              }`}
            >
              <PartIcon />
              {name}
            </button>
          ))}
        </div>

        <div className="rounded-xl border border-line bg-white p-6 sm:p-8">
          <div className="flex flex-wrap items-center gap-3">
            <h3 className="font-serif text-2xl text-ink">{part.name}</h3>
            <Code>{part.code}</Code>
          </div>
          <p className="mt-2">{part.summary}</p>
          <ul className="mt-6 space-y-3 text-sm">
            {part.points.map((point, i) => (
              <li key={i} className="flex gap-3">
                <span className="mt-2 size-1.5 shrink-0 rounded-full bg-muted/50" />
                <span>{point}</span>
              </li>
            ))}
          </ul>
          {part.commands && <Terminal commands={part.commands} className="mt-6" />}
        </div>
      </div>
    </Section>
  )
}

// ---- A live lookup ------------------------------------------------------------------

const HOPS = [
  { id: 'browser', name: 'Your browser', icon: Icon.Browser },
  { id: 'api', name: 'REST API', icon: Icon.Server },
  { id: 'redis', name: 'Redis', icon: Icon.Bolt },
  { id: 'postgres', name: 'PostgreSQL', icon: Icon.Database },
  { id: 'github', name: 'GitHub API', icon: Icon.GitHub },
]

const PRESETS = [
  ['mojombo', "GitHub's first user"],
  ['MOJOMBO', 'same user, in capitals'],
  ['torvalds', 'probably not stored'],
  ['no-such-user-7f3a9c', "doesn't exist"],
  ['not a username', 'invalid'],
]

const HOP_LABELS = {
  api: { ok: 'Username is valid', rejected: 'Refused the username', error: 'No response' },
  redis: { hit: 'Cache hit', miss: 'Cache miss', error: 'Unavailable, skipped' },
  postgres: { found: 'Has the user', missing: 'Not stored here', error: 'Query failed' },
  github: { found: 'Has the user', missing: 'No such user', error: 'Request failed' },
}

const STATE_TONES = { ok: 'green', hit: 'green', found: 'green', miss: 'yellow', missing: 'yellow', rejected: 'red', error: 'red' }

// What a hop shows once the response is in
function describeHop(hop, { res, path }) {
  if (hop === 'browser') {
    const tone = res.ok ? 'green' : res.status && res.status < 500 ? 'yellow' : 'red'
    return { tone, label: res.status ? `HTTP ${res.status} in ${res.ms} ms` : 'No response' }
  }
  const state = path[hop]
  if (state === 'skipped') return { skipped: true, label: path.api === 'ok' ? 'Not needed' : 'Not reached' }
  const label = hop === 'redis' && state === 'hit' && res.status === 404 ? 'Cache hit: no such user' : HOP_LABELS[hop][state]
  return { tone: STATE_TONES[state], label }
}

function explain(login, res, path) {
  const cached = path.stored ? ' The API saved the answer in Redis, so sending it again will be a cache hit.' : ''
  if (path.api === 'rejected') {
    return `The API refused "${login}" without contacting anything else. GitHub usernames only use letters, numbers, hyphens and underscores.`
  }
  if (path.api === 'error') return 'The API did not answer. Is ./golang-redis serve running?'
  if (path.redis === 'hit') {
    return res.status === 404
      ? `Redis remembered that nobody is called ${login}, so GitHub wasn't asked again.`
      : `Redis already had ${login}, so neither PostgreSQL nor GitHub was needed.`
  }
  if (path.postgres === 'found') return `Redis didn't have ${login}, but PostgreSQL did.${cached}`
  if (path.github === 'found') {
    return `${login} isn't stored in PostgreSQL, so the API fetched the profile from GitHub.${cached} To keep it permanently, use Save to database on the Users page.`
  }
  if (path.github === 'missing') {
    return `Nobody on GitHub is called ${login}.${path.stored ? " The API cached that answer too, so asking again won't use up GitHub's rate limit until the entry expires." : ''}`
  }
  return errorMessage(res)
}

function LookupTracer() {
  const [login, setLogin] = useState('mojombo')
  const [busy, setBusy] = useState(false)
  const [run, setRun] = useState(null)
  const [revealed, setRevealed] = useState(0)

  // Light the hops up one by one, so the order of the checks is easy to follow
  useEffect(() => {
    if (!run || revealed >= HOPS.length) return
    const timer = setTimeout(() => setRevealed((n) => n + 1), 300)
    return () => clearTimeout(timer)
  }, [run, revealed])

  async function send(name = login) {
    const target = name.trim()
    if (!target) return
    setLogin(target)
    setBusy(true)
    setRun(null)
    const res = await api.getUser(target)
    setRevealed(0)
    setRun({ login: target, res, path: lookupPath(res) })
    setBusy(false)
  }

  const done = run && revealed >= HOPS.length

  return (
    <Section
      eyebrow="Try it"
      title="Follow a lookup"
      lead="Send a real request to the API and watch which parts of the program it goes through. Send the same name twice: the second answer comes straight from Redis."
    >
      <div className="overflow-hidden rounded-xl border border-line bg-white">
        <form
          onSubmit={(event) => {
            event.preventDefault()
            send()
          }}
          className="flex flex-col gap-3 p-6 sm:flex-row"
        >
          <div className="flex flex-1 items-center rounded-md border border-line transition focus-within:border-ink/30 focus-within:ring-4 focus-within:ring-ink/5">
            <span className="select-none pl-3 font-mono text-sm text-muted">GET /users/</span>
            <input
              value={login}
              onChange={(event) => setLogin(event.target.value)}
              name="username"
              aria-label="GitHub username"
              autoComplete="off"
              spellCheck="false"
              className="min-w-0 flex-1 bg-transparent py-2 pr-3 font-mono text-sm text-ink outline-none"
            />
          </div>
          <Button type="submit" disabled={busy || !login.trim()}>
            {busy ? 'Sending…' : 'Send request'}
            <Icon.ArrowRight />
          </Button>
        </form>

        <div className="flex flex-wrap items-center gap-2 px-6 pb-6">
          <span className="mr-1 text-xs text-muted">Try</span>
          {PRESETS.map(([name, hint]) => (
            <button
              key={name}
              type="button"
              disabled={busy}
              onClick={() => send(name)}
              className="rounded-md border border-line px-2.5 py-1 text-xs transition hover:border-ink/20 hover:bg-bone disabled:opacity-50"
            >
              <span className="font-mono text-ink">{name}</span> <span className="text-muted">{hint}</span>
            </button>
          ))}
        </div>

        <div className="border-t border-line bg-canvas p-6">
          <div className="flex flex-col gap-2 md:flex-row">
            {HOPS.map((hop, i) => {
              const info = run && i < revealed ? describeHop(hop.id, run) : null
              return (
                <Fragment key={hop.id}>
                  {i > 0 && <Connector active={Boolean(info) && !info.skipped} />}
                  <Hop hop={hop} info={info} waiting={busy && i === 0} />
                </Fragment>
              )
            })}
          </div>

          <div aria-live="polite" className="mt-6 min-h-24">
            {done ? (
              <Outcome run={run} />
            ) : (
              <p className="text-sm text-muted">
                {busy || run ? 'Following the request…' : 'Pick a name above or type your own, then send the request.'}
              </p>
            )}
          </div>
        </div>
      </div>
    </Section>
  )
}

const HOP_STYLES = {
  green: 'border-green-ink/15 bg-green-pale text-green-ink',
  yellow: 'border-yellow-ink/15 bg-yellow-pale text-yellow-ink',
  red: 'border-red-ink/15 bg-red-pale text-red-ink',
}

function Hop({ hop, info, waiting }) {
  const style = !info
    ? 'border-line bg-white text-muted'
    : info.skipped
      ? 'border-dashed border-line text-muted opacity-60'
      : HOP_STYLES[info.tone]

  return (
    <div className={`flex-1 rounded-lg border p-4 transition-colors duration-300 ${style} ${waiting ? 'animate-pulse' : ''}`}>
      <div className="flex items-center gap-2 text-sm font-medium text-ink">
        <hop.icon />
        {hop.name}
      </div>
      <p className="mt-1 text-xs">{info ? info.label : waiting ? 'Sending…' : ' '}</p>
    </div>
  )
}

function Connector({ active }) {
  return (
    <div
      aria-hidden="true"
      className={`flex items-center justify-center transition-colors duration-300 ${active ? 'text-ink' : 'text-line'}`}
    >
      <Icon.ArrowRight className="rotate-90 md:rotate-0" />
    </div>
  )
}

function Outcome({ run: { login, res, path } }) {
  return (
    <div className="space-y-4">
      <p className="max-w-3xl">{explain(login, res, path)}</p>
      <div className="flex flex-wrap gap-x-6 gap-y-1 font-mono text-xs text-muted">
        {res.cacheStatus && <span>X-Cache-Status: {res.cacheStatus}</span>}
        {res.data.source && <span>source: {res.data.source}</span>}
        {path.stored && <span>cached as user:{login.toLowerCase()}</span>}
      </div>
      <details className="rounded-lg border border-line bg-white">
        <summary className="cursor-pointer select-none px-4 py-2 text-sm text-muted hover:text-ink">Response body</summary>
        <pre className="max-h-72 overflow-auto border-t border-line p-4 font-mono text-xs">
          {JSON.stringify(res.data, null, 2)}
        </pre>
      </details>
    </div>
  )
}

// ---- How syncing works ----------------------------------------------------------------

// The first GitHub accounts in ID order. The missing numbers belong to deleted accounts.
const FIRST_ACCOUNTS = [
  [1, 'mojombo'],
  [2, 'defunkt'],
  [3, 'pjhyett'],
  [4, 'wycats'],
  [5, 'ezmobius'],
  [6, 'ivey'],
  [7, 'evanphx'],
  [17, 'vanpelt'],
  [18, 'wayneeseguin'],
  [19, 'brynary'],
  [20, 'kevinclark'],
  [21, 'technoweenie'],
  [22, 'macournoyer'],
  [23, 'takeo'],
  [25, 'caged'],
  [26, 'topfunky'],
  [27, 'anotherjesse'],
  [28, 'roland'],
  [29, 'lukas'],
  [30, 'fanvsfan'],
  [31, 'tomtt'],
  [32, 'railsjitsu'],
  [34, 'nitay'],
  [35, 'kevwil'],
]
const DEMO_PAGE_SIZE = 6
const PAGE_STYLES = [TONES.blue, TONES.green]

function SyncDemo() {
  const [pages, setPages] = useState([])
  const latest = pages.findLastIndex((page) => page.ids.length > 0)
  const since = latest >= 0 ? pages[latest].ids.at(-1) : 0
  const finished = pages.length > 0 && pages.at(-1).ids.length < DEMO_PAGE_SIZE

  // What GET /users?since=...&per_page=... would return
  function fetchPage() {
    const ids = FIRST_ACCOUNTS.map(([id]) => id)
      .filter((id) => id > since)
      .slice(0, DEMO_PAGE_SIZE)
    setPages([...pages, { since, ids }])
  }

  return (
    <Section
      eyebrow="Under the hood"
      title="How syncing works"
      lead={
        <>
          GitHub hands out its user list in ID order, one page at a time. Each request asks for the users after a given ID,
          called <Code>since</Code>. Deleted accounts leave gaps in the numbering, so the next <Code>since</Code> can't be
          worked out in advance: it's the last ID on the page that just arrived. That's why <Code>sync-users</Code> fetches
          pages one after another. Step through it:
        </>
      }
    >
      <div className="grid gap-6 lg:grid-cols-[1fr_22rem]">
        <div>
          <ol className="grid grid-cols-4 gap-2 sm:grid-cols-6">
            {FIRST_ACCOUNTS.map(([id, login]) => {
              const page = pages.findIndex(({ ids }) => ids.includes(id))
              const style =
                page < 0
                  ? 'border-line bg-white text-muted'
                  : `border-transparent ${PAGE_STYLES[page % 2]} ${page === latest ? 'ring-1 ring-ink/25' : ''}`
              return (
                <li key={id} className={`rounded-md border px-2.5 py-2 transition-colors duration-300 ${style}`}>
                  <div className="font-mono text-sm font-medium">{id}</div>
                  <div className="truncate text-[11px]">{login}</div>
                </li>
              )
            })}
          </ol>
          <p className="mt-3 text-xs text-muted">Numbers 8 to 16, 24 and 33 are missing: those accounts were deleted.</p>
        </div>

        <div className="flex flex-col rounded-xl border border-line bg-white p-5">
          <div className="flex flex-wrap gap-2">
            <Button onClick={fetchPage} disabled={finished}>
              Fetch next page
            </Button>
            {pages.length > 0 && (
              <Button variant="secondary" onClick={() => setPages([])}>
                Start over
              </Button>
            )}
          </div>
          <ol className="mt-4 space-y-2 font-mono text-xs">
            {pages.map((page, i) => (
              <li key={i} className="rounded-md border border-line bg-canvas px-3 py-2">
                <div className="text-ink">{`GET /users?since=${page.since}&per_page=${DEMO_PAGE_SIZE}`}</div>
                <div className="mt-0.5 text-muted">
                  {page.ids.length > 0
                    ? `${page.ids.length} users (${page.ids[0]}–${page.ids.at(-1)}) saved. Next since=${page.ids.at(-1)}`
                    : `No users: fewer than ${DEMO_PAGE_SIZE}, so the sync stops`}
                </div>
              </li>
            ))}
          </ol>
          <p className="mt-auto pt-4 text-xs text-muted">
            {finished
              ? `Done. Running sync-users -store again later starts after ID ${since}, the highest one already stored.`
              : `Pages here hold ${DEMO_PAGE_SIZE} users; the real sync asks for 100 at a time.`}
          </p>
        </div>
      </div>
    </Section>
  )
}

// ---- The rest of the API ----------------------------------------------------------------

const ENDPOINTS = [
  ['GET', '/users?limit=30&offset=0', 'Lists stored users in ID order, a page at a time, with the total count.'],
  ['POST', '/users', 'Adds a user. The ID and the login must both be new; a duplicate gets 409 Conflict.'],
  [
    'PUT',
    '/users/{username}',
    'Changes the name, company, location, website or bio, then clears the cached copy so the next lookup reads the new values.',
  ],
  ['DELETE', '/users/{username}', 'Removes the user from PostgreSQL and Redis. Looking them up afterwards falls back to GitHub.'],
]
const METHOD_TONES = { GET: 'blue', POST: 'green', PUT: 'yellow', DELETE: 'red' }

function Reference() {
  return (
    <Section
      eyebrow="Reference"
      title="The rest of the API"
      lead="Every change made through the API goes to PostgreSQL first and then clears that user's Redis entry, so the next lookup reads fresh data. Re-syncing from GitHub only refreshes the login, avatar, URL and type, so your edits survive it."
    >
      <div className="grid gap-4 sm:grid-cols-2">
        {ENDPOINTS.map(([method, path, text]) => (
          <div key={method} className="rounded-xl border border-line bg-white p-6">
            <div className="flex flex-wrap items-center gap-3">
              <Badge tone={METHOD_TONES[method]}>{method}</Badge>
              <code className="font-mono text-xs text-ink">{path}</code>
            </div>
            <p className="mt-3 text-sm">{text}</p>
          </div>
        ))}
      </div>
      <p className="mt-10 text-sm text-muted">
        Ready to try it?{' '}
        <a href="#/" className="text-ink underline decoration-line underline-offset-4 hover:decoration-ink">
          Go to the Users page
        </a>
        .
      </p>
    </Section>
  )
}
