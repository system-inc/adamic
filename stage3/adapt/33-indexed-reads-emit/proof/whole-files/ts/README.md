# transformers/ts.ts at zero

All-code pinned latent census **3 -> 0**; clean files **38 -> 39 of 78**.
This source is byte-identical to the previous reviewed wave before these edits.
The two existing U-position ledger declines become assertions at their reads.

findSuperStatementIndexPath starts with an empty index list and records i only
inside i < statements.length. A super-call statement terminates the path. A try
prefix is prepended only if its recursive tryBlock search succeeded. The caller
enters the worker at depth zero only when that path is nonempty. By induction,
a try statement always has a next path component; the terminal component selects
the super-call statement. Each recursion supplies the same tryBlock.statements
used by the producer, so each recorded index selects a present input statement.
Parameter-property generation clones parameter identifiers, and visitor/factory
updates create output arrays without modifying these input statement arrays.
Both reads occur before the worker's callbacks and remain at their original points.

All stock JavaScript bytes and adapter idempotence pass. The path-index ! -> ?? 0
mutant fails both independent checks. Default oracle: **106367
passing**, zero failing/pending, **empty baseline diff**, 219.657s.
Mechanical API projection is exactly 20's 189 plus 40's 28 lines; every other
reference byte and the complete file set match pristine.
Reproduce using before snapshot /tmp/emit33-close-source-before-ts and the same
source-only census/oracle commands as emitter, with MUTANT_FILE=transformers/ts.ts.
