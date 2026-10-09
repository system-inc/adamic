import pathlib,json,subprocess,difflib,time,os
P=pathlib.Path('review/test-audit/internal-lower-arguments_length')
mutants=[]
def add(id,file,old,new,switch,kind):
 file='internal/lower/'+file+'.go'; source=pathlib.Path(file).read_text(); assert source.count(old)==1,(id,source.count(old))
 mutants.append(dict(id=id,file=file,line=source[:source.index(old)].count('\n')+1,old=old,new=new,switch=switch,kind=kind))
add('M01','arguments_length','if !ast.IsAssignmentTarget(length) {','if ast.IsAssignmentTarget(length) {','if auditBool("M01", !ast.IsAssignmentTarget(length)) {','flip condition')
add('M02','arguments_length','if owner.Kind == ast.KindArrowFunction {','if owner.Kind != ast.KindArrowFunction {','if auditBool("M02", owner.Kind == ast.KindArrowFunction) {','flip condition')
add('M03','arguments_length','access.Name().Text() == "length"','access.Name().Text() == "size"','access.Name().Text() == auditString("M03", "length", "size")','change constant')
add('M04','arguments_length','l.function.ReadsArguments = true','l.function.ReadsArguments = false','l.function.ReadsArguments = auditBool("M04", true)','change constant')
add('M05','library_array_predicate','if proven.Flags()&checker.TypeFlagsUnknown != 0 {\n\t\treturn true','if proven.Flags()&checker.TypeFlagsUnknown != 0 {\n\t\treturn false','if proven.Flags()&checker.TypeFlagsUnknown != 0 {\n\t\treturn auditBool("M05", true)','change constant')
add('M06','library_array_predicate','if !l.arrayPredicateDomain(l.checker.GetTypeAtLocation(argument)) && !l.exactObject(argument, 0) {','if l.arrayPredicateDomain(l.checker.GetTypeAtLocation(argument)) || l.exactObject(argument, 0) {','if auditBool("M06", !l.arrayPredicateDomain(l.checker.GetTypeAtLocation(argument)) && !l.exactObject(argument, 0)) {','flip condition')
add('M07','library_array_predicate','predicateReadonlyArray(l.checker, member, true)','predicateReadonlyArray(l.checker, member, false)','predicateReadonlyArray(l.checker, member, auditBool("M07", true))','change option')
add('M08','predicates','return p.reference(ast.SkipParentheses(node.AsCallExpression().Arguments.Nodes[0])) && (p.l.arrayPredicateDomain(p.declared) || p.declared.Flags()&checker.TypeFlagsUnknown != 0)','return false','if auditMutant("M08") { return false }; return p.reference(ast.SkipParentheses(node.AsCallExpression().Arguments.Nodes[0])) && (p.l.arrayPredicateDomain(p.declared) || p.declared.Flags()&checker.TypeFlagsUnknown != 0)','return early')
add('M09','cast_proof','if !l.castUnionWrites(source) {','if l.castUnionWrites(source) {','if auditBool("M09", !l.castUnionWrites(source)) {','flip condition')
add('M10','cast','ir.NumberConstant{Value: value.Float()}','ir.NumberConstant{Value: value.Float() + 1}','ir.NumberConstant{Value: auditNumber("M10", value.Float())}','change constant')
add('M11','cast','return cast, nil','return value, nil','if auditMutant("M11") { return value, nil }; return cast, nil','return early')
add('M12','cast_proof','if declaration == nil || len(declaration.TypeParameters()) == 0 {','if declaration == nil || len(declaration.TypeParameters()) != 0 {','if auditBool("M12", declaration == nil || len(declaration.TypeParameters()) == 0) {','flip condition')
add('M13','invariance','if !l.checker.IsTypeAssignableTo(source, target) {','if l.checker.IsTypeAssignableTo(source, target) {','if auditBool("M13", !l.checker.IsTypeAssignableTo(source, target)) {','flip condition')
add('M14','invariance','if !l.enumAssignable(given, takes) || !l.checker.IsTypeAssignableTo(given, takes) {','if l.enumAssignable(given, takes) || l.checker.IsTypeAssignableTo(given, takes) {','if auditBool("M14", !l.enumAssignable(given, takes) || !l.checker.IsTypeAssignableTo(given, takes)) {','flip condition')
add('M15','census_small','return l.classAssignable(from, to) && l.widened(from, to, map[[2]*checker.Type]bool{}) == nil','return true','if auditMutant("M15") { return true }; return l.classAssignable(from, to) && l.widened(from, to, map[[2]*checker.Type]bool{}) == nil','return early')
add('M16','expression','if declaration.Kind == ast.KindFunctionDeclaration && declaration.Body() == nil && l.censusImplementation(declaration) != nil {','if declaration.Kind == ast.KindFunctionDeclaration && declaration.Body() != nil && l.censusImplementation(declaration) != nil {','if declaration.Kind == ast.KindFunctionDeclaration && auditBool("M16", declaration.Body() == nil) && l.censusImplementation(declaration) != nil {','flip condition')
add('M17','expression','if len(signatures) == 1 && l.censusNeverRestSignature(signatures[0]) {','if len(signatures) == 1 && !l.censusNeverRestSignature(signatures[0]) {','if len(signatures) == 1 && auditBool("M17", l.censusNeverRestSignature(signatures[0])) {','flip condition')
(P/'mutants.json').write_text(json.dumps(mutants,indent=2)+'\n')
originals={m['file']:pathlib.Path(m['file']).read_text() for m in mutants}
originals['internal/lower/lower.go']=pathlib.Path('internal/lower/lower.go').read_text()
(P/'originals.json').write_text(json.dumps(originals))
(P/'diffs').mkdir(exist_ok=True)
for m in mutants:
 diff=''.join(difflib.unified_diff(originals[m['file']].splitlines(True), originals[m['file']].replace(m['old'],m['new']).splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file']))
 (P/'diffs'/(m['id']+'.diff')).write_text(diff)
# Menu and locations are fixed here, before any mutant run.
(P/'menu-fixed.txt').write_text(time.strftime('%Y-%m-%dT%H:%M:%SZ',time.gmtime())+'\n')
def run(cmd,log,env=None):
 start=time.monotonic()
 with (P/'logs'/log).open('w') as out: r=subprocess.run(cmd,stdout=out,stderr=subprocess.STDOUT,env=env)
 return dict(command=' '.join(cmd),exit=r.returncode,wall=time.monotonic()-start,log=log)
validation={}
try:
 for m in mutants:
  assert subprocess.run(['git','apply','--check',str(P/'diffs'/(m['id']+'.diff'))],capture_output=True).returncode==0
  pathlib.Path(m['file']).write_text(originals[m['file']].replace(m['old'],m['new']))
  validation[m['id']]=run(['timeout','90','go','vet','./internal/lower/'],m['id']+'-vet.log')
  pathlib.Path(m['file']).write_text(originals[m['file']])
  if validation[m['id']]['exit']!=0: raise RuntimeError('vet failed '+m['id'])
finally:
 for file,s in originals.items(): pathlib.Path(file).write_text(s)
 (P/'validation.json').write_text(json.dumps(validation,indent=2)+'\n')
for file,s in originals.items():
 for m in mutants:
  if m['file']==file: s=s.replace(m['old'],m['switch'])
 if file.endswith('/lower.go'): s=s.replace('files := program.Files()', 'if auditMutant("P01") { return nil, nil }; files := program.Files()',1)
 pathlib.Path(file).write_text(s)
helper=pathlib.Path('internal/lower/audit_u026.go')
helper.write_text('''package lower
import "os"
func auditMutant(id string) bool { return os.Getenv("ADAMIC_MUTANT") == id }
func auditBool(id string, value bool) bool { if auditMutant(id) { return !value }; return value }
func auditString(id, value, changed string) string { if auditMutant(id) { return changed }; return value }
func auditNumber(id string, value float64) float64 { if auditMutant(id) { return value+1 }; return value }
''')
subprocess.run(['gofmt','-w',*originals.keys(),str(helper)],check=True)
(P/'switch.diff').write_bytes(subprocess.check_output(['git','diff','--',*originals.keys()]))
(P/'switch-helper.go.txt').write_text(helper.read_text())
results={}; rows=json.loads((P/'rows.json').read_text()); regex=(P/'rows.regex').read_text()
try:
 for m in mutants:
  env=dict(os.environ,ADAMIC_MUTANT=m['id'],ADAMIC_BUILD_CACHE_DIR='/tmp/u026/cache/'+m['id'])
  results[m['id']]=run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.'],m['id']+'.log',env)
  contents=(P/'logs'/(m['id']+'.log')).read_text()
  if 'panic:' in contents or results[m['id']]['exit']==124:
   results[m['id']]['individual']={}
   for row in rows:
    results[m['id']]['individual'][row['test']]=run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','^'+row['test']+'$'],m['id']+'-'+row['test']+'.log',env)
  (P/'matrix-runs.json').write_text(json.dumps(results,indent=2)+'\n')
  print(m['id'],results[m['id']]['exit'],round(results[m['id']]['wall'],3),flush=True)
 probes={}
 for row in rows:
  env=dict(os.environ,ADAMIC_MUTANT='P01',ADAMIC_BUILD_CACHE_DIR='/tmp/u026/cache/P01')
  probes[row['test']]=run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','^'+row['test']+'$'],'P01-'+row['test']+'.log',env)
 (P/'probe-runs.json').write_text(json.dumps(probes,indent=2)+'\n')
finally:
 for file,s in originals.items(): pathlib.Path(file).write_text(s)
 helper.unlink()
