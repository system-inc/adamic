#!/usr/bin/env python3
"""Hold declared numeric listeners to pinned production Run maps."""
import json
import pathlib
import re
import subprocess
import time

ROOT = pathlib.Path(__file__).resolve().parents[4]
UNIT = pathlib.Path(__file__).resolve().parent
OUT = pathlib.Path('/workspace/wave-07-listeners')
OUT.mkdir(exist_ok=True)
records = []

def run(name, command):
    started = time.monotonic_ns()
    with (OUT/(name+'.stdout')).open('wb') as stdout, (OUT/(name+'.stderr')).open('wb') as stderr:
        result = subprocess.run([str(value) for value in command], cwd=ROOT, stdout=stdout, stderr=stderr)
    records.append(dict(name=name,command=[str(value) for value in command],exit=result.returncode,
                        elapsed_ns=time.monotonic_ns()-started))
    (OUT/'commands.json').write_text(json.dumps(records,indent=2)+'\n')
    assert result.returncode == 0, (name,result.returncode)
    return (OUT/(name+'.stdout')).read_bytes(), (OUT/(name+'.stderr')).read_bytes()

virtual=ROOT/'cohere/adamic_wave07_listeners.go'
overlay=OUT/'oracle.json'
overlay.write_text(json.dumps({'Replace':{str(virtual):str(UNIT/'testdata/listener_oracle.go')}}))
# The existing independent loader is compiled inside cohere, not the bridge.
with (OUT/'oracle-build.stdout').open('wb') as stdout,(OUT/'oracle-build.stderr').open('wb') as stderr:
    result=subprocess.run(['go','build','-overlay',str(overlay),'-o',str(OUT/'oracle'),str(virtual)],cwd=ROOT/'cohere',stdout=stdout,stderr=stderr)
assert result.returncode==0
(OUT/'input.a').write_text('// exit enables the conditional SourceFile listener.\nexport {};\n')
(OUT/'ambient.d.ts').write_text('export {};\n')
(OUT/'tsconfig.json').write_text('{"compilerOptions":{"target":"ES2022","strict":true},"files":["ambient.d.ts"]}')
(OUT/'manifest').write_text(str(OUT/'input.a')+'\n')
truth,_=run('go',[OUT/'oracle',OUT/'tsconfig.json',OUT/'manifest'])
assert len(truth.splitlines())==9
stage0=pathlib.Path('/workspace/wave-07-next-rest/adamic')
archive=pathlib.Path('/workspace/wave-07-next-rest/checker.a')
entry=UNIT/'listeners_probe.a'
run('native-build',[stage0,'build',entry,'-o',OUT/'native','--tsgo',archive])
actual,errors=run('native',[OUT/'native'])
assert actual==truth and not errors
print('numeric listener map: '+str(len(truth))+' identical production Go bytes',flush=True)
run('sanitized-build',[stage0,'build',entry,'-o',OUT/'native-asan','--tsgo','/workspace/wave-07-next-rest/checker-asan.a','--sanitize'])
actual,errors=run('sanitized',[OUT/'native-asan'])
assert actual==truth and not errors
print('numeric listener map: ASan/UBSan/LSan pass',flush=True)

imports=list(re.finditer(r"import \{ listenerKinds as (\w+) \} from '([^']+)';",entry.read_text()))
assert len(imports)==9

def absolute_imports(source,directory):
    return re.sub(r"(from )'([^']+)'",lambda match:match[1]+"'"+str((directory/match[2]).resolve())+"'" if match[2].startswith('.') else match[0],source)

for position,matched in enumerate(imports):
    original=(UNIT/matched[2]).resolve()
    source=original.read_text()
    token=re.search(r'(export const listenerKinds: readonly number\[\] = \[)(\d+)',source)
    assert token
    changed=source[:token.start(2)]+str(int(token[2])+1)+source[token.end(2):]
    mutant=OUT/('listener-mutant-'+str(position)+'.a')
    mutant.write_text(absolute_imports(changed,original.parent))
    probe=OUT/('listener-probe-'+str(position)+'.a')
    text=absolute_imports(entry.read_text(),UNIT)
    exact="from '"+str(original)+"'"
    assert text.count(exact)==1
    probe.write_text(text.replace(exact,"from '"+str(mutant)+"'"))
    binary=OUT/('mutant-'+str(position))
    run('mutant-'+str(position)+'-build',[stage0,'build',probe,'-o',binary,'--tsgo',archive])
    actual,errors=run('mutant-'+str(position),[binary])
    assert actual!=truth and not errors
    print(matched[1]+': valid numeric mutant exits 0, caught only by production Go bytes',flush=True)
print('PASS numeric listener declarations and all nine mutants',flush=True)
