from pathlib import Path
import subprocess,json,time
root=Path.cwd(); out=root/'review/compiler/agree-modules'; results=[]
def run(label,selector):
 cmd=['go','test','./internal/lower','-run',selector,'-count=1','-timeout','90s','-v']; start=time.monotonic()
 with (out/(label+'.log')).open('w') as log:r=subprocess.run(cmd,stdout=log,stderr=log,timeout=110)
 text=(out/(label+'.log')).read_text(); result=dict(label=label,command=cmd,exit=r.returncode,seconds=round(time.monotonic()-start,3),failures=[line for line in text.splitlines() if '--- FAIL:' in line]); results.append(result);print(result,flush=True);return r.returncode
for label,path,old,new,testfile,test in [
 ('enum-value','internal/lower/enums.go','return ir.NumberConstant{Value: value}, nil','return ir.NumberConstant{Value: value + 1}, nil','enums_test.go','TestConstEnumErasesRuntimeObject'),
 ('nested-rest','internal/lower/functions.go','function.RestElement = element','function.RestElement = 0','nested_functions_test.go','TestNestedRestIsSupported')]:
 p=root/path; saved=p.read_text(); changed=saved.replace(old,new,1);assert changed!=saved
 import difflib
 (out/(label+'.diff')).write_text(''.join(difflib.unified_diff(saved.splitlines(True),changed.splitlines(True),fromfile='a/'+path,tofile='b/'+path)))
 t=root/'internal/lower'/testfile; converted=t.read_text()
 try:
  p.write_text(changed)
  t.write_text(subprocess.check_output(['git','show','571e74cf555b9db994c5dec6c2f8dbee676e5111:internal/lower/'+testfile],text=True))
  assert run(label+'-old','^'+test+'$')==0
  t.write_text(converted)
  assert run(label+'-converted','^'+test+'$')==1
 finally:p.write_text(saved);t.write_text(converted)
# Rerun every local-copy guard mutant that the converted guard files held.
for directory,label,selector in [
 ('enum-guards','M01','^TestEnumMemberValuesAndReverseNameMatchNode$'),
 ('enum-guards','M02','^TestEnumMemberValuesAndReverseNameMatchNode$'),
 ('enum-guards','M04','^TestEnumReverseMappingUsesSingleSlot$'),
 ('enum-guards','M14','^TestEnumFlagProofsWithoutObservableLoweringEffect$'),
 ('enum-guards','M15','^TestEnumFlagProofsWithoutObservableLoweringEffect$'),
 ('enum-guards','M16','^TestEnumFlagProofsWithoutObservableLoweringEffect$'),
 ('statics-guards','M06','^TestPrivateAndPublicStaticsAgreeWithNode$'),
 ('statics-guards','M12','^Test(NamespaceFactoryBindingsAgreeWithNode|ParserFactoryBindingHoisting)$'),
 ('statics-guards','M13','^TestReadinessErrorIncludesReceiverExpression$')]:
 patch=root/'review/compiler'/directory/(label+'.diff')
 subprocess.run(['git','apply','--check',str(patch)],check=True,timeout=10)
 subprocess.run(['git','apply',str(patch)],check=True,timeout=10)
 try:assert run(directory+'-'+label,selector)==1
 finally:subprocess.run(['git','apply','-R',str(patch)],check=True,timeout=10)
(out/'mutants.json').write_text(json.dumps(results,indent=2)+'\n')
