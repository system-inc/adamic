import os,sys,time,json,subprocess,shutil,hashlib,signal,datetime
from pathlib import Path
R=Path('/tmp/family5-land');W=Path('/tmp/family5-alternating');SHAS={'main':'fbe5b72ba115a8f3060d4b5229fd578574f4505f','c02':'28a2e22015d0cf51667fcefa7acb246cab553817'}
os.chdir(R);os.environ.update(ADAMIC_TYPESCRIPT_SOURCE='/workspace/mapset-upstream',ADAMIC_BUILD_CACHE_DIR='/workspace/family5-build-cache')
initial=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip();baseline=json.loads((W/'artifact-manifest.json').read_text());expected={side:{x['file']:x[side+'_sha256'] for x in baseline if x[side+'_present']} for side in SHAS}
results=[]
def stamp():return datetime.datetime.now(datetime.timezone.utc).isoformat()
def snapshot_artifacts(tree,side,dest):
 rows=[]
 for root in ['built/local','lib']:
  for p in sorted((tree/root).rglob('*')):
   if p.is_file():
    key=str(p.relative_to(tree));data=p.read_bytes();h=hashlib.sha256(data).hexdigest();rows.append({'file':key,'bytes':len(data),'sha256':h})
 (dest/'artifacts.json').write_text(json.dumps(rows,indent=2)+'\n')
 jsdiff=[x for x in rows if x['file'].endswith(('.js','.cjs','.mjs')) and expected[side].get(x['file'])!=x['sha256']]
 (dest/'artifact-validation.json').write_text(json.dumps({'javascript_count':sum(x['file'].endswith(('.js','.cjs','.mjs')) for x in rows),'javascript_differences_from_initial_snapshot':jsdiff},indent=2)+'\n')
 if jsdiff:raise RuntimeError('New executable artifact drift: '+json.dumps(jsdiff))

def one(index,side):
 sha=SHAS[side];out=Path('/workspace/family5-alt-run-'+str(index)+'-'+side);dest=W/('run-'+str(index)+'-'+side)
 dest.mkdir(parents=True,exist_ok=False);assert not out.exists()
 subprocess.run(['git','switch','--detach',sha],check=True)
 subprocess.run(['ps','-eo','pid,etime,rss,args','--sort=-rss'],stdout=(dest/'idle-before.txt').open('w'),check=True)
 cmd=['bash','stage3/lane/run.sh',str(out)];start=time.monotonic();started=stamp();print('START alternating '+str(index)+' '+side,flush=True)
 (W/'current.json').write_text(json.dumps({'run':index,'side':side,'sha':sha,'started_utc':started,'result_directory':str(out)},indent=2)+'\n')
 input_recorded=False
 with (dest/'stdout.log').open('wb') as stdout,(dest/'stderr.log').open('wb') as stderr:
  child=subprocess.Popen(cmd,stdout=stdout,stderr=stderr,start_new_session=True)
  while True:
   try:code=child.wait(timeout=5);break
   except subprocess.TimeoutExpired:
    # Small scheduling-input snapshot while install/build is active, before timed cases start.
    if not input_recorded and (out/'oracle/install.log').exists():
     inp={}
     for name in ['.parallelperf.json','mytest.config']:
      p=out/'adapted-tree'/name;inp[name]={'present':p.exists()}
      if p.exists():shutil.copyfile(p,dest/('before-'+name.lstrip('.')));inp[name]['sha256']=hashlib.sha256(p.read_bytes()).hexdigest()
     (dest/'scheduling-inputs.json').write_text(json.dumps(inp,indent=2)+'\n');input_recorded=True
    elapsed=time.monotonic()-start
    if int(elapsed)//50>int(elapsed-5)//50:
     print(f'RUNNING alternating {index} {side}: {elapsed:.0f}s',flush=True)
  try:os.killpg(child.pid,signal.SIGTERM)
  except ProcessLookupError:pass
 status={'side':side,'sha':sha,'command':cmd,'cwd':str(R),'started_utc':started,'ended_utc':stamp(),'exit_code':code,'wall_seconds':round(time.monotonic()-start,3),'environment':{k:v for k,v in os.environ.items() if k.startswith(('ADAMIC_','LATENT_','CENSUS_'))}}
 (dest/'status.json').write_text(json.dumps(status,indent=2)+'\n')
 if not (out/'report.json').exists():raise RuntimeError('Lane has no complete report; real child exit '+str(code))
 lane=json.loads((out/'report.json').read_text());oracle=json.loads((out/'oracle/report.json').read_text()) if (out/'oracle/report.json').exists() else {}
 row={'id':'new-'+str(index),'side':side,'sha':sha,'context':'alone, standard eight workers','counts':lane.get('counts'),'exit_code':code,'wall_seconds':status['wall_seconds'],'phases':oracle.get('phases'),'failures':[{k:f.get(k) for k in ['title','cause','elapsed_ms','timeout_ms','message']} for f in oracle.get('failures',[])],'execution':lane.get('execution')}
 (dest/'result.json').write_text(json.dumps(row,indent=2)+'\n');results.append(row);(W/'new-runs.json').write_text(json.dumps(results,indent=2)+'\n')
 print('RESULT '+json.dumps(row),flush=True)
 for p in out.rglob('*'):
  if not p.is_file():continue
  rel=p.relative_to(out)
  if rel.parts[0]=='adapted-tree' or 'node_modules' in rel.parts:continue
  if p.suffix not in ['.json','.jsonl','.log','.md','.txt','.diff']:continue
  to=dest/'lane'/rel;to.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(p,to)
 snapshot_artifacts(out/'adapted-tree',side,dest)
 # Retain all receipts and manifests; discard only this run's generated, reproducible tree to keep disk available.
 shutil.rmtree(out/'adapted-tree')
 if lane.get('execution',{}).get('apply_exit')!=0 or any(oracle.get('phases',{}).get(k,{}).get('exit')!=0 for k in ['install','build']):raise RuntimeError('Infrastructure phase failed in '+side+' run '+str(index))
 if oracle.get('phases',{}).get('tests',{}).get('exit')==124:raise RuntimeError('Standard oracle 2400-second test limit reached; partial output retained')
try:
 plan=json.loads((W/'plan.json').read_text());sequence=plan['sequence']
 for index,side in enumerate(sequence,1):one(index,side)
 if json.loads((W/'plan.json').read_text()).get('extra_main'):one(len(sequence)+1,'main')
 (W/'finished.json').write_text(json.dumps({'complete':True,'runs':len(results),'pushed':False},indent=2)+'\n')
except Exception as error:
 import traceback
 (W/'STOP.txt').write_text(type(error).__name__+': '+str(error)+'\n'+traceback.format_exc());print('STOP '+str(error),flush=True)
finally:
 subprocess.run(['git','switch','--detach',initial],check=True)
 (W/'current.json').write_text(json.dumps({'finished':True,'runs_complete':len(results),'pushed':False},indent=2)+'\n')
