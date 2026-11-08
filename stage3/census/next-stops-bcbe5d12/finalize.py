import sys,json,time
from pathlib import Path
from supervise import run,scratch
out=Path(__file__).resolve().parent
while not (scratch/'groups-status.json').exists():time.sleep(1)
assert json.loads((scratch/'groups-status.json').read_text())['exit_code']==0
for name,cmd in [
 ('summarize',[sys.executable,str(out/'summarize.py')]),
 ('source-excerpts',['node',str(out/'source-excerpts.cjs')]),
 ('map-emitters',[sys.executable,str(out/'map-emitters.py')]),
 ('root-order',[sys.executable,str(out/'classify-root-order.py')]),
 ('cleanup',[sys.executable,str(out/'cleanup.py')]),
 ('write-report',[sys.executable,str(out/'write-report.py')]),
 ('archive-evidence',[sys.executable,str(out/'archive-evidence.py')]),
 ('validation',[sys.executable,str(out/'validate.py')])]:
 result=run(name,cmd,limit=1800)
 if result['exit_code']:sys.exit(result['exit_code'] if result['exit_code']>0 else 128-result['exit_code'])
