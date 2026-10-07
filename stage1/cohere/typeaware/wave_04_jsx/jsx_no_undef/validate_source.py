#!/usr/bin/env python3
"""Compare native undefined-name source findings with the unmodified Go production rule."""
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
virtual=r/'cohere/adamic_wave04_undef_oracle.go';overlay=out/'oracle-overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(own/'oracle.go')}}))
run('oracle-build',['go','build','-overlay',overlay,'-o',out/'oracle',virtual],r/'cohere')
run('native-build',[compiler,'build',own/'source_main.a','-o',out/'native','--tsgo',archive])
cases=['<Missing />','<Missing></Missing>','<div />','<Foo-bar />','<_foo />','<$foo />','<테스트 />','<app.Foo />','<app.Foo.Bar />','<this.foo />','<ns:Tag />','const App=null; const x=<App />','{ const App=null; } const x=<App />','enum A {App} const x=<App />',"import App from './missing'; const x=<App />",'declare global { var G:any; } const x=<G />','<Map />','<Promise />','/* 😀 é */ <Missing />','const app={Foo:null}; const x=<app.Foo />']
paths=[]
for i,case in enumerate(cases):
 f=out/f'control-{i:02}.tsx';f.write_text('export {};\n'+(case if 'const x=' in case else 'const x='+case)+';\n');paths.append(str(f))
manifest=out/'manifest';manifest.write_text('\n'.join(paths)+'\n');config=out/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'target':'ESNext','module':'ESNext','jsx':'preserve','strict':True,'types':[]},'files':[Path(paths[0]).name]}))
for mode in ['local','globals']:
 expected=run('controls-'+mode+'-go',[out/'oracle',config,manifest,mode]);actual=run('controls-'+mode+'-native',[out/'native',config,manifest,mode])
 assert not (out/('controls-'+mode+'-native.stderr')).read_bytes()
 if actual!=expected:
  first=next((i for i,(x,y) in enumerate(zip(actual,expected)) if x!=y),min(len(actual),len(expected)));raise RuntimeError(f'{mode}: byte mismatch at {first}')
 print(f'{mode}: {len(actual)} bytes equal',flush=True)
for name,config_,manifest_ in [('compiler',Path('/workspace/typeaware-wave-04/typescript/src/compiler/tsconfig.json'),Path('/workspace/typeaware-wave-04/compiler.manifest')),('repository',r/'tsconfig.json',Path('/workspace/typeaware-wave-04/repository.manifest'))]:
 expected=run(name+'-go',[out/'oracle',config_,manifest_,'local']);actual=run(name+'-native',[out/'native',config_,manifest_,'local'])
 assert not (out/('controls-'+mode+'-native.stderr')).read_bytes()
 assert not (out/(name+'-native.stderr')).read_bytes()
 if actual!=expected:raise RuntimeError(name+': corpus byte mismatch')
 print(f'{name}: {len(actual)} bytes equal',flush=True)
run('sanitize-build',[compiler,'build',own/'source_main.a','-o',out/'sanitize','--tsgo',a.sanitized_archive,'--sanitize'])
for mode in ['local','globals']:
 actual=run('sanitize-'+mode,[out/'sanitize',config,manifest,mode]);assert actual==(out/('controls-'+mode+'-go.stdout')).read_bytes();assert not (out/('sanitize-'+mode+'.stderr')).read_bytes()
# A compiled, successful mutant is rejected only by the byte comparison.
import re
source=(own/'source.a').read_text();before='this.rules.byte(name.end)';assert source.count(before)==1
source=source.replace(before,'this.rules.byte(name.end)+1')
source=re.sub(r"from '(\.[^']+)'",lambda m:"from '"+str((own/m[1]).resolve())+"'",source)
mutant=out/'span-mutant.a';mutant.write_text(source)
driver=(own/'source_main.a').read_text();driver=re.sub(r"from '(\.[^']+)'",lambda m:"from '"+str((own/m[1]).resolve())+"'",driver)
entry=out/'span-mutant-main.a';entry.write_text(driver.replace(str(own/'source.a'),str(mutant)))
run('mutant-build',[compiler,'build',entry,'-o',out/'mutant','--tsgo',a.sanitized_archive,'--sanitize'])
actual=run('mutant-run',[out/'mutant',config,manifest,'local']);assert actual!=(out/'controls-local-go.stdout').read_bytes();assert not (out/'mutant-run.stderr').read_bytes()
# The exact source visitor must refuse checker queries after releasing the program.
released=out/'released-main.a';released.write_text(driver.replace('let findings = 0;','tsgoRelease(program);\nlet findings = 0;'))
run('released-build',[compiler,'build',released,'-o',out/'released','--tsgo',archive])
with (out/'released.stdout').open('wb') as so,(out/'released.stderr').open('wb') as se: code=subprocess.run([str(out/'released'),str(config),str(manifest),'local'],stdout=so,stderr=se).returncode
assert code==70 and b'released' in (out/'released.stderr').read_bytes();results['released-run']={'exit':code};(out/'results.json').write_text(json.dumps(results,indent=2))
print('PASS: sanitizer controls; comparison-only span mutant; released program rejected with exit 70.',flush=True)
for name,config_,manifest_ in [('compiler',Path('/workspace/typeaware-wave-04/typescript/src/compiler/tsconfig.json'),Path('/workspace/typeaware-wave-04/compiler.manifest')),('repository',r/'tsconfig.json',Path('/workspace/typeaware-wave-04/repository.manifest'))]:
 for mode in ['local','globals']:
  expected=run(name+'-'+mode+'-go',[out/'oracle',config_,manifest_,mode]);actual=run(name+'-'+mode+'-sanitize',[out/'sanitize',config_,manifest_,mode]);assert actual==expected;assert not (out/(name+'-'+mode+'-sanitize.stderr')).read_bytes()
print('PASS: both corpus modes under ASan/UBSan/LeakSanitizer.',flush=True)
commonjs=out/'commonjs.cjs';commonjs.write_text('const x=<Map />;\nconst y=<Missing />;\n');commonmanifest=out/'commonjs.manifest';commonmanifest.write_text(str(commonjs)+'\n')
commonconfig=out/'commonjs-tsconfig.json';commonconfig.write_text(json.dumps({'compilerOptions':{'target':'ESNext','jsx':'preserve','allowJs':True,'types':[]},'files':[commonjs.name]}))
expected=run('commonjs-go',[out/'oracle',commonconfig,commonmanifest,'local']);actual=run('commonjs-sanitize',[out/'sanitize',commonconfig,commonmanifest,'local']);assert actual==expected;assert not (out/'commonjs-sanitize.stderr').read_bytes()
print('PASS: .cjs global-scope exception matches Go under sanitizers.',flush=True)
