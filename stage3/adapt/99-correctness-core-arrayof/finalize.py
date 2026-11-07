from pathlib import Path
import json,hashlib,shutil,subprocess,os
s=Path('/workspace/scratch/arrayof');repo=Path('/workspace/adamic');own=repo/'stage3/adapt/99-correctness-core-arrayof';read=lambda n:json.loads((s/n).read_text());gate=read('RECONCILED.json');pair=read('PAIRED-REPLAY.json');census=read('CENSUS-COMPARISON.json');comparison=read('COMPARISON.json');contract=read('contract.json');types=read('TYPES-COMPARISON.json');checkpoint=read('CHECKPOINT.json');provenance=read('provenance.json')
assert gate['complete'] and gate['counterexamples']==0 and gate['returnedOrThrew']==860;assert pair['pass_'] and pair['adapters']==84;assert not census['newSites'] and len(census['removedSites'])==1 and census['mutantAllSitesEqualBaseline'];assert comparison['changed']==['compiler/core.ts'] and all(d['identical'] for d in comparison['declarations']);assert contract['pass'];assert (s/'apply-before-proof.exit').read_text().strip()==(s/'apply-after.exit').read_text().strip()=='0'
for name in ['checkpoint.py','post-build.py','after-checks.py','pair-replay.py','finalize.py']:shutil.copyfile(s/name,own/name)
# Detect task workers left behind, excluding this finalizer and its ancestors.
excluded={os.getpid()};pid=os.getppid()
while pid>1:
 excluded.add(pid)
 try:pid=int(Path('/proc',str(pid),'stat').read_text().split(') ',1)[1].split()[1])
 except (FileNotFoundError,PermissionError):break
remaining=[]
for p in Path('/proc').iterdir():
 if not p.name.isdigit() or int(p.name) in excluded:continue
 try:
  state=p.joinpath('stat').read_text().split(') ',1)[1].split()[0]
  if state=='Z':continue
  cmd=p.joinpath('cmdline').read_bytes().replace(b'\0',b' ').decode(errors='replace');cwd=str(p.joinpath('cwd').resolve())
 except (FileNotFoundError,PermissionError,ProcessLookupError):continue
 if str(s) in cmd or cwd.startswith(str(s)+'/'):remaining.append(dict(pid=int(p.name),command=cmd,cwd=cwd))
assert not remaining,remaining
(s/'PROCESS-CLEANUP.json').write_text(json.dumps(dict(taskRoot=str(s),remainingWorkers=remaining,allMeasurementBuildCensusReplayWorkersStopped=True),indent=2)+'\n')
with (s/'pack-proof.log').open('w') as log:
 subprocess.run(['python3',str(s/'pack.py'),str(own/'proof.tar.gz'),'COMPARISON.json','compare.log','*.diff','core-*-before.js','core-*-after.js','after-build-phases.json','post-build.log','install-after.log','build-after.log','types-after*.json','types-after.log','roots-after.json','TYPES-COMPARISON.json','types-command.go.txt','after-census*.json*','after-unit-times.jsonl','census-after-driver.log','mutant-*.json*','mutant-unit-times.jsonl','census-mutant-driver.log','CENSUS-COMPARISON.json','compare-census.log','after-checks.log','replay-*.json','replay-*.log','replay-after.exit','PAIRED-REPLAY.json','pair-replay.log','overlay','census-build.log','build-census-driver.log','overlay.log','adamic-build.log','binary-sha256.txt','*-build-info.log','census-tool-reference.diff','CHECKPOINT.json','checkpoint-driver.log','push-checkpoint.log','pack-checkpoint.log','frozen-adapter-sha256.txt','frozen-adapter-check.log','frozen-adapter-final-check.log','apply-after-wrapper-result.json','apply-after.log','apply-after.exit','after-product-key.txt','after-product.manifest.json','checkpoint.py','post-build.py','after-checks.py','pair-replay.py','finalize.py','PROCESS-CLEANUP.json'],stdout=log,stderr=subprocess.STDOUT,check=True)
