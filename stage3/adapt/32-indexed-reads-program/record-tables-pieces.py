from pathlib import Path
import json,hashlib,os,subprocess,signal,shutil,time,tarfile,collections,fcntl
from concurrent.futures import ThreadPoolExecutor,as_completed
repo=Path('/workspace/adamic');s=Path('/workspace/scratch/gov-b');out=repo/'stage3/adapt/32-indexed-reads-program';plan=json.loads((repo/'stage3/lane/shard-plan.json').read_text());units=plan['units'];results=s/'results';results.mkdir(exist_ok=True)
summary={'ruling':'b5w3ycg, Nov 7 00:55Z, road (b)','base':subprocess.check_output(['git','rev-parse','HEAD'],cwd=repo,text=True).strip(),'units':len(units),'completedUnits':[],'pieces':[],'calls':0,'beforeArrayCalls':0,'afterArrayCalls':0,'baselineSparseCalls':0,'returnedElements':0,'idWrites':0,'nonCanonicalKeys':0,'nonIntegerIds':0,'valueDifferences':0,'identityProjectionFailures':0,'oracleCounts':{side:{'pass':0,'fail':0,'skip':0} for side in ['before','after']},'complete':False}
if (out/'RECORD-TABLE-COUNT.json').exists():summary=json.loads((out/'RECORD-TABLE-COUNT.json').read_text())
def snapshot(side):return {str(p.relative_to(s/side)):hashlib.sha256(p.read_bytes()).hexdigest() for base in ['src','built/local'] for p in sorted((s/side/base).rglob('*')) if p.is_file()}
frozen={side:snapshot(side) for side in ['before','after']};(s/'raw/frozen-manifests.json').write_text(json.dumps(frozen,indent=2)+'\n')
def git(*args):
 with (s/'git-pieces.log').open('a') as log:subprocess.run(['git',*args],cwd=repo,stdout=log,stderr=log,check=True)
