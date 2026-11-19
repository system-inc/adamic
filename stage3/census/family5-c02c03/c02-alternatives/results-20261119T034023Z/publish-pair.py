import os,sys,json,subprocess,shutil,gzip,collections
from pathlib import Path
W=Path('/tmp/family5-c02-alternatives');R=Path('/tmp/family5-land');state=json.loads((W/'state.json').read_text());d=json.loads((W/'comparisons/whole-entry/comparison.json').read_text());named=json.loads((W/'entry/c02/callable-sites.json').read_text());phrase=named['phrase']
def fs(side):return [f for l in (W/'entry'/side/'latent.jsonl').read_text().splitlines() for f in json.loads(l).get('findings',[])]
base_named=[f for f in fs('base') if phrase in f.get('reason','') or phrase in f.get('text','')]
lines=['# C02 whole-entry first-error reach','','Base `'+state['main']+'`; C02 `'+state['c02']+'`. No landing push. Each side used its own rebuilt entry tool, the pinned v6.0.3 adapted tree, and ran alone. LATENT_FULL=0, LATENT_ASSERT_NO_OUTPUT=1, LATENT_TAIL=1, CommonJSBundle, tsc.ts entry, no LATENT_ONLY_FILE. Each raw stream was pushed before the next side started.','','## Counts','','C02 phrase: `'+phrase+'`. C02 has '+str(named['events'])+' matching events, '+str(named['all_distinct_positions'])+' distinct positions including Boundary mirrors, and '+str(named['primary_distinct_positions'])+' distinct primary diagnostic positions. Baseline has '+str(len(base_named))+' matching events at '+str(len({f['where'] for f in base_named if f['kind']!='Boundary'}))+' distinct primary positions.','','```json',json.dumps(d['files'],indent=2),'```','','| Run | Real exit | Wall seconds | Stream commit |','|---|---:|---:|---|']
for side in ['base','c02']:
 s=json.loads((W/'entry'/side/'status.json').read_text());lines.append('| '+side+' | '+str(s['exit_code'])+' | '+str(s['wall_seconds'])+' | `'+state['entry_pushes'][side]+'` |')
lines+=['','Pairs use kind, phase, reason and full message with positions removed and adapted roots normalized. Equal keys pair by unit/site first. A changed primary first-error stop can move inside the same unit. Reached requires an old G02 stop that is cleared in that unit; an unrelated Boundary or a remaining G02 is insufficient. Full raw records and before/after messages are retained.','','## Every C02 site naming the callable field','','| Site | Unit | Kind | Phase |','|---|---|---|---|']
for f in named['sites']:lines.append('| `'+f['where']+'` | `'+f.get('unit','')+'` | '+f['kind']+' | '+f['phase']+' |')
lines+=['','## Every difference, with full words','']
for i,x in enumerate(d['changes'],1):
 lines+=['','### '+str(i)+'. '+x['grade'],'']
 for side in ['before','after']:
  f=x.get(side)
  if not f:continue
  lines += [side+': `'+f['where']+'`; unit `'+f.get('unit','')+'`; '+f['kind']+' / '+f['phase']+'.','', 'Reason: '+f['reason'],'','```text',f['text'],'```']
 if x.get('reach_proof'):lines+=['','Reach proof:','```json',json.dumps(x['reach_proof'],indent=2),'```']
(W/'WHOLE-ENTRY.md').write_text('\n'.join(lines)+'\n')
root=R/'stage3/census/family5-c02c03/c02-alternatives';dest=root/'whole-entry';dest.mkdir(parents=True,exist_ok=False)
shutil.copyfile(W/'WHOLE-ENTRY.md',dest/'REPORT.md');shutil.copyfile(W/'entry/c02/callable-sites.json',dest/'callable-sites.json');shutil.copyfile(W/'comparisons/whole-entry/comparison.json',dest/'comparison.json')
for n in ['compare.py','publish-pair.py'] :shutil.copyfile(W/n,dest/n)
for sub in ['setup/build-base','setup/build-c02','comparisons/whole-entry']:
 for p in (W/sub).iterdir():
  if not p.is_file():continue
  q=dest/sub/p.name;q.parent.mkdir(parents=True,exist_ok=True)
  if p.suffix=='.log':
   with p.open('rb') as a,gzip.GzipFile(filename=str(q)+'.gz',mode='wb',mtime=0) as b:shutil.copyfileobj(a,b)
  else:shutil.copyfile(p,q)
subprocess.run(['git','add',str(root)],cwd=R,check=True);subprocess.run(['git','commit','-m','stage3: pair the complete C02 whole-entry reach'],cwd=R,check=True);subprocess.run(['git','push','origin','HEAD:refs/heads/typescript/family5-c02c03-evidence'],cwd=R,check=True)
sha=subprocess.check_output(['git','rev-parse','HEAD'],cwd=R,text=True).strip();(W/'pair-publication.json').write_text(json.dumps({'sha':sha,'report':str(dest.relative_to(R))+'/REPORT.md'},indent=2)+'\n');print('PUBLISHED whole-entry pair '+sha,flush=True)
