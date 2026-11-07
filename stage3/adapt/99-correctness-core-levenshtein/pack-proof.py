import hashlib,json,shutil,subprocess,tarfile,io,re
from pathlib import Path
s=Path('/workspace/scratch/levenshtein');repo=Path('/workspace/adamic');out=repo/'stage3/adapt/99-correctness-core-levenshtein'
def read(name):return json.loads((s/name).read_text())
count=read('COUNT.json');reconciled=read('reconciled.json');apply=read('after-apply.json');post=read('post-apply.json');replay=read('replay-summary.json');census=read('census-comparison.json');checkpoint=read('checkpoint30-six.json');comparison=read('comparison.json');witness=read('accepted-witness-final.json');frozen=read('adapter-frozen.json')
assert count['complete'] and reconciled['complete'] and count['differences']==0
assert apply['exit']==0 and all(v['exit']==0 for v in post.values())
assert replay['onlyOurOutcomeDiffers'] and replay['adapter30Identical'] and replay['finalSourcesByteIdentical']
assert checkpoint['identical'] and checkpoint['bothSuccessful']
assert not census['added'] and not census['ownerAfterSites'] and set(census['removedHistoricalIds'])=={'F'+str(i) for i in range(22,37)}
assert all(d['identical'] for d in comparison['declarations'])
assert hashlib.sha256((out/'adapt.cjs').read_bytes()).hexdigest()==frozen['adaptSha256']
assert hashlib.sha256((out/'owners.json').read_bytes()).hexdigest()==frozen['ownersSha256']
for name in ['finalize-measurement.py','resume-count.py','await-count.py','checkpoint.py','count.py','control.py','reconcile.py','instrument.cjs','preload.cjs','measure-prepare.py','witness.cjs','replay.py','compare.cjs','contract.cjs','prepare-census.sh','run-census.py','census.py','checked-sites.cjs','implement.py','adapter-template.cjs','after-apply.py','post-apply.py','checkpoint30-six.py','pack-proof.py']:
 shutil.copyfile(s/name,out/name)
for name in ['COUNT.json','reconciled.json','comparison.json','replay-summary.json','census-comparison.json','checked-read-sites.json']:
 shutil.copyfile(s/name,out/name)
shutil.copyfile(s/'accepted-witness-final.json',out/'witness.json')
readme=(out/'README.md').read_text().replace('Both row allocations become new Array<number>(s2.length + 1).fill(0), and six','Both row allocations become `new Array<number>(s2.length + 1).fill(0)`, and six').replace('indexed reads gain !.','indexed reads gain `!`.').replace('predecessor is therefore nonnegative. The final index n is always in the row.','predecessor is therefore nonnegative. `minJ` is 1 or the ceiling of the\npositive difference `i - max`, so it is at least 1. `maxJ` is the floor of\na value no larger than n, so it is at most n. The final index n is always\nin the row.').replace('identity, baseline, API and\ncount guards.','identity, baseline, API and\ncount guards.');(out/'README.md').write_text(readme)
(out/'types.json').write_text(json.dumps({'rootsEachSide':79,'beforeDiagnostics':0,'afterDiagnostics':0,'beforeExit':post['before-final-types']['exit'],'afterExit':post['after-final-types']['exit']},indent=2)+'\n')
# All emitted JavaScript hunks, with no elided changed files.
changedJS=[r['name'] for r in comparison['emitted'] if not r['identical']];assert changedJS
hunks='# Complete emitted JavaScript hunks\n\nAll nine emitted compiler JavaScript files were compared. Changed files: '+', '.join(changedJS)+'. Both complete core.ts transpiles (comments kept and\nremoved) are also scoped to the same two initializers. The six ! assertions\nerase completely. Raw full-context diffs and hashes are in evidence.tar.gz.\n\n'
for name in changedJS:
 a=s/'whole/adapted-tree/built/local'/name;b=s/'after/built/local'/name;r=subprocess.run(['diff','--unified=0',str(a),str(b)],capture_output=True,text=True);assert r.returncode==1
 hunks+='## '+name+'\n\n```diff\n'+r.stdout+'```\n\n'
for tag in ['false','true']:
 r=subprocess.run(['diff','--unified=0',str(s/('core-'+tag+'-before.js')),str(s/('core-'+tag+'-after.js'))],capture_output=True,text=True);assert r.returncode==1
 hunks+='## core.js removeComments='+tag+'\n\n```diff\n'+r.stdout+'```\n\n'
