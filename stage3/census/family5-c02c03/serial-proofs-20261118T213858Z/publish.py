import os,sys,json,gzip,shutil,subprocess,datetime
from pathlib import Path
w=Path('/tmp/family5-final-proofs');r=Path('/tmp/family5-land');os.chdir(r)
s=json.loads((w/'state.json').read_text());C02=s['c02'];C03=s.get('c03')
lines=['# Family 5 C02/C03: accepted lane and serial proofs','',('Stopped: '+s['stop']) if s.get('stop') else ('Both landings proved and pushed.' if s.get('c03_pushed') else 'Continuation status; uncompleted gates are not claimed green.'),'',f"Main `{s['main']}`; C02 `{C02}`; C03 `{C03 or 'not created'}`. TypeScript v6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`.",'','C02 lane is green by the owner’s acceptance: both alternating candidate runs passed 106,366 cases with only the one sanctioned failure, real lane exit 0. The earlier reds were nondeterminism. The 19 JavaScript bundles were byte-identical. See [the full alternating report](../alternating-runs/REPORT.md), hunks, manifests and all run receipts. Its archive was not committed: SHA-256 `773ed33da75936a4d4c89b40cff991fd75cdbfc4c37b6663a53f7f04a2605b06` (2,084,049 bytes).','','C02 was amended from measured `28a2e22015d0cf51667fcefa7acb246cab553817` only to add the requested README lane receipt. `readme-only-amendment.patch.gz` proves the adapter, pins, TypeScript source and JavaScript are unchanged. Previously green checks are carried with their original real exit receipts; fresh remaining proofs ran serially, never beside the lane or another census.','','## Proof receipts','','| Proof | Real child exit | Wall seconds |','|---|---:|---:|']
for p in sorted(w.rglob('status.json')):
 if 'published' in p.parts:continue
 d=json.loads(p.read_text());lines.append(f"| {p.parent.relative_to(w)} | {d['exit_code']} | {d.get('wall_seconds','—')} |")
lines+=['','## Blocking finding','',(w/'BLOCKER.md').read_text(),'','Stock checks require 79 roots and zero diagnostics in both configurations; adamic types exits 0. The public API cmp is byte-for-byte. TestProduct uses -timeout 45m. Full entry censuses use LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, LATENT_TAIL=1, CommonJSBundle, LATENT_TIMES=/dev/stderr, exact absolute LATENT_ONLY_FILE and tsc.ts as root. Complete stdout/stderr and JSONL streams are losslessly retained. No probe or internal/ edit.','','## Census comparisons and every reached stop','']
for p in sorted(w.rglob('comparison*.json')):
 d=json.loads(p.read_text())
 if 'changes' not in d:continue
 lines += ['','`'+str(p.relative_to(w))+'`:','```json',json.dumps(d['files'],indent=2),'```']
 for ch in d['changes']:
  if ch['grade']=='gone':continue
  f=ch.get('after',{});lines+=['',f"- **{ch['grade']}**: `{f.get('where')}`; unit `{f.get('unit')}`; kind `{f.get('kind')}`; phase `{f.get('phase')}`. Full reason: {f.get('reason')}",'',str(f.get('text'))]
for p in sorted(w.rglob('mutant-result.json')):lines+=['','`'+str(p.relative_to(w))+'`:','```json',p.read_text(),'```']
lines+=['','## Publication','','```json',json.dumps(s.get('pushes',{}),indent=2),'```','','Each landing is one commit containing only its adaptation directory and one appended ledger line. C03 must be directly parented by C02. A missing receipt is not a pass. The old five-result gap was six skipped case bodies minus one failed hook counted as a failure; the alternating report accounts for it.']
(w/'REPORT.md').write_text('\n'.join(lines)+'\n')
subprocess.run(['git','switch','typescript/family5-c02c03-evidence'],check=True);subprocess.run(['git','pull','--ff-only','origin','typescript/family5-c02c03-evidence'],check=True)
root=r/'stage3/census/family5-c02c03';dest=root/('serial-proofs-'+datetime.datetime.now(datetime.timezone.utc).strftime('%Y%m%dT%H%M%SZ'));dest.mkdir()
for p in w.rglob('*'):
 if not p.is_file() or p.name=='workflow.lock':continue
 rel=p.relative_to(w);q=dest/rel;q.parent.mkdir(parents=True,exist_ok=True)
 if p.suffix in ['.log','.jsonl','.patch','.diff']:
  with p.open('rb') as src,gzip.GzipFile(filename=str(q)+'.gz',mode='wb',mtime=0) as dst:shutil.copyfileobj(src,dst)
 else:shutil.copyfile(p,q)
old=root/'REPORT.md';shutil.copyfile(old,root/'accepted-lane-pending-report.md')
old.write_text('# Family 5 C02/C03\n\n[Current serial proof report]('+dest.name+'/REPORT.md)\n\n'+(w/'REPORT.md').read_text().replace('../alternating-runs/','alternating-runs/'))
subprocess.run(['git','add',str(root)],check=True);subprocess.run(['git','commit','-m','stage3: record serial family 5 adaptation proofs'],check=True);subprocess.run(['git','push','origin','HEAD:refs/heads/typescript/family5-c02c03-evidence'],check=True)
sha=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip();(w/'publication.json').write_text(json.dumps({'sha':sha,'report':str(dest.relative_to(r))+'/REPORT.md'},indent=2)+'\n');print('EVIDENCE '+sha,flush=True)
