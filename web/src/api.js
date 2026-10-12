// Calls the Go REST API through the dev server's /api proxy (see vite.config.js).
// Every call resolves to a result instead of throwing, so the UI can show what happened.
async function request(method, path, body) {
  const started = performance.now()
  try {
    const res = await fetch(`/api${path}`, {
      method,
      headers: body ? { 'Content-Type': 'application/json' } : undefined,
      body: body ? JSON.stringify(body) : undefined,
    })
    return {
      ok: res.ok,
      status: res.status,
      data: await res.json().catch(() => ({})),
      cacheStatus: res.headers.get('X-Cache-Status'),
      cacheError: res.headers.has('X-Cache-Error'),
      cacheStoreError: res.headers.has('X-Cache-Store-Error'),
      ms: Math.round(performance.now() - started),
    }
  } catch {
    return { ok: false, status: 0, data: {}, ms: Math.round(performance.now() - started) }
  }
}

const userPath = (login) => `/users/${encodeURIComponent(login)}`

export const api = {
  health: () => request('GET', '/health'),
  listUsers: (limit, offset) => request('GET', `/users?limit=${limit}&offset=${offset}`),
  getUser: (login) => request('GET', userPath(login)),
  createUser: (fields) => request('POST', '/users', fields),
  updateUser: (login, fields) => request('PUT', userPath(login), fields),
  deleteUser: (login) => request('DELETE', userPath(login)),
}

// errorMessage turns a failed result into a sentence for the UI. The API always
// explains its own errors in JSON, so a 5xx without one means it never answered.
export function errorMessage(res) {
  if (res.data?.error) return res.data.error
  if (res.status === 0 || res.status >= 500) return "Can't reach the API. Is ./golang-redis serve running?"
  return `Request failed with status ${res.status}`
}
