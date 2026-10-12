# golang-redis

A CLI and REST API that syncs GitHub users into PostgreSQL and serves them through a Redis cache.

## Requirements

- Go 1.26+
- PostgreSQL 13+ and Redis. `docker compose up -d --wait` starts both with settings that match the defaults.

## Setup

```sh
cp .env.example .env   # optional: the defaults work with docker compose; add GITHUB_TOKEN for a higher rate limit
docker compose up -d --wait
go build
./golang-redis setup-db
```

## CLI

```sh
./golang-redis sync-users -count 100 -store   # fetch from GitHub and store; run again to continue after the last stored user
./golang-redis list-users -limit 10 -format json
./golang-redis serve                          # REST API on API_HOST:API_PORT (default localhost:8080)
./golang-redis <command> -h                   # options for a command
```

Logs and status messages go to stderr, so `-format json` output can be piped, for example into `jq`.

## REST API

| Method | Path | Description |
| --- | --- | --- |
| GET | `/health` | Health check |
| GET | `/users?limit=30&offset=0` | List users stored in PostgreSQL |
| POST | `/users` | Create a user (`id`, `login` and `type` are required) |
| GET | `/users/:username` | Get a user from Redis, then PostgreSQL, then GitHub |
| PUT | `/users/:username` | Update `name`, `blog`, `location`, `company` or `bio` |
| DELETE | `/users/:username` | Delete a user from PostgreSQL |

Usernames are case-insensitive, like on GitHub. `GET /users/:username` caches what it finds, including "not found", for `REDIS_TTL_SECONDS`. Writes clear the user's cache entry, so the next GET sees the change. After a DELETE, GET falls back to GitHub (`"source": "github"`).

## Tests

```sh
go test ./...
TEST_POSTGRES=1 go test ./internal/database   # also run the PostgreSQL tests; they use the DB_* settings and roll back their changes
```
