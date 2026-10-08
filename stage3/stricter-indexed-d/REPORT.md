Built inline sentinel predicates, address-safe nullable Map/Set keys, null-byte backstop, JSON null classification and literal-null call fitting.
Commits: inline predicates and migration assertion 3a69e75c; Map/Set, conversions and new Node controls in this commit; prior witness completion 71897d7e.
Commands: all runtime-review controls and compiler packages pass; full merged witness/runtime regression results follow.
Mutants: erased Map hash/comparison address checks and erased JSON sentinel classification both caught; prior null/undefined and indexed guard mutants retained.
Not covered: full repository gate, whole TypeScript build and WebAssembly timing not run; mutable/cyclic record payloads stay outside the published representation.

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


## D151 representation investigation

Observed refusal: `src/compiler/sourcemap.ts:216`,
`map.sourcesContent[raw.sourceIndex]`, receiver `(string | null)[]`.
The minimized witness refuses at its local binding (line 2, column 7):
`stage 0 can't lower a value of type string | null yet`.
`internal/lower/expression.go`, `(*lowering).typeOf` (line 15), emits that
message at line 27 after `(*lowering).representation` (line 30) rejects the
union. The nullable-union branch (lines 63-74) supports only nullable regexp
match arrays and explicitly rejects a union containing both null and undefined.

Existing `string | undefined` representation is `ir.String`, an
`adamic_string *`: a live string pointer means string and NULL means undefined.
The union loop skips undefined and shares the remaining reference representation
(expression.go lines 78-117); native cType is in emit_values.go. There is no
MaybeString present/value representation. Mixed `ir.Union` also uses NULL for
undefined and does not provide a separate null tag. Existing nullable regexp
references reuse NULL for null, with static type information distinguishing it
from undefined. Equality lowering makes `=== undefined` always false for a
nullable operand; that cannot classify all three runtime states in one value.

There is an existing lookup-specific distinction: emit_slots.go,
`(*emitter).typeOfReference`, keeps `adamic_value *slot` long enough for direct
`typeof` to distinguish absent slot from present NULL payload. That presence
information is not carried in an ordinary string local. The next explicit
refusal is indexed_checks.go, `(*lowering).checkedIndexedRead`, line 31:
`an indexed presence check on nullable array elements (the lookup needs to
retain its presence slot)`.

Inference: merely admitting string | null as ir.String is insufficient for the
requested runtime distinction. D151 needs the guard to test slot presence,
rather than whether the payload is NULL. Beyond that guard, a value that can
hold string, null and undefined needs a representation decision: carry presence
alongside the pointer through IR/storage/observations, or introduce a distinct
null tag/sentinel alongside the existing undefined NULL. The temporary slot
used by typeof is not already such a general value representation. Per the
user's instruction, stop here without choosing or implementing either approach.
No null/undefined conflation mutant is claimed because that implementation has
not been selected. Existing dense-read erased-guard results remain unchanged.

All five blocked rows and dependencies:

| Row | Read | Waits on |
|---|---|---|
| D119 | moduleNameResolver.ts:474, typesVersions[key] | Nested record/index-signature representation and presence checks from stricter-indexed-c, then its witness and mutant |
| D129 | program.ts:4157, options.paths[key][i] | Record of string arrays representation from stricter-indexed-c, then its witness and mutant |
| D130 | program.ts:4160, options.paths[key][i] | Same record representation, then its individually observed witness and mutant |
| D131 | program.ts:4160, options.paths[key][i] | Same record representation, then its individually observed witness and mutant |
| D151 | sourcemap.ts:216, map.sourcesContent[raw.sourceIndex] | Null/undefined representation decision and slot-presence guard, then both-backend release/sanitized Node comparisons and conflation mutant |

Validation: reran `go test ./stage3/stricter-indexed-d -run
'^TestLedgerWitnesses/D151$' -count=1 -timeout 3m -v`, with output in
`evidence/D151-investigation.log`. This checks Node string/null/absent controls
and pins the current lower refusal; it does not prove native nullable support.
Only this report and the focused evidence log change in this investigation.


## Nullable strings, step 1

