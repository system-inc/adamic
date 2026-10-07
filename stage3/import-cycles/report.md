Built: ESM cycle ordering, graph-wide linking, runtime TDZ fallback, and merged all ten tsc cycle fixtures; public-runner integration remains blocked.
Commits: implementation 56b95a9, original report 650a221, fixture merge a852810 (bucket a4ec90d).
Checks: prior build/load/lower/oracles passed; latest lowering refused 10/10 fixtures, fresh Node observations matched 10/10 records, ledger audit passed.
Mutants: prior eight caught; merged ledger audit caught a removed row and a late provider-end event.
Not covered: real runner/new-pin shims and original tsc lowering; 58 ledger statements and 914 observed reads remain unassessed, not proven safe.

The latest instruction supersedes the interim implementation. The working tree still holds that earlier, tested checkpoint; it is not a completed public-runner integration. No cohere source was copied. The replacement adapter is drafted at /tmp/import-cycles-public-adapter.go, with the exact new error return, Finding type, rule name, and a mutex protecting calls across programs. I have not installed an unavailable dependency or claimed it builds.

## Fetch blocker

Repeated fetches of cohere main, including the final delivery retry, returned e7cfe4d1aceb524bc6f5936564284d54ca63d771. `git -C cohere ls-tree FETCH_HEAD rule_runner` printed nothing. `git -C cohere show d3ac7d1e:rule_runner/rule_runner.go` returned `fatal: invalid object name 'd3ac7d1e'.` Advertised refs contain no runner ref or supplied commit prefix. GitHub's public commit endpoint returned HTTP 422:

```
{"message":"No commit found for SHA: d3ac7d1e","status":"422"}
```

The cohere checkout remains 715ba94f3608a6500086b1076ce5cb7e51b836db. No submodule bump was made, so there is no observation of a shim break. A fetchable full SHA or published ref is needed. Once available: pin it in its own commit; run `go build ./...` and all load/lower tests; stop if the shims break; otherwise add `replace github.com/system-inc/cohere => ./cohere` and the requirement, install the drafted public adapter, remove the interim, and rerun the oracles and mutants against the real rule. Its findings must preserve the original rule id, message (including newlines), and 1-based byte columns.

The census branch also could not be read: its fetch returned `fatal: could not read Username for 'https://github.com': No such device or address`, and the public raw report URL returned 404. The latest ruling describes a 76-file cycle and 77 export-list lines; these are user-supplied facts, not observations from that report.

## Stage 0 today, observed at ef3d907

`internal/load/load.go` checks the graph using the in-process checker, accepting cyclic imports. Lowering then refuses every back edge in `internal/lower/modules.go`:

```go
case 1:
    return &Refused{Where: l.program.Where(from), What: "an import cycle", Fix: "move what both modules need into a third that neither imports"}
```

The old declaration loop invokes `declareModule` separately for each module. That function registers its declarations and signatures, then lowers its function bodies immediately. An earlier module cannot lower a function referring to a later module's function. Imported bindings resolve through `GetAliasedSymbol` and `GetExportSymbolOfSymbol` in `locals.go`, so their globals are live shared slots, not copied import values. Function reads of globals carry the existing temporal-dead-zone check.

The working changes skip modules already on the DFS stack, place each module once in import-statement order after its dependencies, and concatenate declarations for a single registration/signature/body pass before any module body is lowered. Explicit `import type` and `export type` have no evaluation edge. Inline type specifiers leave the named import's module evaluation edge, as Node's type stripping and verbatimModuleSyntax do.

## Independent Node observations

The source oracle uses Node's native ESM evaluator with only extension loading and type stripping. A separate raw Node loader, without Adamic's runtime normalization, produced:

- Body order: `shared`, `c`, `b`, `a`, `inline`, `main`, each once, exit 0. The explicit type-only target printed nothing; the inline-type import's target printed `inline`.
- Two modules with mutually recursive functions: bodies `odd body`, `even body`; calls made after evaluation print `1`, `1`. The native and JavaScript IR harness invokes the same exports after their module bodies and agrees with Node. Both native release and sanitized builds agree, and LeakSanitizer is clean.
- Three class modules entered at Leaf: `base`, `middle`, `leaf`, exit 0. Entered at Base: `ReferenceError: Cannot access 'Base' before initialization`, exit 1.
- Value-read cycle: `ReferenceError: Cannot access 'value' before initialization`, exit 1. The normal oracle runtime normalizes this to exit 70; the interim lowerer refuses it.

## Interim checkpoint and rule limitations

The interim was independently implemented behind the earlier RunRule interface and serial lock. It refuses every direct top-level imported value read anywhere in a reachable cyclic program, including hoisted functions, initialized exporters, and imports outside the reader's cycle. If deferred code uses imports, it additionally allows only inert module declarations and literal console output, rejecting user calls, callbacks, getters, coercions, dynamic output, and executable class shapes. This avoids accepting indirect load-time reads through a local function while awaiting the public rule.

