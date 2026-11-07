import sys
sys.dont_write_bytecode=True
from run import *
import shutil
root=S/'lane-head2';probe=S/'lane-serial-probe';name='stage3-lane-async-function-cases-00';target=root/'attempts/four-way';target.mkdir(parents=True,exist_ok=True)
for p in [root/name,root/(name+'.log')]:
 if p.exists():shutil.move(str(p),str(target/p.name))
shutil.move(str(probe/name),str(root/name));shutil.copyfile(S/'head2-serial-async-control.log',root/(name+'.log'))
if (root/name/'adapted-tree').exists():shutil.rmtree(root/name/'adapted-tree')
r=json.loads((root/name/'report.json').read_text());assert r['status']=='pass'
u=next(u for u in json.loads((ROOT/'stage3/lane/shards.json').read_text())['units'] if u['name']==name)
row=dict(name=name,command=u['command'],exit=0,seconds=json.loads((S/'head2-serial-async-control-status.json').read_text())['seconds'],status=r['status'],counts=r['counts'],errors=r['errors'],attempt='standalone-control')
with (root/'ledger.jsonl').open('a') as stream:stream.write(json.dumps(row)+'\n')
for index in range(4):
 assert run('head2-lane-piece-'+str(index),['python3','-B',ROOT/'stage3/adapt/99-correctness-core-sys-timer-slot/run-shards.py',S/'head2',root])==0
 result=json.loads((root/'ledger.json').read_text())
 if result['all_commands_executed']:break
assert result['all_commands_executed']
print('complete full manifest child_exit=0',flush=True)
