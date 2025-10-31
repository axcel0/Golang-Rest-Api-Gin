# 🧭 GitHub Copilot Guide for Go Projects

**Go Version**: 1.25.3 (Latest Stable)

## 🎯 Tujuan

Panduan untuk GitHub Copilot agar menulis, memelihara, dan mengembangkan proyek Go secara idiomatik, aman, maintainable, dan sesuai roadmap resmi Go.

---

## 1️⃣ Fundamental Concepts

Copilot harus mematuhi dasar berikut saat menulis atau merefaktor kode:

- **Idiomatic Go**: gunakan pendekatan standar (`defer`, `error`, `context`, package modular).
- **Explicit errors**: tidak boleh silent failure.
- **Context everywhere**: semua I/O (HTTP, DB, goroutine) wajib menerima `context.Context`.
- **Small focused funcs**: satu fungsi = satu tanggung jawab.
- **gofmt/goimports**: selalu jaga format konsisten.

---

## 2️⃣ Go Learning & Development Roadmap (urut logis)

| Tahap | Fokus | Catatan |
|-------|-------|---------|
| **1. Basics** | Setup Go env, `go run`, variabel, konstanta, tipe data, `fmt`, dokumentasi | Gunakan stdlib |
| **2. Control Flow & Data** | Arrays, Slices, Maps, Structs, `if`, `for`, `switch` | "comma-ok idiom", `range` |
| **3. Functions & Pointers** | Variadic, closures, named returns, pointers, GC | Hindari pointer berlebihan |
| **4. Methods & Interfaces** | Receiver pointer/value, interface composition | Minimalist interface |
| **5. Generics** | Gunakan hanya bila ada real benefit | Go 1.18+ |
| **6. Errors** | `errors.New`, `fmt.Errorf("%w")`, unwrap, panic/recover (exceptional only) | Explicit handling |
| **7. Modules & Packages** | `go mod init`, `go mod tidy`, `go doc` | Pisahkan `internal/`/`pkg/` |
| **8. Concurrency** | Goroutine, Channels, `select`, `sync`, `context`, worker pool | "share memory by communicating" |
| **9. Stdlib & Testing** | `io`, `os`, `time`, `log/slog`, `encoding/json`, `testing`, `httptest`, benchmarks | Table-driven tests |
| **10. Ecosystem** | CLI (Cobra), Web (chi/gin), ORM (pgx/sqlc/gorm), logging (zerolog/slog) | Pilih lib sesuai kebijakan |
| **11. Tooling** | `go vet`, `staticcheck`, `golangci-lint`, `govulncheck`, `pprof`, `trace` | Gunakan CI |
| **12. Advanced Topics** | reflection, unsafe, CGO, build tags, plugins | Hanya jika benar-benar dibutuhkan |

---

## 3️⃣ Library Usage & Maintenance Policy

### Prinsip Utama

1. **Prefer maintained library** bila fiturnya sudah tersedia & stabil.
2. Sebelum menambahkan library baru, Copilot wajib cek:
   - ✅ Terakhir commit/rilis **< 12 bulan**.
   - ✅ Tidak ada label **Deprecated** di README/pkg.go.dev.
   - ✅ Kompatibel dengan Go version di `go.mod`.
   - ✅ CI lint/vet/vuln/test lulus.
3. **Selalu gunakan stdlib dulu** jika memadai (jangan reinvent wheel).
4. Setiap library baru harus dicatat alasan pemilihannya di PR.

---

## 🧰 Curated Safe Libraries (2025)

| Kategori | Rekomendasi | Catatan |
|----------|-------------|---------|
| **Router** | `github.com/go-chi/chi/v5` | ringan, idiomatik |
| **Middleware** | `chi/middleware` | requestID, recoverer, timeout |
| **Validation** | `github.com/go-playground/validator/v10` | aktif & maintain |
| **Config** | `github.com/knadh/koanf` atau `github.com/spf13/viper` | pilih sesuai kebutuhan |
| **Logging** | `log/slog` (Go ≥ 1.21) atau `github.com/rs/zerolog` | structured |
| **Tracing** | `go.opentelemetry.io/otel` | OTel standard |
| **Metrics** | `github.com/prometheus/client_golang` | industry standard |
| **DB Driver** | `github.com/jackc/pgx/v5` | prefer native |
| **Query Gen** | `github.com/sqlc-dev/sqlc` | type-safe |
| **Migration** | `github.com/golang-migrate/migrate/v4` | actively maintained |
| **Cache** | `github.com/redis/go-redis/v9` | official client |
| **Messaging** | `github.com/nats-io/nats.go` / `github.com/segmentio/kafka-go` | pure Go |
| **Auth** | `github.com/golang-jwt/jwt/v5` | resmi successor |
| **Testing** | `github.com/stretchr/testify` / `go.uber.org/mock` | assert & mocks |
| **Retry** | `github.com/cenkalti/backoff/v4` | exponential backoff |
| **Rate Limit** | `golang.org/x/time/rate` | official extended lib |
| **DI** | `github.com/google/wire` (compile-time) | optional |

