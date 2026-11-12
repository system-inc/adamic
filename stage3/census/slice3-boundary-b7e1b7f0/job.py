"""Launch one compiler child under the recorded bound. No measurement seam source is here."""
import json,os,subprocess,sys
from pathlib import Path
D=Path(__file__).resolve().parent;i=int(sys.argv[1]);side=sys.argv[2];assert side in ['base','head'];group=json.loads((D/'groups.json').read_text())[i];prefix=D/f'raw/group-{i}-{side}'
assert not prefix.with_suffix('.json').exists()
env=dict(os.environ,LATENT_FULL='0',LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE='\n'.join(group['files']),BOUNDARY_UNIT_RANGES=json.dumps(group['ranges']),BOUNDARY_STREAM=str(prefix)+'-records.jsonl',LATENT_TIMES=str(prefix)+'-times.jsonl',GOMAXPROCS='1',GOMEMLIMIT='2800MiB',ADAMIC_TYPESCRIPT_SOURCE='/workspace/roads-typescript',PYTHONDONTWRITEBYTECODE='1')
for k in ['BOUNDARY_FORCE_REFUSE','BOUNDARY_DEBUG_UNIT']:env.pop(k,None)
with Path(str(prefix)+'-wrapper.log').open('wb') as log:
 child=subprocess.run(['python3',str(D/'run.py'),str(prefix),'1490',f'/workspace/slice3-{side}-census','/workspace/slice3-adapted/src/tsc/tsc.ts',str(prefix)+'-normal.jsonl'],env=env,stdout=log,stderr=subprocess.STDOUT)
print(json.dumps({'group':i,'side':side,'wrapper_child_exit':child.returncode}),flush=True);sys.exit(child.returncode)
