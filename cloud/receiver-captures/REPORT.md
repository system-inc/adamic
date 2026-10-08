6cc45783: reproduced reductions of c6ea4131 core.ts:2100:26 compare and debug.ts:376:26 toString; neither detached-method site is newly accepted.
Built regression coverage for lexical receiver captures already implemented on compiler base 06877146; prerequisite merge 67f7e002, current-main merge ced06993 (efe9f404).
Focused lowering PASS (0.401s), uncached oracle PASS (0.627s), counts refresh PASS (24.453s); source Node exit 0 with nine lines.
Null receiver mutant caught by native UBSan and JavaScript TypeError; omitted receiver retain caught by ASan heap-use-after-free.
Uncovered: the 195 detached-method sites, dynamic function-expression receivers, receiver self-cycles, full TypeScript census, and unrelated kinds; no production compiler or runtime changes.

The ranking's name and the ruling's description differ. Every exact receiver row in RANKED.md says `a method read as a value (... would lose its object, and this with it)`. REMAINING.md is the separate explicit-any contract ledger. The two reductions reproduce that refusal through `adamic c`, exit 1, and the diagnostic recommends an arrow that keeps the object. These are reductions of the refusal kind, not full-project census replays or a claim that the historical source methods use this.

Assumption: implement the ruling's lexical arrow meaning and preserve exact JavaScript. Reading `object.method` does not bind object in JavaScript. Automatically capturing its receiver would change that meaning. Consequently no census-site acceptance reduction is claimed. TypeScript's worker should rerun its census and distinguish detached-method values from arrows using lexical this.

Existing implementation: `internal/lower/expression.go`'s ThisKeyword arm calls `touch(l.this)`. `internal/lower/locals.go` records the receiver in every intervening arrow's environment, marks it Captured, and shares its cell. The native parameter prologue in `emit_functions.go` makes an owning cell through `emit_locals.go`; ordinary environment capture owns that cell. JavaScript carries the same cell in its closure environment. The function-expression visitor refuses its own dynamic this, including an arrow inside that function, and resets this while lowering its body. None of these production files needed an edit.

The new receiver_captures.a fixture holds:

- A method's returned arrow reading and writing number and dynamically built string fields after both method and factory return.
- Two live, distinct receivers, called alternately, so substituting another receiver changes output.
- A callback stored in a separate holder's field and called twice after its factory returns.
- An arrow passed to Array.map, with mutations visible on its next invocation.
- Nested arrows with an outer callback reused to produce another inner callback.
- An inherited method's arrow and nested arrows on a subclass, including overridden method dispatch.

A callback stored on the same receiver can make an owning reference cycle. This unit's field holder is separate and cannot be reached by its callback's receiver. It does not relax the compiler's cycle refusal.

The null mutant replaces captured receiver reads in closure bodies with reads of a typed local initialized to undefined. Native C compiles with its normal strict flags and fails at runtime with `UndefinedBehaviorSanitizer: undefined-behavior`, member access within null pointer of type adamic_object, exit 1. The JavaScript backend reports `TypeError: Cannot read properties of undefined (reading 'count')`, exit 70. Source Node exits 0. The first development attempt used an inline undefined IR expression and produced invalid NULL->slots C; that compile failure is discarded and is not counted as a caught mutant.

The ownership mutant changes only generated C: the two receiver-cell initializers omit adamic_retain(receiver), leaving the closure holding freed instances. C compiles and ASan reports heap-use-after-free. Tests preserve production sources and require a runtime failure, not a build warning.

New lowering tests preserve the separate ordinary-function rule, both for direct this and for an arrow using that ordinary function's this. They also preserve the detached compare/toString refusals. Test registration is in the new receiver_captures_test.go, so internal/oracle/oracle_test.go is untouched.

Counts refresh adds exactly one row, in allocations/frees/retains/releases/peak/regions order: 76/76/75/133/32/0. Every existing row is unchanged.

Commands, with every test's output redirected to a log:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/receiver-setup.log 2>&1
bash cloud/setup.sh > /tmp/receiver-setup-retry.log 2>&1
source /workspace/adamic-tools/env.sh
go run ./cmd/adamic c cloud/receiver-captures/core-2100-compare.a > /tmp/receiver-core-2100-compare.log 2>&1
go run ./cmd/adamic c cloud/receiver-captures/debug-376-toString.a > /tmp/receiver-debug-376-toString.log 2>&1
go test ./internal/lower -run 'TestReceiverCapture|TestLibraryLanguageBoundaries/dynamic_receiver' -count=1 -v -timeout 10m > /tmp/receiver-lower.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestReceiverCapture|TestNativeAgreesWithNode/internal/oracle/testdata/receiver_captures' -count=1 -v -timeout 10m > /tmp/receiver-oracle-complete.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/receiver-counts.log 2>&1
node --input-type=module-typescript < internal/oracle/testdata/receiver_captures.a > /tmp/receiver-node.log 2>&1
git fetch origin main
git merge --no-edit origin/main
go test ./internal/lower -run 'TestReceiverCapture|TestLibraryLanguageBoundaries/dynamic_receiver' -count=1 -v -timeout 10m > /tmp/receiver-landing-lower.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestReceiverCapture|TestNativeAgreesWithNode/internal/oracle/testdata/receiver_captures|TestCountsAreRecorded' -count=1 -v -timeout 30m > /tmp/receiver-landing-oracle.log 2>&1
gofmt -l internal/lower/receiver_captures_test.go internal/oracle/receiver_captures_test.go
git diff --check
```

Initial setup failed with `l.nonNull undefined` and `l.checkedAssertionSource undefined` because its compiler build overlapped checkout/merge. Restarting on the stable branch succeeded: go ready 0.023s, Node ready 0.024s, submodules ready 0.064s, markdown dependencies ready 0.071s, clang ready 0.171s, go build ready 44.879s, test binaries deferred 44.986s, cache warm 44.988s, done 45.016s. nproc=5, cpu.max=400000 100000. Go 1.27.1, clang 20.1.8, Node 24.19.0. The printed env.sh was sourced for subsequent toolchain commands. An initial gofmt call before sourcing failed; the subsequent configured gofmt succeeded.

No whole-package or full gate was run. Focused tests cover the two packages changed by this unit. Current main's unrelated child-process guard and JSON test changes were merged without editing them. No other worker's internal/lower function was edited; no runtime C helper was added. No cohere code was copied. No prior skipped-kind list exists in this session, so unrelated kinds were not invented as extra territory.

Full command outputs are preserved beside this report in evidence/. Final branch SHA is supplied in the handoff, since this report is committed with the tests.

Post-main validation passed: lower 0.162s, uncached oracle including both mutants and complete recorded-count checks 23.928s. Formatting and diff checks printed nothing. Only the new counts row remains in the counts diff.
