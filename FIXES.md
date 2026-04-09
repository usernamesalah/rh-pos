# FIXES.md — rh-pos Code Review Fix Plan

This document tracks all fixes identified during code review. 24 issues total (8 CRITICAL, 8 HIGH, 8 MEDIUM).

---

## Phase 0 (P0) — CRITICAL: Fix Before Production

### Step 1: CRIT-7 — Fix Double Password Hashing

**File:** `internal/handler/admin.go:100-105`

**Problem:** `AdminHandler.CreateUser` hashes the password, then passes it to `authService.CreateUser` which hashes it again. The stored value is `bcrypt(bcrypt(plain))`, so login always fails for admin-created users.

**Fix:** Remove the `HashPassword` call from `admin.go:100-105`. The service layer at `auth_service.go:130` already handles hashing. Pass the plain password through to the service.

**Risk:** Low. Existing admin-created users have irrecoverable passwords — they need password resets.

---

### Step 2: CRIT-2 + CRIT-3 — Typed Context Keys and Fail-Fast Tenant Extraction

**New file:** `internal/pkg/ctxkey/ctxkey.go`

**Modified files:**
- `internal/server/router.go:89`
- `internal/repository/transaction_repository.go:44,63,70,98,109,119`
- `internal/repository/product_repository.go:39,50,68,96,123,135`
- `internal/repository/user_repository.go:34,70,85,97,110,114`
- `internal/usecase/transaction_service.go:44`
- `internal/usecase/product_service.go:68,115,143`
- `internal/pkg/storage/minio/client.go:66`

**Problem:** String context key `"tenant_id"` violates `go vet SA1029`. More critically, methods like `transaction_repository.go:63` pass `ctx.Value("tenant_id")` directly into GORM's `Where("tenant_id = ?", ...)` without type assertion. If the key is missing, this produces `WHERE tenant_id = NULL`, returning unexpected data.

**Fix:**
1. Create `internal/pkg/ctxkey/ctxkey.go`:
   - Define `type contextKey string`
   - Export `const TenantID contextKey = "tenant_id"`
   - Export `func TenantIDFromContext(ctx context.Context) (uint, error)` — two-value assertion, returns error if missing or wrong type.
2. In `router.go:89`, replace `context.WithValue(ctx, "tenant_id", ...)` with `context.WithValue(ctx, ctxkey.TenantID, ...)`.
3. In every repository/service/storage method that reads tenant_id, replace `ctx.Value("tenant_id")` with `ctxkey.TenantIDFromContext(ctx)`. Return error immediately if it fails (fail-fast).

**Risk:** Medium. Touches 10+ files. Missed callsites cause compile errors (typed key mismatch is caught at compile time — safe net).

---

### Step 3: CRIT-8 — Safe JWT Claim Type Assertions

**Files:**
- `internal/server/router.go:78`
- `internal/usecase/auth_service.go:88-91`

**Problem:** `claims["user_id"].(float64)` panics on missing or wrong-type claim. Same for `username` and `role` in `auth_service.go`.

**Fix:** Use two-value assertion form everywhere:
```go
userIDFloat, ok := claims["user_id"].(float64)
if !ok {
    return // or return error — do not proceed
}
userID := uint(userIDFloat)
```
Apply to all three claim extractions in `auth_service.go:88-91`.

**Risk:** Low. Purely defensive.

---

### Step 4: CRIT-5 — Atomic Stock Deduction to Prevent Oversell

**File:** `internal/usecase/transaction_service.go:64-92`

**Problem:** Read stock → check → update is TOCTOU. Concurrent requests can both pass the stock check and both deduct, causing oversell.

**Fix:** Replace with atomic conditional UPDATE inside the DB transaction:
```go
result := tx.Model(&entities.Product{}).
    Where("id = ? AND tenant_id = ? AND stock >= ?", item.ProductID, tenantID, item.Quantity).
    Update("stock", gorm.Expr("stock - ?", item.Quantity))
if result.Error != nil {
    return fmt.Errorf("failed to update product stock: %w", result.Error)
}
if result.RowsAffected == 0 {
    return fmt.Errorf("insufficient stock for product %s", product.Name)
}
```
Keep `GetByID` to read `product.HargaJual` for price calculation. Remove the explicit stock check (`if product.Stock < item.Quantity`).

**Dependencies:** Step 2 (typed context key for tenant_id extraction).
**Risk:** Medium. Load test under concurrency before deploying.

---

### Step 5: CRIT-1 — Move HashID Salt to Config

**Files:**
- `internal/pkg/hash/hash.go:13`
- `internal/config/config.go`
- `.env.example`

