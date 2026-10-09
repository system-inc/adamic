from pathlib import Path
import subprocess,time,json,difflib,os
root=Path.cwd();p=root/'review/test-audit/cmd-adamic-test262-adapt';prefix='cmd/adamic-test262/';paths=[prefix+x for x in ['adapt.go','classify.go','run.go']];base={x:subprocess.check_output(['git','show','HEAD:'+x],text=True) for x in paths}
a,c,r=paths
sig='func adaptSource(source string) adapted {\n';csig='func classify(path string, source string, adapt bool) classified {\n';rsig='func withCrashPath(path string, reason string) string {\n'
changes=[
('M1',a,sig,sig+'\treturn adapted{Source: source}\n','return early',True),
('M2',a,'end: stmt.keyword.end, text: "let"','end: stmt.keyword.end, text: "const"','change constant',False),
('M3',a,'if decls != 1 {','if decls != 2 {','off-by-one bound',False),
('M4',a,'if ref.inInit == decl.binding {\n\t\t\treturn false','if ref.inInit == decl.binding {\n\t\t\treturn true','change constant',False),
('M5',a,'if decl.loopVar && decl.loop == scope {\n\t\t\treturn true','if decl.loopVar && decl.loop == scope {\n\t\t\treturn false','change constant',False),
('M6',a,'if inner == outer {','if inner != outer {','flip condition',False),
('M7',a,'text: ": " + rendered','text: " = " + rendered','change constant',False),
('M8',a,'func (p *parser) callbackFn(call callNote) *fnInfo {\n\tindex := 0','func (p *parser) callbackFn(call callNote) *fnInfo {\n\tindex := 1','off-by-one bound',False),
('M9',a,'return &jtype{kind: "array", elem: elem}','return &jtype{kind: "object", elem: elem}','change constant',False),
('M10',a,'func sameType(left *jtype, right *jtype) bool {\n\tif left == nil || right == nil || left.kind != right.kind {','func sameType(left *jtype, right *jtype) bool {\n\tif left == nil || right == nil || left.kind == right.kind {','flip condition',False),
('M11',a,'func strictEqualOK(left *jtype, right *jtype) bool {\n\tif left == nil || right == nil || left.kind != right.kind {','func strictEqualOK(left *jtype, right *jtype) bool {\n\tif left == nil || right == nil || left.kind == right.kind {','flip condition',False),
('M12',a,'text := "==="','text := "=="','change constant',False),
('M13',a,'end: throw.end, text: "Error"','end: throw.end, text: "TypeError"','change constant',False),
('M14',a,'&& !p.instanceOfTest262 {','&& p.instanceOfTest262 {','flip condition',False),
('M15',a,'if len(expr.elems) == 0 {','if len(expr.elems) <= 1 {','off-by-one bound',False),
('M16',a,"source[next] == '+' || source[next] == '-'","source[next] == '*' || source[next] == '-'",'change constant',False),
('M17',a,'p.skipBraced()\n\t\tp.unsafeVars = true','p.skipBraced()','drop statement',False),
('M18',c,'if adapt {','if !adapt {','flip condition',False),
('M19',c,'\t\tresult.Adaptations = rewritten.Counts\n','','drop statement',False),
('M20',r,'return path + ": " + reason','return path + " / " + reason','change constant',False),
('P1',a,sig,sig+'\treturn adapted{}\n','empty-answer probe',True),
('P2',c,csig,csig+'\treturn classified{}\n','empty-answer probe',True),
('P3',r,rsig,rsig+'\treturn ""\n','empty-answer probe',True)]
def function_end(s,start):
 # These entries terminate just before a known next declaration.
 name=s[start:s.index('\n',start)]
 marker='\ntype edit struct' if 'adaptSource' in name else ('\nfunc directoryOf' if 'classify(' in name else '\nfunc test262Commit')
 return s.index(marker,start)
def diff(path,before,after):return ''.join(difflib.unified_diff(before.splitlines(True),after.splitlines(True),fromfile='a/'+path,tofile='b/'+path))
(p/'diffs').mkdir(exist_ok=True);plan=[];vet=[]
for mid,path,old,new,kind,entry in changes:
 assert base[path].count(old)==1,(mid,base[path].count(old))
 line=base[path][:base[path].index(old)].count('\n')+1
 altered=base[path].replace(old,new,1)
 if entry:
  start=base[path].index(old);end=function_end(base[path],start);altered=base[path][:start]+new+'}\n'+base[path][end:]
 dest=p/'diffs'/(mid+'.diff');dest.write_text(diff(path,base[path],altered));plan.append({'id':mid,'file':path,'line':line,'old':old,'new':new,'kind':kind})
(p/'plan.json').write_text(json.dumps(plan,indent=2))
for item in plan:
 mid=item['id'];dest=p/'diffs'/(mid+'.diff');start=time.monotonic()
 try:
  with (p/(mid+'-vet.log')).open('w') as out:
   for cmd in [['git','apply','--check',str(dest)],['git','apply',str(dest)],['timeout','90','go','vet','./cmd/adamic-test262/']]:
    result=subprocess.run(cmd,stdout=out,stderr=subprocess.STDOUT)
    assert result.returncode==0,(mid,cmd,result.returncode)
 finally:subprocess.run(['git','restore','--source=HEAD','--',item['file']],check=True)
 vet.append({'id':mid,'seconds':time.monotonic()-start,'exit':0});(p/'vet-status.json').write_text(json.dumps(vet,indent=2))
