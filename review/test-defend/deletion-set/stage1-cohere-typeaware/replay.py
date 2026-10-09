import json,pathlib,subprocess,os,time,re,signal,shutil
p=pathlib.Path(__file__).resolve().parent
base_skip='^(TestVolumeProfileCorpora)($|_|/)'
known_semantic=re.compile(r'^(TestVolumeAgreementRepository_[0-9]+|TestVolumeProfileControls_[0-9]+)$')
records=json.loads((p/'matrix.json').read_text()) if (p/'matrix.json').exists() else []
def stop_tree(process):
 table=[l.split() for l in subprocess.check_output(['ps','-e','-o','pid=,ppid=']).decode().splitlines()]
 descendants={process.pid}
 while True:
  more={int(pid) for pid,parent in table if int(parent) in descendants}
  if more<=descendants:break
  descendants|=more
 for pid in reversed([int(x[0]) for x in table if int(x[0]) in descendants]):
  try:os.kill(pid,signal.SIGTERM)
  except ProcessLookupError:pass
 try:process.wait(timeout=5)
 except subprocess.TimeoutExpired:
  for pid in descendants:
   try:os.kill(pid,signal.SIGKILL)
   except ProcessLookupError:pass
  process.wait()
def observe(m,attempt,panics):
 label=m['branch'].split('/')[0]+'-'+m['mutant']+f'-{attempt}-enabled'
 skip=base_skip+('|'+'|'.join('^'+re.escape(n)+'($|/)' for n in panics) if panics else '')
 cmd=['go','test','-json','-count=1','-timeout','30m','./stage1/cohere/typeaware/','-skip',skip]
 cache='/tmp/deletion-typeaware/cache/'+label
 env=dict(os.environ,ADAMIC_GATE_UNCACHED='1',ADAMIC_BUILD_CACHE_DIR=cache)
 events=[];stopped=False;offset=0;pending='';semantic_rows=set();start=time.monotonic()
 log=p/(label+'.log')
 with log.open('w') as out:
  process=subprocess.Popen(cmd,stdout=out,stderr=subprocess.STDOUT,env=env,start_new_session=True)
  while True:
   with log.open() as inp:inp.seek(offset);data=inp.read();offset=inp.tell()
   pending+=data;lines=pending.split('\n');pending=lines.pop()
   for line in lines:
    try:e=json.loads(line)
    except ValueError:continue
    events.append(e)
    if re.search(r'content=agreement/',e.get('Output','')) and e.get('Test'):
     semantic_rows.add(e['Test'].split('/')[0])
    if e.get('Action')=='fail' and (known_semantic.fullmatch(e.get('Test','')) or e.get('Test') in semantic_rows):
     stopped=True
   if stopped:
    stop_tree(process);break
   if process.poll() is not None:break
   time.sleep(0.5)
 # Parse final flushed output again, including anything written during shutdown.
 events=[]
 for line in log.read_text().splitlines():
  try:events.append(json.loads(line))
  except ValueError:pass
 failed=sorted({e['Test'] for e in events if e.get('Action')=='fail' and e.get('Test') and '/' not in e['Test']})
 panic=[]
 for e in events:
  if e.get('Output','').lstrip().startswith('panic:'):
   panic.append(e.get('Test','').split('/')[0])
 r=dict(mutant=m['mutant'],branch=m['branch'],attempt=attempt,command=cmd,cache=cache,exit=process.returncode,wall=time.monotonic()-start,stopped_at_clean_failure=stopped,failed=failed,panicking_tests=sorted(set(panic)),passed=sorted({e['Test'] for e in events if e.get('Action')=='pass' and e.get('Test') and '/' not in e['Test']}),skipped=sorted({e['Test'] for e in events if e.get('Action')=='skip' and e.get('Test')}),binary_seconds=next((e.get('Elapsed') for e in reversed(events) if e.get('Action') in ['pass','fail'] and not e.get('Test')),None),known_non_witness_catchers=[n for n in failed if known_semantic.fullmatch(n) or n in semantic_rows],failing_lines=[e['Output'].strip() for e in events if '.go:' in e.get('Output','') and e.get('Test','').split('/')[0] in failed])
 r['environment']={k:env[k] for k in ['ADAMIC_GATE_UNCACHED','ADAMIC_BUILD_CACHE_DIR','ADAMIC_VOLUME_REPOSITORY_MANIFEST'] if k in env};records.append(r);(p/'matrix.json').write_text(json.dumps(records,indent=2)+'\n');print(label,r['wall'],r['failed'],r['panicking_tests'],flush=True)
 shutil.rmtree(cache,ignore_errors=True)
 return r
for m in json.loads((p/'mutant-list.json').read_text()):
 if m['stale'] or m['mutant'] in ['M1','M4']:continue
 done=subprocess.run(['git','apply',m['local_diff']],capture_output=True,text=True)
 if done.returncode:raise RuntimeError(done.stderr)
 try:
  panics=[]
  for attempt in range(1,20):
   r=observe(m,attempt,panics)
   if r['panicking_tests']:
    if '' in r['panicking_tests']:raise RuntimeError('Cannot identify panicking top-level test; see '+str(r))
    panics=sorted(set(panics+r['panicking_tests']));continue
   break
 finally:
  subprocess.run(['git','apply','-R',m['local_diff']],check=True)
