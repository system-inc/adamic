"""Compare native promise-demand decisions with Go's real checker predicate."""
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
virtual=repo/'cohere/adamic_wave01_await_demand.go'
overlay=out/'overlay.json'
overlay.write_text(json.dumps({'Replace':{str(virtual):str(source/'await_demand_main.go'),str(repo/'cohere/internal/lint/rules/core/adamic_wave01_demand.go'):str(source/'await_demand_capture.go')}}))
oracle=out/'oracle';run('go-build',['go','build','-overlay',overlay,'-o',oracle,virtual],repo/'cohere')
control=out/'controls.a';control.write_text('''
const promise: () => Promise<number> = async () => 1;
const mixed: () => number | Promise<number> = async () => 1;
const voidPosition: () => void = async () => 1;
const unknownPosition: () => unknown = async () => 1;
const anyPosition: () => any = async () => 1;
const unconstrained = async () => 1;
interface PromiseOverloads {(): Promise<number>; (n:number): Promise<string>}
const overloaded: PromiseOverloads = async () => 1;
interface MixedOverloads {(): Promise<number>; (n:number): number}
const mixedOverloads: MixedOverloads = async () => 1;
const promiseUnion: (() => Promise<number>) | (() => Promise<string>) = async () => 1;
const mixedUnion: (() => Promise<number>) | (() => number) = async () => 1;
interface Thenable {then(callback:(value:number)=>unknown):unknown}
const thenable: () => Thenable = async () => 1;
const numberPosition: () => number = async () => 1;
const expression: () => Promise<number> = async function named(){return 1;};
interface Adapter {run(): Promise<number>}
const adapter: Adapter = {async run(){return 1;}};
export {};
''')
config=out/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ESNext','lib':['esnext']},'files':[str(control)]}))
data,error=run('go',[oracle,config,control]);assert not error
cases=json.loads(data);assert len(cases)==14
assert any(c['Expected'] for c in cases) and any(not c['Expected'] for c in cases)
module=source.parent/'require_await/demand.a'
code=['import { awaitDemandsPromise } from '+json.dumps(str(module))+';']
expected=[]
for i,c in enumerate(cases):
    code.append('console.log(`'+str(i)+' ${awaitDemandsPromise('+json.dumps(c['Signatures'] or [])+')}`);')
    expected.append(f"{i} {str(c['Expected']).lower()}\n")
entry=out/'entry.a';entry.write_text('\n'.join(code)+'\n')
expected=''.join(expected).encode();(out/'expected.stdout').write_bytes(expected)
binary=out/'native';run('native-build',[compiler,'build',entry,'-o',binary,'--sanitize'])
actual,error=run('native',[binary]);assert not error and actual==expected
mutant=out/'mutant.a';mutant.write_text(module.read_text().replace('if(!thenable) {','if(false) {'))
mutant_entry=out/'mutant-entry.a';mutant_entry.write_text(entry.read_text().replace(str(module),str(mutant)))
binary=out/'mutant';run('mutant-build',[compiler,'build',mutant_entry,'-o',binary,'--sanitize'])
actual,error=run('mutant',[binary]);assert not error and actual!=expected
print('PASS',len(cases),'real-checker promise-demand controls; sanitizer-clean mixed-return mutant caught by bytes')
