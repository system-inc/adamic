#!/usr/bin/env python3
"""Render the final receipts without mixing historical transformation scopes."""
import json
from pathlib import Path
p=Path(__file__).resolve().parent
def read(name):return json.loads((p/name).read_text())
replay=read('owned/paired-replay/status.json');assert replay['complete'] and not replay['differences'] and replay['baseline_adapters_compared']==80
assert read('owned/types.json')['diagnostics']==0
assert read('owned/javascript/comparison.json')['pass']
assert not read('owned/census/comparison.json')['changed']
union=read('owned/oracle/union.json')
assert union['status']=='pass' and union['units']==398
whole=read('owned/oracle/whole/report.json')
assert whole['status']=='pass' and whole['counts']==dict(passing=106366,failing=1,pending=0)
latent=read('owned/latent/comparison.json')
identity=read('owned/oracle/source-identity.json')
# The producer and standard checkers, rather than this renderer, certify the run.
proof='''# 99x proof: 28 rewrites, 18 earlier-owner handoffs

The current ownership revision passes checks 1 through 6. All current receipts
are under owned/. Receipts outside owned/ describe earlier 46-site experiments
and superseded stopping rules; they do not certify this scope.

javascriptChanging true. No earlier adaptation directory is edited. The exact
46-site inventory and every handoff reason are in README.md and inventory.json.
Six sites belong to 95-checker-axolotl-h: two checker-host calls and four calls
under program.ts createProgram/getEmitHost. Twelve belong to the complete
watch-host runtime pins in 32-indexed-reads-program. Their adapter directories
remain untouched. The runtime-input audit and ownership.json record these pins.
The current adapter edits ten owners in five files and removes three imports.
core.ts, binder.ts and checker.ts source are byte-identical to the baseline.

## Results

| Check | Current result |
| --- | --- |
| 1 full applies with --build-product | Before and after exit 0 |
| 1 paired replay | All 80 baseline outcomes identical; no differences |
| Direct replay | Zero writes and edits; bytes and nanosecond mtimes preserved |
| Owner contracts | Six malformed states refuse before writes; CRLF and unrelated edits pass |
| 2 types | All 79 roots together, exit 0 and zero diagnostics |
| 3 emitted JavaScript | Nine products pass; seven byte-identical, two with exactly 28 rewritten sites each |
| 4 upstream oracle | Whole lane and all 398 standard shards pass; their exact observation union and API bytes agree |
| 5 negative control | Exactly one restored source site detected, in both compiler bundles |
| 6 ordinary census | All 79 first stops unchanged; 73 stop at core.ts:810:89 |
| 6 CW051 | All six units measured on both trees; table below |

## Provenance and full applies

Base main: 3e395f9e5a7038d859ec8930be72cde68d8be3ff.
Upstream TypeScript 6.0.3: 050880ce59e30b356b686bd3144efe24f875ebc8.
99w's complete adapt.cjs, README.md and PROOF.md were read at
06dad75d0bb482b47046370a60341c054899afed. 99k was read on main. The branch
remains based on main, never 99w. No Apple compiler overlay was used.

cloud/setup.sh and npm ci --prefix stage3/api completed after sourcing
/workspace/adamic-tools/env.sh. Setup did not supply the upstream source,
so the authorized pinned checkout was cloned to /workspace/typescript-source.
Setup and package logs are setup.log.gz and npm-api.log.gz. Both complete
applications used --build-product:

```sh
# Isolated unmodified main worktree /workspace/bind-main:
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source stage3/apply.sh --build-product /workspace/bind-before
# This branch with the current adapter:
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source stage3/apply.sh --build-product /workspace/bind-owned-after
NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-x/contract.cjs /workspace/bind-before
```

baseline-apply.log.gz records the unchanged before build. owned/after-apply.log.gz
records the current after build. owned/source-delta.json confirms the five
changed files. owned/contract.log.gz records direct replay, byte/mtime preservation,
atomic refusals and CRLF checks. owners.json contains exact before/after texts
and SHA-256 values for each edited owner, never whole-file hash gates.

## Check 1: all adapter outcomes

paired_replay.py runs the real Python order over independent copies of both
complete adapted trees. owned/paired-replay/before.json and after.json each
contain 81 records: 80 baseline adapters and the separately handled 99x entry.
Every baseline outcome matches: exit status, files written, exact edits and
exact refusal text. Only tree root paths are normalized. records.diff is empty.
No early stopping is used. 99x's after replay exits 0, writes zero files and
makes zero edits; its 18 handed-off calls remain inventoried.

Shared refusals are baseline outcomes. In particular 40-explicit-any reports
`Error: missing reviewed brands: src/compiler/builder.ts` on both trees.
30-indexed-reads reports `declined site changed: core.ts:992:21` on both.
The complete exact text and all other outcomes are retained in the records
and separate compressed adapter logs, including 95h's shared program owner
refusal. None is treated as a stop when identical on both trees.

A passive filesystem journal records sync and async writes, including unchanged
byte writes and inherited child-process writes. Whole-tree byte/mtime manifests
before and after verify zero unjournaled changes. owned/journal/ retains the
control and runtime write audit; owned/paired-replay/ retains every journal,
manifest and coverage.json. This changes observation cost, not adapter behavior.

## Check 2: 79-root types

The unchanged main Adamic binary receives all 79 sorted compiler TypeScript
roots in one invocation. It exits 0 with zero diagnostics. owned/types.json
records every argument and timing; owned/types.log.gz is the complete output.
No type-driven fallback or cast is needed for the 28 remaining rewrites.

## Check 3: emitted products

Both ordinary npm run build:compiler builds exit 0 with the same locked
upstream dependencies. owned/javascript/ contains every raw diff, build log,
comparison report and declaration hash. Both public and internal declaration
files are byte-identical. Seven normal JavaScript products have no changes.
_tsc.js and typescript.js each contain the same 28 expected source-site changes.

The normal emitter lowers each optional binding into a nullish conditional
and a generated temporary. compare.cjs uses stock TypeScript 6.0.3 to match
only inventoried owner/member pairs with receiver/temporary symbol identity.
It reconstructs those expressions and normalizes generated temporary names,
their declarations and resulting parentheses. The remaining complete ASTs
must agree. All 18 handed-off core helper calls remain, so the helper stays
in both bundles and binder's local helper name stays unchanged. There is no
helper deletion or binder rename exemption. The complete raw hunks appear below.

## Check 4: unchanged whole lane and all shards

The frozen oracle worktree commit is FROZEN_COMMIT. Its five functional adapter
inputs match this branch byte for byte. All 746 source files in the prepared
artifact match /workspace/bind-owned-after. owned/oracle/source-identity.json
records those assertions. Evidence commits do not alter frozen oracle inputs.

The unchanged standard prepare_shards.py producer completes all apply, locked
install, build and harness phases successfully. Its elapsed time is recorded
in owned/oracle/prepare-metrics.json. The unchanged lane command is:

```sh
stage3/lane/run.sh /workspace/scratch/bind-owned-lane-whole --artifact ARTIFACT
```

The whole lane passes with 106,366 passing, one sanctioned public API failure
and zero pending. All runners, eight workers, the normal 40-second timeout and
main's sanctioned API lines are unchanged. Its complete reports, task
observations, API diff and logs are under owned/oracle/whole/.

All 398 standard manifest shards pass in ten sequential pieces of at most
40 shards each. Each piece runs two standard shards concurrently, preserves
their normal 30-second limit, and is committed and plainly pushed immediately
after completion. Preparation, whole lane and each piece have a 1440-second
outer limit, below 25 minutes. Exact timings are in the piece receipts and this table.

ORACLE_TIMINGS

Shards use main's supported STAGE3_ADAPTED_TREE product path with symlink-only
views of the same artifact and its exact locked dependency installation.
No expected count, runner, API line, timeout or test is changed. Standard
shard.py performs its usual identity, input hash, count, timeout and API checks.
owned/oracle/shards/ retains all unit reports, observations and logs;
owned/oracle/pieces/ retains all ten result lists, driver logs and timings.
Successful redundant scratch views are deleted after their receipts are saved.

check_oracle_union.py independently verifies all 398 standard verdicts,
manifest coverage, runtime and product identity, every test identity/status
multiset against the whole lane, the summed 106,366/1/0 counts, and the exact
sanctioned failure title and paths. Concatenated shard API diff bytes equal
the whole lane's bytes. Main's unchanged check_results also passes.
owned/oracle/union.json is the complete union receipt.

## Check 5: restored-site negative control

The negative fixture restores the builder.ts host.writeFile call in
createBuilderProgram.emitBuildInfo and its required import. The fixture patch
matches the actual source delta. The normal build exits 0, then the identical
JavaScript comparison exits 1 and reports exactly that one unique original
site in both bundles. Each bundle has 27 bound sites and one original call;
the other seven products stay byte-identical. Its dedicated fixture, patch,
build log and comparison receipt are retained under their negative-control
file names, including owned/mutant-comparison.json. No extra change is hidden
by the normalizer.

## Check 6: ordinary and latent census

owned/census/ contains all before/after records, logs, summaries and the
comparison. The unchanged /workspace/bind-before baseline measurement is
reused; the current after measurement was rerun. All 79 single-root TypeScript
first-stop records agree exactly. There are zero changes. Seventy-three stop
at core.ts:810:89, Refused: a cast the runtime cannot check. The other six
records are unchanged too. Extra JSON and whole-program rows are retained
but excluded from the 79-root comparison.

The six CW051 unit identities come from core-walls-sort at
ad29102e9c145c5b480629227574d2ffa235a4b1. The scheduling selector comes from
container-fit-census at 654092b95416580c3e5be1f968fb8b14755b4481. The census-only
overlay uses main's latent harness. Its two hooks select files and top-level
units without changing registration, dependency bodies or first-error behavior.
Full and speculative modes are disabled. Every added Go source ends in .go.txt.

All 12 bounded pieces exit 0, select exactly one requested unit and report
checker_rejected false. File-wide refusal scans remain separate from selected
unit lowering stops. The maximum piece wall time is LATENT_MAX seconds, below
25 minutes. owned/latent/ retains selections, raw streams, logs, timers and
comparison.json. These are the actual current-main locations, all NotYet:

| Unit | Before | After |
| --- | --- | --- |
LATENT_ROWS

The checker unit remains at its method-presence boundary because checker.ts
is unchanged. createWatchHost and createIncrementalCompilerHost move from
core's optional call to direct optional calls, the form Apple's head lowers.
The three command-line units stay at core's optional call because the relevant
watch host owners belong to 32 and remain untouched. No native-lowering
completion is claimed. Historical core-walls line 1523 is current line 1533.

## All emitted JavaScript diff hunks

Seven files have no hunks. These are the complete raw hunks for the other two.
'''
proof=proof.replace('FROZEN_COMMIT',identity['frozen_commit']).replace('LATENT_MAX',f"{max(max(r['before_seconds'],r['after_seconds']) for r in latent):.3f}")
timings=[('Preparation',read('owned/oracle/prepare-metrics.json')['seconds']),('Whole lane',read('owned/oracle/whole/piece-metrics.json')['seconds'])]+[('Shard piece '+str(i),read('owned/oracle/pieces/'+str(i)+'/metrics.json')['seconds']) for i in range(10)]
assert all(seconds<1500 for _,seconds in timings)
proof=proof.replace('ORACLE_TIMINGS','| Piece | Seconds |\n| --- | ---: |\n'+'\n'.join('| '+label+' | '+f'{seconds:.3f}'+' |' for label,seconds in timings))
proof=proof.replace('LATENT_ROWS','\n'.join('| '+r['name']+' | '+r['before']['where']+' | '+r['after']['where']+' |' for r in latent))
for name in ['_tsc.js','typescript.js']:
 proof+='\n### '+name+'\n\n```diff\n'+(p/'owned/javascript'/(name+'.diff')).read_text()+'```\n'
