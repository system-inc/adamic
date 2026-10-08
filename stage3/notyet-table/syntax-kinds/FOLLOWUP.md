Lowered primitive string substr through existing IR, including its length argument.
Base refreshed to d577dd0d; merge 694f14d3; implementation is the commit containing this report.
Focused lower/oracle tests passed; uncached fixture and seven mutants passed; counts refresh passed in 20.117s.
All seven mutants are caught by source-Node stdout comparison in native and JavaScript; no compiler/sanitizer failure is counted.
Five table sites belong to this kind; one example call now passes its statement, another remains behind Path. Other 21 kinds remain unresolved.

The expanded territory permission is understood: IR, backends, walkers and kind-specific lowering helpers are allowed. No runtime, IR or backend edits were needed for substr. The prior report's skips were prerequisite blocks rather than backend territory exclusions. Refreshing area/compiler to d577dd0d brings only non-compiler changes and still lacks the table compiler's checked-view, nested-function, predicate-proof and Node fs implementations.

Added stringSubstr in internal/lower/string_substr.go and admitted it through libraryString and libraryStringMethod. A generated ordinary IR function receives the string, start and length once in source order. It applies ToIntegerOrInfinity, normalizes a negative start relative to UTF-16 length, clamps length to a nonnegative count of remaining units, and slices from start to start + count. Undefined length defaults to the full length. Numeric and optional numeric arguments are supported; other representations retain named NotYet stops. No code was copied from cohere.

The fixture builds a string containing a supplementary character and lone surrogate, sweeps negative/fractional/NaN/infinite starts and lengths, checks omitted/undefined length and char codes, and records receiver/start/length side effects. Registered from its own syntax_substr_test.go. The ordinary oracle checks source Node, generated JS, release native, ASan/UBSan and leaks. Recorded row: 747 allocations, 747 frees, 138 retains, 869 releases, peak 10, regions 0. No existing count changed.

Every mutant keeps valid IR and finishes without leaks or sanitizer failures, but differs from source Node in both backends:

| Mutant | Wrong behavior | Catcher |
| --- | --- | --- |
| relative-start | reverses negative-start branch | stdout comparison |
| truncate | uses floor instead of truncation | stdout comparison |
| NaN | omits NaN normalization | stdout comparison |
| undefined-length | defaults missing length to zero | stdout comparison |
| length-not-end | treats count as slice end | stdout comparison |
| length-clamp | allows negative count | stdout comparison |
| evaluation-order | exchanges effectful start and length calls | stdout comparison |

The first length-clamp mutant survived because lengths had no negative integer inside the string's range. Adding -1 and -4.9 supplied the missing witness; all seven then failed for the intended output disagreement. An initial fixture included substr() and an undefined start, which the bundled TypeScript signature rejects. Removed those calls before final validation; the fixture now matches that signature.

Commands (all output redirected to logs):

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/adamic-syntax-resume-setup.log 2>&1
source /workspace/adamic-tools/env.sh
go test ./internal/lower ./internal/oracle -run '^TestLibraryStringRefusals$|^TestLibraryStringMutants$|^TestNativeAgreesWithNode/internal/oracle/testdata/syntax_substr.a$' -count=1 -timeout 5m > /tmp/adamic-syntax-substr-focus.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestSyntaxSubstrMutants$|^TestNativeAgreesWithNode/internal/oracle/testdata/syntax_substr.a$' -count=1 -v -timeout 5m > /tmp/adamic-syntax-substr-mutants.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/adamic-syntax-substr-counts.log 2>&1
```

Setup: Go 0.015s, Node 0.015s, submodules 0.060s, markdown 0.063s, clang 0.164s, build 32.710s, warm 32.806s, total 32.833s; nproc=5. Focused lower 0.376s, oracle 3.324s. Uncached fixture/mutants oracle 0.944s. Counts 20.117s. No whole package/full gate ran, following the unit instruction. git diff --check passed.

Replayed executeCommandLine.ts:348:30 before and after with the exact table reason. Both commands exit 1 because this base has a different initial reason. Before: the exact call stopped at structural method dispatch in a program with statics. After: that call's stop disappears. The unit still has independent an array of never, an array of any and reading res findings. The second example moduleNameResolver.ts:1266:20 remains behind a parameter of type Path at 1246:34. Neither replay outputs executable IR; the fixture establishes the semantics. Evidence preserves complete selected findings, with only declaration catalogs omitted.

Ownership was checked from fetched origin/codex/notyet-* histories. libraryString and libraryStringMethod have no other worker's unique changes. Array spread belongs to arrayLiteral, changed by object-small (43af2f34). Expression kind dispatch is enumNeverValue, changed by this-outside (b96e6356) and void-value; therefore ClassExpression and PostfixUnaryExpression cannot be admitted there by this worker. Property lowering belongs to object-property (5a870515). Existing prefix changes are confined to prefix, not the whole expression.go file.

Remaining per-kind status is skipped, with no design ruling inferred:

| Sites | Kind | Reason |
| ---: | --- | --- |
| 46 | checked write without source-slot certificate | checked-view write prerequisites absent |
| 25 | generic or unnamed nested function declaration | nested-function prerequisites absent |
| 10 | SpreadElement | arrayLiteral is being edited by object-small; kind dispatch also owned |
| 9 | ModuleDeclaration | namespace lowering already differs from measured base; no exact lesson established |
| 7 | overload argument with different implementation representation | measured nested/phantom overload paths absent |
| 5 | overloaded function read as value | measured phantom overload proof path absent |
| 4 | first-class nested reference from another nested function | nested-function prerequisites absent |
| 3 | dictionary checked view | checked-view prerequisites absent |
| 1 | ClassExpression | enumNeverValue is being edited by this-outside/void-value |
| 1 | PostfixUnaryExpression | enumNeverValue is being edited by this-outside/void-value |
| 1 | field of boolean or undefined | measured logical assignmentReference absent |
| 1 | optional comparator requiring recursive default conversion | view_array_consumers.go absent |
| 1 | uninitialized object field without supported declared slot | measured objectLiteral path differs; no exact reduction established |
| 1 | untyped length Array escaping without element proof | array-hole prerequisites absent |
| 1 | checked predicate overload target intersection | predicates_proof.go absent |
| 1 | generic overload result proofs | phantom_overload_results.go absent |
| 1 | node:fs.unwatchFile | Node fs prerequisites absent |
| 1 | node:fs.watch | Node fs prerequisites absent |
| 1 | node:fs.watchFile | Node fs prerequisites absent |
| 1 | statSync dynamic options | Node fs prerequisites absent |
| 1 | toLocaleTimeString | Node fs prerequisites absent |

These are uncompleted kinds, not claims that the expanded backend territory makes them impossible. Resuming the missing families requires resolving the measured compiler-lineage mismatch. There are no new C runtime helpers for review.
