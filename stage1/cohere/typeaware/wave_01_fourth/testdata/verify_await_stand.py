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
virtual=repo/'cohere/adamic_wave01_await_stand.go'
overlay=out/'overlay.json'
overlay.write_text(json.dumps({'Replace':{str(virtual):str(source/'await_stand_main.go'),str(repo/'cohere/internal/lint/rules/core/adamic_wave01_stand.go'):str(source/'await_stand_capture.go')}}))
oracle=out/'oracle';run('go-build',['go','build','-overlay',overlay,'-o',oracle,virtual],repo/'cohere')
control=out/'controls.a';control.write_text("""declare function pick<T>(a:T,b:T):T;
declare function rest<T>(a:T,...b:T[]):T;
declare function nested<T>(a:{x:T},b:T):T;
declare function many<T,U>(a:T,b:U,c:T):U;
pick(1,2);
const typed:number=pick(3,4);
rest(1,2,3);
nested({x:1},2);
many(1,'x',2);
export {};
""")
config=out/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ESNext','lib':['esnext']},'files':[str(control)]}))
data,error=run('go',[oracle,config,control]);assert not error
cases=json.loads(data);assert len(cases)==12
module=source.parent/'require_await/stand_ins.a'
code=['import { awaitStandIns } from '+json.dumps(str(module))+';', 'const empty: number[] = [];']
expected=[]
for i,c in enumerate(cases):
    facts={k[0].lower()+k[1:]:c[k] for k in ['Parameters','Declared','Resolved','ArgumentCount','ArgumentIndex','Rest','ReturnType','PositionType']}
    code.append('const result'+str(i)+'=awaitStandIns('+json.dumps(facts)+');')
    for identity in c['Parameters']:
        code.append('console.log(`'+str(i)+' '+str(identity)+' ${ (result'+str(i)+'.get('+str(identity)+") ?? empty).join(',')}`);")
        expected.append(str(i)+' '+str(identity)+' '+','.join(map(str,c['Expected'].get(str(identity),[])))+'\n')
entry=out/'entry.a';entry.write_text('\n'.join(code)+'\n')
expected=''.join(expected).encode();(out/'expected.stdout').write_bytes(expected)
binary=out/'native';run('native-build',[compiler,'build',entry,'-o',binary,'--sanitize'])
actual,error=run('native',[binary]);assert not error and actual==expected
mutant=out/'mutant.a';mutant.write_text(module.read_text().replace('index === facts.argumentIndex || ',''))
mutant_entry=out/'mutant-entry.a';mutant_entry.write_text(entry.read_text().replace(str(module),str(mutant)))
binary=out/'mutant';run('mutant-build',[compiler,'build',mutant_entry,'-o',binary,'--sanitize'])
actual,error=run('mutant',[binary]);assert not error and actual!=expected
print('PASS',len(cases),'real-signature stand-in cases; sanitizer-clean self-argument mutant caught by bytes')
