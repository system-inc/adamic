"""Measure helper entries per stock-checker source call site across the frozen lane."""
from pathlib import Path
import json,os,sys,time,subprocess,signal,collections,tarfile,hashlib,shutil
from concurrent.futures import ThreadPoolExecutor,as_completed
repo=Path('/workspace/adamic');s=Path('/workspace/scratch/getfunctionname');out=repo/'stage3/census/getfunctionname';plan=json.loads((repo/'stage3/lane/shard-plan.json').read_text());sites=[f"{r['file']}:{r['line']}:{r['column']}" for r in json.loads((s/'raw/references.json').read_text())['callSites']];results=s/'lane-results';results.mkdir(exist_ok=True);evidence=s/'evidence';evidence.mkdir(exist_ok=True)
summaryPath=evidence/'runtime-counts.json';summary={'main':'3214dd465e1155d7bf4d41653e95b29a3f1060a6','planSha256':hashlib.sha256((repo/'stage3/lane/shard-plan.json').read_bytes()).hexdigest(),'units':[],'pieces':[],'counts':dict.fromkeys(sites,0),'attempts':dict.fromkeys(sites,0),'unknownEntries':{},'totalEntries':0,'nodeProcesses':0,'oracle':{'pass':0,'fail':0,'skip':0},'complete':False}
if summaryPath.exists():summary=json.loads(summaryPath.read_text())
def run(unit,piece):
 name=unit['name'];directory=piece/'counts'/name;directory.mkdir(parents=True);log=piece/(name+'.log');env=dict(os.environ,STAGE3_ADAPTED_TREE=str(s/'measure'),STAGE3_SHARD_SECONDS='90',NODE_OPTIONS='--require='+str(out/'preload.cjs'),NODE_DISABLE_COMPILE_CACHE='1',GFN_COUNTS_DIR=str(directory),PYTHONDONTWRITEBYTECODE='1');begin=time.monotonic()
 with log.open('w') as f:
  child=subprocess.Popen(['python3',str(repo/'stage3/lane/shard.py'),name,str(results)],cwd=repo,env=env,stdout=f,stderr=subprocess.STDOUT,start_new_session=True)
  try:
   try:code=child.wait(timeout=120)
   except subprocess.TimeoutExpired:os.killpg(child.pid,signal.SIGKILL);child.wait();code=child.returncode
  finally:
   if child.poll() is None:os.killpg(child.pid,signal.SIGKILL);child.wait()
 report=results/name/'report.json';report=json.loads(report.read_text()) if report.exists() else {'status':'missing','errors':['no report']};processes=[json.loads(p.read_text()) for p in sorted(directory.glob('*.json'))];raw=[]
 for p in sorted((results/name).rglob('*')):
  if p.is_file() and 'adapted-tree' not in p.relative_to(results/name).parts:raw.append({'path':str(p.relative_to(results/name)),'text':p.read_text(errors='backslashreplace')})
 raw.append({'path':'driver.log','text':log.read_text()});return {'unit':name,'childExit':code,'seconds':time.monotonic()-begin,'report':report,'processes':processes,'raw':raw}
batch=int(sys.argv[1]) if len(sys.argv)>1 else 8;done=0
for offset in range(0,len(plan['units']),20):
 if offset<len(summary['units']):continue
 begin=time.monotonic();number=offset//20+1;piece=s/f'piece-{number:02d}';piece.mkdir();subset=plan['units'][offset:offset+20]
 with ThreadPoolExecutor(max_workers=4) as pool:rows=[f.result() for f in as_completed([pool.submit(run,u,piece) for u in subset])]
 rows.sort(key=lambda r:r['unit']);failures=[];counts=collections.Counter(dict.fromkeys(sites,0));attempts=collections.Counter(dict.fromkeys(sites,0));unknown={};entries=0;processCount=0;oracle=collections.Counter()
 for row in rows:
  if row['childExit']!=0 or row['report'].get('status')!='pass' or not row['processes']:failures.append({'unit':row['unit'],'exit':row['childExit'],'report':row['report']})
  if row['report'].get('counts'):oracle.update(row['report']['counts'])
  for process in row['processes']:
   assert set(process['counts'])==set(sites) and set(process['attempts'])==set(sites);assert process['entries']==sum(process['counts'].values())+sum(r['count'] for r in process['unknown'].values());counts.update(process['counts']);attempts.update(process['attempts']);entries+=process['entries'];processCount+=1
   for key,event in process['unknown'].items():
    if key not in unknown:unknown[key]={**event,'count':0}
    unknown[key]['count']+=event['count']
 packed=piece/'packed';packed.mkdir()
 with (packed/'observations.jsonl').open('w') as f:
  for row in rows:f.write(json.dumps(row)+'\n')
 record={'piece':number,'units':[u['name'] for u in subset],'seconds':time.monotonic()-begin,'counts':dict(counts),'attempts':dict(attempts),'unknownEntries':unknown,'totalEntries':entries,'nodeProcesses':processCount,'oracle':dict(oracle),'failures':failures};assert record['seconds']<1490
 (packed/'piece.json').write_text(json.dumps(record,indent=2)+'\n');(packed/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+p.name+'\n' for p in sorted(packed.iterdir()) if p.name!='SHA256SUMS'))
 archive=evidence/f'lane-{number:02d}.tar.gz'
 with tarfile.open(archive,'w:gz') as tf:
  for p in sorted(packed.iterdir()):tf.add(p,arcname=p.name)
 assert archive.stat().st_size<20_000_000
 summary['pieces'].append(dict(record,archive=archive.name,sha256=hashlib.sha256(archive.read_bytes()).hexdigest()))
 if not failures:
  summary['units']+=record['units'];summary['totalEntries']+=entries;summary['nodeProcesses']+=processCount
  for key in sites:summary['counts'][key]+=counts[key];summary['attempts'][key]+=attempts[key]
  for key,n in oracle.items():summary['oracle'][key]+=n
  for key,event in unknown.items():
   if key not in summary['unknownEntries']:summary['unknownEntries'][key]={**event,'count':0}
   summary['unknownEntries'][key]['count']+=event['count']
 summary['complete']=len(summary['units'])==len(plan['units']);summaryPath.write_text(json.dumps(summary,indent=2)+'\n');print(json.dumps({'piece':number,'acceptedUnits':len(summary['units']),'counts':summary['counts'],'unattributedEntries':sum(x['count'] for x in summary['unknownEntries'].values()),'seconds':record['seconds'],'failures':len(failures)}),flush=True)
 for row in rows:shutil.rmtree(results/row['unit']/'adapted-tree',ignore_errors=True)
 if failures:raise SystemExit(1)
 done+=1
 if done>=batch:break
if summary['complete']:
 assert summary['units']==[u['name'] for u in plan['units']];assert summary['oracle']==plan['counts'];print('WHOLE ORACLE COMPLETE',json.dumps(summary['counts']),flush=True)
else:print('BATCH COMPLETE',len(summary['units']),flush=True)
