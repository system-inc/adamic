"""Owned parking gate: fresh fixture oracles and independently compiled semantic mutants."""
from pathlib import Path
import subprocess,json,shutil,concurrent.futures,os
HERE=Path(__file__).resolve().parent;ROOT=HERE.parents[4];S=Path('/tmp/wave11-parking-oracles');S.mkdir(exist_ok=True)
groups=[(slug,[slug]) for slug in ['adamic-no-definite-assignment','base-security-require-context-access','nexus-consistency-no-screaming-snake-case','nexus-import-no-forbidden-source','nexus-import-require-path-alias']]
def run(path,cmd):
 with path.open('wb') as out,path.with_suffix('.stderr').open('wb') as err:r=subprocess.run(list(map(str,cmd)),cwd=ROOT,stdout=out,stderr=err)
 assert r.returncode==0 and not path.with_suffix('.stderr').read_bytes(),str(path)
 return path.read_bytes()
def group(item):
 slug,mutants=item;directory=HERE.parent/slug;scratch=S/slug;scratch.mkdir(exist_ok=True)
 run(scratch/'validation.log',['python3',directory/'validate.py','--scratch',scratch])
 want=(scratch/'answer.log').read_bytes();caught=[]
 for name in mutants:
  owned=HERE.parent/name;change=json.loads((owned/'mutant.json').read_text());copy=scratch/change['name'];shutil.copytree(ROOT/'stage1',copy/'stage1',dirs_exist_ok=True)
  file=copy/owned.relative_to(ROOT)/change['file'];text=file.read_text();assert text.count(change['from'])==1;file.write_text(text.replace(change['from'],change['to'],1))
  source=copy/directory.relative_to(ROOT)/('validation.ts' if (directory/'validation.ts').exists() else 'validation.a');run(scratch/(change['name']+'-build.log'),['go','run',(directory/'validation_build.go' if (directory/'validation_build.go').exists() else HERE.parent/'nexus-consistency-no-screaming-snake-case/validation_build.go'),source,copy/'native',copy/'emitted.mjs'])
  for backend,cmd in [('Node',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',source,scratch/'ast.log']),('emitted',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',copy/'emitted.mjs',scratch/'ast.log']),('native',[copy/'native',scratch/'ast.log'])]:
   got=run(scratch/(change['name']+'-'+backend+'.log'),cmd);assert got!=want,'mutant survived '+name+' '+backend
  caught.append(change['name'])
 result={'suite':slug,'cases':len(json.loads((scratch/'upstream.json').read_text())),'bytes':len(want),'mutants':caught,'backends':['Node','emitted','sanitized native'],'byteOnly':True};print(json.dumps(result),flush=True);return result
selected=[item for item in groups if not os.getenv('WAVE11_PARK_SUITE') or item[0]==os.getenv('WAVE11_PARK_SUITE')]
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:results=list(pool.map(group,selected))
(S/('summary-'+os.getenv('WAVE11_PARK_SUITE','all')+'.json')).write_text(json.dumps(results,indent=2));print('PASS owned parking fixture and mutant gates',flush=True)
