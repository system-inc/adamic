"""Finish the remaining lane in separately receipted, bounded 40-shard pieces."""
import os,subprocess,sys
from pathlib import Path
root=Path(__file__).parent
env=dict(os.environ,TIMER_PROOF_REPO='/workspace/timer-ambient-delete',TIMER_PROOF_WORKERS='4')
for piece in range(5,11):
 command=['python3',str(root/'run-command.py'),f'lane-piece-{piece:02d}','python3',str(root/'run-shards.py'),'/workspace/scratch/timer-road5-after','/workspace/scratch/timer-road5-lane']
 result=subprocess.run(command,cwd='/workspace/timer-ambient-delete',env=env)
 print(f'piece {piece}: child exit {result.returncode}',flush=True)
 if result.returncode:sys.exit(result.returncode)
