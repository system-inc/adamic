import concurrent.futures,datetime,gzip,json,os,pathlib,queue,re,shutil,subprocess,threading,time
P=pathlib.Path(__file__).resolve().parent;ROOT=P.parent
ITEMS=json.loads((P/'inventory.json').read_text());SKIP=(P/'skip.regex').read_text().strip()
EXTRA=set(json.loads((P/'skipped-extra.json').read_text()));LOCK=threading.Lock();Q=queue.Queue()
RESULTS={r['diff']:r for r in json.loads((P/'results.json').read_text())} if (P/'results.json').exists() else {}
for item in ITEMS:
 if item['diff'] not in RESULTS:Q.put(item)
(P/'logs').mkdir(exist_ok=True)
def save():
 t=P/'results.json.tmp';t.write_text(json.dumps([RESULTS[i['diff']] for i in ITEMS if i['diff'] in RESULTS],indent=2)+'\n');t.replace(P/'results.json')
def cleanup(path):
 if not path.exists():return
 for root,dirs,files in os.walk(path):
  for name in dirs:
   q=os.path.join(root,name)
   if not os.path.islink(q):
    try:os.chmod(q,0o700)
    except OSError:pass
 shutil.rmtree(path,ignore_errors=True)
def parse(log):
 events=[]
 for line in log.read_text().splitlines():
  try:events.append(json.loads(line))
  except ValueError:pass
 fail={e['Test'].split('/')[0] for e in events if e.get('Action')=='fail' and e.get('Test')}
 output=[(e.get('Test','').split('/')[0],e.get('Output','')) for e in events if e.get('Action')=='output'];panics=[]
 for n,(top,s) in enumerate(output):
  if not s.startswith(('panic:','fatal error:')):continue
  trace=''.join(x[1] for x in output[n:]);tests=re.findall(r'github\.com/system-inc/adamic/internal/oracle\.(Test[A-Za-z0-9_]+)',trace);name=tests[0] if tests else top
  frame='unknown';lines=trace.splitlines()
  for j,line in enumerate(lines):
   if line.startswith('github.com/system-inc/adamic/') and ('/internal/' in line or '/stage1/' in line):
    fn=line.rsplit('(',1)[0];loc=lines[j+1].strip().split(' +')[0] if j+1<len(lines) else ''
    loc=re.sub(r'^.*?/w[01]/','',loc);frame=fn+' at '+loc;break
  record={'test':name or '<unknown>','frame':frame}
  if record not in panics:panics.append(record)
 ended=not any(e.get('FailedBuild') for e in events) and any(e.get('Action') in ('pass','fail','skip') and not e.get('Test') for e in events)
 return events,fail,panics,ended

def worker(n):
 w=pathlib.Path('/workspace/adamic')
 while True:
  try:item=Q.get_nowait()
  except queue.Empty:return
  name=item['diff'];diff=P/'diffs'/(name+'.diff');scratch=ROOT/'applied'/name;result={**item,'all_catchers':[],'panics':[],'completed':False,'runs':[]};applied=False
  start=time.monotonic();gocache=ROOT/'go-cache'/('w'+str(n))
  try:
   cleanup(gocache);shutil.copytree(ROOT/'seed-cache',gocache,copy_function=os.link)
   subprocess.run(['git','apply','--check',str(diff)],cwd=w,check=True,capture_output=True)
   
   files=[line.split('\t')[-1] for line in subprocess.check_output(['git','apply','--numstat',str(diff)],cwd=w,text=True).splitlines()]
   for file in files:
    target=scratch/file;target.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(w/file,target)
   subprocess.run(['git','apply','--unsafe-paths','--directory='+str(scratch),str(diff)],cwd=w,check=True);applied=True
   overlay=ROOT/'overlays'/(name+'.json');overlay.parent.mkdir(exist_ok=True);overlay.write_text(json.dumps({'Replace':{str(w/file):str(scratch/file) for file in files}},indent=2)+'\n')
   excluded=[];caught=set()
   for round in range(199):
    skip=SKIP+('|'+'^('+ '|'.join(re.escape(x) for x in excluded)+')($|/)' if excluded else '')
    key=name+'-r'+str(round);cache=ROOT/'cache'/key;tmp=ROOT/'tmp'/key;tmp.mkdir(parents=True,exist_ok=True);os.chmod(tmp,0o1777)
    env=os.environ.copy();env.update(GOMAXPROCS='2',GOFLAGS='-p=2',GOCACHE=str(gocache),ADAMIC_GATE_UNCACHED='1',ADAMIC_BUILD_CACHE_DIR=str(cache),TMPDIR=str(tmp),GOTMPDIR=str(tmp))
    cmd=['go','test','-overlay='+str(overlay),'-count=1','-timeout','60m','-json','./internal/oracle/','-skip',skip];log=P/'logs'/(key+'.log');begin=time.monotonic()
    print(json.dumps({'start':name,'round':round,'worker':n,'panic_skips':excluded}),flush=True)
    with log.open('w') as out:exitcode=subprocess.run(cmd,cwd=w,env=env,stdout=out,stderr=subprocess.STDOUT).returncode
    events,failed,panics,ended=parse(log);panic_names={x['test'] for x in panics};caught.update(failed-panic_names-EXTRA)
    run={'command':cmd,'environment':{k:env[k] for k in ('GOMAXPROCS','GOFLAGS','GOCACHE','ADAMIC_GATE_UNCACHED','ADAMIC_BUILD_CACHE_DIR','TMPDIR','GOTMPDIR')},'log':'logs/'+log.name+'.gz','exit':exitcode,'wall_seconds':round_time(time.monotonic()-begin),'failed_top_level':sorted(failed),'panics':panics,'package_completed':ended}
    result['runs'].append(run);result['all_catchers']=sorted(caught)
    for panic in panics:
     if panic not in result['panics']:result['panics'].append(panic)
    # Save every completed attempt, including a panic retry, before another run starts.
    with LOCK:(P/(name+'-progress.json')).write_text(json.dumps(result,indent=2)+'\n')
    with log.open('rb') as src,gzip.open(log.with_name(log.name+'.gz'),'wb',compresslevel=6) as dst:shutil.copyfileobj(src,dst)
    log.unlink();cleanup(cache);cleanup(tmp)
    if panics:
     new=[p['test'] for p in panics if p['test'] not in excluded]
     if not new or '<unknown>' in new:result['broken']='Unattributed or repeated test-binary panic';break
     excluded.extend(new);continue
    result['completed']=ended and (exitcode in (0,1))
    if not result['completed']:result['broken']='Build/process failure without a completed package result'
    break
  except Exception as e:result['broken']=repr(e)
  finally:
   if applied:cleanup(scratch)
   cleanup(gocache)
  result['wall_seconds']=round_time(time.monotonic()-start)
  with LOCK:
   RESULTS[name]=result;save();print(json.dumps({'done':len(RESULTS),'diff':name,'catchers':result['all_catchers'],'panics':result['panics'],'completed':result['completed'],'seconds':result['wall_seconds'],'broken':result.get('broken')}),flush=True)
  Q.task_done()
def round_time(x):return float(f'{x:.3f}')
start=time.monotonic()
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:list(pool.map(worker,range(2)))
(P/'status.json').write_text(json.dumps({'wall_seconds':round_time(time.monotonic()-start),'completed':len(RESULTS),'ended_utc':datetime.datetime.now(datetime.timezone.utc).isoformat()},indent=2)+'\n')
