# 🧭 GitHub Copilot Guide for Go Projects

**Go Version**: 1.25.3 (Latest Stable)  
**Last Updated**: October 31, 2025

## 🎯 Tujuan
Panduan untuk GitHub Copilot agar menulis, memelihara, dan mengembangkan proyek Go secara idiomatik, aman, maintainable, dan mengikuti roadmap Go terbaru.

---

## 1️⃣ Fundamental
- **Idiomatic Go** (`defer`, explicit `error`, `context`, packages).
- **Explicit error handling**; jangan silent.
- **Gunakan `context.Context`** di semua I/O & goroutine.
- **Fungsi kecil & fokus**; nama jelas.
- **Format konsisten**: `gofmt`/`goimports`.

---

## 2️⃣ Roadmap (ringkas)
1. **Basics** → variabel, tipe data, `fmt`, tools `go`.
2. **Data & Control Flow** → arrays, slices, maps, structs, `if/for/switch`, comma-ok.
3. **Functions & Pointers** → variadic, closures, named returns, GC aware.
4. **Methods & Interfaces** → small interfaces, value vs pointer receiver.
5. **Generics** → gunakan bila ada manfaat nyata.
6. **Errors** → `%w`, `errors.Is/As`, panic/recover hanya exceptional.
7. **Modules & Packages** → `go mod`, layout `/cmd`, `/internal`, `/pkg`.
8. **Concurrency** → goroutine, channels, `select`, `sync`, `context`.
9. **Stdlib & Testing** → `testing`, table-driven, `httptest`, benchmarks.
10. **Ecosystem** → CLI, HTTP router, DB, logging/OTel.
11. **Tooling** → `vet`, `staticcheck`, `golangci-lint`, `govulncheck`, `pprof`.
12. **Advanced** → reflection, unsafe, CGO, build tags (only if needed).

---

## 3️⃣ Library Usage Policy

### ⚠️ CRITICAL: Deprecation Prevention
**WAJIB melakukan double-check sebelum menggunakan/menambahkan library:**

1. **Cek dokumentasi resmi** di pkg.go.dev
2. **Verifikasi di GitHub**: last commit < 12 bulan
3. **Cek deprecation notice** di README/godoc
4. **Pastikan compatible** dengan Go 1.25.3
5. **Jangan gunakan library deprecated** (SA1019 harus bersih)

### Curated Safe Libraries (Verified 2025)
**Router**: `github.com/gin-gonic/gin` v1.10+  
**Validation**: `github.com/go-playground/validator/v10`  
**Config**: `github.com/spf13/viper` v1.19+  
**Logging**: `log/slog` (stdlib Go ≥1.21) - PREFER THIS  
**DB Driver**: `github.com/lib/pq` (PostgreSQL)  
**ORM**: `gorm.io/gorm` v1.25+  
**Query Gen**: `github.com/sqlc-dev/sqlc`  
**Migration**: `github.com/golang-migrate/migrate/v4`  
**Auth**: `github.com/golang-jwt/jwt/v5`  
**Testing**: `github.com/stretchr/testify`  
**Mock**: `go.uber.org/mock`  
**Rate Limit**: `golang.org/x/time/rate`  
**Swagger**: `github.com/swaggo/swag`  
**Crypto**: `golang.org/x/crypto/bcrypt`

---

## 4️⃣ Copilot Directives

### WAJIB Verification Flow
```go
// copilot:lib-check
// 1. Search pkg.go.dev for library
// 2. Verify GitHub last commit < 12 months
// 3. Check godoc for deprecation warnings
// 4. Ensure Go 1.25.3 compatible
// 5. Run staticcheck for SA1019
```

---

## 5️⃣ Quality Gates (CI)
```bash
golangci-lint run  # SA1019 check
go vet ./...
govulncheck ./...
staticcheck ./...
go test ./... -race -shuffle=on -cover
```

---

## 6️⃣ Definition of Done (PR)
- ✅ **NO DEPRECATED CODE** (SA1019 clean)
- ✅ **All imports verified** at pkg.go.dev
- ✅ Build & tests pass with `-race`
- ✅ Coverage ≥ 70%
- ✅ Docs updated

---

**Reminder**: ALWAYS verify at pkg.go.dev before importing!

