import argparse,json,subprocess,shutil
from pathlib import Path
HERE=Path(__file__).resolve().parent;ROOT=HERE.parents[4]
p=argparse.ArgumentParser();p.add_argument('--scratch',type=Path,required=True);a=p.parse_args();S=a.scratch.resolve();S.mkdir(parents=True,exist_ok=True)
def run(label,cmd,cwd=ROOT):
 with (S/(label+'.log')).open('wb') as out,(S/(label+'.stderr')).open('wb') as err:r=subprocess.run(list(map(str,cmd)),cwd=cwd,stdout=out,stderr=err)
 assert r.returncode==0 and not (S/(label+'.stderr')).read_bytes(),label
 return (S/(label+'.log')).read_bytes()
virtual=ROOT/'cohere/wave11_alias_components.go';exports=ROOT/'cohere/internal/lint/rules/nexus/wave11_alias_components.go';overlay=S/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(HERE/'component_oracle.go.txt'),str(exports):str(HERE/'component_exports.go.txt')}}))
run('Go-build',['go','build','-overlay='+str(overlay),'-o',S/'oracle',virtual],ROOT/'cohere');want=run('Go',[S/'oracle',HERE/'component_cases.json'])
builder=HERE.parent/'nexus-consistency-no-screaming-snake-case/validation_build.go'
def check(label,source,native,js,mutant=False):
 run(label+'-build',['go','run',builder,source,native,js])
 for name,cmd in [('Node',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',source,HERE/'component_cases.json']),('emitted',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',js,HERE/'component_cases.json']),('native',[native,HERE/'component_cases.json'])]:
  got=run(label+'-'+name,cmd);assert (got!=want)==mutant,name;print(label,name,'byte-only catch' if mutant else 'equal',flush=True)
check('baseline',HERE/'components.a',S/'native',S/'emitted.mjs')
change=json.loads((HERE/'mutant.json').read_text());copy=S/'mutant';shutil.copytree(ROOT/'stage1',copy/'stage1',dirs_exist_ok=True);file=copy/HERE.relative_to(ROOT)/change['file'];text=file.read_text();assert text.count(change['from'])==1;file.write_text(text.replace(change['from'],change['to'],1));check(change['name'],copy/HERE.relative_to(ROOT)/'components.a',copy/'native',copy/'emitted.mjs',True)
print('component cases',len(json.loads((HERE/'component_cases.json').read_text())),'bytes',len(want),flush=True)
