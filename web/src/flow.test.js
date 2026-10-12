import assert from 'node:assert/strict'
import { test } from 'node:test'
import { lookupPath } from './flow.js'

const untouched = { redis: 'skipped', postgres: 'skipped', github: 'skipped', stored: false }

test('an invalid username stops at the API', () => {
  assert.deepEqual(lookupPath({ status: 400, data: { error: 'Invalid username' } }), { api: 'rejected', ...untouched })
  // A "/" in the name doesn't match any route, so Gin answers 404 without a cache header
  assert.deepEqual(lookupPath({ status: 404 }), { api: 'rejected', ...untouched })
})

test('an unreachable API is an error', () => {
  assert.deepEqual(lookupPath({ status: 0 }), { api: 'error', ...untouched })
  assert.deepEqual(lookupPath({ status: 502 }), { api: 'error', ...untouched })
})

test('a cache hit is answered by Redis, including a cached "not found"', () => {
  const expected = { api: 'ok', ...untouched, redis: 'hit' }
  assert.deepEqual(lookupPath({ status: 200, cacheStatus: 'hit', data: { source: 'cache' } }), expected)
  assert.deepEqual(lookupPath({ status: 404, cacheStatus: 'hit' }), expected)
})

test('a stored user comes from PostgreSQL and is cached', () => {
  assert.deepEqual(lookupPath({ status: 200, cacheStatus: 'miss', data: { source: 'database' } }), {
    api: 'ok',
    redis: 'miss',
    postgres: 'found',
    github: 'skipped',
    stored: true,
  })
})

test('a user who is not stored comes from GitHub and is cached', () => {
  assert.deepEqual(lookupPath({ status: 200, cacheStatus: 'miss', data: { source: 'github' } }), {
    api: 'ok',
    redis: 'miss',
    postgres: 'missing',
    github: 'found',
    stored: true,
  })
})

test('an unknown user is looked for everywhere, and the miss is cached', () => {
  assert.deepEqual(lookupPath({ status: 404, cacheStatus: 'miss' }), {
    api: 'ok',
    redis: 'miss',
    postgres: 'missing',
    github: 'missing',
    stored: true,
  })
})

test('a 500 points at the store that failed', () => {
  const database = lookupPath({ status: 500, cacheStatus: 'miss', data: { error: 'Failed to fetch user from database' } })
  assert.equal(database.postgres, 'error')
  assert.equal(database.github, 'skipped')

  const github = lookupPath({ status: 500, cacheStatus: 'miss', data: { error: 'Failed to fetch user from github' } })
  assert.equal(github.postgres, 'missing')
  assert.equal(github.github, 'error')
})

test('Redis problems are reported without stopping the lookup', () => {
  const path = lookupPath({
    status: 200,
    cacheStatus: 'miss',
    cacheError: true,
    cacheStoreError: true,
    data: { source: 'database' },
  })
  assert.equal(path.redis, 'error')
  assert.equal(path.postgres, 'found')
  assert.equal(path.stored, false)
})
