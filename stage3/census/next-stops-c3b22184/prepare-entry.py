import json,sys,time
from pathlib import Path
from supervise import run,scratch
out=Path(__file__).resolve().parent
while not (scratch/'setup-status.json').exists():time.sleep(1)
assert json.loads((scratch/'setup-status.json').read_text())['exit_code']==0
for name,cmd in [('overlay',[sys.executable,str(out/'build-entry.py')]),('entry-build',['go','build','-buildvcs=false','-overlay='+str(scratch/'entry-overlay/overlay.json'),'-o',str(scratch/'entry-census'),'./stage3/census/latent/tool'])]:
 status=run(name,cmd,limit=1800)
 if status['exit_code']:
  (scratch/'entry-preparation-status.json').write_text(json.dumps(status,indent=2)+'\n');sys.exit(status['exit_code'] if status['exit_code']>0 else 128-status['exit_code'])
while not (scratch/'ordinary-build-status.json').exists():time.sleep(1)
assert json.loads((scratch/'ordinary-build-status.json').read_text())['exit_code']==0
status=run('entry-runner',[sys.executable,str(out/'run-entry.py')],limit=5400,watch=[scratch/'entry-unit-times.jsonl',scratch/'artifacts/tsc/latent.jsonl'])
sys.exit(status['exit_code'] if status['exit_code']>=0 else 128-status['exit_code'])
