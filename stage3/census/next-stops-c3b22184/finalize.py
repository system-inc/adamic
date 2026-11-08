import json,sys,time
from pathlib import Path
from supervise import run,scratch
out=Path(__file__).resolve().parent
while not (scratch/'groups-status.json').exists():time.sleep(1)
assert json.loads((scratch/'groups-status.json').read_text())['exit_code']==0
while not (scratch/'entry-checkpoint-status.json').exists():time.sleep(1)
assert json.loads((scratch/'entry-checkpoint-status.json').read_text())['exit_code']==0
for name,cmd in [
 ('analyze-stops',['node',str(out/'analyze-stops.cjs')]),
 ('node-probes',[sys.executable,str(out/'probe-readiness.py')]),
 ('map-emitters',[sys.executable,str(out/'map-emitters.py')]),
 ('root-order',[sys.executable,str(out/'classify-root-order.py')]),
 ('cleanup',[sys.executable,str(out/'cleanup.py')]),
 ('write-report',[sys.executable,str(out/'write-report.py')]),
 ('archive-evidence',[sys.executable,str(out/'archive-evidence.py')]),
 ('validation',[sys.executable,str(out/'validate.py')])]:
 status=run(name,cmd,limit=1800)
 if status['exit_code']:sys.exit(status['exit_code'] if status['exit_code']>0 else 128-status['exit_code'])
