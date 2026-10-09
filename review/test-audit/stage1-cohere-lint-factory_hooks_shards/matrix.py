from pathlib import Path
import subprocess,json,time,os
root=Path('/workspace/adamic');out=root/'review/test-audit/stage1-cohere-lint-factory_hooks_shards'
regex='^(TestFactoryHooks(Union|PlantedFailure|_[0-9]{3}|_NativeBuildWorker)|TestNestedConstructorGap|TestProduct_(DotARenameSource|DotARenameOracle|RulesAgreeOracle)|TestRulesAgreeLoweringCacheKey)$'
(out/'matrix-selection.txt').write_text(regex+'\nExcluded as over budget at clean baseline: DotARename family, RulesAgree family and its shared-product witness. All other package tests outside named slice remain unknown.\n')
fixture=Path('/tmp/u108/witness.ts');fixture.write_text('debugger;\n');manifest=Path('/tmp/u108/witness.tsv');manifest.write_text(str(fixture)+'\tno-debugger\t\t\tfalse\t{"Number":-2,"Payload":{"enabled":true}}\n')
subprocess.run(['go','run','./cmd/lint-registry'],cwd=root,stdout=(out/'logs'/'registry.log').open('w'),stderr=subprocess.STDOUT,check=True)
node=['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(root/'stage1/cohere/lint/main.ts'),'--manifest',str(manifest)]
with (out/'logs'/'survivor-before.log').open('w') as log:subprocess.run(node,cwd=root,stdout=log,stderr=subprocess.STDOUT,check=True)
metrics=[]
for m in json.loads((out/'plan.json').read_text()):
 p=root/m['file'];s=p.read_text();p.write_text(s.replace(m['old'],m['new']));env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u108/cache/'+m['id']
 try:
  if m['file'].endswith('.go'):
   with (out/'logs'/f"{m['id']}-vet.log").open('w') as log:subprocess.run(['go','vet','./internal/lower/'],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT,check=True)
  start=time.monotonic()
  with (out/'logs'/f"{m['id']}.log").open('w') as log:r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/lint/','-run',regex],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
  metrics.append(dict(id=m['id'],wall=time.monotonic()-start,exit=r.returncode,regex=regex));(out/'matrix-times.json').write_text(json.dumps(metrics,indent=2))
  if m['id']=='M02':
   with (out/'logs'/'survivor-after-node.log').open('w') as log:subprocess.run(node,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT,check=True)
   for f in Path(env['ADAMIC_BUILD_CACHE_DIR']).glob('*.inputs'):
    binary=f.with_suffix('')/'scanner'
    if 'factory-hooks-native' in f.read_text() and binary.exists():
     with (out/'logs'/'survivor-after-native.log').open('w') as log:subprocess.run([str(binary),'--manifest',str(manifest)],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT,check=True)
 finally:p.write_text(s)
