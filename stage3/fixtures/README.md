# Stage 3 fixtures

Each bucket owns its .a programs and status.json. The schema is the shared Stage
3 fixture contract: file, tsc source spans, census reason, recorded Node stdout,
stderr and exit, and stage0 outcome plus full diagnostic text. Diagnostics use
repository-relative paths, omit the CLI's adamic prefix and final newline, and
preserve diagnostic content. TypeScript attribution is in NOTICE.

This unit seeds runner/ with NotYet, Refused, and matching Compiles observations.
Run the parallel fixture gate from the repository root:

```sh
source /workspace/adamic-tools/env.sh
go test ./stage3/fixtures -count=1 -timeout 10m -v > /tmp/stage3-fixtures.log 2>&1
```

Every `*/status.json` is read, with buckets and fixtures run in parallel. Each
fixture has named `node`, `stage0`, and, when it currently compiles, `native`
checks. Node is run afresh from source through the current oracle/node.mjs;
stdout, stderr, and exit must equal the recorded observation. stage0 checks the
outcome and complete diagnostic. A change says `gap changed: update status.json
and check the native output`. Newly compiling gaps run natively even when their
record still says NotYet or Refused, so a gap closure cannot hide wrong output.

Native checks build the oracle test binary once and invoke its small exported
`TestStage3FixtureHook`. The hook reuses `lowered`, `nativelyUncached` and
`leaksUncached`: the same ASan/UBSan flags, bounded process execution and platform
leak checks as internal/oracle. No helper implementation was copied and no
oracle_test.go or production compiler file was changed. Observations are uncached.
The hook transports bytes as base64 JSON and is dormant in ordinary oracle runs.
Node and native execute with the same repository-root working directory.

```sh
go test ./stage3/fixtures -count=1 -timeout 10m -args -update > /tmp/stage3-fixtures-update.log 2>&1
```

`-update` refreshes only currently compiling fixtures whose native output equals
current Node and whose recorded Node observation is still correct. It replaces
only the stage0 JSON value; all bytes outside that value, including node,
provenance, extra fields, ordering and whitespace, remain unchanged. Noncompiling
fixtures cannot be refreshed automatically because they have no matching native
output to prove the change. Miscompiles and stale Node expectations fail and are
never updated. Each changed status file is replaced atomically after its parallel
fixtures finish, and a concurrent external change aborts the write.

For audits, `-args -fixtures /tmp/copied-fixtures` selects a scratch fixture root.
The test still uses the repository's compiler, runtime and Node runner. The
three requested mutants and additional diagnostic/update guards were run on
scratch copies; see ../meter/REPORT.md. The native mutant uses a Go overlay of
runtime/string_build_impl.h, changing only "true" to "truf".

No other worker's bucket has been edited.
