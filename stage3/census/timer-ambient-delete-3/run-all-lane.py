"""Run the named manifest as a conservative superset within a 1700-second lane budget."""
import json,os,subprocess,sys,time
from pathlib import Path
p=Path(__file__).parent;repo=Path('/workspace/timer-ambient-delete')
manifest=json.loads((repo/'stage3/lane/shards.json').read_text())
selection={'selection':'All 398 named manifest shards, conservatively covering every shard that reads the compiler built from sys.ts.','prior_full_lane_seconds':696.5383166630008,'budget_seconds':1700,'shards':[r['name'] for r in manifest['units']]}
(p/'lane-selection.json').write_text(json.dumps(selection,indent=2)+'\n')
env=dict(os.environ,TIMER_PROOF_REPO=str(repo),TIMER_PROOF_WORKERS='4');start=time.monotonic()
for piece in range(1,11):
 if time.monotonic()-start>=1650:
  print('Stopped: insufficient remaining total lane budget; Loom must finish the lane.',flush=True);sys.exit(124)
 result=subprocess.run(['python3',str(p/'run-command.py'),f'lane-piece-{piece:02d}','python3',str(p/'run-shards.py'),'/workspace/scratch/timer-ambient-delete-3-after','/workspace/scratch/timer-ambient-delete-3-lane'],cwd=repo,env=env)
 print(f'piece {piece}: child exit {result.returncode}',flush=True)
 if result.returncode:sys.exit(result.returncode)
(p/'lane-total.json').write_text(json.dumps({'seconds':time.monotonic()-start,'budget_seconds':1700,'pieces':10},indent=2)+'\n')
