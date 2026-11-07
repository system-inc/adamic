"""Resume complete accepted pieces if the initial bounded controller expires."""
from pathlib import Path
import sys,time,json,os,signal,subprocess,shutil,tarfile
sys.dont_write_bytecode=True;sys.path.insert(0,'/workspace/scratch/debug-messages')
from work import run,s,repo
summary=s/'evidence/before-lane.json';result=s/'raw/before-lane.json'
while not result.exists():time.sleep(2)
if json.loads(summary.read_text())['complete']:raise SystemExit(0)
# No discarded partial unit may outlive its bounded parent.
killed=[]
for line in subprocess.check_output(['ps','-eo','pid,args'],text=True).splitlines():
 fields=line.strip().split(None,1)
 if len(fields)==2 and fields[1].startswith('python3 /workspace/adamic/stage3/lane/shard.py ') and fields[1].endswith(' /workspace/scratch/debug-messages/lane-before'):
  pid=int(fields[0]);killed.append(pid)
  try:os.killpg(pid,signal.SIGKILL)
  except ProcessLookupError:pass
state=json.loads(summary.read_text());completed=set(state['units']);partial=[]
results=s/'lane-before'
for directory in results.iterdir():
 if directory.name not in completed:
  rows=[]
  for p in sorted(directory.rglob('*')):
   if p.is_file() and 'adapted-tree' not in p.relative_to(directory).parts:rows.append({'path':str(p.relative_to(directory)),'text':p.read_text(errors='backslashreplace')})
  partial.append({'unit':directory.name,'raw':rows});shutil.rmtree(directory)
for directory in s.glob('lane-before-piece-*'):
 number=int(directory.name.rsplit('-',1)[1])
 if number>len(state['pieces']):shutil.rmtree(directory)
(s/'raw/discarded-incomplete-lane.json').write_text(json.dumps({'killedProcessGroups':killed,'parentResult':json.loads(result.read_text()),'discardedUnacceptedUnits':partial},indent=2)+'\n')
number=0
while not json.loads(summary.read_text())['complete']:
 number+=1
 run('before-lane-resume-'+str(number),['python3',repo/'stage3/adapt/99-correctness-core-debug-messages/lane-pieces.py','before','8'])
run('lane-reconciliation-run',['python3',repo/'stage3/adapt/99-correctness-core-debug-messages/reconcile.py'])
print('WHOLE ORACLE COMPLETE',flush=True)
