#!/usr/bin/env python3
"""Reproduce the shared JSX frontend gap without editing the frontend or harness."""
import json,os,re,subprocess,gzip
from pathlib import Path
ROOT=Path(__file__).resolve().parents[4];HERE=Path(__file__).resolve().parent
OUT=Path(os.environ.get('WAVE05_REACT_ARTIFACTS','/workspace/wave-05-react-probe'));OUT.mkdir(exist_ok=True)
def run(name,args,cwd=ROOT):
    with (OUT/(name+'.stdout')).open('wb') as stdout,(OUT/(name+'.stderr')).open('wb') as stderr:
        result=subprocess.run(list(map(str,args)),cwd=cwd,stdout=stdout,stderr=stderr)
    (OUT/(name+'.status')).write_text(str(result.returncode)+'\n')
    return result.returncode,(OUT/(name+'.stdout')).read_bytes(),(OUT/(name+'.stderr')).read_bytes()
stage0=OUT/'adamic'
assert run('stage0',['go','build','-o',stage0,'./cmd/adamic'])[0]==0
native=OUT/'native';assert run('native-build',[stage0,'build',HERE/'gaps/jsx_probe.a','-o',native])[0]==0
virtual=ROOT/'cohere/adamic_wave05_react_oracle.go';overlay=OUT/'oracle-overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(HERE/'testdata/oracle.go')}}))
oracle=OUT/'oracle';assert run('oracle-build',['go','build','-overlay',overlay,'-o',oracle,virtual],ROOT/'cohere')[0]==0
# This is the production Go rule tests' exact React declaration stub.
text=(ROOT/'cohere/internal/lint/rules/react/set_state_in_effect_test.go').read_text()
match=re.search(r'^const reactStub = (".*")$',text,re.M);assert match
(OUT/'react.d.ts').write_text(json.loads(match[1]))
cases=[
    ('globals','let globalCount=0;export function Component(){globalCount=1;return <div />;}\n','react-hooks/globals'),
    ('immutability','export function Component(props:{value:number}){props.value=1;return <div />;}\n','react-hooks/immutability'),
    ('derived','import {useState,useEffect} from "./react";export function Component({value}:{value:number}){const [state,setState]=useState(0);useEffect(()=>{setState(value);},[value]);return <div />;}\n','react-hooks/no-deriving-state-in-effects'),
]
config=OUT/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','lib':['ES2022','DOM'],'jsx':'preserve'},'files':[str(OUT/'react.d.ts')]}))
for name,text,rule in cases:
    source=OUT/(name+'.tsx');source.write_text(text);manifest=OUT/(name+'.manifest');manifest.write_text(str(source)+'\n')
    status,truth,error=run(name+'-go',[oracle,config,manifest]);assert status==0 and ('\t'+rule+'\t').encode() in truth,(name,status,truth,error)
    status,actual,error=run(name+'-native',[native,source]);assert status==70 and not actual and b'expected GreaterThanToken, got SlashToken' in error,(name,status,actual,error)
    print(name,'Go reports; native parser refuses JSX, exit 70',flush=True)
negative=OUT/'negative.ts';negative.write_text('let g=0;export function helper(){g=1;return null;}\n')
status,actual,error=run('negative-native',[native,negative]);assert status==0 and actual==b'jsx nodes 0\n' and not error
negative_manifest=OUT/'negative.manifest';negative_manifest.write_text(str(negative)+'\n')
status,truth,error=run('negative-go',[oracle,config,negative_manifest]);assert status==0 and truth.endswith(b'findings 0\n')
print('non-JSX control parses natively and is clean in Go',flush=True)
mutant=OUT/'mutant.a';text=(HERE/'gaps/jsx_probe.a').read_text();assert text.count('parser.file();')==1
text=text.replace('parser.file();','// mutant: bypass parsing');text=text.replace("'../../../../typescript/", "'"+str(ROOT/'stage1/typescript')+'/');mutant.write_text(text)
exe=OUT/'mutant';assert run('mutant-build',[stage0,'build',mutant,'-o',exe])[0]==0
status,actual,error=run('mutant-run',[exe,OUT/'globals.tsx']);assert status==0 and not error
assert status!=70,'refusal assertion did not catch skipped-parser mutant'
print('probe no-op mutant exits 0; expected-refusal assertion catches it (not a rule mutant)',flush=True)
print('PASS exact JSX gap reproduction; no React port claimed',flush=True)
