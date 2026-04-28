# Feature Updates Completion Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bring the backend fully in line with `docs/superpowers/specs/2026-04-13-feature-updates-design.md`, based on the repo's current state rather than re-implementing already completed work.

**Planning note:** Most of the spec is already implemented in the current branch. This plan focuses on verification, the remaining Feature 5 gap in `UpdateProduct`, and API behavior/documentation alignment.

---

## Current Status Audit

| Spec requirement | Status | Evidence |
|---|---|---|
| SKU nullable | Implemented | `internal/domain/entities/product.go`, `migrations/015_sku_nullable.sql`, `internal/handler/product_handler.go` |
| Tenant `terms_of_service` field | Implemented | `internal/domain/entities/tenant.go`, `migrations/016_add_tenant_terms_of_service.sql`, `internal/handler/admin.go`, `internal/handler/auth_handler.go` |
| Cashier can update discount campaigns | Implemented | `internal/server/router.go` |
| Admin-only product delete endpoint | Implemented | `internal/domain/interfaces/services.go`, `internal/usecase/product_service.go`, `internal/handler/product_handler.go`, `internal/server/router.go` |
| Nullable prices on create | Implemented | `internal/domain/entities/product.go`, `migrations/017_price_fields_nullable.sql`, `internal/usecase/product_service.go` |
| Copy-on-null for `CreateProduct` | Implemented | `internal/usecase/product_service.go`, `internal/usecase/product_service_test.go` |
| Copy-on-null for `UpdateProduct` | Not implemented | `internal/usecase/product_service.go` currently updates fields independently |
| Validation error status for invalid price input | Not aligned | `internal/handler/product_handler.go` currently returns `500` from `CreateProduct`/`UpdateProduct` service validation failures |
| Frontend/API doc reflects actual final behavior | Not aligned | `docs/2026-04-13-frontend-api-changes.md` still says both missing prices returns `500` |

---

## File Map

| File | Change |
|---|---|
| `internal/domain/apperrors/errors.go` | MODIFY - add one sentinel error for invalid product price input |
| `internal/usecase/product_service.go` | MODIFY - apply Feature 5 rules to `UpdateProduct` and return wrapped validation error |
| `internal/usecase/product_service_test.go` | MODIFY - add tests for update-time copy-on-null and both-nil validation |
| `internal/handler/product_handler.go` | MODIFY - map product validation failures to `400 Bad Request` |
| `docs/2026-04-13-frontend-api-changes.md` | MODIFY - document final status code and update semantics |
| `docs/superpowers/plans/2026-04-13-feature-updates.md` | MODIFY - this plan |

---

## Task 1: Verify The Already-Implemented Spec Items

**Goal:** Confirm the repo baseline before changing the remaining gaps.

**Files:**
- No code changes expected

- [ ] Run the focused usecase tests:

```bash
go test ./internal/usecase/... -v
```

Expected: existing product tests pass, including SKU-null, delete, and create-time copy-on-null cases.

- [ ] Run a full build:

```bash
go build ./...
```

Expected: clean build.

- [ ] Confirm the current spec coverage manually in code:

```text
internal/domain/entities/product.go
internal/domain/entities/tenant.go
internal/handler/admin.go
internal/handler/auth_handler.go
internal/handler/product_handler.go
internal/server/router.go
migrations/015_sku_nullable.sql
migrations/016_add_tenant_terms_of_service.sql
migrations/017_price_fields_nullable.sql
```

Completion criteria: only Feature 5 update semantics and validation-response handling remain open.

---

## Task 2: Finish Feature 5 For `UpdateProduct`

**Goal:** Apply the same nullable-price rules from the spec to product updates, not just creates.

**Files:**
- Modify: `internal/domain/apperrors/errors.go`
- Modify: `internal/usecase/product_service.go`
- Modify: `internal/usecase/product_service_test.go`

- [ ] Add one sentinel error for invalid product price input in `internal/domain/apperrors/errors.go`.

Recommended shape:

```go
var (
    ErrUserNotFound       = errors.New("user not found")
    ErrInvalidPassword    = errors.New("invalid current password")
    ErrTenantNotInContext = errors.New("tenant_id not found in context")
    ErrInvalidProductPrice = errors.New("invalid product price input")
)
```

Use one error only. Do not introduce a custom error type unless needed.

- [ ] In `internal/usecase/product_service.go`, update `UpdateProduct` so price normalization happens after the existing product is loaded and after request fields are applied.

Required behavior from the spec:

```text
If only harga_jual is provided in the update, set harga_modal to the same value.
If only harga_modal is provided in the update, set harga_jual to the same value.
If both are provided, use both values as-is.
If the resulting product still has both prices nil, return a validation error.
```

Implementation guidance:

```text
1. Load the existing product.
2. Track whether `harga_modal` and `harga_jual` were present in the update map.
3. Apply the requested field updates.
4. Normalize prices based on which price fields were provided.
5. If both resulting fields are nil, return `fmt.Errorf("...: %w", apperrors.ErrInvalidProductPrice)`.
6. Save through `productRepo.Update`.
```

