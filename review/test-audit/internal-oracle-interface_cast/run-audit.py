import pathlib,json,subprocess,os,time,difflib
root=pathlib.Path.cwd(); p=root/'review/test-audit/internal-oracle-interface_cast'; scope=json.loads((p/'scope.json').read_text()); plan=json.loads((p/'plan.json').read_text()); witnesses=['TestInterfaceCastRuntimeMutants','TestArrayFamilyMutants','TestArrayWithBoundsMutant','TestJSONStringifyOracleCatchesKeyOrder','TestLibraryMapSetMutants','TestMathNumberOracleCatchesMutants','TestLibraryStringMutants','TestLiteralOptionalOracleCatchesMutant']; production=[x for x in scope if x not in witnesses]; assert len(json.loads((p/'timings.json').read_text()))==15
files={x[1] for x in plan}|{'internal/load/source_fs.go','internal/load/load.go','internal/lower/lower.go','internal/native/emit.go','internal/javascript/javascript.go','internal/oracle/oracle_test.go'}; original={f:(root/f).read_text() for f in files}
def replace_body(text,signature,contents):
 start=text.index(signature); opening=text.index('{',start); end=opening+1; depth=1
 while depth:
  if text[end]=='{': depth+=1
  if text[end]=='}': depth-=1
  end+=1
 return text[:opening+1]+'\n'+contents+'\n'+text[end-1:]
