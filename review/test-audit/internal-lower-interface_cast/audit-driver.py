import pathlib,json,time,subprocess,difflib,os,re
root=pathlib.Path('/workspace/adamic');out=root/'review/test-audit/internal-lower-interface_cast';rows=json.loads((out/'rows.json').read_text()); pattern=(out/'scope-pattern.txt').read_text().strip()
# Fixed menu from reached production functions, before checking any mutant kills.
menu=[
('M01','interface_cast.go',69,'field.Flags&ast.SymbolFlagsOptional != 0','field.Flags&ast.SymbolFlagsOptional == 0','flip condition','view'),
('M02','iteration.go',54,'Receiver: true','Receiver: false','change option','objectMethod'),
('M03','iteration.go',192,'done == nil','done != nil','flip condition','planIteration'),
('M04','iteration.go',311,'for _, module := range modules {\n\t\tmodule.AsNode().ForEachChild(visit)\n\t}', '', 'drop whole loop','planIteration'),
('M05','refusals.go',150,'if generator {','if generator && false {','change condition constant','refuse'),
('M06','cycles.go',63,'func (l *lowering) findCycles(modules []*ast.SourceFile) error {','func (l *lowering) findCycles(modules []*ast.SourceFile) error {\n if true { return nil }','return early','findCycles'),
('M07','class_inheritance.go',690,'func (l *lowering) nominalMismatch(from, to *checker.Type, seen map[[2]*checker.Type]bool) *checker.Type {','func (l *lowering) nominalMismatch(from, to *checker.Type, seen map[[2]*checker.Type]bool) *checker.Type {\n if true { return nil }','return early','nominalMismatch'),
('M08','class.go',417,'member.Flags&ast.SymbolFlagsMethod != 0','member.Flags&ast.SymbolFlagsMethod == 0','flip condition','setProperty'),
('M09','iteration_consume.go',42,'parameter != ir.Number','parameter == ir.Number','flip condition','collectIteration'),
('M10','class_features.go',77,'!isClassInstance(proven)','isClassInstance(proven)','flip condition','objectKeys'),
('M11','iteration_origin.go',80,'if isCallee(where) {','if !isCallee(where) {','flip condition','erasedMethodField'),
('M12','iteration_consume.go',167,'!l.includesUndefined(valueType)','l.includesUndefined(valueType)','flip condition','destructureIterator'),
('P01','lower.go',20,'func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {\n if true { return nil, nil }','empty-answer probe','Lower'),
('P02','class.go',410,'func (l *lowering) setProperty(target *ast.Node, valueNode *ast.Node) ([]ir.Statement, error) {','func (l *lowering) setProperty(target *ast.Node, valueNode *ast.Node) ([]ir.Statement, error) {\n if true { return nil, nil }','empty-answer probe','setProperty')]
bases={file:(root/'internal/lower'/file).read_text() for _,file,*_ in menu}
records=[]
for mid,file,line,old,new,kind,func in menu:
 base=bases[file]
 if mid=='M04':
  # Find the module walk inside planIteration, not another identical loop.
  start=base.index('func (l *lowering) planIteration(');pos=base.index(old,start);line=base[:pos].count('\n')+1
 else:
  ls=base.splitlines(True)
  assert old in ls[line-1],(mid,line,ls[line-1]); pos=sum(len(x) for x in ls[:line-1])+ls[line-1].index(old)
 assert re.search(r'\b'+func+r'\s', (out/'code-under-test-functions.txt').read_text()),func
 changed=base[:pos]+new+base[pos+len(old):]
 diff=''.join(difflib.unified_diff(base.splitlines(True),changed.splitlines(True),fromfile='a/internal/lower/'+file,tofile='b/internal/lower/'+file))
 (out/(mid+'.diff')).write_text(diff)
 records.append(dict(id=mid,file='internal/lower/'+file,line=line,old=old,new=new,kind=kind,function=func,pos=pos))
(out/'menu.json').write_text(json.dumps(records,indent=2)+'\n')
# timings on pristine code, one row alone three independent count=1 commands
trial=[]
for row in rows:
 for i in range(1,4):
  start=time.monotonic()
  with (out/(row+f'.time{i}.log')).open('w') as f:r=subprocess.run(['timeout','120','go','test','-count=1','-timeout','90s','./internal/lower/','-run','^'+row+'$'],cwd=root,stdout=f,stderr=subprocess.STDOUT)
  trial.append(dict(test=row,run=i,wall=time.monotonic()-start,exit=r.returncode))
  if r.returncode:raise SystemExit('timing baseline failed '+row)
