#!/usr/bin/env python3
"""Complete wire comparison against an independent unmodified production Go rule."""
import argparse,json,subprocess,time,re
from pathlib import Path
from controls import PREFIX,controls
p=argparse.ArgumentParser();p.add_argument('directory',type=Path);p.add_argument('--adamic',type=Path,default=Path('/workspace/typeaware-wave-04-landing-d65/adamic'));p.add_argument('--archive',type=Path,default=Path('/workspace/typeaware-wave-04-landing-d65/next/checker-normal.a'));p.add_argument('--sanitized-archive',type=Path,default=Path('/workspace/typeaware-wave-04-landing-d65/next/checker-asan.a'));p.add_argument('--normal-only',action='store_true');a=p.parse_args()
own=Path(__file__).resolve().parent;r=own.parents[4];out=a.directory.resolve();out.mkdir(parents=True,exist_ok=True);results={}
def run(name,cmd,cwd=r,expected=0):
 start=time.monotonic()
 with (out/(name+'.stdout')).open('wb') as so,(out/(name+'.stderr')).open('wb') as se:code=subprocess.run(list(map(str,cmd)),cwd=cwd,stdout=so,stderr=se).returncode
 results[name]={'seconds':time.monotonic()-start,'exit':code};(out/'results.json').write_text(json.dumps(results,indent=2))
 if code!=expected:raise RuntimeError(f'{name}: exit {code}; see {out/(name+".stderr")}')
 return (out/(name+'.stdout')).read_bytes()
virtual=r/'cohere/adamic_wave04_context_oracle.go';overlay=out/'oracle-overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(own/'oracle.go')}}))
run('oracle-build',['go','build','-overlay',overlay,'-o',out/'oracle',virtual],r/'cohere')
run('native-build',[a.adamic,'build',own/'source_main.a','-o',out/'native','--tsgo',a.archive])
paths=[]
for at,case in enumerate(controls()):
 f=out/f'control-{at:03}.tsx';f.write_text(PREFIX+case+'\n');paths.append(str(f))
helper=out/'helper.tsx';helper.write_text('export function make(){return {};}\nexport const ImportedCtx={};\n');paths.append(str(helper))
f=out/'foreign.tsx';f.write_text(PREFIX+"import {make,ImportedCtx} from './helper'; function Component(){const dep=make();const value=useMemo(()=>({dep}),[dep]);return <ImportedCtx.Provider value={value}/>;}\n");paths.append(str(f))
f=out/'imported-provider.tsx';f.write_text(PREFIX+"import {ImportedCtx} from './helper'; function Component(){return <ImportedCtx value={{}}/>;}\n");paths.append(str(f))
manifest=out/'manifest';manifest.write_text('\n'.join(paths)+'\n');config=out/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'target':'ESNext','module':'ESNext','jsx':'preserve','strict':True,'types':[]},'files':[Path(paths[0]).name]}))
corpora=[('controls',config,manifest),('compiler',Path('/workspace/typeaware-wave-04/typescript/src/compiler/tsconfig.json'),Path('/workspace/typeaware-wave-04/compiler.manifest')),('repository',r/'tsconfig.json',Path('/workspace/typeaware-wave-04/repository.manifest'))]
truths={}
for name,cfg,roots in corpora:
 truth=run(name+'-go',[out/'oracle',cfg,roots]);actual=run(name+'-native',[out/'native',cfg,roots]);assert not (out/(name+'-native.stderr')).read_bytes();truths[name]=truth
 if actual!=truth:
  first=next((i for i,(x,y) in enumerate(zip(actual,truth)) if x!=y),min(len(actual),len(truth)));raise RuntimeError(f'{name}: first differing byte {first}')
 print(f'PASS {name}: {len(actual)} bytes identical',flush=True)
for message in ['defaultMsg','defaultMsgFunc','withIdentifierMsg','withIdentifierMsgFunc','memoWithoutDependenciesMsg','unstableDependencyMsg']:assert ('\t'+message+'\t').encode() in truths['controls'],message
print(f'PASS positive controls: {len(paths)} roots, all six message IDs',flush=True)
if a.normal_only:raise SystemExit(0)
run('sanitize-build',[a.adamic,'build',own/'source_main.a','-o',out/'sanitize','--tsgo',a.sanitized_archive,'--sanitize'])
for name,cfg,roots in corpora:
 actual=run(name+'-sanitize',[out/'sanitize',cfg,roots]);assert actual==truths[name] and not (out/(name+'-sanitize.stderr')).read_bytes()
print('PASS corpora under ASan/UBSan/LeakSanitizer',flush=True)
# Copy all owned modules together, preserving nominal class identities in each mutant.
mutations=[('span','source.a','origin.source.rules.byte(origin.source.node(origin.index).end)','origin.source.rules.byte(origin.source.node(origin.index).end)+1'),('depth','stability.a','if(depth>8||this.active.has(key)||function_.source.body','if(depth>1||this.active.has(key)||function_.source.body'),('escape','stability.a','if(result.fresh!==undefined&&result.other&&this.anyEscapes(own,false))','if(false)'),('primitive','stability.a','if(this.primitive(node))','if(false)')]
for label,module,before,after in mutations:
 directory=out/(label+'-modules');directory.mkdir(exist_ok=True)
 for source in own.glob('*.a'):
  text=source.read_text();text=re.sub(r"from '(\.\.[^']+)'",lambda m:"from '"+str((source.parent/m[1]).resolve())+"'",text)
  if source.name==module:assert text.count(before)==1;text=text.replace(before,after)
  (directory/source.name).write_text(text)
 run(label+'-build',[a.adamic,'build',directory/'source_main.a','-o',out/label,'--tsgo',a.sanitized_archive,'--sanitize'])
 actual=run(label+'-run',[out/label,config,manifest]);assert actual!=truths['controls'] and not (out/(label+'-run.stderr')).read_bytes(),label
 print(f'PASS {label}: compiled, exit zero, clean sanitizer stderr; only byte comparison catches mutant',flush=True)
driver=(own/'source_main.a').read_text();driver=re.sub(r"from '(\.[^']+)'",lambda m:"from '"+str((own/m[1]).resolve())+"'",driver);released=out/'released-main.a';released.write_text(driver.replace('let findings = 0;','tsgoRelease(program);\nlet findings = 0;'))
run('released-build',[a.adamic,'build',released,'-o',out/'released','--tsgo',a.archive]);run('released-run',[out/'released',config,manifest],expected=70);assert (out/'released-run.stderr').read_bytes()==b'adamic: panic: invalid or released checker handle\n'
print('PASS released checker handle rejected with exit 70',flush=True)
