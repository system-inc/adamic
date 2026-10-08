Built: adaptation 41 removes 39 explicit-any tokens, types five untyped locals and adapts 15 void expressions; no compiler edits.
Commits: timer 613159a5, constructors 53216ca4, copies a10a1321, JSON 38a54ffe, locals a2df3565, data d6e5cff9; JSON owner repair/evidence 97d79455.
Commands/results: fresh apply exit 0; stock tsc zero diagnostics; full oracle 106366 passing/one sanctioned API failure/zero pending, lane PASS; any 135 ->96 tokens and 37 ->20 direct observations.
Mutants: real restored any/void, missing constructor/copy fields, changed validator/comparer, drift/duplicate guards, private JSON owners, debugger deletion and real fixture-input changes all fail their stated checks in MUTANTS.md.
Not completed: 96 explicit tokens and 20 direct observations remain, with exact decisions in RESIDUE.md; debugger retained, array storage dropped as supported, with absent, variance and ranked source fixes untouched.

## Scope, inputs and measurements

Branch codex/stage3-explicit-any-2 began at ef3141e9, resumed the accepted
c6ea4131 and merged current origin/main 5fda2d26 at 68a6f7db. The supplied
notyet TABLE.md is dc6b1529. Main changed scheduling/docs/worker policy, with
no compiler, adaptation or census-input change. Provenance and exact changed
paths are in evidence/finished/main-provenance.json. The original main census
and oracle therefore remain the unchanged-input comparison. Main’s oracle used
five workers; the final oracle uses the now-required eight, with identical
counts and the same sole sanctioned API failure.

Main’s direct census is 37 NotYet: 30 values, six returns, one field store; zero
array-any or call-return-any rows. Files/counts:

| File under src/compiler | Main | Final |
| --- | ---: | ---: |
| commandLineParser.ts |21|15|
| utilities.ts |4|2|
| sourcemap.ts |3|0|
| sys.ts |3|2|
| emitter.ts |1|0|
| moduleNameResolver.ts |1|0|
| moduleSpecifiers.ts |1|0|
| resolutionCache.ts |1|0|
| scanner.ts |1|0|
| tsbuildPublic.ts |1|1|

Final is 20 observations: 16 value-any, three return-any, one field store.
Nineteen still have source any contracts. sys.ts:52:31 is already explicitly
NodeJS.Timeout|undefined; stock TypeScript resolves two non-any constituents,
while latent still calls it any. Its cause is not diagnosed.

The stock parser counts 135 explicit AnyKeyword tokens in 23 files on main and 96
in 21 files after41. These are different from latent counts: the measurement
skips checker-diagnosed bodies and stops at each eligible unit’s first error.
The historical table’s85 value sites are not the reproduced current-main 37.
All counts remain labelled measured on a checker-rejected program.

## Family receipts

| Family | Explicit tokens removed | Direct any before -> after -> mutant | Other census observation |
| --- | ---: | --- | --- |
| Initial accepted source owners |11|37 ->36 ->37|Actual scanner-state restoration|
| Node timer globals |3|36 ->35 ->36|True Node handle at captured globals; public hosts deferred|
| Complete allocator stages |4|35 ->35 ->35|Refused5157 ->5156 ->5157, restored links cast|
| Shallow copies |8|35 ->35 ->35|Refused5156 ->5151 ->5152, restored clone source cast|
| JSON/config |11|35 ->26 ->27|Refused5151 ->5159; NotYet1468 ->1470 exposes union/generic lessons|
| Untyped locals |0|26 ->21 ->22|Refused5159 ->5161; NotYet1470 ->1468|
| Settings comparison |2|21 ->20 ->21|Original two-any parameter mutant; body unchanged|

The 53 exact type-owner rules include staged fields and completion steps ; 15 void
owners are separately guarded. The generated incremental row is16 files,81
lines added/67 removed. Only that row and Total are refreshed in patch-set.md.
No compiler source, shared lane/oracle/census harness or counts.md is changed.
There are no new fixtures.

Overall Refused5162 ->5161 and NotYet1468 ->1468. Any removals expose other
lessons; this is not a native-tsc completion claim. JSON uses a recursive actual
value union, not a top-type alias. Copies use partial construction and actual
own-enumerable caller records; constructor headers are partial until completion.
Actual bodies/callers and discarded proposals are documented per family.

## Behavior proof and the declaration repair

Every touched continuation family file emits byte-identical standalone JavaScript.
Actual Symbol, Signature and SourceMapSource construction, filesystem-entry
creation/copying, config recovery, map validation and settings comparison run
on Node. Actual missing-field/version/comparison mutants fail those proofs.
Idempotence and LF reconstruction pass with real guard mutants.

