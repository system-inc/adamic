"""Actual source scoped constants against independent Go checker and reference helpers."""
import pathlib,json,subprocess,time,hashlib
root=pathlib.Path(__file__).resolve().parent;repo=root.parents[4]
work=pathlib.Path('/workspace/wave-09-bindings');work.mkdir(exist_ok=True)
def write(name,text):p=work/name;p.write_text(text);return p
def run(name,args,cwd=repo):
 start=time.monotonic()
 with (work/(name+'.stdout')).open('wb') as out,(work/(name+'.stderr')).open('wb') as err:r=subprocess.run(list(map(str,args)),cwd=cwd,stdout=out,stderr=err,timeout=120)
 print(name,r.returncode,round(time.monotonic()-start,6),flush=True)
 assert r.returncode==0,(name,(work/(name+'.stderr')).read_text()[:2000])
 return (work/(name+'.stdout')).read_bytes()
sources=[
 'const a="[Á]";probe(a);',
 'let a="[👍]";probe(a);',
 'var a="u";probe(a);',
 'let a="u";a="v";probe(a);',
 'let a="u";a+="v";probe(a);',
 'let a="u";a++;probe(a);',
 'let a="u";({x:a}={x:"v"});probe(a);',
 'let a="u";[a]=["v"];probe(a);',
 'let a="u";[...a]=["v"];probe(a);',
 'let a="u";[...(a)]=["v"];probe(a);',
 'let a="u";for(a of ["v"]){}probe(a);',
 'let a="u";for(let a of ["v"]){}probe(a);',
 'let a="u";a.x="v";probe(a);',
 'let a="u";obj[a]="v";probe(a);',
 'const a="u",b=a,c=b;probe(c);',
 'const a=a;probe(a);',
 'const a=b,b=a;probe(a);',
 'const a="u";{const a="v";probe(a);}probe(a);',
 'let a="u";function f(){let a="v";a="x";}probe(a);',
 'let a="u";function f(){a="v";}probe(a);',
 'const a=/[👍]/;const b=a;probe(a);probe(b);probe(b+"");',
 'let a=/[👍]/;a=/[Á]/;probe(a);',
 'const {a}={a:"u"};probe(a);',
 'const [a]=["u"];probe(a);',
 'let a;probe(a);',
 'var a="u";var a;probe(a);',
 'const a="u";probe(`head${a}${true}${null}tail`);',
 'const a=1;probe(a);probe(a+"u");probe(a+2);',
 'function f(a:string){probe(a);}probe(missing);',
 'const a="u";probe((a));probe(a as string);',
 'const a="u";a="v";probe(a);',
 'import {a} from "./helper.js";probe(a);',
 'const a="u";({files=a}=obj);probe(a);',
 'const a="u";probe({a});',
 'const a="u";probe(String.raw`u`);'
]
paths=[str(write(f'control-{i}.a','/* 世界 🌍 */\r\n'+source+'\r\nexport {};\r\n')) for i,source in enumerate(sources)]
write('helper.a','export const a="u";')
manifest=write('manifest','\n'.join(paths)+'\n');write('root.d.ts','declare function probe(value:any):void;declare const obj:any;')
config=write('tsconfig.json',json.dumps({'compilerOptions':{'target':'ES2022','module':'NodeNext','moduleResolution':'NodeNext','lib':['es2022','dom']},'files':['root.d.ts']}))
virtual=repo/'cohere/wave09_binding_oracle.go';overlay=write('overlay.json',json.dumps({'Replace':{str(virtual):str(root/'testdata/binding_oracle.go')}}))
go=work/'go';run('go-build',['go','build','-overlay',overlay,'-o',go,virtual],repo/'cohere');truth=run('go',[go,config,manifest])
for mode in ['normal','sanitized','write-mutant','cycle-mutant']:
 source=root/'binding_main.a'
 if mode.endswith('mutant'):
  text=(root/'constant_bindings.a').read_text()
  for file in ['bindings.ts','reassign.ts','declaration_file_flags.a']:text=text.replace("'../../"+file+"'",json.dumps(str(repo/'stage1/cohere/typeaware'/file)))
  text=text.replace("'./constant_value.a'",json.dumps(str(root/'constant_value.a'))).replace("'../../rules.ts'",json.dumps(str(repo/'stage1/cohere/typeaware/rules.ts')))
  anchor='!this.writes.writes(at)' if mode=='write-mutant' else 'following.includes(declaration)'
  replacement='true' if mode=='write-mutant' else 'following.length>0'
  assert anchor in text
  mutated=write(mode+'.a',text.replace(anchor,replacement,1))
  text=(root/'binding_main.a').read_text().replace("'./constant_bindings.a'",json.dumps(str(mutated)))
  text=text.replace("'../../", "'"+str(repo/'stage1/cohere/typeaware')+'/').replace(str(repo/'stage1/cohere/typeaware')+'/../../typescript/',str(repo/'stage1/typescript')+'/')
  source=write(mode+'-main.a',text)
 binary=work/mode;args=['/workspace/wave-09-core/adamic','build',source,'-o',binary,'--tsgo','/workspace/wave-09-core/'+('checker-asan.a' if mode=='sanitized' else 'checker.a')]
 if mode=='sanitized':args.append('--sanitize')
 run(mode+'-build',args);actual=run(mode,[binary,config,manifest]);assert not (work/(mode+'.stderr')).read_bytes()
 if mode.endswith('mutant'):
  assert actual!=truth;diff=next(i for i,(a,b) in enumerate(zip(actual,truth)) if a!=b);print(mode,'caught at byte',diff,flush=True)
 elif actual!=truth:
  diff=next((i for i,(a,b) in enumerate(zip(actual,truth)) if a!=b),min(len(actual),len(truth)));raise AssertionError((mode,diff,actual[max(0,diff-80):diff+100],truth[max(0,diff-80):diff+100]))
 else:print(mode,'identical bytes',len(truth),flush=True)
write('results.json',json.dumps({'files':len(sources),'bytes':len(truth),'sha256':hashlib.sha256(truth).hexdigest()},indent=2)+'\n')
print('PASS native scoped constant strings and regexp identity, independent Go bytes, sanitizers and eligibility mutants',flush=True)
