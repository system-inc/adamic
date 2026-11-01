# 95d: measure three checker recovery reads

Only checker.ts:12034 is adapted. The first CommonJS export declaration is
required at its producer; one `!` removes exactly one diagnostic and erases
on Node. Rows 42849 and 50153 both have real absences. Their explicit-branch
proposals remain unbuilt and require a Correctness ruling.

lane: run as part of a Correctness wave

Base: `origin/typescript/axolotl-checker-a`,
`ef4d26a4a38b7890317e1ba71f8ad928fd8588f6`, containing `d5c15cbf` and 95a's
survey correction. The branch was created from fetched main first, then based
on 95a as the head-specific instruction requires. Pinned TypeScript 6.0.3:
`050880ce59e30b356b686bd3144efe24f875ebc8`. Only this adaptation directory is
added. No public API line, lane, other adaptation, patch table, fixture registry,
meter or Go source is edited. The patch table under evidence is generated output.

Proof artifacts, fixtures and supporting scripts live on the separate
`typescript/axolotl-checker-d-evidence` branch under
`stage3/evidence/95-checker-axolotl-d/`. This adaptation directory contains
only adapt.cjs, verify.cjs, sites.json and README.md; every file here feeds
apply's input key. The commands below refer to this evidence checkout.

## Runtime measurements

The built adapted 95a compiler is instrumented in scratch by
[evidence/observe.cjs](../../evidence/95-checker-axolotl-d/observe.cjs). Its Node compile hook inserts a
value-preserving observation at each original read, leaving the built files
on disk and all compiler source unchanged. Every patch spelling must occur
exactly once per loaded bundle. It records present/undefined, fixture or test
context, relevant source metadata, and stacks. Process-exit files cover all
workers; zero-hit processes and patched bundle addresses are retained in
[evidence/oracle-observations.json.gz](../../evidence/95-checker-axolotl-d/oracle-observations.json.gz).
No instrumented compiler is committed. The hook is not used by adapt.cjs.

| Read | Compiler/conformance present / absent | Final fixtures present / absent | Total present / absent | Disposition |
| --- | ---: | ---: | ---: | --- |
| 12034 file | 7 / 0 | 4 / 0 | 11 / 0 | Required first declaration at producer |
| 42849 relevantTypeParameter | 0 / 0 | 6 / 8 | 6 / 8 | Explicit absence proposal only |
| 50153 first declaration in fallback | 0 / 0 | 2 / 2 | 2 / 2 | Explicit absence proposal only |

[evidence/measurements.json](../../evidence/95-checker-axolotl-d/measurements.json) contains these exact
counts. Counts exclude exploratory runs and proposal runs. Each absence is
attributed in [evidence/fixture-observation.json](../../evidence/95-checker-axolotl-d/fixture-observation.json):

- 42849: excess.js, arguments two and three of C<string, number, boolean>,
  in JSDoc @type and @extends, with noImplicitAny false and true. Eight hits;
  Node throws TypeError reading `resolvedBaseConstraint`.
- 50153: consumer.ts importing exported.js, querying v.xyz on an augmented
  CommonJS export with a computed static index, with noImplicitAny false and
  true. Two hits; Node throws TypeError reading `parent`. Neither fixture has
  compiler diagnostics. The declared-index control reaches this read twice
  with a first declaration present.

The ordinary compiler runner does not exercise either API query read. The
fixture runner explicitly walks TypeNodes, including JSDoc, and asks the built
checker's getTypeArgumentConstraint and getSymbolAtLocation APIs. Raw fixture
sources are `.ts.txt` and `.js.txt`, not Adamic programs. All 14 fixture program
observations and their diagnostics/queries are retained. Search attempts and
unreached shapes are in [evidence/exploration.md](../../evidence/95-checker-axolotl-d/exploration.md).

Per-read notes name the observation, code explanation, and disposition:
[12034](../../evidence/95-checker-axolotl-d/12034.md), [42849](../../evidence/95-checker-axolotl-d/42849.md), [50153](../../evidence/95-checker-axolotl-d/50153.md).

## Required-value invariant and adapter