The representation decision is implemented: string | null and string | null |
undefined remain ir.String / adamic_string *, with no conversion between the
two. NULL stays undefined. The single immortal `adamic_null_string` sentinel is
static storage in `internal/native/runtime/nullable.c`; this is its only
declaration/definition site for runtime review. Shared declarations in adamic.h
expose adamic_reference_null, adamic_reference_is_null and sentinel observation
helpers. `internal/native/nullable.go` centralizes native null emission/testing,
and `ir.Type.UsesNullSentinel` identifies migrated kinds. Other reference kinds
retain their existing representation for z00sxvc to generalize.

Node-held project .ts controls cover null, a runtime-built string, undefined,
empty string and the real string "null". They exercise === null, === undefined,
== null, typeof, String(), direct console output, string equality, nullish
coalescing and movement through a string | null function into a string | null |
undefined parameter. Native release, ASan/UBSan and JS match stdout/stderr/exit.
The compiled sentinel-to-NULL mutant exits 0 but prints undefined for null;
the Node comparison catches it. This is a real runtime mutant, not a build kill.

Project .ts nullish ==/!= with a null operand is lowered explicitly; .a keeps
its existing loose-equality refusal. Indexed-presence remains the next step:
D151 now reaches the later nullable indexed-check refusal, pinned in sites.json.
The guard implementation and D151 proof are not claimed by this step.


## Nullable strings, step 2

Representation step pushed as eff7e8e9. Indexed checks now admit sentinel-backed
nullable string arrays. Their Coalesce IR sets UndefinedOnly, so native tests
NULL alone and JS uses the existing adamicDefined helper (=== undefined), not
JavaScript's nullish ?? operator. Ordinary ?? still treats both null and
undefined as missing. Null is an in-range value, not an absence.

`TestNullableStringPresenceGuard` runs independent [null], ["ma" + "de"] and []
programs. Source Node prints the three equality observations, typeof, String()
and direct console output. Present cases match Node in JS, native release and
ASan/UBSan. Absent cases stop in all three with exit 70, empty stdout and exact
named site stderr. IR identifies precisely one indexed-presence check. Erasing
that one native panic builds and runs in release and sanitized modes, then exits
0 with Node's undefined observations; the named-stop assertion catches both.

D151's existing witness now also passes string, null, absent, explain and
sanitized erase-guard cases. Its hole control remains an explicit constructor
refusal. To keep the manifest accurate as soon as support is enabled, D151 is
marked unblocked in this step; the next step expands its observations.

Commands: focused guard suite (1.613s), focused D151 witness (7.347s) and
`go test ./stage3/stricter-options -count=1 -timeout 5m` (5.778s), all pass.
Logs are committed as nullable-step2*.log. Current totals: 23 dense rows proven,
4 record rows blocked (D119, D129, D130, D131), 23 hole variants blocked.


## Nullable strings, step 3: D151 complete

D151 keeps `(string | null)[]`, the nested sourcesContent receiver and the
raw.sourceIndex property index. Its witness now explicitly prints === null,
=== undefined, == null, typeof, String(value) and direct console output.
Source Node's ["7"] control prints false/false/false/string/7/7; [null] prints
true/false/true/object/null/null; [] prints
false/true/true/undefined/undefined/undefined. JS and native release and
ASan/UBSan match both present controls. For [] all backends stop at the read
with empty stdout, exit 70 and exact stderr:
`adamic: panic: indexed read is absent: <witness-path>:2:30` plus newline.
The existing CLI assertion verifies checked indexed-presence=1, trusted=0,
and the independently computed read position; IR agrees.

Two real runtime mutants are caught:

* Erase the single indexed panic: D151's native sanitized binary exits 0 and
  prints Node's undefined observations, failing the required named exit-70
  stop. The generic presence fixture additionally runs this mutant in both
  release and sanitized native builds, with the same result.
* In an isolated copy of the runtime archive, remove nullable.o and recompile
  nullable.c with `adamic_reference_null(adamic_kind_string)` returning NULL.
  The generated program is unchanged. Under ASan/UBSan the general three-state
  control exits 0 with incorrect null/undefined classifications and output,
  failing its byte-for-byte Node assertion. D151's in-range null control
  incorrectly stops with exit 70 at its guard, failing its Node exit-0 result.
  Both mutants compile successfully. Runtime cache and repository files are
  untouched by the mutant; only temporary archive/object files change.

Extra representation controls cover reverse and negated null comparisons,
null == null and undefined == null, actual empty and "null" strings,
a runtime-built string, normal ??, and optional string length. Optional length
uses the same null helper so null?.length stays undefined, rather than reading
the sentinel's payload. Nullable-only values pass into a string | null |
undefined parameter through a function without conversion. All these controls
match Node in both backends, including native release and sanitizers.

