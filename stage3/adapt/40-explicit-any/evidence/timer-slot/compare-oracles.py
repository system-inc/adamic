import collections,gzip,hashlib,json,sys
from pathlib import Path
repo,control,retyped,out=map(Path,sys.argv[1:]);out.mkdir(parents=True,exist_ok=True);plan=json.loads((repo/'stage3/lane/shard-plan.json').read_text());manifest=(repo/'stage3/lane/shards.json').read_bytes()
def zipped(name,data):
 with (out/name).open('wb') as stream:
  with gzip.GzipFile(fileobj=stream,filename='',mode='wb',mtime=0) as z:z.write((json.dumps(data)+'\n').encode())
def inspect(label,root):
 assert (root/'shards.json').read_bytes()==manifest
 reports=[];records=[];ledger=json.loads((root/'ledger.json').read_text());assert ledger['all_commands_executed'] and len(ledger['units'])==len(plan['units'])
 assert all(r['exit']==0 and r['status']=='pass' and r['seconds']<30 for r in ledger['units'])
 for unit in plan['units']:
  r=json.loads((root/unit['name']/'report.json').read_text());assert r['status']=='pass' and not r['errors'],r
  assert r['counts']==unit['expect'] and r['inputs_hash']==unit['inputs_hash'] and r['seconds']<30
  observed=json.loads((root/unit['name']/'observations.json').read_text());reports.append(r);records.extend(observed)
 counts={k:sum(r['counts'][k] for r in reports) for k in ['pass','fail','skip']};assert counts==plan['counts']
 failures=[x for r in reports for x in r['failed_tests']];expected=json.loads((repo/'stage3/lane/expected.json').read_text());assert failures==expected['failed_tests']
 identities=collections.Counter((r['task']['runner'],r['task']['file'],tuple(r['name']),r['status']) for r in records)
 paths=[x for r in reports for x in r['baseline_diffs']];assert paths==expected['baseline_diffs']
 summary=dict(units=len(reports),tests=len(records),counts=counts,failed_tests=failures,baseline_diffs=paths,maximum_seconds=ledger['maximum'],wall_seconds=ledger['wall_seconds'],concurrency=ledger['concurrency'])
 zipped(label+'-reports.json.gz',reports);zipped(label+'-identities.json.gz',sorted((*key,value) for key,value in identities.items()));zipped(label+'-ledger.json.gz',ledger)
 return summary,identities
cs,ci=inspect('control',control);rs,ri=inspect('retyped',retyped);assert ci==ri,'test pass/fail identities or multiplicities changed'
base='stage3-lane-byte-comparisons';a=(control/base/'oracle/baseline.diff').read_bytes();b=(retyped/base/'oracle/baseline.diff').read_bytes();assert a==b,'API baseline diff changed'
for filename in ['reference.d.ts','local.d.ts','built.d.ts']:assert (control/base/'saved-api'/filename).read_bytes()==(retyped/base/'saved-api'/filename).read_bytes()
result=dict(status='pass',manifest_sha256=hashlib.sha256(manifest).hexdigest(),control=cs,retyped=rs,identities_equal=True,multiplicity_checked=True,api_baseline_diff_identical=True,api_baseline_diff_sha256=hashlib.sha256(a).hexdigest(),public_api_diff=[],timer_state_public=False)
(out/'oracle-comparison.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
