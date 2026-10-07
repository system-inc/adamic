import argparse,json,subprocess
from pathlib import Path
HERE=Path(__file__).resolve().parent;ROOT=HERE.parents[4]
p=argparse.ArgumentParser();p.add_argument('--scratch',type=Path,required=True);p.add_argument('--source',type=Path,default=HERE/'validation.ts');a=p.parse_args();S=a.scratch.resolve()
summary=json.loads((S/'summary.json').read_text());observations=[]
for scope in summary['corpora']:
 for number in range(len(scope['batches'])):
  label=scope['name']+'-'+str(number);raw=S/(label+'.json');want=(S/(label+'-go.log')).read_bytes()
  for backend,cmd in [('Node',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',a.source,raw,'--parse']),('emitted',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',S/'emitted.mjs',raw,'--parse']),('native',[S/'native',raw,'--parse'])]:
   with (S/('raw-'+label+'-'+backend+'.log')).open('wb') as out,(S/('raw-'+label+'-'+backend+'.stderr')).open('wb') as err:r=subprocess.run(list(map(str,cmd)),cwd=ROOT,stdout=out,stderr=err)
   actual=(S/('raw-'+label+'-'+backend+'.log')).read_bytes();error=(S/('raw-'+label+'-'+backend+'.stderr')).read_bytes();assert r.returncode==0 and not error and actual==want,label+' '+backend
   observations.append({'scope':scope['name'],'batch':number,'backend':backend,'equal':True,'exit':0,'stderrBytes':0})
  if number%10==0:print(scope['name'],'raw batches',number+1,flush=True)
(S/'raw-summary.json').write_text(json.dumps(observations,indent=2));print('raw corpus observations',len(observations),'PASS',flush=True)
