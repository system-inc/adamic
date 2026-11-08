"""Collect the four standard full-core runs before the long entry scans finish."""
import json,subprocess,sys
from pathlib import Path
out=Path(__file__).resolve().parent;repo=out.parents[2]
for label in ['A','B']:
 tree=Path('/tmp/branded-'+label)
 for side in ['parent','head']:
  name=f'{label}-{side}-core'
  cmd=['env','-u','LATENT_FULL','-u','LATENT_SPECULATIVE','-u','LATENT_ONLY_FILE','LATENT_ASSERT_NO_OUTPUT=1',f'LATENT_TIMES={out}/{name}-units.jsonl','LATENT_FULL=1',f'LATENT_ONLY_FILE={tree}/src/compiler/core.ts',f'/tmp/branded-{side}-core',str(tree/'src/compiler'),str(out/(name+'.jsonl'))]
  rc=subprocess.run([sys.executable,str(out/'run.py'),name,str(repo),*cmd]).returncode
  if rc:raise SystemExit(rc)
  records=[json.loads(line) for line in (out/(name+'.jsonl')).read_text().splitlines()]
  assert records[0]['status']=='measurement' and records[0]['latent_mode']=='full'
  assert len(records)==2 and records[1]['file']==str(tree/'src/compiler/core.ts')
  print(name,'checker_diagnostics='+str(len(records[0]['diagnostics'])),flush=True)
rc=subprocess.run([sys.executable,str(out/'core-ledger.py')]).returncode
raise SystemExit(rc)
