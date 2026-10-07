#!/usr/bin/env python3
"""Compare complete raw-source corpora through Adamic's own parser; preserve refusals and differences."""
import argparse,json,subprocess
from pathlib import Path
HERE=Path(__file__).resolve().parent;ROOT=HERE.parents[4]
p=argparse.ArgumentParser();p.add_argument('--scratch',type=Path,required=True);a=p.parse_args();S=a.scratch.resolve();results=[]
summary=json.loads((S/'summary.json').read_text())
for entry in summary['corpora']:
 corpus=entry['name'];batches=len(entry['batches'])
 for batch in range(batches):
  prefix=corpus+'-'+str(batch);wanted=(S/(prefix+'-go.log')).read_bytes();raw=S/(prefix+'.json')
  for name,cmd in [('Node',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',HERE/'validation.ts',raw,'--parse']),('emitted',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',S/'emitted.mjs',raw,'--parse']),('native',[S/'native',raw,'--parse'])]:
   label='independent-full-'+prefix+'-'+name
   with (S/(label+'.log')).open('wb') as out,(S/(label+'.stderr')).open('wb') as err:r=subprocess.run(list(map(str,cmd)),cwd=ROOT,stdout=out,stderr=err)
   got=(S/(label+'.log')).read_bytes();stderr=(S/(label+'.stderr')).read_bytes();result={'corpus':corpus,'batch':batch,'name':name,'exit':r.returncode,'stderrBytes':len(stderr),'equal':r.returncode==0 and not stderr and got==wanted};results.append(result)
   if not result['equal']:print('DIFFERENCE',result,stderr[:120],flush=True)
  (S/'independent-full.json').write_text(json.dumps(results,indent=2))
  if batch%10==0:print(corpus,'independent batches',batch+1,flush=True)
print('completed',len(results),'observations;',sum(r['equal'] for r in results),'equal',flush=True)
