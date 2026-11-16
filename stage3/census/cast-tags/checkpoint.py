from pathlib import Path
import json,time,subprocess
r=Path('/workspace/scratch/cast-tags');e=r/'evidence'
while True:
 ps=subprocess.check_output(['ps','-eo','pid,etimes,time,rss,args'],text=True)
 processes=[line.strip() for line in ps.splitlines() if '/workspace/scratch/cast-tags/latent /workspace/scratch/cast-tags/' in line]
 record={'utc':time.strftime('%Y-%m-%dT%H:%M:%SZ',time.gmtime()),'processes':processes,'outputs':{p.name:p.stat().st_size for p in e.glob('*.jsonl')},'completed':{p.name:json.loads(p.read_text()) for p in e.glob('*-status.json')}}
 with (e/'background-checkpoints.jsonl').open('a') as f:f.write(json.dumps(record)+'\n')
 if (e/'mutant-checker-status.json').exists() and (e/'mutant-generators-status.json').exists():break
 time.sleep(15)
