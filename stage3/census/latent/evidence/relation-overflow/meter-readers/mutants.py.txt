"""Run the two meter reader mutants in isolated copies; retain child exit codes."""
import json
from pathlib import Path
import subprocess
import sys

repository = Path(__file__).resolve().parents[6]
meter = repository / 'stage3/meter'
output = Path(sys.argv[1]).resolve()
output.mkdir(parents=True, exist_ok=True)
original = (meter / 'report.py').read_text()
mutants = [
    ('trailer-as-file', 'records = records[:-1]', 'records = records[:]'),
    ('drop-trailer-findings', 'for record in [*records[1:], *([trailer] if trailer else [])]:',
     'for record in records[1:]:'),
]
results = {}
for name, old, new in mutants:
    assert original.count(old) == 1, name
    destination = output / name
    destination.mkdir(exist_ok=True)
    (destination / 'report.py').write_text(original.replace(old, new))
    for file in ('report_test.py', 'owners.json'):
        (destination / file).write_bytes((meter / file).read_bytes())
    code = ('import sys, unittest; sys.path.append(' + repr(str(meter)) + '); '
            "unittest.main(module=None, argv=['test', 'report_test.RelationTrailerReportTests'])")
    with (destination / 'test.log.txt').open('w') as log:
        result = subprocess.run(['python3', '-E', '-S', '-B', '-c', code],
                                cwd=destination, stdout=log, stderr=subprocess.STDOUT)
    results[name] = result.returncode
    print(f'{name}_exit={result.returncode}', flush=True)
(output / 'results.json').write_text(json.dumps(results, indent=2) + '\n')
assert all(code == 1 for code in results.values()), results
