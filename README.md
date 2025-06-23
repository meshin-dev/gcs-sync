# GCS Sync

**gcs-sync** is a lightweight daemon that keeps one or more local folders in-sync with Google Cloud Storage (GCS).  
It supports one-way or two-way replication, ignore patterns, debounce logic, and runs either as a standalone binary or in a minimal Docker container.

---

## Features

| Feature | Details |
|---------|---------|
| **Bi-directional sync** | `local_to_remote`, `remote_to_local`, `full` (two-way + delete) |
| **Recursive watch** | Any newly-created sub-directory is picked up automatically |
| **Debounce** | Burst file events collapse into a single `gsutil rsync`, default 3 s |
| **Ignore list** | Glob patterns (`**`, `*`, `?`) compiled to regex for both watcher **and** `gsutil -x` |
| **Pluggable logging** | [logrus] levels (`trace`-`error`) via `--log-level` |
| **Cobra CLI** | Simple flag handling (`--config`, `--log-level`) |
| **Uber Fx DI** | Clean lifecycle & graceful shutdown |
| **Container-ready** | ≤ 150 MB image based on `google/cloud-sdk:alpine` |

---

## Quick start

```bash
git clone https://github.com/meshin-dev/gcs-sync && cd gcs-sync

# build & run natively
go run ./main.go --config ./settings/config.yaml --log-level=debug
````

### Using Docker Compose

```bash
docker compose up -d          # builds image & starts the daemon
docker compose logs -f        # tail logs
```

> **Note**
> The Compose file mounts `./settings` **read-only** so your service-account JSON and `config.yaml` never land inside the image.

---

## Configuration

Create `settings/config.yaml` (name/path is configurable via `--config`).

```yaml
sync:
  - src: ~/Projects/book       # local folder (tilde expanded)
    dst: gs://my-bucket/book   # GCS bucket or path
    directions: [local_to_remote]   # or remote_to_local, full
    ignore:                    # glob patterns, relative to src
      - "**/*.tmp"
      - cache/**
    enabled: true
```

### Sync directions

| Value             | Effect                                           |
| ----------------- | ------------------------------------------------ |
| `local_to_remote` | One-way `LOCAL ➜ GCS`                            |
| `remote_to_local` | One-way `GCS ➜ LOCAL`                            |
| `full`            | Two-way, with **`-d`** (delete) on the push step |

Add multiple rules to sync several folders concurrently.

---

## CLI

```text
gcs-sync --help
Bi-directional Google Cloud Storage synchronizer

Usage:
  gcs-sync [flags]

Flags:
  -c, --config     Path to YAML configuration (default "/app/settings/config.yaml")
  -l, --log-level  Log level: trace|debug|info|warn|error (default "info")
  -h, --help       Print help
```

---

## Building from source

```bash
go version          # requires Go 1.22+
go mod tidy
go build -o gcs-sync ./main.go
```

---

## Container image

The provided **multi-arch** Dockerfile produces a fully-static AMD64 binary, then copies it into `google/cloud-sdk:alpine`.

```bash
# local build
docker build -t gcs-sync:dev .

# run
docker run --rm \
  -v "$(pwd)/settings":/app/settings:ro \
  -e GOOGLE_APPLICATION_CREDENTIALS=/app/settings/service-account.json \
  gcs-sync:dev --log-level=debug
```

---

## Advanced usage

* **Multiple instances on the same host**
  Duplicate the Compose file (`docker-compose.dev.yaml`), point it to a different `settings/` folder, and start with
  `docker compose -f docker-compose.dev.yaml up -d`.
  
  **Or simply copy service inside existing yaml to a new service which points to different settings.**

* **Custom debounce window**
  Change `debounceWindow` constant in `internal/watcher/watcher.go`, re-compile.

* **Fine-grained gsutil flags**
  Edit `internal/gsutil/gsutil.go` to tweak parallelism or add canned ACLs.

---

## Contributing

1. Fork the repo & create your branch (`git checkout -b feature/foo`).
2. Run `go test ./...` (coming soon) and `golangci-lint run`.
3. Commit and open a PR. All contributions & bug reports are welcome!

---

## License

Distributed under the **MIT License**. See [`LICENSE`](LICENSE) for details.
