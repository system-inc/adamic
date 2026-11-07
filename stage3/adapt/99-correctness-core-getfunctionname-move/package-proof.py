from pathlib import Path
import json,hashlib,tarfile,shutil,subprocess,difflib
s=Path('/workspace/scratch/core-finish');raw=s/'raw';repo=Path('/workspace/adamic');h1repo=s/'head1-repo';h1=h1repo/'stage3/adapt/99-correctness-core-cast-message';h2=repo/'stage3/adapt/99-correctness-core-getfunctionname-move';load=lambda n:json.loads((raw/n).read_text());sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
required=['head1-normal-comparison.json','head2-normal-comparison.json','head1-paired-replay.json','head2-paired-replay.json','head1-core-census-comparison.json','head2-debug-census-comparison.json','normal-cast-witness.json','cast-restoration-mutant.json','moved-name-witness.json','moved-real-six-unit.json','whole-lane-pair-reconciliation.json','head2-no-deprecatedCompat.json','earlier-apply-deltas.json','final-validation.json','contracts.json']
for name in required:
 x=load(name);assert x.get('passed',True),name
assert load('head2-text-reference-cross-check.json')['passedForAdaptedSymbol']
for side in ['baseline','head1','head2']:assert load(side+'-full-apply-result.json')['exit']==0
# Preserve all actual emitted JavaScript and both public declarations, with inner hashes.
def archive_js(side,dest):
 root=s/side/'built/local';files=sorted(root.rglob('*.js'))+[root/'typescript.d.ts',root/'tsserverlibrary.d.ts'];manifest={str(p.relative_to(root)):sha(p) for p in files};data=json.dumps(manifest,indent=2).encode();tmp=s/(side+'-js-manifest.json');tmp.write_bytes(data)
 with tarfile.open(dest,'w:gz') as tf:
  for p in files:tf.add(p,arcname=str(p.relative_to(root)))
  tf.add(tmp,arcname='manifest.json')
 assert dest.stat().st_size<20_000_000
for side,name in [('baseline','before-javascript.tar.gz'),('head1','after-javascript.tar.gz')]:archive_js(side,h1/name)
for side,name in [('head1','before-javascript.tar.gz'),('head2','after-javascript.tar.gz')]:archive_js(side,h2/name)
head1json=['head1-normal-comparison.json','head1-paired-replay.json','head1-replay-noop-write-proof.json','head1-core-census-comparison.json','normal-cast-witness.json','cast-restoration-mutant.json','whole-lane-pair-reconciliation.json']
head2json=['head2-normal-comparison.json','head2-paired-replay.json','head2-replay-difference.json','head2-debug-census-comparison.json','head2-stock-references.json','head1-stock-references.json','head2-text-reference-cross-check.json','head2-no-deprecatedCompat.json','cast-restoration-mutant.json','moved-name-witness.json','moved-real-six-unit.json','whole-lane-pair-reconciliation.json','earlier-apply-deltas.json','final-validation.json','contracts.json','pair-positive-control.json','pair-negative-control.json','pair-instrumentation.json','pair-after-instrumentation.json']
for out,names in [(h1,head1json),(h2,head2json)]:
 for name in names:shutil.copy2(raw/name,out/name)
