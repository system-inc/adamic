from pathlib import Path
import subprocess,json,difflib,hashlib
out=Path(__file__).parent/'resolved'/'refusal-mutants';out.mkdir(exist_ok=True)
items=[
 ('omit-creation-admission','internal/lower/interface_cast.go','\tif err := l.checkViewMembers(node, target); err != nil {\n\t\treturn nil, err\n\t}\n',''),
 ('wrong-creation-location','internal/lower/view_lazy.go','Where: l.program.Where(node), What: "view type has an unsupported member: "','Where: "wrong:1:1", What: "view type has an unsupported member: "'),
 ('wrong-unsupported-member','internal/lower/view_lazy.go','What: "view type has an unsupported member: " + property.Name','What: "view type has an unsupported member: wrong"')]
results=[]
for name,path,needle,replacement in items:
 p=Path(path);original=p.read_text();assert original.count(needle)==1,(name,path)
 try:
  mutated=original.replace(needle,replacement)
  (out/(name+'.patch')).write_text(''.join(difflib.unified_diff(original.splitlines(True),mutated.splitlines(True),fromfile='a/'+path,tofile='b/'+path)))
  p.write_text(mutated)
  command=['go','test','./internal/lower','-run','^TestTrainPlusRefusalP(26|33|34|35)$','-count=1','-v','-timeout','90s']
  with (out/(name+'.log')).open('w') as log:
   r=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,timeout=120)
  output=(out/(name+'.log')).read_text()
  assert r.returncode==1 and '[build failed]' not in output
  for row in ['26','33','34','35']:
   assert '--- FAIL: TestTrainPlusRefusalP'+row in output
   assert 'native_p'+row+'_' in output
  assert '.a:' in output and '.ts:' in output,output
  results.append({'name':name,'command':command,'exit':r.returncode,'caught':True,'source_paths':['.a','.ts'],'catchers':['TestTrainPlusRefusalP'+n for n in ['26','33','34','35']]})
  (out/'results.json').write_text(json.dumps(results,indent=2)+'\n')
  print(name+': caught in all four .a/.ts cases',flush=True)
 finally:p.write_text(original)
