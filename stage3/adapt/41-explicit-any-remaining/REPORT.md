Built: partial adaptation 41 removes 11 explicit-any tokens and all 15 void expressions in nine source files; no compiler edits.
Commits: any 07aacdf9; debugger decision 79674909; return handoff a4978dfd; void 760ccaa6; call handoff d96ef1e0; dropped kinds ffc600a8.
Commands/results: apply and builds exit 0; main/final oracle 106,366 passing, one sanctioned failure, zero pending; lane PASS; tokens 135 to 124, direct any 37 to 36, void refusals four to zero.
Mutants: restored scanner any and restored source void caught by census; wrong scanner state, callback return, deleted debugger, any-return control artifact mutations and a real fixture-input mutation caught by their dedicated checks.
Not completed: 124 explicit-any tokens, 36 observed any blockers, complete return/call owner work and its call-specific mutant; debugger retained; variance untouched; other refusal fixes listed as candidates.

## Scope and pins

Started from fetched origin/main ef3141e9b1152ab51b51497f8ce3a2799449c8a3 on
codex/stage3-explicit-any-2. The requested dc6b1529 table is capitalized
stage3/notyet-table/TABLE.md. It has seven adaptation rows including variance;
variance is excluded by the addition. With is mentioned in the instruction but
has no compiler-source syntax site or table row. No compiler files, shared lane,
shared census, oracle harness, baseline reference or counts.md are changed.

The final patch table records 41 as nine files, 25 lines added and 25 removed.
Total composition remains 78 modified upstream files, 5,153 added and 5,127
removed lines. New source-tooling files follow adaptation 40's CJS/Python pattern;
no new TypeScript or Adamic fixture is introduced.

## Measurements

All latent observations are **measured on a checker-rejected program**. The tool
skips diagnosed bodies and returns the first lowering error per eligible unit.
It does not produce usable IR or establish whole-program acceptance. The table's
historical 85 any-value observations are not reproduced on this main/tool pair.
The source census independently parses every AnyKeyword, including skipped code.

| Measurement | Main | Any-only | Final |
| --- | ---: | ---: | ---: |
| Explicit any tokens | 135 | 124 | 124 |
| Files containing explicit any | 23 | 22 | 22 |
| Direct any lowering blockers | 37 | 36 | 36 |
| Void syntax expressions | 15 | 15 | 0 |
| Void refusals observed | 4 | 4 | 0 |
| All NotYet | 1468 | 1468 | 1468 |
| All Refused | 5162 | 5161 | 5157 |

Direct any means exact reasons a value of type any, an array of any, a function
returning any, a call returning any, and storing any in a field. It excludes
variance rows merely containing the word any. Main has 30 value observations,
six function-return observations and one store. Direct any is reported as
NotYet, not Refused, by current main. No direct any Refused row is observed.

Main's direct-any files and counts:

| File | Sites |
| --- | ---: |
| commandLineParser.ts | 21 |
| utilities.ts | 4 |
| sourcemap.ts | 3 |
| sys.ts | 3 |
| emitter.ts | 1 |
| moduleNameResolver.ts | 1 |
| moduleSpecifiers.ts | 1 |
| resolutionCache.ts | 1 |
| scanner.ts | 1 |
| tsbuildPublic.ts | 1 |

Only scanner's direct-any observation disappears. Other edited tokens are behind
prior blockers, type-only declarations or casts. A one-token reduction in a
source annotation is not equated with a one-site lowering reduction.

All before/after sites, reason rankings, source hashes and raw compressed streams
are in evidence/. REMAINING.md lists each of the 124 current token locations with
its historical owner analysis. These are unfinished proofs, not 124 impossibility
claims. The unresolved groups include JSON/config 38, timers 16, staged
constructors 11, enumerable copies 10 and correlated AST/callback/generic owners.

## Behavior and artifacts

The any-only upstream build and stock TypeScript 6.0.3 check have zero diagnostics.
The independent proof reconstructs all 79 compiler source files, compares each
whole-file JavaScript emission for six changed files, compares all ten built
JavaScript artifacts exactly, and independently projects all 715 declarations.
Only four internal declaration artifacts change through the recorded owner types.
The public API is byte-identical. No unknown substitution or new cast is added.

The void proof checks eight call statements and seven arrow bodies against Node,
including call order, object writes and undefined callback results. A complete
upstream oracle then compares the final compiler to main through its actual
regression, conformance, project, fourslash, transpile and unit suites.

Both full oracle runs use five workers, all runners and no filter on Node
v24.19.0. Main takes 438.809 seconds, including install/build; its tests take
412.495 seconds. The final report retains its own measured timing. Both have
106,366 passing, one failing and zero pending. Only the existing sanctioned
api/typescript.d.ts acknowledgement differs, with identical baseline.diff bytes.
The standalone oracle correctly exits 1 for that failure; it is not called green.
The landing lane recognizes exactly this sanctioned result and exits 0 with PASS.
It takes 531.884 seconds and records all 222 sanctioned declarations, with 28
already accepted reference declarations. Its adapted compiler sources are
byte-identical to the independently measured final tree, all 79 files.

