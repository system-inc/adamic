#!/usr/bin/env python3
"""Go HIR versus native control analysis, isolated from the shared lint harness."""
import hashlib,json,os,subprocess,sys,time
from pathlib import Path
source=Path(__file__).resolve().parent;repo=source.parents[3]
d=Path(sys.argv[1]).resolve();d.mkdir(parents=True,exist_ok=True)
compiler=Path(os.environ.get('ADAMIC_COMPILER','/workspace/wave29-regex-controls/adamic'))
commands=[]
def run(name,args,cwd=repo,env=None):
 start=time.monotonic_ns()
 with (d/(name+'.stdout')).open('wb') as out,(d/(name+'.stderr')).open('wb') as err:
  result=subprocess.run([str(x) for x in args],cwd=cwd,env=env,stdout=out,stderr=err)
 commands.append(dict(name=name,args=[str(x) for x in args],exit=result.returncode,elapsed_ns=time.monotonic_ns()-start))
 (d/'commands.json').write_text(json.dumps(commands,indent=2)+'\n')
 if result.returncode:raise RuntimeError(f'{name}: exit {result.returncode}, see {d/(name+".stderr")}')
 return (d/(name+'.stdout')).read_bytes()
virtual=repo/'cohere/adamic_wave29_control_oracle.go'
(d/'overlay.json').write_text(json.dumps({'Replace':{str(virtual):str(source/'testdata/control_oracle.go')}}))
run('go-build',['go','build','-overlay',d/'overlay.json','-o',d/'oracle',virtual],cwd=repo/'cohere')
run('go',[d/'oracle',d,repo/'cohere/internal/lint/ecmascript/high_level_intermediate_representation/postdominator_test.go']);truth=(d/'go.expected').read_bytes();frames=d/'graphs.frames'
run('native-build',[compiler,'build',source/'control_core.a','-o',d/'native'])
assert run('native',[d/'native',frames])==truth
assert not (d/'native.stderr').read_bytes()
assert run('node',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',source/'control_core.a',frames])==truth
assert not (d/'node.stderr').read_bytes()
javascript=run('javascript',[compiler,'js',source/'control_core.a'])
(d/'control_core.mjs').write_bytes(javascript)
assert run('javascript-node',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',d/'control_core.mjs',frames])==truth
assert not (d/'javascript.stderr').read_bytes() and not (d/'javascript-node.stderr').read_bytes()
run('asan-build',[compiler,'build',source/'control_core.a','-o',d/'asan','--sanitize'])
env=dict(os.environ,ASAN_OPTIONS='detect_leaks=1:halt_on_error=1',UBSAN_OPTIONS='halt_on_error=1')
assert run('asan',[d/'asan',frames],env=env)==truth
assert not (d/'asan.stderr').read_bytes()
kernel=(source/'post_dominators.a').read_text();driver=(source/'control_core.a').read_text()
for name,before,after in [
 ('throw-exit',"if(block.terminal === 'Return')", "if(block.terminal === 'Return' || block.terminal === 'Throw')"),
 ('switch-case','for(const test of block.tests)', 'for(const test of block.tests.slice(0, 1))'),
]:
 assert kernel.count(before)==1
 local=d/('mutant-'+name);local.mkdir(exist_ok=True)
 (local/'post_dominators.a').write_text(kernel.replace(before,after))
 (local/'control_core.a').write_text(driver.replace("'../frames.ts'",repr(str(source.parent/'frames.ts'))))
 run(name+'-build',[compiler,'build',local/'control_core.a','-o',local/'native'])
 got=run(name,[local/'native',frames])
 assert got!=truth and not(d/(name+'.stderr')).read_bytes()
 differences=sum(a!=b for a,b in zip(truth.splitlines(),got.splitlines()))
 print(f'{name}: compiled, exit 0, empty stderr; only bytes catch {differences} changed graph results',flush=True)
print('PASS:',(d/'go.stdout').read_text().strip(),'sha256='+hashlib.sha256(truth).hexdigest(),flush=True)
print('Go/native/Node/emitted JavaScript/sanitized native agree; these are control graphs, not full rule findings',flush=True)
