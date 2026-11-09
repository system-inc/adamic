#!/usr/bin/env python3
"""Run from the repository root; every subprocess's own exit code is recorded."""
from pathlib import Path
import json, os, shlex, subprocess, sys
from concurrent.futures import ThreadPoolExecutor
import hashlib
root=Path(__file__).resolve().parents[3]
folder=Path(__file__).resolve().parent
node=os.environ.get('FIXTURE_NODE','node')
tsc=os.environ.get('FIXTURE_TSC',str(root/'stage3/api/node_modules/typescript/bin/tsc'))
logs=folder/'receiver-logs'; logs.mkdir(exist_ok=True)
def run(args,log):
 with log.open('w') as f:
  child=subprocess.run(args,stdout=f,stderr=subprocess.STDOUT)
 print(f'{shlex.join(args)}: exit={child.returncode}',flush=True)
 return child.returncode,log.read_text()
versions={}
for label,args in [('node',[node,'--version']),('tsc',[node,tsc,'--version'])]:
 rc,output=run(args,logs/f'{label}-version.log')
 if rc: sys.exit(rc)
 versions[label]=output.strip()
if not versions['node'].startswith('v24.') or versions['tsc']!='Version 6.0.3':
 raise SystemExit('Required Node v24 and TypeScript 6.0.3')
prefixes=sys.argv[1:]
fixtures=[f for f in sorted((folder/'receivers').glob('*.ts')) if not prefixes or f.name.startswith(tuple(prefixes))]
def check(fixture):
 args=[node,tsc,'--ignoreConfig','--strict','--noUncheckedIndexedAccess','--exactOptionalPropertyTypes','--noEmit','--lib','es2020,dom','--typeRoots',str(folder/'empty-types'),str(fixture)]
 rc,diagnostic=run(args,logs/(fixture.stem+'.tsc.log'))
 runtime=None; runtime_rc=None
 if rc==0:
  runtime_rc,runtime=run([node,'--experimental-strip-types',str(fixture)],logs/(fixture.stem+'.node.log'))
 return dict(sha256=hashlib.sha256(fixture.read_bytes()).hexdigest(),fixture=fixture.name,command=shlex.join(args),exit=rc,diagnostic=diagnostic,runtime=runtime,runtime_exit=runtime_rc)
with ThreadPoolExecutor(max_workers=4) as pool:
 rows=list(pool.map(check,fixtures))
rows.sort(key=lambda r:r['fixture'])
Path(os.environ.get('FIXTURE_RESULTS',str(folder/'receiver-results.json'))).write_text(json.dumps(dict(versions=versions,rows=rows),indent=2)+'\n')
