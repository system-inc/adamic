# W1 file claim

Branch: codex/wasm-exports

Existing files: cmd/adamic/request.go, cmd/adamic/request_test.go,
internal/native/request.go, cmd/adamic/tsgo.go (flag parsing),
internal/native/target.go (link flags).

New files: docs/wasm-abi.md, internal/native/wasm/exports.mjs,
cmd/adamic/exports.go, cmd/adamic/exports_test.go,
internal/native/wasm_exports.go, internal/native/runtime/wasm_exports.c,
internal/native/wasm/exports/types.a,
internal/native/wasm/exports/oracle.mjs,
cloud/reports/wasm-exports/report.md.

User-authorized follow-up: cmd/adamic/main.go, minimal prefix --reactor parsing hook.
