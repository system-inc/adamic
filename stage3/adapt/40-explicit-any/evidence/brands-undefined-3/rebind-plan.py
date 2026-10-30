import sys,json
from pathlib import Path
sys.dont_write_bytecode=True
lane=Path(sys.argv[1]);sys.path.insert(0,str(lane))
import shard_plan as s
p=lane/'shard-plan.json';plan=json.loads(p.read_text());old=plan['runtime'];plan['runtime']=s.runtime_hash()
for unit in plan['units']:
 definition={k:v for k,v in unit.items() if k!='inputs_hash'}
 unit['inputs_hash']=s.digest(dict(artifact=plan['artifact'],api=plan['api_digest'],runtime=plan['runtime'],unit=definition))
p.write_text(json.dumps(plan,indent=2)+'\n');(lane/'shards.json').write_text(json.dumps(s.gate_manifest(plan),indent=2)+'\n')
print(json.dumps({'old_runtime':old,'new_runtime':plan['runtime'],'units':len(plan['units']),'assignments_and_expectations_unchanged':True}))
