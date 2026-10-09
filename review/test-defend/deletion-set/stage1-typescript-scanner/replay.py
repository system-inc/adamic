import pathlib,json,subprocess,os,time,re
r=pathlib.Path(__file__).resolve().parent;main=subprocess.check_output(['git','rev-parse','HEAD']).decode().strip();items=json.loads((r/'mutant-list.json').read_text());results=[];cs=['TestProfileSnapshotsAgree','TestScannerAgreesWithTypescriptGo']
for i,x in enumerate(items):
 key=f'{i+1:02}-{x["mutant"]}';row=dict(x,still_caught_by=[],witness_failures=[],panicking_tests=[],runs=[])
 if x['stale']:results.append(row);continue
 diff=r/x['diff'];subprocess.run(['git','apply',str(diff)],check=True)
 files=[line[6:] for line in diff.read_text().splitlines() if line.startswith('+++ b/')];subprocess.run(['git','add','--',*files],check=True)
 commit=subprocess.run(['git','commit','-m','Scratch scanner replay '+key],capture_output=True,text=True);(r/(key+'-commit.log')).write_text(commit.stdout+commit.stderr);assert commit.returncode==0
 row['scratch_commit']=subprocess.check_output(['git','rev-parse','HEAD']).decode().strip()
 try:
  skips=list(cs)
  for attempt in range(20):
   env=os.environ.copy();env.update(ADAMIC_GATE_UNCACHED='1',ADAMIC_BUILD_CACHE_DIR=f'/tmp/deletion-scanner/cache/{key}-{attempt}',ADAMIC_TYPESCRIPT_SOURCE='/workspace/u111-typescript',ADAMIC_SCANNER_BENCH='1',ADAMIC_SCANNER_PROFILE_DIR=f'/tmp/deletion-scanner/artifacts/{key}-{attempt}')
   cmd=['go','test','-json','-count=1','-timeout','30m','./stage1/typescript/scanner/','-skip','^('+ '|'.join(re.escape(s) for s in skips)+')($|_|/)'];log=r/f'{key}-{attempt}.log';start=time.monotonic()
   with log.open('w') as out:run=subprocess.run(cmd,env=env,stdout=out,stderr=subprocess.STDOUT)
   ev=[]
   for line in log.read_text().splitlines():
    try:ev.append(json.loads(line))
    except:pass
   failed=sorted(set(e['Test'].split('/')[0] for e in ev if e.get('Action')=='fail' and e.get('Test')));witnesses=[f for f in failed if re.search('Mutants?|MutantKilled|Planted|Witness',f)];panics=[e for e in ev if 'panic:' in e.get('Output','') or 'fatal error:' in e.get('Output','')]
   row['runs'].append(dict(command=cmd,log=log.name,wall_seconds=time.monotonic()-start,exit=run.returncode,rows_failed=failed));row['witness_failures']=sorted(set(row['witness_failures']+witnesses))
   if panics:
    pt=next((e.get('Test','').split('/')[0] for e in panics if e.get('Test')),None)
    if not pt:row['broken']='Panic owner not identified';break
    row['panicking_tests'].append(pt);skips.append(pt);continue
   row['still_caught_by']=sorted(set(failed)-set(witnesses));row['complete']=any(e.get('Action') in ['pass','fail'] and not e.get('Test') for e in ev)
   if run.returncode and not failed:row['broken']='Package or build failure without a test failure'
   break
 finally:subprocess.run(['git','checkout','--detach',main],check=True,capture_output=True)
 results.append(row);(r/'matrix.json').write_text(json.dumps(results,indent=2)+'\n');print(key,row['still_caught_by'],round(sum(z['wall_seconds'] for z in row['runs']),2),flush=True)
