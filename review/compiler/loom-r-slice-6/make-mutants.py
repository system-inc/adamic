import ast,json
from pathlib import Path
root=Path.cwd();catalog=[]
def add(name,changes,selection,catcher,package='internal/oracle',all=False):
 catalog.append(dict(name=name,changes=changes,selection=selection,catcher=catcher,package=package,replace_all=all))
source=root/'review/compiler/chain-slice-6'
for family in ['run-mutants','run-library-mutants','run-subclass-mutants','run-adopted-mutants','run-main-mutants']:
 tree=ast.parse((source/(family+'.py')).read_text())
 rows=next(ast.literal_eval(n.value) for n in tree.body if isinstance(n,ast.Assign) and any(isinstance(t,ast.Name) and t.id=='mutants' for t in n.targets))
 for row in rows:
  if family=='run-subclass-mutants':name,path,old,new,catcher=row;selection='TestNativeAgreesWithNode/internal/oracle/testdata/step21_error_subclasses[.]a$'
  else:
   name,path,old,new,selection,catcher=row
   if family=='run-library-mutants':selection='TestNativeAgreesWithNode/internal/oracle/testdata/step21_'+selection+'[.]a$'
  if name=='error-message-not-retained':
   old='error->slots[1].reference = adamic_retain(message == NULL ? &empty_message : message);';new='error->slots[1].reference = message == NULL ? &empty_message : message;'
  if name=='soundness-check-is-catchable':selection='^TestStep21NarrowedParameterUseStops$';catcher='per-use stop'
  if name=='uncaught-is-panic':catcher='uncaught diagnostic'
  if name=='implicit-string-conversion':catcher='uncaught diagnostic'
  add(name,[(path,old,new)],selection,catcher,all=family in ['run-library-mutants','run-subclass-mutants'])
# Earlier saved-identity controls remain; their old whole-type/renderer controls
# were superseded by the final per-use campaign below.
tree=ast.parse((source/'run-saved-mutants.py').read_text());rows=next(ast.literal_eval(n.value) for n in tree.body if isinstance(n,ast.Assign) and any(isinstance(t,ast.Name) and t.id=='mutants' for t in n.targets))
for name,path,old,new,selection,catcher in rows:
 if name in ['rethrow-wrapped','saved-error-wrapped']:add(name,[(path,old,new)],selection,catcher)
 else:
  if name=='builtin-narrow-forgets-subtype':add(name,[(path,old,new)],'^TestStep21NarrowedParameterUseStops$','per-use stop')
# Final refinement's campaign constructs one anchor from emitted JS.
ns={};s=(source/'use-checks/run-mutants.py').read_text();s=s[:s.index('results=json.loads')];exec(s,ns)
for name,path,old,new,pkg,selection,catcher in ns['mutants']:
 if name=='native-scalar-properties':catcher='stale narrowing use failed'
 if name=='whole-narrowed-type':
  old='if l.narrowingPropertyReceiver(node) || l.narrowingWiderCall(node) {';new='if l.narrowingWiderCall(node) {'
 add(name,[(path,old,new)],selection,catcher,package=pkg.removeprefix('./'))
# Inheritance's three independent refusal/representation/empty-answer controls.
tree=ast.parse((source/'run-inherit-mutants.py').read_text());rows=next(ast.literal_eval(n.value) for n in tree.body if isinstance(n,ast.Assign) and any(isinstance(t,ast.Name) and t.id=='mutants' for t in n.targets))
for name,path,old,new,selection,catcher in rows:add(name,[(path,old,new)],selection,catcher,package='internal/lower')
add('readiness-old-effects',[('internal/lower/readiness.go','defer readinessExceptions(program)','// mutant: omit readiness exception propagation')],'TestNativeAgreesWithNode/internal/oracle/testdata/lowering_chain_tdz_catch[.]a$','stdout differs')
changes=[]
for relative in ['internal/lower/cast_proof.go','internal/lower/cast.go']:
 anchor='as := node.AsAsExpression()'
 if relative.endswith('cast_proof.go'):new=anchor+'; if l.checker.GetTypeAtLocation(as.Expression).Flags() & checker.TypeFlagsUnknown != 0 { return castProof{}, nil }'
 else:new=anchor+'; if l.checker.GetTypeAtLocation(as.Expression).Flags() & checker.TypeFlagsUnknown != 0 { value, err := l.expression(as.Expression); if err != nil { return nil, err }; return ir.Narrow{Value: value, To: ir.Object}, nil }'
 changes.append((relative,anchor,new))
