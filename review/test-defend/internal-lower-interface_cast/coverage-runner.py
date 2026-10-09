import pathlib,subprocess,time,json,concurrent.futures
R=pathlib.Path('/workspace/adamic');P=R/'review/test-defend/internal-lower-interface_cast';(P/'coverage').mkdir(exist_ok=True)
pairs=[('TestDefaultTaggedInterfaceNeedsNoFlag','TestDefaultTaggedInterfaceAdmission'),('TestIteratorGapsAreExplicit','TestIteratorViewsCannotEraseReceivers'),('TestIteratorViewsCannotHideReturn','TestIteratorViewsCannotEraseReceivers'),('TestIteratorViewsCannotEraseReceivers','TestIteratorViewsCannotHideReturn'),('TestLiteralMethodCapturesCannotMakeCycles','TestInheritanceRefusesUnsoundOverrides'),('TestLiteralMethodViewsDoNotLoseThis','TestDestructuredMethodsCannotLoadOwnSlots'),('TestRepresentedMethodReplacementIsNotYet','TestIteratorDescriptorReasons'),('TestDestructuredMethodsCannotLoadOwnSlots','TestLiteralMethodViewsDoNotLoseThis'),('TestGenericIteratorViewsPreserveNativeArguments','TestUncheckableCastsStayRefused'),('TestIteratorSymbolKeysAreNotStringKeys','TestClassWrongOutputKeysRepair')];(P/'pairs.json').write_text(json.dumps(pairs,indent=2));res=[]
def run(name):
 cmd=['timeout','120','go','test','-count=1','-timeout','90s','-coverpkg=github.com/system-inc/adamic/internal/lower','-coverprofile='+str(P/'coverage'/(name+'.cover')),'./internal/lower/','-run','^'+name+'$'];start=time.monotonic()
 with (P/'coverage'/(name+'.log')).open('w') as f:r=subprocess.run(cmd,cwd=R,stdout=f,stderr=subprocess.STDOUT)
 return dict(test=name,command=cmd,seconds=time.monotonic()-start,exit=r.returncode)
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
 for r in pool.map(run,dict.fromkeys(n for pair in pairs for n in pair)):
  res.append(r);(P/'coverage-times.json').write_text(json.dumps(res,indent=2));print(r['test'],r['exit'],round(r['seconds'],2),flush=True)
def covered(name):
 out={}
 for l in (P/'coverage'/(name+'.cover')).read_text().splitlines()[1:]:
  block,stmt,count=l.split()
  if int(count):out[block]=int(stmt)
 return out
results=[]
for row,sub in pairs:
 a,b=covered(row),covered(sub);blocks=sorted(set(a)-set(b));lines=set()
 for k in blocks:
  file,span=k.rsplit(':',1);start,end=span.split(',');lines.update(file+':'+str(n) for n in range(int(start.split('.')[0]),int(end.split('.')[0])+1))
 exclusive=set()
 for k in b:
  file,span=k.rsplit(':',1);start,end=span.split(',');exclusive.update(file+':'+str(n) for n in range(int(start.split('.')[0]),int(end.split('.')[0])+1))
 lines=sorted(lines-exclusive)
 results.append(dict(test=row,subsumer=sub,row_only_blocks=blocks,row_only_lines=lines));(P/'coverage'/(row+'-difference.txt')).write_text('\n'.join(blocks)+'\nEXCLUSIVE SOURCE LINES\n'+'\n'.join(lines)+'\n')
(P/'coverage-differences.json').write_text(json.dumps(results,indent=2));print([(r['test'],len(r['row_only_blocks']),len(r['row_only_lines'])) for r in results],flush=True)
