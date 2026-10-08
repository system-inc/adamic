Built probes for all 27 corrected ledger rows: 22 dense read shapes proven, 5 representation refusals.
Commits: original series through c019ca20; corrected-scope extension is the next branch commit.
Commands: corrected census, complete witness suite, filtered oracle and vet pass; outputs below.
Mutants: 22 erased guards caught, including new D170 exit 0 with undefined stdout; D107 UBSan exit 1.
Not covered: 4 record rows awaiting representation, nullable D151, 22 hole variants, whole-program compilation and full gate.

## Scope and assumptions

The exact seven-file filter at ledger commit a1a16427 produces 24 rows in both
rows.csv and rows-2026-10-08.csv, rather than the requested 27. This unit follows
the exact named-file and option filter. The user corrected ownership to add moduleNameResolver.ts, sourcemap.ts and
transformers/declarations.ts. The current census uses rows-2026-10-08.csv and
contains exactly 27 rows across those ten files.

The tests generate .ts sources under a project tsconfig with
noUncheckedIndexedAccess=false. They run that actual source with Node's type
stripping, then compare native release, ASan/UBSan and backend JavaScript.
The present control prints 7; the absent read prints undefined on source Node.
Native/backend failure pins stdout, stderr and exit 70 independently, including
an independently computed source line and column. The CLI --explain-checks must
list the inserted site and count one checked indexed guard and zero trusted.
Each supported absent case has its sole emitted-C panic erased, must still build
under sanitizers, and must lose the exact named exit-70 result.

Receiver categories and index expressions are preserved. TypeScript's large AST
interfaces are reduced to the field needed for observation; NodeArray metadata
is omitted while preserving a readonly array receiver. These are minimal read
shape witnesses, not whole transformer executions or whole-program compilation.
Stored new source files are Go and JSON; project .ts inputs are generated tests,
following stage3/stricter-options/indexed_test.go.

The rows in program.ts use options.paths[key][i], a record-to-array chain.
D184, D185 and D186 all originate in superPath[superPathDepth]. D184's TS2538
reports the resulting undefined index; D185 and D186 report arithmetic uses of
that same index. A statement-array-only probe would test a different cause.
D231's tuple position [1] is static; the unchecked read is its nested declarations[0].
No receiver in this exact slice is a Map, typed array or indexed string.

## Progress

Toolchain setup is complete. The first attempted test exited 1
before compilation: cohere/TypeScript/tsc/go.mod was not yet available during
setup's recursive submodule checkout. This is not a witness or mutant result.

Group 1 passes: 5 ledger rows assessed, 2 dense read witnesses proven, 3 record
chains blocked, 19 rows remaining. D107 and D171 hole variants are blocked.
Record refusal: `Adamic 0.1 refuses an index signature; use a Map, which keeps
keys in the order they were added`. Hole refusal: `stage 0 can't lower new an
Identifier yet`. Node prints 7 for present controls and undefined for empty/hole
reads. Group command: `go test ./stage3/stricter-indexed-d -count=1 -timeout 10m -v`;
exit 0, package 6.652s, log `/tmp/stricter-indexed-d-group1.log`.

Setup completed successfully in 502.808s: Go 0.044s, Node 0.024s, clang 0.226s,
markdown dependencies 1.785s, submodules 295.588s, Go build 502.683s, cache warm
502.780s. nproc=5, cpu.max=400000 100000. Environment sourced from
`/workspace/adamic-tools/env.sh`. Go 1.27.1, Node 24.19.0, clang 20.1.8.

Group 1 commit: 8f28caf6, pushed to codex/stricter-indexed-d.
Group 2 adds D172, D173, D174, D175 and D184: present and absent proofs pass in
all modes; all five erase-panic mutants build and exit 0, caught by the required
named stop. Their hole variants all produce undefined on Node and the same
constructor refusal natively. Total: 7 dense reads proven, 3 record rows blocked,
14 remaining. Filtered group command uses
`-run 'TestLedgerWitnesses/(D172|D173|D174|D175|D184)$'`; exit 0, package 13.101s,
log `/tmp/stricter-indexed-d-group2.log`.

