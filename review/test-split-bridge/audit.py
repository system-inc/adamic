#!/usr/bin/env python3
import csv,hashlib,json,re,subprocess
from pathlib import Path
root=Path(__file__).resolve().parents[2];review=root/'review/test-split-bridge'
before=json.loads((review/'before.json').read_text());assert len(before)==6
units=json.loads((review/'units.json').read_text());assert len(units)==36
inputs=json.loads((review/'inputs-before.json').read_text())
after_inputs={}
for name in subprocess.check_output(['git','ls-files','--cached','--others','--exclude-standard'],cwd=root,text=True).splitlines():
 p=root/name
 if p.is_file() and name.startswith(('bridge/tsgo/','internal/','oracle/','go.')):after_inputs[name]=hashlib.sha256(p.read_bytes()).hexdigest()
for name,digest in inputs.items():
 if not name.endswith('_test.go'):assert after_inputs[name]==digest,('production or fixture changed',name)
(review/'inputs-after.json').write_text(json.dumps(after_inputs,indent=2)+'\n')
def events(log):
 for line in Path(log).read_text().splitlines():
  try:yield json.loads(line)
  except json.JSONDecodeError:pass
def output(log):return ''.join(e.get('Output','') for e in events(log))
summary={'before':before,'modes':{}}
rows=[]
for mode,expected_pass,pieces,positions in [('after',23,16,162),('after-corpus',38,31,1600)]:
 measured=json.loads((review/(mode+'.json')).read_text());assert len(measured)==43
 assert len({(r['package'],r['test']) for r in measured})==43
 assert all(r['exit']==0 and r['invocation_seconds']<30 for r in measured)
 assert sum(r['action']=='pass' for r in measured)==expected_pass
 expected={('bridge/tsgo',r['name']) for r in units}|{('bridge/tsgo','TestTSGoRequiresLink'),('bridge/tsgo','TestBridgeUnitsCoverEveryPiece'),('bridge/tsgo','TestBridgeProductCacheIsVerified')}|{(r['package'],r['test']) for r in before if r['package'].endswith('/checker')}
 assert {(r['package'],r['test']) for r in measured}==expected
 baseline=next(r for r in before if r['test']=='TestBridge')['log'] if mode=='after' else '/tmp/test-split-bridge-before-corpus.jsonl'
 old=re.search(r'oracle: (\d+) positions across (\d+) files, (\d+) bytes identical',output(baseline));assert old
 assert int(old[1])==positions
 oracle_rows=[r for r in measured if r['test'].startswith('TestBridgeOracle') and r['action']=='pass']
 byte_count=sum(int(re.search(r'oracle: [^,]+, (\d+) bytes identical',output(r['log']))[1]) for r in oracle_rows)
 assert byte_count==int(old[3]),('oracle byte coverage',mode,byte_count,old[3])
 query_count=0
 for row in measured:
  for match in re.finditer(r'queries: [^,]+, (\d+) positions, (\d+) unchanged program roots',output(row['log'])):
   query_count+=int(match[1]);assert int(match[2])==int(old[2])
 assert query_count==positions*5
 coverage=next(r for r in measured if r['test']=='TestBridgeUnitsCoverEveryPiece')
 assert f'coverage: {pieces} active pieces, {positions} query positions' in output(coverage['log'])
 region=next(r for r in measured if r['test']=='TestBridgeRegion')
 previous=re.search(r'region probe: Go and native answer (.+)',output(baseline))[1]
 current=re.search(r'region probe: Go and native answer (.+)',output(region['log']))[1]
 assert current==previous,('region result/counts changed',previous,current)
 active={r['name'] for r in units if not r['file'] or (r['file']=='sample.ts')==(mode=='after')}
 selected=[r for r in measured if r['test'] in active];assert len(selected)==pieces
 maximum=max(selected,key=lambda r:r['invocation_seconds'])
 previous_seconds=next(e['Elapsed'] for e in events(baseline) if e.get('Test')=='TestBridge' and e['Action']=='pass')
 summary['modes'][mode]={'passed':expected_pass,'skipped':43-expected_pass,'pieces':pieces,'query_positions':positions,'five_analysis_query_positions':query_count,'oracle_bytes':byte_count,'region_result':current,'max_test_seconds':max(r['seconds'] for r in measured),'max_invocation_seconds':max(r['invocation_seconds'] for r in measured),'slowest_bridge_piece':maximum,'before_seconds':previous_seconds}
 rows.append(dict(test='TestBridge '+('sample' if mode=='after' else 'compiler corpus'),before=previous_seconds,after=max(r['seconds'] for r in selected),after_with_fetch=maximum['invocation_seconds'],units=pieces))
 for r in before:
  if r['test']=='TestBridge':continue
  current=next(n for n in measured if (n['package'],n['test'])==(r['package'],r['test']))
  if mode=='after':rows.append(dict(test=r['package']+'/'+r['test'],before=r['seconds'],after=current['seconds'],after_with_fetch=current['invocation_seconds'],units=1))
with (review/'timings.csv').open('w') as f:
 writer=csv.DictWriter(f,fieldnames=list(rows[0]),lineterminator='\n');writer.writeheader();writer.writerows(rows)
mutants=json.loads((review/'mutants.json').read_text())
planted=[r for r in mutants if r['name'].startswith('changed-native-query-')];assert len(planted)==4
assert [(r['shard'],r['failed']) for r in planted]==[(f'{i}/4',['TestBridgeOracleSample'] if i==3 else []) for i in range(4)]
assert all(r['exit']==1 for r in mutants if r['name'] not in [p['name'] for p in planted] and r['name']!='poison-build-callback')
assert next(r for r in mutants if r['name']=='poison-build-callback')['exit']==0
summary['mutants']=mutants
line=Path('/tmp/test-split-bridge-products-final.log').read_text();match=re.search(r'key=([a-f0-9]+) directory=(\S+)',line);assert match
products=Path(match[2]);hashes=json.loads((products/'products.json').read_text());assert len(hashes)==19
for name,digest in hashes.items():assert hashlib.sha256((products/name).read_bytes()).hexdigest()==digest
summary['products']={'key':match[1],'directory':str(products),'hashes':hashes}
metadata={'base':subprocess.check_output(['git','merge-base','HEAD','54cbc125'],cwd=root,text=True).strip(),'submodules':subprocess.check_output(['git','submodule','status','--recursive'],cwd=root,text=True).strip(),'nproc':subprocess.check_output(['nproc'],text=True).strip(),'cpu_max':Path('/sys/fs/cgroup/cpu.max').read_text().strip(),'corpus_commit':subprocess.check_output(['git','-C','/tmp/test-split-bridge-typescript','rev-parse','HEAD'],text=True).strip()}
assert metadata['base'].startswith('54cbc125');assert metadata['corpus_commit']=='050880ce59e30b356b686bd3144efe24f875ebc8'
assert subprocess.check_output(['git','-C','/tmp/test-split-bridge-typescript','status','--porcelain'],text=True)==''
corpus=Path('/tmp/test-split-bridge-typescript')
metadata['corpus_source_sha256']={str(p.relative_to(corpus)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((corpus/'src').rglob('*')) if p.is_file()}
for name,command in {'go':['go','version'],'clang':['clang','--version'],'node':['node','--version']}.items():metadata[name]=subprocess.check_output(command,text=True).strip()
(review/'environment.json').write_text(json.dumps(metadata,indent=2)+'\n')
(review/'audit.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps({k:v for k,v in summary.items() if k not in ('before','mutants','products')},indent=2))
print('Verified 19 build products under key',match[1])
