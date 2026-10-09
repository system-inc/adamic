import pathlib,subprocess,json,time,os,difflib,re
r=pathlib.Path('/workspace/adamic');e=r/'review/test-audit/internal-lower-prototype';fs=['prototype.go','object.go','expression.go','proven_relations.go','non_null.go','refusals.go','regexp.go','lower.go'];bases={f:(r/'internal/lower'/f).read_text() for f in fs};menu=[]
def add(f,old,new,kind,change):
 assert bases[f].count(old)==1,(f,old,bases[f].count(old));mid='M%02d'%(len(menu)+1);off=bases[f].index(old);menu.append(dict(id=mid,file='internal/lower/'+f,line=bases[f][:off].count('\n')+1,old=old,new=new,kind=kind,change=change))
add('prototype.go','if load.IsLibrary(ast.GetSourceFileOfNode(declaration)) {','if !load.IsLibrary(ast.GetSourceFileOfNode(declaration)) {','flip condition','invert inherited-library declaration recognition')
add('prototype.go','if name == "isPrototypeOf" {\n\t\treturn &Refused','if name == "constructor" {\n\t\treturn &Refused','change constant','select constructor instead of isPrototypeOf diagnostic')
add('prototype.go','if name == "isPrototypeOf" {\n\t\treturn nil, true, &Refused','if name != "isPrototypeOf" {\n\t\treturn nil, true, &Refused','flip condition','invert prototype-chain call refusal')
add('prototype.go','return reason\n}', 'return ""\n}', 'change constant','discard whole-program hazard result')
add('prototype.go','} else if result != of {','} else if result == of {','flip condition','invert valueOf representation compatibility')
add('prototype.go','JavaScript uses locale-sensitive number formatting; use toString for deterministic formatting','','change constant','clear locale-sensitive number repair')
add('object.go','if l.inheritedLibraryMember(node) && !l.regexRuntimeProperty(access.Expression, name)', 'if !l.inheritedLibraryMember(node) && !l.regexRuntimeProperty(access.Expression, name)','flip condition','invert property inherited guard')
add('object.go','if l.inheritedLibraryMember(node) && !l.regexRuntimeProperty(access.Expression, index.Text())', 'if !l.inheritedLibraryMember(node) && !l.regexRuntimeProperty(access.Expression, index.Text())','flip condition','invert element inherited guard')
add('expression.go','"RegExpIndicesArray")):\n\t\treturn ir.Array, true','"RegExpIndicesArray")):\n\t\treturn ir.Array, false','change constant','mark array representation unknown')
add('expression.go','// An object with call signatures is a function, held as a closure.\n\t\treturn ir.Closure, true','// An object with call signatures is a function, held as a closure.\n\t\treturn ir.Closure, false','change constant','mark callable representation unknown')
add('proven_relations.go','field != "" {\n\t\treturn refuse("optional field "+field','field == "" {\n\t\treturn refuse("optional field "+field','flip condition','invert optional relation failure')
add('proven_relations.go','\tif mismatch := l.nominalMismatch(source, target, map[[2]*checker.Type]bool{}); mismatch != nil {\n\t\treturn refuse("nominal ancestry for "+l.checker.TypeToString(mismatch), "construct that class or a subclass; use an interface for structural values (adamic/nominal-class)")\n\t}\n','', 'drop statement','drop nominal relation refusal block')
add('non_null.go','counts.Checked++','counts.Checked += 2','change constant','double eager assertion count')
add('refusals.go', 'write ?? panic(\'why it can\'t be missing\'), or narrow and handle the missing case','','change constant','clear non-null repair')
add('refusals.go','if len(module.CommentDirectives) > 0 {','if len(module.CommentDirectives) > 1 {','off-by-one bound','require two suppression directives')
add('refusals.go','What: name + " suppression directive", Fix: "remove it and fix the type error"','What: name + " suppression directive", Fix: ""','change constant','clear suppression repair')
add('regexp.go','if len(args) > 2 {','if len(args) > 0 {','change constant','reject any RegExp constructor argument')
add('regexp.go','return "(?:)"','return "()"','change constant','change empty RegExp source')
add('regexp.go',"if c == '/' && !escaped && depth == 0 {", "if c == '/' && !escaped && depth > 0 {",'flip condition','invert slash character-class condition')
add('regexp.go','escaped, sets := false, strings.Contains(flags, "v")','escaped, sets := false, strings.Contains(flags, "u")','change constant','use u for nested character sets')
(e/'menu.json').write_text(json.dumps(menu,indent=2));rows=json.loads((e/'requested-rows.json').read_text());cov=(e/'functions-coverage.txt').read_text();(e/'reached-functions.txt').write_text('\n'.join(l for l in cov.splitlines() if not re.search(r'\s0\.0%$',l) and not l.startswith('total:'))+'\n')
# Freeze standalone changes and entry probes before any matrix result.
probes=[('PLower','lower.go','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {','return nil, nil', [t for t in rows if t not in ['TestInheritedLibraryReadsNeverLoadOwnFields','TestNullishPrototypeReadsAreRejectedByChecker','TestRegExpSourceNode','TestRepresentationClockSourceCheckedTypes']]),('PProperty','object.go','func (l *lowering) property(node *ast.Node) (ir.Expression, error) {','return nil, nil',['TestInheritedLibraryReadsNeverLoadOwnFields']),('PElement','object.go','func (l *lowering) elementAccess(node *ast.Node) (ir.Expression, error) {','return nil, nil',['TestInheritedLibraryReadsNeverLoadOwnFields']),('PRepresentation','expression.go','func (l *lowering) representation(proven *checker.Type) (ir.Type, bool) {','return 0, false',['TestRepresentationClockSourceCheckedTypes']),('PRegexSource','regexp.go','func escapeRegexSource(pattern, flags string) string {','return ""',['TestRegExpSourceNode'])]
(e/'probes.json').write_text(json.dumps([dict(id=id,file='internal/lower/'+f,line=bases[f][:bases[f].index(sig)].count('\n')+1,signature=sig,answer=answer,rows=ts) for id,f,sig,answer,ts in probes],indent=2))
def diff(f,s):return ''.join(difflib.unified_diff(bases[f].splitlines(True),s.splitlines(True),fromfile='a/internal/lower/'+f,tofile='b/internal/lower/'+f))
for m in menu:
 f=m['file'].split('/')[-1];(e/(m['id']+'.diff')).write_text(diff(f,bases[f].replace(m['old'],m['new'],1)))
