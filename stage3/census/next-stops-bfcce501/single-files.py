"""Two separate full-mode file censuses using the meter latent binary."""
import json,os,sys
from pathlib import Path
out=Path(__file__).resolve().parent
sys.path.insert(0,str(out));from run import run
for name in ['debug','semver','tracing']:
 source='/workspace/scratch/next-stops-adapted/src/compiler/'+name+'.ts'
 os.environ.update(LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE=source,LATENT_TIMES=str(out/(name+'-full-times.jsonl')))
 run(name+'-full',['/workspace/scratch/next-stops-latent','/workspace/scratch/next-stops-adapted/src/compiler',str(out/(name+'-full.jsonl'))])
 rows=[json.loads(line) for line in (out/(name+'-full.jsonl')).read_text().splitlines()]
 findings=[f for r in rows for f in r.get('findings') or []]
 def order(f):
  loc=f.get('where','').rsplit(':',2)
  return (loc[0],int(loc[1]),int(loc[2])) if len(loc)==3 else (f.get('where',''),0,0)
 findings.sort(key=order)
 (out/(name+'-full-findings.json')).write_text(json.dumps(findings,indent=2)+'\n')
 print(name, 'findings',len(findings),flush=True)
