import os,json,time,subprocess,difflib,pathlib
root=pathlib.Path('/workspace/adamic');out=pathlib.Path(__file__).resolve().parent;scratch=pathlib.Path('/tmp/defend-enums')
ref='origin/test-audit/internal-oracle-enums_open';ap='review/test-audit/internal-oracle-enums_open/'
for f in ['PLAN.md','REPORT.md','rows.json','scope.json','matrix.json']:
 (out/('audit-'+f)).write_bytes(subprocess.check_output(['git','show',ref+':'+ap+f],cwd=root))
rows=['TestNumericEnumNeverPathsPinned','TestNumericEnumNeverPinned','TestStage3EnumBoundaries','TestStage3EnumSparseArrayBoundary','TestEnumNameEnumerationMutant','TestEnumSemanticMutants','TestEnumCleanupMutant','TestFallthroughMutants','TestFieldReadinessRepresentation','TestFreshWriteProbesStayRefused','TestFunctionValueBoundaryBoxing','TestImportCycleRuntimeCalls','TestImportCycleLoadTimeReads','TestImportedNonliteralConstCaseIsNotYet','TestInputAgreesWithNode','TestEnumInitializationNode','TestEnumInitializationUnknownPinned','TestModuleNamespaceReadsMatchNode','TestModuleNamespaceReadinessMutants','TestModuleNamespaceLiveBindingMutant','TestNestedSiblingCycleMutantIsCaught','TestNestedCycleRefusal','TestStage3EnumFallthrough','TestTypedArrayWriteStopIsPinned']
pattern='^('+'|'.join(rows)+')$'
general='^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^(enums_open.*|import_cycles|write_after_shrink.a|writes_past_end.a|typed_arrays_stop.a)$'
variants=[('D1', [('internal/lower/fresh.go','return "the write at " +','return "the assignment at " +')]),('D2',[('internal/native/runtime/array.c','is outside an array of length','is beyond an array of length')]),('D3',[('internal/javascript/javascript.go','is outside an array of length','is beyond an array of length')]),('D4',[('internal/native/runtime/array.c','is outside an array of length','is beyond an array of length'),('internal/native/runtime/typed_array.c','is outside an array of length','is beyond an array of length'),('internal/javascript/javascript.go','is outside an array of length','is beyond an array of length')]),('D5',[('internal/lower/expression.go','ast.KindMinusToken:            ir.Subtract,','ast.KindMinusToken:            ir.Remainder,')]),('D6',[('internal/native/emit_expressions.go','return fmt.Sprintf("(%s %s %s)", left, cOperators[operator], right)','return fmt.Sprintf("(%s %s %s)", right, cOperators[operator], left)')]),('D7',[('internal/native/emit_branches.go','e.line("if (%s) {", unwrap(condition))','e.line("if (!(%s)) {", unwrap(condition))')])]
files={f for _,changes in variants for f,_,_ in changes};base={f:(root/f).read_text() for f in files}
(out/'base.txt').write_text(subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True))
(out/'plan.json').write_text(json.dumps({'rows':rows,'general_selector':general,'variants':variants},indent=2))
results=[]
def run(id,label,cmd):
 env=os.environ.copy();env['ADAMIC_GATE_UNCACHED']='1';env['ADAMIC_BUILD_CACHE_DIR']=str(scratch/'cache'/id)
 t=time.monotonic()
 with (out/(id+'-'+label+'.log')).open('w') as log:r=subprocess.run(cmd,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
 result={'id':id,'label':label,'command':cmd,'wall_seconds':time.monotonic()-t,'exit':r.returncode};results.append(result);(out/'commands.json').write_text(json.dumps(results,indent=2));print(result,flush=True)
try:
 for id,changes in [('clean',[])]+variants:
  for f,s in base.items():(root/f).write_text(s)
  diff=''
  for f,old,new in changes:
   s=(root/f).read_text();assert s.count(old)==1,(id,f,s.count(old));(root/f).write_text(s.replace(old,new))
  if changes:
   for f in files:
    n=(root/f).read_text()
    if n!=base[f]:diff+=''.join(difflib.unified_diff(base[f].splitlines(True),n.splitlines(True),fromfile='a/'+f,tofile='b/'+f))
   (out/(id+'.diff')).write_text(diff)
   run(id,'vet',['timeout','120','go','vet','./internal/lower/','./internal/native/','./internal/javascript/'])
  run(id,'matrix',['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',pattern])
  run(id,'general',['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',general])
finally:
 for f,s in base.items():(root/f).write_text(s)
