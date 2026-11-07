"""Run every frozen lane manifest unit with ordinary guards and message counters."""
from pathlib import Path
import json,os,subprocess,signal,time,sys,hashlib,tarfile,collections,shutil
from concurrent.futures import ThreadPoolExecutor,as_completed
repo=Path('/workspace/adamic');s=Path('/workspace/scratch/debug-messages');unit=repo/'stage3/adapt/99-correctness-core-debug-messages';side=sys.argv[1];tree=s/('measure-'+side);plan=json.loads((repo/'stage3/lane/shard-plan.json').read_text());results=s/('lane-'+side);results.mkdir(exist_ok=True);evidence=s/'evidence';evidence.mkdir(exist_ok=True)
names=['assertEachNode','assertNode','assertNotNode','assertOptionalNode'];summary={'side':side,'plan_sha256':hashlib.sha256((repo/'stage3/lane/shard-plan.json').read_bytes()).hexdigest(),'units':[],'counts':dict.fromkeys(names,0),'oracle':{'pass':0,'fail':0,'skip':0},'pieces':[],'complete':False}
path=evidence/(side+'-lane.json');batchLimit=int(sys.argv[2]) if len(sys.argv)>2 else 1000;batchPieces=0
if path.exists():summary=json.loads(path.read_text())
def run(item,piece):
 name=item['name'];counts=piece/'counts'/name;counts.mkdir(parents=True);log=piece/(name+'.log')
 env=dict(os.environ,STAGE3_ADAPTED_TREE=str(tree),STAGE3_SHARD_SECONDS='90',NODE_OPTIONS='--require='+str(unit/'preload.cjs'),NODE_DISABLE_COMPILE_CACHE='1',DEBUG_MESSAGES_COUNTS=str(counts),PYTHONDONTWRITEBYTECODE='1')
 start=time.monotonic()
 with log.open('w') as f:
  p=subprocess.Popen(['python3',str(repo/'stage3/lane/shard.py'),name,str(results)],cwd=repo,env=env,stdout=f,stderr=subprocess.STDOUT,start_new_session=True)
  try:
   try:code=p.wait(timeout=120)
   except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);p.wait();code=p.returncode
  finally:
   try:os.killpg(p.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   p.wait()
 reportPath=results/name/'report.json';report=json.loads(reportPath.read_text()) if reportPath.exists() else {'status':'missing','errors':['no report']}
 processes=[json.loads(p.read_text()) for p in sorted(counts.glob('*.json'))];total=dict.fromkeys(names,0)
 for row in processes:
  for key in names:total[key]+=row['counts'][key]
 record={'unit':name,'child_exit':code,'seconds':time.monotonic()-start,'report':report,'processes':processes,'counts':total}
 # Keep ordinary identity, API, baseline and test-count guards, not only the summary.
 raw=[]
 for p in sorted((results/name).rglob('*')):
  if p.is_file() and 'adapted-tree' not in p.relative_to(results/name).parts:raw.append({'path':str(p.relative_to(results/name)),'text':p.read_text(errors='backslashreplace')})
 raw.append({'path':'driver.log','text':log.read_text()});record['raw']=raw
 return record
for offset in range(0,len(plan['units']),20):
 if offset<len(summary['units']):continue
 start=time.monotonic();number=offset//20+1;piece=s/f'lane-{side}-piece-{number:02d}';piece.mkdir();subset=plan['units'][offset:offset+20]
 with ThreadPoolExecutor(max_workers=4) as pool:records=[f.result() for f in as_completed([pool.submit(run,item,piece) for item in subset])]
 records.sort(key=lambda r:r['unit']);failures=[]
 for r in records:
  if r['child_exit']!=0 or r['report'].get('status')!='pass' or not r['processes'] or any(r['counts'].values()):failures.append({'unit':r['unit'],'exit':r['child_exit'],'status':r['report'].get('status'),'counts':r['counts'],'errors':r['report'].get('errors')})
 packed=piece/'packed';packed.mkdir()
 with (packed/'observations.jsonl').open('w') as f:
  for record in records:f.write(json.dumps(record)+'\n')
 info={'piece':number,'side':side,'units':[i['name'] for i in subset],'seconds':time.monotonic()-start,'failures':failures}
 (packed/'piece.json').write_text(json.dumps(info,indent=2)+'\n')
 (packed/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+p.name+'\n' for p in sorted(packed.iterdir()) if p.name!='SHA256SUMS'))
 archive=evidence/f'{side}-lane-{number:02d}.tar.gz'
 with tarfile.open(archive,'w:gz') as tf:
  for p in sorted(packed.iterdir()):tf.add(p,arcname=p.name)
 assert archive.stat().st_size<20_000_000
 summary['pieces'].append(dict(info,archive=archive.name,sha256=hashlib.sha256(archive.read_bytes()).hexdigest()))
 if not failures:
  summary['units'] += [i['name'] for i in subset]
  for r in records:
   for key in names:summary['counts'][key]+=r['counts'][key]
   for key,value in r['report']['counts'].items():summary['oracle'][key]+=value
 summary['complete']=len(summary['units'])==len(plan['units']);path.write_text(json.dumps(summary,indent=2)+'\n')
 print(json.dumps({'side':side,'piece':number,'completed':len(summary['units']),'seconds':info['seconds'],'counts':summary['counts'],'failures':failures}),flush=True)
 for r in records:shutil.rmtree(results/r['unit']/'adapted-tree',ignore_errors=True)
 if failures:raise SystemExit(1)
 batchPieces+=1
 if batchPieces>=batchLimit:break
if not summary['complete']:
 print('BATCH COMPLETE',side,len(summary['units']),flush=True);raise SystemExit(0)
assert summary['units']==[u['name'] for u in plan['units']]
assert summary['oracle']==plan['counts'];assert not any(summary['counts'].values());print('COMPLETE',side,summary['oracle'],flush=True)