(out/'JAVASCRIPT-HUNKS.md').write_text(hunks.rstrip()+'\n')
provenance={'baseMain':'f0271c6a7b2e92900eb7046d8246cfb89f52cfb6','initialReportCommit':'81619026099036b6bc0677c9199e2fb00f1453e1','checkpointCommit':subprocess.check_output(['git','rev-parse','HEAD'],cwd=repo,text=True).strip(),'upstream':'050880ce59e30b356b686bd3144efe24f875ebc8','ordinaryAdamicSha256':hashlib.sha256((s/'adamic').read_bytes()).hexdigest(),'latentBinarySha256':hashlib.sha256((s/'latent').read_bytes()).hexdigest(),**frozen,'sourceFiles':comparison['sourceFiles'],'sourceDelta':comparison['changed'],'fullBaselineApply':read('prepare.json')['apply-before'],'fullAmendedApply':apply,'wholeControl':read('control.json'),'acceptedCount':reconciled,'adapter30Checkpoint':checkpoint['bothSuccessful']}
(out/'provenance.json').write_text(json.dumps(provenance,indent=2)+'\n')
(out/'PROOF.md').write_text('''# Completed L2 proof

The user authorized both the two filled number-array initializers and the six
checked indexed reads. The real-call count completed before implementation.
No Adamic compiler source, lane runner, upstream lib or existing adapter changes.
Both original 30 string-read assertions are unchanged.

| Proof item | Result |
| --- | --- |
| Main base | f0271c6a7b2e92900eb7046d8246cfb89f52cfb6 |
| TypeScript pin | 050880ce59e30b356b686bd3144efe24f875ebc8, 6.0.3 |
| Setup and pinned API install | Exit 0 |
| Full main apply, --build-product | Exit 0 |
| Full main plus this adapter apply, --build-product | Exit 0 |
| Real-call differential count | '''+str(count['calls'])+''' calls, 0 differences, 398 accepted shards |
| Unchanged full lane control | Pass; accepted shard identities, counts and baseline diff match exactly |
| Node witnesses | '''+str(len(witness['rows']))+''' of '''+str(len(witness['rows']))+''' return values equal; requested and observed classes covered |
| Owner contract | Pass; exact before/after pins, every partial checked-read refusal, LF/CRLF, foreign text and no-write replay |
| Full paired adapter replay | '''+str(replay['adapters'])+''' adapters; only this adapter's outcome differs |
| Adapter 30 full-tree replay | Identical preexisting addRange decline refusal on both sides; zero candidate-caused refusal |
| Adapter 30 checkpoint replay with all eight edits | Both exit 0, identical outcomes, zero source changes |
| Final replay sources | All '''+str(replay['sourceFiles'])+''' files byte-identical after both replays converge |
| Ordinary adamic types, all 79 roots together | Both exit 0, zero diagnostics |
| Ordinary compiler builds | Both exit 0 |
| Complete core JavaScript, comments kept and removed | Only the two initializers differ; six ! assertions erase |
| All nine emitted JavaScript files | Only '''+', '.join(changedJS)+''' differ, at those same initializers |
| typescript.d.ts | Byte-identical |
| tsserverlibrary.d.ts | Byte-identical |
| Source scope | All '''+str(comparison['sourceFiles'])+''' sources compared; only the complete pinned core.ts owner differs |
| Complete core.ts site census | '''+str(census['beforeSites'])+''' before, '''+str(census['afterSites'])+''' after; F22..F36 removed; zero new sites anywhere |
| Initializer-only mutant | Original allocation stops return; all six ! assertions and every other source byte retained |
| Archive manifests | Every member SHA-256 plus outer archive SHA-256 |
| Worker cleanup | No live task workers at closeout |

F IDs retain the historical baseline ordering from getOwnValues onward; they
are not renumbered after removal. census-comparison.json retains every removed
finding and the complete initializer-only mutant owner findings. Both census
runs use LATENT_FULL=1, LATENT_ONLY_FILE=core.ts, LATENT_ASSERT_NO_OUTPUT=1 and
LATENT_TIMES, exactly as the reference census. No accepted-program findings
are filtered except Boundary markers, and every core.ts site is compared.
The no-output assertions run before the first census record.

The full-tree adapter-30 refusal labels addRange's original from[i] decline
site core.ts:992:21, already changed by later main adapters. It is identical
before and after this edit. The independent post-30 checkpoint pair succeeds
with the exact eight-site owner adapter. No existing adapter is amended.

The original stopping report and its seven initializer-only checker errors
remain historical in initial-report.tar.gz. Two discarded count/control
attempts are in excluded-attempts.tar.gz: the duplicate-module-link setup
failure recorded no calls; the memory-contended partial count recorded 2,611
calls and zero differences, and its parallel full control had three timeouts.
The accepted control and the complete accepted shard union ran in isolation.
No excluded call, seed witness or diagnostic retry inflates COUNT.json.

The existing published branch history is preserved with plain pushes. The
post-apply checkpoint preserves both successful applies and the complete count;
its plain push succeeded after using the provided credential header. Later evidence
files change product keys without changing the frozen adapter or owner pins;
provenance.json checks their hashes against the implementation checkpoint.
Every retained Go source has a .go.txt suffix. All command output is logged.

Reproduction uses the exact scratch paths retained by the scripts, after
source /workspace/adamic-tools/env.sh, bash cloud/setup.sh and npm ci --prefix
stage3/api, with ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source at the
pinned commit. instrument.cjs, measure-prepare.py and count.py build and count
both bodies; reconcile.py checks the archive union against the unchanged full
control. after-apply.py, post-apply.py, replay.py, checkpoint30-six.py and
census.py record the remaining proof steps. The initial report archive is
historical; its runner cleanup transcript contains old PIDs and must not run.
''')
# Pack raw proof data without package installs, generated whole trees, or binaries.
evidence=s/'final-evidence';evidence.mkdir()
patterns=['*.log','*.json','*.jsonl','*.diff','core-*-before.js','core-*-after.js']
excludedPrefixes=('excluded-','count-','report-','initial-report')
for pattern in patterns:
 for p in s.glob(pattern):
  if not p.is_file() or p.name.startswith(excludedPrefixes):continue
  if p.name in ['process-audit.json','full-types.json','type-probe.json','assertion-probe.json','witness.json','native-before-owner.log','native-candidate-owner.log','before-types.log','candidate-types.log','confirm-driver.log','prepare-driver.log']:continue
  shutil.copyfile(p,evidence/p.name)
