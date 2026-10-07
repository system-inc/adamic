"""Check full findings/fixes against Go and Node, plus the sanitized driver."""
from pathlib import Path
import subprocess,hashlib,json,sys,time
before,after,root=map(Path,sys.argv[1:4]);manifest=['--manifest',str(before/'compiler.txt')]
commands={
 'Go':[str(before/'oracle')],
 'before':[str(before/'release')],
 'after':[str(after/'release')],
 'Node':['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(after/'main.ts')],
 'sanitized':[str(after/'sanitized')],
}
summary={};want=None
for name,command in commands.items():
 path=Path('/tmp/borrow-chains-full-'+name+'.stdout');error=Path('/tmp/borrow-chains-full-'+name+'.stderr')
 with path.open('wb') as output,error.open('wb') as stderr:
  start=time.perf_counter();subprocess.run(command+manifest,stdout=output,stderr=stderr,check=True);elapsed=time.perf_counter()-start
 data=path.read_bytes()
 if want is None:want=data
 assert data==want,name
 assert not error.stat().st_size,name
 summary[name]={'bytes':len(data),'sha256':hashlib.sha256(data).hexdigest(),'seconds':elapsed}
Path('/tmp/borrow-chains-parity.json').write_text(json.dumps(summary,indent=2)+'\n');print(json.dumps(summary,indent=2))
