# Proof at main 89614cc0f

Historical receipt. The current product-backed Go controls and verifier are described in [the README](../README.md); the gate-only-anchor statements below describe the earlier workflow.

Base: `89614cc0f9d6aa7b7fe96b56c5c6c1c3eace9825`, fresh detached worktree. TypeScript v6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`. Recursive submodules initialized; locked `npm ci` in stage3/api succeeded. Go 1.27.1; temporary files, compiler cache and build cache were in scratch. No archived scout rule heads or production edits were used.

This records refusal-scan stops only, plus the first declaration, body or tail stop reached, and is blind past that first stop. A zero for a body or tail rule is never a pass. Each example is one root import-closure run.

## Tests and mutants

Since this proof, the eight controls run through `controls.py` from verify.sh rather than as Go tests, because a Go test that skipped without the binaries was a skip the gate's skip census refused (Loom's red on chg_497wcgbs). The checks, fixtures and mutants are the same; the names below are the old test names, each now the control of the same name in lowercase (`count-stop` and so on). The current box run is recorded on #et209ny.

The whole package passed in 1.609 s with the prebuilt plain production and scan binaries supplied, with no skipped controls. Every control first compiled the pragma-free input using plain production `load.Load` / `lower.Lower`: the four positive fixtures stopped with the intended words; all four clean twins lowered. Then the scan tool saw the earlier pragma and the later stop, or completed the clean twin. Positive controls also asserted the exact pipeline boundary.

Each of these passed normally and went red with its own `SCAN_ALL_MUTANT_FIRST_ERROR=1` run. The verifier required a lost-target or lost-completion assertion, not merely a nonzero process status:

- `TestScanAllCountStopAfterEarlierScanStop`: `checkBoundArguments`, `function_bind.go:377`; mutant red in 0.27 s.
- `TestScanAllCountCleanAfterEarlierScanStop`: clean count twin; mutant red in 0.24 s.
- `TestScanAllRestStopAfterEarlierScanStop`: `functionBind`, `function_bind.go:54`, first module-body stop under `moduleStatements`; mutant red in 0.31 s.
- `TestScanAllRestCleanAfterEarlierScanStop`: clean rest twin; mutant red in 0.27 s.
- `TestScanAllRecordStopAfterEarlierScanStop`: `noteRecordSlot`, `graph_records.go:16`, under `findCycles`; mutant red in 0.31 s.
- `TestScanAllRecordCleanAfterEarlierScanStop`: clean record twin; mutant red in 0.30 s.
- `TestScanAllCycleStopAfterEarlierScanStop`: array-of-record write, same production record validator under `findCycles`; mutant red in 0.27 s.
- `TestScanAllCycleCleanAfterEarlierScanStop`: clean array twin; mutant red in 0.33 s.

`TestScanAllMainAnchorsMatchExactly` passed with main's actual anchors. The hashes and `anchors.json` are gone. The unchanged named test was then run against isolated scratch copies of the source, checker and Go test:

- Removing `lowering.refuse(module)` made the named test RED and named that anchor (found 0).
- Duplicating `lowering.refuse(module)` made it RED and named that anchor (found 2).
- Removing `contracts = func(node *ast.Node) bool {` inside `refuse` made it RED and named that anchor (found 0).
- Duplicating that refusal visitor anchor made it RED and named it (found 2).
- Removing every pipeline-site regex match made it RED and named the `pipeline-site` check.
- Adding an unrelated comment line to `lower.go` kept it GREEN.

A PATH trap for Go was untouched by every anchor check: the test runs Python only, with no Go subprocess. The operator runner invokes the named Go test using the real Go path; all source mutations occur in temporary scratch copies, never in the checkout. In gate mode only the anchor test runs.

Both public no-IR guards ran on every scan control invocation. No control is owed. The direct-rest pair proves main's earlier body stop, rather than claiming a tail check it cannot reach.

## Gate packages and isolation

Whole-package results (`go test -count=1 -timeout=10m ./cmd/adamic-gate/... ./internal/compilercensus`):

- `cmd/adamic-gate`: PASS, 2.890 s.
- `cmd/adamic-gate/parallelcensus`: PASS, 3.819 s.
- `internal/compilercensus`: PASS, 3.878 s.

A repository search for `scan-all|SCAN_ALL`, excluding this tool directory and the cohere submodule, found only the compiler-dependency manifest entry. No gate reads the stream or the summary output. `git diff --name-only HEAD -- internal/lower stage3/meter stage3/fixtures` was empty. No Loom command was run.

## Closure examples

See [sys.ts](sys.md) and [corePublic.ts](corePublic.md). The summary tables separate events from distinct `(rule, site, words)` triples and group files from observation sites. Raw streams are not committed. Measurements exclude adaptation and builds.

- `sys.ts`: 366.8239 s wall, 1,900,544 KiB peak RSS; 5,545 events, 4,029 distinct triples. This is 5,544 scan-stop events plus the first declaration stop, the `filter` overload result at `core.ts:261:1`. Body and tail outcomes past it are unknown.
- `corePublic.ts`: 0.1064 s wall, 32,604 KiB peak RSS; 0 events and 0 distinct triples. Traversal completed. These zeros are not body-rule or tail-rule passes.

Both receipts use Linux `wait4` usage for the exact child, reporting kernel peak RSS in KiB. Temporary streams stay in scratch and are deleted; only these summaries are committed. The source adaptation was unchanged by the main-tip carry-forward and was reused.


The sys.ts and corePublic.ts example receipts are retained from `8076318c`: newest main changed only the sealed-job script and fixture tests, not the production compiler or source adaptation. After this anchor-check refactor, both rewritten compiler source copies and the rebuilt scan-all binary were byte-identical to that measured version. Census behavior and its recorded counts therefore have not changed; no example tables were recut. There are no whole-file source hashes in the gate check.
