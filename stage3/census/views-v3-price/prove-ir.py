#!/usr/bin/env python3
"""Run head-owned known-count fixtures, both skip-kind mutants, and meter identity."""
import hashlib,json,os,pathlib,re,subprocess
out=pathlib.Path('/workspace/scratch/v3-ir-proof');out.mkdir(exist_ok=True)
expected={'per-read':(2,0),'map-good':(1,1),'for-of-good':(1,1)}
summary=[]
for label in ('head','main'):
 for fixture,want in expected.items():
  src=pathlib.Path('/workspace/v3-head/internal/oracle/testdata/views_v3')/(fixture+'.ts')
  for variant in ('meter','count','ViewElement','ViewElementCheck'):
   prefix=out/f'{label}-{fixture}-{variant}'
   env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1');env.pop('V3_MUTANT_SKIP_NODE',None)
   if variant not in ('meter','count'):env['V3_MUTANT_SKIP_NODE']=variant
   binary=f'/workspace/scratch/v3-{label}-'+('meter' if variant=='meter' else 'count')+'-bin'
   with prefix.with_suffix('.log').open('w') as log:
    subprocess.run([binary,str(src),str(prefix.with_suffix('.jsonl'))],env=env,stdout=log,stderr=log,check=True,timeout=120)
   raw=prefix.with_suffix('.jsonl').read_bytes()
   if variant=='meter':baseline=raw;continue
   stripped=re.sub(rb'"(?:ir_measured|view_elements|view_element_checks)":(?:true|false|[0-9]+),',b'',raw)
   assert stripped==baseline,(label,fixture,variant,'byte identity')
   units=[u for row in map(json.loads,raw.splitlines()) for u in row.get('units',[])]
   got=tuple(sum(u[key] for u in units) for key in ('view_elements','view_element_checks'))
   target=want if label=='head' else (0,0)
   killed=None
   if variant=='count':assert got==target,(label,fixture,got,target)
   elif label=='head' and target[0 if variant=='ViewElement' else 1]>0:
    try:assert got==target
    except AssertionError:killed=True
    else:raise AssertionError('skip-kind mutant survived')
   summary.append(dict(label=label,fixture=fixture,variant=variant,counts=got,expected=target,byte_identical=True,mutant_killed=killed,baseline_sha256=hashlib.sha256(baseline).hexdigest()))
(out/'proof.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps(summary,indent=2))