for tree in ['before','after','mutant']:shutil.copyfile(s/tree/'src/compiler/core.ts',evidence/(tree+'-core.ts'))
for root,label in [(s/'overlay','overlay'),(s/'whole','whole-control')]:
 for p in root.rglob('*'):
  if not p.is_file() or 'adapted-tree' in p.parts:continue
  target=evidence/label/p.relative_to(root);target.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(p,target)
for name in ['provenance.json','types.json','PROOF.md','README.md']:shutil.copyfile(out/name,evidence/name)
(evidence/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+str(p.relative_to(evidence))+'\n' for p in sorted(evidence.rglob('*')) if p.is_file()))
archive=out/'evidence.tar.gz'
with tarfile.open(archive,'w:gz') as tar:tar.add(evidence,arcname='proof')
(out/'evidence.sha256').write_text(hashlib.sha256(archive.read_bytes()).hexdigest()+'  '+archive.name+'\n')
# Preserve discarded observations separately from the counted union.
excluded=s/'excluded-evidence';excluded.mkdir()
for p in s.glob('excluded-*'):
 if p==excluded:continue
 if p.is_dir():shutil.copytree(p,excluded/p.name,ignore=shutil.ignore_patterns('adapted-tree'))
 else:shutil.copyfile(p,excluded/p.name)
(excluded/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+str(p.relative_to(excluded))+'\n' for p in sorted(excluded.rglob('*')) if p.is_file()))
archive=out/'excluded-attempts.tar.gz'
with tarfile.open(archive,'w:gz') as tar:tar.add(excluded,arcname='excluded')
(out/'excluded-attempts.sha256').write_text(hashlib.sha256(archive.read_bytes()).hexdigest()+'  '+archive.name+'\n')
print(json.dumps({'files':len(list(out.iterdir())),'calls':count['calls'],'differences':0,'proof':'complete'}),flush=True)
