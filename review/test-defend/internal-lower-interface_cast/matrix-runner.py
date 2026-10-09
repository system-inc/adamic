import pathlib,subprocess,time,json,difflib,os
R=pathlib.Path('/workspace/adamic');P=R/'review/test-defend/internal-lower-interface_cast';PKG='./internal/lower/'
plans=[
 dict(id='D01',test='TestIteratorViewsCannotHideReturn',file='internal/lower/iteration.go',old='"an iterator view that can hide a return method"',new='"an iterator view with an omitted close method"',kind='change diagnostic constant',difference='distinct hidden-close refusal path versus erased source receiver'),
 dict(id='D02',test='TestIteratorViewsCannotEraseReceivers',file='internal/lower/iteration.go',old='"an iterable view that erases its method receiver convention"',new='"an iterable view with a different method ABI"',kind='change diagnostic constant',difference='source receiver mismatch path versus hidden-close presence'),
 dict(id='D03',test='TestRepresentedMethodReplacementIsNotYet',file='internal/lower/class.go',old='"replacing a represented method at runtime"',new='"writing a represented method at runtime"',kind='change diagnostic constant',difference='direct setProperty safeguard; subsumer is refused earlier by iterator protocol planning'),
 dict(id='D04',test='TestIteratorSymbolKeysAreNotStringKeys',file='internal/lower/class_features.go',old='"Object.keys on a literal with symbol-key storage"',new='"Object.keys on a literal with a hidden symbol slot"',kind='change diagnostic constant',difference='literal symbol-key guard; repaired subsumer passes a fresh plain object'),
 dict(id='D05',test='TestLiteralMethodCapturesCannotMakeCycles',file='internal/lower/iteration.go',old='\tl.closureRecords = append(l.closureRecords, closureRecord{proven: l.checker.GetTypeAtLocation(node), function: index, node: node})',new='',kind='drop statement',difference='literal-method closure registration and captured-cell cycle path, versus inherited class fields'),
 dict(id='D06',test='TestDestructuredMethodsCannotLoadOwnSlots',file='internal/lower/iteration_origin.go',old='if isCallee(where) { // Calls carry the runtime closure\'s receiver convention.',new='if true { // Calls carry the runtime closure\'s receiver convention.',kind='change condition constant',difference='allow receiver exception for non-call destructuring through callback view; subsumer only calls'),
 dict(id='D07',test='TestIteratorGapsAreExplicit',file='internal/lower/iteration_consume.go',old='} else if element != plan.element {',new='} else if false && element != plan.element {',kind='change condition constant',difference='collection into optional-number array versus numeric iterator element, absent from receiver-erasure input')]
original={x['file']:(R/x['file']).read_text() for x in plans}
for x in plans:
 s=original[x['file']];assert s.count(x['old'])==1,(x['id'],s.count(x['old']));x['line']=s[:s.index(x['old'])].count('\n')+1
 (P/(x['id']+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),s.replace(x['old'],x['new'],1).splitlines(True),fromfile='a/'+x['file'],tofile='b/'+x['file'])))
(P/'plan.json').write_text(json.dumps(plans,indent=2));res=[]
def run(id,cmd,env=None):
 start=time.monotonic()
 with (P/(id+'.log')).open('w') as f:r=subprocess.run(cmd,cwd=R,stdout=f,stderr=subprocess.STDOUT,env=env)
 item=dict(id=id,command=cmd,seconds=time.monotonic()-start,exit=r.returncode,selector=(env or {}).get('ADAMIC_MUTANT'),build_cache=(env or {}).get('ADAMIC_BUILD_CACHE_DIR'));res.append(item);(P/'runs.json').write_text(json.dumps(res,indent=2));print(id,r.returncode,round(item['seconds'],2),flush=True);return r.returncode
for x in plans:
 f=R/x['file'];f.write_text(original[x['file']].replace(x['old'],x['new'],1))
 try:
  if run(x['id']+'-vet',['timeout','90','go','vet',PKG]):raise SystemExit('invalid standalone')
 finally:f.write_text(original[x['file']])
for f,s in original.items():
 for x in (x for x in plans if x['file']==f):
  if x['id'] in ['D01','D02','D03','D04']:new='defenseDiagnostic("'+x['id']+'", '+x['old']+', '+x['new']+')'
  elif x['id']=='D05':new='if !defenseMutant("D05") { '+x['old'].strip()+' }'
  elif x['id']=='D06':new='if isCallee(where) || defenseMutant("D06") { // Calls carry the runtime closure\'s receiver convention.'
  else:new='} else if !defenseMutant("D07") && element != plan.element {'
  s=s.replace(x['old'],new,1)
 (R/f).write_text(s)
helper=R/'internal/lower/defense_selector.go';helper.write_text('package lower\nimport "os"\nfunc defenseMutant(id string) bool { return os.Getenv("ADAMIC_MUTANT")==id }\nfunc defenseDiagnostic(id,normal,mutated string) string { if defenseMutant(id) { return mutated }; return normal }\n')
try:
 if run('switch-vet',['timeout','90','go','vet',PKG]):raise SystemExit('invalid switched source')
 for x in plans:
  env=os.environ.copy();env['ADAMIC_MUTANT']=x['id'];env['ADAMIC_BUILD_CACHE_DIR']='/workspace/defend-lower-cache/'+x['id']
  run(x['id'],['timeout','120','go','test','-json','-count=1','-timeout','90s',PKG,'-run','.'],env)
finally:
 for f,s in original.items():(R/f).write_text(s)
 helper.unlink()
