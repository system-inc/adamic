"""Audit and publish one report commit. --stop records a bounded incomplete result."""
import datetime,hashlib,json,os,signal,subprocess,sys,time
from pathlib import Path
out=Path(__file__).resolve().parent;repo=out.parents[2]
base='b5c0aa648a7a35b2f17899895d009eabd4147000'
stop='--stop' in sys.argv
reason=sys.argv[sys.argv.index('--reason')+1] if '--reason' in sys.argv else 'The user-provided approximately two-hour session budget was reached before all entry scans completed.'
binaries={f'/tmp/branded-{side}-{kind}' for side in ['parent','head'] for kind in ['census','core','focus']}
controllers={'measure.py','run.py','start-B-head.py','run-core.py','run-focus.py','serializer.py','checkpoint.py','inspect-results.py'}

def processes():
 go=[];control=[]
 for p in Path('/proc').iterdir():
  if not p.name.isdigit() or int(p.name)==os.getpid():continue
  try:
   args=p.joinpath('cmdline').read_bytes().decode().split('\0');pid=int(p.name)
   if args[0] in binaries:go.append((pid,args))
   if len(args)>1 and Path(args[1]).name in controllers and 'branded-call-351fc63f' in args[1]:control.append((pid,args))
  except (FileNotFoundError,ProcessLookupError,PermissionError,UnicodeError):continue
 return go,control

def ready(name):
 p=out/(name+'.status.json')
 return p.exists() and json.loads(p.read_text())['exit']==0

if stop:
 events=[]
 for pid,args in processes()[0]:
  try:os.killpg(pid,signal.SIGTERM);events.append(dict(pid=pid,args=args,signal='SIGTERM'))
  except ProcessLookupError:pass
 (out/'budget-stop-events.json').write_text(json.dumps(dict(utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),reason=reason,processes=events),indent=2)+'\n')
for _ in range(35):
 go,control=processes()
 if not go and not control:break
 if not stop and go:raise SystemExit('census still active; no report commit made')
 time.sleep(1)
go,control=processes()
assert not go and not control,('task processes remain',go,control)
complete=all(ready(f'{a}-{b}-entry') for a in ['A','B'] for b in ['parent','head'])
assert complete or stop,'entry runs incomplete'
if not complete:(out/'STOP_REASON.txt').write_text(reason+'\n')
(out/'processes-final.json').write_text(json.dumps(dict(utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),live_census_processes=go,live_controllers=control),indent=2)+'\n')
(out/'memory-events-final.txt').write_text(Path('/sys/fs/cgroup/memory.events').read_text())

for name,cmd in [('report-build',[sys.executable,str(out/'write-report.py')]),('audit',[sys.executable,str(out/'audit-results.py')]+([] if complete else ['--partial']))]:
 with (out/(name+'.log')).open('w') as log:rc=subprocess.run(cmd,cwd=repo,stdout=log,stderr=subprocess.STDOUT,env=dict(os.environ,PYTHONDONTWRITEBYTECODE='1')).returncode
 print(name,'child_exit='+str(rc),flush=True)
 assert rc==0,(name,rc)
# Bytecode from the earlier pairing tests is unrelated scratch output.
import shutil
shutil.rmtree(repo/'stage3/census/pair/__pycache__',ignore_errors=True)
shutil.rmtree(out/'__pycache__',ignore_errors=True)
assert not list(out.rglob('*.go'))
files=[p for p in out.rglob('*') if p.is_file() and p.name!='SHA256SUMS']
(out/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+p.relative_to(out).as_posix()+'\n' for p in sorted(files)))

def git(*args):return subprocess.check_output(['git',*args],cwd=repo,text=True).strip()
assert git('rev-parse','HEAD')==base,'report branch no longer directly on requested parent'
assert git('branch','--show-current')=='typescript/branded-call-census'
subprocess.run(['git','add','-f','--',str(out)],cwd=repo,check=True)
changed=git('diff','--cached','--name-only',base).splitlines()
assert changed and all(p.startswith('stage3/census/branded-call-351fc63f/') for p in changed)
authored=[str(p.relative_to(repo)) for p in out.glob('*.py')]+[str(p.relative_to(repo)) for p in out.glob('*.cjs')]+[str((out/'REPORT.md').relative_to(repo))]
with Path('/tmp/branded-whitespace-check.log').open('w') as log:rc=subprocess.run(['git','diff','--cached','--check','--',*authored],cwd=repo,stdout=log,stderr=subprocess.STDOUT).returncode
assert rc==0,'authored whitespace check failed'
with Path('/tmp/branded-final-commit.log').open('w') as log:rc=subprocess.run(['git','commit','-m','stage3/census: pair branded-call 351fc63f against b5c0aa64'],cwd=repo,stdout=log,stderr=subprocess.STDOUT).returncode
assert rc==0,'commit failed'
sha=git('rev-parse','HEAD');assert git('rev-parse','HEAD^')==base
with Path('/tmp/branded-final-push.log').open('w') as log:rc=subprocess.run(['git','push','origin','HEAD:refs/heads/typescript/branded-call-census'],cwd=repo,stdout=log,stderr=subprocess.STDOUT).returncode
assert rc==0,'push failed'
remote=git('ls-remote','--heads','origin','refs/heads/typescript/branded-call-census').split()[0]
assert remote==sha
summary={}
for a in ['A','B']:
 p=out/f'{a}-entry-pair.json'
 if p.exists():
  d=json.loads(p.read_text());summary[a]={v:{k:len(d[v][k]) for k in ['lowered','new_findings','moved','regressions']} for v in ['raw_findings','distinct_reads']}
 else:summary[a]={'complete':False}
print(json.dumps(dict(sha=sha,branch='typescript/branded-call-census',complete=complete,entry=summary),indent=2),flush=True)
