# go-hotels

A professional-grade Go backend architecture showcase project for hotel-related operations.

---

## 🎯 Purpose
This project is built to **demonstrate production-ready backend architecture** with a strong focus on:
- Clean and maintainable code structure
- Industry-standard DevSecOps practices
- Full CI/CD automation
- Observability and scalability
- Security-first mindset (OWASP Top 10 mitigations)

---

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

---

## 🧱 Tech Stack

### Language & Framework
- **Go 1.24.x**
- **chi** (HTTP router)
- **oapi-codegen** (OpenAPI 3 code generation)

### Infrastructure
- **Docker** for containerization
- **PostgreSQL 16** as primary database
- **Task** for task orchestration
- **sqlc** for type-safe DB access
- **migrate** for database migrations

### Observability
- **OpenTelemetry SDK** (planned)
- **Grafana Alloy**, **Prometheus**, **Loki**, **Tempo** (planned)

### Developer Experience
- `go fmt` + **goimports** for formatting (CI-enforced)
- **golangci-lint** for static code analysis
- **VS Code + gopls** for rich IDE features
- **Taskfile.yml** for consistent local and CI commands

### DevSecOps
- **gosec** – Static code security scanner
- **govulncheck** – Go vulnerability database checks
- **gitleaks** – Detects secrets in repo
- **Trivy** – Container vulnerability scanning (HIGH + CRITICAL)
- **Caching**:
  - Go modules & build cache
  - Trivy vulnerability database
- **GitHub Container Registry** (GHCR) push after security clearance

---

## ⚙️ CI/CD Pipeline

### 1. **Lint**
- Runs `golangci-lint`
- Runs `go fmt` and `goimports` checks (fails if not formatted)

### 2. **Test**
- Spins up PostgreSQL 16 as a GitHub Actions service
- Installs `migrate` with Postgres driver
- Runs DB migrations and SQL code generation
- Executes full unit test suite

### 3. **Coverage**
- Generates coverage report
- Uploads report as GitHub Actions artifact

### 4. **Build → Scan → Push**
- **Build** Docker image (`docker build`)
- **Local scans**:
  - `gosec`
  - `govulncheck`
  - `gitleaks`
- **Trivy scan** with cached DB
- **Push** image to GHCR (`ghcr.io/robinbaeckman/go-hotels:<sha>`)

---

## 📂 Project Structure
```
go-hotels/
├── cmd/
│   ├── rest/                  # Entrypoint for hotel-api
│   └── invoice/               # Entrypoint for invoice-service
│
├── internal/
│   ├── hotel/                 # Hotel business logic
│   ├── store/                 # Persistence layer
│   ├── transport/rest/        # HTTP handlers
│   ├── infra/                 # Shared infra: DB, NATS, Redis
│   └── platform/              # Observability, logging, security
│
├── migrations/                # Database migrations
├── api/                       # OpenAPI spec & config
├── Taskfile.yml               # All dev and CI tasks
├── go.mod
└── README.md
```

---

## 🚀 Quick Start

### 1. Install Task
```bash
go install github.com/go-task/task/v3/cmd/task@latest
```

### 2. Run Local Dev
```bash
task db-init   # Run migrations + generate SQL code
task run       # Start hotel-api (localhost:8080)
```

### 3. Run Tests
```bash
task test
```

### 4. Build & Scan Locally
```bash
task build-image
task security
```

---

## 🧪 Coming Soon
- gRPC service-to-service communication
- Background workers with exponential backoff
- Distributed tracing across services
- API key authentication
- Kubernetes manifests + GitOps
- OWASP Top 10 security lab

---

Built with ❤️ by [@robinbaeckman](https://github.com/robinbaeckman)

Planned CI/CD Enhancements
	1.	Test Matrix
	•	Run tests against multiple Go versions (e.g., 1.23.x, 1.24.x) and operating systems (Ubuntu, macOS, Windows).
	2.	Build & Test Caching
	•	Improve caching for go build and go test to speed up runs.
	•	Add separate cache for Docker layers using actions/cache or the cache features in docker/build-push-action.
	3.	SBOM Generation
	•	Generate a Software Bill of Materials (SBOM) for both the codebase and Docker images using tools like syft or Trivy’s SBOM mode.
	•	Upload SBOM as a build artifact.
	4.	SAST & DAST
	•	Static Application Security Testing (SAST) — already partially implemented via gosec.
	•	Dynamic Application Security Testing (DAST) — run API security tests against the live service using OWASP ZAP or similar.
	5.	Dependabot Integration
	•	Enable Dependabot for Go modules, Docker images, and GitHub Actions dependencies.
	6.	Release Pipeline
	•	Automatically create GitHub Releases when tags are pushed.
	•	Build and push version-tagged Docker images (:vX.Y.Z).
	•	Upload built binaries as release assets.
	7.	Environment-based Deployments
	•	Separate dev, staging, and production workflows using GitHub Actions environments.
	•	Require manual approval before deploying to production.
	8.	Kubernetes Manifests & GitOps
	•	Generate and validate Kubernetes manifests (using kustomize or Helm).
	•	Optionally push manifests to a GitOps repo (e.g., ArgoCD or Flux).
	9.	Pre-commit Hooks
	•	Run linting, formatting, and security checks locally before commits using the pre-commit framework.
	10.	Advanced Trivy Configuration
	•	Cache Trivy DB (partially implemented).
	•	Scan both Docker images and local filesystem / Go modules for dependency vulnerabilities.