---

## 4️⃣ Project Structure

```
/cmd/<app>         → entrypoint
/internal/...      → business logic (non-exported)
/pkg/...           → reusable packages
/api/...           → handlers / OpenAPI
/configs/...       → config files
/scripts/...       → tools, migrations
```

---

## 5️⃣ Copilot Coding Directives

### General Rules

- ✅ Always handle errors explicitly.
- ✅ Always use `context` for long-running or I/O operations.
- ✅ Use `defer` for resource cleanup (file/db connections).
- ✅ Prefer small, pure functions.
- ✅ Never ignore return errors.
- ✅ Avoid global mutable state.

### Example Comments

```go
// copilot:task
// Goal: add GET /v1/users endpoint with chi, context-aware, structured logging
// Constraints: use slog, handle errors, return JSON, 200/500 properly

// copilot:lib-check
// Need: rate limiter middleware.
// Policy: prefer maintained, non-deprecated lib; check docs/pkg.go.dev
```

---

## 6️⃣ Code Quality & CI

### Mandatory Checks

- ✅ `gofmt`, `goimports`, `golangci-lint`
- ✅ `go vet`, `govulncheck`, `staticcheck`
- ✅ `go test ./... -race -shuffle=on -cover`
- ✅ Lint rule SA1019 → fail if deprecated API used

### Optional Checks

- `pprof`, `trace`, `bench` for optimization
- `goreleaser` for packaging
- Renovate/Dependabot for dependency health

---

## 7️⃣ Testing Practices

- ✅ **Table-driven tests**.
- ✅ `t.Run` subtests.
- ✅ `httptest` for HTTP layer.
- ✅ `testify/assert` or native `testing`.
- ✅ Benchmark critical code (`go test -bench=.`).
- ✅ Use mocks/fakes for DB or external APIs.

---

## 8️⃣ Deployment Best Practices

- ✅ Multi-stage Docker build (distroless)
- ✅ Non-root user
- ✅ HEALTHCHECK endpoint
- ✅ Read-only FS when possible
- ✅ Versioned release via goreleaser

---

## 9️⃣ Copilot Decision Flow

1. **Cek**: bisa pakai stdlib?
2. **Kalau tidak**: gunakan curated lib (lihat tabel di atas).
3. **Kalau fitur belum ada**: riset 1–2 alternatif → pastikan maintained.
4. **Jangan pakai** lib deprecated atau unmaintained.
5. **Pastikan** lint/test/vuln check pass sebelum merge.

---

## 🔐 Security Hygiene

- ✅ Jangan commit secrets. Gunakan env vars atau secret manager.
- ✅ Aktifkan `govulncheck` di CI.
- ✅ Aktifkan Dependabot/Renovate weekly.
- ✅ Gunakan `GOPRIVATE` untuk private modules.
- ✅ Tambahkan SBOM via goreleaser/syft.

---

## ✅ Definition of Done (PR)

- ✅ Build clean + lint/vet/vuln pass
- ✅ Tests & coverage OK (`-race`)
- ✅ Context & error handling lengkap
- ✅ No deprecated/lib unmaintained
- ✅ Docs/README updated
- ✅ Binary/container build pass

---

## ⚙️ Example CI Snippet

```yaml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25.3'
      - run: go mod tidy
      - run: golangci-lint run
      - run: go vet ./...
      - run: govulncheck ./...
      - run: go test ./... -race -shuffle=on -coverprofile=coverage.out
```

---

## 🧠 Reminder to Copilot

1. **Always align** code with the latest Go stable release (1.25.3).
2. **Verify** every imported library is maintained, non-deprecated, and documented.
3. **Use Go idioms** first, library second, custom code last.
4. **Optimize** only after correctness and clarity.

---

**Last Updated**: October 31, 2025  
**Go Version**: 1.25.3  
**Status**: ✅ Active & Maintained
