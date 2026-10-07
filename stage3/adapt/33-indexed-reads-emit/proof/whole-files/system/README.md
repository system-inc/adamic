# transformers/module/system.ts: three optional-cache declines

All-code pinned latent census **4 -> 3**. Current input is byte-identical to the
previously reviewed Wave C file. The sole added assertion is U-position:
dependencyGroups[groupIndex]. The local groupIndices map stores the current
array length immediately before pushing a defined group. No operation deletes,
splices or truncates that private array. Thus every existing map index selects a
populated group. The assertion stays at the original read before externalImports
is mutated; no evaluation point moves.

Every remaining finding is listed in after.json. At lines 1773, 1774 and 1776,
moduleInfo, exportFunction and contextObject receive sparse cache values during
unconditional SourceFile emission notifications. transformSourceFile returns
plain scripts before populating these maps. The values can legitimately be
undefined, so assertions would claim a false invariant. The owning declarations
are in this file, but widening them requires proving numerous required reads in
transformation, helper generation and emission substitution. That complete
context protocol review remains unfinished; each read has an individual decline
reason in sites.json. The optional noSubstitution map is already handled by its
truthful optional owner and guard and receives no assertion.

Stock JavaScript equality and idempotence pass for all 30 owned files. The group
read ! -> ?? 0 mutant fails emitted-JavaScript equality and the site contract.
Census command uses /tmp/emit33-close-meter and source-only
/tmp/emit33-close-meter-source, output /tmp/emit33-close-latent-system.
Verification before snapshot: /tmp/emit33-close-source-before-system.
The default oracle and mechanical API projection reports accompany this proof.

The first default oracle stopped after an empty-output worker exit, with zero
reported test counts and no baseline diff. The observed cgroup OOM-kill count
had increased to 3 from the prior observed 1. This is a failed run, not a pass;
its report and logs are retained. A fresh isolated default run follows.

Fresh default oracle: **106367 passing**, zero failing/pending, empty baseline
diff, 211.299s. Mechanical API projection permits exactly
20's 189 plus 40's 28 lines; every other reference byte and the complete file
set match pristine. Final JavaScript equality and idempotence pass.