The lane started while void was present in the worktree before its separate
commit, so execution.json records 07aacdf9. The preserved final source hashes and
patch table establish which bytes it tested. No subsequent adapter change was
made; later changes strengthen evidence and document limits.

## Every mutant and check

| Mutation actually run | Dedicated observation |
| --- | --- |
| Restore _state: any in an isolated real scanner.ts | Direct any 36 to 37, exactly scanner.ts:952:103 |
| Change scanner's state to string through the stock checker host | TS2345 at existing callers scanner.ts:958 and :962; adapted check has zero errors |
| Restore the real void arrow at moduleNameResolver.ts:1838:35 | Zero to exactly one void census refusal |
| Make the actual extracted diagnostics callback return push's result | Node comparison detects a numeric result instead of undefined |
| Delete debugger from the actual extracted Debug.fail body | Attached Node inspector pauses once before and zero times after; error message unchanged |
| Restore any in the otherwise compiled number-return scratch control | Current compiler rejects a function returning any; this is not a completed tsc owner proof |
| Add bytes to a real copied emitted JavaScript artifact | Exact artifact byte comparison fails |
| Remove a build artifact from the compared file set | Exact artifact file-set comparison fails |
| Append an unrelated declaration to a real copied public API artifact | Exact declaration byte comparison fails |
| Append number = string to actual forInStatement1.ts test input | Filtered control six passing; mutant three passing/three failing, four baseline differences; exact restoration six passing with zero differences |
| Existing census audit's signature/body-range, extra-NotYet and attribution mutants | All caught by the unchanged census audit, whose log is retained |

Adapter second passes change zero source bytes. An initial wrong Array.at contract
and broad ProgramHost<BuilderProgram> contract produce four stock diagnostics;
those edits are rejected. The final host helper carries T extends BuilderProgram
and no caller statement changes. An initial VM proof lacked its result binding;
that harness mistake was corrected before recording the passing result.
One premature census raced with tree preparation and measured pre-void code;
only the stable rerun is retained as final evidence.

## Commands actually run

