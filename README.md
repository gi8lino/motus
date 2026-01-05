# Motus

Motus is a lightweight training tracker with a React SPA front end and a Go JSON API backend. Timers run in the browser, while sessions, workouts, templates, and history are stored in PostgreSQL.

## Features

- Create users and manage local accounts.
- Build workouts with sets, pauses, and timed blocks using Go-style duration strings (e.g. `45s`, `1m30s`).
- Attach multiple exercises with amount/weight to each set; drag/drop steps and exercises.
- Maintain a global exercise catalog with Core + Personal exercises.
- Share workouts as templates and apply them to create new workouts.
- Choose from bundled sound cues for steps and countdowns.
- Track sessions with history, summaries, and per-step timing.
- Export/import workouts as JSON for backup or sharing.
- Support dark/light/auto themes and per-user settings.

## How it works

Motus separates session timing from data storage:

- The browser runs timers and sounds for steps.
- The backend stores workouts, exercises, templates, sessions, and history.
- When a session finishes, a summary is generated for copying into an AI or notes app.

## Using the app

1. **Log in or register** (local mode) or rely on proxy auth (auth-header mode).
2. **Create workouts** in the Workouts tab. Add sets, pauses, or timed steps.
3. **Select a workout** in Sessions and click Start Session to begin.
4. **Pause/resume/advance** steps, and use the summary when finished.
5. **Review history** to inspect per-step target vs. actual timing.
6. **Manage exercises** in the Exercises tab (Core vs Personal).
7. **Export/import** workouts from the Profile page.

## Workout steps

- **Set**: a standard exercise block; target is informational.
- **Pause**: optional auto-advance countdown; great for rest.
- **Timed set**: countdown or stopwatch with optional transition time.

Each step can include multiple exercises and a sound cue. Auto-advance pauses trigger a visible countdown.

## Running locally

1. Set up PostgreSQL and export the connection string. Example using Docker:

   ```bash
   docker run --name motus-db -e POSTGRES_PASSWORD=motus -e POSTGRES_USER=motus -e POSTGRES_DB=motus -p 5432:5432 -d postgres:16
   export MOTUS_DATABASE_URL="postgres://motus:motus@localhost:5432/motus?sslmode=disable"
   ```

   (Adjust the credentials and host if you already have a local Postgres instance.)

2. Install dependencies and build:

   ```bash
   go build ./...
   ```

3. Run the server:

   ```bash
   go run .
   ```

4. Open <http://localhost:8080> to use the UI.

## CLI flags

Motus supports a handful of runtime flags (or environment variables with the `MOTUS_` prefix):

- `--database-url` (required): PostgreSQL connection string.
- `--listen-address` (default `:8080`): server bind address.
- `--route-prefix` (default empty): mount the app under a path (e.g. `/motus`).
- `--site-root` (default `http://localhost:8080`): base URL used in links.
- `--auth-header` (default empty): header name to trust for proxy auth.
- `--allow-registration` (default false): allow local user sign-up.
- `--auto-create-users` (default false): auto-create users when auth-header is enabled.
- `--debug` (default false): enable debug logging.
- `--log-format` (default `json`): `json` or `text`.

## Auth header mode

When `--auth-header` is set, Motus trusts the specified header as the authenticated user ID (email). The UI switches to proxy-auth mode, disables local login, and expects the reverse proxy to inject a valid email address. If you also set `--auto-create-users`, Motus will create missing users on first access. When the header is not set, Motus runs in local-auth mode and requires email + password.

## Local admin bootstrap

To auto-create or update a local admin account at startup, use the admin flags (or env vars with `MOTUS_` prefix). The server logs when it creates or updates the admin user.

## Docker Compose (DB + pgAdmin)

Use `docker-compose.db.yml` for the database and pgAdmin:

```bash
docker compose -f docker-compose.db.yml up -d
```

pgAdmin runs at <http://localhost:5050>.

Connection settings inside pgAdmin:

- Host name/address: `motus-db`
- Port: `5432`
- Maintenance database: `motus`
- Username: `motus`
- Password: `motus`

Default pgAdmin login:

- Email: `admin@motus.app`
- Password: `motus`

## Sound cues

Bundled audio files live in `web/assets/sounds`. The UI exposes them in each step so you can pick a cue without providing external URLs.

## Templates

Use the Workouts tab to design sessions. When you’re ready to share one, select it and hit “Share”. The Templates tab lists all shared workouts and lets you instantiate a new workout in one click.
