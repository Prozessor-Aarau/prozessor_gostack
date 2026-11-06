/# Prozessor GoStack

Web app for running the Prozessor makerspace: visitor registration and check-in, badges (NFC card reader), machines and defect reports, supervision and opening hours, logistics requests and billing (Swiss QR bills), newsletters, statistics and a public calendar.

The UI is in German.

## Tech Stack

- **Go** with [chi](https://github.com/go-chi/chi) router and [scs](https://github.com/alexedwards/scs) sessions
- **[templ](https://templ.guide)** for HTML templates, **HTMX** for interactivity
- **Tailwind CSS v4** + **daisyUI** for styling
- **PostgreSQL** with **[goose](https://github.com/pressly/goose)** migrations
- **[Air](https://github.com/air-verse/air)** for hot reload during development

## Project Structure

| Path               | Contents                                                       |
| ------------------ | -------------------------------------------------------------- |
| `cmd/`             | Entry point, routes, middleware, mail sender                   |
| `internal/`        | Handlers, models, repository (DB access), config, helpers, etc. |
| `view/`            | `.templ` templates, grouped by feature                         |
| `static/`          | CSS, images, manifest; generated files (QR codes, bills, uploads) |
| `email-templates/` | HTML templates for outgoing mails                              |
| `migrations/`      | goose SQL migrations                                           |

## Getting Started

### 1. Install dependencies

With Nix (recommended), everything is provided by the flake:

```sh
nix develop        # or `direnv allow` if you use direnv
```

Without Nix, install manually: Go, Node.js, `templ`, `air`, `goose`, PostgreSQL, and `pcsclite` + `pkg-config` (needed for the card reader, CGO must be enabled).

Then install the Node packages:

```sh
npm install
```

### 2. Configure

```sh
cp .env.example .env
```

Fill in at least the database settings (`DBHOST`, `DBNAME`, `DBUSER`, `DBPASS`). For local development set `PRODUCTION=false` and `CACHE=false`. See `.env.example` for all options (Slack, FTP, SMTP, label printer, billing details).

For local mail testing you can use [Mailpit](https://mailpit.axllent.org) (included in the Nix shell): run `mailpit`, then set `SMTP_HOST=localhost`, `SMTP_PORT=1025`, `SMTP_ENCRYPTION=false`, and open http://localhost:8025 to see sent mails.

### 3. Set up the database

goose reads the `GOOSE_*` variables from `.env`. Set `GOOSE_DBSTRING` to match your database settings, then run:

```sh
goose up
```

Create a new migration with `goose create <name> sql`.

### 4. Run

```sh
air
```

This builds CSS, generates templates, compiles and restarts on every change. The app runs on the `PORT` from `.env` (default http://localhost:4000).

### 5. Create the first admin

Logins are created via an invite link. To bootstrap the first account:

1. Register at `/register`.
2. In the database, give yourself the Superadmin role and a register code:
   ```sql
   UPDATE visitors SET role_id = 8, register_code = 'some-code' WHERE email = 'you@example.com';
   ```
3. Open `/createlogin?email=you@example.com&uid=some-code` and set a password.

After that, further logins can be sent from a visitor's profile in the admin area.

## Common Commands

| Task                | Command                                                       |
| ------------------- | ------------------------------------------------------------- |
| Dev server          | `air`                                                         |
| Build CSS           | `npm run build` (or `npm run watch`)                          |
| Generate templates  | `templ generate`                                              |
| Build binary        | `go build -C ./cmd -o ../tmp/main .`                          |
| Full build          | `npm run build && templ generate && go build -C ./cmd -o ../tmp/main .` |
| Format / vet        | `go fmt ./...` / `go vet ./...`                               |
| Tests               | `go test ./...`                                               |

Generated files (`*_templ.go`, `static/css/tailwind.css`) are git-ignored. Run the full build after pulling.

## External Services (optional)

- **Slack**: notifications for logistics requests (`SLACKTOKEN`, `LOGISTIC_SLACK_CHANNEL`)
- **FTP**: events for the calendar and newsletter are read from Markdown files with frontmatter (`FTP_*`)
- **Label printer**: QR codes and labels are sent to a print server (`QR_PRINT_SERVER`)
- **NFC card reader**: badge scanning via PC/SC on the server machine

The app starts without them, but the related features will not work.

## Deployment

Production runs as a systemd service. See `prozessor.service.example` for the unit file and `deploy.sh` for the deploy steps (pull, build, migrate, restart).

## Contributing

- Work on a branch and open a pull request against `main`; all changes require review by the code owner (see `.github/CODEOWNERS`).
- Run `go fmt ./...` and `go vet ./...` before committing.
- Schema changes always go into a new migration; never edit existing ones.
- See `AGENTS.md` for code style conventions and `TODO.md` for planned work.
