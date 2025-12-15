# UPDATE.md

## Purpose
This file captures code review notes, highlights, perks, and recommendations for the project.

---

## High-level highlights
- Clear request/response DTO separation and mapping helpers make the API layer easier to reason about.
- Pagination metadata is client-friendly (`HasPrev`, `IsFirstPage`, `IsLastPage`, etc.).
- Cache invalidation by tags in the service layer is a strong architectural choice and aligns with a hexagonal approach.

---

## Key findings (priority-ordered)

### P0 — Consolidate duplicated pagination defaults + validation
**Observed**
- Pagination defaults exist in more than one place (e.g., task package and `pkg/utils`).
- Both handler-level and package-level pagination validation exist.

**Why it matters**
- Defaults and validation rules can silently diverge, causing inconsistent behavior across endpoints.

**Recommendation**
- Use a single source of truth for:
  - `DefaultLimit`, `MaxLimit`, `DefaultOrder`
  - query validation (either handler-only or a shared validation helper)

**Acceptance criteria**
- A single set of constants is used everywhere.
- Only one validation path exists for each pagination style.

---

### P0 — Unify ETag generation and formatting
**Observed**
- ETag logic exists in multiple packages with different formats (weak vs strong tags) and different hashing strategies.

**Why it matters**
- Inconsistent ETags can break client caching expectations and increase complexity when debugging cache behavior.

**Recommendation**
- Centralize ETag creation behind one implementation (e.g., a single ETag generator utility).
- Decide and document whether list/item endpoints use weak ETags (`W/"..."`) or strong (`"..."`).

**Acceptance criteria**
- One ETag strategy is used consistently across endpoints.
- ETag inputs include all fields that affect the response representation.

---

### P1 — Fix `Last-Modified` semantics
**Observed**
- `Last-Modified` appears to be set using the current server time.

**Why it matters**
- `Last-Modified` should reflect the resource modification timestamp, not response generation time, otherwise conditional requests lose value.

**Recommendation**
- For single tasks: `Last-Modified = task.UpdatedAt`.
- For lists: `Last-Modified = max(tasks[].UpdatedAt)` (or omit if list semantics are unclear).

**Acceptance criteria**
- `Last-Modified` correlates to actual data changes.

---

### P1 — Avoid mixing cursor and page pagination metadata
**Observed**
- Cursor responses include page-like fields (current/total pages) even when not meaningful.

**Why it matters**
- API consumers can make incorrect assumptions and UI logic becomes harder.

**Recommendation**
- Split metadata into two structs:
  - `CursorPaginationMeta` (cursor, hasMore, hasPrev, isFirstPage, isLastPage, limit, order)
  - `PagePaginationMeta` (page, totalPages, hasMore, hasPrev, limit, order, totalCount)
- Or make non-applicable fields `omitempty` and only populate them in the relevant endpoints.

**Acceptance criteria**
- Cursor endpoints do not imply page counts unless truly supported.

---

### P2 — Improve error handling (typed errors vs string matching)
**Observed**
- Some error mapping depends on substring checks like "not found" / "already exists".

**Why it matters**
- String matching is brittle and can break during refactors or localization of error messages.

**Recommendation**
- Introduce sentinel errors (e.g., `ErrNotFound`, `ErrAlreadyExists`, `ErrValidation`) or typed errors.
- Map those to HTTP status codes in a single place (handler middleware/helper).

**Acceptance criteria**
- Handlers do not rely on string contains checks to set status codes.

---

## Perks you already get (if you keep leaning into this design)
- Easier adapter swapping: caching/logging/DB layers are already trending toward replaceable components.
- Better testability: DTO mapping and service logic are naturally unit-testable.
- Clearer client integration: richer pagination meta reduces guesswork for frontends.

---

## Quick wins (low risk, good payoff)
1. Fix comment drift: if `CacheTTL` is 30 seconds, say 30 seconds everywhere.
2. Ensure cache keys vary by query parameters (page/limit/order/cursor).
3. Replace `Last-Modified = now()` with data-based timestamps.
4. Pick one ETag generator and reuse it consistently.
5. Reduce duplicated pagination validation.

---

## Medium-term improvements
- Context propagation: prefer request context (or derived timeout context) for DB/cache/logging.
- Standardize response envelope usage (when to include `Data: nil`, how to format validation errors).
- Add tests:
  - pagination validation tests
  - ETag determinism tests
  - cache invalidation tests (tag-based)

---

## Notes
This document reflects a snapshot review of the currently inspected files and may not include all modules.