Group 2 commit: e4ebd182, pushed to codex/stricter-indexed-d.
Halfway report, after group 3: 15 rows assessed, 12 dense reads proven,
3 record rows blocked, 9 rows remaining. New proofs are D185, D186, D189, D191
and D192. All five erase-panic mutants build and exit 0; exact exit/stderr
assertions catch them. Their hole variants remain refused. Command filter:
`-run 'TestLedgerWitnesses/(D185|D186|D189|D191|D192)$'`; exit 0, package 13.623s,
log `/tmp/stricter-indexed-d-group3.log`.

Group 3 commit: ee0a5320, pushed to codex/stricter-indexed-d.
Group 4 proves D193, D194, D225, D226 and D227, with five more caught
exit-0 erase-panic mutants and five observed hole-constructor refusals.
Total: 17 dense reads proven, 3 record rows blocked, 4 remaining.
Command filter: `-run 'TestLedgerWitnesses/(D193|D194|D225|D226|D227)$'`;
exit 0, package 13.280s, log `/tmp/stricter-indexed-d-group4.log`.

Group 4 commit: 3c5c85b4, pushed to codex/stricter-indexed-d.
Group 5 proves D228, D229, D230 and D231. Four more erase-panic mutants build
and exit 0, caught by exact named-stop assertions. All four hole variants refuse.
Total: 21 dense reads proven, 3 record rows blocked, 0 known matching rows
remaining. Command filter: `-run 'TestLedgerWitnesses/(D228|D229|D230|D231)$'`;
exit 0, package 11.659s, log `/tmp/stricter-indexed-d-group5.log`.


Group 5 commit: 9359c2d5, pushed to codex/stricter-indexed-d.

## Initial 24-row verification, October 8, 2026

All 24 exact matching rows are assessed. Dense/out-of-range proof is complete for
21. D129, D130 and D131 are blocked by the index-signature refusal above. All
21 array hole variants are separately blocked by the constructor refusal above.
These gaps belong to the representation worker; no implementation is added here.
No matching rows in the original seven-file filter remain unassessed. The three
newly assigned files and their results are recorded in the corrected-scope
verification below.

Every supported present control prints `7\n`, no stderr, exit 0 on source Node,
backend Node, release native and sanitized native. Every supported empty-array
source prints `undefined\n`, no stderr, exit 0 on Node. Both backends instead
produce no stdout, exact `adamic: panic: indexed read is absent: file:line:column\n`
stderr, exit 70. The independently calculated read location must match IR and
CLI explain output, with indexed-presence=1 and trusted=0.

Every mutant erases exactly the one panic in generated C. Native sanitizer
builds succeed. The named-stop assertion catches all 21. D107 then reaches a
null string use: UBSan reports member access within null pointer of type
adamic_string in string_build_impl.h:40:21 and exit 1. The other 20 exit 0.
The first checkpoint summary incorrectly said both initial mutants exited 0;
D107 actually exited 1 in that run too. This report corrects that observation.
No build failure is counted as a mutant kill.

Commands (environment sourced from /workspace/adamic-tools/env.sh):

```sh
python3 stage3/stricter-indexed-d/census.py > /tmp/stricter-indexed-d-census.log 2>&1
go test ./stage3/stricter-indexed-d ./stage3/stricter-options -count=1 -timeout 10m -v > /tmp/stricter-indexed-d-final.log 2>&1
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(indexing|string_index|narrowed_reads|narrowed_numbers)\.a$' -count=1 -timeout 10m -v > /tmp/stricter-indexed-d-oracle.log 2>&1
go test ./stage3/stricter-indexed-d -run 'TestLedgerWitnesses/D107$' -count=1 -timeout 10m -v > /tmp/stricter-indexed-d-mutant-D107.log 2>&1
go vet ./stage3/stricter-indexed-d > /tmp/stricter-indexed-d-vet.log 2>&1
git diff --check
```

All commands exit 0. Census: 24 rows, 0 missing, 0 extra, exact metadata.
New witness package: 47.338s. Inherited stricter-options package: 10.164s.
Filtered oracle: 3.617s, all four named fixtures pass. D107 supplemental run:
3.595s, preserving mutant stderr after adding output logging. Vet and diff check
emit no diagnostics. The supplemental change only logs mutant outputs; it does
not alter the witnesses or assertions verified by the combined run.

The inherited package also reruns its eight indexed and three JSON erase-check
mutants. Those remain generic predecessor proofs, not additional ledger rows
credited to this unit. Full repository gate and whole TypeScript native build
were not run. Minimal object element types retain the receiver category and
index expression, not the full AST union representation. An actual whole-program
emitted guard count or whole-program completion date cannot be established by
these shape witnesses. No compiler implementation or another worker's files
were changed. Only codex/stricter-indexed-d was pushed; no PR was opened.

