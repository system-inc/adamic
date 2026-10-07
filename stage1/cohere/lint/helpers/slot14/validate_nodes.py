#!/usr/bin/env python3
import json,os,re,random,subprocess,tempfile,shutil,time
from pathlib import Path
owned=Path(__file__).resolve().parent
root=owned.parents[4];cohere=root/'cohere'
scratch=Path(tempfile.mkdtemp(prefix='helper14-nodes-'));evidence=owned/'evidence';evidence.mkdir(exist_ok=True)
log=(evidence/'nodes.log').open('w',buffering=1)
def note(s):print(s,file=log,flush=True)
def run(args,cwd=root,env=None,target=None,allow_failure=False):
 out=target or scratch/('out-'+str(time.time_ns()));err=scratch/('err-'+str(time.time_ns()))
 with out.open('wb') as o,err.open('wb') as e:p=subprocess.run(list(map(str,args)),cwd=cwd,env=env,stdout=o,stderr=e)
 if not allow_failure and (p.returncode or err.stat().st_size):
  note('FAILED '+repr(args)+f' exit={p.returncode}');note(err.read_text());note(out.read_text()[:5000]);raise RuntimeError('command failed')
 return out.read_bytes(),p.returncode,err.read_bytes()
def overlay(mapping,name):
 p=scratch/(name+'.json');p.write_text(json.dumps({'Replace':{str(k):str(v) for k,v in mapping.items()}}));return p
try:
 note('scratch='+str(scratch))
 cases=scratch/'cases.json';wantfile=scratch/'want'
 env=dict(os.environ,SLOT14_CASES=str(cases),SLOT14_OUTPUT=str(wantfile))
 oracle=overlay({cohere/'internal/lint/rules/tailwind/collapse/slot14_nodes_test.go':owned/'nodes_oracle_test.go.txt'},'oracle')
 run(['go','test','-overlay='+str(oracle),'./internal/lint/rules/tailwind/collapse','-run','^TestSlot14Nodes$','-count=1','-v','-timeout=10m'],cwd=cohere,env=env,target=evidence/'nodes-oracle.log')
 builder=scratch/'builder';build=overlay({root/'slot14_build.go':owned/'build.go.txt'},'build')
 run(['go','build','-overlay='+str(build),'-o',builder,root/'slot14_build.go'])
 runner=scratch/'nodes_runner.a';runner.write_text((owned/'withdrawn_nodes_runner.a.txt').read_text().replace('../options_json.ts','./options_json.ts'))
 shutil.copy2(owned.parent/'options_json.ts',scratch/'options_json.ts')
 (scratch/'nodes_from_static_declarations.a').write_text((owned/'withdrawn_nodes_from_static_declarations.a.txt').read_text())
 native=scratch/'native';js=scratch/'emitted.mjs'
 run([builder,runner,native,js,'sanitized'])
 node=['node','--disable-warning=ExperimentalWarning',root/'oracle/node.mjs'];want=wantfile.read_bytes()
 for side,args in [('Node',node+[runner,cases]),('native',[native,cases]),('emitted JavaScript',node+[js,cases])]:
  got,_,_=run(args);assert got==want,(side,'byte mismatch');note(side+f': exact {len(json.loads(cases.read_text()))} cases, {len(want)} bytes')
 mutant=scratch/'mutant';mutant.mkdir();(mutant/'nodes_runner.a').write_text(runner.read_text().replace('./options_json.ts','../options_json.ts'));shutil.copy2(owned.parent/'options_json.ts',scratch/'options_json.ts')
 spec=json.loads((owned/'nodes_mutant.json').read_text());text=(owned/'withdrawn_nodes_from_static_declarations.a.txt').read_text();assert text.count(spec['from'])==1
 (mutant/'nodes_from_static_declarations.a').write_text(text.replace(spec['from'],spec['to']))
 mn=scratch/'mutant-native';mj=scratch/'mutant.mjs';run([builder,mutant/'nodes_runner.a',mn,mj,'sanitized'])
 for side,args in [('Node',node+[mutant/'nodes_runner.a',cases]),('native',[mn,cases]),('emitted JavaScript',node+[mj,cases])]:
  got,_,_=run(args);assert got!=want;note(side+': compiling exit-zero clean-stderr presence mutant caught only by Go comparison')
 shutil.copy2(cases,evidence/'nodes-cases.json')
 note('PASS full upstream framework table, field combinations and fresh allocation four-way comparison')
except Exception as e:note('FAIL '+repr(e));raise
finally:log.close()
