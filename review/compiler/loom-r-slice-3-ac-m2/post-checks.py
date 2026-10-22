import importlib.util,json,os,select,subprocess,sys,time,shutil
from pathlib import Path
out=Path(__file__).resolve().parent
pid=int(sys.argv[1]);descriptor=os.pidfd_open(pid);poll=select.poll();poll.register(descriptor,select.POLLIN)
started=time.monotonic()
while not poll.poll(1000):
 if time.monotonic()-started>5400:raise TimeoutError('first ordered suite did not complete within 90 minutes')
os.close(descriptor)
print('Ordered suite driver completed; starting repaired-toolchain verification',flush=True)
spec=importlib.util.spec_from_file_location('suite_helpers',out/'run-suites.py');module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
# Only the raw GCC portability tests need the installed LLVM allocator header.
base_path=os.environ['PATH'];os.environ['PATH']='/workspace/scratch/slice3-gcc/bin:'+base_path
row=module.run('internal-native-repaired',['go','test','./internal/native/...','-skip','^TestProduct_','-count=1','-json','-parallel=4','-timeout=13m','-p','1'])
os.environ['PATH']=base_path
if row['exit'] or row['over_budget']:raise RuntimeError('repaired native verification is not green and within grain')
row=module.run('production-mutants',['python3',str(out/'run-mutants.py')],1440)
if row['exit']:raise RuntimeError('production mutant verification failed')
os.environ['NODE_PATH']=str(Path.cwd()/'stage3/api/node_modules')
row=module.run('assignment-proof-driver',['python3','stage3/fixtures/assignment-proofs/check.py','--require-adamic'],240)
if row['exit']:raise RuntimeError('assignment proof driver failed')
text=(out/'assignment-proof-driver.log').read_text()
if 'logs ' in text:
 logs=Path(text.strip().rsplit('logs ',1)[1]);destination=out/'assignment-proof-logs';destination.mkdir(exist_ok=True)
 for p in logs.iterdir():
  if p.suffix in ['.log','.stdout','.stderr']:shutil.copyfile(p,destination/p.name)
row=module.run('reader-guard-final',['go','test','./internal/ir','-run','^TestCallTargetReaders$','-count=1','-timeout=90s','-json'],120)
if row['exit']:raise RuntimeError('final reader guard failed')
row=module.run('vet-final',['go','vet','-p','1',*['./'+p+'/...' for p in module.packages]],180)
if row['exit']:raise RuntimeError('final vet failed')
print('Post-checks complete',flush=True)
