import sys,os,subprocess,json
from pathlib import Path
phase=sys.argv[1];tree=Path(os.environ.get('FAMILY5_TREE','/tmp/family5-base-tree'));out=Path('/tmp/family5-evidence')/phase
files=sys.argv[2:]
for file in files:
 d=out/file.removesuffix('.ts');d.mkdir(parents=True,exist_ok=True)
 env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_TAIL='1',CENSUS_MODULE_FORMAT='CommonJSBundle',LATENT_ONLY_FILE=str(tree/'src/compiler'/file),LATENT_TIMES='/dev/stderr')
 command=[sys.executable,'/tmp/family5-proof.py',str(d),'/tmp/family5-tools/entry-census',str(tree/'src/tsc/tsc.ts'),str(d/'latent.jsonl')]
 code=subprocess.call(command,env=env)
 if code:sys.exit(code)
 rows=[json.loads(l) for l in (d/'latent.jsonl').read_text().splitlines()]
 if not rows or rows[-1].get('status')!='relation_report' or rows[0].get('checker_rejected'):
  print('FAIL census completeness or checker diagnostics',file,flush=True);sys.exit(1)
 print('PASS complete checker-clean census '+phase+' '+file+' child_exit=0',flush=True)
