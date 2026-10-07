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
virtual=repo/'cohere/adamic_wave01_await_expand.go'
overlay=out/'overlay.json'
overlay.write_text(json.dumps({'Replace':{str(virtual):str(source/'await_expand_main.go'),str(repo/'cohere/internal/lint/rules/core/adamic_wave01_expand.go'):str(source/'await_expand_capture.go')}}))
oracle=out/'oracle';run('go-build',['go','build','-overlay',overlay,'-o',oracle,virtual],repo/'cohere')
control=out/'controls.a';control.write_text('''type First = number | Promise<number>;
type Second = string | Promise<string>;
type Parameter<T> = T;
async function probe(){return 1;}
export {};
''')
config=out/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ESNext','lib':['esnext']},'files':[str(control)]}))
data,error=run('go',[oracle,config,control]);assert not error
cases=json.loads(data);assert len(cases)==16
module=source.parent/'require_await/expand.a'
code=['import { expandAwaitTypes } from '+json.dumps(str(module))+';']
expected=[]
for i,c in enumerate(cases):
    for key,variable in [('Unions','unions'),('Substitutions','substitutions')]:
        name=variable+str(i);code.append('const '+name+' = new Map<number, readonly number[]>();')
        for identity,parts in c[key].items():
            code.append(name+'.set('+identity+','+json.dumps(parts)+');')
    code.append('console.log(`'+str(i)+' ${expandAwaitTypes('+json.dumps(c['Inputs'])+',unions'+str(i)+',substitutions'+str(i)+").join(',')}`);")
    expected.append(str(i)+' '+','.join(map(str,c['Expected']))+'\n')
entry=out/'entry.a';entry.write_text('\n'.join(code)+'\n')
expected=''.join(expected).encode();(out/'expected.stdout').write_bytes(expected)
binary=out/'native';run('native-build',[compiler,'build',entry,'-o',binary,'--sanitize'])
actual,error=run('native',[binary]);assert not error and actual==expected
mutant=out/'mutant.a';mutant.write_text(module.read_text().replace('if(standIns !== undefined) {','if(standIns !== undefined && standIns.length !== 0) {'))
mutant_entry=out/'mutant-entry.a';mutant_entry.write_text(entry.read_text().replace(str(module),str(mutant)))
binary=out/'mutant';run('mutant-build',[compiler,'build',mutant_entry,'-o',binary,'--sanitize'])
actual,error=run('mutant',[binary]);assert not error and actual!=expected
print('PASS',len(cases),'real-type expansion/substitution cases; sanitizer-clean empty-substitution mutant caught by bytes')
