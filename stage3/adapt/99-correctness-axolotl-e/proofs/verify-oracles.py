#!/usr/bin/env python3
"""Check unchanged complete upstream baselines and the existing API sanction."""
import hashlib, json, pathlib, sys
sys.dont_write_bytecode=True
repo=pathlib.Path(__file__).resolve().parents[4]
sys.path.insert(0,str(repo/'stage3/lane'))
from check import api_check, failed_tests
scratch=pathlib.Path(sys.argv[1]);out=pathlib.Path(sys.argv[2]);out.mkdir(parents=True,exist_ok=True)
expected=json.loads((repo/'stage3/lane/expected.json').read_text())
sanctioned=json.loads((repo/'stage3/lane/sanctioned-api.json').read_text())
rows=[]
for label in ['before','after']:
    tree=scratch/('99e-'+label);oracle=scratch/('99e-oracle-'+label)
    report=json.loads((oracle/'report.json').read_text())
    assert report['counts']==dict(passing=106366,failing=1,pending=0),report['counts']
    assert report['runners']=='all' and report['tests'] is None and report['workers']==4
    assert {k:v['exit'] for k,v in report['phases'].items()}==dict(install=0,build=0,tests=1)
    assert report['baseline_diffs']==expected['baseline_diffs']
    assert failed_tests((oracle/'tests.log').read_text())==expected['failed_tests']
    assert report['cause_counts']=={'timeout':0,'baseline-content':1,'baseline-missing':0,'exception':0,'other':0}
    lane=scratch/('99e-api-'+label);lane.mkdir(exist_ok=True)
    for name,target in [('adapted-tree',tree),('oracle',oracle)]:
        link=lane/name
        if not link.exists(): link.symlink_to(target,target_is_directory=True)
    errors,api=api_check(lane,sanctioned);assert not errors,errors
    rows.append(dict(label=label,counts=report['counts'],phases=report['phases'],wallSeconds=report['wall_seconds'],sanctionedApi=api))
a=(scratch/'99e-oracle-before/baseline.diff').read_bytes();b=(scratch/'99e-oracle-after/baseline.diff').read_bytes()
assert a==b,'baseline moved'
c=(scratch/'99e-before/built/local/typescript.d.ts').read_bytes();d=(scratch/'99e-after/built/local/typescript.d.ts').read_bytes()
assert c==d,'public API moved'
result=dict(runs=rows,baseline=dict(moved=False,byteIdentical=True,bytes=len(a),sha256=hashlib.sha256(a).hexdigest()),api=dict(byteIdentical=True,bytes=len(c),sha256=hashlib.sha256(c).hexdigest()))
(out/'oracle-comparison.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result,indent=2))
