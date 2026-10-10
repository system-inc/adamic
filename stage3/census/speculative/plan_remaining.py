"""Plan the published incomplete TypeScript files without executing their walks.
Usage: plan_remaining.py ROOT RESULT_JSON OUTPUT N
"""
import json
from pathlib import Path
import subprocess
import sys
root, result_path, output = (Path(p).resolve() for p in sys.argv[1:4])
count = int(sys.argv[4])
result = json.loads(result_path.read_text())
output.mkdir(parents=True,exist_ok=True)
files = [f for f in result['unexamined'] if Path(f['file']).suffix in ('.ts','.a')]
assert len(files) == 14, 'remaining-file inventory changed'
workload = []
for file in files:
    source = root / file['file']
    assert source.stat().st_size == file['bytes']
    plan_path = output / (file['file'].replace('/','__') + '.plan.json')
    with plan_path.with_suffix('.log').open('w') as log:
        subprocess.run(['node',str(Path(__file__).with_name('plan_pieces.cjs')),str(source),str(count),str(plan_path)],
                       stdout=log,stderr=subprocess.STDOUT,timeout=30,check=True)
    plan = json.loads(plan_path.read_text())
    workload.append(dict(file=file['file'], source_bytes=file['bytes'], plan=plan_path.name,
                         pieces=len(plan['pieces']), atoms=len(plan['atoms']),
                         atom_bytes=sum(a['bytes'] for a in plan['atoms']), largest_piece_bytes=max(p['bytes'] for p in plan['pieces'])))
assert sum(f['source_bytes'] for f in workload) + result['examined_bytes'] == result['typescript_source_bytes']
(output / 'WORKLOAD.json').write_text(json.dumps(workload,indent=2)+'\n')
print(json.dumps(workload,indent=2))