assert '\u2013' not in proof and '\u2014' not in proof
(p/'PROOF.md').write_text(proof)
readme=(p/'README.md').read_text().split('## Current verification')[0]
readme=readme.replace('There are zero unsupported-shape refusals and no type-driven fallback yet.','There are zero unsupported-shape refusals and zero type-driven fallbacks.')
readme+='''## Current verification

Checks 1 through 6 pass for 28 rewrites and 18 earlier-owner handoffs.
All 80 baseline replay outcomes match. Types pass over all 79 roots, nine
emitted JavaScript products match the expected changes, and the one-site
negative control is detected. The unchanged whole lane and all 398 standard
shards pass with exactly matching observations and sanctioned API bytes.
All 79 ordinary first stops are unchanged; all six CW051 units were measured
before and after. See PROOF.md and owned/ for complete current evidence.
Older receipts outside owned/ describe superseded 46-site experiments.
'''
(p/'README.md').write_text(readme)
out=read('outcome.json');out['status']='pass';out['CW051']=latent;out['checks']['4']='pass: whole lane and all 398 shards, exact observation union and sanctioned API bytes';out['oracle_union']=union;out['oracle_whole']=whole
(p/'outcome.json').write_text(json.dumps(out,indent=2)+'\n')
print(json.dumps(dict(proof_bytes=len(proof),rewritten=28,left=18,checks='1-6 pass')))
