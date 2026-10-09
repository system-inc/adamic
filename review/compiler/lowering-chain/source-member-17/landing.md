Rebuilt step 21 exceptions on main, retaining the views frame and predicate fixes.
Base: 7a10c877; source net change: d72728e5..1ac4d401; delivery: compiler/exceptions-21-main.
Build, vet, changed-file a-check, stage-1 probes, stage-3 fixtures and the focused oracle passed.
The mutant evidence records each production, semantic-fixture and census failure separately.
Full tsc compilation, unrepresented Error APIs and allocation exhaustion remain outside this slice.

The compiler admits owned tagged throws, unknown catches with real narrowing, catch reassignment, try/catch/finally across callbacks and returns, Error subclasses with shared layout and nominal ancestry, and saved Error identity through callbacks and rethrow. Represented library failures are catchable. Compiler checks and readiness guards remain terminal. Uncaught values unwind and exit 1 without an implicit renderer. These are the source branch's adopted rulings; this rebuild introduces no new language choice.

The source branch's net patch was applied onto main, without its old base or merge history. Six conflicts were resolved:

- `internal/javascript/javascript.go`: retain main's static-method map and add the source's builtin Error prototype ancestry before class identity registration.
- `internal/lower/exceptions.go`: retain tagged arbitrary payload lowering and remove the obsolete Error-only and constant-origin restriction.
- `internal/lower/expression.go`: retain main's `localRead` dispatch and namespace readiness behavior; place the source's nominal Error ancestry checks in `internal/lower/locals.go`, alongside the existing array predicate check. Update the two mutant runners to target that moved code.
- `internal/lower/library_object.go`: retain main's enumeration and presence support; add the source's refusal for Error reflection whose own descriptors are not represented.
- `internal/lower/lower_test.go`: remove obsolete library and mutable saved-Error refusal expectations, retaining unrelated refusal probes and the source's positive admission tests.
- `internal/oracle/counts.md`: regenerate from the merged compiler rather than select either conflicting numeric snapshot.

The requested stage-1 sweep found three closed gaps still expected to refuse: CSS and JSON's repeat-in-try reductions and GraphQL's Error constructed elsewhere. Their unchanged source programs now have positive Node, native sanitizer, backend and leak assertions. Their gap documents record the admission. Existing adaptation workarounds remain with their owners. The unknown-read proposal explicitly expects TS18046, preserving its intended refusal.

Commands used with `source /workspace/adamic-tools/env.sh`; test output was written directly to files:

- `go build -o /tmp/exceptions-main-adamic ./cmd/adamic`: exit 0.
- `go vet ./internal/...`: exit 0.
- `python3 /tmp/exceptions-main-a-check.py --out /tmp/exceptions-main-a-check`: exit 0 using the gate policy at tools commit 3e339bbf06e0695f1b1223065ee8d4eef4813ba8. Against origin/main, 21 changed .a files: 20 checked and one expected TS18046. The deliberate unknown read is not marked passing.
- `go test ./stage1/... -run 'Gap|Gaps|Probes' -count=1 -timeout=30m`: repaired sweep exit 0, all 32 tested packages passed. The initial sweep's three stale expectations are recorded above.
- `go test ./stage3/fixtures -count=1 -timeout=30m`: passed, 263.876 seconds.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle ./internal/lower -run 'TestStep21|TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat|TestWhatZeroOneRefusesIsRefusedWithAFix|TestPrototype|TestInheritance|TestNativeAgreesWithNode/internal/oracle/testdata/(step21_|exceptions|closures_throw|finally_leaves|regions_throw|reuse_throw|node_buffer_bom)' -count=1 -v -timeout=30m`: passed, oracle 259.897 seconds and lower 16.925 seconds. Own fixtures agree with Node in both backends, with native release, ASan/UBSan and leak checks. The five semantic IR mutants were rejected by their Node output assertions.
- `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout=30m -args -update-counts`: passed, 345.713 seconds. [The row ledger](landing-counts.md) explains all 16 additions and 91 changed existing rows; no rows were removed.
- Production and census mutants: `python3 /tmp/exceptions-main-mutants.py` (preserved as `run-landing-mutants.py`), running the source branch's seven production runners serially, then both census mutants. [Fresh evidence](landing-evidence/) records the catcher for each; historical source evidence is preserved separately.

Toolchain setup used `GOPROXY='https://proxy.golang.org|direct'`. The initial setup build overlapped patch application and failed on temporary conflict markers. After resolving them, `bash cloud/setup.sh` passed: Go 0.041s, Node 0.045s, markdown 0.122s, submodules 0.127s, clang 0.375s, Go build 112.285s, deferred test binaries 122.479s, cache warm 122.503s, done 123.093s. `nproc` returned 5; the box's CPU quota is 4 cores. Reproducible old Go build artifacts were removed to make room for the sweeps; no source or historical evidence was removed.

The fresh tsc driver uses the adapted TypeScript 6.0.3 entry and the stock checker agrees on its 81-file graph. The first native stop is `src/compiler/builder.ts:1246:69`, TS2345: `Path | undefined` is not assignable to `string`. Both native split modes stop there. This checker barrier precedes exception lowering, so the entry probe does not demonstrate full native tsc execution or reach every remaining exception site.

Still refused or unrepresented: Error stack/options/cause and reflection, unrepresented host/library failure families, captured unknown cells and allocation exhaustion. These need their own contract or language decision before implementation. No new ruling is needed for the transported admitted behavior. No full gate or macOS run was performed.

All 40 mutants were caught: 33 production/harness, five semantic IR and two census. The initial census attempt lacked historical object 6c4fc1af and did not count as a catch. After fetching full object `6c4fc1afb019d0a16fea97082e45fef8fcbe1746`, the unmutated census passed and both mutants failed their intended assertions. The production changes were already restored; the final focused suite passed again (oracle 13.203s, lower 6.991s), and the rebuilt compiler produced identical tsc diagnostics in both split modes. Node prints `Version 6.0.3`.

| Production/harness mutant | Intended catcher observed |
|---|---|
| `missing-exception-edge` | runtime error: |
| `assignment-kills-handler-value` | old text is dead |
| `throw-temporaries-leak` | leaks: |
| `pending-finally-leak` | leaks: |
| `error-message-not-retained` | AddressSanitizer: heap-use-after-free |
| `unknown-authorizes-assertion` | admission check |
| `implicit-string-conversion` | uncaught lifetime check: exit 70 |
| `pending-undefined-lost` | stdout differs |
| `catch-assumes-error` | exit codes differ |
| `typeof-boolean-is-number` | stdout differs |
| `null-is-undefined` | stdout differs |
| `finally-releases-twice` | AddressSanitizer: heap-use-after-free |
| `region-payload-freed` | AddressSanitizer: heap-use-after-free |
| `uncaught-is-panic` | uncaught lifetime check: exit 70 |
| `range-error-is-error` | stdout differs |
| `wrong-range-message` | stdout differs |
| `type-error-is-error` | stdout differs |
| `hash-failure-is-panic` | exit codes differ |
| `host-type-error-is-error` | stdout differs |
| `soundness-check-is-catchable` | want the inserted check to fire |
| `error-ancestry-missing` | stdout differs |
| `error-default-name-wrong` | stdout differs |
| `error-prefix-released-twice` | AddressSanitizer: heap-use-after-free |
| `error-fields-out-of-order` | stdout differs |
| `rethrow-wrapped` | stdout differs |
| `builtin-narrow-forgets-subtype` | want the inserted check to fire |
| `saved-error-wrapped` | stdout differs |
| `backend-renders-stack` | uncaught renderer wrote stderr |
| `uncaught-sanitizer-visible` | sanitizer failure despite matching stdout and exit |
| `changed-source-keeps-terminal-mode` | changed source borrowed terminal convention |
| `changed-dependency-keeps-terminal-mode` | changed dependency borrowed terminal convention |
| `backend-drops-entry-source` | JavaScript backend: exit codes differ |
| `readiness-runs-handler` | JavaScript readiness guard ran catch/finally |

Restored-tip `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout=30m` passed without updating the table: ok  	github.com/system-inc/adamic/internal/oracle	73.142s.
