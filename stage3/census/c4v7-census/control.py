"""Validate the scan-only driver's output guards and registered refusal rules."""
import gzip,json,os,subprocess,sys,time
from pathlib import Path
BASE=Path('/workspace/scratch/c4v7-census');OUT=Path(__file__).resolve().parent
label=sys.argv[1];assert label in ('main','head')
controls={'promise-3210':('async function fails(): Promise<void> {}\nconst run: () => unknown = fails;\nrun();\n','an unawaited async task: a Promise<void> at the initializer site'), 'promise-3220':('async function fails(): Promise<void> {}\n[0].map(fails);\n','an unawaited async task: a Promise<void>[] thrown away')}
status=[]
for name,(source,needle) in controls.items():
 root=BASE/'controls'/name;root.mkdir(parents=True,exist_ok=True);(root/'control.ts').write_text(source)
 dest=OUT/'controls';dest.mkdir(exist_ok=True);raw=dest/f'{label}-{name}.jsonl'
 env=dict(os.environ,LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE=str(root/'control.ts'),CENSUS_MODULE_FORMAT='CommonJSBundle')
 start=time.monotonic()
 with (dest/f'{label}-{name}.log').open('wb') as log:
  r=subprocess.run(['timeout','--signal=TERM','--kill-after=10s','120s',str(BASE/f'scan-{label}'),str(root),str(raw)],cwd=BASE/label,env=env,stdout=log,stderr=log)
 assert r.returncode==0,r.returncode
 rows=[json.loads(x) for x in raw.read_text().splitlines()];records=[x for x in rows if x.get('file')]
 assert len(records)==1 and records[0]['units']==[]
 assert any(needle in f['reason'] for f in records[0]['findings'])
 assert rows[-1]['status']=='relation_report'
 status.append(dict(name=name,exit_code=r.returncode,seconds=time.monotonic()-start,zero_lowering_units=True,expected_reason_found=True))
(dest/f'{label}-status.json').write_text(json.dumps(status,indent=2)+'\n')
print(label,'controls passed')
