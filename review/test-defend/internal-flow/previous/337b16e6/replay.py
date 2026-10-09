import pathlib,subprocess,os,time,json,re,gzip
root=pathlib.Path('/workspace/adamic'); ev=pathlib.Path('/tmp/flow-defense/evidence')
base={f:subprocess.check_output(['git','show','HEAD:'+f],cwd=root).decode() for f in ['internal/flow/build.go','internal/flow/liveness.go']}
mutants=[
 ('D1','internal/flow/build.go','\tcase ir.Evaluate:\n\t\tb.emit(at, 0, statement.Value, b.uses(statement.Value), nil)','\tcase ir.Evaluate:\n\t\tif _, literal := statement.Value.(ir.NumberConstant); literal {\n\t\t\treturn\n\t\t}\n\t\tb.emit(at, 0, statement.Value, b.uses(statement.Value), nil)','return early for a literal-number Evaluate'),
 ('D2','internal/flow/liveness.go','position >= 0; position--','position > 0; position--','off by one: omit the first instruction transfer'),
 ('D3','internal/flow/liveness.go','\t\t\t\tfor _, define := range instruction.Defines {\n\t\t\t\t\tdelete(live, declaration(define))\n\t\t\t\t}','','drop the whole definition-kill loop; preserve future-read overapproximations'),
 ('D4','internal/flow/liveness.go','position := len(block.Instructions) - 1','position := len(block.Instructions) - 2','off by one: omit the final instruction transfer, including a return read'),
]
(ev/'planned-mutants.json').write_text(json.dumps([{'mutant':i,'file':f,'file_line':f+':'+str(base[f][:base[f].index(old)].count('\n')+1),'change':desc} for i,f,old,new,desc in mutants],indent=2))
results=[]
try:
 for i,f,old,new,desc in mutants:
  for p,s in base.items():root.joinpath(p).write_text(s)
  assert base[f].count(old)==1,(i,base[f].count(old))
  root.joinpath(f).write_text(base[f].replace(old,new))
  subprocess.run(['gofmt','-w',f],cwd=root,check=True)
  (ev/(i+'.diff')).write_bytes(subprocess.check_output(['git','diff','--',f],cwd=root))
  env=dict(os.environ,TMPDIR='/tmp/flow-defense/tmp',ADAMIC_BUILD_CACHE_DIR='/tmp/flow-defense/cache/'+i)
  start=time.monotonic()
  with (ev/(i+'-vet.log')).open('w') as out: vet=subprocess.run(['timeout','90','go','vet','./internal/flow/'],cwd=root,env=env,stdout=out,stderr=subprocess.STDOUT)
  r={'mutant':i,'change':desc,'vet_status':vet.returncode,'vet_wall':round(time.monotonic()-start,3)}
  if vet.returncode:results.append(r);break
  command=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/flow/','-run','.']
  start=time.monotonic()
  with (ev/(i+'.log')).open('w') as out: run=subprocess.run(command,cwd=root,env=env,stdout=out,stderr=subprocess.STDOUT)
  r.update(status=run.returncode,wall=round(time.monotonic()-start,3),command='ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(command))
  failed=[];passed=[];lines={};panic=False;timeout=False
  with (ev/(i+'.log')).open() as inp:
   for line in inp:
    try:x=json.loads(line)
    except:continue
    t=x.get('Test','')
    if x.get('Action')=='fail' and t and '/' not in t:failed.append(t)
    if x.get('Action')=='pass' and t and '/' not in t:passed.append(t)
    out=x.get('Output','')
    if 'panic:' in out:panic=True
    if 'test timed out' in out:timeout=True
    if t and re.search(r'\w+_test.go:\d+:',out) and t not in lines:lines[t]=out.strip()
  r.update(rows_failed=failed,rows_passed=passed,first_assertions=lines,panic=panic,timeout=timeout)
  results.append(r);(ev/'results.json').write_text(json.dumps(results,indent=2))
  print(i,r['status'],r['wall'],'failed',len(failed),'passed',len(passed),'timeout',timeout,flush=True)
  if (ev/(i+'.log')).stat().st_size>20_000_000:
   with (ev/(i+'.log')).open('rb') as inp,gzip.open(ev/(i+'.log.gz'),'wb') as out:
    import shutil;shutil.copyfileobj(inp,out)
   (ev/(i+'.log')).unlink()
finally:
 for p,s in base.items():root.joinpath(p).write_text(s)
 (ev/'results.json').write_text(json.dumps(results,indent=2))
