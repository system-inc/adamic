#!/usr/bin/env python3
from pathlib import Path
import json,subprocess,shutil,os,re,time
owned=Path(__file__).resolve().parent;repo=owned.parents[4];cohere=repo/'cohere';scratch=Path('/tmp/w05-fourth-boundaries');scratch.mkdir(exist_ok=True)
def command(args,label,cwd=repo,env=None):
 start=time.monotonic()
 with (scratch/(label+'.log')).open('wb') as out,(scratch/(label+'.stderr.log')).open('wb') as err:subprocess.run(args,cwd=cwd,stdout=out,stderr=err,check=True,env=env)
 assert (scratch/(label+'.stderr.log')).stat().st_size==0
 return time.monotonic()-start
replace={str(cohere/'internal/lint/rules/boundaries/adamic_owned_bridge.go'):str(owned/'testdata/decision_bridge.go.txt'),str(cohere/'adamic_owned_decision.go'):str(owned/'testdata/decision_oracle.go.txt')}
(scratch/'oracle-overlay.json').write_text(json.dumps({'Replace':replace}));command(['go','build','-overlay='+str(scratch/'oracle-overlay.json'),'-o',str(scratch/'oracle'),str(cohere/'adamic_owned_decision.go')],'oracle-build',cohere)
test=cohere/'internal/lint/rules/boundaries/dependencies_test.go'
text=test.read_text();anchor='return decoded.(DependenciesOptions)';assert text.count(anchor)==1
(scratch/'capture.go').write_text(text.replace(anchor,'options := decoded.(DependenciesOptions)\n ownedCapture(options)\n return options'))
capture=scratch/'capture';capture.mkdir(exist_ok=True)
for path in capture.glob('*.json'):path.unlink()
(scratch/'capture-overlay.json').write_text(json.dumps({'Replace':{**replace,str(test):str(scratch/'capture.go')}}))
env=os.environ.copy();env['ADAMIC_BOUNDARY_CAPTURE']=str(capture)
command(['go','test','-overlay='+str(scratch/'capture-overlay.json'),'./internal/lint/rules/boundaries','-run','^TestDependencies','-count=1','-v'],'upstream-tests',cohere,env)
# Exact element/policy data, crossing suffixes, dots, descriptor order, overrides and both messages.
paths=['source/modules/chat/ChatModule.ts','source/modules/chat/deep/Thing.ts','source/common/Thing.ts','source/Thing.ts','workers/api/ApiWorker.ts','libraries/base/source/foundation/base/Base.ts','libraries/base/source/modules/account/AccountModule.ts','workers/api/.wrangler/state.ts','workers/.hidden/state.ts','source/modules/chat/workers/api/Thing.ts','libraries/base/source/api/rpc/client/RpcClient.ts','libraries/base/command-line/base-cli.ts','vendor/workers/thing/Thing.ts','workers/😀/Thing.ts']
elements=[{'type':'project-module','pattern':'source/modules/*'},{'type':'project-source','pattern':['source/*']},{'type':'project-worker','pattern':'workers/*'}]
policies=[[],[{'disallow':{'to':{'element':{'type':'project-worker'}}}}],[{'disallow':{'to':{'element':{'type':'project-worker'}}}},{'allow':{'to':{'element':{'type':'project-worker'}}}}],[{'allow':{'to':{'element':{'type':'project-worker'}}}},{'disallow':{'to':{'element':{'type':'project-worker'}}},'message':'Custom policy words.'}],[{'from':{'element':{'type':'project-worker'}},'to':{'element':{'types':{'anyOf':['project-module']}}},'disallow':{'from':{'element':{'type':'project-module'}},'to':{'element':{'types':{'anyOf':['project-worker']}}}}}]]
rows=[]
for default in ['','allow','disallow']:
 for policy in policies:
  for message in ['','Rule message words.']:
   options={'default':default,'policies':policy,'elements':elements,'message':message}
   for a in paths:
    for b in paths:rows.append(json.dumps(options,separators=(',',':'))+'\t'+a+'\t'+b)
for pattern in ['workers/[a-z]*','workers/[^a-z]*','workers/?','workers/\\a*','workers/**','workers/.hidden','workers/[😀]']:
 for a in paths:rows.append(json.dumps({'elements':[{'type':'probe','pattern':pattern}]},separators=(',',':'))+'\t'+a+'\t'+a)
print('upstream decoded configurations',len(list(capture.glob('*.json'))),flush=True)
def omit_nulls(value):
 if isinstance(value,dict):return {k:omit_nulls(v) for k,v in value.items() if v is not None}
 if isinstance(value,list):return [omit_nulls(v) for v in value]
 return value
for path in sorted(capture.glob('*.json')):
 options=omit_nulls(json.loads(path.read_text()))
 for a in paths:
  for b in paths:rows.append(json.dumps(options,separators=(',',':'))+'\t'+a+'\t'+b)
manifest=scratch/'cases.txt';manifest.write_text('\n'.join(rows)+'\n');print('decision cases',len(rows),flush=True)
command([str(scratch/'oracle'),str(manifest)],'Go');want=(scratch/'Go.log').read_bytes()
virtual=owned/'testdata/build_probe.go';(scratch/'build-overlay.json').write_text(json.dumps({'Replace':{str(virtual):str(owned/'testdata/build_probe.go.txt')}}))
def compare(module,label,mutant=False):
 prefix=scratch/(label+'-binary');entry=module/'testdata/decision_probe.a'
 command(['go','run','-overlay='+str(scratch/'build-overlay.json'),str(virtual),str(entry),str(prefix)],label+'-build')
 for side,args in [('Node',['node','--disable-warning=ExperimentalWarning',str(repo/'oracle/node.mjs'),str(entry),str(manifest)]),('JavaScript',['node','--disable-warning=ExperimentalWarning',str(repo/'oracle/node.mjs'),str(prefix)+'.mjs',str(manifest)]),('native',[str(prefix),str(manifest)])]:
  command(args,label+'-'+side);got=(scratch/(label+'-'+side+'.log')).read_bytes()
  if mutant:assert got!=want
  elif got!=want:
   for i,(a,b) in enumerate(zip(got.splitlines(),want.splitlines())):
    if a!=b:raise RuntimeError(side+' line '+str(i+1)+' got '+repr(a)+' want '+repr(b))
   raise RuntimeError(side+' length mismatch')
  print(label,side,'mutant caught' if mutant else 'identical',len(got),'bytes',flush=True)
compare(owned,'baseline')
variant=scratch/'mutant';shutil.copytree(owned,variant,dirs_exist_ok=True)
for file in variant.rglob('*.a'):
 original=owned/file.relative_to(variant)
 def rewrite(match):
  target=(original.parent/match.group(2)).resolve()
  if target.is_relative_to(owned):return match.group(0)
  return match.group(1)+json.dumps(str(target))+match.group(3)
 file.write_text(re.sub(r"(from\s+)[\"']([^\"']+)[\"'](;)",rewrite,file.read_text()))
spec=json.loads((owned/'mutant.json').read_text());target=variant/spec['file'];text=target.read_text();assert text.count(spec['from'])==1;target.write_text(text.replace(spec['from'],spec['to']))
compare(variant,'mutant',True)
