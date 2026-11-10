from pathlib import Path
from datetime import datetime,timezone
import json
r=Path('/workspace/scratch/baobab-evidence');rows=[]
for s in (r/'whole-lane-go.jsonl').read_text().splitlines():
 try:rows.append(json.loads(s))
 except json.JSONDecodeError:pass
passed=[x for x in rows if x.get('Action')=='pass' and x.get('Test')];failed=[x for x in rows if x.get('Action')=='fail' and x.get('Test')]
records=[]
for p in (r/'lane-counts').glob('*.json'):
 try:records.append(json.loads(p.read_text()))
 except json.JSONDecodeError:pass
elapsed=(datetime.now(timezone.utc)-datetime.fromisoformat(rows[0]['Time'].replace('Z','+00:00'))).total_seconds()/60
progress=dict(unitsPassed=len(passed),unitsFailed=len(failed),processRecords=len(records),assertCallsSoFar=sum(x['total'] for x in records),differencesSoFar=sum(x['differences'] for x in records),laneMinutes=round(elapsed,1))
(r/'lane-progress.json').write_text(json.dumps(progress,indent=2)+'\n');print(json.dumps(progress))
if failed:
 for x in rows:
  if x.get('Test')==failed[0]['Test'] and x.get('Output'):print(x['Output'],end='')
