import pathlib,json
out=pathlib.Path(__file__).resolve().parent
summary={}
for name in ['census-before','census-after','mutant62','mutant63']:
 records=[json.loads(l) for l in (out/(name+'.jsonl')).read_text().splitlines()]
 assert not records[0]['diagnostics']
 findings=[f for r in records for f in r.get('findings',[]) if f.get('where','').split('/')[-1].startswith(('tracing.ts:340:','tracing.ts:359:')) and f['kind']!='Boundary']
 if name=='census-after':
  assert not any(f['kind']=='Refused' for f in findings)
  assert {f['where'].split('tracing.ts:')[1] for f in findings}=={'340:50','359:53'}
  assert all(f['kind']=='NotYet' and f['reason']=='JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata)' for f in findings)
 summary[name]=findings
records=[json.loads(l) for l in (out/'core-first-stop.jsonl').read_text().splitlines()]
assert not records[0]['diagnostics']
tracing=[r for r in records if r.get('file','').endswith('/tracing.ts')]
assert len(tracing)==1
summary['core-tracing']=tracing[0]
(out/'census-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps(summary['core-tracing'],indent=2))