Important nuance:

```text
Omitted price fields should not be treated as an instruction to clear them.
Only fields present in the `updates` map count as client-provided update inputs.
```

- [ ] Add or extend tests in `internal/usecase/product_service_test.go`.

Minimum required test cases:

```text
TestUpdateProduct_OnlyHargaJual_CopiesHargaModal
TestUpdateProduct_OnlyHargaModal_CopiesHargaJual
TestUpdateProduct_BothPricesSet_UsesAsIs
TestUpdateProduct_BothPricesNil_ReturnsError
```

Recommended test setup pattern:

```text
1. Seed an existing product in the mock repo.
2. Call `UpdateProduct` with a partial updates map.
3. Assert the returned product and stored product have the normalized values.
4. Assert the both-nil case returns `apperrors.ErrInvalidProductPrice` via `errors.Is`.
```

- [ ] Run the focused tests:

```bash
go test ./internal/usecase/... -run TestUpdateProduct -v
```

Expected: all new update-price tests pass.

---

## Task 3: Return `400` For Product Validation Errors

**Goal:** Product input validation failures should not surface as internal server errors.

**Files:**
- Modify: `internal/handler/product_handler.go`

- [ ] Update `CreateProduct` in `internal/handler/product_handler.go` to map `apperrors.ErrInvalidProductPrice` to `400 Bad Request`.

Recommended behavior:

```text
If `CreateProduct` returns `ErrInvalidProductPrice`, respond with 400 and the spec message.
Keep unexpected service failures as 500.
```

- [ ] Update `UpdateProduct` in `internal/handler/product_handler.go` to map `apperrors.ErrInvalidProductPrice` to `400 Bad Request`.

Recommended response message:

```text
at least one of harga_jual or harga_modal must be provided
```

If the codebase convention prefers generic client-facing text, keep the handler message generic but still return `400`.

- [ ] Run the build after the handler change:

```bash
go build ./...
```

Expected: clean build.

Optional if convenient:

```text
Add a small handler test for create/update validation mapping.
If no handler test harness exists, rely on usecase tests plus a manual API smoke check.
```

---

## Task 4: Align Frontend/API Documentation

**Goal:** Make the consumer-facing doc match the final backend behavior.

**Files:**
- Modify: `docs/2026-04-13-frontend-api-changes.md`

- [ ] Update the price-field section so it no longer says both missing prices returns `500`.

Change the text to `400` and clarify update semantics:

```text
- On create, at least one of `harga_jual` or `harga_modal` must be provided.
- On update, if only one of the two price fields is supplied, the other is synchronized to the same value.
- Requests that would leave both prices unset return `400 Bad Request`.
```

- [ ] Review the rest of the document for consistency with the implemented routes and nullable fields.

Specific checks:

```text
- `sku` optional in create/update
- `terms_of_service` present in tenant APIs
- cashier can `PUT /api/discount-campaigns/:id`
- admin-only `DELETE /api/products/:id`
```

---

## Task 5: Final Verification

**Goal:** Verify the remaining spec work is complete end-to-end.

**Files:**
- No code changes expected

- [ ] Run targeted tests:

```bash
go test ./internal/usecase/... -v
```

- [ ] Run the full test suite:

```bash
go test ./...
```

- [ ] Run a full build:

```bash
go build -o bin/rh-pos cmd/main.go
```

- [ ] If local environment is ready, perform one manual API smoke check for each changed behavior:

```text
POST /api/products with only harga_jual -> success, harga_modal copied
PUT /api/products/:id with only harga_modal -> success, harga_jual copied
POST /api/products with both prices omitted -> 400
PUT /api/products/:id on a both-nil payload/result -> 400
DELETE /api/products/:id as admin -> success
PUT /api/discount-campaigns/:id as cashier -> success
GET /api/my-tenant -> includes terms_of_service
```

Completion criteria:

```text
All five spec features are implemented.
Feature 5 behaves the same on create and update.
Client-visible validation errors return 400 instead of 500.
Docs match the shipped behavior.
```

---

## Spec Coverage Check

| Spec requirement | Plan task |
|---|---|
| SKU nullable | Task 1 verification |
| Migration 015 | Task 1 verification |
| Tenant terms_of_service field | Task 1 verification |
| Migration 016 | Task 1 verification |
| Cashier can update discount campaign | Task 1 verification |
| Admin-only product delete endpoint | Task 1 verification, Task 5 smoke check |
| Migration 017 | Task 1 verification |
| Nullable price fields on create | Task 1 verification |
| Copy-on-null for update | Task 2 |
| Both prices nil returns validation error | Task 2, Task 3, Task 5 |
| Frontend/API consumer guidance | Task 4 |

---

## Out Of Scope

- No schema changes beyond the three migrations already present
- No router or RBAC changes beyond verifying existing behavior
- No retroactive data cleanup for legacy rows with null prices unless a new requirement appears
- No new role or permission model beyond `admin` and `cashier`
