# golang-redis

A CLI, REST API and web UI that sync GitHub users into PostgreSQL and serve them through a Redis cache.

## Quick start with Docker

```sh
cp .env.example .env                                                # optional: change ports or add a GITHUB_TOKEN
docker compose up -d --build --wait                                 # PostgreSQL, Redis, the API and the web UI
docker compose exec api golang-redis sync-users -count 100 -store   # copy the first 100 GitHub users
```

Then open the web UI at http://localhost:8081. The REST API is at http://localhost:8080, and it creates the database schema when it starts.

If `docker compose up` says a port is already allocated (a local PostgreSQL on 5432, or another Redis on 6379), set `DB_PORT`, `REDIS_PORT`, `API_PORT` or `WEB_PORT` in `.env`. Docker Compose and the app both read it, so they stay in agreement.

## Developing locally

Needs Go 1.26+ and Node 22+. Run only the databases in Docker, and the API and web UI on your machine:

```sh
docker compose up -d --wait postgres redis
go build
./golang-redis setup-db
./golang-redis serve                    # the API, on localhost:8080
cd web && npm install && npm run dev    # the web UI with live reload, on localhost:5173
```

The dev server forwards `/api` to the API; set `API_URL` if the API runs somewhere else. To use your own PostgreSQL instead of the container, it needs to be version 13 or newer (for the `citext` extension).

## CLI

```sh
./golang-redis sync-users -count 100 -store   # fetch from GitHub and store; run again to continue after the last stored user
./golang-redis list-users -limit 10 -format json
./golang-redis serve                          # REST API on API_HOST:API_PORT (default localhost:8080)
./golang-redis <command> -h                   # options for a command
```

Logs and status messages go to stderr, so `-format json` output can be piped, for example into `jq`.

## Web UI

`web/` holds a React and Tailwind front end. It lists, looks up, adds, edits and deletes users, and its About page explains how the program works, including a live trace of a lookup through Redis, PostgreSQL and GitHub.

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
cd web && npm test                             # the web UI's lookup-tracing logic
```
