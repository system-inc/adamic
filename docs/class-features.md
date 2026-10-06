# Class features

This is the first step of task zek5q21, on codex/class-features, based on inheritance tip 7b30af6. **The full task remains open: static fields, methods, accessors and blocks are not implemented.** Static declarations now produce a diagnostic even if the class is never instantiated, so initialization effects cannot silently disappear.

## Implemented

Instance getters and setters use per-class descriptors, searched from the dynamic class toward its base. Overrides replace an entire descriptor, as JavaScript does. Super property reads and writes call the base accessor with the current receiver. Property reads and writes become IR calls, including through structural views, so borrowing and exception analyses see their targets. Generated setter fallback writes preserve the original write-site metadata.

Literal accessors are closures with an explicit receiver, stored privately in their object. They support this, lexical captures, enumeration and spread. Object.keys enumerates public own properties, with numeric index names first in ascending order. Spread reads enumerable accessors in that order and produces data properties. It drops private storage and accessor closures, preventing plain-object reuse of those layouts. Its receiver remains owned across callbacks, and freshness conservatively accounts for callback escape.

Private fields and methods are qualified by their declaring class. Base and derived classes can use the same private spelling without sharing storage or dispatch. Enumeration and spread exclude private fields. TypeScript rejects private access outside the declaring scope. The fixture includes an inherited private Weak parent reference.

Readonly enforcement remains TypeScript's checker rule, not runtime freezing. Writes outside the declaring constructor are rejected, including ordinary methods and derived constructors. Repeated assignments in the declaring constructor remain valid. Objects held by readonly fields can still have mutable contents.

Override checks reject narrower setter parameters, unsafe getter results and overrides masking an inherited descriptor half, with a repair in the diagnostic. A narrowed getter reread is rejected when it can return a different value: read it once into a local and narrow that local. The cycle finder follows literal accessor capture cells and excludes computed results from physical fields.

Cohere's CSS values parser now uses private identifiers for its methods. Its private-method gap is marked closed and exercised against Node, sanitizers and the leak check.

## Current limits

- Static members and blocks remain unsupported. Constructor objects need a separate inheritance and initialization model, including inherited static data and exact private brands.
- Optional accessor reads, accessor literals containing spreads, computed names and ordinary methods inside accessor literals report NotYet.
- Accessors sharing a name need compatible native representations; this conservatively excludes some generic instantiations and overrides.
- Reading setter-only descriptors and writing getter-only descriptors report NotYet. Conservative dispatch can also reject unrelated same-name members.
- Spreading literal getters that can throw or fail through a runtime library operation is diagnosed until callbacks have explicit exception edges. Setter-only literal spread is diagnosed too.
- Getters without an adamic_value representation and setters requiring two native words are diagnosed. Literal property names beginning #accessor: are reserved for closure storage and diagnosed.

## Verification

Setup succeeded with Go 1.27.1, clang 20.1.8 and Node 24.19.0. Timing lines: go 0s, clang 0s, node 0s, submodules 0s, cache 13s, total 13s. nproc printed 5. The environment file is /workspace/adamic-tools/env.sh.

The fixtures class_features_private.a and class_features_accessors.a compare source Node, generated JavaScript, native release and ASan/UBSan, and require a clean leak check. They cover overrides and super, side effects, captures, literal this, numeric enumeration, spread, exceptions, receiver replacement during compound assignment, private method separation, private enumeration, readonly constructor assignments and a private Weak back-reference.

The initial full gate passed the oracle suite in 854.197s and native tests in 290.866s. It found missing generated write-site metadata and the newly closed cohere gap. Both were repaired and their tests rerun. Focused ownership verification passed: lower 1.409s, oracle 6.693s. Final package verification passed: lower 12.286s, native 148.021s, fresh 39.550s and cohere values 126.530s. The latest focused oracle passed in 6.328s. Vet, formatting and git diff checks were clean. A combined invocation mistakenly passed the oracle counts flag to other packages; those binaries rejected the flag. The corrected package command above passed. Its oracle portion separately passed fixture verification and the complete counts update in 208.722s.

