# BillionMail

Open-source email marketing platform & mail server. Handles bulk sending, campaigns, contact management, warmup, and analytics.

## Project Structure

```
core/
├── internal/
│   ├── cmd/            # CLI entry points
│   ├── controller/     # HTTP handlers (16+ domains)
│   ├── service/        # Business logic (batch_mail, domains, rbac, maillog_stat, warmup, contact, etc.)
│   ├── dao/            # Data access layer
│   ├── model/entity/   # ORM entities
│   └── consts/         # Constants
├── api/                # API route definitions
├── frontend/src/
│   ├── views/          # Page components
│   ├── components/     # Reusable UI components
│   ├── store/          # Pinia store modules
│   ├── api/modules/    # API client modules
│   ├── router/         # Vue Router + module routes
│   ├── hooks/          # Composables
│   ├── utils/          # Utilities (base, data, time, storage)
│   ├── features/       # Feature-specific components (EmailEditor)
│   └── i18n/           # Internationalization (en, zh, ja)
├── template/           # Email templates
└── manifest/           # Config/deployment
conf/                   # Mail service configs (postfix, dovecot, rspamd, redis)
Dockerfiles/            # Container definitions
```

## Stack

- **Backend:** Go 1.22, GoFrame v2, PostgreSQL, Redis
- **Frontend:** Vue 3, TypeScript, Pinia, Naive UI, Vitest, pnpm
- **Mail:** Postfix, Dovecot, Rspamd
- **Deploy:** Docker Compose

## Organization Rules

- Controllers → `core/internal/controller/`, one dir per domain
- Services → `core/internal/service/`, one dir per domain
- API routes → `core/api/`, one dir per domain
- Frontend views → `core/frontend/src/views/`, one dir per feature
- Tests → next to source files (`*_test.go`, `*.test.ts`)
- Single responsibility per file, descriptive names

## Code Quality

After editing ANY file, run:

```bash
# Go
cd core && go vet ./... && gofmt -l .

# Frontend
cd core/frontend && pnpm run lint
```

Fix ALL errors before continuing.

## Testing

```bash
# Go (short mode, no DB required)
cd core && go test -count=1 -short ./internal/service/...

# Frontend
cd core/frontend && pnpm test

# Both
cd core && go test -count=1 -short ./internal/service/... && cd frontend && pnpm test
```

## Bug Memory Rules

- NEVER use bare domain for mail infrastructure (DNS, certs, DKIM, dedicated IPs). Always use `public.FormatMX(domain)` to get the mail hostname (e.g., `mail.example.com`).
- When adding new controllers, add the module name to the RBAC `modules` list in `core/internal/service/middlewares/rbac.go`.

## Commands

- `/test` - Run full test suite
- `/fix` - Lint + typecheck + auto-fix with parallel agents
- `/commit` - Quality checks + AI commit + push
- `/update-app` - Update deps + fix deprecations