**Problem:** Salt `"__next_move_to_config__"` is hardcoded and committed. Anyone with repo access can decode all hashed tenant IDs in JWTs.

**Fix:**
1. Add `HashIDSalt string` to `config.Config`. Load from `HASHID_SALT` env var. Require non-empty with minimum 16 chars — fail startup if missing.
2. Add package-level `Init(salt string)` to `hash.go` called from `cmd/main.go`. Remove hardcoded const.
3. Add `HASHID_SALT=` to `.env.example`.

**Risk:** Medium. Changing salt breaks existing JWTs and hashed ID URLs. Initial deployment: set `HASHID_SALT=__next_move_to_config__` to preserve existing tokens. Rotate on a planned maintenance window.

---

### Step 6: CRIT-6 — Harden Admin Password Comparison

**File:** `internal/pkg/middleware/admin_auth.go:13`

**Problem:** Plain string comparison is vulnerable to timing attacks.

**Fix:**
```go
import "crypto/subtle"

usernameMatch := subtle.ConstantTimeCompare([]byte(username), []byte(cfg.Admin.Username))
passwordMatch := subtle.ConstantTimeCompare([]byte(password), []byte(cfg.Admin.Password))
if usernameMatch != 1 || passwordMatch != 1 {
    return false, nil
}
```

**Risk:** Low.

---

### Step 7: CRIT-4 — Float Equality on Money

**File:** `internal/usecase/transaction_service.go:101`

**Problem:** `req.TotalPrice != calculatedTotal` uses exact float64 equality. Floating-point arithmetic can reject legitimate transactions.

**Fix:**
```go
import "math"

const priceEpsilon = 0.01
if math.Abs(req.TotalPrice-calculatedTotal) > priceEpsilon {
    return fmt.Errorf("total price mismatch: provided %.2f, calculated %.2f", req.TotalPrice, calculatedTotal)
}
```

**Risk:** Low.

---

## Phase 1 (P1) — HIGH: Robustness and Security Hardening

### Step 8: HIGH-1 — Use errors.Is for GORM Error Checking

**Files:** `handler/auth_handler.go:102`, `repository/product_repository.go:51,76`, `repository/transaction_repository.go:45`, `repository/user_repository.go:39,55`

**Problem:** `err == gorm.ErrRecordNotFound` fails for wrapped errors, returning 500 instead of 404.

**Fix:** Replace `err == gorm.ErrRecordNotFound` with `errors.Is(err, gorm.ErrRecordNotFound)` everywhere.

---

### Step 9: HIGH-2 — Replace Error String Comparison with Sentinel Errors

**New file:** `internal/domain/errors.go`

**Modified files:** `handler/auth_handler.go:209,212`, `usecase/auth_service.go:153,159`

**Problem:** `err.Error() == "user not found"` is fragile — any typo or message change silently breaks HTTP status routing.

**Fix:**
1. Create `internal/domain/errors.go`:
   ```go
   var ErrUserNotFound = errors.New("user not found")
   var ErrInvalidPassword = errors.New("invalid current password")
   ```
2. Wrap these in `auth_service.go` with `fmt.Errorf("%w", ErrUserNotFound)`.
3. Use `errors.Is(err, domain.ErrUserNotFound)` in handlers.

---

### Step 10: HIGH-3 — Prevent User Existence Leakage in Password Change

**File:** `usecase/auth_service.go:153`

**Problem:** `UpdatePassword` returns a 404 "user not found" for an authenticated user — leaks whether the account was deleted.

**Fix:** Return a generic error for missing users on the authenticated endpoint: `return fmt.Errorf("password update failed")`. Handler maps all errors to HTTP 400.

**Dependencies:** Step 9.

---

### Step 11: HIGH-4 — Report Endpoint Tenant Leakage (Covered by Step 2)

**File:** `repository/transaction_repository.go:98`

This is addressed by Step 2. The `GetReportData` method must use `ctxkey.TenantIDFromContext(ctx)` and fail-fast before executing the raw SQL.

---

### Step 12: HIGH-5 — Remove TenantID Overwrite in UpdateProduct

**File:** `usecase/product_service.go:98,126`

**Problem:** After fetching the product (which already has correct TenantID), lines 98 and 126 overwrite `product.TenantID = &tenantID` from context. If context tenant_id is wrong, this silently moves the product to a different tenant.

**Fix:** Remove `product.TenantID = &tenantID` from both `UpdateProduct` (line 98) and `UpdateStock` (line 126). The fetched product already has the correct TenantID.

**Dependencies:** Step 2.

---

