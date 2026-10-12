import { useEffect, useState } from 'react'
import { api, errorMessage } from './api.js'
import {
  Avatar,
  Badge,
  Button,
  Dialog,
  Field,
  Icon,
  Notice,
  SourceBadge,
  Terminal,
  formatDate,
  inputClass,
  isStored,
  timeAgo,
} from './ui.jsx'

const PAGE_SIZES = [10, 25, 50]

export default function UsersPage() {
  const [page, setPage] = useState({ limit: 25, offset: 0 })
  const [list, setList] = useState({ status: 'loading' })
  const [reloads, setReloads] = useState(0)
  const [selected, setSelected] = useState(null)
  const [adding, setAdding] = useState(false)
  const [lookupKey, setLookupKey] = useState(0)
  const reload = () => setReloads((n) => n + 1)

  useEffect(() => {
    let current = true
    api.listUsers(page.limit, page.offset).then((res) => {
      if (!current) return
      if (!res.ok) return setList({ status: 'error', error: errorMessage(res) })

      const users = res.data.users ?? []
      // Deleting the only user on the last page leaves it empty, so step back a page
      if (users.length === 0 && page.offset > 0) {
        return setPage((p) => ({ ...p, offset: Math.max(0, p.offset - p.limit) }))
      }
      setList({ status: 'ok', users, total: res.data.total })
    })
    return () => {
      current = false
    }
  }, [page, reloads])

  return (
    <div className="py-12 sm:py-16">
      <div className="flex flex-wrap items-end justify-between gap-6">
        <div>
          <h1 className="font-serif text-4xl tracking-tight text-ink sm:text-5xl">GitHub users</h1>
          <p className="mt-3 max-w-xl text-muted">
            Users copied from GitHub into PostgreSQL. Look anyone up, edit what's stored, or add someone by hand.
          </p>
        </div>
        <Button onClick={() => setAdding(true)}>
          <Icon.Plus />
          Add user
        </Button>
      </div>

      <Lookup key={lookupKey} onOpen={setSelected} onSaved={reload} />
      <UserList list={list} page={page} setPage={setPage} onOpen={setSelected} />

      <Dialog open={Boolean(selected)} onClose={() => setSelected(null)} label="User details">
        {selected && (
          <UserEditor
            key={selected.id}
            user={selected}
            onClose={() => setSelected(null)}
            onChanged={(user) => {
              setSelected(user)
              reload()
              // A deleted user may still be showing in the lookup panel, so clear it
              if (!user) setLookupKey((n) => n + 1)
            }}
          />
        )}
      </Dialog>

      <Dialog open={adding} onClose={() => setAdding(false)} label="Add a user">
        <AddUserForm onClose={() => setAdding(false)} onAdded={reload} />
      </Dialog>
    </div>
  )
}

// ---- Lookup -------------------------------------------------------------------

// The fields POST /users accepts, used to copy a GitHub profile into PostgreSQL
const CREATE_FIELDS = ['id', 'login', 'type', 'avatar_url', 'url', 'name', 'company', 'blog', 'location', 'email', 'bio']

function Lookup({ onOpen, onSaved }) {
  const [login, setLogin] = useState('')
  const [busy, setBusy] = useState(false)
  const [result, setResult] = useState(null)

  async function lookUp(event) {
    event.preventDefault()
    const name = login.trim()
    setBusy(true)
    setResult({ name, res: await api.getUser(name) })
    setBusy(false)
  }

  async function save(user) {
    const fields = Object.fromEntries(CREATE_FIELDS.filter((key) => user[key]).map((key) => [key, user[key]]))
    const res = await api.createUser(fields)
    if (!res.ok) return setResult((current) => ({ ...current, saveError: errorMessage(res) }))
    setResult((current) => ({ ...current, res: { ...current.res, data: { user: res.data.user, source: 'database' } } }))
    onSaved()
  }

  return (
    <section className="mt-10 rounded-xl border border-line bg-white p-6">
      <form onSubmit={lookUp} className="flex flex-col gap-3 sm:flex-row">
        <div className="relative flex-1">
          <Icon.Search className="absolute left-3 top-1/2 -translate-y-1/2 text-muted" />
          <input
            value={login}
            onChange={(event) => setLogin(event.target.value)}
            name="username"
            aria-label="GitHub username"
            placeholder="GitHub username, like torvalds"
            autoComplete="off"
            spellCheck="false"
            className={`${inputClass} pl-9`}
          />
        </div>
        <Button type="submit" disabled={busy || !login.trim()}>
          {busy ? 'Looking up…' : 'Look up'}
        </Button>
      </form>
      <p className="mt-2 text-xs text-muted">
        Checks Redis first, then PostgreSQL, then GitHub.{' '}
        <a href="#/about" className="underline decoration-line underline-offset-4 hover:text-ink">
          See how it works
        </a>
      </p>
      {result && <LookupResult result={result} onOpen={onOpen} onSave={save} />}
    </section>
  )
}

