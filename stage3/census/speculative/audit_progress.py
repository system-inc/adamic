"""Prove progress instrumentation preserves observations and reports completed AST counts.
Usage: audit_progress.py OLD_BINARY PROGRESS_BINARY OUTPUT
"""
from copy import deepcopy
import json
import os
from pathlib import Path
import subprocess
import sys
old,new,output=(Path(x).resolve() for x in sys.argv[1:4])
output.mkdir(exist_ok=True,parents=True)
source=output/'source';source.mkdir(exist_ok=True)
controls=Path(__file__).parent/'control/dependency'
for name in ('main','target'):(source/(name+'.a')).write_bytes((controls/(name+'.a')).read_bytes())
for label,binary in (('old',old),('progress',new)):
    env=dict(os.environ,LATENT_SPECULATIVE='1',LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1')
    env.pop('LATENT_ONLY_FILE',None)
    if label=='progress':env['LATENT_PROGRESS_FILE']=str(output/'progress.json')
    else:env.pop('LATENT_PROGRESS_FILE',None)
    with (output/(label+'.log')).open('w') as log:
        subprocess.run([str(binary),str(source),str(output/(label+'.jsonl'))],env=env,stdout=log,stderr=subprocess.STDOUT,timeout=60,check=True)
assert (output/'old.jsonl').read_bytes()==(output/'progress.jsonl').read_bytes(),'progress changed census observations'
rows=[json.loads(x) for x in (output/'progress.jsonl').read_text().splitlines()]
progress=json.loads((output/'progress.json').read_text())
record=next(r for r in rows[1:] if r['file']==progress['file'])
def check(p):
    expected=record['speculative_coverage']
    assert p['phase']=='completed'
    assert p['walker_unique_target_nodes']==p['examined_unique_target_nodes']==p['total_target_ast_nodes']==expected['total_nodes']
    assert p['source_bytes']==expected['source_bytes']
    assert 0<=p['lowerer_unique_target_nodes']<=p['total_target_ast_nodes']
    assert p['statements_completed']==p['total_statements']
check(progress)
for name,field in (('shifted-node-count','examined_unique_target_nodes'),('shifted-source-bytes','source_bytes')):
    mutant=deepcopy(progress);mutant[field]+=1
    try:check(mutant)
    except AssertionError:print(name+' mutant caught by completed AST recount')
    else:raise AssertionError(name+' survived')
print('PASS: progress-on output byte-identical to prior binary; completed progress counters independently match coverage traversal')
