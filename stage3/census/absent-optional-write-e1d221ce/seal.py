"""Seal verified report evidence and push one report-only commit on the parent."""
import hashlib,json,os,subprocess,sys
from pathlib import Path
out=Path(__file__).resolve().parent
repo=out.parents[2]
base='e3aa0bb7642163ebd004a71ba40a4657d27eab71'
ref='refs/heads/typescript/absent-optional-write-census'
assert json.loads((out/'audit.json').read_text())['source_hashes_unchanged']
assert json.loads((out/'pair.status.json').read_text())['exit']==0
assert json.loads((out/'control.status.json').read_text())['exit']==0
assert json.loads((out/'finish.status.json').read_text())['exit']==0
live=[]
for proc in Path('/proc').iterdir():
 if not proc.name.isdigit():continue
 try:
  args=proc.joinpath('cmdline').read_bytes().decode().split('\0')
  if args[0] in ['/workspace/aow-parent-census','/workspace/aow-head-census','/workspace/aow-unprimed-head-census']:
   live.append(dict(pid=int(proc.name),command=args))
 except (FileNotFoundError,ProcessLookupError,PermissionError,UnicodeError):pass
assert not live,live
(out/'processes-final.json').write_text(json.dumps(dict(live_census_processes=live),indent=2)+'\n')
assert not list(out.rglob('*.go'))
assert not list(out.rglob('*.pyc'))
for p in out.rglob('*.md'):
 s=p.read_text();assert '\u2013' not in s and '\u2014' not in s,p
(out/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+p.relative_to(out).as_posix()+'\n' for p in sorted(out.rglob('*')) if p.is_file() and p.name!='SHA256SUMS'))
def git(*args):return subprocess.check_output(['git',*args],cwd=repo,text=True).strip()
def run(name,args):
 with Path('/workspace/aow-final-'+name+'.log').open('w') as log:
  rc=subprocess.run(args,cwd=repo,stdout=log,stderr=subprocess.STDOUT).returncode
 print(name+' child_exit='+str(rc),flush=True)
 assert rc==0,(name,rc)
assert git('rev-parse','HEAD')==base
run('stage',['git','add','-f','--',str(out)])
files=git('diff','--cached','--name-only',base).splitlines()
assert files and all(f.startswith('stage3/census/absent-optional-write-e1d221ce/') for f in files)
authored=[str(p.relative_to(repo)) for p in out.glob('*.py')]+[str((out/'REPORT.md').relative_to(repo))]
run('diff-check',['git','diff','--cached','--check','--',*authored])
run('commit',['git','commit','-m','stage3/census: pair absent optional write e1d221ce against e3aa0bb7'])
sha=git('rev-parse','HEAD')
assert git('rev-parse','HEAD^')==base
assert git('rev-list','--count',base+'..HEAD')=='1'
run('push',['git','push','origin','HEAD:'+ref])
assert git('ls-remote','--heads','origin',ref).split()[0]==sha
print(json.dumps(dict(sha=sha,ref=ref,files=len(files),parent=base),indent=2),flush=True)
