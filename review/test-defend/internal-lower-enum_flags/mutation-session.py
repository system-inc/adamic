import pathlib,subprocess,json,time,difflib,os
root=pathlib.Path('/workspace/adamic');os.chdir(root);p=root/'review/test-defend/internal-lower-enum_flags';pkg='./internal/lower/'
plans=[
('D1','TestEnumInitializationGraphMemo','internal/lower/namespaces_call_graph.go','if cached := g.functions[function]; cached != nil {','if cached := g.functions[function]; cached != nil && (cached.active || cached.index != 0 || len(cached.reaches) != 1) {','flip condition: reuse active nodes and zero/multiple-target caches, but recompute the first completed single-target cache','Shared cache-hit lines; enum memo row revisits a completed one-target graph including its first cache entry, while its mixed cycle subsumer has two targets.'),
('D2','TestEnumNamespaceSharedCycle','internal/lower/namespaces_call_graph.go','member.reaches = reaches','if member == entry || len(component) != 2 { member.reaches = reaches }','drop statement for nonroot members of a two-node component','Exclusive back-edge blocks at 42-46 and namespace-read block 103; SCC size two differs from the existing three-node namespace cycle.'),
('D3','TestFlagEnumLiteralSpellings','internal/lower/enums.go','return ir.NumberConstant{Value: value}, nil','if initializer := member.AsEnumMember().Initializer; initializer != nil && initializer.Kind == ast.KindNumericLiteral && strings.HasPrefix(initializer.Text(), "0x") { return ir.NumberConstant{Value: value + 1}, nil }; return ir.NumberConstant{Value: value}, nil','change constant by one for hexadecimal numeric enum initializers','Shared enumConstant lines; this row uses hexadecimal direct enum initializers, unlike its inline-iteration subsumer.'),
('D4','TestFlagEnumMemberAliases','internal/lower/enums.go','return ir.NumberConstant{Value: value}, nil','if initializer := member.AsEnumMember().Initializer; initializer != nil && initializer.Kind == ast.KindPropertyAccessExpression { if alias := l.enumMember(initializer); alias != nil && alias.Parent == member.Parent { return ir.NumberConstant{Value: value + 1}, nil } }; return ir.NumberConstant{Value: value}, nil','change constant by one for a qualified same-enum member alias','Shared enumConstant lines; qualified same-enum alias differs from foreign-enum aliases and the subsumer\'s bitwise member values.'),
('D5','TestEnumNameEnumeration','internal/lower/library_for_in.go','if l.enumObject(node) != nil {\n\t\treturn true','if l.enumObject(node) != nil {\n\t\treturn false','flip enum plain-enumerable origin condition','for-in enum origin is reached by name enumeration, unlike the for-of inline-iteration subsumer.'),
('D6','TestStringEnumsStayClosed','internal/lower/cast_proof.go','refused := &Refused{Where: l.program.Where(node), What: "a cast the runtime can\'t check", Fix: castRepair}','refused := &Refused{Where: l.program.Where(node), What: "a cast the runtime can\'t check", Fix: func() string { if source.Flags()&checker.TypeFlagsStringLike != 0 && l.enumIdentity(target) != nil && !l.numericEnum(l.enumIdentity(target)) { return "" }; return castRepair }()}','change diagnostic repair constant to empty for primitive-string to string-enum casts','No exclusive block; shared castProof lines receive a primitive string and a closed string enum. Diagnostic omission is a real production-output mutation, but rejection remains enabled.'),
('D7','TestFlagEnumsOpen','internal/lower/assignments.go','operator, isCompound := compoundAssignments[binary.OperatorToken.Kind]','operator, isCompound := func() (ast.Kind, bool) { operator, ok := compoundAssignments[binary.OperatorToken.Kind]; if operator == ast.KindLessThanLessThanToken && l.flagEnum(l.enumIdentity(l.checker.GetTypeAtLocation(binary.Left))) { return ast.KindGreaterThanGreaterThanToken, ok }; return operator, ok }()','change option: flag-shaped enum compound left shift becomes right shift','Exclusive assignment blocks versus literal-promises subsumer; flag shape distinguishes the open row\'s <<= case from implicit numeric enum shifts.')]
meta=[];runs={}
for ident,row,file,old,new,menu,lead in plans:
 source=(root/file).read_text();assert source.count(old)==1,(ident,source.count(old));line=source[:source.index(old)].count('\n')+1
 changed=source.replace(old,new)
 diff=''.join(difflib.unified_diff(source.splitlines(True),changed.splitlines(True),fromfile='a/'+file,tofile='b/'+file));(p/(ident+'.diff')).write_text(diff)
 meta.append(dict(id=ident,target=row,file=file,line=line,old=old,new=new,menu=menu,coverage_lead=lead))
(p/'defense-plan.json').write_text(json.dumps(meta,indent=2))
def run(n,c,env=None):
 t=time.monotonic()
 with (p/(n+'.log')).open('w') as f:r=subprocess.run(c,stdout=f,stderr=subprocess.STDOUT,env=env)
 runs[n]={'exit':r.returncode,'wall_seconds':time.monotonic()-t,'command':c};return r.returncode
for m in meta:
 s=root/m['file'];original=s.read_text()
 try:
  assert run(m['id']+'-apply',['git','apply','--check',str(p/(m['id']+'.diff'))])==0
  s.write_text(original.replace(m['old'],m['new']))
  assert run(m['id']+'-vet',['go','vet',pkg])==0
  env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/enum-defend/cache/'+m['id']
  run(m['id'],['timeout','120','go','test','-json','-count=1','-timeout','90s',pkg,'-run','.'],env)
  events=[]
  for line in (p/(m['id']+'.log')).read_text().splitlines():
   try:events.append(json.loads(line))
   except:pass
  failed=[e['Test'] for e in events if e['Action']=='fail' and 'Test'in e and '/'not in e['Test']]
  print(m['id'],round(runs[m['id']]['wall_seconds'],3),failed,flush=True)
  (p/'matrix-runs.json').write_text(json.dumps(runs,indent=2))
 finally:s.write_text(original)
(p/'matrix-runs.json').write_text(json.dumps(runs,indent=2))
