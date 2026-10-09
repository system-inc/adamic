#!/usr/bin/env python3
"""Prove reference evidence assertions reject corrupted observations."""
import gzip,json,shutil,subprocess,tempfile
from pathlib import Path
here=Path(__file__).resolve().parent;source=here/'evidence/project-references'
def check(e,log):return subprocess.run(['python3',str(here/'verify-references.py'),str(e)],stdout=log,stderr=subprocess.STDOUT).returncode
def text(e,name,value):(e/(name+'.gz')).write_bytes(gzip.compress(value.encode(),mtime=0))
def edit(e,name,change):
 p=e/name;d=json.loads(p.read_text());change(d);p.write_text(json.dumps(d))
with tempfile.TemporaryDirectory() as t:
 with (Path(t)/'check.log').open('wb') as log:assert check(source,log)==0
mutants=[('build-exit',lambda e:(e/'tsc-build.exit').write_text('1\n')),('build-order',lambda e:text(e,'tsc-build.stdout',"Building project '/app/tsconfig.json'...\nBuilding project '/dependency/tsconfig.json'...\n")),('node-oracle',lambda e:text(e,'node-emitted.stdout','8\n')),('missing-output-error',lambda e:text(e,'adamic-before.stderr','different error\n')),('first-source-stop',lambda e:text(e,'entry-0.stderr','wrong\n')),('first-lowering-stop',lambda e:text(e,'entry-types-0.stderr','wrong\n')),('referenced-root',lambda e:edit(e,'stock-entry-with-types.json',lambda d:d['roots'].pop())),('source-diagnostic',lambda e:edit(e,'stock-entry-without-types.json',lambda d:d['diagnostics'][0].update(code=9999))),('ambient-options',lambda e:edit(e,'stock-entry-with-types.json',lambda d:d['options'].update(types=[]))),('source-identity',lambda e:edit(e,'identity.json',lambda d:d['after'].update(fake='changed'))),('compiler-territory',lambda e:edit(e,'identity.json',lambda d:d.update(publication_compiler_diff='changed'))),('unchecked-dependency',lambda e:text(e,'mutant-flat-load.stdout',json.dumps({'loaded':True,'error':''}))),('source-reference-trace',lambda e:edit(e,'loader-trace.json',lambda d:d.update(loadInput_references='no project reference')))]
for name,mutate in mutants:
 with tempfile.TemporaryDirectory(prefix='tsc-reference-mutant-') as t:
  root=Path(t);e=root/'evidence';shutil.copytree(source,e);mutate(e)
  with (root/'check.log').open('wb') as log:assert check(e,log)!=0,name
  print(name+': caught by verify-references.py',flush=True)
print('PASS: thirteen evidence mutants rejected; real unimported-source mutant also caught by stock tsc and scratch loader')
