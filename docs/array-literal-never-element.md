# Empty array literals inside arrays

Built: fresh empty literals use their contextual or best-common array element type in the existing elementType path.
Base: codex/array-literal-never-element starts at origin/main 48c05d091f0a43c31cbe051b1d6578d99eeedf19.
Proof: four fixtures compare source Node, sanitized native, release native and JavaScript output, with leak and count checks.
Mutant: force the empty literal's element kind to Boolean; layout assertions fail and LeakSanitizer reports a 74-byte leak.
Limits: the node:buffer adapter is absent from this base; the full repository gate and complete host fixture are not proven here.

The original checker type of [] may stay never[] inside [[], [1]], despite
number[][] being the enclosing array's best-common type. elementType now
consults its contextual destination, its parent's contextual element type,
and the existing impliedTarget best-common destination. The regular arrayLiteral
operation still constructs the value; no backend or runtime path is added.
This reuses the contextual element path used for [] as T[] in the generic-return
unit codex/parser-generic-empty-array 4e022aaef7d24085fd0ded0c11b539843022fc47.
The number and object generic instantiations have an independent IR regression
check. No unrelated generic fallback or readonly-alias changes are imported.

For an empty literal in a union of arrays, every member accepts the empty value.
Choose one concrete member's supported element layout. A union of arrays is not
an array of union elements; the latter remains unsupported by the existing slot
rules. Mixed object/array contexts and tuple storage are not treated as arrays.
Only fresh empty literals get this specialization. Nonempty weak literals
continue using the contextual weak-handle representation.

Fixtures cover [[], [1]], [[1], []], a lone [] in number[][], and [] in
(number[] | string[])[], plus a direct union-array argument and an optional
array destination. Number, string and object fixtures append after construction
and read the elements. Strings are built at runtime so counted ownership matters.
The union fixture observes lengths, with no unsupported union-array indexing or
mutating-method semantics inferred from it.

The exact host reproducer is preserved in
cloud/reports/array-literal-never-element/host_reproducer.a. Node, run as an ES
module through stdin, exits 0 and prints a blank line followed by a. Adamic on
this base stops at its import with TS2591: Cannot find name 'node:buffer'. The
array lowering is independently proven; this does not claim the host adapter is
present or that the full host fixture compiles.

## Observations

The original object.go from 48c05d09 was tested through a Go overlay against
TestNestedEmptyArrayElementKinds. First, last, strings, objects and union all
fail with stage 0 can't lower an array of never yet; the already-contextual typed
case passes. The final layout/generic tests pass in 0.144s. The four original
unformatted fixtures pass the uncached oracle in 0.628s.

A production Go overlay changes only the successful empty-literal element result
from held to ir.Boolean. Every layout probe fails its assertion, including both
generic instantiations. The native mutant compiles successfully and the oracle's
source/backends behavior comparisons pass before its leak check fails:
LeakSanitizer detects 74 bytes leaked in one allocation after a fresh string is
pushed into the empty array. The wrong numeric kind marks that array's buffer as
not holding references, so destruction misses the appended string. This is not a
clang warning or lowering refusal. The overlay leaves repository files unchanged.

Counts update passes in 20.733s and adds only four rows, with no existing row
changed. Allocations/frees/retains/releases/peak/regions:

| Fixture | Counts |
|---|---|
| first | 25/25/17/39/11/0 |
| last | 28/28/21/43/14/0 |
| typed | 17/17/13/31/10/0 |
| union | 22/22/10/29/13/0 |

Setup's first run overlapped file edits and failed with
internal/lower/object.go:269:20: l.literalArrayElement undefined. Retrying after
all files were present passed: Go ready 0.024s, Node 0.027s, submodules 0.056s,
markdown dependencies 0.080s, clang 0.174s, build 25.975s, cache warm 26.125s,
done 26.153s. nproc is 5 and cpu.max is 400000 100000, a four-CPU quota.
Go 1.27.1, clang 20.1.8 and Node 24.19.0; environment sourced from
/workspace/adamic-tools/env.sh in every build/test shell.

All test output is redirected to logs. Commands:

```sh
go test ./internal/lower -run 'TestNestedEmptyArrayElementKinds|TestEmptyLiteralGenericReturnUsesSamePath' -count=1 -v
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/array_literal_empty_' -count=1 -v -timeout 30m
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -count=1 -timeout 30m
go test -overlay=/tmp/empty-literal-wrong-kind-overlay.json ./internal/lower -run 'TestNestedEmptyArrayElementKinds|TestEmptyLiteralGenericReturnUsesSamePath' -count=1 -v
ADAMIC_GATE_UNCACHED=1 go test -overlay=/tmp/empty-literal-wrong-kind-overlay.json ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/array_literal_empty_first' -count=1 -v -timeout 30m
gofmt -l cmd internal
go vet ./...
```

From the cohere submodule, the pinned checker is invoked through its original
source, with no copied code. The four new .a files pass:

```sh
go run ./command/cohere --directory .. --no-fix internal/oracle/testdata/array_literal_empty_first.a internal/oracle/testdata/array_literal_empty_last.a internal/oracle/testdata/array_literal_empty_typed.a internal/oracle/testdata/array_literal_empty_union.a
```

Its first run reported only four format findings. The format-only command
applied them, then the no-fix run passed with 276 rules and four files checked.
Formatting and vet logs have no diagnostics. The final formatted fixtures and
wrong-kind runtime mutant were rerun. Logs are retained under
cloud/reports/array-literal-never-element.

Final complete touched-package gate exits 0: lowering 24.851s and uncached
oracle 198.044s. The formatted-fixture oracle passes in 0.607s. The repeated
runtime mutant fails in 0.497s with the same 74-byte LeakSanitizer report.
The complete oracle includes its recorded-count comparison. No native emitter,
runtime, compiler orchestration or shared oracle harness file was changed.
