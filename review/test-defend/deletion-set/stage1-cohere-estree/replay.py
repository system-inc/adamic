import json,os,pathlib,re,signal,subprocess,time,sys
p=pathlib.Path('review/test-defend/deletion-set/stage1-cohere-estree')
candidates=['TestInterfaceDefaultGap','TestInterfaceTypeMethodGap','TestUnattachedDecoratorControl']
items=json.loads((p/'apply-check.json').read_text());records=json.loads((p/'matrix.json').read_text()) if (p/'matrix.json').exists() else []
def witness(test):
 return bool(re.search('Mutant|Planted|Disagreement',test)) or test in ['TestAcceptanceDiagnosticControl','TestLossyInputControl','TestPortStallControl_000','TestPortStallControlUnion','TestProduct_LossyInputControlLowered','TestProduct_LossyInputControlNative']
def stop_tree(proc):
 data=subprocess.check_output(['ps','-eo','pid=,ppid='],text=True);pairs=[tuple(map(int,l.split())) for l in data.splitlines()]; descendants={proc.pid}
 while True:
  new={pid for pid,ppid in pairs if ppid in descendants};n=len(descendants);descendants.update(new)
  if len(descendants)==n:break
 for pid in descendants-{proc.pid}:
  try:os.kill(pid,signal.SIGTERM)
  except ProcessLookupError:pass
 try:os.killpg(proc.pid,signal.SIGTERM)
 except ProcessLookupError:pass
 try:proc.wait(timeout=10)
 except subprocess.TimeoutExpired:os.killpg(proc.pid,signal.SIGKILL);proc.wait()
for x in items:
 if len(sys.argv)>1 and x['slug'] not in sys.argv[1:]:continue
 if any(r['slug']==x['slug'] for r in records):continue
 record=dict(x,still_caught_by=[],witness_failures=[],panicking_tests=[],runs=[])
 if x['stale']:
  records.append(record);(p/'matrix.json').write_text(json.dumps(records,indent=2)+'\n');continue
 checked=subprocess.run(['git','apply','--check',x['local_diff']],capture_output=True,text=True)
 if checked.returncode:record['stale']=True;record['apply_check']=checked.stderr;records.append(record);(p/'matrix.json').write_text(json.dumps(records,indent=2)+'\n');continue
 subprocess.run(['git','apply',x['local_diff']],check=True)
 try:
  extra=[]
  for attempt in range(300):
   names=candidates+extra;skip='^('+'|'.join(re.escape(n) for n in names)+')($|_|/)'
   cmd=['go','test','-json','-count=1','-timeout','30m','./stage1/cohere/estree/','-skip',skip]
   cache='/tmp/estree-deletion-set/cache/replay-v3/'+x['slug']+'-'+str(attempt);env=dict(os.environ,ADAMIC_GATE_UNCACHED='1',ADAMIC_BUILD_CACHE_DIR=cache)
   logpath=p/(x['slug']+'-'+str(attempt)+'.log');start=time.monotonic();statuses={};panics=[];last='';early=False;fail_lines={}
   proc=subprocess.Popen(cmd,env=env,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,text=True,start_new_session=True)
   with logpath.open('w') as log:
    for raw in proc.stdout:
     log.write(raw);log.flush()
     try:e=json.loads(raw)
     except:continue
     test=e.get('Test',''); top=test.split('/')[0]
     if top:last=top
     out=e.get('Output','')
     if re.match(r'^panic: ',out.strip()):
      panics.append(top or last);record['panicking_tests'].append(top or last)
     if top and out and e.get('OutputType')=='error':fail_lines.setdefault(top,[]).append(out.strip())
     if test and '/' not in test and e.get('Action') in ['pass','fail','skip']:
      action=e['Action'];statuses[test]=action
      if action=='fail':
       if witness(test):record['witness_failures'].append(test)
       elif not panics:
        # Go's panic recovery prints FAIL before the panic stack. Wait briefly,
        # stop remaining work, then drain the pipe before declaring a clean catch.
        time.sleep(0.5); stop_tree(proc)
        remainder=proc.stdout.read(); log.write(remainder); log.flush()
        for tail in remainder.splitlines():
         try:extra_event=json.loads(tail)
         except:continue
         extra_test=extra_event.get('Test',''); extra_top=extra_test.split('/')[0]
         extra_out=extra_event.get('Output','')
         if re.match(r'^panic: ',extra_out.strip()):
          panic_test=extra_top or test;panics.append(panic_test);record['panicking_tests'].append(panic_test)
         if extra_top and extra_out and extra_event.get('OutputType')=='error':fail_lines.setdefault(extra_top,[]).append(extra_out.strip())
         if extra_test and '/' not in extra_test and extra_event.get('Action') in ['pass','fail','skip']:statuses[extra_test]=extra_event['Action']
        if not panics:
         record['still_caught_by'].extend(sorted(n for n,a in statuses.items() if a=='fail' and not witness(n))); early=True
        record['witness_failures'].extend(n for n,a in statuses.items() if a=='fail' and witness(n))
        break
   code=proc.wait(); record['runs'].append(dict(command=' '.join(cmd),cache=cache,log=str(logpath),wall_seconds=time.monotonic()-start,exit=code,early_stopped=early,completed=not early and not panics,statuses=statuses,failure_lines=fail_lines))
   (p/'matrix.json').write_text(json.dumps(records+[record],indent=2)+'\n')
   if record['still_caught_by'] or not panics:break
   new=[n for n in panics if n and n not in names]
   if not new:record['broken']='Panic could not be assigned to an additional test';break
   extra.extend(new)
  record['witness_failures']=sorted(set(record['witness_failures']));record['panicking_tests']=sorted(set(record['panicking_tests']));records.append(record);(p/'matrix.json').write_text(json.dumps(records,indent=2)+'\n');print(x['slug'],record['still_caught_by'],'panics',record['panicking_tests'],flush=True)
 finally:subprocess.run(['git','apply','-R',x['local_diff']],check=True)
