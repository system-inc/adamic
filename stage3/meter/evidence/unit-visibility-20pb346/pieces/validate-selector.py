"""Check empty-selector identity and exact two-piece inventory coverage."""
from pathlib import Path
import json, os, subprocess, hashlib
root=Path('/workspace/scratch/20pb346-2')
source=root/'adapted/src/compiler'
file=str(source/'corePublic.ts')
out=root/'selector-validation'
out.mkdir()
inventory=json.loads((root/'unit-inventory.jsonl').read_text().splitlines()[1])[file]
assert len(inventory)==len(set(inventory))
def run(label,binary,units=None):
 env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')}
 env.update(LATENT_FULL='1',LATENT_ONLY_FILE=file,LATENT_ASSERT_NO_OUTPUT='1',GOMEMLIMIT='8GiB')
 if units is not None:env['LATENT_PROOF_UNITS']='\n'.join(units)
 with (out/(label+'.txt')).open('w') as log:
  subprocess.run([str(root/binary),str(source),str(out/(label+'.jsonl'))],env=env,stdout=log,stderr=subprocess.STDOUT,check=True,timeout=120)
 data=(out/(label+'.jsonl')).read_bytes()
 rows=[json.loads(l) for l in data.splitlines()]
 actual=[u['where']+'|'+u['kind'] for r in rows[1:] for u in r['units']]
 assert actual==(inventory if units is None else units),(label,actual,units)
 return data
result={}
for side in ('before','after'):
 original=run(side+'-original','pieces-'+side)
 bounded=run(side+'-unfiltered','bounded-'+side)
 assert original==bounded
 for i in range(2):run(side+'-half-'+str(i),'bounded-'+side,inventory[i::2])
 result[side]=dict(empty_selector_sha256=hashlib.sha256(bounded).hexdigest(),units=len(inventory),halves=[len(inventory[0::2]),len(inventory[1::2])])
assert result['before']==result['after']
Path(__file__).with_name('selector-validation.json').write_text(json.dumps(result,indent=2)+'\n')
