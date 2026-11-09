import subprocess,json,hashlib,gzip,shutil,sys
from pathlib import Path
out=Path(__file__).resolve().parent;repo=out.parents[2]
def git(*args):return subprocess.check_output(['git',*args],cwd=repo,text=True).strip()
base=git('rev-parse','1060360b');assert git('rev-parse','HEAD')==base
for name in ['apply','a-entry-build','b-entry-build','a-entry-run','b-entry-run']:
 row=json.loads((out/(name+'.status.json')).read_text());assert row['exit']==0 and row['stopped'] is None,(name,row)
assert json.loads((out/'source-hash-check.json').read_text())['match']
assert json.loads((out/'comparison.json').read_text())['headers_equal']
assert json.loads((out/'comparison.json').read_text())['units_equal']
active=[]
for p in Path('/proc').iterdir():
 if not p.name.isdigit():continue
 try:
  args=p.joinpath('cmdline').read_bytes().decode().split('\0')
  if args and args[0] in ['/workspace/scratch/fa-'+side+'-entry-bin' for side in ['a','b']]:active.append(int(p.name))
 except (OSError,UnicodeError):pass
assert not active,active
(out/'processes-final.json').write_text(json.dumps(dict(live_census_processes=active),indent=2)+'\n')
for source in ['split-setup.log','split-npm.log','split-fetch.log','split-adapt-fetch.log','split-source.log','split-build-controller.log','split-launch.log']:
 p=Path('/tmp')/source
 if p.exists():(out/source.removeprefix('split-')).write_bytes(p.read_bytes())
manifest={}
for p in sorted(out.glob('*.jsonl')):
 data=p.read_bytes();manifest[p.name]=dict(bytes=len(data),sha256=hashlib.sha256(data).hexdigest())
 with p.with_suffix(p.suffix+'.gz').open('wb') as raw:
  with gzip.GzipFile(filename='',mode='wb',fileobj=raw,mtime=0) as target:target.write(data)
 assert gzip.decompress(p.with_suffix(p.suffix+'.gz').read_bytes())==data
 p.unlink()
(out/'stream-hashes.json').write_text(json.dumps(manifest,indent=2)+'\n')
shutil.rmtree(out/'__pycache__',ignore_errors=True)
assert not list(out.rglob('*.go'))
assert '\u2013' not in (out/'REPORT.md').read_text() and '\u2014' not in (out/'REPORT.md').read_text()
(out/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+p.relative_to(out).as_posix()+'\n' for p in sorted(out.rglob('*')) if p.is_file() and p.name!='SHA256SUMS'))
subprocess.run(['git','add','-f','--',str(out)],cwd=repo,check=True)
prefix='stage3/census/find-ancestor-split/'
files=git('diff','--cached','--name-only',base).splitlines();assert files and all(f.startswith(prefix) for f in files),files
with Path('/tmp/split-final-commit.log').open('w') as f:
 rc=subprocess.run(['git','commit','-m','stage3/census: isolate the two stacked findAncestor scan changes'],cwd=repo,stdout=f,stderr=subprocess.STDOUT).returncode
print('commit child_exit='+str(rc),flush=True)
if rc:sys.exit(rc)
sha=git('rev-parse','HEAD');assert git('rev-parse','HEAD^')==base
assert all(f.startswith(prefix) for f in git('diff-tree','--no-commit-id','--name-only','-r','HEAD').splitlines())
with Path('/tmp/split-final-push.log').open('w') as f:
 rc=subprocess.run(['git','push','origin','HEAD:refs/heads/typescript/find-ancestor-split-census'],cwd=repo,stdout=f,stderr=subprocess.STDOUT).returncode
print('push child_exit='+str(rc),flush=True)
if rc:sys.exit(rc)
assert git('ls-remote','origin','refs/heads/typescript/find-ancestor-split-census').split()[0]==sha
summary=dict(sha=sha,branch='typescript/find-ancestor-split-census',variants=json.loads((out/'comparison.json').read_text())['variants'])
Path('/tmp/split-final-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
print('sha='+sha,flush=True)
