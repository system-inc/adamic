# Qualified predicate call mutants (#3hxh5jm)

These back the qualified predicate admission and call-check fixtures in
`internal/oracle/testdata/predicate_qualified`, and the live false-direction
fixture held by `TestQualifiedPredicateFalseSideReadLowers`.

Prepare the runtime products, then run the overlays against the same checkout
and product cache. No production source is edited by this tool:

```sh
go test ./internal/oracle -count=1 -parallel 1 -run '^TestProduct_NativeRuntimes(Sanitized|Release|Counted)[0-2]$'
python3 internal/lower/testdata/predicate_qualified_mutants.py --output /tmp/qualified-predicate-mutants
```

The run used a clean environment, a fresh HOME, GOMAXPROCS=8, and serial Go
package builds. Both mutated Go test commands exited 1. The restored qualified
oracle units and false-direction fixture passed.

## Bypass the qualified call check

In `internal/lower/expression.go`, make `callFunction` return its raw `ir.Call`
for a property-access callee, before its existing `censusOverloadResult` call.
The tool pins the exact original and replacement text. Each of these goes red:

- `TestQualifiedPredicateNamespaceAssertions`
- `TestQualifiedPredicateReexportAssertions`
- `TestQualifiedPredicateModuleAssertions`

All three lying fixtures lower, but both backends now execute the unchecked
source answer. Each test independently reports both failures; captured lines:

```text
JavaScript qualified lie did not stop at the asserts call check: exit 0 stdout "called\nafter\n" stderr ""
ASan/UBSan native qualified lie did not stop at the asserts call check: exit 0 stdout "called\nafter\n" stderr ""
```

## Make the false-direction walk identifier-only

In `internal/lower/predicate_false_flow.go`, replace valueCall's shared helper
condition with `ast.IsIdentifier(callee) &&` that condition. Leave admission,
check insertion, and the escape switch unchanged.

`TestQualifiedPredicateFalseSideReadLowers` goes red. Its explicit escaped
predicate keeps the walk live. The captured diagnostic, with the temporary
source directory shortened, is:

```text
main.ts:7:56: stage 0 can't lower a read of n narrowed on the false side of ts.isX(n) at main.ts:7:33, a call through a predicate value whose false answer isn't checked yet
```

This is the walk's NotYet at the field read, not a checker failure or a build
failure. The ordinary fixture lowers and agrees with Node on both backends.
