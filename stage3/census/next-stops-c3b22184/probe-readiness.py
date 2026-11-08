import json,sys
from pathlib import Path
sys.dont_write_bytecode=True
repo=Path('/workspace/adamic');out=Path(__file__).resolve().parent;tree=Path('/workspace/c3b22184-adapted');scratch=Path('/workspace/scratch/c3b22184')
sys.path.insert(0,str(repo/'stage3/meter'))
from node_readiness import run_roots
from report import READINESS_REASONS
# Node throws do not by themselves establish report.root_order_checks eligibility.
rows=sum([json.loads((out/f'group-{i}-rows.json').read_text()) for i in range(1,5)],[])
roots=[row['roots'][0] for row in rows if row.get('reason') in READINESS_REASONS]
result=run_roots(tree,roots,scratch/'readiness-candidates')
(out/'node-probes.json').write_text(json.dumps(result,indent=2)+'\n')
for root,r in result.items():print(Path(root).name,'build_child_exit='+str(r['build_exit']),'node_child_exit='+str(r['exit']),'files_child_exit='+str(r['program_files_exit']),r['first_error'],flush=True)
