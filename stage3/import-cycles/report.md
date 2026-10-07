Built: public RunRule integration, ESM-order proof with no checks on proven reads, TDZ fallback without cycle-read refusals, re-export barrels, and original-ledger classifications.
Commits: main merge 8017f4c, integration shim merge cac8c24, runtime acceptance 2f5277e, immediate scanner unblock push 824de024ac3093fab75f95b1f53824118ac1f795.
Checks: build, full load/lower packages, all uncached Node/native fixtures (50.753s), counts, formatting and vet passed; exact enum probe prints 1 with its existing enum dependency.
Mutants: seven current mutants caught for order, TDZ fallback, class readiness, source error names, proof/check elimination and public-runner invocation; earlier interim mutants are historical.
Not covered: full native tsc or unexecuted branches; ledger binding obligations are 55 proven/3 checked/0 refused statements and 905 proven/9 checked/0 refused reads.

The final ruling and current results are in the final section below. Earlier sections retain historical checkpoints and are superseded where they describe the interim guard or blocked dependencies.

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

## ca939db5 compatibility stop, October 7

The public runner is now available. Fetched cohere main and checked out ca939db5cc348dc7dabbf4e6fd89dfc1dd9865f4. Recursive submodule update succeeded and moved its TypeScript dependency to d92d9bfee114c80be2c375d72edae966176e3a4f, from Adamic's verified 8d550c837c90bd1805b047b7eeccc2baac2d5e7a. Read the actual public API and the rule source/documentation. No cohere code was copied.

Compatibility commands, with /workspace/adamic-tools/env.sh sourced:

- `go build ./... > /tmp/import-cycles-new-pin-build.log 2>&1`: exit 1.
- `go test -count=1 -timeout 30m ./internal/load ./internal/lower > /tmp/import-cycles-new-pin-tests.log 2>&1`: exit 1; both packages fail compilation, so no test bodies execute.

The compiler-host API no longer takes the current-directory argument. Source files and VFS now use typed RootedFilePath/RootedDirectoryPath values. SourceFile.Path, tspath.Path, tspath.ToPath and ComparePathsOptions are absent. NewParsedCommandLine takes five arguments rather than four; the VFS no longer exposes UseCaseSensitiveFileNames. Adamic's sourceFS and regexpLibraryFS implement old string-based methods and no longer implement the VFS interface. The same transition breaks bridge/tsgo/checker and bridge/tsgo/oracle, including config-file parsing and root/source-file lookup. Complete output is retained under runner-pin-failure/.

This is the previously specified stop condition: “Check that the submodule bump doesn't break the shims (go build ./... and the load/lower tests); if it does, stop and tell me what broke.” I did not repair the wider shim migration, change go.mod, replace the interim, update the verified-pin constant, or push an unbuildable dependency bump. The cohere checkout is left at the tested requested pin as an uncommitted submodule change; the report commit retains the previous recorded pin. The TDZ fallback code remains intact.

The original ledger totals remain 58 statements and 914 observed reads unassessed. No classification as refused/checked/proven safe was obtained for these original source sites; the ten reduced-fixture interim refusals are separate evidence. Continuing requires a shim-compatible public-runner pin or a separate authorized migration of the loader and bridge APIs. Only codex/import-cycles is pushed; no main or area/ branch is modified.

## Final ruling: accept or check, never refuse a cyclic read

Merged current origin/main c7991b9 into codex/import-cycles. The only conflict was counts.md: kept both import-cycle rows and main's devirtualize row; regeneration moved the latter to its actual registration order without changing its values. Then merged integration's origin/codex/cohere-bump-ca939db tip 1ef4b34, which records cohere 7945d102. No independent cohere bump was made. The integrated loader/bridge adaptations remove the earlier shim blocker.

Removed the conservative implementation completely. The small adapter calls the public `rule_runner.RunRule(program, checker, files, "nexus/correctness-no-import-cycle-load-time-read")` under a process-wide mutex while Lower holds the checker exclusively. The rule returns hazardous-read findings, not positive proof certificates. Its findings are retained as evidence, and no finding causes a Refused result. Absence of a finding does not waive a TDZ check.

The actual entry's DFS post-order proves a direct module-body read initialized when the declaration's provider module finished earlier. Only module-level function declarations receive the unconditional hoisting proof; a function nested in a namespace does not. Reads in deferred function/constructor/instance-initializer code remain checked because this pass does not propagate call-time context. Unproved global reads keep the existing ready checks, including indirect calls, class construction and extends. Proven class reads also omit redundant guards. Import/re-export declarations are linked through checker symbols and schedule their dependencies; the declarations themselves have no body instruction. ExportAssignment remains governed by the existing policy.

### Immediate scanner publication

The exact four-file probe from codex/stage3-scanner-proof was copied unchanged. Current main still rejects its const-enum syntax with TS1294: that is the existing enum unit's prerequisite, not a cycle refusal. An isolated worktree combined cycles with codex/flag-enums ec67b02, preserving main's predicate/invariance behavior and the enum checks. Its re-export barrels required the now-published no-op admission through the graph linker. The isolated combination was never pushed.

