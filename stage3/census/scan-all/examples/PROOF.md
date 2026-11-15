# Proof at main ccf3cd88e

Base: `ccf3cd88eb04da8334ab3919fd46dd46464682f9`, fresh detached worktree. TypeScript v6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`. Recursive submodules initialized; locked `npm ci` in stage3/api succeeded. Go 1.27.1; temporary files, compiler cache and build cache were in scratch. No archived scout rule heads or production edits were used.

This records refusal-scan stops only, plus the first declaration, body or tail stop reached, and is blind past that first stop. A zero for a body or tail rule is never a pass. Each example is one root import-closure run.

## Tests and mutants

The whole package passed in 1.717 s with the prebuilt plain production and scan binaries supplied, with no skipped controls. Every control first compiled the pragma-free input using plain production `load.Load` / `lower.Lower`: the four positive fixtures stopped with the intended words; all four clean twins lowered. Then the scan tool saw the earlier pragma and the later stop, or completed the clean twin. Positive controls also asserted the exact pipeline boundary.

Each of these passed normally and went red with its own `SCAN_ALL_MUTANT_FIRST_ERROR=1` run. The verifier required a lost-target or lost-completion assertion, not merely a nonzero process status:

- `TestScanAllCountStopAfterEarlierScanStop`: `checkBoundArguments`, `function_bind.go:377`; mutant red in 0.22 s.
- `TestScanAllCountCleanAfterEarlierScanStop`: clean count twin; mutant red in 0.33 s.
- `TestScanAllRestStopAfterEarlierScanStop`: `functionBind`, `function_bind.go:54`, first module-body stop under `moduleStatements`; mutant red in 0.22 s.
- `TestScanAllRestCleanAfterEarlierScanStop`: clean rest twin; mutant red in 0.24 s.
- `TestScanAllRecordStopAfterEarlierScanStop`: `noteRecordSlot`, `graph_records.go:16`, under `findCycles`; mutant red in 0.24 s.
- `TestScanAllRecordCleanAfterEarlierScanStop`: clean record twin; mutant red in 0.28 s.
- `TestScanAllCycleStopAfterEarlierScanStop`: array-of-record write, same production record validator under `findCycles`; mutant red in 0.29 s.
- `TestScanAllCycleCleanAfterEarlierScanStop`: clean array twin; mutant red in 0.25 s.

`TestScanAllMainAnchorsMatchExactly` passed with main's actual sources. Replacing the expected lower.go digest with an incorrect digest made that named test red, explicitly naming `internal/lower/lower.go`; restoring the digest made it green again (0.135 s). No production file was altered for this mutant.

Both public no-IR guards ran on every scan control invocation. No control is owed. The direct-rest pair proves main's earlier body stop, rather than claiming a tail check it cannot reach.

## Gate packages and isolation

Whole-package results (`go test -count=1 -timeout=10m ./cmd/adamic-gate/... ./internal/compilercensus`):

- `cmd/adamic-gate`: PASS, 5.760 s.
- `cmd/adamic-gate/parallelcensus`: PASS, 8.866 s.
- `internal/compilercensus`: PASS, 8.642 s.

A repository search for `scan-all|SCAN_ALL`, excluding this tool directory and the cohere submodule, found only the compiler-dependency manifest entry. No gate reads the stream or the summary output. `git diff --name-only HEAD -- internal/lower stage3/meter stage3/fixtures` was empty. No Loom command was run.

## Closure examples

See [sys.ts](sys.md) and [corePublic.ts](corePublic.md). The summary tables separate events from distinct `(rule, site, words)` triples and group files from observation sites. Raw streams are not committed. Measurements exclude adaptation and builds.

- `sys.ts`: 366.8239 s wall, 1,900,544 KiB peak RSS; 5,545 events, 4,029 distinct triples. This is 5,544 scan-stop events plus the first declaration stop, the `filter` overload result at `core.ts:261:1`. Body and tail outcomes past it are unknown.
- `corePublic.ts`: 0.1064 s wall, 32,604 KiB peak RSS; 0 events and 0 distinct triples. Traversal completed. These zeros are not body-rule or tail-rule passes.

Both receipts use Linux `wait4` usage for the exact child, reporting kernel peak RSS in KiB. Temporary streams stay in scratch and are deleted; only these summaries are committed. The source adaptation was unchanged by the main-tip carry-forward and was reused.