function LookupResult({ result: { name, res, saveError }, onOpen, onSave }) {
  if (!res.ok) {
    const text =
      res.status === 404
        ? `Nobody on GitHub is called "${name}".`
        : res.status === 400
          ? `"${name}" isn't a valid GitHub username. They only use letters, numbers, hyphens and underscores.`
          : errorMessage(res)
    return (
      <div className="mt-5">
        <Notice tone={res.status === 404 || res.status === 400 ? 'gray' : 'red'}>{text}</Notice>
      </div>
    )
  }

  const { user, source } = res.data
  const stored = source === 'database' || isStored(user)
  const details = [user.name, user.company, user.location].filter(Boolean).join(' · ')

  return (
    <div className="mt-5 flex flex-wrap items-center gap-4 rounded-lg border border-line bg-canvas p-4">
      <Avatar user={user} size={48} />
      <div className="min-w-0 flex-1">
        <div className="flex flex-wrap items-center gap-2">
          <span className="font-medium text-ink">{user.login}</span>
          <SourceBadge source={source} />
          <span className="font-mono text-xs text-muted">{res.ms} ms</span>
        </div>
        <p className="truncate text-sm text-muted">{details || `${user.type} · ID ${user.id}`}</p>
        {saveError && <p className="mt-1 text-sm text-red-ink">{saveError}</p>}
      </div>
      {stored ? (
        <Button variant="secondary" onClick={() => onOpen(user)}>
          Edit
        </Button>
      ) : (
        <Button variant="secondary" onClick={() => onSave(user)}>
          <Icon.Plus />
          Save to database
        </Button>
      )}
    </div>
  )
}

// ---- Stored users -------------------------------------------------------------

function UserList({ list, page, setPage, onOpen }) {
  if (list.status === 'loading') {
    return <div className="mt-6 h-72 animate-pulse rounded-xl border border-line bg-white" aria-label="Loading users" />
  }
  if (list.status === 'error') {
    return (
      <div className="mt-6">
        <Notice tone="red">{list.error}</Notice>
      </div>
    )
  }
  if (list.total === 0) return <EmptyState />

  const { limit, offset } = page
  const { users, total } = list

  return (
    <section className="mt-6 overflow-hidden rounded-xl border border-line bg-white">
      <div className="flex items-baseline justify-between border-b border-line px-6 py-4">
        <h2 className="font-medium text-ink">Stored in PostgreSQL</h2>
        <span className="text-sm text-muted">{total.toLocaleString()} users</span>
      </div>

      <table className="w-full text-left text-sm">
        <thead className="text-xs uppercase tracking-wider text-muted">
          <tr>
            <th className="px-6 py-3 font-medium">User</th>
            <th className="px-3 py-3 font-medium">Type</th>
            <th className="hidden px-3 py-3 font-medium md:table-cell">ID</th>
            <th className="hidden px-6 py-3 text-right font-medium sm:table-cell">Last synced</th>
          </tr>
        </thead>
        <tbody>
          {users.map((user) => (
            <tr
              key={user.id}
              onClick={() => onOpen(user)}
              className="cursor-pointer border-t border-line transition hover:bg-canvas"
            >
              <td className="px-6 py-3">
                <div className="flex items-center gap-3">
                  <Avatar user={user} />
                  <div className="min-w-0">
                    <button
                      type="button"
                      onClick={(event) => {
                        event.stopPropagation()
                        onOpen(user)
                      }}
                      className="font-medium text-ink hover:underline focus-visible:underline"
                    >
                      {user.login}
                    </button>
                    {user.name && <div className="truncate text-muted">{user.name}</div>}
                  </div>
                </div>
              </td>
              <td className="px-3 py-3">
                <Badge tone={user.type === 'User' ? 'gray' : 'yellow'}>{user.type}</Badge>
              </td>
              <td className="hidden px-3 py-3 font-mono text-xs text-muted md:table-cell">{user.id}</td>
              <td className="hidden px-6 py-3 text-right text-muted sm:table-cell">{timeAgo(user.synced_at)}</td>
            </tr>
          ))}
        </tbody>
      </table>

      <div className="flex flex-wrap items-center justify-between gap-3 border-t border-line px-6 py-3 text-sm text-muted">
        <span>
          {offset + 1}–{offset + users.length} of {total.toLocaleString()}
        </span>
        <div className="flex items-center gap-2">
          <label className="mr-2 flex items-center gap-2">
            Rows
            <select
              name="rows"
              value={limit}
              onChange={(event) => setPage({ limit: Number(event.target.value), offset: 0 })}
              className="rounded-md border border-line bg-white px-2 py-1 text-ink"
            >
              {PAGE_SIZES.map((size) => (
                <option key={size}>{size}</option>
              ))}
            </select>
          </label>
          <Button
            variant="secondary"
            disabled={offset === 0}
            onClick={() => setPage({ limit, offset: Math.max(0, offset - limit) })}
          >
            Previous
          </Button>
          <Button variant="secondary" disabled={offset + limit >= total} onClick={() => setPage({ limit, offset: offset + limit })}>
            Next
          </Button>
        </div>
      </div>
    </section>
  )
}

