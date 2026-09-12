<p align="center">
  <img src=".dev/assets/logo.png" width="250">
</p>


<div align="center">

**PrismGo - Write Go Like Laravel**

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Module](https://img.shields.io/badge/module-github.com%2Fprismgo%2Fframework-blue)](https://github.com/prismgo/framework)
[![Coverage](https://codecov.io/gh/prismgo/framework/branch/main/graph/badge.svg)](https://codecov.io/gh/prismgo/framework)
[![Latest Version](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fproxy.golang.org%2Fgithub.com%2Fprismgo%2Fframework%2F%40latest&query=%24.Version&label=version)](https://pkg.go.dev/github.com/prismgo/framework?tab=versions)
[![License](https://img.shields.io/badge/license-MIT-green)](./LICENSE)

[简体中文](README.md) | English

</div>

---

## Introduction

PrismGo Demo is the framework's **runnable example and local integration project**. A local `replace` in `go.mod` points directly to the sibling `framework/` source.

The Demo project has four main responsibilities:

- **Demonstrate usage** through real application code for routes, commands, models, repositories, transactions, queued jobs, scheduled tasks, and more.
- **Maintain the coverage catalog**, mapping Framework version → feature → documentation section → example → test as a queryable progress map.
- **Provide isolated verification** with a test Application that defaults to SQLite, memory/file drivers, and temporary directories without reading the development `.env`.
- **Support local integration** through `./dev`, which initializes the workspace and manages MySQL, PostgreSQL, SQL Server, Redis, and RabbitMQ as needed.

| Path | Purpose |
|---|---|
| `app/demo/catalog/` | Catalog data, feature progress, and documentation mappings |
| `app/demo/` | Runnable feature examples and scenario tests |
| `app/models/`, `app/repositories/`, `app/services/` | Model, repository, and business-service examples for the order domain |
| `app/demo/testing/` | Test Application isolated from the local development environment |
| `database/migrations/` | Demo database migrations |
| `.dev/`, `dev` | Workspace initialization, local dependencies, and integration tools |

## Catalog Progress Map

The catalog currently covers **29 modules and 1432 entries**: 158 implemented, 1271 planned, and 3 manually verified. The table below is a snapshot taken when this README was updated. `app/demo/catalog/` is the single source of truth; use `demo:list` for live progress.

| Module | Coverage | Progress | Remaining | Status |
|---|---|---:|---:|---|
| `cache` | Cache stores, tags, and locks | 57/57 | 0 | Implemented |
| `commands` | Application commands and options | 1/101 | 100 | In progress |
| `config` | Environment and application configuration | 0/62 | 62 | Planned |
| `console` | Artisan-style commands and terminal IO | 0/70 | 70 | Planned |
| `container` | Dependency binding and resolution | 0/47 | 47 | Planned |
| `cookie` | Cookie values and queued writes | 0/50 | 50 | Planned |
| `database` | Connections, models, migrations, and indexes | 0/62 | 62 | Planned |
| `encryption` | Application data encryption | 0/1 | 1 | Planned |
| `event` | Synchronous, asynchronous, and queued events | 0/57 | 57 | Planned |
| `exception` | Exception reporting and rendering | 0/77 | 77 | Planned |
| `facade` | Framework facade access | 0/1 | 1 | Planned |
| `filesystem` | Local, public, and cloud filesystems | 1/67 | 66 | In progress |
| `horizon` | Queue monitoring and worker management | 1/1 | 0 | Implemented |
| `http-server` | HTTP server startup and lifecycle | 0/29 | 29 | Planned |
| `installation` | Application installation workflow | 0/1 | — | Manual |
| `lens` | Lens development workflow | 0/1 | — | Manual |
| `lifecycle` | Application bootstrap and shutdown | 0/57 | 57 | Planned |
| `logger` | Multi-channel application logging | 0/40 | 40 | Planned |
| `queue` | Queues, jobs, and workers | 59/59 | 0 | Implemented |
| `ratelimit` | Request and action rate limiting | 0/74 | 74 | Planned |
| `redis` | Redis connections and operations | 0/93 | 93 | Planned |
| `route` | HTTP route registration | 0/87 | 87 | Planned |
| `schema` | Database schema and migration builder | 0/143 | 143 | Planned |
| `service-provider` | Service provider registration and lifecycle | 0/42 | 42 | Planned |
| `session` | Server-side session storage | 0/50 | 50 | Planned |
| `starter` | Generated application starter | 0/1 | — | Manual |
| `support` | General framework helpers | 0/1 | 1 | Planned |
| `timer` | Scheduled task definitions | 0/62 | 62 | Planned |
| `translation` | Application translation and pluralization | 39/39 | 0 | Implemented |

From the workspace root, inspect the live map and entry details with:

```bash
go run ./demo demo:list
go run ./demo demo:show queue
go run ./demo demo:show queue --status=planned
go run ./demo demo:show queue --level=integration
go run ./demo demo:list --json
go run ./demo demo:cache list
go run ./demo demo:cache get --json
```

Status meanings: `Implemented` means every entry in the module is complete; `In progress` means some entries are complete; `Planned` means no entries are implemented yet; `Manual` means an external workflow such as the installer or Lens performs verification, so it is not counted as remaining work.

## Local Development Workspace and the `dev` Script

PrismGo local development uses three sibling repositories. **`demo/`, `framework/`, and `docs/` live directly under the same workspace root and are not nested inside one another.** Each directory is an independent Git repository; the workspace root itself is not a Git repository.

### Installation and Initialization

Initialization requires Go 1.25+, Git, GitHub SSH access, and `flock`. Docker Engine and Compose v2 are also required when running databases, Redis, or RabbitMQ.

Create an empty workspace directory and clone only the Demo repository. Its directory name must be `demo`:

```bash
mkdir workspace
cd workspace
git clone git@github.com:prismgo/demo.git demo
./demo/dev init
```

`init` automatically completes the remaining setup:

1. Clone missing `framework/` and `docs/` repositories alongside `demo/`.
2. Create `demo/.env` when absent and generate a random local key when `APP_KEY` is still empty; existing non-empty configuration is preserved.
3. Create or update `go.work` with `demo/` and `framework/`.
4. Create workspace Agent-instruction and project-skill links.
5. Download the Framework's Go dependencies.

### Three-Repository Directory Layout

After `./demo/dev init` completes, the directory layout is:

```text
workspace/                           # Workspace root (any name; not a Git repository)
├── demo/                            # Repository 1: examples, Catalog, integration tools
│   ├── .git/
│   ├── dev                          # Workspace management entry point
│   └── go.mod                       # replace framework => ../framework
├── framework/                       # Repository 2: framework source and tests
│   └── .git/
├── docs/                            # Repository 3: Chinese and English user docs
│   └── .git/
├── go.work                          # Loads both demo/ and framework/
├── AGENTS.md -> demo/AGENTS.md
└── CLAUDE.md -> demo/AGENTS.md
```

`demo/go.mod` uses the local framework source through `../framework`. The Catalog then maps Demo examples and tests to documentation sections in `docs/`. Changes and commits must be made separately in the corresponding repository.

| Repository | What belongs here | What does not belong here |
|---|---|---|
| `demo/` | Demos, the Catalog, scenario tests, local dependencies, and development tools | Real implementations of framework public APIs |
| `framework/` | Framework implementations, public APIs, components, and framework tests | Demo-specific business examples |
| `docs/` | Chinese and English user guides and API documentation | Framework or Demo implementation code |

After initialization, inspect the environment, view the Catalog, and start the Demo HTTP server:

```bash
./demo/dev doctor              # Requires Docker; checks dependencies and Compose config
go run ./demo demo:list        # Show feature coverage progress
go run ./demo serve            # Start the Demo HTTP server
```

If no external services are needed, skip `doctor` and `dev up` and run the Catalog, Demo, or default tests directly.

### Common `dev` Commands

| Command | Usage |
|---|---|
| `./demo/dev help` | Show commands and available service names |
| `./demo/dev up [service...]` | Start all or selected dependencies and wait for health checks |
| `./demo/dev status` | Show service health, local ports, and connection details |
| `./demo/dev logs [service]` | Follow logs for all services or one service |
| `./demo/dev env` | Regenerate and print `demo/.dev/runtime/test.env` |
| `./demo/dev test [go-test-args]` | Start dependencies and run selected integration tests in `framework/` |
| `./demo/dev test-horizon` | Start Redis and RabbitMQ and run the real Horizon dual-queue consumption test |
| `./demo/dev down` | Stop services while retaining local data |
| `./demo/dev reset [--yes]` | Delete all local test data and rebuild the environment; use with care |

Available service names are `mysql`, `postgres`, `sqlserver`, `redis`, `rabbitmq`, and `sqlite`. Usually, start only the dependencies needed by the current task:

```bash
./demo/dev up mysql redis
./demo/dev status
./demo/dev logs redis
```

Load the generated test connection variables, or run framework integration tests through the unified entry point:

```bash
./demo/dev env
source demo/.dev/runtime/test.env

./demo/dev test ./queue/...
./demo/dev test-horizon
./demo/dev down
```

The Demo's default tests use isolated drivers and do not require Docker:

```bash
cd demo
GOWORK=off go test . ./app/... ./bootstrap/... ./config/... ./database/... ./routes/...
```

See the [local test environment guide](.dev/docs/local-test-environment.md) for service addresses, environment variables, security notes, and troubleshooting.
