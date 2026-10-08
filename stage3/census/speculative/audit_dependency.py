"""Dependency recovery spans retain global depth and honest per-source coverage."""
import json
import os
from pathlib import Path
import subprocess
import sys
binary,output=map(lambda x:Path(x).resolve(),sys.argv[1:3])
source=output/'source';source.mkdir(parents=True,exist_ok=True)
for name in ('main','target'):
    (source/(name+'.a')).write_bytes((Path(__file__).parent/('control/dependency/'+name+'.a')).read_bytes())
def run(name,**extra):
    env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1')
    for key in ('LATENT_SPECULATIVE','LATENT_MUTANT_DEPTH_ZERO'):
        env.pop(key,None)
    env.update(extra)
    destination=output/(name+'.jsonl')
    with (output/(name+'.log')).open('w') as log:
        subprocess.run([str(binary),str(source),str(destination)],env=env,stdout=log,stderr=subprocess.STDOUT,check=True)
    return [json.loads(line) for line in destination.read_text().splitlines()]
def check(rows):
    main=next(r for r in rows[1:] if Path(r['file']).name=='main.a')
    coverage=main['speculative_coverage']
    assert coverage['unvisited_nodes']==0 and coverage['visited_nodes']>coverage['total_nodes']
    foreign=[f for f in main['findings'] if '/target.a:' in f.get('site_where','')]
    assert {(f['reason'],f['depth']) for f in foreign}=={('a WithStatement',1),('a DebuggerStatement',2)}
    assert all(r['speculative_coverage']['unvisited_nodes']==0 for r in rows[1:])
    return len(foreign)
assert check(run('speculative',LATENT_SPECULATIVE='1'))==2
try:
    check(run('depth-zero-mutant',LATENT_SPECULATIVE='1',LATENT_MUTANT_DEPTH_ZERO='1'))
except AssertionError:
    print('depth-zero mutant caught by cross-file failure ancestry')
else:
    raise AssertionError('depth-zero mutant survived')
print('PASS: dependency sites retain depths 1 and 2 after later signature failure; all source nodes visited; visited set honestly includes dependency nodes')
