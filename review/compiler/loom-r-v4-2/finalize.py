import csv,json,re,subprocess
from pathlib import Path
e=Path(__file__).resolve().parent
raw=[json.loads(l) for l in (e/'focused-results.jsonl').read_text().splitlines()]
selected=json.loads((e/'selected-focused.json').read_text())
selected['internal/oracle']=[x for x in selected['internal/oracle'] if x!='TestOriginalArrayWitnessNode']+['TestOriginalArrayWitnessNode_'+str(i).zfill(4) for i in range(16)]
rows=[r for r in raw if r.get('test') and r['test']!='TestOriginalArrayWitnessNode']
for r in [json.loads(l) for l in (e/'split-results.jsonl').read_text().splitlines()]:r.update(grain=r['seconds']<60,skip=False);rows.append(r)
assert {(p,t) for p,ts in selected.items() for t in ts}=={(r['package'],r['test']) for r in rows},'unfinished focused leaves'
assert all(r['exit']==0 and r['grain'] for r in rows)
(e/'initial-selected-focused.json').write_text((e/'selected-focused.json').read_text())
(e/'selected-focused.json').write_text(json.dumps(selected,indent=2)+'\n')
(e/'final-focused-results.json').write_text(json.dumps(rows,indent=2)+'\n')
readers=[json.loads(l) for l in (e/'reader-results.jsonl').read_text().splitlines()];assert len([r for r in readers if r.get('build')])==53 and len([r for r in readers if r.get('test')])==48 and all(r['exit']==0 and r.get('grain',True) for r in readers)
flow=[json.loads(l) for l in (e/'flow-corpus.jsonl').read_text().splitlines()];assert flow[-1]['Action']=='pass' and not any(r['Action']=='fail' for r in flow)
leaves=[r for r in flow if r['Action']=='pass' and r.get('Test')];assert all(r['Elapsed']<60 for r in leaves)
mutants=[json.loads(l) for l in (e/'source-mutants.jsonl').read_text().splitlines()];assert len(mutants)==19 and all(r['killed'] for r in mutants)
assert json.loads((e/'witness-output-mutant.json').read_text())['killed']
summary=dict(focused=len(rows),focused_skips=[r['test'] for r in rows if r['skip']],focused_max=max(r['seconds'] for r in rows),flow_leaves=len(leaves),flow_max=max(r['Elapsed'] for r in leaves),flow_seconds=flow[-1]['Elapsed'],reader_packages=53,reader_tests=48,reader_max=max(r['seconds'] for r in readers if r.get('test')),external_mutants=20)
(e/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
with (e/'test-seconds.csv').open('w') as out:
 w=csv.writer(out);w.writerow(['suite','package','test','seconds','status'])
 for r in rows:w.writerow(['focused',r['package'],r['test'],r['seconds'],'skip' if r['skip'] else 'pass'])
 for r in readers:
  if r.get('test'):w.writerow(['reader',r['package'],r['test'],r['seconds'],'skip' if r.get('skip') else 'pass'])
 for r in leaves:w.writerow(['flow',r['Package'],r['Test'],r['Elapsed'],'pass'])
 for name in ['final-guards','post-split-guards','main-compatibility']:
  for line in (e/(name+'.jsonl')).read_text().splitlines():
   r=json.loads(line)
   if r.get('Action')=='pass' and r.get('Test'):w.writerow([name,r['Package'],r['Test'],r['Elapsed'],'pass'])
p=e/'REPORT.md';s=p.read_text();s+='\nFinal validated inventory: '+str(summary)+'\n\nSeven focused controls remain explicitly skipped for the prior documented pending runtime/readonly/factory coverage; these are listed in summary.json and their logs. No full package tests, full gate, ahra invocation, PR or Loom submission was run. All logs are retained, including setup mistakes and the single superseded over-budget aggregate; the final inventory has no over-budget leaf. All implementation Go files are gofmt-clean and source whitespace checks pass.\n';p.write_text(s)
for p in e.glob('*.test'):p.unlink()
print(json.dumps(summary,indent=2))