function EmptyState() {
  return (
    <section className="mt-6 rounded-xl border border-dashed border-line bg-white px-6 py-14 text-center">
      <h2 className="font-serif text-2xl text-ink">No users stored yet</h2>
      <p className="mx-auto mt-2 max-w-md text-muted">
        Copy the first hundred GitHub users into PostgreSQL from the command line, then refresh this page. Or look someone
        up above and save them.
      </p>
      <Terminal commands={['./golang-redis sync-users -count 100 -store']} className="mx-auto mt-8 max-w-lg text-left" />
    </section>
  )
}

// ---- Edit and delete ----------------------------------------------------------

// The fields PUT /users/:username accepts
const EDITABLE = [
  ['name', 'Name'],
  ['company', 'Company'],
  ['location', 'Location'],
  ['blog', 'Website'],
  ['bio', 'Bio'],
]

function UserEditor({ user, onClose, onChanged }) {
  const [form, setForm] = useState(() => Object.fromEntries(EDITABLE.map(([key]) => [key, user[key] ?? ''])))
  const [message, setMessage] = useState(null)
  const [confirmingDelete, setConfirmingDelete] = useState(false)
  const [busy, setBusy] = useState(false)

  // Send only what changed: the API refuses an update without any fields
  const changes = Object.fromEntries(
    EDITABLE.filter(([key]) => form[key] !== (user[key] ?? '')).map(([key]) => [key, form[key]]),
  )
  const dirty = Object.keys(changes).length > 0

  async function save(event) {
    event.preventDefault()
    setBusy(true)
    const res = await api.updateUser(user.login, changes)
    setBusy(false)
    if (!res.ok) return setMessage({ tone: 'red', text: errorMessage(res) })
    setMessage({ tone: 'green', text: 'Saved. The copy in Redis was cleared, so the next lookup reads this version.' })
    onChanged(res.data.user)
  }

  async function remove() {
    if (!confirmingDelete) return setConfirmingDelete(true)
    setBusy(true)
    const res = await api.deleteUser(user.login)
    setBusy(false)
    if (!res.ok) return setMessage({ tone: 'red', text: errorMessage(res) })
    onChanged(null)
  }

  // autoComplete off: these fields describe a GitHub user, not the person typing
  const input = (key) => ({
    name: key,
    autoComplete: 'off',
    value: form[key],
    onChange: (event) => setForm({ ...form, [key]: event.target.value }),
    className: inputClass,
  })

  return (
    <form onSubmit={save}>
      <div className="flex items-start gap-4 border-b border-line p-6">
        <Avatar user={user} size={56} />
        <div className="min-w-0 flex-1">
          <h2 className="truncate font-serif text-2xl leading-tight text-ink">{user.login}</h2>
          <div className="mt-2 flex flex-wrap items-center gap-2">
            <Badge>{user.type}</Badge>
            <Badge>ID {user.id}</Badge>
            <a
              href={`https://github.com/${user.login}`}
              target="_blank"
              rel="noreferrer"
              className="inline-flex items-center gap-1 text-xs text-muted hover:text-ink"
            >
              github.com/{user.login}
              <Icon.External className="size-3" />
            </a>
          </div>
        </div>
        <CloseButton onClick={onClose} />
      </div>

      <div className="grid gap-4 p-6 sm:grid-cols-2">
        {EDITABLE.slice(0, 4).map(([key, label]) => (
          <Field key={key} label={label}>
            <input {...input(key)} />
          </Field>
        ))}
        <Field label="Bio" className="sm:col-span-2">
          <textarea rows={3} {...input('bio')} />
        </Field>

        <dl className="grid grid-cols-[auto_1fr] gap-x-6 gap-y-1 text-xs sm:col-span-2">
          {user.email && (
            <>
              <dt className="text-muted">Email</dt>
              <dd>{user.email}</dd>
            </>
          )}
          <dt className="text-muted">Added</dt>
          <dd>{formatDate(user.created_at)}</dd>
          <dt className="text-muted">Last synced</dt>
          <dd>{timeAgo(user.synced_at)}</dd>
        </dl>

        {message && (
          <div className="sm:col-span-2">
            <Notice tone={message.tone}>{message.text}</Notice>
          </div>
        )}
      </div>

      <div className="flex flex-wrap items-center justify-between gap-3 border-t border-line bg-canvas px-6 py-4">
        <Button variant="danger" onClick={remove} disabled={busy}>
          <Icon.Trash />
          {confirmingDelete ? 'Click again to delete' : 'Delete'}
        </Button>
        <div className="flex gap-2">
          <Button variant="secondary" onClick={onClose}>
            Close
          </Button>
          <Button type="submit" disabled={!dirty || busy}>
            Save changes
          </Button>
        </div>
      </div>
    </form>
  )
}

