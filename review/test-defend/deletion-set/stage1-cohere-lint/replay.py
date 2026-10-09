from pathlib import Path
import subprocess,os,json,time,signal,re
p=Path('review/test-defend/deletion-set/stage1-cohere-lint');skip='^(TestNodeTableIsLinkOnlyFamily|TestNodeTableIsLinkOnly_[0-9]{3}|TestProfileCompilationBuildLower)($|_|/)'
def events(path):
 result=[]
 for line in path.read_text().splitlines():
  try:result.append(json.loads(line))
  except ValueError:pass
 return result
def witness(name):return bool(re.search('Mutant|Planted|Killed',name))
b=events(p/'baseline.log');assert b[-1].get('Action')=='pass' and not b[-1].get('Test'), 'Clean baseline must pass before replay'
items=json.loads((p/'mutant-list.json').read_text());results=[]
for i,item in enumerate(items):
 if item['stale']:
  results.append(dict(**item,status='stale',still_caught_by=[]));continue
 diff=item['local_diff'];subprocess.run(['git','apply',diff],check=True)
 try:
  extra=[];attempts=[]
  for repeat in range(20):
   label=f'R{i+1}-{item["mutant"]}-{repeat}';path=p/(label+'.log');env=os.environ.copy();env['ADAMIC_GATE_UNCACHED']='1';env['ADAMIC_BUILD_CACHE_DIR']='/tmp/lint-deletion/cache/'+label
   effective=skip if not extra else skip+'|^(?:'+'|'.join(re.escape(x) for x in extra)+')($|/)'
   cmd=['go','test','-json','-count=1','-timeout','30m','./stage1/cohere/lint/','-skip',effective];start=time.monotonic();stopped=False
   with path.open('w') as log:
    process=subprocess.Popen(cmd,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
    while process.poll() is None:
     es=events(path);panic=any('panic:' in e.get('Output','') for e in es)
     clean=[e['Test'] for e in es if e.get('Action')=='fail' and e.get('Test') and '/' not in e['Test'] and not witness(e['Test'])]
     if clean and not panic:
      stopped=True;os.killpg(process.pid,signal.SIGTERM);process.wait();break
     time.sleep(.5)
   es=events(path);fail=sorted({e['Test'] for e in es if e.get('Action')=='fail' and e.get('Test') and '/' not in e['Test']});panics=[]
   for n,e in enumerate(es):
    if 'panic:' in e.get('Output',''):
     test=e.get('Test')
     if not test:
      test=next((x.get('Test') for x in reversed(es[:n]) if x.get('Test')),None)
     if test:panics.append(test.split('/')[0])
   panics=sorted(set(panics));caught=[x for x in fail if not witness(x) and x not in panics];a=dict(command=cmd,cache=env['ADAMIC_BUILD_CACHE_DIR'],log=str(path),wall_seconds=time.monotonic()-start,exit=process.returncode,failed_tests=fail,witness_failures=[x for x in fail if witness(x)],panicking_tests=panics,still_caught_by=caught,stopped_after_first_failure=stopped);attempts.append(a)
   (p/(label+'.json')).write_text(json.dumps(a,indent=2)+'\n')
   if panics:
    fresh=[x for x in panics if x not in extra]
    if not fresh:raise RuntimeError('Could not identify new panicking test')
    extra+=fresh;continue
   if not caught and not any(e.get('Action') in ['pass','fail'] and not e.get('Test') for e in es):raise RuntimeError('Run did not complete or establish a clean catcher')
   results.append(dict(**item,attempts=attempts,still_caught_by=caught,status='caught' if caught else 'last_catcher_lost'));break
  (p/'matrix.json').write_text(json.dumps(results,indent=2)+'\n');print(item['branch'],item['mutant'],results[-1]['status'],results[-1]['still_caught_by'],flush=True)
 finally:subprocess.run(['git','apply','--reverse',diff],check=True)
