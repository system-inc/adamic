Validated fixed optional-boolean field indexing; no additional lowering function changed.
Checked non-null c41c0e06 remains merged at c7095bc7; this step continues from f2873010.
Focused Node/native release/native sanitizers/JavaScript oracle passed; counts refreshed.
Missing indexed absence mutant failed: native exit 70 where Node returns undefined and exits 0.
Remaining: all 72 derived-array roots require owned functions/representation decisions; two open dictionaries remain refused.

Receiver shapes were examined largest first. Ownership was refreshed with
`git fetch origin '+refs/heads/codex/notyet-*:refs/remotes/origin/codex/notyet-*'`.
`git log origin/codex/notyet-representations -- internal/lower/expression.go`
shows representation work through 95a72c07 and the c41c0e06 merge at 3cead3fd.
The elementType owner is origin/codex/notyet-element-type, including 024f7856.
Neither function was edited. No IR/backend/runtime change was made.

| Shape | Raw CSV roots | Required work / reason to move on |
| --- | ---: | --- |
| NodeArray | 62 | representation must preserve indexed storage and pos/end/hasTrailingComma/transformFlags across MutableNodeArray construction and readonly views; then elementType must recover generic base element storage. |
| Readonly<PathPathComponents> | 6 | representation/objectIntersection must establish a plain-array view for the intersection with __pathComponensBrand: void. The field is void, not a proven absent undefined field: decide how brand presence and writes behave before treating it as phantom. elementType must unwrap that established base. |
| TemplateStringsArray | 2 | representation must preserve raw array metadata and identity alongside numeric string slots; elementType must recognize the readonly base. |
| SortedReadonlyArray | 1 | representation must decide the void-valued " __sortedArrayBrand" contract for the array interface; elementType must recover MappedPosition storage. |
| JSDocArray | 1 | representation must preserve mutable optional jsDocCache alongside indexed JSDoc storage; elementType must recover the Array base. |

The exact declarations inspected are adapted compiler/types.ts:1589 (NodeArray),
path.ts:457 (PathPathComponents), lib/es5.d.ts:603 (TemplateStringsArray),
corePublic.ts:17 (SortedReadonlyArray) and types.ts:963 (JSDocArray).
These are implementation blockers, not language refusals. Building just the read
by reinterpreting Object as Array would use an incompatible layout. A representation
and conversion contract must come from the function owner; the unit then resumes
at elementAccess. No derived shape is claimed implemented or Node-validated.

Both binder examples were replayed with checked non-null merged. Each original
ElementAccessExpression signature reproduces (exit 0), at 1746:21 and 1757:28.
The subsequent reading-clause findings are also recorded in shape-replays.json.

The three prior dictionary classifications were checked separately. CompilerOptions
at utilities.ts:9370:12 uses StrictOptionName, a union of eight names of explicitly
declared optional boolean fields. It is covered by the existing finite-field rule,
without accepting open keys or index-signature source syntax. Its selected replay
has no lowering findings (exit 1 means the old signature did not reproduce).
The new own registration element_access_boolean_test.go runs a reduced fixed-shape
.a fixture through Node, native release, ASan/UBSan and JavaScript Node. It checks
all eight names as absent, true and false, strict fallback, and a different present
field next to an absent field. The assumed scope is fixed declared fields only.

Record<number, FlowGraphNode> at debug.ts:996:29 and MapLike<string> at
utilities.ts:8605:43 remain open dictionaries. docs/0.1.md:92 refuses index signatures
and Record<string,T>; refusals.go retains the index-signature refusal. A ruling is
needed on dynamic own keys, missing-key undefined, numeric conversion, prototype
names, integer ordering and shape growth before either is implemented.

Raw CSV classification is now 72 derived arrays, 16 finite fields, one tuple union,
two open dictionaries: 17/91 rule shapes covered. This adds one validated root shape
to the prior 16, not a claim of complete stage3 program compilation.
Run `python3 stage3/notyet-element-access/coverage.py` to reproduce the count.

Commands (output redirected to /tmp/element-access-boolean-*.log):

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run TestNativeAgreesWithNode/internal/oracle/testdata/element_access_boolean -v -count=1 -timeout 10m
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 10m -args -update-counts
```

Oracle: ok 0.527s, three uncached native misses and two Node misses; counts: ok
22.911s. Only the new counts row changed: 38 allocations, 38 frees, 17 retains,
50 releases, peak 7, regions 0. The mutant replaces the helper's optional-field
Absent flag with false; go test exits 1 on a semantic exit-code mismatch (native
panic 70 versus Node 0), not a Go/clang build failure. Full log and command are
in boolean-mutant.json. Source was restored and the final oracle passed.
The first mutant invocation omitted the toolchain environment and failed before
running Go tests with "Go: Unknown option: test"; it was not counted. The sourced
rerun is the recorded killed mutant. No whole-package test or full gate ran.
