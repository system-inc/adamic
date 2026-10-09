from pathlib import Path
import subprocess,json
out=Path('review/compiler/miscompile-train-plus-2/resolved/extra-mutants')
results=[]
for name,path,needle,test,pkg in [('metadata-isolated','internal/native/emit_objects.go','\t\tif cast, ok := expression.(ir.CheckedCast); ok && cast.CheckedFields {\n\t\t\tneeded = true\n\t\t}\n','^TestCheckedCastWithoutReadSummary$','native'),('historical-single-filter','internal/lower/view_unions_untagged.go','\t\t\tfunction := l.result.Functions[producer.function]\n\t\t\t// Named functions also have a direct-call record. Only their closure\n\t\t\t// thunk has the code convention registered by the callable adapter.\n\t\t\tif !function.Closure || function.Receiver {\n\t\t\t\tcontinue\n\t\t\t}\n','^TestCallableProducerCertificatesUseClosureThunks$','lower')]:
 p=Path(path);original=p.read_text();assert original.count(needle)==1
 try:
  p.write_text(original.replace(needle,''))
  with (out/(name+'.log')).open('w') as log:r=subprocess.run(['go','test','./internal/'+pkg,'-run',test,'-count=1','-v','-timeout','90s'],stdout=log,stderr=subprocess.STDOUT,timeout=120)
  output=(out/(name+'.log')).read_text();results.append({'name':name,'exit':r.returncode,'caught':r.returncode==1 and 'exit status 70' in output if name=='metadata-isolated' else r.returncode==1});print(results[-1],flush=True)
 finally:p.write_text(original)
(out/'isolated-results.json').write_text(json.dumps(results,indent=2)+'\n')
assert results[0]['caught'],results
