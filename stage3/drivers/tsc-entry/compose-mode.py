#!/usr/bin/env python3
"""Reproduce this experiment's disposable, explicitly reconciled compiler."""
import argparse,json,subprocess
from pathlib import Path
p=argparse.ArgumentParser();p.add_argument('scratch',type=Path);a=p.parse_args();here=Path(__file__).resolve().parent;repo=here.parents[2];s=a.scratch.resolve()
if s.exists():raise SystemExit('scratch must not exist')
v=json.loads((here/'evidence/mode-c09c22b8/provenance.json').read_text());pins=v['pins']
def run(args,cwd):subprocess.run(args,cwd=cwd,check=True)
run(['git','worktree','add','--detach',str(s),pins['origin/area/compiler']],repo)
run(['git','merge','--no-edit','-X','theirs',pins['c09c22b8']],s)
for ref in ['origin/codex/stricter-records','origin/codex/stricter-indexed-all','origin/codex/stricter-catch-variables']:run(['git','merge','--no-edit','-X','ours',pins[ref]],s)
for name,digest in v['uncommitted_c09_selected_paths'].items():
 if digest is None:(s/name).unlink(missing_ok=True)
 else:run(['git','restore','--source',pins['c09c22b8'],'--',name],s)
print('Scratch compiler core changes remain uncommitted; initialize pinned submodules and build from this directory.')
