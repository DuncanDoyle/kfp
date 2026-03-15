# Phase 1.3 — Complete Record

**Date:** 2026-03-15
**Status:** Complete

## Summary

Phase 1.3 was the third patch cycle for Phase 1 (Envoy Config Viewer). It resolved documentation debt, a deep-copy correctness bug, and missing renderer test coverage introduced during Phases 1.1 and 1.2.

---

## Issues Resolved

### #7 — Design doc stale (P1)

**Type:** Design gap — Phases 1.1/1.2 expanded the model beyond what the design doc described.

**Changes:**
- `docs/plans/2026-03-08-phase-1-envoy-viewer-design.md` — updated Data Model section:
  - Added `PathSeparatedPrefix` and `QueryParams` to `RouteMatch`
  - Added `Regex bool` to `HeaderMatch`
  - Added `QueryParamMatch` type
  - Added `RouteRewrite` type
  - Added `Rewrite *RouteRewrite` to `Route`
- Updated Testing Strategy section to include: 01_x matcher scenarios, 02_5 URLRewrite, 02_6 CORS, 02_8 RateLimit, URLRewrite rendering tests, new match type rendering tests.

**Commit:** see phase commit

---

### #8 — cloneRouteConfig deep-copy bug (P1)

**Type:** Implementation bug — `TypedPerFilterConfig` and `Metadata` maps were not deep-copied, leaving clones sharing the same underlying map as the original.

**Root cause:** `cloneRoute := r` copies the struct by value but maps are reference types — both the original and the clone pointed to the same underlying map. Mutations in Phase 2 (filter expansion) would corrupt other HCMs sharing the same route_config_name.

**Fix:** Added shallow map copy (new map, same value references) for both `TypedPerFilterConfig` and `Metadata` in `cloneRouteConfig`. Used `maps.Copy` (Go 1.21+ stdlib). Values are opaque JSON (`any`) and are only read by the renderer, so shallow value copy is safe.

**New tests:** `internal/parser/clone_test.go`
- `TestCloneRouteConfig_MapsAreIndependent` — verifies that mutating clone1 map keys does not affect original or clone2
- `TestCloneRouteConfig_NilMaps` — verifies that nil maps remain nil in clones (no spurious allocation)

**Files:** `internal/parser/parser.go`, `internal/parser/clone_test.go`

---

### #9 — No renderer test for URLRewrite (P2)

**Type:** Test gap.

**New test:** `TestRender_RoutePolicies_URLRewrite` in `internal/renderer/renderer_test.go`
- Verifies route with `Rewrite *RouteRewrite` is rendered with `rewrite:`, regex pattern, and substitution in the Route Policies block.

---

### #10 — No renderer tests for new match types (P2)

**Type:** Test gap.

**New tests** in `internal/renderer/renderer_test.go`:
- `TestRender_MatchTypes_PathSeparatedPrefix` — `(path-prefix)` label
- `TestRender_MatchTypes_Regex` — `(regex)` label
- `TestRender_MatchTypes_HeaderExact` — `header(name=value)` notation
- `TestRender_MatchTypes_HeaderRegex` — `header(name~value)` notation
- `TestRender_MatchTypes_QueryParamExact` — `query(name=value)` notation
- `TestRender_MatchTypes_QueryParamRegex` — `query(name~value)` notation

Shared helper `routeSnapshotWithMatch` extracted to reduce boilerplate.

---

### #14 — Roadmap phase status and scenario table stale (P3)

**Type:** Doc gap.

**Changes:** `docs/plans/kfp-roadmap.md`
- Phase 1 status: `Designing` → `Complete (patch cycles 1.1, 1.2, 1.3 applied)`
- Scenario 02_8: `Not yet` → `Yes`

---

## Test Results

All tests pass (`go test ./...`):
- `internal/parser` — 0.831s
- `internal/renderer` — 0.371s
- `internal/model` — cached
- `internal/envoy` — cached

## Deferred Issues (future cycles)

- **#11** (P2) — `prefix_rewrite` route action not captured
- **#12** (P3) — Parser silently drops malformed config sections
- **#13** (P3) — `weighted_clusters` not captured (traffic splitting)
