from pathlib import Path
import difflib
import subprocess

evidence = Path('review/compiler/fx7-callable-producers')
mutants = [
 ('unfiltered-native', 'internal/native/view_unions_untagged.go',
  '\t\t\tif !e.program.Functions[function].Closure || e.program.Functions[function].Receiver {\n\t\t\t\tcontinue\n\t\t\t}\n',
  '^TestCallableProducerRegistryFiltersDirectFunctions$', 'comparison of distinct pointer types'),
 ('direct-producer-certificate', 'internal/lower/view_unions_untagged.go',
  '\t\t\tfunction := l.result.Functions[producer.function]\n\t\t\t// Named functions also have a direct-call record. Only their closure\n\t\t\t// thunk has the code convention registered by the callable adapter.\n\t\t\tif !function.Closure || function.Receiver {\n\t\t\t\tcontinue\n\t\t\t}\n',
  '^TestCallableProducerCertificatesUseClosureThunks$', 'producer certificate includes direct function'),
]
for name, filename, before, test, wanted in mutants:
 path=Path(filename); original=path.read_text()
 assert original.count(before)==1
 mutated=original.replace(before,'',1)
 (evidence/(name+'.diff')).write_text(''.join(difflib.unified_diff(original.splitlines(True),mutated.splitlines(True),fromfile='a/'+filename,tofile='b/'+filename)))
 try:
  path.write_text(mutated)
  with (evidence/(name+'.log')).open('w') as log:
   result=subprocess.run(['go','test','./internal/lower','-run',test,'-count=1','-v','-timeout','90s'],stdout=log,stderr=subprocess.STDOUT,timeout=180)
  output=(evidence/(name+'.log')).read_text()
  assert result.returncode==1 and wanted in output and '[build failed]' not in output,output
  print(name,'caught; test exit',result.returncode,flush=True)
 finally:path.write_text(original)