At 12034 the only getFlowTypeFromCommonJSExport caller is guarded by
exportSymbol?.declarations and isDuplicatedCommonJSExport of that same list.
The predicate requires length greater than one and every declaration to be a
bound JavaScript export access. The binder's dense declaration array contains
parser nodes with parent chains to their source files. Thus its first
Declaration is present and getSourceFileOfNode returns its attached SourceFile.
The alias's separate declaration-length test is not the export-list proof.

The adapter adds `!` to symbol.declarations![0] at the existing producer call.
It selects the nonoptional overload; it does not assert the optional alias
file, and it does not require an endFlowNode to exist. Each expression keeps
its original evaluation point. No default, optional chain, cast or public
signature change is introduced.

Following 95a, 31 and 91, stock TypeScript 6.0.3 through 47-host-errors/compiler
parses and binds the current tree. The adapter requires the private nested
owner, exact indexed expression, the owner parameter's Symbol type and binding,
the read's Declaration | undefined type, the resolved utilities.ts producer,
its before/after overload type, the single symbol-bound caller and its guard,
and the exact duplicate-export predicate. It pins only the owner function's
normalized text with ownerBeforeSha256
`57faba456412535bb02760aad9cdbcf7569cc83624729da3f6ddf40b6ac5a80c`.
There is no whole-checker hash gate, so independent 95 heads can compose.

All edits are planned, parsed and compared as stock JavaScript before the one
write. The input is reread before writing, no printer is used, and all 54,434
CRLFs are preserved. The second run reports files 0, assertions 0. --check
writes nothing. The independent verifier compares all 711 source files and
requires exactly the single parsed producer assertion, with unchanged prefix
and suffix; it does not call the adapter planner.

The adapter applies on a tree with 95a alone and on fresh full apply. At both
remote-ref checks, origin exposed only the A branch; local branches contained
only D. B and C composition cannot be run against nonexistent refs. D on D is
covered by the zero-edit reruns. No claim is made about unprovided B/C code.

## Diagnostic, source and mutant proofs

The production census tool loads only src/tsc/tsc.ts using internal/load options.
Following 95a's compare-census.py, compare-census.py normalizes root paths,
requires root and aggregate observations to match, counts one, and counts
checker diagnostics only by the primary file location. Full sets are compared,
not only totals. The dependency-free control exactly reproduces 95a's complete
after set: [evidence/reproduction.json](../../evidence/95-checker-axolotl-d/reproduction.json).

| Tree | Whole program | checker.ts own file |
| --- | ---: | ---: |
| 95a control | 131 | 17 |
| 95d after | 130 | 16 |
| One-edit reversion mutant | 131 | 17 |

Removed: `src/compiler/checker.ts:12034:30 TS18048: 'file' is possibly 'undefined'.`
Added anywhere: **0**. [evidence/census.json](../../evidence/95-checker-axolotl-d/census.json) and compressed
raw census observations retain the complete proof. Installing upstream npm
dependencies removes the unrelated existing sys.ts:1598:69 TS2307 for
source-map-support; this explains the initial post-install baseline 130/17,
not a 95d adaptation change. The table uses dependency-free controls throughout.

The mutant undoes only this producer `!` in an isolated after copy. It restores
exactly row 12034 and nothing else. The independent verifier exits 1 with
`exact required declaration read missing or uninventoried edit`; the adapter's
--check exits 1 with `required declaration read missing`. The census comparison
requires that exact restored finding and no other difference. No mutant enters
an oracle tree. Raw failures and exit codes are under evidence.

Stock transpileModule checker JavaScript is byte-identical in both modes:

| Comments | Bytes | SHA-256 before and after |
| --- | ---: | --- |
| Retained | 2,892,419 | `786ee138ff00765ab7c7ea7651de48df32c5981ec2c0810c68e302570d566f6e` |
| Removed | 2,354,716 | `3664a2f5fb2aba0e839a3153e6ff4d093b530db890b5b2433be6cd38e6e91aac` |

[evidence/source.json](../../evidence/95-checker-axolotl-d/source.json) records the first comparison;
[evidence/final-source.json](../../evidence/95-checker-axolotl-d/final-source.json) independently checks the
fresh full apply output against the same control. Both have one changed file
out of 711 and exactly one assertion. Fresh apply exits 0, with incremental
row `95-checker-axolotl-d | 1 | 1 | 1`; its rerun also edits nothing.

