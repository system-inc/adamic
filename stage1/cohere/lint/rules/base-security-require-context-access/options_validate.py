#!/usr/bin/env python3
import json,subprocess,argparse,shutil
from pathlib import Path
HERE=Path(__file__).resolve().parent;ROOT=HERE.parents[4]
p=argparse.ArgumentParser();p.add_argument('--scratch',type=Path,required=True);a=p.parse_args();S=a.scratch.resolve();S.mkdir(parents=True,exist_ok=True)
def run(label,cmd,cwd=ROOT):
 with (S/(label+'.log')).open('wb') as out,(S/(label+'.stderr')).open('wb') as err:r=subprocess.run(list(map(str,cmd)),cwd=cwd,stdout=out,stderr=err)
 assert r.returncode==0 and not (S/(label+'.stderr')).read_bytes(),label
 return (S/(label+'.log')).read_bytes()
virtual=ROOT/'cohere/wave11_options.go';overlay=S/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(HERE/'options_oracle.go.txt')}}));run('oracle-build',['go','build','-overlay='+str(overlay),'-o',S/'oracle',virtual],ROOT/'cohere')
run('baseline-build',['go','run',HERE.parent/'adamic-no-definite-assignment/validation_build.go',HERE/'options_validation.ts',S/'native',S/'emitted.mjs']);want=run('Go',[S/'oracle',HERE/'options_cases.json'])
def check(prefix,source,native,js,mutated=False):
 for name,cmd in [('Node',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',source,HERE/'options_cases.json']),('emitted',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',js,HERE/'options_cases.json']),('native',[native,HERE/'options_cases.json'])]:
  got=run(prefix+'-'+name,cmd);assert (got!=want)==mutated,name;print(prefix,name,'byte comparison', 'caught' if mutated else 'equal',flush=True)
check('baseline',HERE/'options_validation.ts',S/'native',S/'emitted.mjs')
copy=S/'mutant';shutil.copytree(ROOT/'stage1',copy/'stage1',dirs_exist_ok=True);owned=copy/'stage1/cohere/lint/rules'/HERE.name;file=owned/'options.ts';text=file.read_text();old="shape.check(raw) !== 'valid'";assert text.count(old)==1;file.write_text(text.replace(old,"false && ("+old+")"));run('mutant-build',['go','run',HERE.parent/'adamic-no-definite-assignment/validation_build.go',owned/'options_validation.ts',copy/'native',copy/'emitted.mjs']);check('unknown-field-mutant',owned/'options_validation.ts',copy/'native',copy/'emitted.mjs',True)