Raw successful output is preserved under [evidence](evidence/).

## Per-site result and mutant inventory

Each dense row below also has a separately observed hole-constructor refusal.
Each record row has present and empty-inner-array Node observations plus the
native representation refusal, and has no emitted guard or mutant claim.

| Ledger ID | Source file and line | Traced read | Receiver | Result | Erase-panic mutant |
|---|---|---|---|---|---|
| D107 | src/compiler/commandLineParser.ts:2077 | `args[i]` | array of string | Dense read proven; hole blocked | Caught: UBSan exit 1 |
| D129 | src/compiler/program.ts:4157 | `options.paths[key][i]` | record of string arrays | Blocked: index signature | Not emitted |
| D130 | src/compiler/program.ts:4160 | `options.paths[key][i]` | record of string arrays | Blocked: index signature | Not emitted |
| D131 | src/compiler/program.ts:4160 | `options.paths[key][i]` | record of string arrays | Blocked: index signature | Not emitted |
| D171 | src/compiler/transformers/es2015.ts:2803 | `node.declarations[0]` | NodeArray of object | Dense read proven; hole blocked | Caught: exit 0 |
| D172 | src/compiler/transformers/es2015.ts:3046 | `declarations[0]` | array of object | Dense read proven; hole blocked | Caught: exit 0 |
| D173 | src/compiler/transformers/es2015.ts:4437 | `funcStatements[classBodyStart]` | NodeArray of object | Dense read proven; hole blocked | Caught: exit 0 |
| D174 | src/compiler/transformers/es2015.ts:4691 | `segments[0]` | array of object | Dense read proven; hole blocked | Caught: exit 0 |
| D175 | src/compiler/transformers/es2015.ts:4693 | `segments[0]` | array of object | Dense read proven; hole blocked | Caught: exit 0 |
| D184 | src/compiler/transformers/esDecorators.ts:1164 | `superPath[superPathDepth]` | readonly array of number | Dense read proven; hole blocked | Caught: exit 0 |
| D185 | src/compiler/transformers/esDecorators.ts:1165 | `superPath[superPathDepth]` | readonly array of number | Dense read proven; hole blocked | Caught: exit 0 |
| D186 | src/compiler/transformers/esDecorators.ts:1192 | `superPath[superPathDepth]` | readonly array of number | Dense read proven; hole blocked | Caught: exit 0 |
| D189 | src/compiler/transformers/esnext.ts:183 | `node.statements[pos]` | NodeArray of object | Dense read proven; hole blocked | Caught: exit 0 |
| D191 | src/compiler/transformers/esnext.ts:344 | `statementsIn[i]` | readonly array of object | Dense read proven; hole blocked | Caught: exit 0 |
| D192 | src/compiler/transformers/esnext.ts:384 | `statementsIn[i]` | readonly array of object | Dense read proven; hole blocked | Caught: exit 0 |
| D193 | src/compiler/transformers/esnext.ts:767 | `statements[i]` | readonly array of object | Dense read proven; hole blocked | Caught: exit 0 |
| D194 | src/compiler/transformers/esnext.ts:767 | `statements[i]` | readonly array of object | Dense read proven; hole blocked | Caught: exit 0 |
| D225 | src/compiler/transformers/jsx.ts:309 | `nonWhitespaceChildren[0]` | readonly array of object | Dense read proven; hole blocked | Caught: exit 0 |
| D226 | src/compiler/transformers/jsx.ts:502 | `expressions[0]` | array of object | Dense read proven; hole blocked | Caught: exit 0 |
| D227 | src/compiler/transformers/module/esnextAnd2015.ts:100 | `node.arguments[0]` | NodeArray of object | Dense read proven; hole blocked | Caught: exit 0 |
| D228 | src/compiler/transformers/module/esnextAnd2015.ts:163 | `importsAndRequiresToRewriteOrShim[0]` | array of object | Dense read proven; hole blocked | Caught: exit 0 |
| D229 | src/compiler/transformers/module/esnextAnd2015.ts:194 | `node.arguments[0]` | NodeArray of object | Dense read proven; hole blocked | Caught: exit 0 |
| D230 | src/compiler/transformers/module/esnextAnd2015.ts:196 | `node.arguments[0]` | NodeArray of object | Dense read proven; hole blocked | Caught: exit 0 |
| D231 | src/compiler/transformers/module/esnextAnd2015.ts:251 | `importRequireStatements[1].declarationList.declarations[0]` | NodeArray of object inside tuple | Dense read proven; hole blocked | Caught: exit 0 |


