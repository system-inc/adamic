#!/usr/bin/env python3
"""Build the completed 95h fold combinations without importing unrelated branch edits."""
import json
from pathlib import Path
import shutil
import subprocess
import time
p=Path(__file__).resolve().parent;repo=p.parents[2];scratch=Path('/workspace/scratch/bind2');target=Path('/workspace/bind2-main322');fold='c2714b29c5937f553d79e98e9fae80aab8ed995a'
f=scratch/'apply-main/metrics.json'
while not f.exists():time.sleep(5)
assert json.loads(f.read_text())['exit']==0
# The original main apply is complete; subsequent matrix builds read its output tree.
shutil.copytree('/workspace/bind2-x-repo/stage3/adapt/99-correctness-axolotl-x',target/'stage3/adapt/99-correctness-axolotl-x',dirs_exist_ok=True)
subprocess.run(['git','restore','--source='+fold,'--staged','--worktree','stage3/adapt/95-checker-axolotl-h'],cwd=target,check=True)
subprocess.run(['git','add','stage3/adapt/99-correctness-axolotl-x'],cwd=target,check=True)
subprocess.run(['git','commit','-q','-m','Scratch composition snapshot: completed 95h fold and 99x'],cwd=target,check=True)
heads={}
for label in ['95hx','both']:
 if label=='both':
  subprocess.run(['git','restore','--source=origin/typescript/32-maybebind-fold','--staged','--worktree','stage3/adapt/32-indexed-reads-program'],cwd=target,check=True)
  subprocess.run(['git','commit','-q','-m','Scratch composition snapshot: both owner folds and 99x'],cwd=target,check=True)
 heads[label]=subprocess.check_output(['git','rev-parse','HEAD'],cwd=target,text=True).strip()
 (scratch/'95h-matrix-heads.json').write_text(json.dumps(dict(fold95h=fold,heads=heads),indent=2)+'\n')
 subprocess.run(['python3',str(p/'measure_piece.py'),str(scratch/('apply-'+label)),'--','bash',str(target/'stage3/apply.sh'),'--build-product','/workspace/bind2-'+label+'-tree'],cwd=target,check=True)
