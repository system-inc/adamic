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
record=json.loads((out/'core-production-first-stop.json').read_text())
assert record['checker_diagnostics']==0 and not record['output']
assert record['kind']=='Refused' and record['where'].endswith('/tracing.ts:380:1'),record
assert record['reason']=='an ambient value declaration',record
assert pathlib.Path('/workspace/tracing-measured-after/src/compiler/tracing.ts').read_text().splitlines()[379].startswith('declare const require:')
summary['core-first-stop']=record
(out/'census-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps(record,indent=2))
