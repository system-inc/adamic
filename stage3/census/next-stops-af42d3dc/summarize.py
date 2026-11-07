import json,collections,sys
sys.dont_write_bytecode=True
from pathlib import Path
repo=Path('/workspace/adamic');out=Path(__file__).resolve().parent;tree=Path('/workspace/after-wave-adapted');scratch=Path('/workspace/scratch/after-wave')
rows=[r for p in sorted(out.glob('group-*-rows.json')) for r in json.loads(p.read_text())]
assert len(rows)==79,len(rows)
assert len({r['roots'][0] for r in rows})==79
prefix=str(tree/'src/compiler')+'/'
def short(s):return s.removeprefix(prefix)
groups=collections.defaultdict(list)
for r in rows:groups[(short(r.get('where','-')),r['kind'],r.get('reason','-'))].append(short(r['roots'][0]))
old=json.loads((out/'previous-grouped-first-stops.json').read_text());oldmap={r:(g['where'],g['kind'],g['reason']) for g in old for r in g['roots']}
result=[]
for (where,kind,reason),roots in sorted(groups.items(),key=lambda x:(-len(x[1]),x[0])):
 result.append(dict(where=where,kind=kind,reason=reason,count=len(roots),roots=roots,new_group=not any((g['where'],g['kind'],g['reason'])==(where,kind,reason) for g in old)))
(out/'grouped-first-stops.json').write_text(json.dumps(result,indent=2)+'\n')
(out/'comparison.json').write_text(json.dumps([{'root':short(r['roots'][0]),'previous':oldmap[short(r['roots'][0])],'current':[short(r.get('where','-')),r['kind'],r.get('reason','-')]} for r in rows],indent=2)+'\n')
# Use the meter's own artifact checks only with a completed entry ledger.
sys.path.insert(0,str(repo/'stage3/meter'));import report
statusfile=scratch/'entry-latent-status.json';status=json.loads(statusfile.read_text()) if statusfile.exists() else None
if status and status['exit_code']==0:
 checks=report.root_order_checks(tree,scratch/'artifacts',rows)
 artifacts=[r for r in checks if r['node']['throws']]
 classification=dict(status='complete',readiness_node_checks=checks,root_order_artifacts=artifacts,totals=dict(source_files=79,ordinary_lowering=sum(r['kind']=='accepted' for r in rows),lowering=sum(r['kind']=='accepted' for r in rows)+sum(report.artifact_lowers_whole(r) for r in artifacts)))
else:
 # Incomplete entry evidence cannot prove eligibility under report.root_order_checks.
 pending=[dict(root=short(r['roots'][0]),where=short(r['where']),reason=r['reason']) for r in rows if r.get('reason') in report.READINESS_REASONS]
 classification=dict(status='incomplete',reason='The bounded entry census did not complete; the meter requires its attempted-source and finding ledger before classifying root-order artifacts.',entry_status=status,pending=pending,root_order_artifacts=[],totals=dict(source_files=79,ordinary_lowering=sum(r['kind']=='accepted' for r in rows),lowering=sum(r['kind']=='accepted' for r in rows)))
(out/'root-order.json').write_text(json.dumps(classification,indent=2)+'\n')
print(json.dumps(result,indent=2));print('artifact status',classification['status'])
