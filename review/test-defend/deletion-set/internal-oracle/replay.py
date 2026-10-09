import concurrent.futures,json,os,pathlib,queue,re,signal,subprocess,threading,time
ROOT=pathlib.Path('/workspace/adamic')
OUT=ROOT/'review/test-defend/deletion-set/internal-oracle'
ITEMS=json.loads((OUT/'mutant-list.json').read_text())
WITNESSES=set(json.loads((OUT/'witnesses.json').read_text()))
SKIP=(OUT/'skip.regex').read_text().strip()
LOCK=threading.Lock(); RESULTS={r['replay_index']:r for r in json.loads((OUT/'matrix.json').read_text())}; QUEUE=queue.Queue()
for i,item in enumerate(ITEMS):
 if i not in RESULTS:QUEUE.put((i,item))
(OUT/'logs').mkdir(exist_ok=True)

def save():
 tmp=OUT/'matrix.json.tmp';tmp.write_text(json.dumps([RESULTS[i] for i in sorted(RESULTS)],indent=2)+'\n');tmp.replace(OUT/'matrix.json')

def descendants(pid):
 result=set();front={pid}
 while front:
  nexts=set()
  for p in pathlib.Path('/proc').glob('[0-9]*'):
   try:
    stat=(p/'stat').read_text();parent=int(stat[stat.rfind(')')+2:].split()[1])
    if parent in front:nexts.add(int(p.name))
   except (OSError,ValueError,IndexError):pass
  result.update(nexts);front=nexts
 return result

def stop(proc):
 children=descendants(proc.pid)
 try:os.killpg(proc.pid,signal.SIGTERM)
 except ProcessLookupError:pass
 for p in children:
  try:os.kill(p,signal.SIGTERM)
  except ProcessLookupError:pass
 try:proc.wait(timeout=3)
 except subprocess.TimeoutExpired:
  try:os.killpg(proc.pid,signal.SIGKILL)
  except ProcessLookupError:pass
  proc.wait()
 for p in children:
  try:os.kill(p,signal.SIGKILL)
  except ProcessLookupError:pass

def run(w,index,item,round,extra):
 key=f'{index:03d}-{item["mutant"]}-r{round}'
 log=OUT/'logs'/f'{key}.log'
 skip=SKIP if not extra else SKIP+'|^('+ '|'.join(re.escape(x) for x in extra)+')($|/)' 
 cmd=['go','test','-json','-count=1','-timeout','30m','./internal/oracle/','-skip',skip]
 cache=f'/workspace/deletion-replay/cache/{key}-resumed-{os.getpid()}'
 env=os.environ.copy();env.update(ADAMIC_GATE_UNCACHED='1',ADAMIC_BUILD_CACHE_DIR=cache, GOTMPDIR='/workspace/deletion-replay/tmp', TMPDIR='/workspace/deletion-replay/tmp')
 start=time.monotonic();failed=set();wfailed=set();panics=set();passed=set();evidence=[];active='';early=False;package_completed=False;pending=''
 with log.open('w') as target, log.open('r') as reader:
  proc=subprocess.Popen(cmd,cwd=w,env=env,stdout=target,stderr=subprocess.STDOUT,start_new_session=True)
  while True:
   chunk=reader.read()
   pending+=chunk
   lines=pending.split('\n');pending=lines.pop()
   for line in lines:
    try:event=json.loads(line)
    except json.JSONDecodeError:continue
    test=event.get('Test','');top=test.split('/')[0];action=event.get('Action');output=event.get('Output','')
    if action=='run' and top:active=top
    if output.startswith('panic:') or output.startswith('fatal error:'):
     panics.add(top or active or '<unknown>');evidence.append(output.strip())
    if action=='fail' and top:
     if top in WITNESSES:wfailed.add(top)
     else:failed.add(top)
     evidence.append(f'{test}: fail')
    if action=='pass' and top and test==top:passed.add(top)
    if action in ('pass','fail','skip') and not test:package_completed=True
   if failed and not panics and proc.poll() is None:
    early=True;stop(proc)
   if proc.poll() is not None:
    # read final buffered JSON in one additional iteration
    if not chunk:break
   time.sleep(.15)
 exitcode=proc.wait()
 result={'command':cmd,'environment':{'ADAMIC_GATE_UNCACHED':'1','ADAMIC_BUILD_CACHE_DIR':cache,'GOTMPDIR':env['GOTMPDIR'],'TMPDIR':env['TMPDIR']},'log':str(log.relative_to(OUT)),'wall_seconds':round_time(time.monotonic()-start),'exit':exitcode,'stopped_early':early,'package_completed':package_completed,'still_caught_by':sorted(failed-panics),'witness_failures':sorted(wfailed),'panicking_tests':sorted(panics),'observed_passes':sorted(passed),'evidence':evidence}
 if not early and not package_completed:result['broken']='test binary did not complete with a package result'
 if exitcode!=0 and not failed and not wfailed and not panics:result['broken']='build/process failure without a failing test'
 # Each cache is task scratch and can be removed after evidence is captured.
 import shutil
 shutil.rmtree(cache,ignore_errors=True)
 return result

def round_time(value):return float(f'{value:.3f}')

def worker(n):
 w=pathlib.Path(f'/workspace/deletion-replay/w{n}')
 while True:
  try:i,item=QUEUE.get_nowait()
  except queue.Empty:return
  result=dict(item);result['replay_index']=i;result['runs']=[]
  diff=OUT/item['saved_diff'];applied=False
  try:
   check=subprocess.run(['git','apply','--check',str(diff)],cwd=w,capture_output=True,text=True)
   if check.returncode:result.update(stale=True,apply_check=check.stderr)
   else:
    subprocess.run(['git','apply',str(diff)],cwd=w,check=True);applied=True
    extra=[]
    for r in range(199):
     attempt=run(w,i,item,r+int(os.getenv("REPLAY_ROUND_OFFSET","0")),extra);result['runs'].append(attempt)
     if attempt['panicking_tests']:
      for x in attempt['panicking_tests']:
       if x not in extra:extra.append(x)
      if '<unknown>' in extra:result['broken']='unattributed test-binary panic';break
      continue
     break
    last=result['runs'][-1]
    result['still_caught_by']=last['still_caught_by']
    result['panicking_tests']=sorted({x for r in result['runs'] for x in r['panicking_tests']})
    result['witness_failures']=sorted({x for r in result['runs'] for x in r['witness_failures']})
    result['wall_seconds']=round_time(sum(r['wall_seconds'] for r in result['runs']))
    result['completed_without_catcher']=last['package_completed'] and not last['still_caught_by'] and not last.get('broken')
    if last.get('broken'):result['broken']=last['broken']
  except Exception as e:result['broken']=repr(e)
  finally:
   if applied:subprocess.run(['git','apply','-R',str(diff)],cwd=w,check=True)
  with LOCK:
   RESULTS[i]=result;save()
   print(json.dumps({'done':len(RESULTS),'total':len(ITEMS),'index':i,'mutant':item['mutant'],'caught':result.get('still_caught_by',[]),'seconds':result.get('wall_seconds'),'broken':result.get('broken'),'lost':result.get('completed_without_catcher',False)}),flush=True)
  QUEUE.task_done()

start=time.monotonic()
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:list(pool.map(worker,range(2)))
(OUT/'replay-status.json').write_text(json.dumps({'wall_seconds':round_time(time.monotonic()-start),'completed':len(RESULTS)},indent=2)+'\n')
