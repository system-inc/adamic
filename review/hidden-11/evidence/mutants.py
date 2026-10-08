from pathlib import Path
import subprocess,json,os
root=Path('/workspace/adamic')
mutants=[('captured-derived-default','internal/lower/expression.go','\t\tread := ir.Expression(ir.Read{Local: local,','\t\tif node.Text() == "derived" && l.result.Locals[local].Captured { return ir.NumberConstant{}, nil }\n\t\tread := ir.Expression(ir.Read{Local: local,',['go','test','./internal/oracle','-run','^TestNativeAgreesWithNode/internal/oracle/testdata/hidden_boundary_derived_binding.a$','-count=1'],'stdout'),('self-initializer-accepted','internal/lower/locals.go','if name, isInitializing := l.initializing[local]; isInitializing && l.unlowerable == nil {','if name, isInitializing := l.initializing[local]; false && isInitializing && l.unlowerable == nil {',['go','test','./internal/lower','-run','^TestHiddenDerivedSelfInitializerStaysNotYet$','-count=1'],'assertion')]
results=[]
for name,file,old,new,args,catch in mutants:
 path=root/file;original=path.read_text();assert original.count(old)==1
 log=Path('/tmp/hidden-11-mutant-'+name+'.log')
 try:
  path.write_text(original.replace(old,new))
  with log.open('w') as stream:done=subprocess.run(args,cwd=root,env=dict(os.environ,ADAMIC_GATE_UNCACHED='1'),stdout=stream,stderr=subprocess.STDOUT)
 finally:path.write_text(original)
 text=log.read_text();assert done.returncode==1 and '--- FAIL:' in text and '[build failed]' not in text,(name,text)
 if catch=='stdout':assert 'native stdout' in text or 'stdout' in text,text
 results.append(dict(mutant=name,exit=done.returncode,catcher=catch))
Path('/tmp/hidden-11-mutants.json').write_text(json.dumps(results,indent=2)+'\n')
