// lookupPath works out which parts of the program a GET /users/:username request
// went through, using only what the API sent back:
//   - the X-Cache-Status header says whether Redis answered (hit) or not (miss)
//   - "source" in the body says who supplied the user: cache, database or github
// Each hop ends up as one of: ok, rejected, hit, miss, found, missing, error, skipped.
// "stored" says whether the answer was saved to Redis for next time.
export function lookupPath({ status, data = {}, cacheStatus, cacheError, cacheStoreError }) {
  const path = { api: 'ok', redis: 'skipped', postgres: 'skipped', github: 'skipped', stored: false }

  // The API sets X-Cache-Status on every lookup it handles, so without it the request
  // stopped at the API: a username it refused, or a server that isn't running
  if (!cacheStatus) {
    return { ...path, api: status === 0 || status >= 500 ? 'error' : 'rejected' }
  }
  if (cacheStatus === 'hit') return { ...path, redis: 'hit' }

  path.redis = cacheError ? 'error' : 'miss'
  const stored = !cacheStoreError
  if (status === 200 && data.source === 'database') return { ...path, postgres: 'found', stored }

  // A 500 names the store that failed: "Failed to fetch user from database" or "... github"
  if (status === 500 && !/github/i.test(data.error ?? '')) return { ...path, postgres: 'error' }

  // PostgreSQL didn't have the user, so the API asked GitHub
  path.postgres = 'missing'
  if (status === 200) return { ...path, github: 'found', stored }
  if (status === 404) return { ...path, github: 'missing', stored }
  return { ...path, github: 'error' }
}
