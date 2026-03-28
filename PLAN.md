# [PLAN] Support Dual Receipt Modes

## Impact Analysis
- `backend/internal/config/config.go`: Add `USE_OPTIMIZED_RECEIPT` toggle.
- `backend/internal/utils/receipt.go`: Support switching logic.
- `backend/internal/app/app.go`: Pass config.
- `backend/tests/receipt_test.go`: Update test init.

## Breaking Changes
- `utils.InitReceiptRenderer` signature change (private change, but affects test).

## Verification
- Run `go test -v ./tests/receipt_test.go`.
- Manually check both modes.
