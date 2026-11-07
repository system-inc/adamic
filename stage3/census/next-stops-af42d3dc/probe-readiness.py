import json,sys
from pathlib import Path
sys.dont_write_bytecode=True
repo=Path('/workspace/adamic');out=Path(__file__).resolve().parent;tree=Path('/workspace/after-wave-adapted');scratch=Path('/workspace/scratch/after-wave')
sys.path.insert(0,str(repo/'stage3/meter'))
from node_readiness import run_roots
# Probe the three readiness-root candidates in the reference table. These
# observations alone never establish report.root_order_checks eligibility.
roots=[tree/'src/compiler'/n for n in ['debug.ts','diagnosticInformationMap.generated.ts','types.ts']]
result=run_roots(tree,roots,scratch/'readiness-candidates')
(out/'node-probes.json').write_text(json.dumps(result,indent=2)+'\n')
for root,r in result.items():print(Path(root).name,'build_child_exit='+str(r['build_exit']),'node_child_exit='+str(r['exit']),'files_child_exit='+str(r['program_files_exit']),r['first_error'],flush=True)
