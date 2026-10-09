#!/usr/bin/env python3
"""Run a real one-function change, missed-function mutant and stale-map mutant."""
import argparse
from collections import Counter
import copy
import json
import os
from pathlib import Path
import shutil
import shlex
import subprocess
import sys
import time
from differential import load_map, save_map, covered_indices, compact_map, identity, SEED, sha, inventory
from run import PIN
ROOT=Path(__file__).resolve().parent
p=argparse.ArgumentParser(description=__doc__)
p.add_argument('map',type=Path);p.add_argument('source',type=Path);p.add_argument('output',type=Path)
a=p.parse_args();out=a.output.resolve();out.mkdir(parents=True,exist_ok=False);source=a.source.resolve()
map_path=a.map.resolve();mapping=load_map(map_path)
selection=json.loads((ROOT/'selection.json').read_text())['cases']
sample=sorted(selection,key=lambda row:sha((SEED+identity(row)).encode()))[:8]
sampled_functions={index for row in sample for index in covered_indices(mapping['configurations'][identity(row)])}
frequency=Counter(index for row in mapping['configurations'].values() for index in covered_indices(row))
candidates=[f for f in mapping['inventory']['functions'] if f['block'] and f['index'] in frequency and f['index'] not in sampled_functions]
function=min(candidates,key=lambda f:(frequency[f['index']],f['id']))
env=dict(os.environ,PYTHONDONTWRITEBYTECODE='1',STAGE3_VERDICT_UPSTREAM=str(source))
def command(argv,label,expected=0):
 with (out/(label+'.log')).open('wb') as log:
  result=subprocess.run([str(x) for x in argv],stdout=log,stderr=log,env=env)
 if result.returncode!=expected:raise RuntimeError(f'{label}: {result.returncode} != {expected}; see log')
 return result.returncode
candidate=out/'source'
command(['git','-C',source,'worktree','add','--detach',candidate,PIN],'checkout')
shutil.copyfile(source/'src/compiler/diagnosticInformationMap.generated.ts',candidate/'src/compiler/diagnosticInformationMap.generated.ts')
file=candidate/function['file'];original=file.read_bytes().decode()
def patch_body(statement):
 # Stock AST offsets are UTF-16 units, including astral code points.
 raw=original.encode('utf-16-le');offset=(function['bodyStart']+1)*2
 file.write_text((raw[:offset]+statement.encode('utf-16-le')+raw[offset:]).decode('utf-16-le'))
bundle=out/'compiler';bundle.mkdir();binary=bundle/'tsc'
binary.write_text('#!/usr/bin/env bash\nexec node '+shlex.quote(str(bundle/'tsc.cjs'))+' "$@"\n');binary.chmod(0o755)
def build(label):
 command(['node',ROOT/'bundle_cli.cjs',source,candidate,bundle],label+'.bundle')
 artifacts=[bundle/'tsc.cjs',*sorted(bundle.glob('lib*.d.ts'))]
 argv=[sys.executable,ROOT/'bind_source.py',binary,candidate]
 for path in artifacts:argv+=['--artifact',path]
 command(argv,label+'.bind')
def verdict(label,coverage,expected):
 started=time.monotonic()
 command([ROOT/'run.sh','--since',PIN,'--tsc',binary,'--coverage-map',coverage,out/label],label,expected)
 return time.monotonic()-started
patch_body('\n/* step34 one-function timing change */\n')
full=inventory(candidate,out/'full-inventory.json')
fast=inventory(candidate,out/'fast-inventory.json',mapping['inventory'])
assert full==fast, 'cached inventory differs from full stock-parser inventory'
build('timing')
wall=verdict('one-function',map_path,0)
timing=json.loads((out/'one-function/differential.json').read_text())
assert timing['changed_functions']==[function['id']]
assert timing['affected']==frequency[function['index']]
patch_body('\nglobalThis.process.stdout.write("step34 wrong function\\n");\n');build('wrong-function')
verdict('mapped-mutant',map_path,1)
wrong=json.loads((out/'mapped-mutant/differential.json').read_text())
assert wrong['baseline_report']['failed']>0 and not wrong['coverage_gaps']
missed=copy.deepcopy(mapping)
for key,row in missed['configurations'].items():
 row.pop('functions_bitmap',None)
 row['functions']=[index for index in covered_indices(mapping['configurations'][key]) if index!=function['index']]
compact_map(missed);missing_map=out/'missing-map.json.gz';save_map(missing_map,missed)
verdict('missing-function-mutant',missing_map,2)
missing=json.loads((out/'missing-function-mutant/differential.json').read_text())
assert missing['coverage_gaps']==[function['id']] and missing['affected']==0 and missing['baseline_report']['failed']==0
stale=copy.deepcopy(mapping);stale['inventory']['files'][function['file']]['sha256']='0'*64
stale_map=out/'stale-map.json.gz';save_map(stale_map,stale)
verdict('stale-source-mutant',stale_map,2)
error=json.loads((out/'stale-source-mutant/differential-error.json').read_text())
assert 'base source hash mismatch' in error['harness_error']
report={'function':function['id'],'callers':frequency[function['index']], 'one_function_wall_seconds':round(wall,3),
 'one_function':timing,'mapped_mutant':wrong,'missing_function_mutant':missing,'stale_source_mutant':error}
(out/'report.json').write_text(json.dumps(report,indent=2)+'\n')
print(json.dumps({'function':function['id'],'callers':frequency[function['index']], 'seconds':timing['seconds'],
 'mapped_mutant':'caught by output comparison','missing_function_mutant':'coverage gap','stale_source_mutant':'source hash mismatch'}))