## Unbuilt JavaScript proposals

[evidence/proposed-42849.diff.txt](../../evidence/95-checker-axolotl-d/proposed-42849.diff.txt) explicitly
throws on an absent type parameter.
[evidence/proposed-50153.diff.txt](../../evidence/95-checker-axolotl-d/proposed-50153.diff.txt) retains the
ordered aliasSymbol/symbol/fallback branches, checks a saved first declaration
only in the fallback, and explicitly throws when it is absent. Each preserves
the observed caught TypeError name and message instead of replacing the crash
with a result. Neither proposal is part of adapt.cjs or any built source tree.

[evidence/proposal.cjs](../../evidence/95-checker-axolotl-d/proposal.cjs) evaluates those branches in memory
in the existing Node build, without building or writing a compiler file.
Every diagnostic and query result in all 14 final fixture programs is equal
before and after, including all eight excess-argument errors, both empty-list
errors, and the populated-list controls. Exact outputs and comparison scope are
in [evidence/node-comparison.json](../../evidence/95-checker-axolotl-d/node-comparison.json). Throw sites and
stack frames move: uncaught stderr/stack identity is not claimed. These are
reviewable proposals pending a ruling, not approved behavior-preserving native
adaptations. No native build, diagnostic-removal claim or oracle acceptance is
made for them.

## Commands and environment

Every build/test shell sources /workspace/adamic-tools/env.sh. Setup succeeded:
Node 0.021s, Go 0.033s, clang 0.311s, tsc 0.567s, markdown ready 1.552s,
submodules 167.139s, Go build 440.773s, cache warm 440.897s, done 440.933s.
nproc=5; cpu.max=400000 100000. Node v24.19.0, Go 1.27.1, clang 20.1.8.
The first apply attempt failed because stage3/api/node_modules was absent.
`npm ci --prefix stage3/api --no-audit --no-fund` supplied the locked parser;
retry passed without changing any tracked dependency file. Setup log is retained.

All test output goes to log files. Principal commands, from the repository root:

```sh
source /workspace/adamic-tools/env.sh
export NODE_PATH="$PWD/stage3/api/node_modules"
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/95d-upstream
bash stage3/apply.sh /workspace/scratch/95d-before > /tmp/95d-apply-before.log 2>&1
# The before run used 95a alone. Make dependency-free control/after copies.
go build -o /workspace/scratch/95d-census ./stage3/census/tool > /tmp/95d-census-build.log 2>&1
/workspace/scratch/95d-census CONTROL/src/tsc/tsc.ts BEFORE.jsonl > /tmp/95d-control-census.log 2>&1
node stage3/adapt/95-checker-axolotl-d/adapt.cjs AFTER > /tmp/95d-adapt.log 2>&1
node stage3/adapt/95-checker-axolotl-d/verify.cjs CONTROL AFTER > /tmp/95d-verify.log 2>&1
node stage3/adapt/95-checker-axolotl-d/adapt.cjs AFTER > /tmp/95d-idempotence.log 2>&1
/workspace/scratch/95d-census AFTER/src/tsc/tsc.ts AFTER.jsonl > /tmp/95d-after-census.log 2>&1
# Copy AFTER to MUTANT and undo only getSourceFileOfNode(symbol.declarations![0]!).
node stage3/adapt/95-checker-axolotl-d/verify.cjs CONTROL MUTANT > /tmp/95d-mutant-verify.log 2>&1
node stage3/adapt/95-checker-axolotl-d/adapt.cjs --check MUTANT > /tmp/95d-mutant-check.log 2>&1
/workspace/scratch/95d-census MUTANT/src/tsc/tsc.ts MUTANT.jsonl > /tmp/95d-mutant-census.log 2>&1
python3 stage3/evidence/95-checker-axolotl-d/compare-census.py BEFORE.jsonl AFTER.jsonl MUTANT.jsonl > /tmp/95d-census-comparison.json 2>&1
bash stage3/apply.sh /workspace/scratch/95d-fresh > /tmp/95d-apply-fresh.log 2>&1
```

