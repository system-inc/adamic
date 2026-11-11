from pathlib import Path
import json,subprocess,sys,collections,datetime
S=Path(__file__).resolve().parent;R=Path('/workspace/adamic')
run=json.loads((S/'counted-lane.json').read_text());assert run['exit']==0,run
rows=[]
for line in (S/'counted-lane.log').read_text().splitlines():
 try:rows.append(json.loads(line))
 except ValueError:pass
passed=[r['Test'] for r in rows if r.get('Action')=='pass' and 'Test' in r]
assert len(passed)==len(set(passed))==398,len(passed)
assert not [r for r in rows if r.get('Action') in ['fail','skip']]
assert any(r.get('Action')=='pass' and 'Test' not in r for r in rows)
manifest=json.loads((R/'stage3/lane/shards.json').read_text());assert len(manifest['units'])==398
expected={k:sum(u['expect'][k] for u in manifest['units']) for k in ['pass','fail','skip']};assert expected=={'pass':106366,'fail':1,'skip':0}
for name in ['lane-counts','project-counts','stock-late-counts']:
 p=subprocess.run([sys.executable,S/'summarize.py',name]);print('summary child exit:',p.returncode,flush=True);assert p.returncode==0
lane=json.loads((S/'lane-counts-summary.json').read_text());project=json.loads((S/'project-counts-summary.json').read_text())
assert lane['calls']==lane['visits']==0,lane
assert project['calls']==3 and project['visits']==84642
for result in [lane,project]:
 assert result['intrinsicFlaggedWithoutString']==result['descriptorFlaggedWithoutString']==result['objectFlaggedWithoutNumber']==0
 assert result['maskDifferences']==[0]*5 and result['intrinsicDifferences']==result['intrinsicAtDescriptorDifferences']==result['inAtDescriptorDifferences']==0
processes=[json.loads(p.read_text()) for p in (S/'lane-counts').glob('*.json')]
coverage={name:sum(name in '\n'.join(p['argv']) for p in processes) for name in passed};assert all(coverage.values()),[k for k,v in coverage.items() if not v]
(S/'lane-counter-process-coverage.json').write_text(json.dumps(coverage,indent=2)+'\n')
for side in ['base','flags-road']:
 types=json.loads((S/(side+'-types.json')).read_text());assert types['exit']==0 and len(types['command'])-2==79
 rs=[json.loads(l) for l in (S/(side+'-census.jsonl')).read_text().splitlines()];assert not rs[0]['checker_rejected'] and not rs[0]['diagnostics'];assert rs[-1]['status']=='relation_report'
 files=[r for r in rs if 'file' in r];assert len(files)==1 and len(files[0]['units'])==23
 if side=='flags-road':assert not [f for f in files[0]['findings'] if f['reason']=="a cast the runtime can't check"]
assert all(r['equal'] for r in json.loads((S/'stock-byte-witness-results.json').read_text()))
assert all(r['equal'] for r in json.loads((S/'late-instrumentation-identity-results.json').read_text()))
assert all(r['lowering_exit']==0 and not r['findings'] and not r['checker_rejected'] for r in json.loads((S/'stage0-probe-results.json').read_text()).values())
integrity=json.loads((S/'counted-product-integrity-results.json').read_text());assert integrity['corpusFiles']==82870 and integrity['corpusPathDifferences']==integrity['corpusContentDifferencesOutsideCounterHooks']==0
result=dict(completedAt=datetime.datetime.now(datetime.timezone.utc).isoformat(),laneShards=398,oracleCounts=expected,lane=lane,adaptedProjects=project,completedNodeProcesses=len(processes),coveredShards=len(coverage),objectFlagsRoad='getObjectFlags(type)',intrinsicNameRoad='type.flags & TypeFlags.Intrinsic ? (type as IntrinsicType).intrinsicName : undefined',writtenValueDifferences=0)
(S/'step1-results.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps({k:v for k,v in result.items() if k not in ['lane','adaptedProjects']}),flush=True)
