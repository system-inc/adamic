import ast,json,pathlib,sys
from verify import run,root,out
source=ast.parse((root/'review/compiler/loom-views-fx-b/run-mutants.py').read_text())
cases=ast.literal_eval(next(n.value for n in source.body if isinstance(n,ast.Assign) and any(isinstance(t,ast.Name) and t.id=='cases' for t in n.targets)))
results=[]
for name,file,before,after,packages,selector,marker in cases:
 original=(root/file).read_text();assert original.count(before)==1,(name,original.count(before))
 replacement=out/(name+'.go.txt');replacement.write_text(original.replace(before,after,1))
 overlay=out/(name+'-overlay.json');overlay.write_text(json.dumps({'Replace':{str(root/file):str(replacement)}}))
 code=run(name+'-mutant',['go','test','-overlay',str(overlay),*packages,'-run',selector,'-count=1','-json','-timeout','90s'],180)
 log=(out/(name+'-mutant.log')).read_text()
 caught=code==1 and marker in log and 'build failed' not in log and 'panic: test timed out' not in log
 results.append(dict(mutant=name,file=file,selector=selector,marker=marker,caught=caught))
 (out/'mutants.json').write_text(json.dumps(results,indent=2)+'\n')
 if not caught: sys.exit('mutant was not caught: '+name)
