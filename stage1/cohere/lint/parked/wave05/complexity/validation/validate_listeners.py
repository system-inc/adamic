"""Owned numeric listener declaration comparison. Does not certify rule traversal."""
from pathlib import Path
import os,json,subprocess,shutil,hashlib
here=Path(__file__).resolve().parent;rules=here.parents[1];repo=here.parents[5]
compiler=Path(os.environ.get('ADAMIC_COMPILER_REPO',str(repo))).resolve();cohere=compiler/'cohere'
scratch=Path('/tmp/w05-numeric-listeners');scratch.mkdir(exist_ok=True)
def run(args,label,cwd=compiler):
 with (scratch/(label+'.log')).open('wb') as out,(scratch/(label+'.stderr.log')).open('wb') as err:subprocess.run(args,cwd=cwd,stdout=out,stderr=err,check=True)
 assert (scratch/(label+'.stderr.log')).stat().st_size==0,label
virtual=cohere/'adamic_numeric_listeners.go';overlay=scratch/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(here/'listeners_oracle.go.txt')}}))
run(['go','run','-overlay='+str(overlay),str(virtual)],'Go',cohere);want=(scratch/'Go.log').read_bytes()
build=compiler/'adamic_owned_listener_build.go';overlay=scratch/'build-overlay.json';overlay.write_text(json.dumps({'Replace':{str(build):str(here/'build.go.txt')}}))
def compare(entry,label,mutant=False):
 prefix=scratch/(label+'-binary');run(['go','run','-overlay='+str(overlay),str(build),str(entry),str(prefix)],label+'-build')
 for side,args in [('Node',['node','--disable-warning=ExperimentalWarning',str(compiler/'oracle/node.mjs'),str(entry)]),('JavaScript',['node','--disable-warning=ExperimentalWarning',str(compiler/'oracle/node.mjs'),str(prefix)+'.mjs']),('native',[str(prefix)])]:
  run(args,label+'-'+side);got=(scratch/(label+'-'+side+'.log')).read_bytes();assert got!=want if mutant else got==want,(label,side)
  print(label,side,'mutant caught' if mutant else 'identical',len(got),'bytes',flush=True)
compare(here/'listeners_driver.a','baseline')
manifest=json.loads((here/'listeners_manifest.json').read_text())
for changed in manifest:
 variant=scratch/'mutant'/changed['directory']
 for row in manifest:
  target=variant/row['directory'];target.mkdir(parents=True,exist_ok=True);shutil.copyfile(rules/row['directory']/'listeners.a',target/'listeners.a')
 driver=variant/'complexity/validation';driver.mkdir(parents=True,exist_ok=True);shutil.copyfile(here/'listeners_driver.a',driver/'listeners_driver.a')
 file=variant/changed['directory']/'listeners.a';text=file.read_text();number=changed['numericKinds'][0];kind=changed['kinds'][0];anchor='    '+str(number)+', // '+kind;assert text.count(anchor)==1;file.write_text(text.replace(anchor,'    '+str(number+1)+', // deliberately wrong '+kind))
 compare(driver/'listeners_driver.a','mutant-'+changed['directory'],True)
evidence={}
for file in scratch.glob('*.log'):
 if file.name.endswith('.stderr.log'):continue
 data=file.read_bytes();err=scratch/(file.name[:-4]+'.stderr.log');evidence[file.name]={'bytes':len(data),'sha256':hashlib.sha256(data).hexdigest(),'stderrBytes':err.stat().st_size if err.exists() else None}
(scratch/'evidence.json').write_text(json.dumps(evidence,indent=2)+'\n')
print('declarations',len(manifest),'listener entries',sum(len(row['numericKinds']) for row in manifest),flush=True)
