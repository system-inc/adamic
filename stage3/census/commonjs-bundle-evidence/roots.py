"""One unchanged ordinary census invocation per child; bounded 20-root checkpoint."""
import json,os,subprocess,sys
from pathlib import Path
out=Path('/workspace/scratch/995spev')
label=sys.argv[1];start=int(sys.argv[2]);end=int(sys.argv[3])
assert end-start<=20
roots=sorted((out/'adapted/src/compiler').rglob('*.ts'));assert len(roots)==79,len(roots)
(out/'roots.json').write_text(json.dumps(list(map(str,roots)),indent=2)+'\n')
for i in range(start,end):
 name=f'{label}-root-{i+1:02}'
 dest=out/(name+'.jsonl')
 rc=subprocess.call([sys.executable,str(out/'run.py'),name,str(out/(label+'-census')),str(roots[i]),str(dest)],env=dict(os.environ, CENSUS_CHECKPOINT_CHILD='1'))
 print(name,'wrapper_exit='+str(rc),flush=True)
 if rc:sys.exit(rc)
 rows=[json.loads(s) for s in dest.read_text().splitlines()]
 assert len(rows)==2
 assert {k:v for k,v in rows[0].items() if k!='seconds'}=={k:v for k,v in rows[1].items() if k!='seconds'}
 print(name,rows[0]['kind'],rows[0].get('where',''),flush=True)
