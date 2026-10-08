# Unit 2 report — paused at parser recovery

Unit 2 is unfinished and has not been pushed. Static-components is not registered
or certified: **0 upstream rule cases compared**. No green rule claim is made.

## Last passing construction certificate

**Native = Node: 752 / 1,465 original corpus functions**, plus **57 / 57 path
probes**, 809 / 1,522 records overall. This adds four original corpus functions
and one probe to the shared-index checkpoint. The 23 Flow graphs remain excluded;
the original denominator and all 45 classified Go skips remain intact.

The passing command was:

```
HIR_CENSUS_EXPORT=/tmp/hir-templates go test -count=1 -v -timeout=25m ./stage1/cohere/high_level_intermediate_representation -run 'TestWholeConstructionCensus/catches_(regex_flags_disappear|template_operand_disappears|tagged_template_changes_tag|this_changes_global_name)'
```

`/tmp/hir-templates.log` passed in 95.43 seconds. Both direct and cached
ForFunction construction match every admitted row. New lowering covers template
literals, tagged templates, regular expressions, `this`, and non-null erasure.
Four new semantic mutants compile, execute and disagree on both runtimes: dropped
regex flags, dropped template operand, changed tag operand, and changed `this`
global name. The checked FunctionIndex off-by-one mutant stops on both. Combined
with the 56 previously certified semantic mutants, 60 semantic mutants have
passing certificates; the entire combined matrix has not yet been rerun.

## Current unverified extension and first failure

Casts and deletes were added with three further mutants; await's instruction
shape is present, but async function admission remains unfinished. These changes
are **not certified**: the latest full comparison stops before native coverage
and before their mutants run. Its 822 admitted records are not a coverage claim.

Carrying Go's source ScriptKind into replay exposes a parser contract difference.
Five previously declined `ranges_test.go` graphs contain JSX text in a `.ts`
fixture. Go recovers that text as type assertions, object expressions and a regex;
replaying it as TSX instead yields JSX and different identifiers. The manifest now
preserves TS/TSX/JS/JSX mode. Independent TS parsing stops on a missing semicolon.
No row was excluded or reclassified to hide this failure.

The shortest reproducer checked on both runtimes is
[testdata/parser-recovery-gap.a](testdata/parser-recovery-gap.a). Its diagnostic is
`adamic: panic: parser slice expected semicolon at 26 in input.ts`.
This is a Stage 1 parser behavior gap, **not an Adamic language/compiler gap**.
The compiler successfully builds the reproducer.

Automatic approval review rejected the proposed missing-token recovery change
before execution because it interpreted that change as conflicting with the
user's stop-on-gap instruction. The shared parser was not changed. Approval is
required to extend its recovery contract. The proposed behavior is to preserve
Go's missing-token AST recovery in HIR parsing, retaining strict parsing by
default, with exact oracle comparisons and recovery witnesses before landing.

## Allocation delta

The last measured 12-function native probe remains **3,212 -> 3,869 allocations
(+657)**; frees equal allocations. Retains rise by 2,168, releases by 1,905, and
peak live objects by 24 (175 -> 199); regions remain zero.

Known contributors are the boxed concrete index objects, canonical handle arrays,
and per-function SSA adapter closures. The latter close over each HIRFunction so
the imported SSA callback, which does not receive a function argument, can recover
its canonical IdentifierIndex. These are named causes of the conversion's growth;
the measurement does not isolate each contributor. No new allocation measurement
is claimed for the unverified cast/delete extension.

After approval: finish parser recovery certification, then remaining function,
method, binding/pattern and optional-chain variants; complete rule ownership and
compilation-unit integration; register static-components in its own directory
under docs/lint-registration.md; run the complete construction and rule corpora,
all mutants and counted native build; push once to stage1-hir/wip.
