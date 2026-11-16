"""Run each group with its own bound, checkpointing before the next group."""
import subprocess,sys
from pathlib import Path
b=Path(__file__).resolve().parent
start=int(sys.argv[1]) if len(sys.argv)>1 else 1
for first,last in [(1,10),(11,20),(21,30),(31,40),(41,50),(51,60),(61,70),(71,79)]:
 if first<start:continue
 for args in [[sys.executable,str(b/'run.py'),f'group-{first}-{last}',sys.executable,str(b/'group.py'),str(first),str(last)],[sys.executable,str(b/'checkpoint.py')]]:
  rc=subprocess.call(args);print('group driver child_exit='+str(rc),flush=True)
  if rc:sys.exit(rc)