for side,out in [('head1',h1),('head2',h2)]:shutil.copy2(raw/(side+'-all-emitted-javascript-hunks.patch'),out/'all-emitted-javascript-hunks.patch')
for side,out in [('baseline',h1),('head1',h1),('head2',h2)]:shutil.copy2(raw/(side+'-source-manifest.json'),out/(side+'-source-manifest.json'))
for side,out in [('baseline',h1),('head1',h1),('head2',h2)]:shutil.copy2(raw/(side+'-replay.tar.gz'),out/(side+'-replay.tar.gz'))
for p in sorted((s/'evidence').glob('lane-*.tar.gz')):shutil.copy2(p,h2/p.name)
shutil.copy2(s/'evidence/runtime-counts.json',h2/'runtime-counts.json');shutil.copy2(raw/'moved-real-six-unit.tar.gz',h2/'moved-real-six-unit.tar.gz')
for name in ['run.py','phase.py','compare-normal.cjs','compare-replays.py','normal-cast-witness.cjs']:shutil.copy2(s/name,h1/name)
for name in ['run.py','phase.py','compare-normal.cjs','compare-replays.py','compare-censuses.py','references.cjs','trace-types.py','mutants.py','instrument-pairs.cjs','instrument-pairs-after.cjs','pair-preload.cjs','pair-control.cjs','pair-lane.py','run-pair-lane.py','after-unit.py','moved-name-witness.cjs','reconcile-pairs.py','final-validation.cjs','contracts.cjs','package-proof.py']:shutil.copy2(s/name,h2/name)
# Raw preparation/failure history remains visible; no normal generated control library duplication.
excluded={'moved-control-library.cjs'}
with tarfile.open(h2/'raw-proof.tar.gz','w:gz') as tf:
 for p in sorted(raw.rglob('*')):
  if p.is_file() and p.name not in excluded and not p.name.endswith('.tar.gz'):tf.add(p,arcname=str(p.relative_to(raw)))
with tarfile.open(h1/'raw-proof.tar.gz','w:gz') as tf:
 for p in sorted(raw.iterdir()):
  if p.is_file() and not p.name.endswith('.tar.gz') and p.name.startswith(('baseline-','head1-','adamic-build','normal-cast','cast-restoration','final-validation','whole-lane')):tf.add(p,arcname=p.name)
 for name in ['setup.log','setup-result.json','api-install.log','api-install-result.json','build-census.log','build-census-result.json']:
  p=Path('/workspace/scratch/core-cast/raw')/name
  if p.exists():tf.add(p,arcname='setup/'+name)
for side,out in [('head1',h1),('head2',h2)]:
 names=['src/compiler/core.ts'] if side=='head1' else ['src/compiler/debug.ts','src/deprecatedCompat/deprecate.ts'];before='baseline' if side=='head1' else 'head1';parts=[]
 for name in names:parts+=difflib.unified_diff((s/before/name).read_bytes().decode().splitlines(True),(s/side/name).read_bytes().decode().splitlines(True),fromfile='before/'+name,tofile='after/'+name)
 (out/'source-hunks.patch').write_text(''.join(parts))
