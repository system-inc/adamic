import sys
sys.dont_write_bytecode=True
from run import *
assert run('branch',['git','switch','-c','typescript/core-sys-timer-slot-2'])==0
assert run('stack',['git','reset','--soft','5ba70a9486a9d297f42a6a29fd9b2388caaf093d'])==0
assert run('generate-api',['python3','stage3/lane/generate-api.py','/workspace/typescript-source'])==0
sys.path.insert(0,str(ROOT/'stage3/lane'))
import shard_plan
p=shard_plan.load_plan();old=p['runtime'];p['runtime']=shard_plan.runtime_hash();(ROOT/'stage3/lane/shard-plan.json').write_text(json.dumps(p,indent=2)+'\n')
assert run('generate-shards-check',['python3','stage3/lane/generate_shards.py','--check'])==0
(S/'runtime.json').write_text(json.dumps({'before':old,'after':p['runtime']},indent=2)+'\n')