### Step 13: HIGH-6 — Don't Commit Image Key Before Upload Succeeds

**File:** `usecase/product_service.go:179-199`

**Problem:** `GetProductUploadURL` commits the image key to the database before the client uploads anything. If the client never uploads, the product has a broken image reference.

**Fix:** Remove the `UpdateProduct` call from `GetProductUploadURL`. Change the method signature to return `(url string, key string, error)`. Return the key alongside the URL so the client can call a confirmation endpoint (`PUT /api/products/:id` with `{"image": key}`) after a successful upload.

Update `product_handler.go:438` to include the key in the JSON response.

**Risk:** Medium. API contract change — coordinate with frontend.

---

### Step 14: HIGH-7 — Replace fmt.Printf with Structured Logging in MinIO Client

**File:** `pkg/storage/minio/client.go:28,45,48,185`

**Problem:** `fmt.Printf` bypasses `slog` and cannot be level-filtered in production.

**Fix:** Add `logger *slog.Logger` to the `Client` struct and `NewClient` constructor. Replace all `fmt.Printf` calls with `logger.Info`/`logger.Debug` calls with structured fields.

---

### Step 15: HIGH-8 — Restrict CORS Origins

**Files:** `server/router.go:45`, `internal/config/config.go`

**Problem:** `e.Use(echoMiddleware.CORS())` allows all origins (`*`).

**Fix:**
1. Add `AllowedOrigins []string` to `ServerConfig` in `config.go`. Load from `CORS_ALLOWED_ORIGINS` env var (comma-separated). Default: `["http://localhost:3000"]`.
2. Replace in `router.go`:
   ```go
   e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
       AllowOrigins: cfg.Server.AllowedOrigins,
       AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
       AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
   }))
   ```

**Risk:** Medium. Must list all legitimate frontend origins or existing clients break.

---

## Phase 2 (P2) — MEDIUM: Code Quality and Edge Cases

### Step 16: MED-1 — Remove Incorrect tenant_id = 0 in Admin Middleware

**File:** `internal/pkg/middleware/admin_auth.go:18`

**Problem:** `c.Set("tenant_id", uint(0))` sets tenant_id to 0 for admin ops. This is a latent bug if any admin handler calls a tenant-scoped repository method.

**Fix:** Remove line 18 entirely. Admin handlers receive tenant_id explicitly from request bodies. The fail-fast helper from Step 2 will correctly error if admin code paths accidentally read tenant_id from context.

**Dependencies:** Step 2.

---

### Step 17: MED-2 — Remove Redundant GORM Hooks

**File:** `internal/domain/entities/base.go:17-28`

**Problem:** `BeforeCreate`/`BeforeUpdate` hooks duplicate GORM's built-in auto-timestamp handling. Hooks use local timezone instead of UTC.

**Fix:** Delete `BeforeCreate` and `BeforeUpdate` methods (lines 17-28). GORM handles `CreatedAt`/`UpdatedAt` automatically. Verify no entity embeds `Base` before deleting the entire struct if unused.

---

### Step 18: MED-3 — Fix Average Transaction Calculation

**File:** `usecase/report_service.go:47`

**Problem:** `totalRevenue / float64(len(details))` divides by product-row count, not transaction count. Result is wrong.

**Fix:**
1. Add `GetTransactionCount(ctx context.Context, startDate, endDate time.Time) (int64, error)` to `TransactionRepository` interface and implementation.
2. In `report_service.go`:
   ```go
   txCount, err := s.transactionRepo.GetTransactionCount(ctx, startDate, endDate)
   if txCount > 0 {
       averageTransaction = totalRevenue / float64(txCount)
   }
   ```

---

### Step 19: MED-4 — Limit io.ReadAll on MinIO Download

**File:** `internal/pkg/storage/minio/client.go:127`

**Problem:** `io.ReadAll` reads entire object into memory with no size cap — OOM risk.

**Fix:**
```go
const maxImageSize = 10 << 20 // 10 MB
limitedReader := io.LimitReader(reader, maxImageSize+1)
data, err := io.ReadAll(limitedReader)
if len(data) > maxImageSize {
    return nil, fmt.Errorf("object exceeds maximum size of %d bytes", maxImageSize)
}
```

---

### Step 20: MED-5 — Extension Allowlist on Upload URL

**File:** `internal/handler/product_handler.go:425`

**Problem:** `Extension` field accepted without validation — `../../etc/passwd` is valid.

