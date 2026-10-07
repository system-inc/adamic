"""Compare native head ranges with unchanged Go head selection."""
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
virtual=repo/'cohere/adamic_wave01_await_head.go'
overlay=out/'overlay.json'
overlay.write_text(json.dumps({'Replace':{str(virtual):str(source/'await_head_main.go'),str(repo/'cohere/internal/lint/rules/core/adamic_wave01_head.go'):str(source/'await_head_capture.go')}}))
oracle=out/'oracle';run('go-build',['go','build','-overlay',overlay,'-o',oracle,virtual],repo/'cohere')
data,error=run('go',[oracle]);assert not error
cases=json.loads(data);assert len(cases)==32
module=source.parent/'require_await/head.a'
code=['import { AwaitModifier, awaitHead } from '+json.dumps(str(module))+';',
      'import { asyncWhitespace, awaitReplacement } from '+json.dumps(str(source.parent/'require_await/repair.a'))+';',
      'import { awaitDescription, awaitDiagnostic } from '+json.dumps(str(source.parent/'require_await/report.a'))+';']
expected=[]
def written(value):
    if not isinstance(value,str):return str(value)
    data=value.encode('utf-16-le',errors='surrogatepass')
    result=''
    for offset in range(0,len(data),2):
        unit=int.from_bytes(data[offset:offset+2],'little')
        result+=chr(unit) if 32<=unit<=126 and unit!=92 else '\\u'+format(unit,'04x')
    return result
for i,c in enumerate(cases):
    args=[json.dumps(c[k]) for k in ['Bytes','NodeStart','NodeEnd','ParentStart','ArrowStart','ArrowEnd','NameEnd']]
    modifiers='['+','.join('new AwaitModifier('+json.dumps(text,ensure_ascii=False)+','+str(end)+')' for text,end in (c['Modifiers'] or []))+']'
    code.append('const modifiers'+str(i)+': AwaitModifier[] = '+modifiers+';')
    args.append('modifiers'+str(i))
    code.append('const head'+str(i)+' = awaitHead('+','.join(args)+');')
    code.append('console.log(`'+str(i)+' ${head'+str(i)+'.start} ${head'+str(i)+'.end}`);')
    expected.append(f"{i} {c['Start']} {c['End']}\n")
    code.append('let keyword'+str(i)+' = -1;')
    code.append("for(const modifier of modifiers"+str(i)+") { if(modifier.text === 'async') {keyword"+str(i)+" = modifier.end;}}")
    desc='awaitDescription('+','.join(json.dumps(c[k],ensure_ascii=False) for k in ['Arrow','Method','Constructor','Named','Name'])+')'
    suffix=json.dumps(c['After'],ensure_ascii=False)
    replacement='awaitReplacement('+','.join(json.dumps(c[k],ensure_ascii=False) for k in ['After','PreviousProperty','PreviousInitializer','PreviousStatement','PreviousText'])+')'
    arguments=['head'+str(i)+'.start','head'+str(i)+'.end',desc,'keyword'+str(i)+' < 0 ? -1 : keyword'+str(i)+' - 5','keyword'+str(i)+' + asyncWhitespace('+suffix+')',replacement]
    code.append('console.log(awaitDiagnostic('+','.join(arguments)+').written());')
    expected.append('\t'.join(written(value) for value in c['Expected'])+'\n')

entry=out/'entry.a';entry.write_text('\n'.join(code)+'\n')
expected=''.join(expected).encode();(out/'expected.stdout').write_bytes(expected)
binary=out/'native';run('native-build',[compiler,'build',entry,'-o',binary,'--sanitize'])
actual,error=run('native',[binary]);assert not error and actual==expected
mutant=out/'mutant.a';mutant.write_text(module.read_text().replace('if(parentStart >= 0) {','if(false) {'))
mutant_entry=out/'mutant-entry.a';mutant_entry.write_text(entry.read_text().replace(str(module),str(mutant)))
binary=out/'mutant';run('mutant-build',[compiler,'build',mutant_entry,'-o',binary,'--sanitize'])
actual,error=run('mutant',[binary]);assert not error and actual!=expected
print('PASS',len(cases),'Go head ranges and full finding/suggestion bytes; sanitizer-clean property-parent mutant caught by bytes')