def write_diff(id,f,text):
 (p/(id+'.diff')).write_text(''.join(difflib.unified_diff(original[f].splitlines(True),text.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
def restore():
 for f,text in original.items(): (root/f).write_text(text)
def command(args,log,env=None):
 start=time.monotonic()
 with (p/log).open('w') as out: result=subprocess.run(args,env=env,stdout=out,stderr=subprocess.STDOUT)
 return result.returncode,round(time.monotonic()-start,3)
compiles=[]
for id,f,before,after,menu in plan:
 assert before in original[f],(id,before); text=original[f].replace(before,after,1); write_diff(id,f,text); line=original[f][:original[f].index(before)].count('\n')+1; (p/(id+'.location')).write_text(f'{f}:{line}: {before} -> {after}\n'); (root/f).write_text(text)
 package='./'+str(pathlib.Path(f).parent)+'/'
 try:
  rc,secs=command(['go','vet',package],id+'-compile.log'); compiles.append({'id':id,'seconds':secs,'exit':rc,'command':'go vet '+package}); assert rc==0,id
 finally: restore()
(p/'compiles.json').write_text(json.dumps(compiles,indent=2))
f='internal/oracle/oracle_test.go'; sig='func disagreement(oracle run, native run) string'; w1=replace_body(original[f],sig,'\treturn ""'); write_diff('W01',f,w1); w2=original[f].replace('case !bytes.Equal(oracle.stdout, native.stdout):','case false && !bytes.Equal(oracle.stdout, native.stdout):',1); write_diff('W02',f,w2)
probes=[('P01',101,'internal/lower/lower.go','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error)','return nil, nil'),('P02',102,'internal/load/load.go','func Load(paths []string) (*Program, error)','return nil, nil'),('P03',103,'internal/native/emit.go','func C(program *ir.Program) string','return ""'),('P04',104,'internal/javascript/javascript.go','func JavaScript(program *ir.Program) string','return ""')]
for id,n,f,sig,answer in probes:
 text=replace_body(original[f],sig,'\t'+answer)
 if id=='P01': text=text.replace('\n\t"fmt"','').replace('\n\t"path/filepath"','')
 write_diff(id,f,text)
changed=original.copy()
for n,(id,f,before,after,menu) in enumerate(plan,1):
 if n==1:
  replacement=before.replace('cString(property.Name)','auditChoose(1, cString(property.Name), cString(property.View))').replace('cString(property.View))','auditChoose(1, cString(property.View), cString(property.Name)))')
  # Replace the two arguments independently, preserving expression order.
  replacement=before.replace('cString(property.Name)','AUDIT_FIELD').replace('cString(property.View)','auditChoose(1, cString(property.View), cString(property.Name))').replace('AUDIT_FIELD','auditChoose(1, cString(property.Name), cString(property.View))')
 elif n==2: replacement=before.replace('ir.Equal','auditChoose(2, ir.Equal, ir.NotEqual)')
 elif n==3: replacement='character >= auditChoose[byte](3, 0x20, 0x21)'
 elif n in [4,6,7,9]: replacement=f'auditChoose({n}, {before}, {after})'
 elif n==5: replacement=before # changed separately below, including quotes
 elif n==8: replacement='if auditChoose(8, isIterator(source), !isIterator(source)) {'
 elif n==10: continue
 elif n==11: replacement=before.replace('core.TSTrue','auditChoose(11, core.TSTrue, core.TSFalse)')
 elif n==12: replacement=before
 changed[f]=changed[f].replace(before,replacement,1)
for n in [5,12]:
 id,f,before,after,menu=plan[n-1]; old=json.dumps(before); new=json.dumps(after); assert old in changed[f]; changed[f]=changed[f].replace(old,f'auditChoose({n}, {old}, {new})',1)
changed['internal/load/source_fs.go']=changed['internal/load/source_fs.go'].replace('return prelude, true','return auditPrelude(prelude), true',1)
f='internal/oracle/oracle_test.go'; sig='func disagreement(oracle run, native run) string'; text=changed[f]; opening=text.index('{',text.index(sig))+1; changed[f]=text[:opening]+'\n\tif auditSelect() == 201 { return "" }\n'+text[opening:]; changed[f]=changed[f].replace('case !bytes.Equal(oracle.stdout, native.stdout):','case auditSelect() != 202 && !bytes.Equal(oracle.stdout, native.stdout):',1)
for id,n,f,sig,answer in probes:
 text=changed[f]; opening=text.index('{',text.index(sig))+1; changed[f]=text[:opening]+f'\n\tif auditSelect() == {n} {{ {answer} }}\n'+text[opening:]
helpers=[]
for package in ['load','lower','native','javascript','oracle']:
 helper=root/f'internal/{package}/audit_selector.go'; content=f'package {package}\nimport("os";"strconv")\nfunc auditSelect() int {{ n,_:=strconv.Atoi(os.Getenv("ADAMIC_MUTANT"));return n }}\nfunc auditChoose[T any](id int,before,after T) T {{if auditSelect()==id{{return after}};return before}}\n'
 if package=='load':
  content=content.replace('"strconv")','"strconv";"strings")'); before,after=plan[9][2:4]; content+=f'func auditPrelude(text string) string {{if auditSelect()==10{{return strings.Replace(text,{json.dumps(before)},{json.dumps(after)},1)}};return text}}\n'
 helper.write_text(content); helpers.append(helper)
for f,text in changed.items(): (root/f).write_text(text)
command(['gofmt','-w']+[str(root/f) for f in files if f.endswith('.go')]+[str(x) for x in helpers],'format.log')
(p/'switch.diff').write_text(subprocess.run(['git','diff','--','internal/load','internal/lower','internal/native','internal/javascript','internal/oracle'],capture_output=True,text=True).stdout)
(p/'selector-source.txt').write_text('\n'.join(x.read_text() for x in helpers))
env=os.environ|{'ADAMIC_GATE_UNCACHED':'1'}; records=[]
def run(id,n,rows):
 pattern='^('+'|'.join(rows)+')$'; args=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',pattern]; given=env|{'ADAMIC_MUTANT':str(n),'ADAMIC_BUILD_CACHE_DIR':'/tmp/u061/cache/'+id}; rc,secs=command(args,id+'.log',given); events=[]
 for line in (p/(id+'.log')).read_text().splitlines():
  try: events.append(json.loads(line))
  except: pass
 failed=sorted(set(e['Test'].split('/')[0] for e in events if e.get('Action')=='fail' and 'Test' in e)); passed=sorted(set(e['Test'].split('/')[0] for e in events if e.get('Action')=='pass' and 'Test' in e)); result={'id':id,'seconds':secs,'exit':rc,'matrix_rows':rows,'failed':failed,'passed':passed,'command':'ADAMIC_GATE_UNCACHED=1 ADAMIC_MUTANT='+str(n)+' ADAMIC_BUILD_CACHE_DIR='+given['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(args),'log':id+'.log'}; print(id,rc,secs,failed,flush=True); return result
try:
 baseline=run('inactive',0,scope); assert baseline['exit']==0,'inactive baseline failed'
 for n,(id,f,before,after,menu) in enumerate(plan,1):
  result=run(id,n,production)
  if 'panic:' in (p/(id+'.log')).read_text():
   result['reruns']=[run(id+'-alone-'+row,n,[row]) for row in production]; result['failed']=sorted(set(row for part in result['reruns'] for row in part['failed'])); result['passed']=sorted(set(row for part in result['reruns'] for row in part['passed']))
  records.append(result); (p/'matrix.json').write_text(json.dumps(records,indent=2))
 for id,n in [('W01',201),('W02',202)]:
  records.append(run(id,n,scope)); (p/'matrix.json').write_text(json.dumps(records,indent=2))
 for id,n,f,sig,answer in probes:
  rows=production if id=='P01' else ['TestJSONStringifyResultMayBeUndefined'] if id=='P02' else production[:5]
  result=run(id,n,rows)
  if 'panic:' in (p/(id+'.log')).read_text():
   result['reruns']=[run(id+'-alone-'+row,n,[row]) for row in rows]; result['failed']=sorted(set(row for part in result['reruns'] for row in part['failed'])); result['passed']=sorted(set(row for part in result['reruns'] for row in part['passed']))
  records.append(result); (p/'matrix.json').write_text(json.dumps(records,indent=2))
finally:
 restore()
 for helper in helpers: helper.unlink()
