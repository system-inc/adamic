#!/usr/bin/env python3
"""Executable blocker evidence, never a passing lint port."""
import json, os, subprocess, sys
from pathlib import Path
repo=Path(__file__).resolve().parents[4];source=Path(__file__).resolve().parent
artifacts=Path(sys.argv[1]).resolve();artifacts.mkdir(parents=True,exist_ok=True)
compiler=Path(os.environ.get('ADAMIC_COMPILER', '/workspace/wave29-regex-controls/adamic'));records=[]
def run(name,args,cwd=repo,expected=0):
 with (artifacts/(name+'.stdout')).open('wb') as out,(artifacts/(name+'.stderr')).open('wb') as err:
  p=subprocess.run([str(x) for x in args],cwd=cwd,stdout=out,stderr=err)
 records.append(dict(name=name,args=[str(x) for x in args],exit=p.returncode))
 (artifacts/'commands.json').write_text(json.dumps(records,indent=2)+'\n')
 assert expected is None or p.returncode==expected,(name,p.returncode)
 return (artifacts/(name+'.stdout')).read_text()
virtual=repo/'cohere/adamic_wave29_third_oracle.go'
(artifacts/'overlay.json').write_text(json.dumps({'Replace':{str(virtual):str(source/'testdata/oracle.go')}}))
run('oracle-build',['go','build','-overlay',artifacts/'overlay.json','-o',artifacts/'oracle',virtual],cwd=repo/'cohere')
ambient=artifacts/'react.d.ts'
ambient.write_text('declare module "react" {export type Dispatch<A>=(value:A)=>void;export type SetStateAction<S>=S|((prev:S)=>S);export function useState<S>(initial:S):[S,Dispatch<SetStateAction<S>>];export function useEffect(effect:()=>void):void;}\n')
paths=[]
for name in ['set_state_in_effect','set_state_in_render','static_components']:
 text=(source/'gaps'/(name+'.a')).read_text().split("const parser = new Parser('",1)[1].split("', 'input.tsx');",1)[0]
 # These are reference TypeScript inputs; the executable probes remain .a.
 path=artifacts/(name+'.tsx');path.write_text(text+'\n');paths.append(str(path))
(artifacts/'manifest').write_text('\n'.join(paths)+'\n')
(artifacts/'config.json').write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','lib':['ES2022'],'jsx':'preserve','module':'ESNext'},'files':paths+[str(ambient)]}))
truth=run('go-positive',[artifacts/'oracle',artifacts/'config.json',artifacts/'manifest'])
for rule in ['react-hooks/set-state-in-effect','react-hooks/set-state-in-render','react-hooks/static-components']:
 assert '\t'+rule+'\t' in truth,rule
assert truth.splitlines()[-1]=='findings 3',truth
print('independent Go oracle: three inputs, three positive findings; fixes and suggestions zero',flush=True)
for name in ['set_state_in_effect','set_state_in_render','static_components']:
 run(name+'-build',[compiler,'build',source/'gaps'/(name+'.a'),'-o',artifacts/name])
 kinds=run(name+'-parser',[artifacts/name],expected=None)
 error=(artifacts/(name+'-parser.stderr')).read_text()
 status=records[-1]['exit']
 assert status in [0,70],(name,status)
 if status==70:
  assert 'parser slice expected' in error,error
  print(name,'probe compiled; native parser exited 70:',error.strip(),flush=True)
 else:
  assert 'Jsx' not in kinds and 'TypeAssertionExpression' in kinds,kinds
  print(name,'probe compiled and exited 0, but produced TypeAssertionExpression and no JSX nodes',flush=True)
# Hook-only positive controls prove that adding JSX alone is insufficient.
for name,text in [
 ('effect-hook','import {useState,useEffect} from "react"; export function useThing(){const [s,setS]=useState(0);useEffect(()=>setS(1));return s;}'),
 ('render-hook','import {useState} from "react"; export function useThing(){const [s,setS]=useState(0);setS(1);return s;}'),
]:
 path=artifacts/(name+'.a');path.write_text(text+'\n')
 (artifacts/'hook-manifest').write_text(str(path)+'\n')
 (artifacts/'hook-config.json').write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','lib':['ES2022'],'module':'ESNext'},'files':[str(path),str(ambient)]}))
 truth=run(name+'-go',[artifacts/'oracle',artifacts/'hook-config.json',artifacts/'hook-manifest'])
 assert truth.splitlines()[-1]=='findings 1',truth
 probe=artifacts/(name+'-parse.a')
 probe.write_text("import { Parser } from '"+str(repo/'stage1/typescript/parser/parser.ts')+"';const parser=new Parser('"+text+"','input.a');console.log(`${parser.file()}`);\n")
 run(name+'-parse-build',[compiler,'build',probe,'-o',artifacts/(name+'-parse')])
 run(name+'-parse',[artifacts/(name+'-parse')])
 print(name,'Go reports one; native parser succeeds, but React SSA lowering/analysis has no native entry',flush=True)