// ---- Add a user ----------------------------------------------------------------

const BLANK_USER = { id: '', login: '', type: 'User', name: '', company: '', location: '', blog: '', email: '', bio: '' }

function AddUserForm({ onClose, onAdded }) {
  const [form, setForm] = useState(BLANK_USER)
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  async function submit(event) {
    event.preventDefault()
    // Leave empty optional fields out, so they're stored as NULL rather than ""
    const fields = Object.fromEntries(
      Object.entries({ ...form, id: Number(form.id) }).filter(([, value]) => value !== ''),
    )
    setBusy(true)
    const res = await api.createUser(fields)
    setBusy(false)
    if (!res.ok) return setError(errorMessage(res))
    onAdded()
    onClose()
  }

  // autoComplete off: these fields describe a GitHub user, not the person typing
  const input = (key) => ({
    name: key,
    autoComplete: 'off',
    value: form[key],
    onChange: (event) => setForm({ ...form, [key]: event.target.value }),
    className: inputClass,
  })

  return (
    <form onSubmit={submit}>
      <div className="flex items-start justify-between gap-4 border-b border-line p-6">
        <div>
          <h2 className="font-serif text-2xl leading-tight text-ink">Add a user</h2>
          <p className="mt-1 text-sm text-muted">
            Saved in PostgreSQL next to synced users. The ID and the login must both be unused.
          </p>
        </div>
        <CloseButton onClick={onClose} />
      </div>

      <div className="grid gap-4 p-6 sm:grid-cols-2">
        <Field label="GitHub ID">
          <input {...input('id')} type="number" min="1" required />
        </Field>
        <Field label="Login" hint="Letters, numbers, hyphens and underscores">
          <input {...input('login')} required spellCheck="false" />
        </Field>
        <Field label="Type">
          <select {...input('type')}>
            <option>User</option>
            <option>Organization</option>
            <option>Bot</option>
          </select>
        </Field>
        <Field label="Name">
          <input {...input('name')} />
        </Field>
        <Field label="Company">
          <input {...input('company')} />
        </Field>
        <Field label="Location">
          <input {...input('location')} />
        </Field>
        <Field label="Website">
          <input {...input('blog')} />
        </Field>
        <Field label="Email">
          <input {...input('email')} type="email" />
        </Field>
        <Field label="Bio" className="sm:col-span-2">
          <textarea rows={3} {...input('bio')} />
        </Field>
        {error && (
          <div className="sm:col-span-2">
            <Notice tone="red">{error}</Notice>
          </div>
        )}
      </div>

      <div className="flex justify-end gap-2 border-t border-line bg-canvas px-6 py-4">
        <Button variant="secondary" onClick={onClose}>
          Cancel
        </Button>
        <Button type="submit" disabled={busy}>
          Add user
        </Button>
      </div>
    </form>
  )
}

function CloseButton({ onClick }) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-label="Close"
      className="rounded-md p-1.5 text-muted transition hover:bg-bone hover:text-ink"
    >
      <Icon.Close />
    </button>
  )
}
