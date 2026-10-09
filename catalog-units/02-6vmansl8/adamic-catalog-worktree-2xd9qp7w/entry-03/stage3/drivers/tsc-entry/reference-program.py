#!/usr/bin/env python3
"""Recreate the two-project reference witness using stock Node TypeScript."""
import argparse,json,os,shutil,subprocess
from pathlib import Path
p=argparse.ArgumentParser();p.add_argument('compiler',type=Path);p.add_argument('loader_probe',type=Path);p.add_argument('output',type=Path);a=p.parse_args();here=Path(__file__).resolve().parent;out=a.output.resolve();out.mkdir();w=out/'two-project';w.mkdir()
for name in ['app','dependency']:
 folder=w/name;folder.mkdir()
 shutil.copy2(here/'evidence/mode-c09c22b8/witnesses'/name/'tsconfig.json',folder/'tsconfig.json')
 original=here/'probes/missing-project-output'/name/('main.a' if name=='app' else 'value.a')
 (folder/('main.ts' if name=='app' else 'value.ts')).write_text(original.read_text().replace('../dependency/value.a','../dependency/value.js'))
(w/'package.json').write_text('{"type":"module"}\n')
def run(name,cmd):
 with (out/(name+'.stdout')).open('wb') as stdout,(out/(name+'.stderr')).open('wb') as stderr:r=subprocess.run(cmd,cwd=w,stdout=stdout,stderr=stderr)
 (out/(name+'.exit')).write_text(str(r.returncode)+'\n')
tsc=str(Path(os.environ['SCANNER_TYPESCRIPT']).with_name('tsc.js'))
run('adamic-before',[str(a.compiler.resolve()),'build',str(w/'app/main.ts'),'-o',str(out/'before-native')])
run('loader-before',[str(a.loader_probe.resolve()),str(w/'app/main.ts')])
run('tsc-single',['node',tsc,'--ignoreConfig','--noEmit','--strict','--target','es2020','--module','esnext','--moduleResolution','bundler',str(w/'app/main.ts'),str(w/'dependency/value.ts')])
run('tsc-build',['node',tsc,'--build',str(w/'app/tsconfig.json'),'--verbose'])
run('node-emitted',['node',str(w/'out/app/main.js')])
run('loader-after',[str(a.loader_probe.resolve()),str(w/'app/main.ts')])
print('Two-project observations saved; compile failure exits are evidence, not success.')
