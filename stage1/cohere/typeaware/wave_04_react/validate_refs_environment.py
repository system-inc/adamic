#!/usr/bin/env python3
"""Independent production Go environment traces versus native and Node."""
from pathlib import Path
import argparse,json,os,subprocess,hashlib
p=argparse.ArgumentParser();p.add_argument('directory',type=Path);p.add_argument('--adamic',default='/workspace/typeaware-wave-04-landing-current/adamic');a=p.parse_args();own=Path(__file__).resolve().parent;repo=own.parents[3];out=a.directory.resolve();out.mkdir(parents=True,exist_ok=True);runs=[]
def run(label,args,cwd=repo,env=None):
 with (out/(label+'.stdout')).open('wb') as so,(out/(label+'.stderr')).open('wb') as se:
  result=subprocess.run([str(x) for x in args],cwd=cwd,env=env,stdout=so,stderr=se)
 actual=(out/(label+'.stdout')).read_bytes();error=(out/(label+'.stderr')).read_bytes();runs.append(dict(label=label,exit=result.returncode,stdout_bytes=len(actual),stderr_bytes=len(error)))
 if result.returncode:raise RuntimeError(str(runs[-1]))
 return actual,error
module=own/'refs_environment.a';overlay=out/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(repo/'cohere/internal/lint/rules/react/wave04_refs_environment_test.go'):str(own/'testdata/refs_environment_test.go')}}))
run('go',['go','test','-overlay',overlay,'./internal/lint/rules/react','-run','^TestWave04RefsEnvironment$','-count=1','-v'],cwd=repo/'cohere',env=dict(os.environ,WAVE04_REFS_ENV=str(out/'controls.json')))
data=json.loads((out/'controls.json').read_text());truth=data['Expected'].encode();(out/'expected.txt').write_bytes(truth)
driver="import { RefBinding, RefEnvironment } from '"+str(module)+"';\nimport { RefValue } from '"+str(own/'refs_value.a')+"';\n"
def boolean(x):return str(x).lower()
bindings=', '.join('new RefBinding('+json.dumps(x['Name'])+','+str(x['Declaration'])+','+boolean(x['Present'])+')' for x in data['Bindings'])
for si,ops in enumerate(data['Scenarios']):
 driver+='{\nconst env = new RefEnvironment(['+bindings+']);\n'
 for i,(kind,rid,has,span,has_span) in enumerate([(0,0,False,0,False),(3,11,True,0,False),(3,12,True,0,False),(4,11,True,9,True)]):
  driver+=f"const v{i} = new RefValue({kind}); v{i}.refId = {rid}; v{i}.hasRefId = {boolean(has)}; v{i}.span = {span}; v{i}.hasSpan = {boolean(has_span)}; env.arena.add(v{i});\n"
 for step,op in enumerate(ops):
  k=op['Key'];v=op['Value'];kind=op['Kind'];txt=json.dumps(op['Text'])
  expression={'set':f'env.set({k},{v})','reset':'env.changed = false','note':f'env.noteDeclaration({k})','define':f'env.define({k},{v})','access':f'env.accessNodes.set({k},{v})','carry':f'env.carryAccessNode({k},{v})','name':f'env.setName({k},{v})','name2':f'env.setName2({k},{txt})','property':f'env.setPropertyName({k},{txt})','mint':'env.nextRefId()'}[kind]
  driver+=expression+';\n'+f"console.log(`step {si} {step} ${{env.changed}} ${{env.refIdSeed}}`);\n"
  driver+='for(let id = 0; id < 10; id++) { console.log(`key ${id} ${env.operandId(id)} ${env.arena.signature(env.get(id))} ${env.nameOf(id)} ${env.propertyNames.get(id) ?? \'\'} ${env.accessNodes.get(id) ?? -1}`); }\n'
 driver+='}\n'
entry=out/'probe.a';entry.write_text(driver)
for variant in ['normal','asan']:
 command=[a.adamic,'build',entry,'-o',out/variant]
 if variant=='asan':command+=['--sanitize']
 run('build-'+variant,command);actual,error=run('run-'+variant,[out/variant])
 if actual!=truth or error:raise RuntimeError('environment differs '+variant)
actual,error=run('node',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',entry])
if actual!=truth or error:raise RuntimeError('source Node differs')
js,_=run('emit',[a.adamic,'js',entry]);(out/'probe.js').write_bytes(js);actual,error=run('js',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',out/'probe.js'])
if actual!=truth or error:raise RuntimeError('emitted JS differs')
mutants=[('alias','return this.temporaries.get(key) ?? key;','return key;'),('declaration','this.byDeclaration.has(declaration)','this.byDeclaration.has(declaration + 1000)'),('changed','this.changed = true;','this.changed = false;'),('initial-none','this.data.set(resolved,widened); return;','this.data.set(resolved,widened); this.changed = true; return;'),('widen','this.arena.join(value,current)','this.arena.join(current,current)'),('access','this.accessNodes.set(key,node)','this.accessNodes.set(key,-1)')]
for label,before,after in mutants:
 source=module.read_text().replace("from './refs_value.a'","from '"+str(own/'refs_value.a')+"'")
 if source.count(before)!=1:raise RuntimeError('nonunique mutant '+label)
 path=out/(label+'.a');path.write_text(source.replace(before,after));probe=out/(label+'-probe.a');probe.write_text(driver.replace(str(module),str(path)))
 run(label+'-build',[a.adamic,'build',probe,'-o',out/(label+'-native'),'--sanitize']);actual,error=run(label+'-run',[out/(label+'-native')])
 if actual==truth or error or len(actual.splitlines())!=len(truth.splitlines()):raise RuntimeError('mutant survived or failed outside comparison '+label)
(out/'runs.json').write_text(json.dumps(runs,indent=2)+'\n');(out/'source-sha256.json').write_text(json.dumps({str(x.relative_to(own)):hashlib.sha256(x.read_bytes()).hexdigest() for x in [module,own/'refs_value.a',own/'testdata/refs_environment_test.go',Path(__file__).resolve()]},indent=2)+'\n')
print(f'PASS partial refs environment: {len(truth.splitlines())} records, {len(truth)} bytes; native, sanitizers, source Node, emitted JS; six comparison-only mutants.')
