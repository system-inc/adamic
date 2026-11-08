import hashlib,json,sys,time
from pathlib import Path
from supervise import run,scratch
out=Path(__file__).resolve().parent
while not (scratch/'setup-status.json').exists():time.sleep(1)
assert json.loads((scratch/'setup-status.json').read_text())['exit_code']==0
status=run('ordinary-build',['go','build','-o',str(scratch/'census'),'./stage3/census/tool'],limit=1800)
if status['exit_code']:sys.exit(status['exit_code'] if status['exit_code']>0 else 128-status['exit_code'])
status=run('ordinary-buildinfo',['go','version','-m',str(scratch/'census')],limit=1800)
(out/'census-buildinfo.txt').write_bytes((scratch/'ordinary-buildinfo.log').read_bytes())
(out/'binaries.json').write_text(json.dumps({'ordinary':{'path':str(scratch/'census'),'sha256':hashlib.sha256((scratch/'census').read_bytes()).hexdigest()}},indent=2)+'\n')
while not (scratch/'apply-checkpoint-status.json').exists():time.sleep(1)
assert json.loads((scratch/'apply-checkpoint-status.json').read_text())['exit_code']==0
status=run('groups',[sys.executable,str(out/'run-groups.py')],limit=5400,watch=[out/f'group-{i}.jsonl' for i in range(1,5)])
sys.exit(status['exit_code'] if status['exit_code']>=0 else 128-status['exit_code'])
