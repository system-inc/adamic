import importlib.util,os,json,tempfile,sys,subprocess,time
from types import SimpleNamespace
root='/workspace/adamic'
spec=importlib.util.spec_from_file_location('gate',root+'/cloud/fast-gate/run.py');mod=importlib.util.module_from_spec(spec);spec.loader.exec_module(mod)
phase=sys.argv[1]
out='/tmp/split-evidence/planted-'+phase
os.makedirs(out,exist_ok=True)
sha=mod.git(root,'rev-parse','HEAD')
a=SimpleNamespace(tree=root,tools=root,sha=sha,base=sha,out=out,parallel=2,full=True,complete=True,branch='devtools/split-gate-phases',branch_source='',session='',session_source='')
g=mod.Gate(a);g.planned=[phase];g.kind='fast'
wrappers=tempfile.mkdtemp(prefix='fault-',dir=out)
if phase=='catalog':
 os.environ['GOCACHE']='/tmp/split-catalog-plant-cache'
 realGo='/workspace/adamic-tools/go/bin/go'
 with open(wrappers+'/go','w') as f:f.write('#!/bin/bash\nif [ "$1" = test ]; then case "$*" in *shared_slice_append*) echo "planted unit failure" >&2; exit 2;; esac; fi\nexec '+realGo+' "$@"\n')
 os.chmod(wrappers+'/go',0o755)
 original=g.spawn
 def spawn(command,stdout,stderr=mod.subprocess.STDOUT,directory=None,environment=None):
  env=dict(environment or {},PATH=wrappers+':'+os.environ['PATH'])
  return original(command,stdout,stderr,directory,env)
 g.spawn=spawn
 original_units=g.phaseUnits
 def units(phase,names,commands,execute):return original_units(phase,names[:2],commands[:2],execute)
 g.phaseUnits=units
 g.catalogFull()
 assert g.failure['step']=='catalog/01 shared-slice-append',g.failure
 assert not g.result['catalog_units'][0]['ok'] and g.result['catalog_units'][1]['ok']
else:
 os.environ['WASI_SYSROOT']='/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot'
 realNode='/workspace/adamic-tools/bin/node'
 with open(wrappers+'/node','w') as f:f.write('#!/bin/bash\nlast="${!#}"\ncase "$last" in */01_hello.ts) exit 17;; esac\nexec '+realNode+' "$@"\n')
 os.chmod(wrappers+'/node',0o755)
 names=mod.wasiFixtures(root)[:2]
 mod.wasiFixtures=lambda tree:names
 with open(out+'/test.jsonl','w') as log:
  g.wasiSplit(log,{'PATH':wrappers+':/workspace/adamic-tools/wasi-sdk/bin:'+os.environ['PATH']})
 assert g.failure['step']=='wasi/'+names[0],g.failure
 assert not g.result['wasi_units'][0]['ok'] and g.result['wasi_units'][1]['ok']
 assert any('TestWASI/'+names[0] in name for name in g.failedTests)
g.result['failure_probe']={'phase':phase,'injection':'Go test control exits 2 for shared_slice_append' if phase=='catalog' else 'Node exits 17 only for 01_hello.ts source oracle','real_candidate_tests':True,'warm_compiler_cache':True,'exactly_one_semantic_failure':True}
g.finish()
print('PLANTED FAILURE CAUGHT:',g.failure['step'],flush=True)