# Head1 cites the immutable result bytes from head2, avoiding a circular commit hash.
mutantDigest=sha(raw/'cast-restoration-mutant.json');lane=load('whole-lane-pair-reconciliation.json');final=load('final-validation.json');census=load('head2-debug-census-comparison.json');checkpoints={side:load(side+'-checkpoint.json') for side in ['baseline','head1','head2']}
head1proof={'status':'PASS','base':'2a95448b8fadbcd30549a148bd28da5f8b75f907','publishedDraftParent':'d718a08df495c4d28da3228279ecf549d8dd010d','rule':'JavaScript-changing, approved failure text removal','fullApplies':{side:load(side+'-full-apply-result.json') for side in ['baseline','head1']},'checkpoints':{side:checkpoints[side] for side in ['baseline','head1']},'normalComparison':load('head1-normal-comparison.json'),'pairedReplay':load('head1-paired-replay.json'),'normalNodeWitness':True,'types79BothExit0':True,'coreCensus':load('head1-core-census-comparison.json'),'freshWholeLaneCastFailureCalls':0,'firingControl':1,'getFunctionNameSourceUntouched':True,'revisedMutant':{'head':'typescript/core-getfunctionname-move','file':'stage3/adapt/99-correctness-core-getfunctionname-move/cast-restoration-mutant.json','sha256':mutantDigest,'stockTscExit':2,'adamicTypesExit':1,'actualLocation':'core.ts:1789:97','olderForecastLocation':'core.ts:1791:97','diagnostic':'TS2339: getFunctionName does not exist on typeof Debug'}}
(h1/'PROOF.json').write_text(json.dumps(head1proof,indent=2)+'\n')
(h1/'README.md').write_text(f'''# cast keeps its value and drops only the predicate name

Complete, JavaScript-changing. Adamic task b5w3ycg, Nov 7 22:06:52Z,
Compiler accepts under the delegation; the user subsequently revised the
mutant proof to run on head 2. Base: 2a95448b8fadbcd30549a148bd28da5f8b75f907.
The published stopped checkpoint is retained as history, with a normal
continuation commit. No force-push is used.

The chosen wording is `Invalid cast. The supplied value ${{value}} did not pass the test.`
Only the space, quoted predicate name and Debug.getFunctionName call disappear.
The period stays. `${{value}}` belongs to #09dvwny and is unchanged. The
condition, return, Debug.fail call and its throw implementation stay byte for
byte. Current source: core.ts:1789; the earlier forecast called it 1791 before
the elementAt declaration cut removed two preceding lines.

## Ownership and adapter

No earlier active whole-owner or exact-message pin covers cast. The audit of
224 adapter/pin inputs includes 99q, 99r and the earlier core adaptations.
40 and 41's old whole-cast text is historical evidence, not an active guard.
The new directory pins the complete before/after owner, refuses drift and
performs zero writes on an already-adapted tree. No earlier adapter changes
in this head. Debug and getFunctionName are byte-identical to the baseline.

## Counts and witnesses

| Measurement | Real executions | Firing control |
| --- | ---: | ---: |
| Original getFunctionName expression in cast's failure message, cited fcb70da5e | 0 | 1 |
| Cast failure branch on this exact cut, fresh whole lane | 0 | 1 |

This does not claim cast itself is never called. Its failure branch is zero.
The fresh lane covers all 398 units, 106366 passes, one expected failure and
zero skips, with all identities and baseline guards preserved. The full raw
pieces and independent reconciliation live in head 2's move directory;
whole-lane-pair-reconciliation.json here retains the checked result.

The actual completed normal builds are used by normal-cast-witness.cjs.
Six failing value classes preserve Error class, condition evaluation count,
value interpolation and exact throw point. Four passing cases retain identity;
a predicate-thrown sentinel retains its identity. Only the shorter text differs.

## Revised mutant, from head 2

cast is an uninstantiated generic in the core census. Its before/after core
censuses have identical site sets, no added site, and zero diagnostics. No
lowering success is inferred from deferral. As the user ruled, the mutant is
instead run on the finished moved tree: restore the original complete cast
owner while Debug no longer exports getFunctionName.

Stock tsc exits 2 and adamic types over all 79 roots exits 1. Each reports only
TS2339 at core.ts:1789:97: Property 'getFunctionName' does not exist on type
of Debug. This is the earlier 1791 site after the elementAt cut. The source
restoration hash also matches the complete original core.ts bytes.

Citation: typescript/core-getfunctionname-move,
stage3/adapt/99-correctness-core-getfunctionname-move/cast-restoration-mutant.json.
The exact result is copied here and pinned by SHA256 `{mutantDigest}`.
This replaces the superseded absent-core-site mutant requirement.

## Proof results

- Both full applies: exit 0, {load('baseline-full-apply-result.json')['seconds']:.3f}s baseline and {load('head1-full-apply-result.json')['seconds']:.3f}s head 1.
- Post-apply checkpoints: pushed after each apply, recorded in PROOF.json.
- Paired replay: all 97 adapters; only cast's transformation differs.
- Adapter 49 rewrites core.ts unchanged: same exit, output, paths and four
  writes. Its raw no-op hashes inherit exactly our message difference. Both
  writes are identity operations; the raw hashes and proof remain visible.
- Both 79-root adamic types checks: exit 0, zero diagnostics.
- Both own compiler/harness builds: exit 0, zero diagnostics.
- Every JavaScript hunk: only one message deletion in each of _tsc.js,
  typescript.js and run.js; all other emitted JavaScript byte-identical.
- typescript.d.ts and tsserverlibrary.d.ts: byte-identical.
- Core full-mode census: identical sites, no additions, cast deferred.
- Actual normal-build Node witness: pass, as described above.
- Revised missing-member mutant on head 2: rejected by stock tsc and Adamic.
- getFunctionName and all other source files: unchanged.
- Ledger: records the ruling and the head-2 mutant proof.

See PROOF.json, all-emitted-javascript-hunks.patch, source-hunks.patch and the
raw archives. The original STOP.json and raw-stopped-proof.tar.gz are marked
historical. The old -9 applies were manually stopped, not killed by the
2-hour limit or an observed memory failure. Fresh applies all pass. Peak
child RSS is recorded; it is not represented as an aggregate process-tree peak.
''')
(h1/'PROOF.md').write_text('''# Proof index

Each result is recorded in PROOF.json and README.md. Normal JavaScript archives
include every .js file and both public declarations with a per-file manifest.
All hunks are shown in all-emitted-javascript-hunks.patch. Raw command logs,
actual child exits, source manifests and full census streams are archived.
The paired replay no-op write nuance is independently proved in
head1-replay-noop-write-proof.json, without rewriting any raw journal.
The revised mutant is cast-restoration-mutant.json, cited from head 2 and
pinned by its SHA256 in README.md.
''')
stop=json.loads((h1/'STOP.json').read_text());stop.update(historical=True,supersededBy='PROOF.json',resolution='user revised mutant to missing Debug member on head 2; all fresh proofs now complete');(h1/'STOP.json').write_text(json.dumps(stop,indent=2)+'\n')
ledger=h1repo/'stage3/upstream/LEDGER.md';text=ledger.read_text();start=text.index('- 99-correctness-core-cast-message:');text=text[:start]+'''- 99-correctness-core-cast-message: JavaScript-changing. Adamic, task b5w3ycg, Nov 7 03:20Z: established-shape adaptations, Compiler accepts; specific approval Nov 7 22:06:52Z and the subsequent head-2 mutant ruling. cast keeps its condition, Debug.fail call and ${value} (#09dvwny); only the quoted predicate name is removed. The cited and fresh whole-lane failure-message counts are 0, each with a firing control; no real-call counterexample. Restoring the original cast on the moved head-2 tree is rejected by stock tsc and adamic types with missing Debug.getFunctionName, recorded in the adaptation README.
''';ledger.write_text(text)
head2proof={'status':'PASS','base':'2a95448b8fadbcd30549a148bd28da5f8b75f907','head1':'to be set to the finished head before publishing','fullApplies':{side:load(side+'-full-apply-result.json') for side in ['head1','head2']},'checkpoints':checkpoints,'bodyRawSha256':final['sourceBodyRawSha256'],'sourceBodyByteIdentical':True,'earlierApplyDeltas':load('earlier-apply-deltas.json'),'pairedReplay':load('head2-paired-replay.json'),'types79ActualNoDeprecatedCompat':load('head2-no-deprecatedCompat.json'),'normalComparison':load('head2-normal-comparison.json'),'debugCensus':census,'realCalls':6,'returnedNames':{'noop':6},'counterexamples':0,'wholeLaneUnits':398,'actualMovedSixCallsMatched':True,'normalTsDebugLosesOnlyHelper':True,'noReexportThroughRootOrDebug':True,'stockReferencesOnlyLocalDefinitionAndCall':True,'exhaustiveGrepScopeNote':'15 separate implementation matches remain in unchanged lib bootstrap tools; zero old member names in src and all newly emitted artifacts. These bootstrap objects are independently verified, not readers of the adapted helper.','mutants':{'castMissingMember':True,'reflectionStopRestored':True}}
(h2/'PROOF.json').write_text(json.dumps(head2proof,indent=2)+'\n')
(h2/'PROOF.md').write_text(f'''# getFunctionName move proof

Complete, JavaScript-changing. Base 2a95448b. All fresh full applies pass.
Historical stop records remain visible and are superseded by this proof.

| Proof item | Result |
| --- | --- |
| Head 1 and head 2 full applies | exit 0; {load('head1-full-apply-result.json')['seconds']:.3f}s and {load('head2-full-apply-result.json')['seconds']:.3f}s |
| Checkpoint after each full apply | pushed; SHAs in PROOF.json |
| Source function body | byte-identical; raw SHA256 {final['sourceBodyRawSha256']} |
| Earlier apply path | 30, 41, 45 and road (b) Git patches byte-identical to baseline |
| Full paired replay | 98 adapters; all existing exits and writes identical |
| Recognition output | only 41's source positions change from line 16 to 18; same rule-31 refusal |
| Move replay | exit 0, zero writes |
| Drift and half-state contracts | reject at both homes; missing and duplicate homes reject |
| Road (b) original contract mutants | all eight still reject without writes |
| Actual adamic types | all 79 roots, exit 0, zero diagnostics |
| Actual compiler import boundary | all 79 roots seen in file trace; zero deprecatedCompat accesses |
| Stock compiler graph | zero deprecatedCompat imports |
| Own compiler and harness build | exit 0, zero diagnostics; adds 0 diagnostics |
| All normal emitted JavaScript | only moved helper, old namespace assignment and caller wiring differ |
| Public declarations | both byte-identical |
| Stock find-references | only module-local definition and deprecate call |
| Old Debug name in src and all own emitted files/maps | zero |
| Normal ts object | Debug loses only getFunctionName; no root re-export |
| Full debug census | 105 to 104 units; 355:26 gone, no new sites |
| Reflection restoring mutant | original Debug bytes restored; every original site returns |
| Cast restoring mutant | stock tsc exit 2 and adamic types exit 1, TS2339 at 1789:97 |
| Fresh whole lane | 398 units exactly once; 106366 pass, one expected failure, zero skip |
| Actual helper calls | 6, all own-data-name functions; name noop in every case |
| Original/proposed body counterexamples | 0 on all six actual arguments |
| Actual moved-tree call witness | same six input/name records and ordinary unit guards pass |
| Cast failure-branch executions | 0, firing control 1 |
| Comparison negative control | deliberately wrong name caught; excluded from real counts |
| Edge name witnesses | named, arrow, bound, empty name, deleted own name and nonfunction agree |
| Ledger | exact ruled destination and reflection deferral recorded |

All 20 lane archives, 797 process/thread records, outer and inner hashes and
ordinary reports were independently reconciled. Sequential piece times sum to
{lane['sumPieceSeconds']:.3f}s; largest piece {lane['largestPieceSeconds']:.3f}s.
The after move replays stage3-lane-file-293, where all six calls occur, against
the actual moved normal build. Each actual call also evaluates the old pinned
body on that exact function object and lexical hasProperty, with no difference.
Function text, name descriptors and returned-name multisets match before/after.

## Census interpretation

The unchanged guarded meter uses LATENT_FULL=1, LATENT_ONLY_FILE=debug.ts and
LATENT_ASSERT_NO_OUTPUT=1. No meter logic changes. Physical scratch Go files
are .go.txt. Every run has zero checker diagnostics. Raw historical labels are
retained even when their diagnostic list is empty.

The removed declaration contains {census['removedNewlineCount']} newlines; subsequent source positions
are mapped by that exact amount for site comparison. No reason, kind or phase
is normalized. Six distinct non-Boundary findings disappear, all recorded by
the getFunctionName unit: five in its Debug body and its hasProperty dependency
at core.ts:1275:12. No site is added. Restoring the entire original Debug file
restores the entire original site set, including both 355:26 observations.
The exact refusal-scan reason is `a method read as a value (toString would lose
its object, and this with it)`.

cast itself remains an uninstantiated generic, so its revised mutant restores
the original complete owner on head 2, not a nonexistent core census stop.
Both real checkers reject the missing Debug member. The actual line is 1789;
1791 was the older forecast before elementAt's declaration cut.

## Bootstrap grep scope, reported rather than hidden

An exhaustive grep with ignore files disabled finds 15 old implementation
matches in unchanged lib/_tsc.js and lib/typescript.js. These are TypeScript's
bootstrap tools, not the new emitted product. Their bytes match before the
move, their Debug object is distinct from the normal product's Debug, and they
cannot reference this module-local source symbol. All matches and hashes are
in head2-text-reference-cross-check.json. src and every newly emitted artifact,
including maps and declarations, have zero old Debug member names. Installed
node_modules copies are separate build tooling as well. No literal zero is
claimed over those independent compiler copies.

## Evidence and verification details

Every normal JavaScript hunk is in all-emitted-javascript-hunks.patch. The
normal artifacts and public declarations have inner SHA manifests in the
before/after archives. The JavaScript check parses AST tokens, preserving
literal contents; a naive scanner initially misread a template tail and its
failure is retained. Corrected comparison passes with only the approved
move/wiring and code trivia differences. Normal artifacts were rehashed after
all tests, and all 746 source inputs per tree still match each apply checkpoint.

The full replay keeps raw outcomes and write bytes. Only repository/source
roots are canonicalized. Adapter 41's exact source positions 16 to 18 are
checked separately, including its error header and stack, with no reason or
exit normalization. Head 1's adapter-49 no-op write hashes reflect only its
inherited cast-message input difference and are independently proved there.

The old -9 applies were manually stopped after the old proof blockers. They
were not observed memory failures or environment timeout kills. Fresh steps
all finish. Peak child RSS and actual child exits are in their result records.
No force-push or main push is used; already-published checkpoint history is
preserved when the two finished heads are joined.
''')
(h2/'README.md').write_text('''# getFunctionName belongs to its compatibility user

Complete. JavaScript-changing, Adamic task b5w3ycg, Nov 7 22:06:52Z and the
subsequent explicit replay-pin ruling, Compiler accepts under the delegation.
This head includes the finished cast-message head on road (b)'s cut 2a95448b.
See PROOF.md for one line per required proof and PROOF.json for exact results.

## Move and exact owners

The adapter removes the exported helper from namespace Debug, changes the
whole pinned deprecate owner to call its module-local helper, adds the pinned
type AnyFunction and hasProperty imports, and appends the function. The entire
function body moves byte for byte, including indentation, name lookup,
Function.prototype.toString.call and match[1]!. No algorithm or fallback changes.
The only signature change is removal of export. No helper is re-exported.

owners.json pins the old declaration, whole caller before/after, whole import
before/after, local function and body hashes. All four pieces must be wholly
original or wholly moved. Missing, duplicate, drifted or half-applied owners
refuse before writing. Planned-tree validation, body-byte equality and source
rereads precede writes. Finished replay writes nothing.

Current references are deprecate.ts:73's local call and line 77's definition.
Stock find-references over all 711 src roots returns only those two references.
The actual 79-root adamic file trace reads every root and no deprecatedCompat
path. The function leaves the current compiler graph rather than moving its wall.

## Owning amendments

30, 41 and 45 each retain their strict home-recognition code and pins in their
own directory. Their ordinary apply patches are byte-identical to main's:
30 still asserts the original capture, 41 still erases its original .name
cast, and 45 still performs its regexp proof at the original home. Finished
recognition accepts only the fully adapted local body. Missing or duplicate
homes, changed bodies, and partially restored moved bodies refuse.

30 recognizes only the exact moved match[1] site as already asserted. 41
recognizes only rule 6 at the moved home, preserving all other rules. 45 runs
its existing capture proof on a strictly checked virtual declaration; any
attempted edit to that virtual debug source refuses, and no virtual code is
written. Capture 1 must still be mandatory.

Road (b)'s condition-4 whole-helper guard follows getFunctionName to
 deprecate.ts, pinning the same body byte for byte and still rejecting drift
on replay. The original whole Debug owner is accepted unchanged before the
move. Road (b)'s README records the same amendment. Its four actual assertion
message edits and all other helper guards remain unchanged. Its original
eight contract mutants still reject without writes.

## Counts and names

| Input class | Real calls | Equal original/proposed names | Counterexamples |
| --- | ---: | ---: | ---: |
| Function with an own data name | 6 | 6, all noop | 0 |
| All other real input classes | 0 | not observed | 0 |

All 398 ordinary lane units pass their identity, API, count and baseline guards.
Raw before/proposed comparisons use each exact function object, not a parsed
approximation. The actual moved tree replays all six calls in unit 293; input
and name multisets match. Node witnesses also cover named, bound, arrow,
empty/deleted-name and nonfunction cases. The normal built ts.Debug loses only
getFunctionName and the ts root gains no re-export. A wrong-name comparison
mutant is caught. Seeded controls are excluded from real counts.

## Scope and reflection ruling

By Adamic's word: deprecatedCompat ships in typescript.js, not tsc, so when
typescript.js comes home after 1.0.0, getFunctionName's reflection
(Function.prototype.toString) meets us there. No compiler/library reflection
admission changes in this head.

No old Debug member name remains in src or any newly emitted file, map or
declaration. Exhaustive checkout grep also reports separate, unchanged lib
bootstrap compiler copies with their own Debug objects; those 15 matches are
listed in the proof, not silently omitted or claimed to be readers of this
private function. The product's normal object and stock symbol checks pass.

## Evidence

PROOF.md indexes every item. Both normal JavaScript archives include all .js
files and both public declarations with hashes. All hunks, source manifests,
full replay streams, census streams, raw child exits and all 20 lane pieces
are retained. SHA256SUMS and ARCHIVES.json identify the evidence. Historical
stop files remain marked as history. The revised cast mutant is recorded here
and copied/cited by head 1 without a circular commit-hash dependency.
'''.replace('\n deprecate.ts','\ndeprecate.ts'))
for name in ['stop-adapter-result.json','move-probe.json']:
 x=json.loads((h2/name).read_text());x.update(historical=True,supersededBy='PROOF.json');(h2/name).write_text(json.dumps(x,indent=2)+'\n')
