# go-hotels

A professional-grade Go backend architecture showcase project for hotel-related operations.

## 🎯 Purpose
This project is not about demonstrating business logic complexity — it’s about showcasing:
- Clean architecture
- Microservices
- Observability
- DevSecOps
- Infrastructure-as-Code
- CI/CD pipelines
- Pub/Sub
- REST + gRPC
- Security best practices (incl. OWASP Top 10)
- End-to-end traceability

## 🏗️ Architecture
```
            ┌────────────┐           ┌────────────────┐
            │  Clients   │ ───────▶ │  hotel-api      │
            └────────────┘           └────────────────┘
                                           │
       REST + OpenAPI                     │  NATS Event: InvoiceRequested
                                           ▼
                                  ┌──────────────────┐
                                  │ invoice-service   │
                                  └──────────────────┘
```

## 🧱 Tech Stack

### Language & Framework
- **Go 1.24.x**
- **chi** (router)
- **oapi-codegen** (OpenAPI 3 codegen)

### Infrastructure
- **Docker** for containerization
- **Docker Compose** for local orchestration
- **PostgreSQL** as primary DB
- **NATS** for event-driven pub/sub

### Observability
- **Grafana Alloy** for logs, metrics, and tracing
- **Prometheus** as metrics backend
- **Loki** and **Tempo** for logs and traces
- **OpenTelemetry** SDK in Go

### Developer Experience
- **Task** (make-like runner)
- **golangci-lint** for linting
- **go tool** with `tool` directives in go.mod
- **VS Code + gopls** for LSP

### DevSecOps
- **Trivy** or **grype** for container scanning
- **gosec** or `golangci-lint run --enable=gosec`
- GitHub Dependabot
- (Planned) GitHub Actions CI

### Security
- **API key auth** (planned)
- **OWASP Top 10** mitigations (to be covered)
- **mTLS between services** (planned)

## 🧩 Microservices

### 1. hotel-api (Monolith-style HTTP API)
Handles:
- List hotels
- Search by city/checkin/checkout
- Create bookings
- Emits `InvoiceRequested` event via NATS

### 2. invoice-service (Worker-style async service)
Handles:
- Listens to `InvoiceRequested`
- Generates invoice (fake JSON or PDF)
- Emits logs/metrics/traces

## 🔁 API Endpoints
Defined in `openapi/openapi.yaml`, generated with `oapi-codegen`:
- `GET /hotels`
- (soon) `POST /bookings`

## 📂 Structure
```
go-hotels/
├── cmd/
│   ├── api/                 # Entrypoint for hotel-api
│   └── invoice/             # Entrypoint for invoice-service
│
├── internal/
│   ├── hotel/               # Hotel business logic
│   ├── invoice/             # Invoice logic (worker)
│   ├── infra/               # Shared infra: DB, NATS, Redis
│   ├── platform/            # Observability, logging, security
│   └── api/                 # REST handlers
│
├── openapi/                # OpenAPI spec & config
├── proto/                  # gRPC definitions (future)
├── Taskfile.yml            # All dev commands
├── go.mod
└── README.md
```

## 🚀 Tasks
```sh
go install github.com/go-task/task/v3/cmd/task@latest
```
Then run:
```sh
task generate   # Regenerates OpenAPI server
task serve      # Starts hotel-api (localhost:8080)
```

## 🧪 Coming Soon
- gRPC between hotel-api and invoice-service
- Background workers with exponential backoff
- Distributed tracing across services
- API key authentication
- Kubernetes manifests + GitOps
- CI pipeline + SBOM + DevSecOps
- OWASP Top 10 vulnerable app (lab environment)

---

This is a **feature-rich** yet **minimalistic** backend showcase designed to impress senior engineers and hiring managers. Built with love by @robinbaeckman ❤️

