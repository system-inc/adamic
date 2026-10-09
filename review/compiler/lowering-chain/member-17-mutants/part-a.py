import ast,json,os,subprocess,time
from pathlib import Path
root=Path.cwd();out=Path(__file__).resolve().parent
text=(root/'review/compiler/lowering-chain/source-member-14/mutants.py').read_text()
tree=ast.parse(text);probes=next(ast.literal_eval(n.value) for n in tree.body if isinstance(n,ast.Assign) and any(isinstance(t,ast.Name) and t.id=='probes' for t in n.targets))
rules=[(name,'internal/lower/known_lowering_gaps.go',before,after,'./internal/oracle','^TestMiscompile2A'+selector+'$',('want path-bearing stop','panic: Unhandled case in Node.Text')) for name,before,after,selector in probes]
rules += [('optional-error','internal/native/runtime/exceptions.c','adamic_retain(message == NULL ? &empty_message : message)','adamic_retain(message)','./internal/oracle','^TestMiscompile2A(OptionalError|MarkerUndefined)$',('runtime error:','member access within null pointer')),('error-own-message','internal/native/runtime/exceptions.c','adamic_object_new(message == NULL ? &absent_message_shape : &error_shape)','adamic_object_new(&error_shape)','./internal/native','^TestErrorUndefinedMessageHasNoOwnProperty$',('native ','Node ','[] true true'))]
results=[]
for name,relative,before,after,package,test,witness in rules:
 original=root/relative;contents=original.read_text();assert contents.count(before)==1,(name,contents.count(before))
 contents=contents.replace(before,after,1)
 if name=='optional-error':contents=contents.replace('static adamic_string empty_message = ADAMIC_STRING("");\n','')
 directory=out/name;directory.mkdir(exist_ok=True)
 mutant=directory/(original.name+'.txt');mutant.write_text(contents)
 overlay=directory/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(original):str(mutant)}}))
 command=['go','test','-p','1','-parallel','4','-overlay',str(overlay),package,'-run',test,'-count=1','-timeout','90s','-v']
 start=time.monotonic()
 with (directory/'test.log').open('w') as log:r=subprocess.run(command,cwd=root,stdout=log,stderr=subprocess.STDOUT,timeout=120,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1','UBSAN_OPTIONS':'symbolize=0'})
 observed=(directory/'test.log').read_text();caught=r.returncode!=0 and (any(w in observed for w in witness) if relative.endswith('.go') else all(w in observed for w in witness)) and '[build failed]' not in observed and 'clang failed' not in observed
 results.append(dict(mutant=name,caught=caught,exit=r.returncode,seconds=round(time.monotonic()-start,3),command=command));(out/'results.json').write_text(json.dumps(results,indent=2)+'\n');print(name,'caught' if caught else 'NOT CAUGHT',results[-1]['seconds'],flush=True)
 assert caught,(name,observed)
