from pathlib import Path
import json,sys
r=json.loads(Path(sys.argv[1]).read_text());expected=json.loads(Path(sys.argv[2]).read_text())
assert {k:v['exit'] for k,v in r['phases'].items()}==expected['phase_exits'],r['phases']
assert r['counts']==expected['platforms']['linux/x64/'+r['node']],r['counts']
assert r['baseline_diffs']==expected['baseline_diffs'],r['baseline_diffs']
assert r['workers']==8 and r['runners']=='all' and r['tests'] is None
assert len(r['failures'])==1,r['failures']
sys.path.insert(0,str(Path(sys.argv[2]).parent))
from check import failed_tests
titles=failed_tests((Path(sys.argv[1]).parent/'tests.log').read_text())
assert titles==expected['failed_tests'],titles
print(json.dumps({'status':'PASS','counts':r['counts'],'onlyBaselineDiff':r['baseline_diffs'],'failures':r['failures'],'workers':r['workers'],'runners':r['runners']},indent=2))