Compared with the pinned cohere rule, the interim additionally refuses harmless hoisted-function reads and calls, reads outside the same strongly connected component, and harmless top-level local computations or calls in programs with deferred imports. Its safe-entry class refusal is conservative against Node, but is also a refusal the pinned real rule makes: that rule reports a class/variable read anywhere in the same cycle regardless of the chosen entry's actual evaluation order. Its documentation explicitly says it does not follow calls into hoisted functions and misses IIFEs. Its graph applies TypeScript import elision rather than retaining all verbatimModuleSyntax edges. These are observations of the old pinned source, not claims about the unavailable new package. The landed runner must be inspected for these limits before claiming the full requested semantics.

The public adapter draft calls `nexus/correctness-no-import-cycle-load-time-read` and returns the real Finding's Rule and Message in the refusal. The final integration must replace the checkpoint and update the hoisted-function and indirect-call expectations to the real package's observed behavior, keeping any unmet semantics explicit.

## Commands and outputs

All test output was written to logs before reading it. Setup used `bash cloud/setup.sh` followed by `source /workspace/adamic-tools/env.sh`.

```
go version go1.27.1 linux/amd64
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (105s)
setup: done in 105s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
nproc: 5
```

`gofmt -l cmd internal` printed nothing. `go vet ./...` exited 0. `go test -count=1 -timeout 30m ./internal/load ./internal/lower` passed: load 3.307s, lower 42.779s. `ADAMIC_GATE_UNCACHED=1 go test ./internal/load ./internal/lower ./internal/oracle -run 'TestImportCycle|TestNativeAgreesWithNode/internal/oracle/testdata/import_cycles' -count=1 -v` passed; the oracle took 6.627s. The load/lower packages have no tests matching this filter; their complete tests were run separately. `go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -args -update-counts` passed in 11.807s. Only the two added fixture rows changed; each is all zeroes.

The old-pin full command `ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./...` was stopped with exit 143 when the new runner instruction arrived, before changing any submodule checkout. This is not a successful full gate. Logs are under /tmp/import-cycles-*.log, especially setup, touched, final-focused, counts, gate, vet, and node-read.

Each mutant used a Go overlay, with `ADAMIC_GATE_UNCACHED=1`, `-count=1`, and output in its own log. The exact mutation and observation:

| Mutant | Check and observation |
|---|---|
| Replace post-order append with prepend | Filtered order oracle failed, exit 1: Node printed `shared c b a inline main`, native and generated JS printed `main inline a b c shared`. Both builds exited 0, so compilation or sanitizers did not catch it. |
| Change `if cyclic` to `if cyclic && false` | `TestImportCycleLoadTimeRefusals` failed, exit 1: every expected refusal became `<nil>`. |
| Force every ImportClause to have a runtime edge | Filtered order oracle failed, exit 1: native and generated JS additionally printed `type-only body must not run`. |
| Restore per-module declareModule calls | `TestImportCycleRuntimeCalls` failed, exit 1: `odd.a:3:23: stage 0 can't lower reading even yet`. |
| Disable only the deferred-import guard | Indirect refusal subtest failed, exit 1: expected refusal became `<nil>`. |

Overlays are /tmp/import-cycles-mutants/{wrong-order,drop-refusal,type-import-runs,module-at-a-time,drop-indirect}.json. Their logs are /tmp/import-cycles-mutant-{order,refusal,type,linking,indirect}.log. These prove the interim checkpoint's checks only; the refusal mutant must be rerun after public integration, and the interim-only indirect mutant will no longer apply.

## Multiple roots and remaining scope

Lower still takes one root. A single .a entry with side-effect imports can make the desired runtime graph reachable without a multi-entry API. Supporting several roots would require an explicit host policy for root evaluation order, one shared DFS state across roots, one shared declaration-linking pass, and tests for shared modules and cycles reached from different roots. This unit adds none of that.

The public package, submodule bump, compatibility build and updated real-rule tests remain blocked. This is a delivery checkpoint, not the finished public-rule integration. The tested checkpoint and a separate integration draft are available for review; neither is represented as the requested finished change.

## Additional ruling: undecided reads are checked

The entry is `src/tsc/tsc.ts`. The supplied description is one 76-file strongly connected component through `_namespaces` barrels, with 77 export-list lines. The export-list count is not a count of program roots. These remain user-supplied facts pending the fixture ledger.

The fallback is now implemented independently of the rule. A cyclic graph marks every global read and write with the existing ready check, including its module bodies. Function-body reads remain checked as before. Classes in cyclic graphs get ready storage even with no static members; construction uses the existing class-ready thunk. An extends read checks the public source binding rather than its internal storage, preserving both the TDZ point and the name in the ReferenceError. Function declarations remain hoisted and require no class/variable ready flag.

The conservative interim still refuses production load-time reads while public integration is blocked. To exercise the fallback without claiming results from unavailable cohere, `TestUndecidedCycleReadsUseReadyChecks` injects a test runner that returns no findings. Seven cases compare source Node, generated JavaScript, and sanitized native execution: an early imported value; an initialized imported value; an indirect read through a local function; early/initialized construction of a class without static members; and early/initialized extends. All passed. Finished native cases run with LeakSanitizer on; panicking cases stop with the same ReferenceError line and exit 70 as the Node oracle normalization. This validates fallback behavior, not public-rule recall.

