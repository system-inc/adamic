import pathlib,subprocess,time,json,os,difflib,re
r=pathlib.Path('/workspace/adamic');e=r/'review/test-audit/cmd-adamic-test262-cache';pkg='cmd/adamic-test262';paths={f:r/pkg/f for f in ['cache.go','compiler.go','run.go']};bases={f:p.read_text() for f,p in paths.items()}
# The menu is fixed before checking any mutant failure.
menu=[]
def add(mid,f,old,new,kind,description,mode='expression'):
 assert old in bases[f],(mid,old)
 offset=bases[f].index(old);line=bases[f][:offset].count('\n')+1
 menu.append(dict(id=mid,file=pkg+'/'+f,line=line,old=old,new=new,kind=kind,change=description,mode=mode,offset=offset))
add('M01','cache.go','cacheKey("test262-compiler-v1", program, compiler, command, context)','cacheKey("test262-compiler-v1", "", compiler, command, context)','change constant','omit compiler source key dimension')
add('M02','cache.go','cacheKey("test262-node-v1", program, version, adaptation, command, context)','cacheKey("test262-node-v1", "", version, adaptation, command, context)','change constant','omit Node source key dimension')
add('M03','cache.go','cacheKey("test262-native-v1", code, library, command, context)','cacheKey("test262-native-v1", "", library, command, context)','change constant','omit native C key dimension')
add('M04','cache.go','fmt.Fprintf(hash, "%d:", len(part))','','drop statement','drop key part length boundaries','statement')
add('M05','cache.go','recordedExecution{[]byte(value.Stdout), []byte(value.Stderr), value.Exit, value.Signal, value.TimedOut}','recordedExecution{[]byte(value.Stderr), []byte(value.Stdout), value.Exit, value.Signal, value.TimedOut}','swap two arguments','swap stdout and stderr in stored execution')
add('M06','cache.go','envelope.Digest == cacheKey(string(envelope.Value))','true','change constant','accept envelope without digest validation')
add('M07','cache.go','!reusable || result.TimedOut','reusable || result.TimedOut','flip condition','invert reusable eligibility')
# M08 targets a disjoint token within the M07 region; combine both edits explicitly below.
add('M08','cache.go','result.TimedOut || (result.Exit == -1','!result.TimedOut || (result.Exit == -1','flip condition','invert timeout eligibility')
add('M09','cache.go','os.Getenv("ADAMIC_GATE_UNCACHED") == "1"','os.Getenv("ADAMIC_GATE_UNCACHED") == "0"','change constant','invert bypass selector')
add('M10','cache.go','_ = os.Rename(temporary.Name(), path)','','drop statement','drop atomic cache publication','statement')
add('M11','run.go','importedProgram.MatchString(codeOnly(program)) || referencedProgram.MatchString(program)','importedProgram.MatchString(codeOnly(program)) && referencedProgram.MatchString(program)','flip condition','require both dependency forms')
add('M12','run.go','directory = e.work\n\t\tif err := os.MkdirAll(directory, 0700)','','drop statement','drop cache mkdir fallback directory assignment','fallback')
add('M13','run.go','if e.compiler != nil {\n\t\t// The worker executes','if e.compiler == nil {\n\t\t// The worker executes','flip condition','invert stable worker command selection','condition-prefix')
add('M14','compiler.go','execution{Exit: -1, TimedOut: true}','execution{Exit: -1, TimedOut: false}','change constant','drop timeout indication')
add('M15','compiler.go','worker.command = nil','','drop statement','retain command after close','statement')
add('M16','compiler.go','execution{Exit: 1, Stderr: diagnostic.String()}','execution{Exit: 0, Stderr: diagnostic.String()}','change constant','report checker diagnostics as successful compilation')
add('M17','run.go','attempted >= limit','attempted > limit','off-by-one bound','allow one extra attempted test')
add('M18','run.go','(index+1)%50 == 0','(index+1)%49 == 0','change constant','change progress interval')
add('M19','run.go','report.Pass++','report.Pass += 2','off-by-one bound','double pass counter increment','statement')
add('M20','cache.go','!strings.HasPrefix(variable, "ADAMIC_GATE_UNCACHED=")','true','change constant','include bypass environment in cache context')
# Standalone M12 drops only the assignment, retaining the next statement.
menu[11]['new']='if err := os.MkdirAll(directory, 0700)'
probes=[]
def probe(mid,f,signature,empty):
 assert signature in bases[f]
 probes.append(dict(id=mid,file=pkg+'/'+f,line=bases[f][:bases[f].index(signature)].count('\n')+1,signature=signature,empty=empty))