# Each older owner keeps its amendment receipt beside its own guard.
for directory in ['30-indexed-reads','41-explicit-any-remaining','45-regex-captures','99-correctness-core-debug-messages']:
 target=repo/'stage3/adapt'/directory;receipt={'status':'PASS','amendment':'finished-tree recognition accepts exact adapted getFunctionName at original or module-local compatibility home','sourceBodyRawSha256':final['sourceBodyRawSha256'],'ordinaryApplyPatchByteIdentical':True,'fullReplayExistingExitAndWritesUnchanged':True,'driftAtBothHomesRefused':True,'halfAppliedMovedBodyRefused':True,'proof':'../99-correctness-core-getfunctionname-move/PROOF.md'};(target/'getfunctionname-move-proof.json').write_text(json.dumps(receipt,indent=2)+'\n')
# Check old and new archives, retain their exact bytes, and produce a complete archive index.
for out in [h1,h2]:
 archives=[{'file':p.name,'bytes':p.stat().st_size,'sha256':sha(p)} for p in sorted(out.glob('*.tar.gz'))];assert all(a['bytes']<20_000_000 for a in archives);(out/'ARCHIVES.json').write_text(json.dumps(archives,indent=2)+'\n');print(out.name,'files',len(list(out.iterdir())),'archives',len(archives),'largest',max(a['bytes'] for a in archives))
print('all proof packets prepared; final commit references and checksums still to seal')
