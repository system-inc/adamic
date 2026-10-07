"""Compare native repair predicates with unchanged Go repair helpers."""
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
virtual=repo/'cohere/adamic_wave01_await_repair.go'
overlay=out/'overlay.json'
overlay.write_text(json.dumps({'Replace':{str(virtual):str(source/'await_repair_main.go'),str(repo/'cohere/internal/lint/rules/core/adamic_wave01_repair.go'):str(source/'await_repair_capture.go')}}))
oracle=out/'oracle';run('go-build',['go','build','-overlay',overlay,'-o',oracle,virtual],repo/'cohere')
data,error=run('go',[oracle]);assert not error
cases=json.loads(data);assert len(cases)==24
module=source.parent/'require_await/repair.a'
code=['import { asyncWhitespace, awaitReplacement } from '+json.dumps(str(module))+';']
expected=[]
for i,c in enumerate(cases):
    args=','.join(json.dumps(c[k],ensure_ascii=False) for k in ['After','PreviousProperty','PreviousInitializer','PreviousStatement','PreviousText'])
    code.append('console.log(`'+str(i)+' ${asyncWhitespace('+json.dumps(c['After'],ensure_ascii=False)+')}\\t${awaitReplacement('+args+')}`);')
    expected.append(f"{i} {c['ExpectedWhitespace']}\t{c['Replacement']}\n")
entry=out/'entry.a';entry.write_text('\n'.join(code)+'\n')
expected=''.join(expected).encode();(out/'expected.stdout').write_bytes(expected)
binary=out/'native';run('native-build',[compiler,'build',entry,'-o',binary,'--sanitize'])
actual,error=run('native',[binary]);assert not error and actual==expected
mutant=out/'mutant.a';mutant.write_text(module.read_text().replace('code === 36','false'))
mutant_entry=out/'mutant-entry.a';mutant_entry.write_text(entry.read_text().replace(str(module),str(mutant)))
binary=out/'mutant';run('mutant-build',[compiler,'build',mutant_entry,'-o',binary,'--sanitize'])
actual,error=run('mutant',[binary]);assert not error and actual!=expected
print('PASS',len(cases),'Go whitespace and semicolon decisions; sanitizer-clean identifier-boundary mutant caught by bytes')
