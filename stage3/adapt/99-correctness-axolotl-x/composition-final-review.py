#!/usr/bin/env python3
"""Validate shipping scope, packed bytes, frozen inputs and completed process cleanup."""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import tarfile
p=Path(__file__).resolve().parent;repo=p.parents[2];base='a6d5f494e46262f805e5be76a4d7c108fe6a3a6c'
def git(*a):return subprocess.check_output(['git',*a],cwd=repo,text=True).strip()
assert git('branch','--show-current')=='typescript/maybebind-adapt-2'
assert git('merge-base',base,'HEAD')==base
prefix=str(p.relative_to(repo))+'/'
files=set(git('log','--format=','--name-only',base+'..HEAD').splitlines())-{''};assert files and all(f.startswith(prefix) for f in files)
assert not any(f.endswith('.go') for f in files)
for name in ['owners.json','inventory.json','ownership.json']:assert not git('diff',base,'--',prefix+name)
subprocess.run(['git','diff','--check'],cwd=repo,check=True);subprocess.run(['git','diff','--check',base,'HEAD'],cwd=repo,check=True)
bad=[]
for f in p.rglob('*'):
 if not f.is_file() or f.suffix=='.gz':continue
 try:s=f.read_text()
 except UnicodeError:continue
 if '\u2013' in s or '\u2014' in s:bad.append(str(f.relative_to(p)))
assert not bad,bad
hashes={}
for label in ['x','32x']:
 frozen=Path('/workspace/bind2-'+label+'-repo/stage3/adapt/99-correctness-axolotl-x')
 for name in ['adapt.cjs','engine.cjs','owners.json','inventory.json','ownership.json']:
  assert (p/name).read_bytes()==(frozen/name).read_bytes();hashes[name]=hashlib.sha256((p/name).read_bytes()).hexdigest()
archives=[]
for archive in sorted((p/'composition').rglob('*.tar.gz')):
 digest=hashlib.sha256(archive.read_bytes()).hexdigest();assert Path(str(archive)+'.sha256').read_text()==digest+'  '+archive.name+'\n'
 manifest=archive.with_suffix('').with_suffix('.sha256');expected=dict(line.split('  ',1)[::-1] for line in manifest.read_text().splitlines());observed={}
 with tarfile.open(archive) as stream:
  for member in stream.getmembers():
   assert member.isfile() and not member.name.startswith('/') and '..' not in Path(member.name).parts
   assert member.name not in observed
   observed[member.name]=hashlib.sha256(stream.extractfile(member).read()).hexdigest()
 assert expected==observed,str(archive)
 archives.append(dict(path=str(archive.relative_to(p)),files=len(observed),bytes=archive.stat().st_size,sha256=digest))
ancestors=set();pid=os.getpid()
while pid and pid not in ancestors:
 ancestors.add(pid)
 try:pid=int(Path('/proc',str(pid),'stat').read_text().rsplit(')',1)[1].split()[1])
 except (OSError,ProcessLookupError):break
active=[]
for d in Path('/proc').iterdir():
 if not d.name.isdigit() or int(d.name) in ancestors:continue
 try:
  if (d/'stat').read_text().rsplit(')',1)[1].split()[0]=='Z':continue
  command=(d/'cmdline').read_bytes().replace(b'\0',b' ').decode(errors='replace');cwd=str((d/'cwd').resolve())
  if any(s in command or s in cwd for s in ['/workspace/bind2-','/workspace/scratch/bind2','composition-oracle.py','composition-matrix.py','composition-latent.py','composition-95h-matrix.py']):active.append(dict(pid=int(d.name),command=command,cwd=cwd))
 except (OSError,ProcessLookupError):pass
assert not active,active
result=dict(branch=git('branch','--show-current'),base=base,reviewed_head=git('rev-parse','HEAD'),all_commit_paths_under_99x=True,own_pins_unchanged=True,no_added_go_files=True,git_diff_check='pass',dash_check='pass',functional_sha256=hashes,archives=archives,active_task_processes=active)
(p/'composition/final-review.json').write_text(json.dumps(result,indent=2)+'\n')
f=p/'outcome.json';r=json.loads(f.read_text());r['processes_remaining']=0;r['packed_archives_verified']=len(archives);f.write_text(json.dumps(r,indent=2)+'\n')
print(json.dumps(dict(status='pass',archives_verified=len(archives),active_task_processes=active)))