archives={name:dict(bytes=(own/name).stat().st_size,sha256=hashlib.sha256((own/name).read_bytes()).hexdigest()) for name in ['measurement.tar.gz','checkpoint.tar.gz','proof.tar.gz']}
report=dict(status='pass',task='b5w3ycg',base=provenance['base'],typescript=provenance['source'],ruleClass='JavaScript-changing established-shape, counted domain',ruling='Adamic, task b5w3ycg, Nov 7 03:20Z: established-shape adaptations, Compiler accepts',count=gate,fullApplies=dict(beforeExit=0,afterExit=0,freshBaselineInputsCompared=81504,baselineInputDifferences=0,afterOuterShellStatusRetainedSeparately=True),ownerContract=dict(pass_=True,checks=contract['checks']),pairedReplay=pair,types=types,source=dict(files=comparison['sourceFiles'],changed=comparison['changed'],ownerOnly=True),javascript=dict(completeCoreCommentsKeptAndRemovedOwnerOnly=True,compilerFiles=len(comparison['emitted']),changed=[r['name'] for r in comparison['emitted'] if not r['identical']],allChangesOnlyArrayOf=True),publicDeclarations=comparison['declarations'],census=census,checkpoint=checkpoint,archives=archives,processes=read('PROCESS-CLEANUP.json'))
(own/'PROOF.json').write_text(json.dumps(report,indent=2)+'\n')
(own/'README.md').write_text('''# arrayOf builds its result by push

JavaScript-changing established-shape adaptation for task b5w3ycg, site F2 at
adapted core.ts:1339:12. The complete arrayOf owner is pinned before and after.
An explicit T[] empty array and push replace the untyped new Array(count) and
indexed writes. This removes the any[] relation to T at the return site. The
function signature is unchanged and no cast or import is added.

Ruling: Adamic, task b5w3ycg, Nov 7 03:20Z: established-shape adaptations, Compiler accepts.

Base: origin/main f0271c6a7b2e92900eb7046d8246cfb89f52cfb6.
Upstream: TypeScript 6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8.

## Measurement before construction

All 398 unchanged lane units pass their normal identity, baseline and API
guards: 106,366 passing tests, one sanctioned failure, zero skipped. The
unchanged stage3/lane/run.sh control passes too. Independent reconciliation
checks every raw count, every unit report, the exact test identity multiset,
and byte-identical baseline diff against that control. The largest count
piece takes 156.217 seconds. Each unit uses the runner-supported 90-second
measurement bound; no semantic guard or test selection is changed.

| Count class | Calls |
| --- | ---: |
| Non-negative integer | 860 |
| Negative | 0 |
| Fractional | 0 |
| NaN | 0 |
| Infinite | 0 |

Observed counts are 0, 1, 2, 5, 50, 100, 150 and 250. Negative means negative
integers; fractional means finite non-integers. All 860 calls return, with
zero callback exceptions and zero counterexamples.

For each counted n, both bodies call f(i) once in order for 0 <= i < n and
return a dense length-n array of those exact values. The scratch observer
feeds each actual callback value into a push-built shadow and checks length,
own-index presence and Object.is at every position. It never calls f twice.
The shipped adapter installs no observer. Every measured source and built
file is frozen by hash across all pieces; JavaScript outside the measured
owner is unchanged. Excluded seeds detect reversal, object substitution and
holes, and verify the VM entry points.

## Node edge witnesses

The witness extracts the actual baseline owner and the exact candidate body.
It records results, keys, callback order and exceptions for both versions.

| Count | Original | Push body | Equal |
| --- | --- | --- | --- |
| 0 | [] | [] | yes |
| 3 | three callback values | same three callback values | yes |
| -1 | RangeError | [] | no |
| 1.5 | RangeError | two callback values | no |
| NaN | RangeError | [] | no |

Negative, fractional, NaN and infinite counts each occur zero times in the
complete lane count. The differing named edges are outside that counted
domain, as recorded in the ledger. The accepted proof covers tsc's counted
caller inputs. No upstream message or issue is filed.

## Owner and proof contract

The adapter requires exactly one top-level owner and exact full before/after
text and hashes. Missing, duplicate, moved, changed, half-adapted and invalid
syntax states refuse before writing. Replacement preserves all bytes outside
the owner, LF/CRLF style, and foreign edits. Replay performs zero writes and
preserves bytes and mtime. All 11 contract checks pass.

Both full applies complete, including product packaging. All 84 paired replay
outcomes match after tree-path normalization except arrayof: one edit before,
zero after. Earlier adapters' identical refusals on the fully adapted trees
are retained, rather than suppressed or relabeled as successful replays.

Both ordinary adamic types commands receive all 79 sorted compiler roots and
exit 0 with zero diagnostics. All 746 source files are compared; only the
pinned core.ts owner differs. Both complete core JavaScript emits, comments
kept and removed, differ only there. All nine emitted compiler JavaScript
files are checked: only _tsc.js and typescript.js differ, within arrayOf.
Both typescript.d.ts and tsserverlibrary.d.ts are byte-identical. Every hunk
is displayed in JAVASCRIPT-HUNKS.md.

The full-core census follows the named core-sites reference flags and scope.
Current main adds only a shared type-string cache to that census tool; its
diff is retained. Before and after each have 271 units and zero checker
diagnostics. Distinct sites fall from 59 to 58: exactly F2 disappears, and no
new Refused or NotYet site appears anywhere in core.ts. Restoring the exact
original owner restores the complete 59-site baseline set. This establishes
removal of the refusal-scan finding; the generic unit remains deferred.

## Evidence and reproduction

PROOF.md and PROOF.json index the checks; COUNT.json retains the independently
reconciled count and histogram. measurement.tar.gz contains all raw count
pieces, normal reports, whole control, baseline proof and witnesses.
checkpoint.tar.gz retains the completed after apply and owner contract.
proof.tar.gz retains paired replays, builds, every JavaScript hunk, declaration
hashes, types, full censuses, mutant, compiler provenance and census overlay.
Each archive has an internal SHA256SUMS, an adjacent archive SHA-256, and a
successful read-back verification. Go evidence uses .go.txt names.

The retained scripts use this run's /workspace layout. Setup sources
/workspace/adamic-tools/env.sh, runs bash cloud/setup.sh and npm ci --prefix
stage3/api, and exports ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source
at the pin above. Use the recorded base checkout for the before/count scripts;
construction is gated on RECONCILED.json. The ordinary compiler is unmodified
main; the census overlay is measurement-only and asserts no IR output.

The initial baseline adapter phase completed, but an early compiler build
collided with product packaging. A separate isolated full apply completed
successfully, and all 81,504 applied inputs match the built baseline exactly.
Two layout preflights are retained separately and excluded from the canonical
count. The after full apply records child exit 0 and a complete keyed product;
an outer-shell status discrepancy is retained separately. Accepted results
use the immediately recorded full-apply status, complete log and manifest.

The adapter and owner-pin hashes stay frozen through every proof. Later
proof packaging changes product keys without changing the transformation.
A post-apply checkpoint is pushed separately, so the final feature branch
contains one commit on the recorded main tip. All started measurement, build,
census and replay workers are stopped.
''')
(own/'PROOF.md').write_text('''# arrayOf proof

Base: f0271c6a7b2e92900eb7046d8246cfb89f52cfb6.
TypeScript: 050880ce59e30b356b686bd3144efe24f875ebc8.

| Check | Result |
| --- | --- |
| Prior owner search | 0 matches before writing |
| Count before construction | 398 units, 860 non-negative integer calls, all other classes 0, counterexamples 0 |
| Whole lane and independent reconciliation | Pass; 106366 passes, 1 sanctioned failure, 0 skips; exact identities and baseline diff |
| Full apply on main | Exit 0, product packaged; 81504 inputs match measured baseline |
| Full apply on main plus arrayof | Child exit 0, product packaged; outer-shell discrepancy retained separately |
| Whole-owner and no-write contract | 11 checks pass; LF/CRLF, drift refusals, foreign edit, bytes and mtime |
| Paired replay of every adapter | 84 outcomes; 83 identical, arrayof changes one edit to zero edits |
| Ordinary adamic types before and after | All 79 sorted roots; both exit 0; zero diagnostics |
| Source comparison | 746 files; only the exact pinned arrayOf owner differs |
| Complete core JavaScript, comments kept and removed | Only arrayOf differs |
| All 9 emitted compiler JavaScript files | Only _tsc.js and typescript.js differ, only at arrayOf; all hunks displayed |
| typescript.d.ts and tsserverlibrary.d.ts | Both byte-identical |
| Full core.ts census | 271 units each; 59 sites before, 58 after; exactly F2 gone; 0 new sites anywhere |
| Original-owner mutant | Complete 59-site baseline restored, including core.ts:1339:12 |
| Archive integrity | Internal and external SHA-256 manifests; read-back passes |
| Process cleanup | All measurement, build, census and replay workers stopped |

JAVASCRIPT-HUNKS.md shows every nonempty source and emitted JavaScript hunk.
PROOF.json includes hashes, complete results, the count table and checkpoint.
The three packed evidence archives retain complete raw records and commands.
The failed packaging/layout preflights and the outer-shell status discrepancy
are recorded explicitly and excluded from successful command claims.
''')
count=sum(p.is_file() for p in own.rglob('*'));assert count<120;report['adapterDirectoryFiles']=count;report['totalChangedFiles']=count+1;(own/'PROOF.json').write_text(json.dumps(report,indent=2)+'\n');print(json.dumps(dict(status='pass',adapterDirectoryFiles=count,totalChangedFiles=count+1,archives=archives)),flush=True)
