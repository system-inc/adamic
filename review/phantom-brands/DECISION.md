# Phantom brands: language decision

Historical review, superseded by the approved refinement below. Implementation originally stopped pending this language decision.

Observed with Node 24.19.0, running the exact source using `node --input-type=module-typescript`:

* `undefined-read.a`: stdout is empty; exit is 1; stderr contains `TypeError: Cannot read properties of undefined (reading '__escapedIdentifier')`.
* `prototype-read.a`: stdout is `object\n`; exit is 0.
* The same escaped-identifier member read on a string brand prints `undefined\n`, exit 0.

The first program shows that representing the void arm as undefined is consistent with Node, but replacing its member read with undefined is not. The second shows that void-only members alone are insufficient to establish the promised runtime type: JavaScript can resolve a brand name through the primitive's prototype.

Choice needed: preserve the TypeError for reads through undefined, and restrict phantom members to names proven absent from the primitive's property and prototype chain, or explicitly approve a semantic exception. My judgment is to preserve Node and refuse conflicting member names. Optional access on undefined can return undefined; ordinary access cannot.

These probes are original minimal counterexamples to the ruling, not fixtures cut from tsc. They live under review because they are deliberately unsound. No claim is made that Adamic accepts or executes them.

Located representation refusal: `internal/lower/expression.go`, `representation`, routes intersections through `objectIntersection` after handling WeakBrand. Named values then produce NotYet diagnostics in `typeOf`, function signatures and locals when no representation is known. Cast refusal is in `internal/lower/cast.go`.

The census README and scope/method section were read. Its findings are measurements on a checker-rejected program, deduplicated by kind, location, reason and exact text; lowering reports only the first failure per attempted unit. The supplied 222 is not independently verified here. No before/after census, compiler gate, sanitizer oracle or mutants were run because implementation stopped at the language decision. There is no family movement to report.

## Approved refinement and implementation

On October 7, @system_adamic approved preserving the catchable TypeError through the undefined arm and refusing member names present on the actual primitive or its prototype chain. The implementation now follows that ruling. Optional reads still yield undefined, ordinary reads through undefined throw, and casts are erased. See [REPORT.md](REPORT.md) for fixtures, the original-source review oracle, mutants, counts and validation.
