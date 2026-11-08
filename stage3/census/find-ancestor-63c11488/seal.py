"""Seal only this report directory as one commit directly on the requested parent."""
import hashlib,json,subprocess,shutil
from pathlib import Path
out=Path(__file__).resolve().parent;repo=out.parents[2]
def git(*args):return subprocess.check_output(['git',*args],cwd=repo,text=True).strip()
base=git('rev-parse','1060360b')
assert git('rev-parse','HEAD')==base
assert (out/'audit.json').exists()
active=[]
for proc in Path('/proc').iterdir():
 if not proc.name.isdigit():continue
 try:
  args=proc.joinpath('cmdline').read_bytes().decode().split('\0')
  if args and args[0] in ['/workspace/scratch/fa-'+side+'-'+kind for side in ['parent','head'] for kind in ['full','focus','entry-bin']]:active.append(int(proc.name))
 except (OSError,UnicodeError):pass
assert not active,active
(out/'processes-final.json').write_text(json.dumps(dict(live_compiler_processes=active),indent=2)+'\n')
for src,dst in [('fa-stacked-setup.log','setup.log'),('fa-stacked-builds.log','build-controller.log'),('fa-stacked-census.log','census-controller.log'),('fa-stacked-source-clone.log','source-clone.log'),('fa-stacked-source-checkout.log','source-checkout.log'),('fa-stacked-finish.log','report-render.log')]:
 p=Path('/tmp')/src
 if p.exists():(out/dst).write_bytes(p.read_bytes())
shutil.rmtree(out/'__pycache__',ignore_errors=True)
assert not list(out.rglob('*.go'))
report=(out/'REPORT.md').read_text();assert '\u2013' not in report and '\u2014' not in report
(out/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+p.relative_to(out).as_posix()+'\n' for p in sorted(out.rglob('*')) if p.is_file() and p.name!='SHA256SUMS'))
subprocess.run(['git','add','-f','--',str(out)],cwd=repo,check=True)
files=git('diff','--cached','--name-only',base).splitlines();prefix='stage3/census/find-ancestor-63c11488/'
assert files and all(f.startswith(prefix) for f in files),files
with Path('/tmp/fa-stacked-final-commit.log').open('w') as f:
 rc=subprocess.run(['git','commit','-m','stage3/census: pair stacked findAncestor 63c11488 against road B'],cwd=repo,stdout=f,stderr=subprocess.STDOUT).returncode
print('commit child_exit='+str(rc),flush=True);assert rc==0
sha=git('rev-parse','HEAD');assert git('rev-parse','HEAD^')==base
assert all(f.startswith(prefix) for f in git('diff-tree','--no-commit-id','--name-only','-r','HEAD').splitlines())
with Path('/tmp/fa-stacked-final-push.log').open('w') as f:
 rc=subprocess.run(['git','push','origin','HEAD:refs/heads/typescript/find-ancestor-stacked-census'],cwd=repo,stdout=f,stderr=subprocess.STDOUT).returncode
print('push child_exit='+str(rc),flush=True);assert rc==0
remote=git('ls-remote','origin','refs/heads/typescript/find-ancestor-stacked-census').split()[0];assert remote==sha
summary=dict(sha=sha,branch='typescript/find-ancestor-stacked-census',counts={})
d=json.loads((out/'entry-paired.json').read_text())
for view in ['raw_findings','distinct_reads']:
 summary['counts'][view]={k:len(d[view][k]) for k in ['lowered','new_findings','moved','new_sites','unit_first_stop_delta','regressions']}
Path('/tmp/fa-stacked-final-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps(summary,indent=2),flush=True)
