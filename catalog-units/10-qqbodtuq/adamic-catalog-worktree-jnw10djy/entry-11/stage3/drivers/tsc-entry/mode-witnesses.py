#!/usr/bin/env python3
"""Run project-scope and first-stop witnesses with a disposable compiler binary."""
import argparse,json,os,shutil,subprocess
from pathlib import Path
p=argparse.ArgumentParser();p.add_argument('compiler',type=Path);p.add_argument('probe',type=Path);p.add_argument('output',type=Path);a=p.parse_args();here=Path(__file__).resolve().parent;out=a.output.resolve();out.mkdir();binary=a.compiler.resolve();probe=a.probe.resolve()
source=here/'evidence/mode-c09c22b8/witnesses'
for name in ['app','dependency','outside','ordinary-09','ordinary-14']:
 folder=out/name;folder.mkdir();shutil.copy2(source/name/'tsconfig.json',folder/'tsconfig.json')
 authored={'app':'missing-project-output/app/main.a','dependency':'missing-project-output/dependency/value.a','outside':'project-indexed-read.a','ordinary-09':'09-never-argument.a','ordinary-14':'14-void-result.a'}[name]
 text=(here/'probes'/authored).read_text();text=text.replace('../dependency/value.a','../dependency/value.js')
 (folder/('value.ts' if name=='dependency' else 'main.ts')).write_text(text)
for name in ['app','outside','ordinary-09','ordinary-14']:
 folder=out/name
 for label,cmd in [('build',[str(binary),'build',str(folder/'main.ts'),'-o',str(folder/'native')]),('node',['node',str(here/'node.mjs'),str(folder/'main.ts')]),('load',[str(probe),str(folder/'main.ts')])]:
  with (folder/(label+('.json' if label=='load' else '.stdout'))).open('wb') as stdout,(folder/(label+'.stderr')).open('wb') as stderr:r=subprocess.run(cmd,stdout=stdout,stderr=stderr)
  (folder/(label+'.exit')).write_text(str(r.returncode)+'\n')
