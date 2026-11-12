# jobFlow

A job board and professional social network built in Go. Individuals build
profiles, post updates, follow each other and apply for jobs. Companies post
openings and review applications.

## Features

- Registration, login and password reset by email code
- Profiles with picture upload, followers and following
- Posts, comments, sharing and notifications (live updates over SSE)
- Job listings, applications and a company applications view
- Security: CSRF protection, hashed passwords, session expiry and cleanup,
  login throttling

## Tech stack

- Go (`net/http`, `html/template`), no web framework
- SQLite via `modernc.org/sqlite` (pure Go, WAL mode)
- Docker for deployment, GitHub Actions for CI

## Project layout

| Folder        | Purpose                                        |
| ------------- | ---------------------------------------------- |
| `handlers/`   | HTTP handlers, one file per feature            |
| `middleware/` | Auth and CSRF middleware                       |
| `database/`   | Schema and all SQL queries                     |
| `models/`     | Data structures                                |
| `utils/`      | Template rendering, uploads, client IP helpers |
| `templates/`  | HTML templates                                 |

## Running locally

1. Install Go 1.26 or newer.
2. Copy `.env.example` to `.env` and fill in the values.
3. Start the app:

   ```bash
   go run .
   ```

4. Open http://localhost:8081

## Running with Docker

```bash
docker build -t jobflow .
docker run --rm -p 8081:8081 --env-file .env -v jobflow-data:/data jobflow
```

The database and uploaded files live in the `/data` volume, so they survive
restarts.

## Configuration

| Variable            | Purpose                                         | Default       |
| ------------------- | ----------------------------------------------- | ------------- |
| `PORT`              | Port to listen on                               | `8081`        |
| `DATABASE_PATH`     | SQLite file location                            | `database.db` |
| `UPLOAD_PATH`       | Folder for uploaded files                       | `uploads`     |
| `SMTP_USER`         | Gmail address that sends reset codes            | none          |
| `SMTP_APP_PASSWORD` | Gmail app password (not your account password)  | none          |
| `RESET_CODE_SECRET` | Secret that protects password-reset tokens      | none          |
| `TRUST_PROXY`       | Set to `true` only when behind a trusted proxy  | off           |

Never commit `.env`. Use `.env.example` as the template.

## Tests

```bash
go vet ./...
go test ./...
```

CI runs the same checks on every push.

## Roadmap

AI-assisted CV generation, job recommendations and skill-gap advice for
candidates, plus screening support for companies, always with a human making
the final decision.