"""Compare native diagnostic rendering with unchanged Go reportNonAtomicUpdate."""
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
virtual=repo/'cohere/adamic_wave01_report.go'
overlay=out/'overlay.json'
overlay.write_text(json.dumps({'Replace':{str(virtual):str(source/'atomic_report_main.go'),str(repo/'cohere/internal/lint/rules/core/adamic_wave01_report.go'):str(source/'atomic_report_capture.go')}}))
oracle=out/'oracle';run('go-build',['go','build','-overlay',overlay,'-o',oracle,virtual],repo/'cohere')
data,error=run('go',[oracle]);assert not error
cases=json.loads(data)
module=source.parent/'require_atomic_updates/report.a'
code=['import { atomicDiagnostic } from '+json.dumps(str(module))+';']
def written(text):
    data=text.encode('utf-16-le',errors='surrogatepass')
    result=''
    for i in range(0,len(data),2):
        unit=int.from_bytes(data[i:i+2],'little')
        result+=chr(unit) if 32<=unit<=126 and unit!=92 else '\\u'+format(unit,'04x')
    return result
expected=[]
for c in cases:
    args=[c['Start'],c['End'],c['Name'],c['Target'],c['Property']]
    code.append('console.log(atomicDiagnostic('+','.join(json.dumps(v,ensure_ascii=False) for v in args)+').written());')
    expected.append(f"{c['Start']}\t{c['End']}\trequire-atomic-updates\t{c['ID']}\t{written(c['Message'])}\t0\t0\n")
entry=out/'entry.a';entry.write_text('\n'.join(code)+'\n')
expected=''.join(expected).encode();(out/'expected.stdout').write_bytes(expected)
binary=out/'native';run('native-build',[compiler,'build',entry,'-o',binary,'--sanitize'])
actual,error=run('native',[binary]);assert not error and actual==expected
mutant=out/'mutant.a';mutant.write_text(module.read_text().replace("'nonAtomicObjectUpdate'", "'nonAtomicUpdate'").replace('../../diagnostic.ts',str(source.parent.parent/'diagnostic.ts')))
mutant_entry=out/'mutant-entry.a';mutant_entry.write_text(entry.read_text().replace(str(module),str(mutant)))
binary=out/'mutant';run('mutant-build',[compiler,'build',mutant_entry,'-o',binary,'--sanitize'])
actual,error=run('mutant',[binary]);assert not error and actual!=expected
print('PASS',len(cases),'Go diagnostics, exact spans/messages/fix and suggestion counts; sanitizer-clean property-ID mutant caught by bytes')
