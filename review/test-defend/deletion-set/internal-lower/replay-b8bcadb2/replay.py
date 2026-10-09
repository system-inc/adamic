import subprocess,json,pathlib,time,os,signal,re
P=pathlib.Path('review/test-defend/deletion-set/internal-lower/replay-b8bcadb2');mutants=json.load(open(P/'mutant-list.json'));candidates=json.load(open(P/'candidates.json'));results=json.load(open(P/'matrix.json')) if (P/'matrix.json').exists() else []
# These named agreement witnesses intentionally plant failures in their own comparator.
witnesses={'TestAgreementRejectsWrongLoweredOutput','TestAgreementRejectsEmptyAnswer','TestNativeAgreementRejectsWrongOutput','TestNativeAgreementRejectsWrongExit'}
def excluded(name):return name in witnesses or bool(re.search('Mutants?$|MutantKilled$|Planted(Failure|Disagreement)$',name))
for m in mutants:
 if any(x['key']==m['key'] for x in results):continue
 result=dict(m,still_caught_by=[],witness_failures=[],pin_failures=[],panicking_tests=[],runs=[])
 if m['stale']:results.append(result);continue
 subprocess.run(['git','apply',m['file']],check=True)
 try:
  extra=[]
  for attempt in range(20):
   skip='^('+'|'.join(candidates+extra)+')($|_|/)';cmd=['go','test','-json','-count=1','-timeout','30m','./internal/lower/','-skip',skip]
   log=P/(m['key']+'-'+str(attempt)+'.log');env=dict(os.environ,ADAMIC_GATE_UNCACHED='1',ADAMIC_BUILD_CACHE_DIR='/tmp/lower-deletion/cache/'+m['key']+'/'+str(attempt));start=time.monotonic();offset=0;buffer='';failed=set();pins=set();wit=set();panic=None;stopped=False;lasttest=None;complete=False
   with open(log,'w') as out:
    proc=subprocess.Popen(cmd,env=env,stdout=out,stderr=subprocess.STDOUT,start_new_session=True)
    while True:
     with open(log) as inp:
      inp.seek(offset);chunk=inp.read();offset=inp.tell()
     buffer+=chunk;lines=buffer.split('\n');buffer=lines.pop()
     for line in lines:
      try:e=json.loads(line)
      except ValueError:continue
      test=e.get('Test','').split('/')[0];output=e.get('Output','')
      if test:lasttest=test
      if e.get('Action')=='fail' and test:
       if excluded(test):wit.add(test)
       elif test in pins:pass
       else:failed.add(test)
      if e.get('OutputType')=='error' and test and re.search(r'(scan inputs changed|input.*hash.*changed|census.*hash|fingerprint.*changed)',output,re.I):pins.add(test)
      if output.startswith('panic: ') or output.startswith('fatal error: '):panic=test or lasttest
      if e.get('Action') in ['pass','fail'] and not e.get('Test'):complete=True
     if failed or panic:
      if proc.poll() is None:
       os.killpg(proc.pid,signal.SIGTERM);stopped=True
      break
     if proc.poll() is not None:break
     time.sleep(.2)
    try:code=proc.wait(timeout=10)
    except subprocess.TimeoutExpired:os.killpg(proc.pid,signal.SIGKILL);code=proc.wait()
   # Read any final buffered events after process exit.
   for line in log.read_text().splitlines():
    try:e=json.loads(line)
    except ValueError:continue
    test=e.get('Test','').split('/')[0]
    if e.get('Action')=='fail' and test:
     if excluded(test):wit.add(test)
     elif test not in pins:failed.add(test)
    if e.get('Action') in ['pass','fail'] and not e.get('Test'):complete=True
   if panic:failed.discard(panic)
   result['still_caught_by']=sorted(set(result['still_caught_by'])|failed);result['witness_failures']=sorted(set(result['witness_failures'])|wit);result['pin_failures']=sorted(set(result['pin_failures'])|pins)
   result['runs'].append(dict(command=cmd,cache=env['ADAMIC_BUILD_CACHE_DIR'],log=str(log),seconds=time.monotonic()-start,exit=code,stopped_at_failure=stopped,complete=complete))
   if failed:break
   if panic and panic not in extra:extra.append(panic);result['panicking_tests'].append(panic);continue
   if code!=0 and not (wit or pins):result['broken']='no observed test assertion failure; inspect log'
   break
 finally:subprocess.run(['git','apply','-R',m['file']],check=True)
 results.append(result);json.dump(results,open(P/'matrix.json','w'),indent=2);print(m['key'],'caught',result['still_caught_by'],'panic',result['panicking_tests'],'wall',round(sum(x['seconds'] for x in result['runs']),2),flush=True)
json.dump(results,open(P/'matrix.json','w'),indent=2)