Runtime review location: `internal/native/runtime/nullable.c` contains the
single private static `adamic_null_string` declaration/definition. Its immortal
header has reference count zero. No string-producing operation returns that
address. The common kind-based identity helper is adamic_reference_null;
shared declarations are in adamic.h. Native null emission and comparisons go
through internal/native/nullable.go, not a string-only case in the expression
emitter. Other reference kinds retain their existing representation pending
z00sxvc's nullable-reference generalization.

The original five blocked rows now stand as follows:

| Row | Current state and dependency |
|---|---|
| D119 | Blocked: nested record/index-signature representation and presence checks from stricter-indexed-c, then witness and mutant |
| D129 | Blocked: record of string arrays representation from stricter-indexed-c, then witness and mutant |
| D130 | Blocked: same record representation, then individually observed witness and mutant |
| D131 | Blocked: same record representation, then individually observed witness and mutant |
| D151 | Resolved: sentinel decision implemented, null and string controls proven, absent read checked, both mutants caught |

Fetched stricter-indexed-c again: FETCH_HEAD is still f5a2212c and its diff from
390af985 contains only its three witness/report files, no compiler representation.
There is no record implementation to merge yet. All 27 rows are assessed,
23 dense shapes proven, 4 record rows blocked, 23 additional hole variants
explicitly refused, and 0 rows remaining unassessed.

Final validation (all output logged, no test output piped):

```sh
go test ./internal/lower ./internal/ir ./internal/native ./internal/javascript ./stage3/stricter-indexed-d ./stage3/stricter-options -count=1 -timeout 10m
go test ./internal/lower ./stage3/stricter-indexed-d -count=1 -timeout 10m -v
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(indexing|string_index|narrowed_reads|narrowed_numbers|unions|typeof.*|regexp.*)\.a$' -count=1 -timeout 8m -v
go test ./stage3/stricter-indexed-d -run '^TestNullableString' -count=1 -timeout 3m -v
go test ./stage3/stricter-indexed-d -run '^TestLedgerWitnesses/D151$' -count=1 -timeout 3m -v
go vet ./internal/lower ./internal/ir ./internal/native ./internal/javascript ./stage3/stricter-indexed-d
python3 stage3/stricter-indexed-d/census.py
git diff --check
```

All pass. The broad package run includes native's complete package (135.851s),
lower (32.366s), IR (25.063s), this unit (95.961s), stricter-options (17.306s)
and JS (no standalone tests). After comparison edge-case fixes, lower passes
again in 19.507s and the complete unit in 56.217s, rerunning all 23 site mutants.
After strengthening the runtime mutant, nullable controls pass in 2.460s and
D151 passes in 4.650s. The 30 filtered oracle fixtures pass in 0.636s using the
worker cache; release and sanitized native and backend observations agree.
Census retains exact metadata for 27 rows with zero missing/extra. Vet and diff
check are clean. The whole repository gate remains unrun. Final logs are
preserved under evidence/nullable-final*.log.

## Published representation merges and first witness group

Merged c132be26 (sparse and numeric typed-array support) as 09180ec8 and
28d30cd3 (readonly records) as f7c17f3b. The one expression-entry conflict
preserves both typedArrayExpression and recordUse checks. D151 and nullable
representation controls, including both mutants, pass after both merges
(merged-nullable.log, 18.009s).

D129, D130 and D131 now pass present, absent, hole, outer-present and
outer-absent controls (merged-records.log, 14.816s). Chained reads require two
IR checks and two CLI checked entries, even though both share a source position.
The inner erased-guard mutant preserves the outer check. Separate outer-key
controls observe the record lookup directly, since Node throws if an absent
record key is immediately indexed. Their independently erased outer guard is
caught. All ordinary controls run in JS and native release and ASan/UBSan;
mutants build successfully under sanitizers before their output/exit differs.

D119 still refuses its nested record payload at D119.ts:1:24:
`stage 0 can't lower a record other than a readonly string index signature holding nonnullable scalars or arrays of nonnullable scalars yet`.
Observed in both present and absent programs after merging the implementation.
It waits on nested record payload representation, explicitly outside the
records worker's supported scalar/array payload scope. No flattening or new
record representation is attempted here.