**Fix:** Add allowlist validation after binding the request:
```go
allowedExtensions := map[string]bool{"jpg": true, "jpeg": true, "png": true, "gif": true, "webp": true}
ext := strings.ToLower(strings.TrimPrefix(req.Extension, "."))
if !allowedExtensions[ext] {
    return ErrorResponse(c, http.StatusBadRequest, "Invalid file extension")
}
req.Extension = ext
```

---

### Step 21: MED-6 — Remove Dead Interfaces from User Entity

**File:** `internal/domain/entities/user.go:24-42`

**Problem:** `UserRepository` and `UserUseCase` interfaces defined in the entity file have different signatures from the real interfaces in `domain/interfaces/`. They are dead code and cause confusion.

**Fix:** Delete lines 24-42 from `user.go`.

---

### Step 22: MED-7 — Composite SKU Unique Index Per-Tenant

**Files:** `internal/domain/entities/product.go:12`, new migration `migrations/007_fix_sku_unique_index.sql`

**Problem:** Global unique index on SKU prevents two tenants from using the same SKU.

**Fix:**
1. Change GORM tag:
   ```go
   SKU      string `json:"sku" gorm:"uniqueIndex:idx_tenant_sku;not null"`
   TenantID *uint  `json:"tenant_id" gorm:"uniqueIndex:idx_tenant_sku;index"`
   ```
2. Migration:
   ```sql
   ALTER TABLE products DROP INDEX sku; -- drop global unique index (check exact name first)
   ALTER TABLE products ADD UNIQUE INDEX idx_tenant_sku (tenant_id, sku);
   ```

**Risk:** Medium. Pre-check for cross-tenant SKU conflicts before running:
```sql
SELECT sku, COUNT(DISTINCT tenant_id) FROM products GROUP BY sku HAVING COUNT(DISTINCT tenant_id) > 1;
```

---

### Step 23: MED-8 — Fix Wrong Comment in Config

**File:** `internal/config/config.go:101`

**Problem:** Comment says `// 24 hours default expiry` but value is `time.Hour * 1`.

**Fix:** Change comment to `// 1 hour default expiry`.

---

## Dependency Graph

```
P0 (do first):
  Step 1 ──── independent
  Step 2 ──── foundational → blocks Steps 4, 11, 12, 16
  Step 3 ──── independent
  Step 4 ──── depends on Step 2
  Step 5 ──── independent
  Step 6 ──── independent
  Step 7 ──── independent

P1 (after P0):
  Step 8  ─── independent
  Step 9  ─── independent
  Step 10 ─── depends on Step 9
  Step 11 ─── covered by Step 2
  Step 12 ─── depends on Step 2
  Step 13 ─── independent
  Step 14 ─── independent
  Step 15 ─── independent

P2 (after P1):
  Step 16 ─── depends on Step 2
  Step 17 ─── independent
  Step 18 ─── independent
  Step 19 ─── independent
  Step 20 ─── independent
  Step 21 ─── independent
  Step 22 ─── independent (needs migration pre-check)
  Step 23 ─── independent
```

**Start P0 with Step 2** — it unblocks the most other steps.

---

## New Files to Create

| File | Purpose |
|------|---------|
| `internal/pkg/ctxkey/ctxkey.go` | Typed context key + `TenantIDFromContext` helper |
| `internal/domain/errors.go` | Sentinel errors (`ErrUserNotFound`, `ErrInvalidPassword`) |
| `migrations/007_fix_sku_unique_index.sql` | Drop global SKU index, add composite tenant+SKU index |

## Files to Modify

| File | Steps |
|------|-------|
| `internal/handler/admin.go` | Step 1 |
| `internal/handler/auth_handler.go` | Steps 8, 9, 10 |
| `internal/handler/product_handler.go` | Steps 13, 20 |
| `internal/server/router.go` | Steps 2, 3, 15 |
| `internal/config/config.go` | Steps 5, 15, 23 |
| `internal/pkg/hash/hash.go` | Step 5 |
| `internal/pkg/middleware/admin_auth.go` | Steps 6, 16 |
| `internal/pkg/storage/minio/client.go` | Steps 2, 14, 19 |
| `internal/repository/product_repository.go` | Steps 2, 8 |
| `internal/repository/transaction_repository.go` | Steps 2, 8, 11 |
| `internal/repository/user_repository.go` | Steps 2, 8 |
| `internal/usecase/auth_service.go` | Steps 3, 9, 10 |
| `internal/usecase/product_service.go` | Steps 2, 12, 13 |
| `internal/usecase/report_service.go` | Step 18 |
| `internal/usecase/transaction_service.go` | Steps 4, 7 |
| `internal/domain/entities/base.go` | Step 17 |
| `internal/domain/entities/product.go` | Step 22 |
| `internal/domain/entities/user.go` | Step 21 |
| `internal/domain/interfaces/repositories.go` | Step 18 |
| `.env.example` | Step 5 |

