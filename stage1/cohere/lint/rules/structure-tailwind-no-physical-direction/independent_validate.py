#!/usr/bin/env python3
"""Probe the Adamic parser on every upstream case with no JSX nodes, irrespective of filename."""
import argparse,json,subprocess,shutil
from pathlib import Path
HERE=Path(__file__).resolve().parent;ROOT=HERE.parents[4]
p=argparse.ArgumentParser();p.add_argument('--scratch',type=Path,required=True);a=p.parse_args();S=a.scratch.resolve()
rows=json.loads((S/'ast.log').read_text());plain=[r for r in rows if not any(n['kind'].startswith('Jsx') for n in r['ast'])]
# The Go tree classifies this probe's supported grammar. It does not supply the parsed tree or findings.
corpus=S/'non-jsx-upstream.json';corpus.write_text(json.dumps(plain));copy=S/'independent-non-jsx';shutil.copytree(ROOT/'stage1',copy/'stage1',dirs_exist_ok=True)
driver=copy/'stage1/cohere/lint/rules'/HERE.name/'validation.ts';source=driver.read_text();guard="        if(name.endsWith('.tsx')) { panic('NotYet: stage-1 JSX parser adapter'); }\n";assert source.count(guard)==1;driver.write_text(source.replace(guard,''))
def run(label,cmd,allow=False):
 with (S/(label+'.log')).open('wb') as out,(S/(label+'.stderr')).open('wb') as err:r=subprocess.run(list(map(str,cmd)),cwd=ROOT,stdout=out,stderr=err)
 result={'exit':r.returncode,'stderrBytes':(S/(label+'.stderr')).stat().st_size}
 if not allow:assert r.returncode==0 and result['stderrBytes']==0,label
 return result
run('non-jsx-build',['go','run',HERE/'validation_build.go',driver,copy/'native',copy/'emitted.mjs']);run('non-jsx-go',[S/'oracle',corpus]);wanted=(S/'non-jsx-go.log').read_bytes();results={'cases':len(plain),'rules':{n:sum(r['rule']==n for r in plain) for n in {r['rule'] for r in plain}},'JSXExcluded':len(rows)-len(plain),'backends':[]}
for name,cmd in [('Node',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',driver,corpus,'--parse']),('emitted',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',copy/'emitted.mjs',corpus,'--parse']),('native',[copy/'native',corpus,'--parse'])]:
 result=run('non-jsx-'+name,cmd,allow=True);result|={'name':name,'equal':result['exit']==0 and result['stderrBytes']==0 and (S/('non-jsx-'+name+'.log')).read_bytes()==wanted};results['backends'].append(result)
print(json.dumps(results,indent=2),flush=True);(S/'independent-non-jsx.json').write_text(json.dumps(results,indent=2))
