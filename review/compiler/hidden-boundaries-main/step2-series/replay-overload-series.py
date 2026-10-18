import json,os,pathlib,re,subprocess,time,difflib,concurrent.futures
root=pathlib.Path.cwd();ev=root/'review/compiler/hidden-boundaries-main/step2-series';history=ev/'historical/overload-results';jobs=[]
def sel(old):
 bits=old.replace('^','').replace('$','').split('/')
 if bits[0]=='TestNativeAgreesWithNode':return old
 if bits[0]=='TestHiddenTNodeConstraintRepresentation':
  if len(bits)==1:return '^TestHiddenTNodeConstraint'
  return '^TestHiddenTNodeConstraint'+''.join(x.capitalize() for x in bits[1].split('_'))+'$'
 return '^'+bits[0]+''.join(''.join(x[0].upper()+x[1:] for x in re.split('[^a-zA-Z0-9]+',p) if x) for p in bits[1:])+'$'
for group in ['results','callback','structural','fields','admission']:
 script=history/'run-mutants.py' if group=='results' else history/'groups'/group/'run-mutants.py'
 s=script.read_text();stop=s.index('results = []') if group=='results' else s.index('rows=[]');prefix=s[:stop];prefix=re.sub(r'root\s*=\s*Path\(__file__\).resolve\(\).parents\[\d+\]', 'root = Path.cwd()',prefix)
 ns={'__file__':str(script)};exec(prefix,ns)
 rows=ns['mutations'] if group=='results' else ns['mutants']
 for row in rows:
  if group=='results':name,changed,pkg,selector=row;source=ns['source'];marker=''
  elif group=='callback':name,changed=row;source=ns['source'];pkg='./internal/lower';selector='TestOverloadCallbackFieldStorage' if name=='field-storage' else 'TestOverloadCallbackUnserved/'+name;marker=''
  else:name,source,changed,selector,marker=row;pkg='./internal/lower'
  if changed==source.read_text():print('NO CHANGE',group,name,flush=True);raise SystemExit(2)
  jobs.append((group+'-'+name,source,changed,pkg,sel(selector),marker))
source=root/'internal/lower/overload_results.go';original=source.read_text();needle='body = append(body, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: test}, Then: []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant(message)}}}})';assert original.count(needle)==1
for field in ['kind','value']:jobs.append(('field-hatch-'+field,source,original.replace(needle,'if field == "" { '+needle+' }'),'./internal/oracle','^TestOverloadFieldHatch'+field.capitalize()+'Liar$','unchecked wrong result'))
for name,guard,fixture in [('drop-indirect-check','closure == nil','returned'),('skip-single-signature-check','resolved.Declaration() == overload','narrow')]:jobs.append(('values-'+name,source,original.replace(needle,'if '+guard+' { '+needle+' }'),'./internal/oracle','^TestOverloadValues'+fixture.capitalize()+'Liar$','unchecked wrong result'))
for name,file,needle,replacement,fixture in [('erase-TIn-to-Node','overload_visitor_calls.go','l.localTypes[local] = bindings[i]','l.localTypes[local] = l.concrete(l.checker.GetTypeAtLocation(parameter))','Helper'),('admit-unproven-invocation','overload_visitor_domain.go','declaration.Body().ForEachChild(scan)\n\treturn result','declaration.Body().ForEachChild(scan)\n\tresult.unproven = nil\n\treturn result','OverloadedHelper')]:
 p=root/'internal/lower'/file;s=p.read_text();assert s.count(needle)==1;jobs.append(('visitors-'+name,p,s.replace(needle,replacement),'./internal/oracle','^TestOverloadVisitors'+fixture+'Liar$','unproven visitor input admitted'))
# The moved main binder result must remain guarded.
source=root/'internal/lower/overload_results.go';s=source.read_text();old='if !l.censusProveOverloadResult(implementation, overload) && !l.censusRelated(produced, promised) {';assert old in s
jobs.append(('binder-result-drop-check',source,s.replace(old,'if false && !l.censusProveOverloadResult(implementation, overload) && !l.censusRelated(produced, promised) {'),'./internal/oracle','^TestNativeAgreesWithNode$/internal/oracle/testdata/census_overload_binder_result_checked.a$',''))
def run(job):
 name,p,changed,pkg,selector,marker=job;copy=ev/(name+'.go.txt');copy.write_text(changed);original=p.read_text();(ev/(name+'.diff')).write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile=str(p.relative_to(root)),tofile=str(p.relative_to(root)))));overlay=ev/(name+'-overlay.json');overlay.write_text(json.dumps({'Replace':{str(p):str(copy)}}));cmd=['timeout','180','go','test','-overlay',str(overlay),pkg,'-run',selector,'-count=1','-timeout','90s','-json'];start=time.monotonic()
 with (ev/(name+'.json.log')).open('w') as log:result=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1','GOMAXPROCS':'4'})
 events=[]
 for line in (ev/(name+'.json.log')).read_text().splitlines():
  try:events.append(json.loads(line))
  except ValueError:pass
 failures=[{'test':d['Test'],'seconds':d.get('Elapsed')} for d in events if d.get('Action')=='fail' and d.get('Test')];output=''.join(d.get('Output','') for d in events);caught=result.returncode!=0 and bool(failures) and marker in output
 row=dict(mutant=name,exit=result.returncode,caught=caught,selector=selector,marker=marker,failures=failures,wall_seconds=round(time.monotonic()-start,3));print(name,'caught' if caught else 'NOT CAUGHT',failures,flush=True);return row
results=[]
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
 for row in pool.map(run,jobs):
  results.append(row);(ev/'mutants-current.json').write_text(json.dumps(results,indent=2)+'\n')
if not all(r['caught'] for r in results):raise SystemExit(1)
