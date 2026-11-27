# The family 1 swing (#x2cjvy9, rulings.md:258 as amended for .ts)

In .ts, a refinement by an open numeric enum's literal tag that the walk can't
prove is read through each narrowed object member's checked view of the
declared union (enums.go, narrowedCheckedViews), as an open-enum predicate's
true branch is (#k1t2wp9). .a refuses as before. Each lie stops at its own read
or write under the member's checked view: `m.n.a`, `x.a`, `wide.a` or
`<write>.a` (J-svh9wt1).

The mutants were run by hand, each on a job, and restored. They aren't
products, so this file records each one beside its catcher.

| Mutant | Catching test | What the run showed |
| --- | --- | --- |
| Skip: enumRefusal returns nil where it calls narrowedCheckedViews (Adamic's condition 4, as she took it on her accept) | TestEnumTagSwingLiesStopAtEveryRead, every lie row | J-svh9wt1 at 2394e645: all five lies red (native prints 0.9999981674500305 where Node prints NaN, and the write row trips "a field the checker proved is there is missing"), all five honest twins green |
| No narrows: register the views without `narrows` (the condition 4 Adamic first wrote) | none, by construction | J-k2nxabz at fdd72da6: all rows green. A view without narrows takes view_widened.go's unentered fallback, which only adds checks, and a narrowed field read takes that fallback anyway |
| Fallback: drop a narrows view from view_widened.go's unentered fallback | none, by construction | J-ja4ynd3 at fdd72da6: all rows green. The reads are checked at the member's own view before any widening, so no widened read is weighed |

The rows whose .ts refusal the swing lifts keep their mutants through
enumTagSwingSites, a test-only record of where the swing takes a .ts
refinement, at the site family 1's refusal names. A mutant that wrongly proves
a refinement leaves no site, and the row goes red. That holds the kind-set
FirstKindFit and UnionClaimFit catchers (enum-tag-kind-set.mutants.md) and the
type-only import's declaration-file mutant (TestClosedKindSourcesReadTypeOnlyImports).
