#!/usr/bin/env python3
"""Verify completed independent applies and freeze the entry closure inventory."""
import hashlib,json,os,pathlib,subprocess
p=pathlib.Path(__file__).resolve().parent
scratch=pathlib.Path('/workspace/scratch/v3-plan');scratch.mkdir(exist_ok=True)
manifests={};headers={}
for label in ('main','head'):
 tree=pathlib.Path(f'/workspace/v3-{label}-adapted-complete')
 assert (tree/'patch-set.md').is_file(),'apply incomplete: '+label
 manifests[label]={str(x.relative_to(tree)):hashlib.sha256(x.read_bytes()).hexdigest() for x in sorted((tree/'src').rglob('*')) if x.is_file()}
 env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE='/V3_NO_FILE')
 with (scratch/(label+'.log')).open('w') as log:
  subprocess.run([f'/workspace/scratch/v3-{label}-count-bin',str(tree/'src/tsc/tsc.ts'),str(scratch/(label+'.jsonl'))],env=env,stdout=log,stderr=log,check=True,timeout=120)
 rows=list(map(json.loads,(scratch/(label+'.jsonl')).read_text().splitlines()));assert len(rows)==1
 headers[label]=rows[0]
assert manifests['main']==manifests['head'],'own applies produced different source bytes'
files=sorted(str(pathlib.Path(x).relative_to('/workspace/v3-main-adapted-complete')) for x in headers['main']['sources'])
assert files==sorted(str(pathlib.Path(x).relative_to('/workspace/v3-head-adapted-complete')) for x in headers['head']['sources'])
(p/'source-manifest.json').write_text(json.dumps(manifests['main'],indent=2)+'\n')
(p/'plan.json').write_text(json.dumps(dict(mode='full',files=files,source_files=len(manifests['main']),all_source_bytes_identical=True),indent=2)+'\n')
print(json.dumps(dict(closure_files=len(files),source_files=len(manifests['main']))))