Native and independent Node both printed `1\n`, exit 0. The generated C contains no `ReferenceError: Cannot access` guard. Published codex/import-cycles immediately at 824de024ac3093fab75f95b1f53824118ac1f795 and sent that SHA before continuing. The source probe, its existing enum dependency, output and generated-C hash are recorded in final-evidence/scanner-probe.json. The cycles feature does not claim to implement enums.

The pinned premature variant uses a lexical `const Category = { Error: 1 }` and imports types before the barrel, so the reader reaches Category while its provider is in progress. Source Node, sanitized native, release native and generated JavaScript agree on empty stdout, exit 70, and:

```
adamic: panic: ReferenceError: Cannot access 'Category' before initialization
```

A safe sibling enters the same object-binding graph through the barrel and prints 1. Both are registered in the differential oracle. A premature const-enum constant-folding example would not be a valid TDZ witness, so the early-read fixture deliberately uses a runtime lexical binding.

### Original ledger classification

Fetched pristine TypeScript v6.0.3 at the exact ledger commit 050880ce59e30b356b686bd3144efe24f875ebc8 and generated diagnosticInformationMap with upstream's own script. The audit builds one compiler program rooted at src/tsc/tsc.ts, calls the actual public rule, and uses the same esmModuleOrder/proveModuleReads helpers as Lower. Its computed order of all 80 modules equals the independently verified Node order byte for byte. Every one of the 914 original read occurrences resolves at its source location to the exact declaration/provider recorded by the ledger; missing or mismatched identities fail instead of being counted safe. The direct conditional Debug read is classified too.

| Bounded ledger population | Proven | Checked | Refused |
| --- | ---: | ---: | ---: |
| Original top-level statements | 55 | 3 | 0 |
| Observed original read occurrences | 905 | 9 | 0 |

The checked statements are semver.ts:46, parser.ts:1437 and binder.ts:499. Their checked occurrences are Debug at semver.ts:63,64,65,70,71; emptyArray at semver.ts:68; Debug at scanner.ts:4020 and scanner.ts:1115; and Debug at binder.ts:496. These are inside deferred bodies and retain the runtime fallback rather than inheriting the observed invocation's safe timing. Hoisted module function bindings and direct reads after their provider completes are proven. A row is checked if any of its bounded ledger reads is checked; otherwise it is proven.

These are source binding-initialization proof/check obligations, not a claim that complete native tsc artifacts were emitted. Other feature gaps still prevent that build. A proven row covers its bounded recorded reads and the one inventoried conditional read, not every possible branch of its initializer. Unexecuted branches in deferred code retain checks. The run produced 10 real cohere findings; findings such as timestamp in performance.ts:62 are accepted under the fixed entry-order proof instead of being refused. The full original rule ids and messages are preserved in original-ledger-classification.json.

Reproduce the source audit after cloning the pinned upstream source and running its diagnostic generator:

```
source /workspace/adamic-tools/env.sh
ADAMIC_CYCLE_LEDGER_ROOT=/tmp/import-cycles-original-tsc ADAMIC_CYCLE_LEDGER_OUTPUT=/tmp/cycle-ledger.json go test ./internal/lower -run '^TestOriginalCycleLedger$' -count=1 -v > /tmp/cycle-ledger.log 2>&1
```

### Validation and limits

`go build ./...` passed after the integration merge and adapter update. Full `go test -count=1 -timeout 30m ./internal/load ./internal/lower` passed (load 0.643s, lower 12.135s). `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestImportCycle|TestNativeAgreesWithNode' -count=1 -timeout 30m` passed in 50.753s across the complete registered Node/native fixture corpus. The explicit value-probe filter selected both new fixtures and passed uncached. The seven ready-check cases passed. The original-ledger audit passed in 2.253s.

`go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts` passed in 12.662s. Changes: the safe value probe adds `3 3 0 4 3 0` for its object/formatting lifetime; the premature probe adds all zeroes because it stops before its provider allocates; devirtualize's existing row moves with unchanged counts to match registration order. No other count values changed. `gofmt -l cmd internal` printed nothing, `go vet ./...` exited 0, and git diff --check passed. Test output was retained in log files.

All ten reduced tsc fixtures were rerun through the real runner and fresh Node. They no longer fail on cycle reads. Six stop on the existing boolean|undefined truthiness refusal (01-04, 07, 08); two stop on never[] representation (05, 06); two stop on unknown representation (09, 10). Complete diagnostics and Node outputs are under runner-run/. No unrelated compiler feature was added to close those gaps.

Seven current mutants all exit 1 due to the intended observations, not compilation warnings: wrong module order changes Node's printed sequence; removing ready checks produces undefined/0 instead of the value TDZ panic; keeping a proven value check violates the emitted-C assertion; bypassing RunRule empties the pinned original-source findings; dropping class storage prints 7 instead of the Box TDZ panic; checking the internal base storage says Base_class instead of Base; keeping a proven extends check violates its emitted-C assertion. Logs are in final-evidence/. The earlier drop-refusal mutant is superseded because cycle-read refusal is now explicitly wrong.

The full all-packages gate and full native TypeScript compiler were not run. Multi-entry remains out of scope. A normal workspace build is green; go mod tidy's transitive dependency-test download of klauspost/compress v1.20.0 was denied by the proxy, so tidy is not reported as successful. No cohere implementation was copied. Only codex/import-cycles is pushed; main and area/ branches remain integration-owned.
