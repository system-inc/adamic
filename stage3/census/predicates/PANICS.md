Fixed: qualified cast names, computed relation fields, and generic validation context cleanup.
Commits: the fix commit contains this report; the following census commit records its measured SHA.
Commands/results: affected package tests and vet pass; sanitized qualified cast prints `one`, matching Node.
Mutants: four runs restore the cast guard, widening guard, computed-field guard and early cleanup; each regression catches the original panic.
Limits: generic mutation and computed optional-field contracts remain named refusals; the full gate is not claimed green.

# Programs, panics and fixes

## Qualified cast name

The census refusal scan in `builder.ts` stopped in
`toBuilderStateFileInfoForMultiEmit`, at `fileInfo as BuilderState.FileInfo`.
`castProof` treated every TypeReference name as an identifier while testing for
`as const`. A qualified name has no `Node.Text()` implementation and panicked:
`Unhandled case in Node.Text: *ast.QualifiedName`.

The reduced program is `internal/lower/testdata/census_panics/qualified.a`, with
its type-only `types.a` dependency. Both cast admission and mutable-view admission
now require an identifier before comparing it with `const`. Qualified casts reach
the existing relation proof. This compatible upcast is erased. Its original source
on Node and native with ASan/UBSan print `one\n`, byte for byte.
`TestCensusQualifiedCast` independently pins both admission paths. The original
builder cast has unproven fields and now returns the named refusal `a cast the
runtime can't check` at `builder.ts:2233:9`, rather than panicking.

## Computed relation field

The census refusal scan in `transformers/declarations/diagnostics.ts` stopped in
`createGetIsolatedDeclarationErrors`. Its diagnostic table has computed enum keys
and a `satisfies Partial<Record<SyntaxKind, DiagnosticMessage>>` view.
`relationFieldExpression` tried `property.Name().Text()` on a computed key and
panicked: `Unhandled case in Node.Text: *ast.ComputedPropertyName`.

The reduced program is `internal/lower/testdata/census_panics/computed_relation.a`.
Computed keys now remain structural views when proving nested optional fields;
syntax alone does not certify which runtime field was initialized. An unproven
field returns the existing named diagnostic instead of calling `Text()` on the
computed name. The regression pins `optional field arrow.suggestion has no proven
compatible presence/type`. Node prints `return type\n`. No new key-evaluation or
field-contract implementation is claimed.

## Generic validation inside a closure

The census lowering attempt at `emitter.ts:685:1`,
`getCommonSourceDirectoryOfConfig`, passes an arrow containing a `filter` call.
Instantiating `filter` reaches `core.ts:288:33`, `result.push(item)`, whose
instantiated write is refused. `instantiateFunction` had cleared the caller's
closure stack and changed its substitution context before validation, but installed
its restoration defer only after validation. The refusal returned with an empty
stack; `closure` then popped it and panicked:
`runtime error: slice bounds out of range [:-1]`.

Restoration is now installed immediately after the context changes, before every
validation exit. Depth cleanup remains paired with its later increment. The
original census unit now reports `instantiating a generic function makes a value
of type string | undefined written where string is read`, at `core.ts:288:33`.
The reduced checker-valid witness `generic_closure.a` triggers the same write
validation from an arrow, without relying on the census's rejected bodies. Node
prints `1\n`; `TestCensusGenericRefusalInsideClosure` pins the named refusal.

# Verification

Outputs are retained under `evidence/panics/`. Setup in the repository completed
in 41.084 seconds, including 40.784 seconds warming the Go build cache. `nproc`
reported 5; the cgroup permits four CPUs. Go 1.27.1, Node 24.19.0 and clang 20.1.8
were used. The scratch setup encountered its intentional cohere symlink; rerunning
from the repository succeeded. The commands were:

```sh
go test ./internal/lower ./cmd/adamic ./internal/ir ./internal/load -count=1
go vet ./internal/lower ./cmd/adamic ./internal/ir ./internal/load
go test ./internal/lower -run TestCensus -count=1
node --input-type=module-typescript < internal/lower/testdata/census_panics/qualified.a
go run ./cmd/adamic build internal/lower/testdata/census_panics/qualified.a -o /tmp/panic-qualified-native --sanitize
```

The other two witnesses were also run from original source on Node. Each mutant
restores one actual fault: the qualified-name identifier guard is removed; the
computed-name guard is removed; early generic context restoration is removed.
The duplicate mutable-view guard was removed in a fourth mutant run too.
Every corresponding regression fails with the original panic, rather than a build
or clang error. Mutants were restored before the commit. The qualified-name test
also directly exercises the duplicate guard in `refuseWidening`.

Scratch stack tracing was used only to investigate the original panics. It is not
part of production lowering or the official final measurement overlay. Main was
fetched and was already an ancestor (`48c05d09`). Only `codex/proven-predicates` is
pushed. Existing readiness oracle failures are outside these fixes.
