Built optional ordinary-array indexing; the original 91-site reason remains unlowered.
Implementation 63f7ae5bd676e305fb8ce068c8b4936145aa6172; replay merge 9d8f6a44cf576a2c90c5a121cc0e6784431c59cf.
The focused Node/backend/release/sanitizer oracle passed; counts refresh passed with one new row.
Three valid mutants failed: lost optional flag, eager key, and repeated receiver; an earlier build-only mutant is excluded.
Covered shapes: 7 optional-array roots; 0 of the original 91 roots. Original binder replays retain a NonNullExpression stop.

## Base and observations

The explicit unit base overrides the generic main-base instruction. Resolved
origin/area/compiler to b410340dc8f889b5799c3bc519117c63def3aa24 and created
codex/notyet-element-access there. Merged the newest census-replay tip,
9a1f14c5d994aa855625e7cfa295677060348fec, and pushed the merge. The table tip
was 57b9777c8eb4ee28b1f50220e8c8fb51a2dfadf7. Its compiler base is 44583d32
under a different non-null fallback. Neither that base nor a non-null
implementation was substituted for the requested compiler base.

`bash stage3/apply.sh /tmp/element-access-adapted` exited 0. All 81 adapted
source hashes match the table's source-manifest.json. Thus the replay mismatch
below is not attributable to different adapted source bytes.

The exact requested original signatures do not reproduce: both
binder.ts:1746:21 and binder.ts:1757:28 stop at `a NonNullExpression` before
entering elementAccess. Both original commands exit 1. Replaying those same
positions with the observed reason after the optional-array change exits 0.
They retain the same stop. Production's up-front refusal pass also refuses
non-null assertions. No non-null implementation or permissive cast was added.

## CSV classification and limits

[sites.csv](sites.csv) records every independent root, selected from the table's
roots/raw.csv by exact reason and empty blocked_by_reason, deduplicated by
position. Stock TypeScript 6.0.3 supplied receiver, key, and result types over
the complete adapted tsc entry project. This is independent checker evidence,
not a claim that Adamic successfully represents each receiver.

The original 91 sites comprise 62 NodeArray receivers; 10 other array-derived
receivers (6 Readonly<PathPathComponents>, 2 TemplateStringsArray,
1 SortedReadonlyArray and 1 JSDocArray); 1 union of tuples; 15 objects with
finite keys; and 3 receivers with index signatures (Record<number, ...>,
MapLike<string>, CompilerOptions). The array-derived and tuple sites use
numeric keys. There are no string receivers or Map-entry bracket operations
in this reason's CSV roots. In particular, both binder receivers are
NodeArray<CaseOrDefaultClause>, with number keys and possibly undefined results.

These 91 roots remain uncovered. The finite named-field cases are implementation
work, not a language refusal. Dynamic tuple indexing and structural array
representations also remain implementation work; silently treating an ir.Object
as a native array would be wrong. NodeArray additionally carries metadata fields,
so representing it as a plain array would lose observable state.

A ruling or prerequisite is needed before the binder sites can advance on this
base: supply the approved checked non-null implementation, or authorize the
measurement fallback used by the table. Open dictionary acceptance requires the
index-signature/Record ruling in docs/0.1.md to be superseded with its semantics.
Array metadata and construction need an agreed representation that preserves
all own fields and the declared array behavior. No refusal was weakened.

## Optional ordinary arrays

The additional 7 roots all use ordinary arrays that may be undefined; their
contents are reference values, sometimes also undefined. The rule therefore
matches all 7 receiver/key shapes. This count is shape coverage from the CSV,
not seven successful whole-unit replays.

ArrayIndex now records Optional. Native snapshots the receiver, emits key work
and key-temporary cleanup inside the present branch, and leaves a missing slot
otherwise. Existing element presence and ownership handling then applies.
JavaScript emits `?.[]`, preserving the same evaluation order and short circuit.
Only elementAccess was edited in object.go. No unrelated lowering function or
prohibited file was edited.

The fixture covers missing receivers, present and absent elements, negative,
fractional and NaN indexes, built strings, optional object elements, and a key
that removes the last external array owner. Its receiver and key counts also
prove single evaluation and skipped work. It is registered in its own test file.

Before the change, checker.ts:9101:112 reproduces `?.[] on a value`, exit 0.
Afterward its optional expression lowers; the next failed statement is
checker.ts:9102:17, `a BinaryExpression with a boolean and a value`.
checker.ts:7440:67 is blocked by `reading labeledElementDeclarations` both before
and after, following an earlier non-null stop. That replay never reaches the
optional-index rule. The other five post-change replays contain no failure at
the selected optional-access position. [replays.json](replays.json) preserves
compact findings and command messages for the requested examples.

## Commands, tests, and mutants

Every test wrote directly to a log, without a pipe. The shell sourced
/workspace/adamic-tools/env.sh. Commands actually run:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/element-access-setup.log 2>&1
bash stage3/apply.sh /tmp/element-access-adapted > /tmp/element-access-apply.log 2>&1
go run ./stage3/census/latent/replay -project /tmp/element-access-adapted/src/tsc/tsc.ts -where /tmp/element-access-adapted/src/compiler/binder.ts:1746:21 -kind NotYet -reason 'an ElementAccessExpression' > /tmp/element-access-before-1746.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/element_access_optional' -count=1 -timeout 10m > /tmp/element-access-optional-final.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 10m -args -update-counts > /tmp/element-access-counts.log 2>&1
go vet ./internal/lower ./internal/native ./internal/javascript ./internal/ir > /tmp/element-access-vet.log 2>&1
git diff --check > /tmp/element-access-diff-check.log 2>&1
```

The replay command was also run for binder.ts:1757:28 and both optional examples,
using their exact reasons, then all seven optional sites after the change.
The original two post-change commands used `a NonNullExpression` and reproduced
that reason. Replay exit 1 after removing a signature is expected; findings,
not its exit alone, establish advancement.

The initial oracle passed in 12.593s; the restored final oracle passed in 0.575s,
with ADAMIC_GATE_UNCACHED=1. Source Node, JavaScript backend, release native,
ASan, UBSan, and leak checks all participate in this harness. Counts refresh
passed in 29.317s. The only counts diff is the new row: allocations/frees 35/35,
retains/releases 22/51, peak 9, regions 0. Focused vet and whitespace checks
produced no diagnostics. No whole-package test or full gate was run.

All mutants used that same focused uncached oracle command, each restored before
the next. Logs are /tmp/element-access-mutant-NAME.log:

| Mutant | Result and catcher |
| --- | --- |
| lowering-optional-flag: emit Optional false | Exit 1; UBSan null array access, backend TypeError, and Node disagreement |
| native-eager-key: emit key work before the present branch | Exit 1; native prints key:2 and 2 2 where Node prints no second key and 2 1 |
| native-repeat-receiver-valid: repeat receiver inside a correctly scoped branch | Exit 1; native prints 3 1 where Node prints 2 1 |

The first native-repeat-receiver attempt had undeclared C temporaries and was
caught by compilation. It is explicitly not a valid semantic mutant and is not
included in the three caught mutants above.

Setup exited 0, with nproc=5 and cpu.max=400000 100000. Timing lines: Node ready
0.068s; Go ready 0.076s; clang ready 0.576s; markdown installation step 1.530s
and ready 1.718s; submodules ready 24.672s; Go build ready 264.070s; test binaries
deferred 264.178s; build cache warm 264.180s; done 264.221s. No workaround was
needed beyond the requested GOPROXY setting.