def run(unit,directory):
 name=unit['name'];observation={}
 for side in ['before','after']:
  root=directory/side;calls=root/'calls'/name;calls.mkdir(parents=True)
  env=dict(os.environ,STAGE3_ADAPTED_TREE=str(s/side),STAGE3_RESULTS=str(results/side),GOVB_DIR=str(calls),STAGE3_SHARD_SECONDS='90',NODE_OPTIONS='--require='+str(s/'preload.cjs'),NODE_DISABLE_COMPILE_CACHE='1',PYTHONDONTWRITEBYTECODE='1')
  with (root/(name+'.log')).open('w') as log:
   p=subprocess.Popen(['python3',repo/'stage3/lane/shard.py',name,results/side],cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
   try:code=p.wait(timeout=120)
   finally:
    try:os.killpg(p.pid,signal.SIGKILL)
    except ProcessLookupError:pass
    p.wait()
  report=json.loads((results/side/name/'report.json').read_text());observation[side]={'exit':code,'report':report}
 return unit,observation

def traces(directory,side,name):
 result=[];rows=[]
 for file in sorted((directory/side/'calls'/name).glob('*.jsonl')):
  events=[json.loads(l) for l in file.read_text().splitlines()];rows+=events
  normalized=[{k:v for k,v in row.items() if k not in ['sourceKind','sparse','sequence']} for row in events]
  result.append(normalized)
 result.sort(key=lambda trace:json.dumps(trace,sort_keys=True))
 return result,rows
for start in range(0,len(units),20):
 if start<len(summary['completedUnits']):continue
 begun=time.monotonic();piece=start//20+1;directory=s/f'piece-{piece:02d}';directory.mkdir();subset=units[start:start+20]
 for side in ['before','after']:(directory/side).mkdir()
 with ThreadPoolExecutor(max_workers=4) as pool:observations=[f.result() for f in as_completed([pool.submit(run,u,directory) for u in subset])]
 assert time.monotonic()-begun<1400
 assert all(frozen[side]==snapshot(side) for side in ['before','after'])
 packed=directory/'packed';packed.mkdir();stats=collections.Counter();oracle={side:collections.Counter() for side in ['before','after']};failures=[]
 with (packed/'events.jsonl').open('w') as eventsOut,(packed/'comparisons.jsonl').open('w') as compares,(packed/'shard-evidence.jsonl').open('w') as evidence:
  for unit,obs in sorted(observations,key=lambda row:row[0]['name']):
   name=unit['name'];sides={};allrows={}
   for side in ['before','after']:
    status=obs[side]
    if status['exit']!=0 or status['report']['status']!='pass':failures.append({'unit':name,'side':side,**status})
    oracle[side].update(status['report']['counts']);sides[side],allrows[side]=traces(directory,side,name)
    for file in sorted((results/side/name).rglob('*')):
     if file.is_file() and 'adapted-tree' not in file.relative_to(results/side/name).parts:evidence.write(json.dumps({'unit':name,'side':side,'path':str(file.relative_to(results/side/name)),'text':file.read_text(errors='backslashreplace')})+'\n')
    evidence.write(json.dumps({'unit':name,'side':side,'path':'shard-driver.log','text':(directory/side/(name+'.log')).read_text()})+'\n')
    eventsOut.write(json.dumps({'unit':name,'side':side,'traces':sides[side],'rawRows':allrows[side]})+'\n')
   equal=sides['before']==sides['after'];stats['valueDifferences']+=not equal
   returned=[x for x in allrows['before'] if x['event']=='values'];stores=[x for x in allrows['before'] if x['event']=='write']
   stats['beforeArrayCalls']+=sum(x['sourceKind']=='array' for x in returned);stats['afterArrayCalls']+=sum(x['sourceKind']=='array' for x in allrows['after'] if x['event']=='values');stats['calls']+=len(returned);stats['baselineSparseCalls']+=sum(x['sparse'] for x in returned);stats['returnedElements']+=sum(len(x['values']) for x in returned);stats['idWrites']+=len(stores)
   stats['nonIntegerIds']+=sum(not x['canonical'] for side in allrows.values() for x in side if x['event']=='write')
   stats['nonCanonicalKeys']+=sum(len(x['nonCanonicalKeys']) for side in allrows.values() for x in side if x['event']=='values')
   stats['identityProjectionFailures']+=sum(not x['identityProjectionMatches'] for side in allrows.values() for x in side if x['event']=='values')
   compares.write(json.dumps({'unit':name,'equalOrderedTraces':equal,'calls':len(returned),'baselineSparseCalls':sum(x['sparse'] for x in returned),'beforeEvents':len(allrows['before']),'afterEvents':len(allrows['after'])})+'\n')
   if not equal:failures.append({'unit':name,'reason':'ordered return/write trace differs'})
  for key in ['nonIntegerIds','nonCanonicalKeys','identityProjectionFailures','afterArrayCalls']:
   if stats[key]:failures.append({'reason':key,'count':stats[key]})
 record={'piece':piece,'units':[x['name'] for x in subset],'seconds':time.monotonic()-begun,'stats':dict(stats),'oracle':{side:dict(x) for side,x in oracle.items()},'failed':failures,'frozenInputsUnchanged':True}
 (packed/'piece.json').write_text(json.dumps(record,indent=2)+'\n')
 if piece==1:
  raw={p.name:p.read_text() for p in (s/'raw').iterdir() if p.is_file() and p.suffix in ['.json','.log']};(packed/'preparation.json').write_text(json.dumps(raw,indent=2)+'\n')
 if not failures:
  summary['completedUnits'] += record['units']
  for key,n in stats.items():summary[key]+=n
  for side in oracle:
   for key,n in oracle[side].items():summary['oracleCounts'][side][key]+=n
 summary['pieces'].append(record);summary['complete']=len(summary['completedUnits'])==len(units)
 (out/'RECORD-TABLE-COUNT.json').write_text(json.dumps(summary,indent=2)+'\n')
 (packed/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+p.name+'\n' for p in sorted(packed.iterdir()) if p.name!='SHA256SUMS'))
 archive=out/f'record-tables-measure-{piece:02d}.tar.gz'
 with tarfile.open(archive,'w:gz') as tf:
  for p in sorted(packed.iterdir()):tf.add(p,arcname=f'piece-{piece:02d}/'+p.name)
 (out/f'record-tables-measure-{piece:02d}.sha256').write_text(hashlib.sha256(archive.read_bytes()).hexdigest()+'  '+archive.name+'\n')
 git('add',str(out.relative_to(repo)));git('diff','--cached','--check');git('commit','-m',f'never cut alone: lands with chg_yj14ddxz, because the census reads its stop\n\ngetOwnValues road (b) measurement piece {piece}: paired {len(subset)} lane shards, {stats["calls"]} calls, {stats["baselineSparseCalls"]} previously sparse.');git('push','origin','typescript/getownvalues')
 print(json.dumps({'piece':piece,'completed':len(summary['completedUnits']),'calls':summary['calls'],'sparse':summary['baselineSparseCalls'],'differentUnits':stats['valueDifferences'],'failures':len(failures),'seconds':record['seconds']}),flush=True)
 for unit,_ in observations:
  for side in ['before','after']:shutil.rmtree(results/side/unit['name']/'adapted-tree',ignore_errors=True)
 if failures:raise SystemExit('Measurement failed; checkpoint pushed, no adaptation built.')
assert summary['calls']==1109 and summary['baselineSparseCalls']==445
assert all(v==plan['counts'] for v in summary['oracleCounts'].values())
assert len(summary['completedUnits'])==len(set(summary['completedUnits']))==398
print('COMPLETE',json.dumps(summary),flush=True)
