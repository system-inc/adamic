import sys
sys.dont_write_bytecode=True
from run import *
assert run('head2-generate-api',['node','stage3/lane/generate-api.cjs','/workspace/typescript-source'])==0
import shutil
shutil.copyfile(S/'head2-generate-api.log',ROOT/'stage3/lane/sanctioned-api.json')
sys.path.insert(0,str(ROOT/'stage3/lane'))
import shard_plan
p=shard_plan.load_plan();p['runtime']=shard_plan.runtime_hash();(ROOT/'stage3/lane/shard-plan.json').write_text(json.dumps(p,indent=2)+'\n')
assert run('head2-generate-shards',['python3','stage3/lane/generate_shards.py'])==0
assert run('head2-check-shards',['python3','stage3/lane/generate_shards.py','--check'])==0
