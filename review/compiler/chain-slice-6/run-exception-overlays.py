import ast,json,os,subprocess
from pathlib import Path
root=Path.cwd();out=root/'review/compiler/chain-slice-6/exception-mutants';out.mkdir(exist_ok=True)
rows=[]
for family in ['run-mutants','run-library-mutants','run-saved-mutants','run-subclass-mutants','run-adopted-mutants','run-main-mutants']:
 tree=ast.parse((root/'review/compiler/chain-slice-6'/ (family+'.py')).read_text())
 tuples=next(ast.literal_eval(n.value) for n in tree.body if isinstance(n,ast.Assign) and any(isinstance(t,ast.Name) and t.id=='mutants' for t in n.targets))
 for row in tuples:
  if family=='run-subclass-mutants':
   name,path,old,new,catcher=row;selection='TestNativeAgreesWithNode/internal/oracle/testdata/step21_error_subclasses[.]a$'
  else:
   name,path,old,new,selection,catcher=row
   if family=='run-library-mutants':selection='TestNativeAgreesWithNode/internal/oracle/testdata/step21_'+selection+'[.]a$'
  rows.append((name,[(path,old,new)],selection,catcher, family=='run-library-mutants' or family=='run-subclass-mutants'))
rows.append(('readiness-old-effects',[('internal/lower/readiness.go','defer readinessExceptions(program)','// mutant: omit readiness exception propagation')],'TestNativeAgreesWithNode/internal/oracle/testdata/lowering_chain_tdz_catch[.]a$','stdout differs',False))
# The unknown assertion has independent admission and lowering barriers.
changes=[]
for relative in ['internal/lower/cast_proof.go','internal/lower/cast.go']:
 anchor='as := node.AsAsExpression()'
 if relative.endswith('cast_proof.go'):replacement=anchor+'; if l.checker.GetTypeAtLocation(as.Expression).Flags() & checker.TypeFlagsUnknown != 0 { return castProof{}, nil }'
 else:replacement=anchor+'; if l.checker.GetTypeAtLocation(as.Expression).Flags() & checker.TypeFlagsUnknown != 0 { value, err := l.expression(as.Expression); if err != nil { return nil, err }; return ir.Narrow{Value: value, To: ir.Object}, nil }'
 changes.append((relative,anchor,replacement))
rows.append(('unknown-authorizes-assertion',changes,'^TestStep21UnknownAssertions$','want Refused, got <nil>',False))
results=json.loads((out/'results.json').read_text()) if (out/'results.json').exists() else []
results=[r for r in results if r['caught']]
for name,changes,selection,catcher,replace_all in rows:
 if any(r['name']==name for r in results):continue
 directory=out/name;directory.mkdir(exist_ok=True);replace={}
 for relative,old,new in changes:
  path=root/relative;source=path.read_text();assert old in source,(name,'mutation site absent',relative)
  if not replace_all:assert source.count(old)==1,(name,source.count(old))
  changed=directory/(path.name+'.txt');changed.write_text(source.replace(old,new) if replace_all else source.replace(old,new,1));replace[str(path)]=str(changed)
 overlay=directory/'overlay.json';overlay.write_text(json.dumps({'Replace':replace},indent=2)+'\n')
 command=['go','test','-overlay',str(overlay),'./internal/oracle','-run',selection,'-count=1','-v','-timeout','90s']
 originals={}
 try:
  # External Node reads these directly; Go overlays affect compiler inputs only.
  for relative,old,new in changes:
   if relative.endswith('.mjs'):
    path=root/relative;originals[path]=path.read_bytes();path.write_text(path.read_text().replace(old,new,1))
  with (directory/'test.log').open('w') as log:
   result=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,env=dict(os.environ,ADAMIC_GATE_UNCACHED='1'),timeout=120)
 finally:
  for path,original in originals.items():path.write_bytes(original)
 text=(directory/'test.log').read_text();caught=result.returncode!=0 and catcher in text and '[build failed]' not in text and 'clang failed' not in text
 results.append({'name':name,'exit':result.returncode,'catcher':catcher,'caught':caught,'command':command});(out/'results.json').write_text(json.dumps(results,indent=2)+'\n')
 print(name, 'caught' if caught else 'NOT CAUGHT',catcher,flush=True)
 assert caught,(name,'see '+str(directory/'test.log'))
