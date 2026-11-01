#!/usr/bin/env python3
"""Source, census, idempotence, and one-public-overload undo proof."""
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
before,after,census,output=map(Path,sys.argv[1:])
output.mkdir(parents=True,exist_ok=False)
here=Path(__file__).resolve().parent
adapter=here.parents[1]/'adapt/95-checker-axolotl-e'
env=dict(os.environ,NODE_PATH=str(here.parents[1]/'api/node_modules'))
def run(label,command,expected=0):
    with (output/(label+'.log')).open('w') as log:
        result=subprocess.run(list(map(str,command)),stdout=log,stderr=subprocess.STDOUT,env=env)
    assert result.returncode==expected,(label,result.returncode,expected)
def observe(label,tree):
    path=output/(label+'.jsonl')
    run(label,[census,tree/'src/tsc/tsc.ts',path])
    return path
run('source',['node',adapter/'verify.cjs',before,after])
run('idempotence',['node',adapter/'adapt.cjs',after])
assert json.loads((output/'idempotence.log').read_text())['declarations']==0
run('check',['node',adapter/'adapt.cjs','--check',after])
a,b=observe('before',before),observe('after',after)
mutant=output/'mutant'
shutil.copytree(after,mutant,ignore=shutil.ignore_patterns('.git','node_modules','built'))
row=json.loads((adapter/'sites.json').read_text())[0]
file=mutant/'src/compiler/types.ts'
old=row['ownerBefore']
new=old.replace('typeArguments?: NodeArray<TypeNode>;','typeArguments?: NodeArray<TypeNode> | undefined;')
data=file.read_bytes()
assert data.count(new.encode())==1
file.write_bytes(data.replace(new.encode(),old.encode()))
run('mutant-verifier',['node',adapter/'verify.cjs',before,mutant],1)
run('mutant-check',['node',adapter/'adapt.cjs','--check',mutant],1)
c=observe('mutant',mutant)
run('census',['python3',here/'compare-census.py',a,b,c])
print(json.dumps(dict(status='pass',undo='public overload only',undoVerifierExit=1,undoCheckExit=1),indent=2))
