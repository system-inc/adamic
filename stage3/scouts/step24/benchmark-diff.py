"""Alternate comparator and diff on the same actual successful end mutant."""
from pathlib import Path
import hashlib
import json
import subprocess
import time

here=Path(__file__).resolve().parent
root=Path('/tmp/step24-parser-final')
left,right=root/'node.dump',root/'mutant.dump'
commands={'dumpdiff':['/tmp/step24-dumpdiff',str(left),str(right)],'diff':['diff','-u',str(left),str(right)]}
observations={k:[] for k in commands}
for round in range(5):
    for name in (list(commands) if round%2==0 else list(reversed(commands))):
        start=time.perf_counter()
        with (here/'evidence'/f'{name}-{round}.log').open('wb') as stream:
            result=subprocess.run(commands[name],stdout=stream,stderr=subprocess.STDOUT)
        elapsed=time.perf_counter()-start
        assert result.returncode==1
        observations[name].append(elapsed)
with (here/'evidence/comparator-equal.log').open('wb') as stream:
    result=subprocess.run(['/tmp/step24-dumpdiff',str(left),str(left)],stdout=stream,stderr=subprocess.STDOUT)
assert result.returncode==0
report={'bytes':left.stat().st_size,'sha256':hashlib.sha256(left.read_bytes()).hexdigest(),
        'input':'actual successful Identifier.end +1 mutant from parser run.sh','equal_exit':0,'different_exit':1,
        'seconds':observations,'best':{k:min(v) for k,v in observations.items()},
        'limit':'early mismatch at line 166, warm local file cache; no general speedup or worst-case claim'}
assert '_namespaces/ts.ts/preorder/150' in (here/'evidence/dumpdiff-0.log').read_text()
(here/'evidence/comparator-benchmark.json').write_text(json.dumps(report,indent=2)+'\n')
print(json.dumps(report,indent=2))
