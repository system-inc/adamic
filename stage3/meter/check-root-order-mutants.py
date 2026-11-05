#!/usr/bin/env python3
"""Require one red regression for each direction of the root-order ruling."""
from pathlib import Path
import subprocess
import sys
import tempfile

meter = Path(__file__).resolve().parent
source = (meter / 'report.py').read_text()
mutants = [
    ('ignore-entry-readiness', 'in attempted - readiness', 'in attempted',
     'test_entry_order_readiness_still_counts_against_root_file'),
    ('count-root-order-artifact-as-failure', "artifacts = [row for row in checks if row['node']['throws']]",
     'artifacts = []', 'test_root_only_readiness_with_node_throw_is_visible_artifact'),
]
for name, before, after, test in mutants:
    assert before in source, name
    with tempfile.TemporaryDirectory(prefix='meter-root-order-mutant-') as temporary:
        scratch = Path(temporary)
        (scratch / 'report.py').write_text(source.replace(before, after))
        (scratch / 'owners.json').write_bytes((meter / 'owners.json').read_bytes())
        program = ('import sys,unittest; sys.path.insert(0,' + repr(str(scratch)) + '); '
                   'sys.path.insert(1,' + repr(str(meter)) + '); import report_test; '
                   'unittest.main(module=report_test,defaultTest=' + repr('RootOrderTests.' + test) + ')')
        path = scratch / 'test.log'
        with path.open('w') as output:
            result = subprocess.run([sys.executable, '-E', '-S', '-B', '-c', program],
                                    stdout=output, stderr=output)
        text = path.read_text()
        print(name + ': ' + test)
        print(text, end='')
        if result.returncode == 0 or 'FAIL: ' + test not in text or 'AssertionError:' not in text:
            raise AssertionError(name + ' survived or failed for another reason')
print('Both root-order mutants went red on their named tests.')
