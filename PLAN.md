# Backend Testing Implementation Plan (PH-0 Protocols)

## 1. Impact Analysis
- **Parser**: `internal/parser/parser.go` will be tested against 20+ OCR text variations.
- **UseCases**: `internal/usecase/shipment.go` & `internal/usecase/config.go` will be tested with mocks.
- **API Handlers**: `internal/transport/http/handler/shipment.go` will be tested via `httptest`.
- **Project Structure**: A new `backend/tests/` package will be created for black-box testing.

## 2. Breaking Changes
- **None**. No source code will be modified; only new `.go` test files and mocked interfaces will be added.

## 3. Verification
- **Run `go test -v ./tests/...`** to verify all tests pass.
- **Run `go test -coverprofile=coverage.out ./...`** to check coverage metrics.
- Target: >60% coverage on critical logic packages (`parser`, `usecase`).

## 4. Execution Roadmap
1. Create `backend/tests/` directory.
2. Implement **Deep Parser Tests** (20+ OCR patterns).
3. Implement **UseCase Integration Tests** (Mocking `db.Querier`).
4. Implement **API E2E Tests** (HTPP Handlers + RBAC).
5. Final QA & Cleanup.
