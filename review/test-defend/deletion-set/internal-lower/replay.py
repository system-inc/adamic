import json,subprocess,pathlib,time,re,os
r=pathlib.Path(__file__).resolve().parent
items=json.loads((r/'mutant-list.json').read_text());cs=json.loads((r/'skipped.json').read_text());results=[]
for i,x in enumerate(items):
 key=f'{i+1:02}-{x["mutant"]}';row=dict(x,still_caught_by=[],witness_failures=[],panicking_tests=[],runs=[])
 if x['stale']:results.append(row);continue
 diff=r/x['diff'];subprocess.run(['git','apply',str(diff)],check=True)
 try:
  skips=list(cs)
  for attempt in range(10):
   log=r/f'{key}-{attempt}.log';env=os.environ.copy();env.update(ADAMIC_GATE_UNCACHED='1',ADAMIC_BUILD_CACHE_DIR=f'/tmp/deletion-lower/cache/{key}-{attempt}')
   cmd=['go','test','-json','-count=1','-timeout','30m','./internal/lower/','-skip','^('+ '|'.join(re.escape(s) for s in skips)+')($|_|/)'];t=time.monotonic()
   with log.open('w') as out:p=subprocess.run(cmd,stdout=out,stderr=subprocess.STDOUT,env=env)
   ev=[]
   for l in log.read_text().splitlines():
    try:ev.append(json.loads(l))
    except:pass
   failed=sorted(set(e['Test'].split('/')[0] for e in ev if e.get('Action')=='fail' and e.get('Test')))
   witnesses=[f for f in failed if re.search(r'Mutants?|MutantKilled|Planted|Witness|AgreementRejects|AgreementAccepts',f)]
   panics=[e for e in ev if 'panic:' in e.get('Output','') or 'fatal error:' in e.get('Output','')]
   rr=dict(command=cmd,exit=p.returncode,wall_seconds=time.monotonic()-t,log=log.name,rows_failed=failed,panic=bool(panics));row['runs'].append(rr)
   row['witness_failures']=sorted(set(row['witness_failures']+witnesses))
   if panics:
    pt=next((e.get('Test','').split('/')[0] for e in panics if e.get('Test')),None)
    if not pt:row['broken']='Panic owner could not be identified';break
    row['panicking_tests'].append(pt);skips.append(pt);continue
   row['still_caught_by']=sorted(set(failed)-set(witnesses));row['complete']=any(e.get('Action') in ['pass','fail'] and not e.get('Test') for e in ev)
   if p.returncode and not failed:row['broken']='Build or package failure without a clean test failure'
   break
 finally:subprocess.run(['git','apply','-R',str(diff)],check=True)
 results.append(row);(r/'matrix.json').write_text(json.dumps(results,indent=2)+'\n');print(key,row['still_caught_by'],'seconds',round(sum(z['wall_seconds'] for z in row['runs']),2),flush=True)
