import json,os,shutil,subprocess
from pathlib import Path
root=Path.cwd(); evidence=root/'review/compiler/loom-r-slice-6-m2'
products=Path('/tmp/adamic-gate/loom-r-slice-6-m2/reader-products')
os.sched_setaffinity(0,sorted(os.sched_getaffinity(0))[:4]); results=[]
for part in ['Native','Lower','Fresh','Flow']:
 directory=products/part; environment=dict(os.environ,ADAMIC_BUILD_OUTPUT_DIR=str(directory))
 command=['timeout','100','go','test','-p','1','./internal/ir','-run','^(TestCallTargetExportInputs'+part+'|TestCallTargetReaders'+part+')$','-count=1','-parallel=4','-timeout','90s','-json']
 with (evidence/('readers-'+part+'.jsonl')).open('w') as log:result=subprocess.run(command,env=environment,stdout=log,stderr=subprocess.STDOUT,timeout=105)
 results.append({'part':part,'exit':result.returncode,'command':command});print(part,'exit',result.returncode,flush=True)
 # These are disposable copies of export data, retained only until this guard
 # finishes. No source or active Go cache is removed.
 if directory.exists():shutil.rmtree(directory)
(evidence/'readers.json').write_text(json.dumps(results,indent=2)+'\n')
assert all(result['exit']==0 for result in results)
