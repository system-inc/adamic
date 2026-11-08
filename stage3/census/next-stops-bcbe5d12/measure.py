import json,sys,time
from pathlib import Path
from supervise import run,scratch
out=Path(__file__).resolve().parent
while not Path('/workspace/adamic/cohere/TypeScript/tsc/go.mod').exists(): time.sleep(1)
for name,cmd in [
 ('ordinary-build-complete',['go','build','-o',str(scratch/'census'),'./stage3/census/tool']),
 ('overlay-complete',[sys.executable,str(out/'build-entry.py')]),
 ('entry-build',['go','build','-buildvcs=false','-overlay='+str(scratch/'entry-overlay/overlay.json'),'-o',str(scratch/'entry-census'),'./stage3/census/latent/tool'])]:
 status=run(name,cmd,limit=1800)
 if status['exit_code']:sys.exit(status['exit_code'] if status['exit_code']>0 else 128-status['exit_code'])
while not (scratch/'apply-checkpoint-status.json').exists():time.sleep(1)
assert json.loads((scratch/'apply-checkpoint-status.json').read_text())['exit_code']==0
status=run('groups',[sys.executable,str(out/'run-groups.py')],limit=5400,watch=[out/f'group-{i}.jsonl' for i in range(1,5)])
sys.exit(status['exit_code'] if status['exit_code']>=0 else 128-status['exit_code'])
