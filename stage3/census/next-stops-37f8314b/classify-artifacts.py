from pathlib import Path
import json,sys
sys.dont_write_bytecode=True
repo=Path('/workspace/adamic');out=Path(__file__).resolve().parent
sys.path.insert(0,str(repo/'stage3/meter'))
import report
scratch=Path('/workspace/scratch/next-stops-37f8314b/artifacts')
tree=Path('/workspace/next-stops-37f8314b-adapted')
rows=[r for p in sorted(out.glob('group-*-rows.json')) for r in json.loads(p.read_text())]
probes=json.loads((out/'node-probes.json').read_text())
report.run_roots=lambda tree,roots,logs:{str(Path(r).resolve()):probes[str(Path(r).resolve())] for r in roots}
status=json.loads((out/'entry-latent-status.json').read_text())
if status['exit']==0:
 checks=report.root_order_checks(tree,scratch,rows)
 artifacts=[r for r in checks if r['node']['throws']]
 result=dict(status='complete',root_order_artifacts=artifacts,readiness_node_checks=checks,totals=dict(lowering=sum(r['kind']=='accepted' for r in rows)+sum(report.artifact_lowers_whole(r) for r in artifacts),source_files=79,artifact_total=len(artifacts)))
else:
 result=dict(status='incomplete',reason='Meter entry census did not complete within its 1800-second bound; attempted-source and entry-finding evidence is incomplete.',entry_status=status,root_order_artifacts=[],pending=[dict(root=Path(k).relative_to(tree/'src/compiler').as_posix(),node=v) for k,v in probes.items()],totals=dict(lowering=sum(r['kind']=='accepted' for r in rows),source_files=79,artifact_total=None))
(out/'root-order.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps({k:v for k,v in result.items() if k not in ['pending','readiness_node_checks','root_order_artifacts']},indent=2))
