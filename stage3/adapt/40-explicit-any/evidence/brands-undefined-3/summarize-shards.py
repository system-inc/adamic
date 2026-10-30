import sys,json,collections,hashlib,gzip
from pathlib import Path
sys.dont_write_bytecode=True
lane,results,output=map(Path,sys.argv[1:]);sys.path.insert(0,str(lane))
from shard_plan import names_digest
plan=json.loads((lane/'shard-plan.json').read_text()); reports=[];records=[]
for unit in plan['units']:
 report=json.loads((results/unit['name']/'report.json').read_text())
 assert report['status']=='pass',(unit['name'],report['errors'])
 assert report['inputs_hash']==unit['inputs_hash']
 assert report['counts']==unit['expect']
 assert report['seconds']<30
 observed=json.loads((results/unit['name']/'observations.json').read_text())
 assert names_digest(observed)==unit['names_digest']
 reports.append(report);records.extend(observed)
counts={k:sum(r['counts'][k] for r in reports) for k in ['pass','fail','skip']}
assert counts==plan['counts']
expected=json.loads((lane/'expected.json').read_text())
failures=[x for r in reports for x in r['failed_tests']]
assert failures==expected['failed_tests']
paths=[x for r in reports for x in r['baseline_diffs']]
assert paths==expected['baseline_diffs']
result={'status':'pass','units':len(reports),'tests':len(records),'counts':counts,'failed_tests':failures,'baseline_diffs':paths,'names_digest':names_digest(records),'maximum_seconds':max(r['seconds'] for r in reports),'manifest_sha256':hashlib.sha256((lane/'shards.json').read_bytes()).hexdigest(),'product_path':True,'concurrency':json.loads((results/'ledger.json').read_text())['concurrency']}
output.mkdir(parents=True,exist_ok=True)
(output/'shards-summary.json').write_text(json.dumps(result,indent=2)+'\n')
for name,value in [('shards-reports.json.gz',reports),('shards-observations.json.gz',records)]:
 with (output/name).open('wb') as file:
  with gzip.GzipFile(filename='',mode='wb',fileobj=file,mtime=0) as zipped:zipped.write((json.dumps(value)+'\n').encode())
print(json.dumps(result))