First five hole variants D107, D171, D172, D173 and D174 pass with exact named
exit-70 stderr, source Node undefined, checked explain entries and individually
erased guards caught (holes-group1.log, 6.102s). Further hole groups follow.

Second hole group D175, D184, D185, D186 and D189 passes all backend and
sanitized mutant assertions (holes-group2.log, 6.570s). Total completed hole
variants so far: 10 of 23 standalone shapes, plus the three record-chain holes.

Third hole group D191, D192, D193, D194 and D225 passes (holes-group3.log,
9.382s), including each independently erased guard. 15 standalone holes proven.

Fourth hole group D226, D227, D228, D229 and D230 passes (holes-group4.log,
9.376s). 20 standalone holes proven; each guard mutant caught.


## Final hole group and per-row state

D231, D151 and D170 hole variants pass (holes-group5.log, 4.729s).
The user's original 21 holes are all proven; the added D170 and newly supported
D151 bring standalone array holes to 23. Record-chain D129/D130/D131 contribute
three additional hole controls, for 26 supported hole variants. No assigned row
needs a typed-array witness; its implementation arrived through the requested
merge and is checked in the dependency suite.

All proven rows below have present and absent Node controls, exact named
exit-70 site stderr in JS and native release and ASan/UBSan, independently
computed IR/CLI checked locations and erased-guard mutants caught. Hole controls
use new Array<T>(1), with the same receiver and index form as the dense source.

| Row | State | Checks and additional controls |
|---|---|---|
| D107 | Proven | One check; present/absent/hole and separate absent/hole guard mutants |
| D129 | Proven | Two chained checks; inner absent and hole mutants, separate outer-key present/absent control and outer mutant |
| D130 | Proven | Two chained checks; inner absent and hole mutants, separate outer-key present/absent control and outer mutant |
| D131 | Proven | Two chained checks; inner absent and hole mutants, separate outer-key present/absent control and outer mutant |
| D171 | Proven | One check; present/absent/hole and separate absent/hole guard mutants |
| D172 | Proven | One check; present/absent/hole and separate absent/hole guard mutants |
| D173 | Proven | One check; present/absent/hole and separate absent/hole guard mutants |
| D174 | Proven | One check; present/absent/hole and separate absent/hole guard mutants |
| D175 | Proven | One check; present/absent/hole and separate absent/hole guard mutants |
| D184 | Proven | One check; present/absent/hole and separate absent/hole guard mutants |
| D185 | Proven | One check; present/absent/hole and separate absent/hole guard mutants |
| D186 | Proven | One check; present/absent/hole and separate absent/hole guard mutants |
| D189 | Proven | One check; present/absent/hole and separate absent/hole guard mutants |
| D191 | Proven | One check; present/absent/hole and separate absent/hole guard mutants |
| D192 | Proven | One check; present/absent/hole and separate absent/hole guard mutants |
| D193 | Proven | One check; present/absent/hole and separate absent/hole guard mutants |
| D194 | Proven | One check; present/absent/hole and separate absent/hole guard mutants |
| D225 | Proven | One check; present/absent/hole and separate absent/hole guard mutants |
| D226 | Proven | One check; present/absent/hole and separate absent/hole guard mutants |
| D227 | Proven | One check; present/absent/hole and separate absent/hole guard mutants |
| D228 | Proven | One check; present/absent/hole and separate absent/hole guard mutants |
| D229 | Proven | One check; present/absent/hole and separate absent/hole guard mutants |
| D230 | Proven | One check; present/absent/hole and separate absent/hole guard mutants |
| D231 | Proven | One check; present/absent/hole and separate absent/hole guard mutants |
| D119 | Proven | One checked record read; nested receiver and string-variable key; release and sanitized erased-guard mutants caught |
| D151 | Proven | One presence check; string/null/absent/hole; equality, typeof, String and console; both guard and sentinel-conflation mutants |
| D170 | Proven | One check; present/absent/hole and separate absent/hole guard mutants |

D119 is the sole unresolved row. Its receiver is a readonly string-key record
whose payload is another record of string arrays. The published recordInfo in
internal/lower/records.go admits scalars and arrays of scalars, but does not
admit that nested payload. Both source Node controls are observed and the exact
compiler refusal is pinned. No native result, checked classification or guard
mutant is claimed for D119.

