"""Continue bounded pieces, recording the exact accepted lane unit union."""
import collections, hashlib, json, os, pathlib, subprocess, time, sys
repo=pathlib.Path('/workspace/adamic');s=pathlib.Path('/workspace/scratch/multimap');out=repo/'stage3/adapt/99-correctness-core-multimap';tree=s/'counted';units=json.loads((repo/'stage3/lane/shards.json').read_text())['units']
def manifest():
 return {str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for d in ['src','built','tests/cases','tests/baselines/reference'] for p in sorted((tree/d).rglob('*')) if p.is_file()}
frozen=manifest();manifest_file=s/'counted-manifest.json'
if len(sys.argv)>1:assert json.loads(manifest_file.read_text())==frozen,'count input changed before resume'
else:manifest_file.write_text(json.dumps(frozen,indent=2)+'\n')
def aggregate():
 accepted={};processes={};attempts=[]
 for directory in sorted(s.glob('count-*')):
  if not directory.is_dir() or not (directory/'summary.json').exists():continue
  summary=json.loads((directory/'summary.json').read_text());attempts.append({k:summary[k] for k in ['start','end','seconds','complete']})
  for row in summary['units']:
   if row['exit']==0 and row['report'].get('status')=='pass':
    assert row['unit'] not in accepted,'duplicate accepted unit'
    accepted[row['unit']]=row
    processes[row['unit']]=list((directory/'calls').glob(row['unit']+'.*.json'))
 counts=collections.Counter();failures=[];oracle=collections.Counter()
 for name,row in accepted.items():
  oracle.update(row['report']['counts'])
  for p in processes[name]:
   r=json.loads(p.read_text());counts.update(r['counts']);failures+=r['failures']
 result={'complete':len(accepted)==len(units),'units':len(accepted),'manifestUnits':len(units),'acceptedUnits':[u['name'] for u in units if u['name'] in accepted],'counts':dict(counts),'counterexamples':len(failures),'enumerations':counts['enumerations'],'oracleCounts':dict(oracle),'attempts':attempts,'sourceAndBuiltManifestUnchanged':manifest()==frozen,'excluded':'All failed-unit records, preparation calls, and synthetic witnesses. The initial byte-comparison unit is replaced only by its passing retry.'}
 (out/'COUNT.json').write_text(json.dumps(result,indent=2)+'\n');assert not failures and counts['enumerations']==0 and result['sourceAndBuiltManifestUnchanged'];return result
print(json.dumps(aggregate()),flush=True)
for start in range(int(sys.argv[1]) if len(sys.argv)>1 else 0,len(units),40):
 directory=s/f'count-{start:03d}'
 if directory.exists():raise SystemExit('existing piece: '+str(directory))
 begin=time.monotonic();print('begin piece',start,flush=True)
 with (s/f'piece-{start:03d}.log').open('w') as log:
  r=subprocess.run(['python3',str(out/'measure.py'),str(start),str(min(40,len(units)-start))],cwd=repo,stdout=log,stderr=subprocess.STDOUT,timeout=1380)
 summary=aggregate();print(json.dumps({'piece':start,'exit':r.returncode,'seconds':time.monotonic()-begin,'accepted':summary['units'],'calls':summary['counts'].get('calls',0),'counterexamples':summary['counterexamples'],'enumerations':summary['enumerations']}),flush=True)
 assert r.returncode==0
result=aggregate();assert result['complete'];print('complete',result['units'],result['counts']['calls'],flush=True)
