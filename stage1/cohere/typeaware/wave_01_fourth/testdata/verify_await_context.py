"""Compare native type expansion with unchanged Go requireAwaitExpandTypes."""
import json
import os
from pathlib import Path
import subprocess
import sys
repo=Path(__file__).resolve().parents[5]
source=Path(__file__).resolve().parent
out=Path(sys.argv[1]).resolve();out.mkdir(parents=True,exist_ok=True)
compiler=Path(sys.argv[2]).resolve()
def run(name,args,cwd=repo):
    with (out/(name+'.stdout')).open('wb') as stdout,(out/(name+'.stderr')).open('wb') as stderr:
        p=subprocess.run(list(map(str,args)),cwd=cwd,stdout=stdout,stderr=stderr,env=dict(os.environ,ASAN_OPTIONS='detect_leaks=1',UBSAN_OPTIONS='halt_on_error=1'))
    assert p.returncode==0,(name,p.returncode)
    return (out/(name+'.stdout')).read_bytes(),(out/(name+'.stderr')).read_bytes()
virtual=repo/'cohere/adamic_wave01_await_context.go'
overlay=out/'overlay.json'
overlay.write_text(json.dumps({'Replace':{str(virtual):str(source/'await_context_main.go'),str(repo/'cohere/internal/lint/rules/core/adamic_wave01_context.go'):str(source/'await_context_capture.go')}}))
oracle=out/'oracle';run('go-build',['go','build','-overlay',overlay,'-o',oracle,virtual],repo/'cohere')
control=out/'controls.a';control.write_text("""type Named = {run: () => Promise<void>; [key: string]: unknown};
type Indexed = {[key:string]: number};
type Tuple = [number, string, Promise<void>];
type Array = boolean[];
type Function = {(): number; (x:string): Promise<string>};
type Plain = {};
async function probe(){return 1;}
export {};
""")
config=out/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ESNext','lib':['esnext']},'files':[str(control)]}))
data,error=run('go',[oracle,config,control]);assert not error
cases=json.loads(data);assert len(cases)==12
module=source.parent/'require_await/context_step.a'
code=['import { applyAwaitContextStep } from '+json.dumps(str(module))+';']
expected=[]
for i,c in enumerate(cases):
    facts=[{'property':f['Property'],'stringIndex':f['StringIndex'],'tuple':f['Tuple'],'elements':f['Elements'],'numberIndex':f['NumberIndex'],'returns':f['Returns']} for f in c['Facts']]
    code.append('console.log(`'+str(i)+' ${applyAwaitContextStep('+str(c['Kind'])+','+str(c['Index'])+','+json.dumps(facts)+").join(',')}`);")
    expected.append(str(i)+' '+','.join(map(str,c['Expected']))+'\n')
entry=out/'entry.a';entry.write_text('\n'.join(code)+'\n')
expected=''.join(expected).encode();(out/'expected.stdout').write_bytes(expected)
binary=out/'native';run('native-build',[compiler,'build',entry,'-o',binary,'--sanitize'])
actual,error=run('native',[binary]);assert not error and actual==expected
mutant=out/'mutant.a';mutant.write_text(module.read_text().replace('if(part.property !== 0) {','if(false) {'))
mutant_entry=out/'mutant-entry.a';mutant_entry.write_text(entry.read_text().replace(str(module),str(mutant)))
binary=out/'mutant';run('mutant-build',[compiler,'build',mutant_entry,'-o',binary,'--sanitize'])
actual,error=run('mutant',[binary]);assert not error and actual!=expected
print('PASS',len(cases),'real-type contextual lookup cases; sanitizer-clean property-priority mutant caught by bytes')
