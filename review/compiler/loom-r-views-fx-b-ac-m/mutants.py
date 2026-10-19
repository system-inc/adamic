import ast,json,pathlib,subprocess,time
root=pathlib.Path.cwd();out=root/'review/compiler/loom-r-views-fx-b-ac-m'
a=ast.parse((root/'review/compiler/loom-views-fx-b/run-mutants.py').read_text());cases=ast.literal_eval(next(n.value for n in a.body if isinstance(n,ast.Assign) and any(isinstance(x,ast.Name) and x.id=='cases' for x in n.targets)))
cases=[tuple(list(c[:5])+[c[5].replace('TestFX6P37','TestLowerViewsUnrelatedUnionField').replace('TestFX6P53','TestLowerViewsDestructuredCallableUnion')]+list(c[6:])) for c in cases]
cases += [
 ('literal-key','internal/lower/element_access_fields.go','func (l *lowering) refuseViewElementReads(fields map[string]bool) error {','func (l *lowering) refuseViewElementReads(fields map[string]bool) error {\n fields = nil',['./internal/oracle'],'^TestReviewProgramsRefuse$','accepted refused review program'),
 ('incompatible-producer','internal/lower/collections.go','if err := l.refuseIncompatibleDestructuredView(pattern, initializer); err != nil {','if err := l.refuseIncompatibleDestructuredView(pattern, initializer); false && err != nil {',['./internal/oracle'],'^TestOracleViewsDestructuredInvalidObjectUnionRefuses$','want located incompatible-producer refusal')]
results=[]
for name,file,before,after,packages,selector,marker in cases:
 original=(root/file).read_text();assert original.count(before)==1,(name,original.count(before))
 replacement=out/(name+'.go.txt');replacement.write_text(original.replace(before,after,1));overlay=out/(name+'-overlay.json');overlay.write_text(json.dumps({'Replace':{str(root/file):str(replacement)}}))
 cmd=['go','test','-overlay',str(overlay),*packages,'-run',selector,'-count=1','-json','-timeout','90s'];start=time.monotonic()
 with (out/(name+'-mutant.log')).open('w') as log:
  try:p=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,timeout=150);code=p.returncode
  except subprocess.TimeoutExpired:code=124
 text=(out/(name+'-mutant.log')).read_text();caught=code==1 and marker in text and 'build failed' not in text and 'test timed out' not in text
 results.append({'name':name,'command':cmd,'exit':code,'seconds':time.monotonic()-start,'caught':caught,'marker':marker});(out/'mutants.json').write_text(json.dumps(results,indent=2)+'\n');print(name,code,caught,flush=True)
 assert caught,name
