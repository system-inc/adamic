import pathlib,json,subprocess,os,time,re,signal
r=pathlib.Path(__file__).resolve().parent;main=subprocess.check_output(['git','rev-parse','HEAD']).decode().strip();items=json.loads((r/'mutant-list.json').read_text());results=[]
def terminate_tree(pid):
 children={}
 for p in pathlib.Path('/proc').iterdir():
  if not p.name.isdigit():continue
  try:parts=(p/'stat').read_text().rsplit(')',1)[1].split();parent=int(parts[1]);children.setdefault(parent,[]).append(int(p.name))
  except:pass
 def descend(x):
  a=[]
  for c in children.get(x,[]):a+=descend(c)+[c]
  return a
 for x in descend(pid)+[pid]:
  try:os.kill(x,signal.SIGTERM)
  except ProcessLookupError:pass
for i,x in enumerate(items):
 key=f'{i+1:02}-{x["mutant"]}';row=dict(x,still_caught_by=[],witness_failures=[],panicking_tests=[],runs=[])
 if x['stale']:results.append(row);continue
 diff=r/x['diff'];subprocess.run(['git','apply',str(diff)],check=True);files=[l[6:] for l in diff.read_text().splitlines() if l.startswith('+++ b/')];subprocess.run(['git','add','--',*files],check=True)
 c=subprocess.run(['git','commit','-m','Scratch YAML replay '+key],capture_output=True,text=True);(r/(key+'-commit.log')).write_text(c.stdout+c.stderr);assert c.returncode==0;row['scratch_commit']=subprocess.check_output(['git','rev-parse','HEAD']).decode().strip()
 try:
  skips=['TestLexerMatchesGo']
  for attempt in range(30):
   env=os.environ.copy();env.update(ADAMIC_GATE_UNCACHED='1',ADAMIC_BUILD_CACHE_DIR=f'/tmp/deletion-yaml/cache/{key}-{attempt}',ADAMIC_YAML_LIBRARY='/tmp/deletion-yaml/library')
   cmd=['go','test','-json','-count=1','-timeout','30m','./stage1/cohere/yaml/','-skip','^('+ '|'.join(re.escape(t) for t in skips)+')($|_|/)'];log=r/f'{key}-{attempt}.log';start=time.monotonic();ev=[];stop=False
   with log.open('w') as out, log.open('r') as inp:
    proc=subprocess.Popen(cmd,stdout=out,stderr=subprocess.STDOUT,env=env,start_new_session=True)
    while proc.poll() is None:
     for line in inp.readlines():
      try:e=json.loads(line);ev.append(e)
      except:continue
      if e.get('Action')=='fail' and e.get('Test') and '/' not in e['Test'] and not re.search('Mutants?|MutantKilled|Planted|Witness',e['Test']):
       stop=True;terminate_tree(proc.pid);break
     if stop:break
     time.sleep(.25)
    try:proc.wait(timeout=10)
    except subprocess.TimeoutExpired:terminate_tree(proc.pid);proc.kill();proc.wait()
   ev=[]
   for line in log.read_text().splitlines():
    try:ev.append(json.loads(line))
    except:pass
   failed=sorted(set(e['Test'].split('/')[0] for e in ev if e.get('Action')=='fail' and e.get('Test')));witness=[t for t in failed if re.search('Mutants?|MutantKilled|Planted|Witness',t)];panics=[e for e in ev if 'panic:' in e.get('Output','') or 'fatal error:' in e.get('Output','')]
   row['runs'].append(dict(command=cmd,log=log.name,wall_seconds=time.monotonic()-start,exit=proc.returncode,rows_failed=failed,stopped_at_first_non_witness_failure=stop));row['witness_failures']=sorted(set(row['witness_failures']+witness))
   if panics:
    pt=next((e.get('Test','').split('/')[0] for e in panics if e.get('Test')),None)
    if not pt:row['broken']='Panic owner not identified';break
    row['panicking_tests'].append(pt);skips.append(pt);continue
   row['still_caught_by']=sorted(set(failed)-set(witness));row['complete']=any(e.get('Action') in ['pass','fail'] and not e.get('Test') for e in ev)
   if proc.returncode and not failed:row['broken']='Package or build failure without a clean test failure'
   break
 finally:subprocess.run(['git','checkout','--detach',main],check=True,capture_output=True)
 results.append(row);(r/'matrix.json').write_text(json.dumps(results,indent=2)+'\n');print(key,row['still_caught_by'],round(sum(z['wall_seconds'] for z in row['runs']),2),flush=True)