The private string sentinel remains declared once in
internal/native/runtime/nullable.c for runtime review; both dependency merges
preserve its undefined-only indexed guards and separate null observations.

## Combined merge regression validation

Initial dependency run exposed a diagnostic-order regression:
TestTypedArrayRepresentationBoundaries/cast-view expected the typed-array
storage refusal, but recordUse caught that non-record cast first. The correction
limits recordUse's cast refusal to types with a string record index signature.
The existing cast path still refuses incompatible typed-array storage views;
record cast/view boundaries retain their record refusal. This does not admit
new representations or runtime behavior. The failed run is retained in
merged-dependencies.log, and the corrected full suites pass in
merged-dependencies-fixed.log.

Commands (all test output redirected to evidence logs):

```sh
source /workspace/adamic-tools/env.sh
go test ./internal/lower ./internal/ir ./internal/javascript -count=1 -timeout 10m
go test ./internal/native -count=1 -timeout 10m
go test ./stage3/stricter-indexed-d -count=1 -timeout 10m -v
go test ./internal/lower ./stage3/stricter-options ./stage3/stricter-records -count=1 -timeout 10m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(indexing|string_index|narrowed_reads|narrowed_numbers|unions|typeof.*|regexp.*)\.a$' -count=1 -timeout 8m -v
go vet ./internal/lower ./internal/ir ./internal/native ./internal/javascript ./stage3/stricter-indexed-d ./stage3/stricter-options ./stage3/stricter-records
python3 stage3/stricter-indexed-d/census.py
git diff --check
```

Compiler package results: lower 25.882s, IR 21.118s, JavaScript no standalone
tests, native 141.275s. After the cast correction: lower 30.773s,
stricter-options 39.435s and stricter-records 39.879s, all pass. The complete
unit before that correction passes in 75.668s, with 55 independently erased
guard mutants caught (26 absent, 26 hole, three outer-record absent), plus two
sentinel-to-NULL runtime mutants caught. Generic nullable controls additionally
hold erased guards in release and sanitized builds. The uncached 30-fixture
Node oracle passes in 12.089s (native 0 cache hits/85 misses, Node 0 hits/60
misses). Vet and diff check have empty logs; census reports 27 rows, zero
missing/extra, exact source metadata. Setup was reused from the prior successful
run, with its recorded timing unchanged; nproc is still 5. The complete
repository gate and whole TypeScript native build remain unrun.

Final complete witness rerun after the cast correction passes in 80.072s,
with all 27 row assessments and both nullable runtime mutants rerun; raw output
is merged-final-witnesses.log. No supported row or hole variant remains.

## Runtime review: inline predicates and migration backstop

The sentinel now has one external declaration in adamic.h and one definition
in nullable.c. adamic_reference_is_null and adamic_reference_is_sentinel are
static inline in adamic.h and compare the string sentinel directly by address.
They do not call adamic_reference_null. Unmigrated kinds retain their previous
NULL behavior. TestNullReferenceUsesOnlyMigratedKinds asserts that native
nullReference emits a runtime sentinel constructor only for ir.String, and
emits NULL for Object, Array, Record, Map, Closure, Union and Weak.

A Node-held 100,000,000-iteration guarded `(string | null)[]` read loop alternates
"abc" and null and prints 150000000. Five release trials before inlining:
1.036942, 1.050600, 0.988727, 1.029420, 1.015818 seconds. After inlining:
0.826715, 0.788082, 0.810806, 0.793662, 0.794983 seconds. Median changes from
1.029420 to 0.794983 seconds, approximately 22.8% lower elapsed time. This is an
observation on this worker, not a cross-target performance guarantee. Both
versions assert one IR indexed-presence check and match Node stdout/exit/stderr.
No semantic mutant is claimed for the inline optimization.

Logs: runtime-review-before.log and runtime-review-after-inline.log. Commands:
`ADAMIC_NULL_BENCH=before go test ./stage3/stricter-indexed-d -run '^TestNullableGuardTiming$' -count=1 -v`
and the same with ADAMIC_NULL_BENCH=after-inline. Migration backstop passes in
runtime-review-kinds.log. Existing nullable controls and D151 pass in 17.556s,
including the sentinel-to-NULL and erased-presence mutants
(runtime-review-inline-controls.log). Setup/environment is reused; nproc 5.


## Runtime review: Map/Set identity, bytes and conversions

