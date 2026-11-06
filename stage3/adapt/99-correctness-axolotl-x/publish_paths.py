#!/usr/bin/env python3
"""Serialize evidence-only commits and ordinary pushes from concurrent pieces."""
import fcntl
from pathlib import Path
import subprocess
import sys
repo=Path(__file__).resolve().parents[3];prefix='stage3/adapt/99-correctness-axolotl-x/'
message,*paths=sys.argv[1:];assert paths and all(p.startswith(prefix) for p in paths)
with Path('/workspace/scratch/bind-publish.lock').open('w') as lock:
    fcntl.flock(lock,fcntl.LOCK_EX)
    staged=subprocess.check_output(['git','diff','--cached','--name-only'],cwd=repo,text=True).splitlines()
    assert not staged,staged
    subprocess.run(['git','add',*paths],cwd=repo,check=True)
    subprocess.run(['git','diff','--cached','--check'],cwd=repo,check=True)
    if subprocess.run(['git','diff','--cached','--quiet'],cwd=repo).returncode:
        subprocess.run(['git','commit','-q','-m',message],cwd=repo,check=True)
        subprocess.run(['git','push','origin','HEAD:typescript/maybebind-adapt-2'],cwd=repo,check=True)