Reproduce with output in logs:

    source /workspace/adamic-tools/env.sh
    bash cloud/setup.sh > /tmp/adamic-class-features-setup.log 2>&1
    go vet ./... > /tmp/adamic-class-features-vet.log 2>&1
    gofmt -l cmd internal > /tmp/adamic-class-features-format.log 2>&1
    go test -count=1 -timeout 30m ./internal/lower ./internal/native ./internal/fresh ./stage1/cohere/values > /tmp/adamic-class-features-core-packages.log 2>&1
    go test -count=1 ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/class_features' > /tmp/adamic-class-features-ownership.log 2>&1
    go test -count=1 -timeout 30m ./internal/oracle -run TestCountsAreRecorded -args -update-counts > /tmp/adamic-class-features-counts-final.log 2>&1

Compare counts against the 180 rows at the inheritance tip: allocations, frees, retains, releases and peak must never increase; region counts must never decrease. The final comparisons printed: Compared 180 pre-branch rows; regressions: [] and Compared 173 pre-branch rows; regressions: []. The two new rows are below; columns are allocations, frees, retains, releases, peak and values in regions.

| Fixture | Allocations | Frees | Retains | Releases | Peak | In regions |
|---|---:|---:|---:|---:|---:|---:|
| class_features_private.a | 47 | 47 | 37 | 65 | 15 | 0 |
| class_features_accessors.a | 67 | 67 | 53 | 110 | 21 | 0 |

## Mutants

Each mutation changed real lowering or runtime code, was run, then restored. These failures did not depend on compiler warnings.

| Mutant | What caught it |
|---|---|
| Getter replaced by raw slot read | Node/native stdout: 2.5 instead of 6, getter side effects missing. A reduced fixture kept both native executions at exit 0; release and sanitized output also differed for the misread closure slot. |
| Setter skipped | Node/native stdout: 2 instead of 6, setter trace and setter exception missing. |
| Private fields in public shape | Node/native stdout: qualified private names appeared in enumeration and spread. |
| Derived private method redirected to base | Both backends: privateone/privateone instead of privateone/child1. |
| Readonly diagnostic TS2540 discarded | TestClassFeaturesReadonlyChecker accepted an illegal write and failed. The loader mutation was temporary. |
| Setter contravariance check skipped | TestClassFeaturesAccessorRefusals accepted a narrower setter and failed. |
| Capture-cell edges omitted | TestClassFeaturesAccessorCaptureCycle accepted a reference-count cycle and failed. |
| Narrowed getter check skipped | TestClassFeaturesNarrowedAccessor accepted a changing getter reread and failed. |
| Static diagnostic skipped | TestClassFeaturesStaticDeclarationsAreDiagnosed accepted dropped initialization and failed. This proves the diagnostic, not static support. |
| Setter write-site set to zero | TestEveryWriteIsRecordedAndKnown found unrecorded writes and failed. |
| Numeric keys in insertion order | Node/native stdout: 10,2,label instead of 2,10,label. |

## Files outside the original territory

The original territory is new files under internal/lower and internal/native, plus internal/lower/class.go. Additional files:

| File | Why |
|---|---|
| internal/ir/ir.go | Carry visibility, descriptors, accessor calls and enumeration. |
| internal/javascript/javascript.go | Match dispatch, hidden storage and literal enumeration. |
| internal/fresh/fresh.go | Visit new expressions and account for spread callback escape. |
| internal/lower/lower.go | Small preparation/finalization hooks and static diagnostic. |
| internal/lower/object.go | Literal/property hooks, narrowing, enumeration and spread. |
| internal/lower/refusals.go | Remove the blanket accessor refusal. |
| internal/lower/lower_test.go | Remove its obsolete getter refusal expectation. |
| internal/lower/cycles.go | Follow captures and distinguish physical fields. |
| internal/native/emit.go | Small shape/call/exception hooks; own spread receivers across callbacks. |
| internal/native/reuse.go | Prevent reuse when spread drops private storage or closures. |
| internal/native/runtime/adamic.h | Declare descriptors and helpers. |
| internal/native/runtime/object.c | Copy public fields and materialize accessor results in order. |
| internal/oracle/class_inheritance_test.go | Register differential fixtures. |
| internal/oracle/testdata/class_features_private.a | Hold private/readonly behavior and ownership to Node. |
| internal/oracle/testdata/class_features_accessors.a | Hold accessor behavior and ownership to Node. |
| internal/oracle/counts.md | Record measured memory costs. |
| stage1/cohere/values/gaps_test.go | Mark the private-method gap closed and exercise it. |
| stage1/cohere/values/GAPS.md | Document closure and removed workaround. |
| stage1/cohere/values/parser.ts | Replace private-method workaround and its calls with private identifiers. |
| docs/class-features.md | Publish implementation, verification and remaining scope. |
