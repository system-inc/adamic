from pathlib import Path
import difflib
import subprocess

evidence=Path('review/compiler/fx7-callable-producers')
def mutant(name, changes, test, wanted):
 originals={filename:Path(filename).read_text() for filename in changes}
 patch=''
 try:
  for filename, (before,after) in changes.items():
   original=originals[filename]
   assert original.count(before)==1,(name,filename)
   mutated=original.replace(before,after,1)
   patch+=''.join(difflib.unified_diff(original.splitlines(True),mutated.splitlines(True),fromfile='a/'+filename,tofile='b/'+filename))
   Path(filename).write_text(mutated)
  (evidence/(name+'.diff')).write_text(patch)
  with (evidence/(name+'.log')).open('w') as log:
   result=subprocess.run(['go','test','./internal/lower','-run',test,'-count=1','-v','-timeout','90s'],stdout=log,stderr=subprocess.STDOUT,timeout=180)
  output=(evidence/(name+'.log')).read_text()
  assert result.returncode==1 and wanted in output and '[build failed]' not in output,output
  print(name,'caught; test exit',result.returncode,flush=True)
 finally:
  for filename,original in originals.items():Path(filename).write_text(original)

mutant('assignable-unfiltered-native',{'internal/native/view_unions_untagged.go':('\t\t\tif !e.program.Functions[function].Closure || e.program.Functions[function].Receiver {\n\t\t\t\tcontinue\n\t\t\t}\n','')},'^TestCallableProducerRegistryFiltersDirectFunctions$','comparison of distinct pointer types')
mutant('assignable-direct-certificate',{
 'internal/lower/view_unions_untagged.go':('\t\t\tfunction := l.result.Functions[producer.function]\n\t\t\t// Named functions also have a direct-call record. Only their closure\n\t\t\t// thunk has the code convention registered by the callable adapter.\n\t\t\tif !function.Closure || function.Receiver {\n\t\t\t\tcontinue\n\t\t\t}\n',''),
 'internal/lower/view_callable_producers.go':('!function.Closure || function.Receiver || function.RestElement','function.Receiver || function.RestElement'),
},'^TestCallableProducerCertificatesUseClosureThunks$','producer certificate includes direct function')
mutant('exact-identity',{'internal/lower/view_unions_untagged.go':('l.checker.IsTypeAssignableTo(producer.proven, target) && l.untaggedCallableABI(producer.function, *contract)','checker.Checker_isTypeIdenticalTo(l.checker, producer.proven, target) && l.untaggedCallableABI(producer.function, *contract)')},'^TestCallableProducerLiteralReturn$','JavaScript backend stdout')
mutant('skip-adapter-refusal',{'internal/lower/view_lazy.go':('\t\t\tif err := l.callableProducerViewRefusal(graph, read); err != nil {\n\t\t\t\trefused = err\n\t\t\t\treturn false\n\t\t\t}\n','')},'^TestCallableProducer(FewerParameters|MethodShorthand|ExtraOptional)Refused$','got <nil>')

mutant('skip-result-registry-bound',{'internal/lower/view_callable_producers.go':('\tif function.Returns != 0 && function.Returns != ir.Number && function.Returns != ir.Boolean && function.Returns != ir.String {\n\t\treturn false\n\t}\n','')},'^TestCallableProducerDiscardedObjectResultRefused$','got <nil>')

mutant('eager-tagged-callable',{'internal/lower/view_callable_producers.go':('if contract.Kind == ir.ViewObject {','if false && contract.Kind == ir.ViewObject {')},'^TestCallableProducerTaggedUnreadMethod$','Lower refused acceptance row')
