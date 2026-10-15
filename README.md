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

## Local development ports

With Python 3 installed, `bin/dev-port` saves named ports in the
Git-ignored `.dev-ports.json`. Separate Make commands reuse the same ports:

```sh
make download
make dev-build
make postgres
make serve
```

`make run` starts the application with its dependencies; `make ports` prints
saved addresses. `bin/dev-port app` retrieves an assignment;
`--port 8080` saves an explicit port. Stop local services before `make ports-reset`.
Saved ports are reused even when occupied, and are not reserved between runs.

Shared development tools come from [gi8lino/dev-tools](https://github.com/gi8lino/dev-tools).
`make dev-tools` downloads the pinned `DEV_TOOLS_VERSION` into `bin/`; startup
targets install them automatically. To upgrade, change that version in the
Makefile and run `make dev-tools`. `make open` opens the browser once the app responds.
`make run` starts the app and opens the browser automatically.

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

1. **Choose an authentication mode**: local accounts, a trusted proxy, OpenID Connect, or an explicitly unauthenticated single-user deployment.
2. **Create workouts** in the Workouts tab. Add sets and pauses.
3. **Select a workout** in Training and click Start to begin.
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
   export MOTUS__DATABASE_URL="postgres://motus:motus@localhost:5432/motus?sslmode=disable"
   ```

   (Adjust the credentials and host if you already have a local Postgres instance.)

2. Install dependencies and build:

   ```bash
   go build ./...
   ```

3. Run the server:

   ```bash
   go run ./cmd/motus
   ```

4. Open <http://localhost:8080> to use the UI.

## CLI flags

Motus can be configured using command-line flags or environment variables. Environment variables use the `MOTUS__` prefix.

- `--database-url` (required): PostgreSQL connection string.
- `--listen-address` (default `127.0.0.1:8080`): public application bind address.
- `--management-listen-address` (default `127.0.0.1:8081`): health/readiness bind address.
- `--route-prefix` (default empty): mount the public app under a path (e.g. `/motus`).
- `--public-url` (default `http://localhost:8080`): externally visible base URL used for browser callbacks.
- `--auth-mode` (optional deployment override): one of `local`, `proxy`, `oidc`, or `none`.
- `--trusted-proxy-cidrs`: allowed immediate proxy CIDRs, or an explicit `*` to trust every peer.
- `--trusted-username-headers`, `--trusted-email-headers`: ordered identity headers used in proxy mode; deployment overrides take precedence over persisted header settings.
- `--allow-user-registration` (optional override): control automatic registration of external identities.
- `--session-lifetime` (default `720h`): lifetime of local and OIDC authentication sessions.
- `--oidc-issuer`, `--oidc-client-id`, `--oidc-client-secret`, and `--oidc-session-secret`: configure browser OIDC login in `oidc` mode. The session secret must be at least 32 bytes.
- `--encryption-key`: optional base64-encoded 256-bit key for sensitive persisted settings.
- `--core-exercises-file` (default empty): optional YAML file overriding the embedded core exercise catalog.
- `--debug` (default false): enable debug logging.
- `--access-log` (default false): log completed application requests.
- `--log-format` (default `json`): `json` or `text`.

The management listener is deliberately separate from the public application listener. It exposes `GET /healthz` for process liveness and `GET /readyz` for PostgreSQL readiness, and it is not mounted below `--route-prefix`.

## Local mode and first-run setup

Local authentication is the default. On an empty database, open Motus and use the first-run setup screen to create the initial administrator. The setup endpoint can create the first user exactly once; there are no startup admin flags.

After setup, additional local users can register only when `--allow-user-registration` is enabled. Administrators can manage users from the application.

## Trusted-proxy mode

Use `proxy` mode when an authentication gateway verifies a user's identity and forwards an identity header. For an Envoy Gateway SecurityPolicy that sets `X-Auth-Request-Email` from a verified JWT:

```yaml
- name: MOTUS__AUTH_MODE
  value: proxy
- name: MOTUS__TRUSTED_PROXY_CIDRS
  value: "*"
```

Motus uses appkit's default identity headers when none are persisted: `X-Forwarded-User`, `X-Auth-Request-User`, and `Remote-User` for usernames; `X-Forwarded-Email`, `X-Auth-Request-Email`, and `X-Authentik-Email` for email. Nonempty persisted header lists take precedence over these defaults; explicitly set CLI/environment headers take precedence over both. Empty persisted lists use the defaults without modifying PostgreSQL.

**Security:** Only use `*` when network policy restricts access to the authenticated Gateway. The Gateway must strip or overwrite **every** accepted identity header supplied by clients. To allow only Envoy's verified email header (recommended when other forwarded headers aren't guaranteed to be sanitized), explicitly configure:

```yaml
- name: MOTUS__TRUSTED_USERNAME_HEADERS
  value: X-Auth-Request-Email
- name: MOTUS__TRUSTED_EMAIL_HEADERS
  value: X-Auth-Request-Email
```

With specific ingress IPs, use their CIDRs instead of `*`. Leaving the CIDR list unset fails closed.

Existing Motus users are resolved by email. To provision previously unknown proxy identities automatically, also set:

```yaml
- name: MOTUS__ALLOW_USER_REGISTRATION
  value: "true"
```

New proxy-provisioned users are not administrators automatically. For a new installation, either create the initial administrator in local mode first using the same email address and then switch to trusted-proxy mode, or promote the provisioned user explicitly.

For an Envoy Gateway SecurityPolicy that injects `X-Auth-Request-Email`, the same username and email header settings apply. Its JWT must contain the `email` claim. If proxy Pod addresses change, `MOTUS__TRUSTED_PROXY_CIDRS="*"` is an option **only** when network policy prevents all direct access to the Motus listener.

`--trusted-proxy-cidrs="*"` (or `MOTUS__TRUSTED_PROXY_CIDRS="*"`) explicitly trusts **every valid immediate peer**. Do not use this without a network boundary restricting direct access to the authentication proxy, which must strip incoming identity headers. A wildcard cannot be combined with CIDRs; leaving the list empty still denies proxy authentication.

## OIDC mode

Set `--auth-mode=oidc` and all four OIDC flags to use a single OpenID Connect provider. Motus uses the authorization-code flow with PKCE and nonce/state validation, verifies the provider ID token, and creates its normal opaque Motus session. The provider callback is `<public-url>/auth/oidc/callback`; add that exact URL to the provider client. New verified identities are admitted only when `--allow-user-registration` is enabled.

## No-auth mode

`--auth-mode=none` is for a genuinely single-user deployment behind an already trusted boundary. Requests run as Motus's local `admin` account, created only when this mode first handles a request. Do not expose this mode directly to the internet.

## Core exercises YAML

Motus embeds and reconciles its default core exercises on every startup, including in the container image. No file mount is required. To replace the built-in catalog, pass `--core-exercises-file` (or `MOTUS__CORE_EXERCISES_FILE`) with a mounted YAML file. Every entry has a concise name and optional searchable labels. Set `hasSides: true` for unilateral exercises; when omitted, it defaults to `false`. `previousNames` lets catalog upgrades rename existing entries and linked workout snapshots safely:

```yaml
version: 2
exercises:
  - { name: Push-up, labels: [bodyweight, push, chest] }
  - { name: Swing, labels: [kettlebell, hinge], previousNames: [Kettlebell Swing] }
  - { name: Side Plank, hasSides: true, labels: [bodyweight, core] }
```

See `examples/core-exercises.yaml` for the embedded default and override format.

## Kubernetes

The manifests in `deploy/kubernetes` use the Gateway API and default to local authentication so a fresh deployment can complete first-run setup without an external identity provider.

`deployment.yaml` sets the application and management listeners, the public URL, and `MOTUS__AUTH_MODE=local` explicitly. `httproute.yaml` expects a Gateway named `external` in the `motus` namespace; change the parent reference and hostname to match your cluster.

For oauth2-proxy or another trusted authentication proxy, replace the local auth environment setting with the trusted-proxy variables shown above.

## Docker Compose

Start the local PostgreSQL service with the same Compose file used by the Makefile:

```bash
docker compose -f deploy/motus/docker-compose.db.yml up -d
```

The database listens on `127.0.0.1:5431` by default; set `MOTUS_POSTGRES_PORT` to choose a different host port.

## Sound cues

Bundled audio files live in `web/assets/sounds`. The UI exposes them in each step so you can pick a cue without providing external URLs.

## Templates

Use the Workouts tab to design trainings. When you’re ready to share one, select it and hit “Share”. The Templates tab lists all shared workouts and lets you instantiate a new workout in one click.

## License

This project is licensed under the Apache 2.0 License. See the [LICENSE](LICENSE) file for details.


## Authentication proxy boundary

When using the `proxy` or legacy `trusted-proxy` authentication mode, configure
`--trusted-proxy-cidrs` with the CIDR addresses of the **immediate, verified ingress peers**.
Proxy authentication fails at startup if the list is empty or invalid.

```sh
MOTUS__TRUSTED_PROXY_CIDRS=127.0.0.1/32
```

The ingress must **strip incoming client-supplied identity headers** and inject its own.
Only configure peers controlled by your deployment; do not use unrestricted CIDRs.
The application never trusts `X-Forwarded-For` as evidence of an authenticated peer.

Appkit-managed sessions use purpose-separated hashes; existing Motus session cookies
remain valid through the legacy lookup path until they expire or are revoked.


