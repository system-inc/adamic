#!/usr/bin/env python3
"""Compare native fragment source findings with the unmodified Go production rule."""
import argparse,json,subprocess,time,os
from pathlib import Path
p=argparse.ArgumentParser();p.add_argument('directory',type=Path);p.add_argument('--adamic',type=Path,default=Path('/workspace/typeaware-wave-04-landing-d65/adamic'));p.add_argument('--archive',type=Path,default=Path('/workspace/typeaware-wave-04-landing-d65/next/checker-normal.a'));p.add_argument('--sanitized-archive',type=Path,default=Path('/workspace/typeaware-wave-04-landing-d65/next/checker-asan.a'));a=p.parse_args()
r=Path(__file__).resolve().parents[5];own=Path(__file__).resolve().parent;out=a.directory.resolve();out.mkdir(parents=True,exist_ok=True)
compiler=a.adamic;archive=a.archive
results={}
def run(name,cmd,cwd=r):
 start=time.monotonic()
 with (out/(name+'.stdout')).open('wb') as so,(out/(name+'.stderr')).open('wb') as se: code=subprocess.run(list(map(str,cmd)),cwd=cwd,stdout=so,stderr=se).returncode
 results[name]={'seconds':time.monotonic()-start,'exit':code};(out/'results.json').write_text(json.dumps(results,indent=2))
 if code:raise RuntimeError(f'{name}: exit {code}; see {out/name}.stderr')
 return (out/(name+'.stdout')).read_bytes()
virtual=r/'cohere/adamic_wave04_fragments_oracle.go';overlay=out/'oracle-overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(own/'oracle.go')}}))
run('oracle-build',['go','build','-overlay',overlay,'-o',out/'oracle',virtual],r/'cohere')
run('native-build',[compiler,'build',own/'source_main.a','-o',out/'native','--tsgo',archive])
cases=['<React.Fragment />','<React.Fragment>hello</React.Fragment>','<>hello</>','<React.Fragment key="x" />','<React.Fragment {...props} />','<Other.Fragment />',"import {Fragment} from 'react'; const x=<Fragment />", "import {Fragment as F} from 'react'; const x=<F />", "import {Other as F} from 'react'; const x=<F />", "import {Fragment as F} from 'preact'; const x=<F />",'const F=React.Fragment; const x=<F />','const F=React; const x=<F />',"const F=require('react'); const x=<F />",'const {Fragment:F}=React; const x=<F />','const {Other:F}=React; const x=<F />','const [F]=React; const x=<F />','const F=(React.Fragment); const x=<F />','/* 😀 é */ <React.Fragment />','<><React.Fragment /></>','<React.Fragment value={1} />']
paths=[]
for i,case in enumerate(cases):
 f=out/f'control-{i:02}.tsx';f.write_text('export {};\n'+(case if 'const x=' in case else 'const x='+case)+';\n');paths.append(str(f))
manifest=out/'manifest';manifest.write_text('\n'.join(paths)+'\n');config=out/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'target':'ESNext','module':'ESNext','jsx':'preserve','strict':True,'types':[]},'files':[Path(paths[0]).name]}))
for mode in ['syntax','element']:
 expected=run('controls-'+mode+'-go',[out/'oracle',config,manifest,mode]);actual=run('controls-'+mode+'-native',[out/'native',config,manifest,mode])
 assert not (out/('controls-'+mode+'-native.stderr')).read_bytes()
 if actual!=expected:
  first=next((i for i,(x,y) in enumerate(zip(actual,expected)) if x!=y),min(len(actual),len(expected)));raise RuntimeError(f'{mode}: byte mismatch at {first}')
 print(f'{mode}: {len(actual)} bytes equal',flush=True)
for name,config_,manifest_ in [('compiler',Path('/workspace/typeaware-wave-04/typescript/src/compiler/tsconfig.json'),Path('/workspace/typeaware-wave-04/compiler.manifest')),('repository',r/'tsconfig.json',Path('/workspace/typeaware-wave-04/repository.manifest'))]:
 expected=run(name+'-go',[out/'oracle',config_,manifest_,'syntax']);actual=run(name+'-native',[out/'native',config_,manifest_,'syntax'])
 assert not (out/('controls-'+mode+'-native.stderr')).read_bytes()
 assert not (out/(name+'-native.stderr')).read_bytes()
 if actual!=expected:raise RuntimeError(name+': corpus byte mismatch')
 print(f'{name}: {len(actual)} bytes equal',flush=True)
run('sanitize-build',[compiler,'build',own/'source_main.a','-o',out/'sanitize','--tsgo',a.sanitized_archive,'--sanitize'])
for mode in ['syntax','element']:
 actual=run('sanitize-'+mode,[out/'sanitize',config,manifest,mode]);assert actual==(out/('controls-'+mode+'-go.stdout')).read_bytes();assert not (out/('sanitize-'+mode+'.stderr')).read_bytes()
# A compiled, successful mutant is rejected only by the byte comparison.
import re
source=(own/'source.a').read_text();before='this.rules.byte(node.end),preferElement';assert source.count(before)==1
source=source.replace(before,'this.rules.byte(node.end)+1,preferElement')
source=re.sub(r"from '(\.[^']+)'",lambda m:"from '"+str((own/m[1]).resolve())+"'",source)
mutant=out/'span-mutant.a';mutant.write_text(source)
driver=(own/'source_main.a').read_text();driver=re.sub(r"from '(\.[^']+)'",lambda m:"from '"+str((own/m[1]).resolve())+"'",driver)
entry=out/'span-mutant-main.a';entry.write_text(driver.replace(str(own/'source.a'),str(mutant)))
run('mutant-build',[compiler,'build',entry,'-o',out/'mutant','--tsgo',a.sanitized_archive,'--sanitize'])
actual=run('mutant-run',[out/'mutant',config,manifest,'syntax']);assert actual!=(out/'controls-syntax-go.stdout').read_bytes();assert not (out/'mutant-run.stderr').read_bytes()
# The exact source visitor must refuse checker queries after releasing the program.
released=out/'released-main.a';released.write_text(driver.replace('let findings = 0;','tsgoRelease(program);\nlet findings = 0;'))
run('released-build',[compiler,'build',released,'-o',out/'released','--tsgo',archive])
with (out/'released.stdout').open('wb') as so,(out/'released.stderr').open('wb') as se: code=subprocess.run([str(out/'released'),str(config),str(manifest),'syntax'],stdout=so,stderr=se).returncode
assert code==70 and b'released' in (out/'released.stderr').read_bytes();results['released-run']={'exit':code};(out/'results.json').write_text(json.dumps(results,indent=2))
print('PASS: sanitizer controls; comparison-only span mutant; released program rejected with exit 70.',flush=True)
for name,config_,manifest_ in [('compiler',Path('/workspace/typeaware-wave-04/typescript/src/compiler/tsconfig.json'),Path('/workspace/typeaware-wave-04/compiler.manifest')),('repository',r/'tsconfig.json',Path('/workspace/typeaware-wave-04/repository.manifest'))]:
 for mode in ['syntax','element']:
  expected=run(name+'-'+mode+'-go',[out/'oracle',config_,manifest_,mode]);actual=run(name+'-'+mode+'-sanitize',[out/'sanitize',config_,manifest_,mode]);assert actual==expected;assert not (out/(name+'-'+mode+'-sanitize.stderr')).read_bytes()
print('PASS: both corpus modes under ASan/UBSan/LeakSanitizer.',flush=True)
