"""Require complete lane-shard coverage and audit every recorded split comparison."""
from pathlib import Path
import gzip,hashlib,json,sys
sys.dont_write_bytecode=True
root=Path(__file__).resolve().parents[3]
sys.path.insert(0,str(root/'stage3/lane'))
from shard_plan import names_digest,runtime_hash,unit_inputs_hash,load_plan
results,calls,out,original,counted=map(Path,sys.argv[1:])
plan_path=root/'stage3/lane/shards.json'
plan=load_plan();units=plan['units']
ledger=json.loads((results/'ledger.json').read_text())
assert {r['name'] for r in ledger['units']}=={r['name'] for r in units}
assert len(ledger['units'])==len(units)
assert all(r['exit']==0 and r['status']=='pass' and not r['errors'] for r in ledger['units'])
assert ledger['maximum']<30
counts=dict(pass_=0,fail=0,skip=0);reports=[];failures=[];baseline_paths=[]
with gzip.open(out/'lane-shard-observations.jsonl.gz','wt') as observations_archive:
 for unit in units:
  folder=results/unit['name'];report=json.loads((folder/'report.json').read_text())
  observations=json.loads((folder/'observations.json').read_text())
  assert report['name']==unit['name'] and report['status']=='pass' and report['counts']==unit['expect']
  assert report['artifact']=='stage3-adapted-tree'
  assert report['inputs_hash']==unit_inputs_hash(unit,plan,'stage3-adapted-tree',runtime_hash())
  assert names_digest(observations)==unit['names_digest']
  observed={'pass':sum(r['status']=='pass' for r in observations),'fail':sum(r['status']=='fail' for r in observations),'skip':sum(r['status']=='pending' for r in observations)}
  assert observed==unit['expect']
  counts['pass_']+=observed['pass'];counts['fail']+=observed['fail'];counts['skip']+=observed['skip']
  reports.append(report);failures.extend(report['failed_tests']);baseline_paths.extend(report['baseline_diffs'])
  observations_archive.write(json.dumps({'unit':unit['name'],'observations':observations},separators=(',',':'))+'\n')
expected=json.loads((root/'stage3/lane/expected.json').read_text())
assert failures==expected['failed_tests'] and baseline_paths==expected['baseline_diffs']
assert counts=={'pass_':106366,'fail':1,'skip':0}
with gzip.open(out/'lane-shard-reports.jsonl.gz','wt') as target:
 for report in reports:target.write(json.dumps(report,separators=(',',':'))+'\n')
with gzip.open(out/'lane-ledger.json.gz','wb') as target:target.write((results/'ledger.json').read_bytes())
processes=[];total=0;differences=0
with gzip.open(out/'split-calls.jsonl.gz','wt') as archive:
 for file in sorted(calls.glob('*.jsonl')):
  count=0
  for line in file.read_text().splitlines():
   row=json.loads(line);assert set(row)=={'range','before','after','same'}
   assert isinstance(row['range'],str)
   assert isinstance(row['before'],list) and isinstance(row['after'],list)
   assert all(isinstance(x,str) for x in row['before']+row['after'])
   same=row['before']==row['after']
   assert row['same']==same
   differences+=not same;total+=1;count+=1
   archive.write(json.dumps({'pid':int(file.stem),**row},separators=(',',':'))+'\n')
  processes.append({'pid':int(file.stem),'calls':count,'sha256':hashlib.sha256(file.read_bytes()).hexdigest()})
assert total>0 and differences==0
instrument=json.loads((out/'instrumentation.json').read_text());instrumented=[]
for row in instrument:
 old=(original/'built/local'/row['name']).read_text();new=old
 for edit in row['edits']:
  assert old[edit['start']:edit['end']]==edit['before']
  new=new[:edit['start']]+edit['after']+new[edit['end']:]
 actual=(counted/'built/local'/row['name']).read_text()
 assert actual==new,'temporary instrumentation changed unexpectedly'
 instrumented.append({'file':row['name'],'sites':len(row['edits']),'original_sha256':hashlib.sha256(old.encode()).hexdigest(),'instrumented_sha256':hashlib.sha256(actual.encode()).hexdigest()})
assert {row['file'] for row in instrumented}=={'_tsc.js','typescript.js','run.js'}
summary={'status':'PASS','calls':total,'differences':differences,'units':len(units),'counts':{'passing':counts['pass_'],'failing':counts['fail'],'pending':counts['skip']},'failed_tests':failures,'baseline_diffs':baseline_paths,'maximum_shard_seconds':ledger['maximum'],'wall_seconds':ledger['wall_seconds'],'concurrency':ledger['concurrency'],'coordinator_exit':0,'plan_sha256':hashlib.sha256(plan_path.read_bytes()).hexdigest(),'runtime_hash':runtime_hash(),'instrumented_files':instrumented,'processes':processes,'control_directory_excluded':True,'archive_sha256':hashlib.sha256((out/'split-calls.jsonl.gz').read_bytes()).hexdigest()}
(out/'count-audit.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps({k:v for k,v in summary.items() if k!='processes'},indent=2))
