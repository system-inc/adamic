"""Preserve named full-unit observations while long entry scans continue."""
import json,subprocess,sys
from pathlib import Path
out=Path(__file__).resolve().parent;repo=out.parents[2]
for label in ['A','B']:
 sites=json.loads((out/f'{label}-sites.json').read_text())['checker']
 unique={(r['unit'] if label=='A' else r['observing_unit']):r for r in reversed(sites)}
 for i,(unit,row) in enumerate(sorted(unique.items()),1):
  for side in ['parent','head']:
   name=f'{label}-{side}-focus-{i}'
   selector=row['site'] if label=='A' else unit
   cmd=['env','LATENT_FULL=1','LATENT_ASSERT_NO_OUTPUT=1',f'/tmp/branded-{side}-focus','-project',f'/tmp/branded-{label}/src/tsc/tsc.ts','-where',selector,'-output',str(out/(name+'.jsonl'))]
   rc=subprocess.run([sys.executable,str(out/'run.py'),name,str(repo),*cmd]).returncode
   if rc:raise SystemExit(rc)
rc=subprocess.run([sys.executable,str(out/'serializer.py'),'A']).returncode
raise SystemExit(rc)
