"""Checker-clean failure nesting distinguishes stubbing from diagnostic admission."""
import json
import os
from pathlib import Path
import subprocess
import sys

binary, output = map(lambda x:Path(x).resolve(),sys.argv[1:3])
source=output/'source'
source.mkdir(parents=True,exist_ok=True)
(source/'probe.a').write_bytes((Path(__file__).parent/'control/signature-nested.a').read_bytes())
def run(name,**extra):
    env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1')
    for key in ('LATENT_SPECULATIVE','LATENT_MUTANT_NO_STUBS','LATENT_MUTANT_DEPTH_ZERO'):
        env.pop(key,None)
    env.update(extra)
    destination=output/(name+'.jsonl')
    with (output/(name+'.log')).open('w') as log:
        subprocess.run([str(binary),str(source),str(destination)],env=env,stdout=log,stderr=subprocess.STDOUT,check=True)
    return [json.loads(x) for x in destination.read_text().splitlines()]
def check(rows):
    assert not rows[0]['checker_rejected'] and not rows[0]['diagnostics']
    findings=[f for r in rows[1:] for f in r['findings']]
    for kind,depth in [('KindFunctionDeclaration',0),('KindCallExpression',1),('KindArrowFunction',2),('KindIdentifier',3)]:
        actual=[f['depth'] for f in findings if f.get('site_kind')==kind]
        assert [d for d in actual if d==depth]==[depth],(kind,actual)
    assert len(findings)==5 and all(f['kind']=='NotYet' for f in findings)
    assert all(r['speculative_coverage']['unvisited_nodes']==0 for r in rows[1:])
    assert [f["depth"] for f in findings if f["site_kind"]=="KindIdentifier" and f["reason"]=="a value of type T"]==[1]
    return len(findings)
speculative=run('speculative',LATENT_SPECULATIVE='1')
check(speculative)
full=run('full')
mutant=run('no-stubs-mutant',LATENT_SPECULATIVE='1',LATENT_MUTANT_NO_STUBS='1')
assert mutant==full
assert sum(len(r['findings']) for r in full[1:])==2 # One NotYet plus its historical Boundary ledger row.
assert sum(f['kind']=='NotYet' for r in full[1:] for f in r['findings'])==1
for name,rows in [('no-stubs',mutant),('depth-zero',run('depth-zero-mutant',LATENT_SPECULATIVE='1',LATENT_MUTANT_DEPTH_ZERO='1'))]:
    try:
        check(rows)
    except AssertionError:
        print(name+' mutant caught by checker-clean exact nesting control')
    else:
        raise AssertionError(name+' survived')
print('PASS: checker-clean root signature 0, failed call 1, arrow child 2, captured value 3; five NotYet sites including hidden parameter versus one full; exact no-stubs JSON restoration')
