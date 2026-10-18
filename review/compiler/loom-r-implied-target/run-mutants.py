import json,pathlib,subprocess,time
out=pathlib.Path(__file__).resolve().parent
root=out.parents[2]
cases=[]
for name,condition,test in [
 ('array-per-element-query','ok && parent.Kind != ast.KindArrayLiteralExpression','^TestImpliedTargetLargeArrayQueriesOnce$'),
 ('conditional-per-branch-query','ok && parent.Kind != ast.KindConditionalExpression','^TestImpliedTargetConditionalQueriesOnce$'),
 ('logical-per-operand-query','ok && parent.Kind != ast.KindBinaryExpression','^TestImpliedTarget(Or|And|Nullish)QueriesOnce$'),
 ('nil-cache-miss','ok && target != nil','^TestImpliedTargetTupleCachesNil$')]:
 cases.append((name,'internal/lower/invariance.go','if target, ok := l.impliedTargets[parent]; ok {','if target, ok := l.impliedTargets[parent]; '+condition+' {',test,'more than once'))
for name,before,after,test,marker in [
 ('shape-ignores-readonly', ':%p:%d:%t;", value, field.Flags, f.l.checker.IsReadonlySymbol(field)', ':%p:%d;", value, field.Flags', '^TestCycleScalarShapeKeysPreserveDistinctFields$', 'shared a shape key'),
 ('shape-per-literal','if f.scalarShapes[key] {','if false && f.scalarShapes[key] {','^TestCycleScalarLiteralShapesOnce$','scalar literal shapes = 2'),
 ('shape-ignores-type',':%p:%d:%t;", value, field.Flags',':%d:%t;", field.Flags','^TestCycleScalarShapeKeysPreserveDistinctFields$','shared a shape key'),
 ('shape-ignores-name','key.WriteString(field.Name)','key.WriteString("")','^TestCycleScalarShapeKeysPreserveDistinctFields$','shared a shape key'),
 ('shape-includes-reference','if !boolean && flags & ^(','if false && !boolean && flags & ^(','^TestCycleScalarShapeKeysPreserveDistinctFields$','want 7/4')]:
 cases.append((name,'internal/lower/cycles.go',before,after,test,marker))
results=[]
for name,file,before,after,test,marker in cases:
 source=(root/file).read_text();assert source.count(before)==1,(name,source.count(before))
 mutant=out/(name+'.go.txt');mutant.write_text(source.replace(before,after,1))
 overlay=out/(name+'.json');overlay.write_text(json.dumps({'Replace':{str(root/file):str(mutant)}}))
 command=['go','test','-overlay='+str(overlay),'./internal/lower','-run',test,'-count=1','-json','-timeout','90s']
 start=time.monotonic();log=out/'logs'/(name+'.jsonl')
 with log.open('w') as stream:result=subprocess.run(command,cwd=root,stdout=stream,stderr=subprocess.STDOUT,timeout=150)
 text=log.read_text();caught=result.returncode==1 and marker in text and '"Action":"fail"' in text and '[build failed]' not in text
 results.append(dict(mutant=name,command=command,exit=result.returncode,caught=caught,catcher=test,marker=marker,seconds=time.monotonic()-start,log=str(log)))
 (out/'mutants.json').write_text(json.dumps(results,indent=2)+'\n');print(name,caught,round(results[-1]['seconds'],3),flush=True)
 if not caught:raise SystemExit(1)
