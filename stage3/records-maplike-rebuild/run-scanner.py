#!/usr/bin/env python3
"""Reproduce the 3ea66219 scanner profile using this compiler checkout."""
from pathlib import Path
import os, subprocess, shutil, json
repo=Path(__file__).resolve().parents[2]; out=Path(os.environ.get('RECORDS_SCANNER_OUTPUT', '/tmp/records-maplike-scanner-rebuild'));out.mkdir()
driver=out/'driver';driver.mkdir()
for name in ['run.py','main.a','node.mjs','inventory.cjs','scratch-corpus.cjs']:
 data=subprocess.check_output(['git','show','3ea66219:stage3/drivers/scanner/'+name],cwd=repo)
 if name=='run.py': data=data.replace(b'repository = here.parents[2]',('repository = Path(' + repr(str(repo)) + ')').encode())
 (driver/name).write_bytes(data)
pipeline=out/'pipeline';pipeline.mkdir()
for name in ['apply.sh','apply.py','source.json']: shutil.copyfile(repo/'stage3'/name,pipeline/name)
shutil.copytree(repo/'stage3/api',pipeline/'api',ignore=shutil.ignore_patterns('node_modules'))
(pipeline/'adapt').mkdir()
for name in ['00-setup','10-type-imports','42-scanner-any','50-temporary-scanner-implicit-returns','51-temporary-scanner-fallthrough']:
 if (repo/'stage3/adapt'/name).exists():
  shutil.copytree(repo/'stage3/adapt'/name,pipeline/'adapt'/name)
 else:
  target=pipeline/'adapt'/name;target.mkdir()
  for file in ['adapt.cjs']:
   (target/file).write_bytes(subprocess.check_output(['git','show','3ea66219:stage3/adapt/'+name+'/'+file],cwd=repo))
env=dict(os.environ,STAGE3_CACHE=os.environ.get('STAGE3_CACHE', str(Path.home()/'.cache/adamic-stage3')),GOPROXY='https://proxy.golang.org|direct')
def run(name,command,cwd=repo,extra=None,allow=False):
 with (out/(name+'.log')).open('wb') as log: code=subprocess.run([str(x) for x in command],cwd=cwd,env=dict(env,**(extra or {})),stdout=log,stderr=subprocess.STDOUT).returncode
 print(name,code,flush=True)
 if code and not allow: raise SystemExit(name+' failed; inspect '+str(out/(name+'.log')))
 return code
full=out/'full-tree';run('apply',['bash',pipeline/'apply.sh',full])
inputs=out/'inputs';shutil.copytree(full/'src/compiler',inputs/'src/compiler')
(inputs/'src/compiler/scanner.ts').write_bytes(subprocess.check_output(['git','-C',str(full),'show','HEAD:src/compiler/scanner.ts']))
run('corpus-type-imports',['node',driver/'scratch-corpus.cjs',inputs],extra={'NODE_PATH':env['STAGE3_CACHE']+'/api/node_modules'})
reference=out/'reference';run('reference',['python3',driver/'run.py',reference,'--tree',full,'--inputs',inputs,'--node-only'])
slice=out/'slice';run('slice',['bash',repo/'stage3/slice/run.sh',full,slice,'src/compiler/scanner.ts:createScanner','src/compiler/types.ts:ScriptTarget','src/compiler/types.ts:SyntaxKind'],extra={'SLICE_TYPESCRIPT':env['STAGE3_CACHE']+'/api/node_modules/typescript/lib/typescript.js'})
for split in ['0','1']:
 run('scanner-'+split,['python3',driver/'run.py',out/('split-'+split),'--tree',slice,'--inputs',inputs,'--oracle',reference/'node.stdout','--compiler',os.environ.get('RECORDS_SCANNER_COMPILER', '/tmp/records-maplike-adamic'),'--compiler-cwd',repo],extra={'ADAMIC_NATIVE_SPLIT':split,'ADAMIC_NATIVE_JOBS':'5'},allow=True)
print('Reports retained under '+str(out),flush=True)
