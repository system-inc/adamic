"""Compare native structural predicates with unchanged production Go predicates."""
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
virtual=repo/'cohere/adamic_wave01_syntax.go'
overlay=out/'overlay.json'
overlay.write_text(json.dumps({'Replace':{str(virtual):str(source/'atomic_syntax_main.go'),str(repo/'cohere/internal/lint/rules/core/adamic_wave01_syntax.go'):str(source/'atomic_syntax_capture.go')}}))
oracle=out/'oracle';run('go-build',['go','build','-overlay',overlay,'-o',oracle,virtual],repo/'cohere')
data,error=run('go',[oracle]);assert not error
cases=json.loads(data)
module=source.parent/'require_atomic_updates/syntax.a'
code=['import { AtomicSyntaxNode, AtomicSyntax } from '+json.dumps(str(module))+';']
expected=[]
for i,c in enumerate(cases):
    code.append('function case'+str(i)+'(): void {')
    code.append('const nodes: AtomicSyntaxNode[] = ['+','.join('new AtomicSyntaxNode('+','.join(json.dumps(v) for v in n)+')' for n in c['Nodes'])+'];')
    code.append('const syntax = new AtomicSyntax(nodes);')
    code.append('for(const index of '+json.dumps([r[0] for r in c['Results']])+') {')
    code.append('console.log(`'+str(i)+' ${index} ${syntax.declarationName(index)} ${syntax.assignment(index, false, true) >= 0} ${syntax.assignment(index, true, false)}`);')
    code.append('}}')
    for index,declaration,plain,assignment in c['Results']:
        expected.append(f'{i} {index} {str(declaration).lower()} {str(plain).lower()} {assignment}\n')
code.extend('case'+str(i)+'();' for i in range(len(cases)))
entry=out/'entry.a';entry.write_text('\n'.join(code)+'\n')
expected=''.join(expected).encode();(out/'expected.stdout').write_bytes(expected)
binary=out/'native';run('native-build',[compiler,'build',entry,'-o',binary,'--sanitize'])
actual,error=run('native',[binary]);assert not error and actual==expected
mutant=out/'mutant.a';mutant.write_text(module.read_text().replace('if(parent.memberObject !== current) {','if(false) {'))
mutant_entry=out/'mutant-entry.a';mutant_entry.write_text(entry.read_text().replace(str(module),str(mutant)))
binary=out/'mutant';run('mutant-build',[compiler,'build',mutant_entry,'-o',binary,'--sanitize'])
actual,error=run('mutant',[binary]);assert not error and actual!=expected
print('PASS',len(cases),'parsed controls,',len(expected.splitlines()),'identifier predicates; sanitizer-clean member-name mutant caught by Go bytes')
