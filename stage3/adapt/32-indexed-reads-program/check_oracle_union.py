#!/usr/bin/env python3
"""Verify the standard product-path shards equal the standard whole lane, including statuses."""
from collections import Counter
import json
from pathlib import Path
import sys
repo,results,whole,output=map(lambda s:Path(s).resolve(),sys.argv[1:])
sys.path.insert(0,str(repo/'stage3/lane'))
from shard_plan import load_plan,names_digest,runtime_hash,unit_inputs_hash,title
from check import check_results
plan=load_plan();runtime=runtime_hash();assert runtime==plan['runtime']
records=[];counts=Counter();failures=[];paths=[];diff=b'';reports=[]
for unit in plan['units']:
    root=results/'units'/unit['name'];report=json.loads((root/'report.json').read_text());reports.append(report)
    assert report['status']=='pass' and report['counts']==unit['expect'] and report['seconds']<30,unit['name']
    assert report['artifact']=='stage3-adapted-tree'
    assert report['inputs_hash']==unit_inputs_hash(unit,plan,'stage3-adapted-tree',runtime)
    observed=json.loads((root/'observations.json').read_text());assert names_digest(observed)==unit['names_digest']
    records.extend(observed);counts.update(report['counts']);failures+=report['failed_tests'];paths+=report['baseline_diffs'];diff+=(root/'oracle/baseline.diff').read_bytes()
profile=[json.loads(s) for s in (whole/'whole-tasks.jsonl').read_text().splitlines()]
whole_records=[dict(record,task=row['task'],status=status) for row in profile for status,key in [('pass','passes'),('fail','errors')] for record in row[key]]
identity=lambda rows:Counter((r['task']['runner'],r['task']['file'],title(r['name']),r['status']) for r in rows)
assert identity(records)==identity(whole_records)
assert dict(counts)==plan['counts']
expected=json.loads((repo/'stage3/lane/expected.json').read_text());assert failures==expected['failed_tests'] and paths==expected['baseline_diffs']
assert diff==(whole/'oracle/baseline.diff').read_bytes()
assert check_results(whole)['status']=='pass'
result=dict(status='pass',units=len(reports),counts=dict(counts),observations=len(records),identical_test_identities_and_statuses=True,identical_sanctioned_API_bytes=True,failed_tests=failures,baseline_diffs=paths,maximum_shard_seconds=max(r['seconds'] for r in reports),runtime=runtime,product_source=str(results/'product'))
output.write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