The final lane is /workspace/adaptation41-finished-repaired-lane: fresh apply
exit 0, oracle exit 1 solely for the existing API exception,106366 passing/one
failing/zero pending, and PASS with exactly 222 sanctioned declarations. It
took502.648 seconds; oracle alone385.443 seconds. All79 census source files
match the lane’s fresh apply tree byte for byte. All ten built JavaScript files
and public API are byte-identical to the preceding constructor control. Public
API is also byte-identical to main. Three JavaScript bundles differ from main
because of the already proved void source rewrites; their runtime observations
and full oracle remain identical. The byte comparisons explicitly retain this
difference instead of claiming all final JS bytes identical to main.

The first final lane was a real failure: private JsonConfigObject was omitted
from the internal bundle, yielding APILibCheck plus the API exception,106365
passing/two failing. Private declarations also changed namespace export printing.
The repair exports both shared declarations with @internal. A concrete rebuild
restores public API byte identity; the fresh full lane proves the repaired
internal API fixture passes. The failed run is retained as the ownership mutant.
The final lane execution records d6e5cff9 plus the then-working owner-export
repair; final rules/source hashes identify those exact tested bytes.

A proposed API-fixture command used eight workers; upstream’s parallel host
ignores --tests, so it started the full suite and was cancelled. It is not a
reported fixture pass. No full Go-package gate or confirmation run was made.
The initial composed-constructor guard failures and partial comparer mutant
are also retained as discarded probes. See MUTANTS.md for every actual check.

## Remaining decisions and ranked routes

RESIDUE.md lists all20 final observed rows with file:line:column, why and minimal
programs. JSON has a truthful recursive domain, but public result/input changes
need additional lane sanctions and raw/validated-owner work. The actual option
predicate accepts an array of objects without validating string elements, and
readJson promises object but returns7. Host timers need H across public host,
storage and cancellation contracts; an alias of their any return is rejected.
Node/Type/NodeLinks headers still omit required fields or have undefined parent;
SourceFile copying preserves external enumerable extensions. These are unfinished
owner/consumer or API decisions, not proofs of impossibility.

The 96 syntax tokens include JSON 27, timer 13, constructors 7, SourceFile copy 2 and
47 other historical owner sites not completed in this continuation. All exact
locations/snippets and classes are in evidence/finished-sites.json.
Debug.fail’s debugger is retained: removing it loses an attached-inspector pause.
Any-array storage/length already compiles on main, so that kind is dropped;
element reading is not claimed supported. With has no disposition row/source
site and fails checking. The historical timer any-call row is reviewed and
deferred with H; no completed call-specific adaptation mutant is claimed.

RANKED.md lists every other source-edit candidate row and current sites:
197 receiver captures, 81 boolean conditions, 11 candidate index signatures,11 definite
assignment assertions,11 namespaces, six parameter properties, six var, five
generator functions, five yield. Main has 124 exact candidate rows/328 sites;
final 126 rows/333 sites. No ranked candidate fix or variance fix was made.
Two new unbounded JSON dictionary signatures are compiler lessons, excluded from
the source-candidate ranking; the raw census keeps all13 index-signature rows.
Distinct expando-field and missing-return-annotation rows are absent (zero),
without a claim that the source lacks those issues. The 1389 unchecked casts
are mixed source contracts/dynamic boundaries, not all source-only fixes.

## Commands, toolchain and retained output

Before setup: export GOPROXY='https://proxy.golang.org|direct'. Setup succeeds;
source /workspace/adamic-tools/env.sh. Its cumulative timing lines are Node
0.035s, Go 0.054s, markdown 0.097s, submodules 0.141s, clang 0.287s, go build 38.077s,
done 38.204s. nproc 5, cpu.max 400000 100000, Node 24.19.0, Go 1.27.1, clang 20.1.8.
Raw setup output is evidence/finished/setup.log.gz.

Ran stock tsc.js -p TREE/src/compiler --noEmit per family and final;
family-proof.cjs, constructor-proof.cjs, copy-proof.cjs, json-proof.cjs,
data-proof.cjs, guard-proof.cjs, inventory.cjs, residue-proof.cjs and
timer-type-proof.cjs with their recorded tree inputs; full latent census plus
census-report.py on each control/restored-any source; npm run build for the
JSON ownership repair; bash stage3/lane/run.sh for the final fresh apply and
full oracle. Earlier any/void/debugger/oracle-fixture checks remain in their
original evidence and c6ea4131 history. All test/measurement output goes to
logs, never a pipe. Family docs and README give exact replay commands.

Evidence is under evidence/{timers,constructors,copies,json-config,locals,data,
finished}. Full raw census streams and large/ANSI outputs are compressed;
small JSON receipts remain directly readable. No additional per-family push
follows the user’s rule change. The finished evidence is committed and then
pushed once to the assigned branch; no PR, force push or main push.
