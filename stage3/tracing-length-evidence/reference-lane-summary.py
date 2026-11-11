from pathlib import Path
import json,re
r=Path('/workspace/scratch/baobab-evidence');repo=Path('/workspace/adamic')
assert json.loads((r/'whole-lane-exit.json').read_text())['exit']==0
rows=[json.loads(s) for s in (r/'whole-lane-go.jsonl').read_text().splitlines()]
expected=set(re.findall(r'func (TestShard_\w+)\(', (repo/'stage3/lane/shards_generated_test.go').read_text()))
passed=[x['Test'] for x in rows if x.get('Action')=='pass' and x.get('Test','').startswith('TestShard_')]
assert len(passed)==len(set(passed))==len(expected)==398;assert set(passed)==expected
assert not any(x.get('Action')=='fail' for x in rows)
manifest=json.loads((repo/'stage3/lane/shards.json').read_text());units=manifest['units'];assert len(units)==398
records=[json.loads(p.read_text()) for p in sorted((r/'lane-counts').glob('*.json'))]
assert records
coverage={u['name']:sum(any('/'+u['name']+'/' in arg for arg in p['argv']) for p in records) for u in units}
assert all(coverage.values()),[k for k,v in coverage.items() if not v]
calls=[]
for p in (r/'lane-counts').glob('*.jsonl'):calls.extend(json.loads(s) for s in p.read_text().splitlines())
assert len(calls)==sum(p['total'] for p in records)
differences=sum(p['differences'] for p in records);assert differences==sum(bool(x['difference']) for x in calls)
assert all(x['boolean']==x['double'] and x['beforeFails']==x['afterFails'] for x in calls);assert differences==0
controls={}
for directory in ['baseline-control-counts','control-counts']:
 rs=[json.loads(p.read_text()) for p in (r/directory).glob('*.json')];cs=[]
 for p in (r/directory).glob('*.jsonl'):cs.extend(json.loads(s) for s in p.read_text().splitlines())
 assert sum(x['total'] for x in rs)==len(cs)==2;assert sum(x['differences'] for x in rs)==0
 assert sorted(x['boolean'] for x in cs)==[False,True]
 controls[directory]={'calls':2,'differences':0,'onePassingAndOneFailing':True}
counts={k:sum(u['expect'][k] for u in units) for k in ['pass','fail','skip']}
summary={'wholeLaneExit':0,'units':len(passed),'oracleCountsCheckedByEachGoUnit':counts,'processRecords':len(records),'unitCounterProcessCoverage':coverage,'assertCalls':len(calls),'differences':differences,'oracleNeverCallsStopTracing':len(calls)==0,'controls':controls,'preparedArtifactKey':json.loads((r/'prepared.json').read_text())['key'],'instrumentation':'Temporary emitted bundle copies with VM-aware preload; measurement-only Go overlay selects the copied product and carries the preload through toolchain.Env. Dependency directory removed only from temporary product copy.'}
(r/'lane-summary.json').write_text(json.dumps(summary,indent=2)+'\n');print(json.dumps({k:v for k,v in summary.items() if k!='unitCounterProcessCoverage'},indent=2))