## Corrected-scope verification, October 8, 2026

The three added updated-ledger rows are D119 at moduleNameResolver.ts:474,
D151 at sourcemap.ts:216 and D170 at transformers/declarations.ts:1678.
Their source expression and cause are retained exactly from
rows-2026-10-08.csv at a1a16427. The census now covers all 27 corrected assigned
rows: 0 missing, 0 extra, all metadata exact.

D170 traces to clause.types[0], a NodeArray object read with literal zero index.
Its minimized payload retains the same readonly array and index form. Source
Node prints 7 for present and undefined for absent. Backend Node, native release
and ASan/UBSan agree on the present case; absent cases produce empty stdout,
exact named site stderr and exit 70. IR and CLI explain each identify one checked
indexed-presence site and zero trusted. Its erase-panic mutant successfully builds
under sanitizers, then exits 0, stdout undefined, empty stderr. The exact named
exit-70 assertion catches it. Its new Array<Item>(1) hole variant observes Node
undefined and the existing constructor refusal, without claiming a guard.

D119 traces to typesVersions[key], an index signature whose value is another
index signature of string arrays. Node observes 7 for the present record and
undefined for the absent key. Native lowering refuses at the index signature:
`Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order
they were added`. No native guard or mutant is claimed.

D151 retains the exact (string | null)[] receiver and raw.sourceIndex property
index. Source controls print 7, undefined, and null respectively. All three
native attempts refuse at the required local binding with
`stage 0 can't lower a value of type string | null yet`, before emission of an
indexed guard. The initial test expected the later nullable indexed-presence
refusal from indexed_checks.go; the observed earlier value refusal is now pinned.
Both the initial failure log and final passing refusal probes are preserved.
Using a string-only receiver would test a different contract and is not done.

Dependency check: codex/stricter-indexed-c was fetched twice. Both observations
found f5a2212c, whose diff from 390af985 changes only its REPORT.md, sites.json
and witness_test.go. Its report explicitly says compiler implementation is
unchanged. That push contains no record representation to merge. D129, D130 and
D131 remain blocked awaiting its implementation push, as instructed; D119 also
needs that record representation. No merge of the pending implementation is
claimed. The branch can resume those witnesses when that dependency is pushed.

Commands (the established setup environment is sourced for each test shell):

```sh
python3 stage3/stricter-indexed-d/census.py > /tmp/stricter-indexed-d-census-corrected.log 2>&1
go test ./stage3/stricter-indexed-d -count=1 -timeout 10m -v > /tmp/stricter-indexed-d-corrected-final.log 2>&1
go vet ./stage3/stricter-indexed-d > /tmp/stricter-indexed-d-corrected-vet.log 2>&1
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(indexing|string_index|narrowed_reads|narrowed_numbers)\.a$' -count=1 -timeout 10m -v > /tmp/stricter-indexed-d-corrected-oracle.log 2>&1
git diff --check
```

All commands exit 0. Complete corrected witness package: 40.822s, all 27 rows
assessed and all 22 supported per-site mutants rerun and caught. Filtered oracle:
0.488s, all four indexing fixtures pass. Census: 27 rows, 0 missing, 0 extra.
Vet and diff check have no diagnostics. The complete repository gate and whole
TypeScript native build remain unrun. nproc remains 5; the prior successful setup
and its timing log remain the established toolchain for this continued task.
Only this unit's harness, manifest, census, report and evidence changed.

| Added row | Traced read | Receiver | Result | Mutant |
|---|---|---|---|---|
| D119 | typesVersions[key] | record of records of string arrays | Blocked: index signature | Not emitted |
| D151 | map.sourcesContent[raw.sourceIndex] | array of string or null | Blocked: nullable value lowering | Not emitted |
| D170 | clause.types[0] | NodeArray of object | Dense read proven; hole blocked | Caught: exit 0, undefined stdout |

Current totals: 22 dense reads proven; 5 rows blocked (D119, D129, D130, D131,
D151); 22 additional hole variants blocked; 0 rows remaining unassessed.