All measurement and test output was redirected to named logs. Full reports and
important logs are committed under evidence/. Scratch trees and full build logs
remain at /workspace/adaptation41-* and /tmp/adaptation41-*.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/adaptation41-setup.log 2>&1
source /workspace/adamic-tools/env.sh
bash stage3/apply.sh /tmp/adaptation41-main > /tmp/adaptation41-main-apply.log 2>&1
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/adaptation41-overlay > /tmp/adaptation41-overlay.log 2>&1
gofmt -w /tmp/adaptation41-overlay/*.go
go build -buildvcs=false -overlay=/tmp/adaptation41-overlay/overlay.json -o /tmp/adaptation41-census ./stage3/census/latent/tool > /tmp/adaptation41-census-build.log 2>&1
python3 stage3/census/latent/audit.py /tmp/adaptation41-census > /tmp/adaptation41-census-audit.log 2>&1
LATENT_ASSERT_NO_OUTPUT=1 /tmp/adaptation41-census /tmp/adaptation41-main/src/compiler /tmp/adaptation41-before.jsonl > /tmp/adaptation41-before-census.log 2>&1
bash stage3/apply.sh /workspace/adaptation41-any-tree > /tmp/adaptation41-any-apply.log 2>&1
npm ci --prefix /workspace/adaptation41-any-tree --no-audit --no-fund > /tmp/adaptation41-any-install.log 2>&1
npm run build --prefix /workspace/adaptation41-any-tree > /tmp/adaptation41-any-build.log 2>&1
NODE_PATH="$HOME/.cache/adamic-stage3/api/node_modules" node stage3/adapt/41-explicit-any-remaining/prove.cjs /tmp/adaptation41-main /workspace/adaptation41-any-tree stage3/adapt/41-explicit-any-remaining/evidence/any-proof.json > /tmp/adaptation41-any-proof-final.log 2>&1
LATENT_ASSERT_NO_OUTPUT=1 /tmp/adaptation41-census /workspace/adaptation41-any-tree/src/compiler /tmp/adaptation41-any-after.jsonl > /tmp/adaptation41-any-after-census.log 2>&1
LATENT_ASSERT_NO_OUTPUT=1 /tmp/adaptation41-census /workspace/adaptation41-any-mutant/src/compiler /tmp/adaptation41-any-mutant.jsonl > /tmp/adaptation41-any-mutant-census.log 2>&1
NODE_PATH="$HOME/.cache/adamic-stage3/api/node_modules" node stage3/adapt/41-explicit-any-remaining/void-proof.cjs /workspace/adaptation41-final-tree > /tmp/adaptation41-void-proof.log 2>&1
LATENT_ASSERT_NO_OUTPUT=1 /tmp/adaptation41-census /workspace/adaptation41-final-tree/src/compiler /tmp/adaptation41-final-stable.jsonl > /tmp/adaptation41-final-stable-census.log 2>&1
LATENT_ASSERT_NO_OUTPUT=1 /tmp/adaptation41-census /workspace/adaptation41-void-mutant/src/compiler /tmp/adaptation41-void-mutant.jsonl > /tmp/adaptation41-void-mutant-census.log 2>&1
NODE_PATH="$HOME/.cache/adamic-stage3/api/node_modules" node stage3/adapt/41-explicit-any-remaining/debugger-proof.cjs /tmp/adaptation41-main > /tmp/adaptation41-debugger-proof.log 2>&1
NODE_OPTIONS=--max-old-space-size=2048 bash stage3/oracle/run.sh /tmp/adaptation41-main /workspace/adaptation41-main-oracle > /tmp/adaptation41-main-oracle.log 2>&1
NODE_OPTIONS=--max-old-space-size=2048 bash stage3/oracle/run.sh /workspace/adaptation41-final-tree /workspace/adaptation41-final-oracle > /tmp/adaptation41-final-oracle.log 2>&1
NODE_OPTIONS=--max-old-space-size=2048 bash stage3/lane/run.sh /workspace/adaptation41-lane > /tmp/adaptation41-lane.log 2>&1
```

The independent real-input oracle control and mutant use one worker and only
compiler tests matching forInStatement1. The original and restored control have
six passing tests and no baseline differences. The input mutant has three
passing and three failing tests, with errors, JavaScript, symbols and types
baseline mismatches. Install/build exit zero, and only the oracle kills it.
The fixture is restored byte for byte. Its SHA-256 and phase reports are retained.

```sh
NODE_OPTIONS=--max-old-space-size=2048 bash stage3/oracle/run.sh /workspace/adaptation41-final-tree /workspace/adaptation41-oracle-control --runners=compiler --tests=forInStatement1 > /tmp/adaptation41-oracle-control.log 2>&1
NODE_OPTIONS=--max-old-space-size=2048 node stage3/adapt/41-explicit-any-remaining/oracle-proof.cjs /workspace/adaptation41-final-tree /workspace/adaptation41-oracle-control /workspace/adamic/stage3/oracle/run.sh /workspace/adaptation41-oracle-mutant > /tmp/adaptation41-oracle-mutant-proof.log 2>&1
```

The stock checker probes run through inline Node scripts with the compiler's
complete config and actual callers; their diagnostics are in stock-check.log.txt
and rejected-stock-check.log.txt. Each minimal kind probe uses
`go run ./cmd/adamic c /tmp/adaptation41-kind-probes/KIND.a`, with its own log.
Python census-report.py independently deduplicates each raw stream and records
exact reason/file/site ledgers. Targeted node --check and git diff --check pass.

Setup succeeds without a workaround: Node 0.022s; Go 0.034s; clang 0.277s;
markdown dependencies ready 1.013s; submodules 15.065s; Go build 258.949s;
cache warm 259.050s; done 259.086s. nproc=5; CPU quota=4. The printed environment
/workspace/adamic-tools/env.sh is sourced in build/test shells.

## Ranked other source edit candidates

The complete 124 exact refusal rows representing 328 observed sites, including
all locations, are in RANKED.md. These are source-level adaptation candidates;
this unit does not establish a completed owner/caller proof for any of them.

1. Method receiver captures: 195.
2. Explicit boolean conditions: 78.
3. Definite assignment assertions: 11.
4. Namespaces: 11.
5. Index signatures: 11.
6. Parameter properties: 6.
7. var declarations: 6.
8. Generator functions: 5.
9. Yield expressions: 5.

There are no distinct observed declared-expando-field or missing-return-type
refusal reasons. The 1,396 unchecked casts are mixed and require per-site review;
they are not all presented as source-only fixes. Variance is left untouched.
Non-null, comma, logical-assignment and predicate-verification rows remain
compiler lessons under the supplied table rulings.

## Remaining boundary decisions and limits

DEBUGGER.md gives the actual source behavior counterexample and a minimal program
for @system_adamic_typescript. ANY-RETURNS.md and ANY-CALLS.md give minimal JSON,
staged initialization and generic timer contract programs, while explicitly
identifying unfinished work. DROPPED.md retains actual main compilation receipts
for any-array storage and the checker-rejected, absent with statement.

This is not completion of the requested remaining-any adaptation. No claim is
made that the 124 remaining tokens have no truthful static type. Thirty-six
observed any blockers remain, including all six return owners; no dedicated
call-return mutant or completed call-return owner is provided. Native tsc
execution, source-map/build-metadata identity, Windows host behavior, arbitrary
custom-host inputs and the full Go gate are not covered. No fixture is added,
so counts.md is unchanged. No pull request is opened.
