#!/usr/bin/env python3
"""Compare the supplied-graph kernel only; not an end-to-end rule harness."""
import hashlib, json, os, subprocess, sys, time
from pathlib import Path
source=Path(__file__).resolve().parent
repo=source.parents[3]
output=Path(sys.argv[1]).resolve();output.mkdir(parents=True,exist_ok=True)
compiler=Path(os.environ.get('ADAMIC_COMPILER','/workspace/wave29-regex-controls/adamic'))
commands=[]
def run(name,args,cwd=repo,env=None):
 start=time.monotonic_ns()
 with (output/(name+'.stdout')).open('wb') as out,(output/(name+'.stderr')).open('wb') as err:
  result=subprocess.run([str(x) for x in args],cwd=cwd,env=env,stdout=out,stderr=err)
 commands.append(dict(name=name,args=[str(x) for x in args],exit=result.returncode,elapsed_ns=time.monotonic_ns()-start))
 (output/'commands.json').write_text(json.dumps(commands,indent=2)+'\n')
 if result.returncode:raise RuntimeError(f'{name}: exit {result.returncode}; see its logs')
 return (output/(name+'.stdout')).read_bytes()
virtual=repo/'cohere/internal/lint/rules/react/adamic_wave29_static_core_test.go'
(output/'overlay.json').write_text(json.dumps({'Replace':{str(virtual):str(source/'testdata/static_core_test.go')}}))
env=dict(os.environ,ADAMIC_WAVE29_STATIC_CORE=str(output))
run('go',['go','test','-overlay',output/'overlay.json','./internal/lint/rules/react','-run','^TestWave29StaticCore$','-count=1','-v'],cwd=repo/'cohere',env=env)
truth=(output/'go.expected').read_bytes();frames=output/'graphs.frames'
run('native-build',[compiler,'build',source/'static_core.a','-o',output/'native'])
actual=run('native',[output/'native',frames])
assert actual==truth, 'native graph diagnostics differ from production Go'
assert not (output/'native.stderr').read_bytes(), 'unexpected native stderr'
run('asan-build',[compiler,'build',source/'static_core.a','-o',output/'asan','--sanitize'])
sanenv=dict(os.environ,ASAN_OPTIONS='detect_leaks=1:halt_on_error=1',UBSAN_OPTIONS='halt_on_error=1')
assert run('asan',[output/'asan',frames],env=sanenv)==truth
assert not (output/'asan.stderr').read_bytes(), 'sanitizer diagnostics'
kernel=(source/'static_components.a').read_text()
driver=(source/'static_core.a').read_text()
mutants=[('store-binding','dynamic.set(instruction.binding, creator);','dynamic.set(instruction.binding, instruction.binding);'),('phi','dynamic.set(phi.target, creator);','dynamic.set(phi.target, phi.target);')]
for name,before,after in mutants:
 assert kernel.count(before)==1
 directory=output/('mutant-'+name);directory.mkdir(exist_ok=True)
 text=kernel.replace(before,after).replace("'../../../typescript/scanner/scanner.ts'",repr(str(repo/'stage1/typescript/scanner/scanner.ts')))
 (directory/'static_components.a').write_text(text)
 text=driver.replace("'../frames.ts'",repr(str(source.parent/'frames.ts'))).replace("'../../../typescript/parser/nodes.ts'",repr(str(repo/'stage1/typescript/parser/nodes.ts')))
 (directory/'static_core.a').write_text(text)
 run(name+'-build',[compiler,'build',directory/'static_core.a','-o',directory/'native'])
 got=run(name,[directory/'native',frames])
 assert not (output/(name+'.stderr')).read_bytes(), 'mutant must execute normally'
 assert got!=truth, 'comparison did not catch compiling mutant'
 differences=[i for i,(a,b) in enumerate(zip(truth.splitlines(),got.splitlines())) if a!=b]
 print(f'{name}: compiled and exited 0, comparison caught {len(differences)} differing lines')
print(f'PASS: {len((output/"cases.tsv").read_text().splitlines())} supplied graphs; {sum(b"\treact-hooks/static-components\t" in line for line in truth.splitlines())} findings; {len(truth)} bytes; sha256={hashlib.sha256(truth).hexdigest()}')