(out/'timing.json').write_text(json.dumps(trial,indent=2)+'\n')
# Every standalone diff must compile; restore after each, record costs.
validation=[]
for rec in records:
 file=rec['file'].split('/')[-1];base=bases[file];pos=rec['pos'];path=root/rec['file']
 path.write_text(base[:pos]+rec['new']+base[pos+len(rec['old']):])
 start=time.monotonic()
 with (out/(rec['id']+'.vet.log')).open('w') as f:r=subprocess.run(['timeout','90','go','vet','./internal/lower/'],cwd=root,stdout=f,stderr=subprocess.STDOUT)
 validation.append(dict(id=rec['id'],seconds=time.monotonic()-start,exit=r.returncode));path.write_text(base)
 (out/'validation.json').write_text(json.dumps(validation,indent=2)+'\n')
 if r.returncode:raise SystemExit('vet failed '+rec['id'])
# Switch requires one helper file; standalone diffs contain no helper or switch.
helper=root/'internal/lower/u033_audit_mutant.go';helper.write_text('package lower\nimport "os"\nfunc u033Mutant(id string) bool { return os.Getenv("ADAMIC_MUTANT") == id }\n')
for file,base in bases.items():
 edits=[]
 for rec in records:
  if rec['file']!='internal/lower/'+file:continue
  old,new,mid=rec['old'],rec['new'],rec['id']
  if rec['kind'] in ['empty-answer probe','return early']:
   empty='nil, nil' if mid.startswith('P') else 'nil'
   replace=old+'\n if u033Mutant("'+mid+'") { return '+empty+' }'
  elif mid=='M04':replace='if !u033Mutant("M04") { '+old+' }'
  elif mid=='M02':replace='Receiver: !u033Mutant("M02")'
  elif mid=='M05':replace='if generator && !u033Mutant("M05") {'
  elif mid=='M11':replace='if isCallee(where) != u033Mutant("M11") {'
  else:replace='(('+old+') != u033Mutant("'+mid+'"))'
  edits.append((rec['pos'],old,replace))
 for pos,old,replace in sorted(edits,reverse=True):base=base[:pos]+replace+base[pos+len(old):]
 (root/'internal/lower'/file).write_text(base)
subprocess.run(['gofmt','-w',str(helper)]+[str(root/'internal/lower'/file) for file in bases],check=True)
with (out/'switch.diff').open('w') as f:subprocess.run(['git','diff','--','internal/lower'],cwd=root,stdout=f,check=True)
(out/'switch-helper.go.txt').write_text(helper.read_text())
start=time.monotonic()
with (out/'switch-build.log').open('w') as f:r=subprocess.run(['timeout','90','go','test','-c','-o','/tmp/u033-lower.test','./internal/lower/'],cwd=root,stdout=f,stderr=subprocess.STDOUT)
(out/'switch-build-time.json').write_text(json.dumps({'seconds':time.monotonic()-start,'exit':r.returncode})+'\n')
if r.returncode:raise SystemExit('switch build failed')
runs=[]
for rec in records:
 mid=rec['id'];env=os.environ.copy();env['ADAMIC_MUTANT']=mid;env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u033/cache/'+mid
 runpattern='.' if mid.startswith('M') else pattern
 start=time.monotonic()
 with (out/(mid+'.log')).open('w') as f:r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run',runpattern],cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
 runs.append(dict(id=mid,seconds=time.monotonic()-start,exit=r.returncode,pattern=runpattern));(out/'runs.json').write_text(json.dumps(runs,indent=2)+'\n');print(mid,r.returncode,flush=True)
 text=(out/(mid+'.log')).read_text()
 if 'panic:' in text or r.returncode==124:
  # A cooked/panicked full run is narrowed to all requested rows individually.
  # The already compiled switched binary avoids compilation ambiguity on reruns.
  for row in rows:
   start=time.monotonic()
   with (out/(mid+'.'+row+'.log')).open('w') as f:rr=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','^'+row+'$'],cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
   print(mid,row,rr.returncode,flush=True)
# Leave switch available for survivor witnesses; root restores after collecting witnesses.