add('unknown-authorizes-assertion',changes,'^TestStep21UnknownAssertions$','want Refused, got <nil>')
header='internal/native/runtime/adamic.h';h=Path(header).read_text();line=next(l for l in h.splitlines() if l.startswith('#define ADAMIC_FEATURE_NAME_PARTS('));ref=next(l for l in h.splitlines() if l.startswith('static const void *const adamic_runtime_feature_reference '))
for name,old,new in [('fixed-symbol',line,'#define ADAMIC_FEATURE_NAME_PARTS(a, b, c, d, e) adamic_runtime_features'),('drop-reference',ref,'/* mutant: no unit reference */'),('drop-retain','__attribute__((used, retain))','__attribute__((used))')]:add(name,[(header,old,new)],'^TestRuntimeFeatureMismatchProgramFeature$','mismatched runtime linked successfully',package='internal/native')
old='\t// Keep header bytes unchanged.'
new=''' for _, feature := range runtimeFeatures {
 if slicesContain(flags, "-D"+feature+"=1") {for i:=range files {if files[i].name=="adamic.h" {files[i].contents=append([]byte("#define "+feature+" 1\\n"),files[i].contents...)}}}
 }
 // Keep header bytes unchanged.'''
add('header-forces-runtime-features',[('internal/native/library.go',old,new)],'^TestRuntimeFeatureMismatchRuntimeFeature$','mismatched runtime linked successfully',package='internal/native')
add('split-runtime-default-flags',[('internal/native/units_tsgo.go','flags := append(sourceFlags(source, options), "-DADAMIC_TSGO")','flags := append(Flags(options), "-DADAMIC_TSGO")')],'^TestSplitTSGoRuntimeFeatures$','undefined reference',package='internal/native')
new='''library, err := native.RuntimeLibrary("",options)
 if err != nil {return err}
 path := filepath.Join(filepath.Dir(output),"main.c")
 if err:=os.WriteFile(path,[]byte(source),0644);err!=nil{return err}
 flags:=append(native.SourceFlags(source,options),"-I",filepath.Dir(library),"-o",output,path)
 flags=append(flags,native.RuntimeLinkFlags(library)...);flags=append(flags,"-lm")
 if output,err:=toolchain.Command("clang",flags...).CombinedOutput();err!=nil{return fmt.Errorf("profile clang: %w\\n%s",err,output)}
 return nil'''
add('lint-profile-default-runtime',[('stage1/cohere/lint/profile_compilation_main_test.go','return native.Build(source, output, options)',new)],'^TestProfileRuntimeFeaturesMatchNode$','undefined reference',package='stage1/cohere/lint')
# New main interaction: pointer-null detection loses an undefined winner even
# when another worker supplies a non-null exception later.
add('parallel-undefined-winner-lost',[('internal/native/runtime/parallel.c','if (state.exception_index != SIZE_MAX) {','if (state.exception != NULL) {')],'^TestParallelUndefinedExceptionSanitized$','signal: aborted',package='internal/native')
Path('review/compiler/loom-r-slice-6/mutant-catalog.json.txt').write_text(json.dumps(catalog,indent=2)+'\n')
# Validate anchors without running or classifying a test yet.
issues=[]
for m in catalog:
 for f,old,new in m['changes']:
  count=Path(f).read_text().count(old)
  if not count or count!=1 and not m['replace_all']:issues.append((m['name'],f,count))
print('Catalog',len(catalog),'mutants; anchor issues:',issues)
