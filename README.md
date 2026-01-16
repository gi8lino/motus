# Motus

![motus.png](./web/public/brand.svg)

[![Go Report Card](https://goreportcard.com/badge/github.com/gi8lino/motus?style=flat-square)](https://goreportcard.com/report/github.com/gi8lino/motus)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/gi8lino/motus)
[![Release](https://img.shields.io/github/release/gi8lino/motus.svg?style=flat-square)](https://github.com/gi8lino/motus/releases/latest)
[![GitHub tag](https://img.shields.io/github/tag/gi8lino/motus.svg?style=flat-square)](https://github.com/gi8lino/motus/releases/latest)
![Tests](https://github.com/gi8lino/motus/actions/workflows/tests.yml/badge.svg)
[![Build](https://github.com/gi8lino/motus/actions/workflows/release.yml/badge.svg)](https://github.com/gi8lino/motus/actions/workflows/release.yml)
[![license](https://img.shields.io/github/license/gi8lino/motus.svg?style=flat-square)](LICENSE)

---

Motus is a training companion for circuit and round-based workouts. It keeps track of which exercise comes next, how many reps you need, and when to take a break. Timers run in the browser, while trainings, workouts, templates, and history are stored in PostgreSQL.

## Features

- Create users and manage local accounts.
- Build workouts with sets, pauses, and repeatable steps using Go-style duration strings (e.g. `45s`, `1m30s`).
- Attach multiple exercises to each set; each exercise can track reps/weight or a timed duration.
- Maintain a global exercise catalog with Core + Personal exercises.
- Share workouts as templates and apply them to create new workouts.
- Choose from bundled sound cues for steps and countdowns.
- Track trainings with history, summaries, and per-step timing.
- Export/import workouts as JSON for backup or sharing.
- Support dark/light/auto themes and per-user settings.

## How it works

Motus separates training timing from data storage:

- The browser runs timers and sounds for steps.
- The backend stores workouts, exercises, templates, trainings, and history.
- When a training finishes, a summary is generated for copying into an AI or notes app.

## Using the app

1. **Log in or register** (local mode) or rely on proxy auth (auth-header mode).
2. **Create workouts** in the Workouts tab. Add sets and pauses.
3. **Select a workout** in Train and click Start to begin.
4. **Pause/resume/advance** steps, and use the summary when finished.
5. **Review history** to inspect per-step target vs. actual timing.
6. **Manage exercises** in the Exercises tab (Core vs Personal).
7. **Export/import** workouts from the Profile page.

## Workout steps

- **Set**: a regular exercise block that you complete, then move on. Each exercise inside the set can be rep-based or timed.
- **Pause**: a rest block; can auto-advance on a countdown.
- **Repeat options**: repeat a set multiple times with a configurable pause between repeats.

Each step can include multiple exercises and a sound cue. Auto-advance pauses trigger a visible countdown. Training summaries include target vs. actual time so you can paste the recap into your preferred AI and ask how the training went.

## Running locally

1. Set up PostgreSQL and export the connection string. Example using Docker:

   ```bash
   docker run --name motus-db -e POSTGRES_PASSWORD=motus -e POSTGRES_USER=motus -e POSTGRES_DB=motus -p 5432:5432 -d postgres:18
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
- `--core-exercises-file` (default empty): path to a YAML file of core exercises to seed on startup.
- `--admin-email` (default empty): admin email to bootstrap or update at startup.
- `--admin-password` (default empty): admin password to bootstrap or update at startup.
- `--debug` (default false): enable debug logging.
- `--log-format` (default `json`): `json` or `text`.

## Auth header mode

When `--auth-header` is set, Motus trusts the specified header as the authenticated user ID (email). The UI switches to proxy-auth mode, disables local login, and expects the reverse proxy to inject a valid email address. If you also set `--auto-create-users`, Motus will create missing users on first access. When the header is not set, Motus runs in local-auth mode and requires email + password.

## Local admin bootstrap

To auto-create or update a local admin account at startup, use the admin flags (or env vars with `MOTUS_` prefix). The server logs when it creates or updates the admin user.

To promote an existing user to admin directly in the database:

```sql
UPDATE users SET is_admin = TRUE WHERE id = 'user@example.com';
```

## Core exercises YAML

Use `--core-exercises-file` to seed a set of core exercises at startup. The file must be YAML with a top-level `exercises` list of strings:

```yaml
exercises:
  - Push-up
  - Squat
  - 1 Pump Burpee
```

See `examples/core-exercises.yaml` for a full example.

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

Use the Workouts tab to design trainings. When you’re ready to share one, select it and hit “Share”. The Templates tab lists all shared workouts and lets you instantiate a new workout in one click.

## License

This project is licensed under the Apache 2.0 License. See the [LICENSE](LICENSE) file for details.
