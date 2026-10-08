import importlib.util, os, sys, json, tempfile, time, socket, shutil, glob, subprocess
from types import SimpleNamespace
root='/workspace/adamic'
cpus=sorted(os.sched_getaffinity(0))[:4]
os.sched_setaffinity(0, cpus)
mode,phase=sys.argv[1:]
if mode == 'before' and phase == 'wasi' and os.path.isdir('/tmp/split-catalog-plant-cache'):
 import subprocess
 subprocess.run([sys.executable,'/tmp/split-plant.py','catalog'],check=True)
 shutil.rmtree('/tmp/split-catalog-plant-cache')
if phase == 'wasi' and not os.path.exists('/workspace/adamic-tools/wasi-sdk/bin/clang'):
 import subprocess
 subprocess.run(['bash','-c','mkdir -p /workspace/adamic-tools/wasi-sdk; curl -fsSL https://github.com/WebAssembly/wasi-sdk/releases/download/wasi-sdk-27/wasi-sdk-27.0-x86_64-linux.tar.gz | tar --no-same-owner -xz -C /workspace/adamic-tools/wasi-sdk --strip-components 1'], check=True)
path='/tmp/split-before-run.py' if mode=='before' else root+'/cloud/fast-gate/run.py'
spec=importlib.util.spec_from_file_location('gate',path); mod=importlib.util.module_from_spec(spec);spec.loader.exec_module(mod)
out='/tmp/split-evidence/'+mode+'-'+phase
os.makedirs(out,exist_ok=True)
for previous_cache in glob.glob('/tmp/split-evidence/*/cold-build-*'):
 shutil.rmtree(previous_cache)
os.environ['GOCACHE']=tempfile.mkdtemp(prefix='cold-build-',dir=out)
os.environ['GOMAXPROCS']='4'
os.environ['GOFLAGS']='-p=4'
os.environ['WASI_SYSROOT']='/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot'
sha=mod.git(root,'rev-parse','HEAD')
a=SimpleNamespace(tree=root,tools=root,sha=sha,base=sha,out=out,parallel=2,full=True,complete=True,branch='devtools/split-gate-phases',branch_source='',session='',session_source='')
g=mod.Gate(a);g.planned=[phase]
g.kind='fast'
g.result['measurement']={'instrument':'Python time.monotonic (including process startup, compilation, setup and cleanup)','cold':'fresh empty shared GOCACHE at each phase start; compiler artifacts reused within the phase; -count=1, ADAMIC_GATE_UNCACHED=1; no warmup; does not prove independently cold time for later units','affinity':cpus,'cpu_max':open('/sys/fs/cgroup/cpu.max').read().strip(),'hostname':socket.gethostname(),'mode':mode,'unit_workers':2 if mode=='after' else ('catalog default: 4' if phase=='catalog' else 1)}
if mode == 'before' and phase == 'catalog':
 original_spawn = g.spawn
 def bounded_catalog(command, *args, **kwargs):
  if command[:2] == ['bash','verify/catalog/check.sh']:
   command = command + ['--jobs','2']
  return original_spawn(command,*args,**kwargs)
 g.spawn = bounded_catalog
 g.result['measurement']['catalog_jobs_override'] = 2
start=time.monotonic()
with open(out+'/test.jsonl','w') as log:
 if phase=='catalog':g.catalogFull()
 else:
  env={'PATH':'/workspace/adamic-tools/wasi-sdk/bin:'+os.environ['PATH']}
  if mode=='before':g.test('wasi',['go','test','-count=1','-json','-timeout','3h','-run','^TestWASI$','./internal/native'],log,env)
  else:g.wasiSplit(log,env)
g.result['measurement']['wall_seconds']=time.monotonic()-start
g.finish()
if mode == 'after' and phase == 'wasi':
 subprocess.run([sys.executable,'/tmp/split-plant.py','wasi'],check=True)
shutil.rmtree(os.environ['GOCACHE'],ignore_errors=True)
