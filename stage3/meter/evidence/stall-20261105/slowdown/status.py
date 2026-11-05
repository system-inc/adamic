from pathlib import Path
import json,datetime
r=Path('/tmp/meter-stall-20261105/slowdown')
for label in ['good','bad']:
 d=r/label
 if not (d/'start.json').exists():continue
 start=json.loads((d/'start.json').read_text());elapsed=(datetime.datetime.now(datetime.timezone.utc)-datetime.datetime.fromisoformat(start['started_utc'])).total_seconds()
 if (d/'result.json').exists():elapsed=json.loads((d/'result.json').read_text())['wall_seconds']
 a=[json.loads(s) for s in (d/'times.jsonl').read_text().splitlines()] if (d/'times.jsonl').exists() else []
 print(label,'elapsed_seconds',round(elapsed),'completed_units',len(a),'last_unit',a[-1]['where'].split('/adapted/')[-1] if a else None,'profile_bytes',(d/'cpu.pprof').stat().st_size if (d/'cpu.pprof').exists() else 0)
 if (d/'result.json').exists():print(label,'finished')