for id,f,sig,answer,ts in probes:(e/(id+'.diff')).write_text(diff(f,bases[f].replace(sig,sig+'\n\temptyAnswer := true\n\tif emptyAnswer { '+answer+' }\n',1)))
runs=[];env=os.environ.copy()
def run(cmd,log,id=None):
 ev=env.copy()
 if id:ev.update(ADAMIC_MUTANT=id,ADAMIC_BUILD_CACHE_DIR='/tmp/u043/cache/'+id)
 start=time.monotonic()
 with (e/log).open('w') as out:q=subprocess.run(cmd,cwd=r,env=ev,stdout=out,stderr=subprocess.STDOUT)
 runs.append(dict(command=cmd,log=log,exit=q.returncode,wall=time.monotonic()-start,environment={k:ev[k] for k in ['ADAMIC_MUTANT','ADAMIC_BUILD_CACHE_DIR'] if k in ev}));(e/'mutant-runs.json').write_text(json.dumps(runs,indent=2));print(log,q.returncode,round(runs[-1]['wall'],3),flush=True);return q.returncode
try:
 for m in menu:
  f=m['file'].split('/')[-1];(r/m['file']).write_text(bases[f].replace(m['old'],m['new'],1));assert run(['timeout','90','go','vet','./internal/lower/'],m['id']+'-vet.log')==0;m['vet']=True;(r/m['file']).write_text(bases[f])
 # One switch build. Apply replacements from last position backwards per file.
 for f,base in bases.items():
  patches=[]
  for m in menu:
   if m['file'].endswith('/'+f):
    old,new=m['old'],m['new'];off=base.index(old)
    # Whole statement deletion is controlled by enclosing the original block.
    if m['id']=='M12':replacement='if auditMutant() != "'+m['id']+'" {\n'+old+'\n}\n' if old.startswith('\tif ') else 'auditChoice("'+m['id']+'", "", '+json.dumps(old)+')'
    elif m['id'] in ['M02','M03']:
     replacement=old.replace('name == "isPrototypeOf"', 'name == auditChoice("M02", "constructor", "isPrototypeOf")') if m['id']=='M02' else old.replace('name == "isPrototypeOf"','auditBool("M03", name == "isPrototypeOf")')
    elif m['id'] in ['M04','M06','M14','M16','M18','M20']:
     if m['id'] in ['M06','M14']:replacement='"+auditChoice("'+m['id']+'", "", '+json.dumps(old)+')+"'
     elif m['id']=='M04':replacement='if auditMutant() == "M04" { return "" }; '+old
     elif m['id']=='M16':replacement='What: name + " suppression directive", Fix: auditChoice("M16", "", "remove it and fix the type error")'
     elif m['id']=='M18':replacement='return auditChoice("M18", "()", "(?:)")'
     else:replacement='escaped, sets := false, strings.Contains(flags, auditChoice("M20", "u", "v"))'
    elif m['id'] in ['M09','M10']:
     replacement=old.replace('return ir.Array, true','return ir.Array, auditMutant() != "M09"').replace('return ir.Closure, true','return ir.Closure, auditMutant() != "M10"')
    elif m['id']=='M13':replacement='counts.Checked += auditInt("M13", 2, 1)'
    elif m['id']=='M15':replacement='if len(module.CommentDirectives) > auditInt("M15", 1, 0) {'
    elif m['id']=='M17':replacement='if len(args) > auditInt("M17", 0, 2) {'
    else:
     # Wrap the changed boolean expression in auditBool; targeted substrings here include surrounding syntax.
     common=0
     while common<min(len(old),len(new)) and old[common]==new[common]:common+=1
     if m['id']=='M01':replacement='if auditBool("M01", load.IsLibrary(ast.GetSourceFileOfNode(declaration))) {'
     elif m['id']=='M05':replacement='} else if auditBool("M05", result != of) {'
     elif m['id'] in ['M07','M08']:replacement=old.replace('l.inheritedLibraryMember(node)','auditBool("'+m['id']+'", l.inheritedLibraryMember(node))')
     elif m['id']=='M11':replacement=old.replace('field != ""','auditBool("M11", field != "")')
     elif m['id']=='M19':replacement=old.replace('depth == 0','auditBool("M19", depth == 0)')
     else:raise Exception(m['id'])
    patches.append((off,len(old),replacement))
  for id,pf,sig,answer,ts in probes:
   if pf==f:patches.append((base.index(sig)+len(sig),0,'\n if auditMutant() == "'+id+'" { '+answer+' }\n'))
  for off,n,s in sorted(patches,reverse=True):base=base[:off]+s+base[off+n:]
  (r/'internal/lower'/f).write_text(base);(e/(f+'.switch.txt')).write_text(base)
 helper=r/'internal/lower/u043_mutant.go';helper.write_text('package lower\nimport "os"\nfunc auditMutant() string{return os.Getenv("ADAMIC_MUTANT")}\nfunc auditBool(id string,b bool)bool{if auditMutant()==id{return !b};return b}\nfunc auditChoice(id,a,b string)string{if auditMutant()==id{return a};return b}\nfunc auditInt(id string,a,b int)int{if auditMutant()==id{return a};return b}\n');(e/'switch-helper.go.txt').write_text(helper.read_text())
 assert run(['gofmt','-w']+['internal/lower/'+f for f in fs]+['internal/lower/u043_mutant.go'],'switch-gofmt.log')==0
 assert run(['timeout','90','go','test','-c','-o','/tmp/u043-test','./internal/lower/'],'switch-build.log')==0
 for m in menu:
  id=m['id'];run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.'],id+'.log',id)
  text=(e/(id+'.log')).read_text()
  if 'panic:' in text or 'test timed out' in text:
   for t in rows:run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','^'+t+'$'],id+'-alone-'+t+'.log',id)
 for id,f,sig,answer,ts in probes:
  for t in ts:run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','^'+t+'$'],id+'-'+t+'.log',id)
finally:
 for f,s in bases.items():(r/'internal/lower'/f).write_text(s)
 helper=r/'internal/lower/u043_mutant.go'
 if helper.exists():helper.unlink()
