#!/usr/bin/env python3
"""Go oracle for partial refs predicates and finding metadata."""
from pathlib import Path
import argparse,json,os,subprocess,hashlib
p=argparse.ArgumentParser();p.add_argument('directory',type=Path);p.add_argument('--adamic',default='/workspace/typeaware-wave-04-landing-current/adamic');a=p.parse_args();own=Path(__file__).resolve().parent;repo=own.parents[3];out=a.directory.resolve();out.mkdir(parents=True,exist_ok=True);runs=[]
def run(label,args,cwd=repo,env=None):
 with (out/(label+'.stdout')).open('wb') as so,(out/(label+'.stderr')).open('wb') as se:
  result=subprocess.run([str(x) for x in args],cwd=cwd,env=env,stdout=so,stderr=se)
 actual=(out/(label+'.stdout')).read_bytes();error=(out/(label+'.stderr')).read_bytes();runs.append(dict(label=label,exit=result.returncode,stdout_bytes=len(actual),stderr_bytes=len(error)))
 if result.returncode:raise RuntimeError(str(runs[-1]))
 return actual,error
module=own/'refs_checks.a';overlay=out/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(repo/'cohere/internal/lint/rules/react/wave04_refs_checks_test.go'):str(own/'testdata/refs_checks_test.go')}}))
run('go',['go','test','-overlay',overlay,'./internal/lint/rules/react','-run','^TestWave04RefsChecks$','-count=1','-v'],cwd=repo/'cohere',env=dict(os.environ,WAVE04_REFS_CHECKS=str(out/'controls.json')))
data=json.loads((out/'controls.json').read_text());truth=data['Expected'].encode();(out/'expected.txt').write_bytes(truth)
driver="import { RefFinding, RefChecks } from '"+str(module)+"';\nimport { RefEnvironment } from '"+str(own/'refs_environment.a')+"';\nimport { RefValue, RefFunction } from '"+str(own/'refs_value.a')+"';\nconst checks = new RefChecks();\n"
for index in range(-1,len(data['Values'])):
 driver+='{\nconst env = new RefEnvironment([]);\n'
 for i,v in enumerate(data['Values']):
  driver+=f"const v{i} = new RefValue({v['Kind']}); v{i}.value = {v['Value']}; v{i}.fn = {v['Function']}; env.arena.add(v{i});\n"
 driver+='v3.refId = 11; v3.hasRefId = true; v4.refId = 11; v4.hasRefId = true; v4.span = 12; v4.hasSpan = true;\nconst f0 = new RefFunction(); f0.readRefEffect = true; env.arena.functions.push(f0); const f1 = new RefFunction(); f1.returnType = 4; env.arena.functions.push(f1);\n'
 if index>=0:driver+=f'env.data.set(7,{index});\n'
 driver+='env.accessNodes.set(7,0);\n'
 for seeded in [False,True]:
  seed=str(seeded).lower()
  for method in ['direct','value','passed','update']:
   driver+='{\nconst findings: RefFinding[] = [];\n'
   if seeded:driver+='findings.push(new RefFinding(3,99,false,-1));\n'
   driver+=f'checks.{method}(env,7,'+('1,' if method=='update' else '')+'findings);\n'
   driver+=f'console.log(`case {index} {seed} {method} ${{findings.length}}`);\nfor(const f of findings) {{ console.log(f.written()); }}\n}}\n'
 driver+='}\n'
entry=out/'probe.a';entry.write_text(driver)
for variant in ['normal','asan']:
 command=[a.adamic,'build',entry,'-o',out/variant]
 if variant=='asan':command+=['--sanitize']
 run('build-'+variant,command);actual,error=run('run-'+variant,[out/variant])
 if actual!=truth or error:raise RuntimeError('predicate bytes differ '+variant)
actual,error=run('node',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',entry])
if actual!=truth or error:raise RuntimeError('source Node differs')
js,_=run('emit',[a.adamic,'js',entry]);(out/'probe.js').write_bytes(js);actual,error=run('js',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',out/'probe.js'])
if actual!=truth or error:raise RuntimeError('emitted JS differs')
mutants=[('direct',"panic('missing direct value');\n        if(value.kind === 4)","panic('missing direct value');\n        if(value.kind === 3)"),('effect','if(value.kind === 5 && fn?.readRefEffect === true)','if(value.kind === 5 && fn?.readRefEffect === false)'),('passed','this.append(findings,1,id)','this.append(findings,0,id)'),('update','this.append(findings,2,id,node)','this.append(findings,2,id,-1)')]
for label,before,after in mutants:
 source=module.read_text().replace("from './refs_environment.a'","from '"+str(own/'refs_environment.a')+"'")
 if source.count(before)!=1:raise RuntimeError('nonunique mutant '+label)
 path=out/(label+'.a');path.write_text(source.replace(before,after));probe=out/(label+'-probe.a');probe.write_text(driver.replace(str(module),str(path)))
 run(label+'-build',[a.adamic,'build',probe,'-o',out/(label+'-native'),'--sanitize']);actual,error=run(label+'-run',[out/(label+'-native')])
 if actual==truth or error:raise RuntimeError('mutant survived or failed outside comparison '+label)
(out/'runs.json').write_text(json.dumps(runs,indent=2)+'\n');(out/'source-sha256.json').write_text(json.dumps({str(x.relative_to(own)):hashlib.sha256(x.read_bytes()).hexdigest() for x in [module,own/'refs_environment.a',own/'refs_value.a',own/'testdata/refs_checks_test.go',Path(__file__).resolve()]},indent=2)+'\n')
print(f'PASS partial refs predicates: 96 cases, {len(truth.splitlines())} records, {len(truth)} bytes; native, sanitizers, source Node, emitted JS; four comparison-only mutants.')
