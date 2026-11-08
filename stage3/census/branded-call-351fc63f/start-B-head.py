"""Overlap only the late portion of B parent with its independent head run."""
import datetime,json,subprocess,sys,time
from pathlib import Path
out=Path(__file__).resolve().parent;repo=out.parents[2]
while True:
 rows=[json.loads(line) for line in (out/'B-parent-entry-units.jsonl').read_text().splitlines()]
 files=len({r['file'] for r in rows})
 stat={k:int(v) for k,v,*_ in (line.split() for line in Path('/sys/fs/cgroup/memory.stat').read_text().splitlines())}
 hard=stat['anon']+stat['shmem']
 if files>=65 and hard<12*1024**3:break
 print('waiting for late parent segment: files='+str(files)+' anon_plus_shmem='+str(hard),flush=True)
 time.sleep(10)
trigger=dict(utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),parent_files_timed=files,anon_plus_shmem_bytes=hard)
(out/'B-head-overlap-trigger.json').write_text(json.dumps(trigger,indent=2)+'\n')
(out/'concurrency-note.txt').write_text('Pairings overlap. A parent/head remain sequential. B head starts during the late part of its parent, after at least 65 files have timed attempts and anon plus shmem is below 12 GiB. A per-run lock and binary/command checks prevent a duplicate B head. Completed core and focused streams are reused. Elapsed times are shared-host observations, not performance benchmarks.\n')
name='B-head-entry'
cmd=['env','-u','LATENT_FULL','-u','LATENT_SPECULATIVE','-u','LATENT_ONLY_FILE','LATENT_ASSERT_NO_OUTPUT=1',f'LATENT_TIMES={out}/{name}-units.jsonl','/tmp/branded-head-census','/tmp/branded-B/src/tsc/tsc.ts',str(out/(name+'.jsonl'))]
rc=subprocess.run([sys.executable,str(out/'run.py'),name,str(repo),*cmd]).returncode
print('B early head wrapper child_exit='+str(rc),flush=True)
raise SystemExit(rc)