String-key hashing checks the sentinel address before reading bytes; null gets
its own fixed hash. String-key comparison checks either sentinel address first
and compares sentinel values by identity, then handles undefined NULL and normal
string contents. The normal byte comparison follows the existing canonical
WTF-8 equality used by strings. Set uses the same map storage. The fixture
stores empty string, null, a runtime-built real string "null", and undefined
in one Map and one Set. It exercises size, get/has, overwrite, duplicate add
and delete. All four keys stay distinct in source Node, JS, native release and
ASan/UBSan. Removing exactly the Map hash and comparison sentinel-address
branches from a temporary replacement map.o compiles and runs, but exits 0
with size 3 instead of Node's 4 and wrong lookups/deletions. The byte-for-byte
Node fixture catches it. The original archive and runtime cache are unchanged.

The one adamic_null_string definition in internal/native/runtime/nullable.c
now uses ADAMIC_STRING("null"), retaining its immortal string-kind header.
adamic.h contains its sole extern declaration and both inline predicates.
A direct C runtime backstop checks the header (references 0, string kind),
constructor address and raw sentinel output against Node's String(null), in
release and ASan/UBSan. Null equality remains by address, distinct from a real
string with the same bytes.

Ten independent project .ts programs hold these requested shapes to Node:
Map/Set, narrowed concatenation, narrowed length, narrowed JSON.stringify,
narrowed slice, narrowed equality/order comparison, null template, null +
string in both orders, null JSON.stringify, and String(null). Narrowed controls
pass a runtime-built string, empty string, real "null" and literal null through
a string | null function. Conversion controls include literal-null templates
and both literal-null concatenation orders as well as a nullable variable.
All stdout/stderr/exit observations match in both backends, including native
release and ASan/UBSan (runtime-review-controls-final.log, 5.781s).

The controls first exposed a real pre-existing native call gap: literal null
passed to a string | null parameter still emitted NULL because argument fitting
handled maybe/union/weak but not null's reference kind. This could incorrectly
enter a narrowed branch and panic, or print undefined. Native arguments now
fit ir.Null to the parameter's reference kind before normal evaluation through
nullReference. It is shared across kinds, and UsesNullSentinel continues to
keep unmigrated kinds off the constructor's default path. Literal null in
string concatenation/templates is spelled "null" directly. No mixed numeric
coercion or general object ToPrimitive support is introduced.

JSON schema validation now recognizes null members, allowing the sentinel-backed
string | null representation. Runtime scalar classification identifies the
sentinel as JSON null, so it writes null while the real string "null" is quoted;
NULL remains undefined. An additional isolated JSON runtime mutant erases that
classification branch, compiles and exits 0 with quoted "null" for null. The
null JSON Node fixture catches it. Initial failing observations are retained
in runtime-review-controls.log; the intermediate literal-call fix is retained
in runtime-review-controls-fixed.log. Compiler lower/IR packages pass in
22.345s and 14.473s; JavaScript has no standalone tests.


## D119 completed alongside runtime review

Merged codex/stricter-records b150f83c as a160f053. The finite readonly nested
record validator merged cleanly and preserves this branch's string-record cast
scoping correction. D119's existing source still declares
`{ readonly [key: string]: { readonly [path: string]: string[] } }`, reads
`typesVersions[key]` and observes the record-valued result. The present key
prints 7 under Node, JS, native release and ASan/UBSan. The absent key prints
undefined under source Node; both compiled backends instead stop with empty
stdout, exit 70 and exact independently located indexed-read stderr.
IR inventory and CLI explain both assert exactly one indexed-presence guard
and zero trusted checks. The witness is a record read, so no array-hole variant
is generated for it. Existing record-to-array chains keep their explicit hole
controls and two guards.

D119's erased single guard is compiled and run independently in native release
and ASan/UBSan. Both mutants exit 0, stdout "undefined\n", empty stderr. Both
are caught by the named-stop comparison. D119's complete witness passes in
7.488s (runtime-review-D119.log). All 27 assigned rows are now proven, all 26
array/record-chain hole variants are proven, no rows blocked or remaining.
Earlier refusal descriptions in this report are historical checkpoints;
the per-row table now marks D119 proven. Every row's current state is proven.
The published validator still refuses mutable/cyclic record payloads and
finite nesting beyond its conservative bound; D119 needs none of those.
