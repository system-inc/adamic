import json,subprocess,difflib
from pathlib import Path
out=Path(__file__).resolve().parent;folder=out/'mutants';folder.mkdir(exist_ok=True);source=json.loads((out/'base.json').read_text())['source'];cases=[];edits=[]
def save(topic,file,current,mutated,packages,pattern,description):
 assert current!=mutated,topic
 replacement=folder/(topic+'.go.txt');replacement.write_text(mutated);overlay=folder/(topic+'.json');overlay.write_text(json.dumps({'Replace':{str(Path(file).resolve()):str(replacement)}}))
 cases.append(dict(topic=topic,file=file,overlay=str(overlay),packages=packages,pattern=pattern));edits.append(dict(topic=topic,file=file,description=description,patch=''.join(difflib.unified_diff(current.splitlines(True),mutated.splitlines(True),fromfile='a/'+file,tofile='b/'+file))))
for c in json.loads(Path('review/compiler/chain-followers/host-mutants/cases.json').read_text()):
 file=c['file'];base=subprocess.check_output(['git','show',source+':'+file],text=True);archived=(Path('review/compiler/chain-followers/host-mutants')/(c['topic']+'.go.txt')).read_text();current=Path(file).read_text();mutated=current;ops=[]
 for tag,a,b,x,y in difflib.SequenceMatcher(None,base.splitlines(True),archived.splitlines(True),autojunk=False).get_opcodes():
  if tag!='equal':ops.append((''.join(base.splitlines(True)[a:b]),''.join(archived.splitlines(True)[x:y]),a))
 description='minimal recorded mutation over current source'
 if c['topic']=='generic-values':ops=[op for op in ops if 'return l.genericFunctionValue(node, declaration)' in op[0]];assert len(ops)==1
 if c['topic']=='phantom-overload-interaction':ops=[op for op in ops if op[0].strip() or op[1].strip()]
 if c['topic']=='string-concatenation':
  old='case ir.Null:\n\t\treturn ir.StringConstant{Index: l.constant("null")}';assert current.count(old)==1
  mutated=current.replace(old,'case ir.Null:\n\t\treturn ir.StringConstant{Index: l.constant("wrong-null")}',1)
  description='Isolated type-correct null spelling regression; archived whole-file overlay also reverted unrelated async/template main code, which is preserved here.'
 else:
  for old,new,a in ops:
   if old:
    assert mutated.count(old)==1,(c['topic'],old);mutated=mutated.replace(old,new,1)
   else:
    anchor=''.join(base.splitlines(True)[max(0,a-1):a]);assert mutated.count(anchor)==1,(c['topic'],anchor);mutated=mutated.replace(anchor,anchor+new,1)
 save(c['topic'],file,current,mutated,c['packages'],c['pattern'].replace('^TestNativeAgreesWithNode/','^TestNativeAgreesWithNode_[0-9]+$/'),description)
extra=[
 ('constructor-context','internal/lower/class_inheritance.go','\touterClosures := l.closures\n\tl.closures = nil\n\tdefer func() { l.closures = outerClosures }()\n','', '^TestArrowFieldCaptureFrames$','recorded arrow mutation, held by IR frame assertions before C emission'),
 ('dropped-capture','internal/lower/locals.go','\t\t\tfunction.Environment = append(function.Environment, local)\n','\t\t\t// Mutant: omit captured cell.\n','^TestArrowSuperCaptureFrames$','recorded arrow mutation, held by IR frame assertions before C emission'),
 ('main-implied-target-cache','internal/lower/invariance.go','\tif target, ok := l.impliedTargets[parent]; ok {\n\t\treturn target\n\t}\n','','^TestImpliedTargetOrQueriesOnce$','prove main cache retained across the conflict'),
 ('generic-operand-context','internal/lower/generic_function_value.go','return instantiateSignatureInContextOf(l.checker, target, signatures[0], nil, nil)','return nil','^TestGenericFunctionValueOperandContexts$','prove branch concrete checker-derived operand context')]
for topic,file,old,new,pattern,description in extra:
 current=Path(file).read_text();assert current.count(old)==1,(topic,current.count(old));save(topic,file,current,current.replace(old,new,1),['./internal/lower'],pattern,description)
(folder/'cases.json').write_text(json.dumps(cases,indent=2));(folder/'ported-edits.json').write_text(json.dumps(edits,indent=2));print('Prepared',len(cases),'current-source mutants')
s=subprocess.check_output(['git','show','compiler/loom-r-string-brand-ac-m:review/compiler/loom-r-string-brand-ac-m/validate.py'],text=True);prefix=s.split("standard=['-count=1'")[0].replace('results=[]',"results=json.loads((out/'checks.json').read_text())")
rest="""standard=['-count=1','-timeout','90s','-parallel','4','-json']
for c in json.loads((out/'mutants/cases.json').read_text()):
 run('mutant-'+c['topic'],['go','test','-overlay',c['overlay'],*c['packages'],'-run',c['pattern'],*standard],expected=1,timeout=180)
 text=(out/('mutant-'+c['topic']+'.log')).read_text()
 assert '[build failed]' not in text and 'use of undeclared identifier' not in text,c['topic']
 events=[]
 for line in text.splitlines():
  try:e=json.loads(line)
  except ValueError:continue
  if e.get('Action')=='fail' and 'Test' in e:events.append(e)
 assert events,(c['topic'],'no test catcher')
"""
(out/'mutant-check.py').write_text(prefix+rest)
