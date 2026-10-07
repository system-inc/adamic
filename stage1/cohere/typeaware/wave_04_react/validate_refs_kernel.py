#!/usr/bin/env python3
"""Independent Go oracle for partial refs equality and naming kernels."""
from pathlib import Path
import argparse,json,os,subprocess,hashlib
p=argparse.ArgumentParser();p.add_argument('directory',type=Path);p.add_argument('--adamic',default='/workspace/typeaware-wave-04/final/adamic');a=p.parse_args()
own=Path(__file__).resolve().parent;repo=own.parents[3];out=a.directory.resolve();out.mkdir(parents=True,exist_ok=True);runs=[]
def run(label,args,cwd=repo,env=None):
 with (out/(label+'.stdout')).open('wb') as so,(out/(label+'.stderr')).open('wb') as se:
  proc=subprocess.run([str(x) for x in args],cwd=cwd,env=env,stdout=so,stderr=se)
 result=(out/(label+'.stdout')).read_bytes();error=(out/(label+'.stderr')).read_bytes()
 runs.append(dict(label=label,exit=proc.returncode,stdout_bytes=len(result),stderr_bytes=len(error)))
 if proc.returncode:raise RuntimeError(str(runs[-1]))
 return result,error
virtual=repo/'cohere/internal/lint/rules/react/wave04_refs_kernel_test.go'
overlay=out/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(own/'testdata/refs_kernel_test.go')}}))
run('go',['go','test','-overlay',overlay,'./internal/lint/rules/react','-run','^TestWave04RefsKernel$','-count=1','-v'],cwd=repo/'cohere',env=dict(os.environ,WAVE04_REFS_KERNEL=str(out/'controls.json')))
data=json.loads((out/'controls.json').read_text());truth=data['Expected'].encode();(out/'expected.txt').write_bytes(truth)
module=own/'refs_value.a'
driver="import { RefValue, RefFunction, RefValues } from '"+str(module)+"';\nconst arena = new RefValues();\n"
def boolean(x):return str(x).lower()
for i,row in enumerate(data['Values']):
 driver+=f"const v{i} = new RefValue({row['Kind']});\n"
 for field,key in [('refId','RefId'),('hasRefId','HasRefId'),('span','Span'),('hasSpan','HasSpan'),('value','Value'),('fn','Function')]:
  val=row[key];driver+=f"v{i}.{field} = {boolean(val) if isinstance(val,bool) else val};\n"
 driver+=f"arena.values.push(v{i});\n"
for i,row in enumerate(data['Functions']):
 driver+=f"const f{i} = new RefFunction(); f{i}.readRefEffect = {boolean(row['Effect'])}; f{i}.returnType = {row['Return']}; arena.functions.push(f{i});\n"
driver+="for(let i = -1; i < arena.values.length; i++) { for(let j = -1; j < arena.values.length; j++) { console.log(`equal ${i} ${j} ${arena.equal(i,j)}`); } }\n"
for i,name in enumerate(data['Names']):
 driver+=f"console.log(`name {i} ${{arena.hookName({json.dumps(name,ensure_ascii=False)})}} ${{arena.refName({json.dumps(name,ensure_ascii=False)})}}`);\n"
entry=out/'probe.a';entry.write_text(driver)
for variant in ['normal','asan']:
 cmd=[a.adamic,'build',entry,'-o',out/variant]
 if variant=='asan':cmd+=['--sanitize']
 run('build-'+variant,cmd);actual,error=run('run-'+variant,[out/variant])
 if actual!=truth or error:raise RuntimeError('kernel differs: '+variant)
actual,error=run('node',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',entry])
if actual!=truth or error:raise RuntimeError('source Node kernel differs')
js,error=run('emit',[a.adamic,'js',entry]);(out/'probe.js').write_bytes(js)
actual,error=run('js',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',out/'probe.js'])
if actual!=truth or error:raise RuntimeError('emitted JS kernel differs')
mutations=[('identity','case 3: return true;','case 3: return left.refId === right.refId;'),('guard','left.refId === right.refId && left.hasRefId === right.hasRefId','left.refId === right.refId'),('span','left.span === right.span && left.hasSpan === right.hasSpan','left.span === right.span'),('name','name.length <= 3','name.length < 3')]
for label,before,after in mutations:
 source=module.read_text()
 if source.count(before)!=1:raise RuntimeError('nonunique mutation '+label)
 path=out/(label+'.a');path.write_text(source.replace(before,after));mutant=out/(label+'-probe.a');mutant.write_text(driver.replace(str(module),str(path)))
 run(label+'-build',[a.adamic,'build',mutant,'-o',out/(label+'-native'),'--sanitize']);actual,error=run(label+'-run',[out/(label+'-native')])
 if error or actual==truth:raise RuntimeError('mutant survived or failed before comparison: '+label)
runfile=out/'runs.json';runfile.write_text(json.dumps(runs,indent=2)+'\n')
(out/'source-sha256.json').write_text(json.dumps({str(x.relative_to(own)):hashlib.sha256(x.read_bytes()).hexdigest() for x in [module,own/'testdata/refs_kernel_test.go',Path(__file__).resolve()]},indent=2)+'\n')
print(f'PASS partial refs kernel: {len(truth.splitlines())} records, {len(truth)} bytes; native, sanitizer, source Node, emitted JS; four comparison-only kernel mutants.')