---

## Master Checklist

> **Status: ALL DONE** — `go build ./...` and `go vet ./...` pass clean.

### P0 — CRITICAL
- [x] **Step 1** — Removed double hash in `admin.go` (lines 100-105 deleted)
- [x] **Step 2** — Created `internal/pkg/ctxkey/ctxkey.go`; replaced all `ctx.Value("tenant_id")` with typed `TenantIDFromContext` helper across all repos, services, and MinIO client
- [x] **Step 3** — Two-value assertions in `router.go` and `auth_service.go:ValidateToken`
- [x] **Step 4** — Atomic `UPDATE ... WHERE stock >= ?` + `RowsAffected == 0` check in `transaction_service.go`
- [x] **Step 5** — HashID salt moved to `HASHID_SALT` env var; `hash.Init()` called in `cmd/main.go`; startup fails if missing or < 16 chars
- [x] **Step 6** — `crypto/subtle.ConstantTimeCompare` in `admin_auth.go`; removed `c.Set("tenant_id", uint(0))` side-effect
- [x] **Step 7** — Epsilon (`0.01`) float comparison in `transaction_service.go`

### P1 — HIGH
- [x] **Step 8** — `errors.Is(err, gorm.ErrRecordNotFound)` in all handlers/repos (including `tenant_repository.go` and `transaction_handler.go`)
- [x] **Step 9** — Sentinel errors `ErrUserNotFound`, `ErrInvalidPassword` in `internal/domain/apperrors/errors.go`; used with `errors.Is` in handlers
- [x] **Step 10** — Generic error returned for missing user in `UpdatePassword`; no user-existence leak
- [x] **Step 11** — Report tenant leakage fixed (covered by Step 2 — `GetReportData` now uses typed key + fail-fast)
- [x] **Step 12** — Removed `product.TenantID = &tenantID` overwrite from `UpdateProduct` and `UpdateStock`
- [x] **Step 13** — `GetProductUploadURL` now returns `(url, key, error)`; image key NOT committed to DB until client confirms upload. Interface updated in `services.go`.
- [x] **Step 14** — All `fmt.Printf` removed from `minio/client.go`; `*slog.Logger` injected via `NewClient(config, logger)`; `cmd/main.go` updated
- [x] **Step 15** — CORS restricted to `CORS_ALLOWED_ORIGINS` env var (default `http://localhost:3000`); added to `ServerConfig`

### P2 — MEDIUM
- [x] **Step 16** — Removed `c.Set("tenant_id", uint(0))` from admin middleware (done with Step 6)
- [x] **Step 17** — Removed redundant `BeforeCreate`/`BeforeUpdate` GORM hooks from `entities/base.go`
- [x] **Step 18** — Fixed average transaction: added `GetTransactionCount` to `TransactionRepository` interface + implementation; `report_service.go` now divides by distinct transaction count
- [x] **Step 19** — `io.LimitReader` (10 MB cap) added in `minio/client.go:DownloadBytes`
- [x] **Step 20** — Extension allowlist `[jpg, jpeg, png, gif, webp]` enforced in `product_handler.go:GetUploadURL`
- [x] **Step 21** — Deleted dead `UserRepository`/`UserUseCase` interfaces from `entities/user.go`
- [x] **Step 22** — Composite `uniqueIndex:idx_tenant_sku` on `(tenant_id, sku)` in `entities/product.go`; migration `007_fix_sku_unique_index.sql` created
- [x] **Step 23** — Fixed comment `// 24 hours` → `// 1 hour` in `config.go`

### Additional fixes found during verification
- [x] `tenant_repository.go:41` — `err == gorm.ErrRecordNotFound` → `errors.Is`
- [x] `transaction_handler.go:231` — `err == gorm.ErrRecordNotFound` → `errors.Is`
- [x] `product_service.go:UpdateStock` — restored accidentally dropped `product.Stock = stock` line

### Action Required Before Deploy
1. Add `HASHID_SALT=__next_move_to_config__` to your `.env` (preserves existing tokens)
2. Add `CORS_ALLOWED_ORIGINS=https://your-frontend.com` to your `.env`
3. Run `migrations/007_fix_sku_unique_index.sql` after pre-checking for cross-tenant SKU conflicts:
   ```sql
   SELECT sku, COUNT(DISTINCT tenant_id) FROM products GROUP BY sku HAVING COUNT(DISTINCT tenant_id) > 1;
   ```
