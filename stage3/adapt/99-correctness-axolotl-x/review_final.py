#!/usr/bin/env python3
"""Record final scope, input identity, formatting and process cleanup."""
import hashlib
import json
import os
from pathlib import Path
import subprocess
p=Path(__file__).resolve().parent;repo=p.parents[2]
def git(*args):return subprocess.check_output(['git',*args],cwd=repo,text=True).strip()
base='3e395f9e5a7038d859ec8930be72cde68d8be3ff'
assert git('branch','--show-current')=='typescript/maybebind-adapt'
assert git('merge-base',base,'HEAD')==base
prefix=str(p.relative_to(repo))+'/'
files=sorted(set(git('log','--format=','--name-only',base+'..HEAD').splitlines())-{''})
assert files and all(n.startswith(prefix) for n in files)
assert not any(n.endswith('.go') for n in files)
subprocess.run(['git','diff','--check'],cwd=repo,check=True)
subprocess.run(['git','diff','--check',base,'HEAD'],cwd=repo,check=True)
bad=[]
for f in p.rglob('*'):
 if not f.is_file() or f.suffix=='.gz':continue
 try:t=f.read_text()
 except UnicodeError:continue
 if '\u2013' in t or '\u2014' in t:bad.append(str(f.relative_to(p)))
assert not bad,bad
identity=json.loads((p/'owned/oracle/source-identity.json').read_text())
for name,sha in identity['adapter_sha256'].items():assert hashlib.sha256((p/name).read_bytes()).hexdigest()==sha,name
ancestors=set();pid=os.getpid()
while pid and pid not in ancestors:
 ancestors.add(pid)
 try:pid=int(Path('/proc',str(pid),'stat').read_text().rsplit(')',1)[1].split()[1])
 except (FileNotFoundError,ProcessLookupError):break
active=[]
for entry in Path('/proc').iterdir():
 if not entry.name.isdigit() or int(entry.name) in ancestors:continue
 try:
  state=(entry/'stat').read_text().rsplit(')',1)[1].split()[0]
  if state=='Z':continue
  cmd=(entry/'cmdline').read_bytes().replace(b'\0',b' ').decode(errors='replace')
  cwd=str((entry/'cwd').resolve())
  if any(s in cmd or s in cwd for s in ['/workspace/bind-','/workspace/scratch/bind-','stage3/adapt/99-correctness-axolotl-x/']):active.append(dict(pid=int(entry.name),command=cmd,cwd=cwd))
 except (FileNotFoundError,ProcessLookupError,PermissionError):pass
assert not active,active
out=dict(branch=git('branch','--show-current'),base=base,reviewed_head=git('rev-parse','HEAD'),all_commit_paths_under_adapter=True,committed_paths=len(files),git_diff_check='pass',go_file_names='pass',dash_check='pass',frozen_functional_inputs_identical=True,active_task_processes=active)
(p/'owned/final-review.json').write_text(json.dumps(out,indent=2)+'\n')
status=json.loads((p/'outcome.json').read_text());status['processes_remaining']=0
(p/'outcome.json').write_text(json.dumps(status,indent=2)+'\n')
print(json.dumps(out,indent=2))
