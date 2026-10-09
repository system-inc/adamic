import pathlib,json,subprocess,os,time,difflib,re
p=pathlib.Path('review/test-audit/internal-oracle-arguments_length');rows=json.loads((p/'rows.json').read_text());names=[r['test'] for r in rows];plan=[]
def add(f,old,new,kind):
 s=pathlib.Path(f).read_text();assert s.count(old)==1,(f,old);plan.append(dict(id='M%02d'%(len(plan)+1),file=f,line=s[:s.index(old)].count('\n')+1,old=old,new=new,kind=kind))
add('internal/native/emit_functions.go','count := "(double)argument_count"','count := "(double)argument_count + 1"','off-by-one')
add('internal/lower/cast_proof.go','reflect.DeepEqual(previous.AsLiteralType().Value(), literal.AsLiteralType().Value())','!reflect.DeepEqual(previous.AsLiteralType().Value(), literal.AsLiteralType().Value())','flip condition')
add('internal/native/exceptions.go','if e.program.ClosuresMayThrow {\n\t\te.checkThrown(holds...)','if false {\n\t\te.checkThrown(holds...)','flip condition')
add('internal/javascript/readiness.go','allowed.includes(adamicViewField(object, field, field, type)) ? object : panic(message)','!allowed.includes(adamicViewField(object, field, field, type)) ? object : panic(message)','flip condition')
(p/'mutant-plan.json').write_text(json.dumps(plan,indent=2));(p/'diffs').mkdir(exist_ok=True)
# Test-file helpers are invisible to Go production coverage. Keep a conservative source inventory.
helpers=[]
for f in [pathlib.Path('internal/oracle/oracle_test.go'),pathlib.Path('internal/oracle/cache_test.go'),pathlib.Path('internal/oracle/cast_test.go'),pathlib.Path('internal/oracle/call_targets_guards_test.go'),pathlib.Path('internal/oracle/library_string_test.go')]:
 for line,s in enumerate(f.read_text().splitlines(),1):
  if s.startswith('func ') and not re.match(r'func (Test|init)',s):helpers.append(str(f)+':'+str(line)+' '+s)
(p/'test-helper-inventory.txt').write_text('\n'.join(helpers)+'\n')
original={m['file']:pathlib.Path(m['file']).read_text() for m in plan};results=[]
try:
 for m in plan:
  before=original[m['file']];after=before.replace(m['old'],m['new']);pathlib.Path(m['file']).write_text(after);(p/'diffs'/(m['id']+'.diff')).write_text(''.join(difflib.unified_diff(before.splitlines(True),after.splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file'])))
  start=time.monotonic()
  with (p/'logs'/('vet-'+m['id']+'.log')).open('w') as out:r=subprocess.run(['timeout','90','go','vet','./'+str(pathlib.Path(m['file']).parent)+'/'],stdout=out,stderr=subprocess.STDOUT)
  m['vet_seconds']=time.monotonic()-start;m['vet_exit']=r.returncode;pathlib.Path(m['file']).write_text(before);assert r.returncode==0,m['id']
 switched=dict(original)
 for m in plan:
  if m['id']=='M01':replacement='count := auditString("M01", "(double)argument_count", "(double)argument_count + 1")'
  elif m['id']=='M02':replacement='auditBool("M02", '+m['old']+', '+m['new']+')'
  elif m['id']=='M03':replacement='if e.program.ClosuresMayThrow && !auditIs("M03") {\n\t\te.checkThrown(holds...)'
  else:continue
  switched[m['file']]=switched[m['file']].replace(m['old'],replacement)
 switched['internal/javascript/readiness.go']=switched['internal/javascript/readiness.go'].replace('const fieldReadinessRuntime =','var fieldReadinessRuntime =')
 for f,s in switched.items():pathlib.Path(f).write_text(s)
 for package in ['native','lower']:
  pathlib.Path('internal/'+package+'/audit_switch.go').write_text('package '+package+'\nimport "os"\nfunc auditIs(id string)bool{return os.Getenv("ADAMIC_MUTANT")==id}\nfunc auditBool(id string,a,b bool)bool{if auditIs(id){return b};return a}\nfunc auditString(id,a,b string)string{if auditIs(id){return b};return a}\n')
 m=plan[-1];pathlib.Path('internal/javascript/audit_switch.go').write_text('package javascript\nimport("os";"strings")\nfunc init(){if os.Getenv("ADAMIC_MUTANT")=="M04"{fieldReadinessRuntime=strings.Replace(fieldReadinessRuntime,'+json.dumps(m['old'])+','+json.dumps(m['new'])+',1)}}\n')
 start=time.monotonic()
 with (p/'logs'/'compile-switch.log').open('w') as out:subprocess.run(['go','test','-c','-o','/tmp/u055.test','./internal/oracle/'],stdout=out,stderr=subprocess.STDOUT,check=True)
 (p/'switch-build.json').write_text(json.dumps(dict(seconds=time.monotonic()-start)))
 for m in plan:
  env=dict(os.environ,ADAMIC_MUTANT=m['id'],ADAMIC_BUILD_CACHE_DIR='/tmp/u055/cache/'+m['id']);cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run','^('+'|'.join(names)+')$'];start=time.monotonic();log=p/'logs'/(m['id']+'.log')
  with log.open('w') as out:r=subprocess.run(cmd,env=env,stdout=out,stderr=subprocess.STDOUT)
  es=[]
  for s in log.read_text().splitlines():
   if s.startswith('{'):
    try:es.append(json.loads(s))
    except:pass
  fails=sorted(set(e['Test'].split('/')[0] for e in es if e['Action']=='fail' and e.get('Test')));panic=any(e.get('Output','').startswith('panic:') for e in es)
  result=dict(id=m['id'],exit=r.returncode,kills=fails,command=' '.join(cmd),environment={'ADAMIC_MUTANT':m['id'],'ADAMIC_BUILD_CACHE_DIR':env['ADAMIC_BUILD_CACHE_DIR']},wall_seconds=time.monotonic()-start,panic=panic,bounded=True,matrix_rows=names,events=es)
  assert not panic and r.returncode!=124,'isolated reruns needed'
  results.append(result);(p/'matrix.json').write_text(json.dumps(results,indent=2));print(m['id'],r.returncode,fails,flush=True)
finally:
 for f,s in original.items():pathlib.Path(f).write_text(s)
 for package in ['native','lower','javascript']:pathlib.Path('internal/'+package+'/audit_switch.go').unlink(missing_ok=True)
 (p/'mutant-plan.json').write_text(json.dumps(plan,indent=2))