Three more isolated mutants were run with `go test -overlay ... ./internal/lower -run 'TestUndecidedCycleReadsUseReadyChecks/<case>' -count=1 -v`, all exit 1:

| Mutant | Observed failure |
|---|---|
| Remove cyclic module-body readiness from checked | Early-value case: generated JS prints `undefined`, native prints `0`, both exit 0; source Node throws the required ReferenceError. |
| Omit ready storage for classes without static members | Early-class case: generated JS and native print `7` and exit 0; source Node throws `Cannot access 'Box' before initialization`. |
| Check the internal parent-class storage in extends | Early-extends case: generated JS and native report `Base_class`; Node reports `Base`. |

The corresponding overlays are /tmp/import-cycles-mutants/{drop-top-ready,drop-class-ready,internal-base-ready}.json and logs are /tmp/import-cycles-mutant-{top-ready,class-ready,base-ready}.log. The earlier five mutants were regenerated against this implementation and rerun, all exit 1.

Latest validation on the available old pin:

```
go build ./... > /tmp/import-cycles-old-pin-build.log 2>&1
# exit 0

go test -count=1 -timeout 30m ./internal/load ./internal/lower > /tmp/import-cycles-ready-packages.log 2>&1
# load 1.293s, lower 12.804s, exit 0

ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestImportCycle|TestNativeAgreesWithNode/internal/oracle/testdata/(import_cycles|class_)' -count=1 -v > /tmp/import-cycles-ready-oracle.log 2>&1
# PASS, 5.766s; this regex also selects fixture names containing class_

go vet ./... > /tmp/import-cycles-ready-vet.log 2>&1
# exit 0
```

The latest cohere main retry still exposes no rule_runner directory. `git fetch origin codex/stage3-fixtures-cycles:refs/remotes/origin/codex/stage3-fixtures-cycles` returned `fatal: couldn't find remote ref codex/stage3-fixtures-cycles`. The ledger has not been read, so its total/refused/checked/proven-safe counts are pending, not zero. Once published, run its actual fixtures from the declared entry and tally the ledger against observed rule findings and emitted ready checks. A separate lowering gap must be reported as such, never counted as proven safe.

## Published fixture bucket, October 7 follow-up

Merged origin/codex/stage3-fixtures-cycles at a4ec90d, with merge commit a852810. Read its README and LEDGER whole. Ran `go run ./cmd/adamic js stage3/fixtures/cycles/<fixture>/main.a` for every one of the ten entries, capturing complete output in `ledger-run/*.lower.log`. All ten reached the interim rule and were refused; no JavaScript or native artifact was emitted. These are interim messages, despite carrying the configured nexus rule id; they are not cohere findings.

Ran fresh `node --disable-warning=ExperimentalWarning oracle/node.mjs <entry>` for each entry. All ten stdout/stderr/exit triples equal the bucket records: nine exit 0; fixture 06 exits 70 with `adamic: panic: ReferenceError: Cannot access 'emptyArray' before initialization`. The source bucket's import-order mutant is therefore still distinguished by Node. Our current refusal of both fixtures 05 and 06 does not demonstrate that Adamic distinguishes safe scheduling.

| Population | Refused | Checked | Proven safe | Unassessed |
| --- | ---: | ---: | ---: | ---: |
| Ten reduced fixture programs | 10 | 0 | 0 | 0 |
| 58 original tsc ledger statements | 0 | 0 | 0 | 58 |
| 914 observed original read occurrences | 0 | 0 | 0 | 914 |

The original-read zeros mean no compiler classification was obtained, not acceptance. The bucket states that fixture 05 is a scheduling witness for all ledger rows, not a reproduction of their initializers. It explicitly omits the complete factory initializer and many type/allocator dependencies. Consequently refusing that witness cannot establish that 58 original statements or 914 original reads were refused. Nor can the observed Node ordering be promoted to a certificate for branches not executed. Runtime-check coverage for those original sites requires actual lowered original code and source-site mapping; neither is present in this run. The conditional static read is also unassessed.

`python3 stage3/fixtures/cycles/audit.py > /tmp/import-cycles-ledger-audit.log 2>&1` passed and printed `removed ledger row: caught`, `late provider end: caught`, and `Trace, ledger, ten fixture records and chronology: pass`. The fixture/Node reconciliation asserted ten entries and exact equality for each observation. Detailed results and machine-readable counts are in `ledger-run/results.json` and `ledger-run/summary.json`.

Retried cohere main with `git -C cohere fetch --no-recurse-submodules origin main`: FETCH_HEAD remains e7cfe4d1aceb524bc6f5936564284d54ca63d771 and `git -C cohere ls-tree FETCH_HEAD rule_runner` prints nothing. Thus the requested real-rule replacement is still blocked. No compiler code changed in this follow-up and the full compiler gate was not repeated. No code was copied from cohere.