# Whole conditional replacements preserve originals as the inactive branch.
sw=dict(base)
for mid,path,old,new,kind,entry in changes:
 if entry:
  sw[path]=sw[path].replace(old,old+'\tif auditMutant("'+mid+'") { '+new[len(old):].strip()+' }\n',1)
 else:
  # Each target is a statement/condition, use its enclosing function as the switch unit.
  original=base[path];at=original.index(old)
  starts=[m.start() for m in __import__('re').finditer(r'^func ',original,__import__('re').M) if m.start()<=at];start=starts[-1]
  nextmatch=__import__('re').search(r'^func |^type |^var ',original[at+len(old):],__import__('re').M)
  # Go AST-free exact function boundary: top-level closing brace begins at column zero.
  end=original.index('\n}',at)+2
  body=original[start:end];changed=body.replace(old,new,1)
  # Apply changes in an existing switch function by replacing exact target text with an expression-specific selector.
  # Instead duplicate the original whole function as a mutant helper and dispatch at entry.
  header=body[:body.index('{')+1];signature=header
  method=__import__('re').match(r'func (\([^)]*\) )?(\w+)\((.*?)\)(.*?)\{',header,__import__('re').S)
  name=method[2];helper='audit'+mid+name
  helperbody=changed.replace(' '+name+'(', ' '+helper+'(',1)
  args={'rewriteVars':'','declSafe':'decl','capturedAcrossLoop':'decl, ref','scopeWithin':'inner, outer','annotateCallbacks':'','callbackFn':'call','arrayOf':'elem','sameType':'left, right','strictEqualOK':'left, right','rewriteEquals':'','rewriteThrows':'','parseThrow':'','evalArray':'expr','endOfNumber':'source, index','parseStatement':'','classify':'path, source, adapt','withCrashPath':'path, reason'}[name]
  result_type=method[4].strip()
  call=('p.' if method[1] else '')+helper+'('+args+')'
  dispatch='\n\tif auditMutant("'+mid+'") { '+('return '+call if result_type else call+'; return')+' }'
  sw[path]=sw[path].replace(signature,signature+dispatch,1)
  sw[path]+='\n\n'+helperbody+'\n'
helperpath=prefix+'audit_selector.go';helper='package main\nimport "os"\nfunc auditMutant(id string) bool { return os.Getenv("ADAMIC_MUTANT") == id }\n'
patch=''.join(diff(path,base[path],sw[path]) for path in paths)+diff(helperpath,'',helper)
(p/'selector.diff').write_text(patch);subprocess.run(['git','apply',str(p/'selector.diff')],check=True)
subprocess.run(['gofmt','-w',*paths,helperpath],check=True)
(p/'selector.diff').write_text(''.join(diff(path,base[path],(root/path).read_text()) for path in paths)+diff(helperpath,'',(root/helperpath).read_text()))
statuses=[]
try:
 for mid in ['control']+[x[0] for x in changes]:
  env=dict(os.environ);env.pop('ADAMIC_MUTANT',None)
  if mid!='control':env['ADAMIC_MUTANT']=mid
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./cmd/adamic-test262/','-run','.'];start=time.monotonic()
  with (p/(mid+'.log')).open('w') as out:res=subprocess.run(cmd,env=env,stdout=out,stderr=subprocess.STDOUT)
  status={'id':mid,'exit':res.returncode,'seconds':time.monotonic()-start,'command':cmd};statuses.append(status);(p/'matrix-status.json').write_text(json.dumps(statuses,indent=2));print(status,flush=True)
  if mid=='control' and res.returncode:raise RuntimeError('inactive selector is red')
  if mid in ['control','M15','M16','M17']:
   witness=p/'witness-source';witness.mkdir(exist_ok=True)
   for path in [a,prefix+'rewrite.go',helperpath]:__import__('shutil').copy(root/path,witness/Path(path).name)
   (witness/'main.go').write_text('package main\nimport("encoding/json";"fmt")\nfunc main(){ cases:=[]string{"function cb(x){return x;} [1].forEach(cb);", "var x=1e+2;", "var x=1; class Box { value: number; }"}; out:=map[string]adapted{};for _,s:=range cases {out[s]=adaptSource(s)};b,_:=json.MarshalIndent(out,"","  ");fmt.Println(string(b))}\n')
   with (p/('witness-'+mid+'.json')).open('w') as out, (p/('witness-'+mid+'.log')).open('w') as err:subprocess.run(['go','run',*[str(f) for f in witness.glob('*.go')]],env=env,stdout=out,stderr=err,check=True)
finally:
 subprocess.run(['git','restore','--source=HEAD','--',*paths],check=True)
 (root/helperpath).unlink(missing_ok=True)
