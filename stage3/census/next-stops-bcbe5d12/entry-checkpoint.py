import json,sys,time
from pathlib import Path
from supervise import run,scratch
out=Path(__file__).resolve().parent
while not (scratch/'entry-latent-status.json').exists():time.sleep(1)
status=json.loads((scratch/'entry-latent-status.json').read_text())
label='entry probe complete'
if status['exit_code']:
 times=scratch/'entry-unit-times.jsonl';rows=[json.loads(l) for l in times.read_text().splitlines()] if times.exists() else []
 record=dict(status=status,last_completed_timing=rows[-1] if rows else None,active_unit='not recorded; the last completed unit is not the interrupted unit',raw_records=len((scratch/'artifacts/tsc/latent.jsonl').read_bytes().splitlines()))
 (out/'ENTRY-STOP.json').write_text(json.dumps(record,indent=2)+'\n')
 label='entry probe stopped: '+str(status['stopping_reason'])
 with (out/'REPORT.md').open('a') as f:f.write('\nEntry probe stopped: '+str(status['stopping_reason'])+'; child exit '+str(status['exit_code'])+'. The ordinary root census continues.\n')
result=run('entry-checkpoint',[sys.executable,str(out/'checkpoint.py'),label],limit=1800)
sys.exit(result['exit_code'] if result['exit_code']>=0 else 128-result['exit_code'])
