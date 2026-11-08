import json,sys
from pathlib import Path
sys.dont_write_bytecode=True
out=Path(__file__).resolve().parent;scratch=Path('/workspace/scratch/c3b22184');tree=Path('/workspace/c3b22184-adapted')
sys.path.insert(0,'/workspace/adamic/stage3/meter')
import report
rows=sum([json.loads((out/f'group-{i}-rows.json').read_text()) for i in range(1,5)],[])
status=json.loads((scratch/'entry-latent-status.json').read_text())
if status['exit_code']==0:
 checks=report.root_order_checks(tree,scratch/'artifacts',rows)
 result=dict(status='complete',readiness_node_checks=checks,root_order_artifacts=[r for r in checks if r['node']['throws']])
else:
 probes=json.loads((out/'node-probes.json').read_text())
 candidates=[r for r in rows if r.get('reason') in report.READINESS_REASONS]
 times=scratch/'entry-unit-times.jsonl';timing=[json.loads(l) for l in times.read_text().splitlines()] if times.exists() else []
 stop=dict(status=status,last_completed_timing=timing[-1] if timing else None,active_unit='not recorded; the last completed unit is not the interrupted unit',raw_records=len((scratch/'artifacts/tsc/latent.jsonl').read_bytes().splitlines()))
 (out/'ENTRY-STOP.json').write_text(json.dumps(stop,indent=2)+'\n')
 result=dict(status='incomplete',stopping_reason='The bounded entry census did not complete. report.py requires its attempted-source and finding ledger before classifying root-order artifacts.',entry_status=status,root_order_artifacts=[],pending=[dict(root=r['roots'][0].split('/src/compiler/')[1],where=r['where'].split('/src/compiler/')[1],reason=r['reason'],node=probes.get(r['roots'][0]),classification='artifact classification incomplete') for r in candidates])
(out/'root-order.json').write_text(json.dumps(result,indent=2)+'\n')
print('root_order_status='+result['status'])
