from pathlib import Path
import difflib,json,subprocess
root=Path(__file__).resolve().parents[3];out=Path(__file__).resolve().parent
cases=[
 ('receiver','internal/lower/readiness.go','return strconv.Itoa(receiverTypeID) + ":" + name','return strconv.Itoa(0) + ":" + name',['./internal/lower'],'^TestFX6P37$','JavaScript backend stdout'),
 ('destructure','internal/lower/collections.go','value.ViewTypeID = int(fieldType.Id())','value.ViewTypeID = 0',['./internal/lower'],'^TestFX6P53$','JavaScript backend stdout'),
 ('scalar-check','internal/native/view_unions_read.go','e.checkMaybeUnionView(property, value, snapshot)','// Reverted conversion guard.',['./internal/oracle'],'^TestScalarUnionViewMisfit','native'),
 ('nullable','internal/lower/interface_cast.go','return int(l.checker.GetNonNullableType(receiver).Id())','return int(receiver.Id())',['./internal/oracle'],'TestCheckedViewOptionalReadBoundary|TestCheckedViewObjectPrimitiveSource/optional-receiver','optional checked read escaped'),
 ('generic','internal/lower/interface_cast.go','receiver = l.concrete(receiver)\n\tif receiver.Flags()&checker.TypeFlagsTypeParameter != 0 {\n\t\tif constraint := l.checker.GetBaseConstraintOfType(receiver); constraint != nil {\n\t\t\treceiver = l.concrete(constraint)\n\t\t}\n\t}', '// Leave the receiver binder rigid.',['./internal/oracle'],'TestCheckedViewUntaggedSourceFlows/generic/wrong','exit codes differ'),
 ('union','internal/lower/interface_cast.go','if receiver.Flags()&checker.TypeFlagsUnion != 0 {','if false && receiver.Flags()&checker.TypeFlagsUnion != 0 {',['./stage3/fixtures'],'TestFixturesAssertions/assertions/19_identifier_kind.a/stage0','gap changed')]
results=[]
for name,file,before,after,packages,selector,catcher in cases:
 path=root/file;original=path.read_text();assert original.count(before)==1,(name,original.count(before));changed=original.replace(before,after,1)
 (out/(name+'.patch')).write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
 try:
  path.write_text(changed)
  cmd=['go','test',*packages,'-run',selector,'-count=1','-v','-timeout','90s']
  with (out/(name+'-mutant.log')).open('w') as log: result=subprocess.run(cmd,cwd=root,stdout=log,stderr=subprocess.STDOUT,timeout=120)
  text=(out/(name+'-mutant.log')).read_text();caught=result.returncode!=0 and catcher in text and 'build failed' not in text
  results.append(dict(mutant=name,command=cmd,exit=result.returncode,caught=caught,catcher=catcher));(out/'mutants.json').write_text(json.dumps(results,indent=2)+'\n')
  assert caught,text
 finally:path.write_text(original)