To reproduce the control from a fresh current apply, copy the tree without
node_modules/built, and restore checker.ts from its
`refs/stage3/95-checker-axolotl-d-before` Git tree. This leaves 95a in place.
Do not rerun 90 on a tree already edited by 95.

```sh
export OBSERVE_TREE=/workspace/scratch/95d-before
mkdir -p /workspace/scratch/95d-observations
OBSERVE_OUT=/workspace/scratch/95d-observations \
NODE_OPTIONS="--require=$PWD/stage3/evidence/95-checker-axolotl-d/observe.cjs" \
timeout --kill-after=20s 880 bash stage3/oracle/run.sh "$OBSERVE_TREE" /workspace/scratch/95d-oracle-measure --runners=compiler --limit-seconds 720 > /tmp/95d-oracle-measure.log 2>&1
mkdir -p /workspace/scratch/95d-clean-observations
OBSERVE_OUT=/workspace/scratch/95d-clean-observations \
NODE_OPTIONS="--require=$PWD/stage3/evidence/95-checker-axolotl-d/observe.cjs" \
node stage3/evidence/95-checker-axolotl-d/fixtures.cjs "$OBSERVE_TREE" > /tmp/95d-fixtures-clean-before.json 2>&1
NODE_OPTIONS="--require=$PWD/stage3/evidence/95-checker-axolotl-d/proposal.cjs" \
node stage3/evidence/95-checker-axolotl-d/fixtures.cjs "$OBSERVE_TREE" > /tmp/95d-fixtures-clean-proposed.json 2>&1
timeout --kill-after=20s 880 bash stage3/oracle/run.sh /workspace/scratch/95d-fresh /workspace/scratch/95d-oracle-final --limit-seconds 720 > /tmp/95d-oracle-final.log 2>&1
```

The instrumented compiler/conformance run passed in 213.414 seconds:
install 5.985s, build 24.757s, tests 182.660s; **89,838 passing, 0 failing,
0 pending**, no baseline differences. This compiler-only count is not the
all-runners baseline of 106,366 passing and one standing Public APIs failure.
The full fresh-tree oracle result follows.

Native compiler execution, the full repository Go gate, exhaustive arbitrary
checker API mutations, and composition with unavailable B/C refs are not covered.
Only the one invariant-backed assertion is shipped.

## Completed full oracle

The fresh-tree standalone all-runners oracle completed within the 880-second
hard bound in **350.495 seconds**. Install exited 0 in 2.409s, build exited 0
in 24.999s, and tests exited 1 in 322.998s. Counts are **106,366 passing,
1 failing, 0 pending**, exactly the supplied main and 95a baseline. Raw status
is fail, solely for `unittests:: Public APIs for typescript.d.ts should be
acknowledged when they change`. Baseline differences: one existing API file;
new baseline differences: **0**. Timeouts and exceptions: **0**.

[evidence/oracle-final-summary.json](../../evidence/95-checker-axolotl-d/oracle-final-summary.json) checks
both declaration artifacts against main's recorded bytes:

- Built/local API: 590,996 bytes, SHA-256
  `b901adcc6ec29e299ec21401b2a155b2b2c14c046ec743fed6ee45679b726d5d`.
- Accepted reference API: 588,626 bytes, SHA-256
  `f87d6c30da287293aae741711fa94594b75d49e0977cea6bfd5982f90e2bafcf`.

The actual local baseline equals the built declaration byte for byte. The full
51,123-byte baseline diff has SHA-256
`6d52cd09811c6846c977ea492071c14325816a631706cc79128c4299b9faa031`,
exactly 95a's recorded diff. The full report and compressed phase logs/diff are
retained beside the summary. This head introduces no public API delta.

The post-install fresh census is **129 whole-program, 16 checker own-file**,
compared with **130 and 17** on the installed 95a tree. It independently removes
only row 12034, with nothing added. The dependency-free 131 to 130 comparison
above reproduces 95a's original environment exactly. Both complete paired sets
are retained in [evidence/installed-census.json](../../evidence/95-checker-axolotl-d/installed-census.json)
and its compressed observations.
