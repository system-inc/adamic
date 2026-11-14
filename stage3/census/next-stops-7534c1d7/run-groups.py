"""Run the original bounded groups, checkpointing only completed pairs."""
import subprocess,sys
from pathlib import Path
base=Path(__file__).resolve().parent
for a,z in [(1,10),(11,20),(21,30),(31,40),(41,50),(51,60),(61,70),(71,79)]:
 subprocess.run([sys.executable,str(base/'run.py'),f'group-{a}-{z}',sys.executable,str(base/'group.py'),str(a),str(z)],check=True)
 subprocess.run([sys.executable,str(base/'checkpoint.py')],check=True)