probe('PReuse','cache.go','func (cache *resultCache) reuse(key string, execute func() (execution, bool)) execution {','execution{}')
probe('PObserve','cache.go','func (cache *resultCache) observe(key string, execute func() (execution, bool)) execution {','execution{}')
probe('PCompilerKey','cache.go','func compilerResultKey(program, compiler, command, context string) string {','""')
probe('PNodeKey','cache.go','func nodeResultKey(program, version, adaptation, command, context string) string {','""')
probe('PNativeKey','cache.go','func nativeResultKey(code, library, command, context string) string {','""')
probe('PKey','cache.go','func cacheKey(parts ...string) string {','""')
probe('PFilter','run.go','func (e *engine) runFilter(filter string, limit int, classifyOnly bool) (filterReport, error) {','filterReport{}, nil')
probe('PAttempt','run.go','func (e *engine) attempt(test classified) result {','result{}')
probe('PDependent','run.go','func dependentProgram(program string) bool {','false')
probe('PCompile','compiler.go','func compileInProcess(path string) (result execution) {','execution{}')
probe('PWorker','compiler.go','func (worker *compilerWorker) compile(path string) execution {','execution{}')
(e/'menu.json').write_text(json.dumps(menu,indent=2));(e/'probes.json').write_text(json.dumps(probes,indent=2))
for m in menu:
 f=m['file'].split('/')[-1];changed=bases[f].replace(m['old'],m['new'],1)
 (e/(m['id']+'.diff')).write_text(''.join(difflib.unified_diff(bases[f].splitlines(True),changed.splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file'])))
# Shared switch edits use unique disjoint spans. Resolve M07/M08's overlap as one original expression.
switched={}
for f,base in bases.items():
 edits=[]
 for m in menu:
  if m['file']!=pkg+'/'+f or m['id'] in ['M07','M08']:continue
  old,new=m['old'],m['new'];mid=m['id']
  if m['mode']=='statement':replacement=f'if auditMutant("{mid}") {{ {new} }} else {{ {old} }}'
  elif m['mode']=='fallback':replacement='if !auditMutant("M12") { directory = e.work }; if err := os.MkdirAll(directory, 0700)'
  elif m['mode']=='condition-prefix':replacement='if auditChoice("M13", e.compiler != nil, e.compiler == nil) {\n\t\t// The worker executes'
  else:replacement=f'auditChoice("{mid}", {old}, {new})'
  edits.append((m['offset'],m['offset']+len(old),replacement))
 if f=='cache.go':
  old='!reusable || result.TimedOut';pos=base.index(old);edits.append((pos,pos+len(old),'auditChoice("M07", !reusable, reusable) || auditChoice("M08", result.TimedOut, !result.TimedOut)'))
 for start,end,value in sorted(edits,reverse=True):base=base[:start]+value+base[end:]
 for p in probes:
  if p['file']==pkg+'/'+f:base=base.replace(p['signature'],p['signature']+'\n if auditMutant("'+p['id']+'") { return '+p['empty']+' }',1)
 if f=='cache.go':base+='\nfunc auditMutant(id string) bool { return os.Getenv("ADAMIC_MUTANT") == id }\nfunc auditChoice[T any](id string, original, mutated T) T { if auditMutant(id) { return mutated }; return original }\n'
 switched[f]=base;(e/(f+'.switch.txt')).write_text(base)
for p in probes:
 f=p['file'].split('/')[-1];changed=bases[f].replace(p['signature'],p['signature']+'\n return '+p['empty'],1)
 (e/(p['id']+'.diff')).write_text(''.join(difflib.unified_diff(bases[f].splitlines(True),changed.splitlines(True),fromfile='a/'+p['file'],tofile='b/'+p['file'])))
# Record every package function observed in the clean slice, including preparation and oracle helpers.
coverage=(e/'functions-coverage.txt').read_text().splitlines()
reached=[line for line in coverage if not line.endswith('0.0%') and not line.startswith('total:')]
(e/'reached-functions.txt').write_text('\n'.join(reached)+'\nCompiler-worker init and compileInProcess also execute in children; child coverage is not merged. No lowering/native implementation will be mutated.\n')
if os.environ.get('U012_PREPARE_ONLY')=='1':print('Menu and standalone diffs prepared; no production mutation performed.');raise SystemExit()
env=os.environ.copy();env['ADAMIC_TEST262_MEASURE']='1';runs=[];rows=json.loads((e/'requested-rows.json').read_text());allrows=[x for x in (e/'list.log').read_text().splitlines() if x.startswith('Test')]
def run(cmd,log,extra={}):
 start=time.monotonic()
 with (e/log).open('w') as stream:q=subprocess.run(cmd,cwd=r,env=env|extra,stdout=stream,stderr=subprocess.STDOUT)
 result=dict(command=cmd,log=log,exit=q.returncode,wall=time.monotonic()-start,environment=extra);runs.append(result);(e/'mutant-runs.json').write_text(json.dumps(runs,indent=2));print(log,q.returncode,round(result['wall'],3),flush=True);return q.returncode
try:
 for m in menu:
  f=m['file'].split('/')[-1];paths[f].write_text(bases[f].replace(m['old'],m['new'],1))
  status=run(['timeout','90','go','vet','./cmd/adamic-test262/'],m['id']+'-vet.log');paths[f].write_text(bases[f])
  if status:raise SystemExit('standalone vet failed: '+m['id'])
 for f,source in switched.items():paths[f].write_text(source)
 assert run(['gofmt','-w',*[str(p) for p in paths.values()]],'switch-format.log')==0
 for f,p in paths.items():(e/(f+'.switch.txt')).write_text(p.read_text())
 assert run(['timeout','90','go','test','-c','-o','/tmp/u012-test','./cmd/adamic-test262/'],'switch-build.log')==0
 for mid in [m['id'] for m in menu]+[p['id'] for p in probes]:
  status=run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./cmd/adamic-test262/','-run','.'],mid+'.log',dict(ADAMIC_MUTANT=mid,ADAMIC_BUILD_CACHE_DIR='/tmp/u012/cache/'+mid))
  log=(e/(mid+'.log')).read_text()
  # Panics abort later rows; rerun every requested row alone with a capped binary.
  if 'panic:' in log and ('panic: test timed out' in log or '"Action":"fail"' in log) and ('goroutine ' in log):
   for row in rows:run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./cmd/adamic-test262/','-run','^'+row+'$'],mid+'-alone-'+row+'.log',dict(ADAMIC_MUTANT=mid,ADAMIC_BUILD_CACHE_DIR='/tmp/u012/cache/'+mid))
finally:
 for f,p in paths.items():p.write_text(bases[f])
