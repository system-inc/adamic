"""Compare native require-await rendering with unchanged Go checkRequireAwait."""
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
virtual=repo/'cohere/adamic_wave01_await_report.go'
overlay=out/'overlay.json'
overlay.write_text(json.dumps({'Replace':{str(virtual):str(source/'await_report_main.go'),str(repo/'cohere/internal/lint/rules/core/adamic_wave01_await_report.go'):str(source/'await_report_capture.go')}}))
oracle=out/'oracle';run('go-build',['go','build','-overlay',overlay,'-o',oracle,virtual],repo/'cohere')
data,error=run('go',[oracle]);assert not error
cases=json.loads(data);assert len(cases)==15
module=source.parent/'require_await/report.a'
code=['import { awaitDescription, awaitDiagnostic } from '+json.dumps(str(module))+';']
def written(value):
    if not isinstance(value,str):return str(value)
    data=value.encode('utf-16-le',errors='surrogatepass')
    result=''
    for i in range(0,len(data),2):
        unit=int.from_bytes(data[i:i+2],'little')
        result+=chr(unit) if 32<=unit<=126 and unit!=92 else '\\u'+format(unit,'04x')
    return result
expected=[]
for c in cases:
    desc='awaitDescription('+','.join(json.dumps(c[k],ensure_ascii=False) for k in ['Arrow','Method','Constructor','Named','Name'])+')'
    args=[json.dumps(c['Start']),json.dumps(c['End']),desc,json.dumps(c['AsyncStart']),json.dumps(c['AsyncEnd']),json.dumps(c['Replacement'])]
    code.append('console.log(awaitDiagnostic('+','.join(args)+').written());')
    expected.append('\t'.join(written(v) for v in c['Expected'])+'\n')
entry=out/'entry.a';entry.write_text('\n'.join(code)+'\n')
expected=''.join(expected).encode();(out/'expected.stdout').write_bytes(expected)
binary=out/'native';run('native-build',[compiler,'build',entry,'-o',binary,'--sanitize'])
actual,error=run('native',[binary]);assert not error and actual==expected
mutant=out/'mutant.a';text=module.read_text().replace("'removeAsync'", "'removeAsynchronous'")
for name in ['diagnostic','suggestion','repair']:
    text=text.replace('../../'+name+'.ts',str(source.parent.parent/(name+'.ts')))
mutant.write_text(text)
mutant_entry=out/'mutant-entry.a';mutant_entry.write_text(entry.read_text().replace(str(module),str(mutant)))
binary=out/'mutant';run('mutant-build',[compiler,'build',mutant_entry,'-o',binary,'--sanitize'])
actual,error=run('mutant',[binary]);assert not error and actual!=expected
print('PASS',len(cases),'Go findings with exact suggestions; sanitizer-clean suggestion-ID mutant caught by bytes')
