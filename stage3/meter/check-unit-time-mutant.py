#!/usr/bin/env python3
"""Drop the reported time block and require its report integration test to fail."""
from pathlib import Path
import subprocess
import sys
import tempfile

meter = Path(__file__).resolve().parent
source = (meter / 'report.py').read_text()
old = "(run / 'report.json').write_text(json.dumps(result, indent=2) + '\\n')\n    lines = []"
new = "(run / 'report.json').write_text(json.dumps(dict(result, trees={label: {key: value for key, value in tree.items() if key != 'latent_time'} for label, tree in trees.items()}), indent=2) + '\\n')\n    lines = []"
assert source.count(old) == 1
with tempfile.TemporaryDirectory(prefix='meter-time-mutant-') as temporary:
    scratch = Path(temporary)
    (scratch / 'report.py').write_text(source.replace(old, new))
    (scratch / 'owners.json').write_bytes((meter / 'owners.json').read_bytes())
    program = ('import sys,unittest; sys.path.insert(0,' + repr(str(scratch)) + '); '
               'sys.path.insert(1,' + repr(str(meter)) + '); import report_test; '
               'unittest.main(module=report_test, defaultTest="ReportTests.test_meter_unit_times_block_sorts_files_and_shortens_units")')
    log_path = scratch / 'mutant-test.txt'
    with log_path.open('w') as log:
        result = subprocess.run([sys.executable, '-E', '-S', '-B', '-c', program],
                                stdout=log, stderr=log)
    output = log_path.read_text()
    print(output, end='')
    if result.returncode == 0 or "KeyError: 'latent_time'" not in output:
        raise AssertionError('latent_time mutant survived or failed for another reason')
print("Caught drop-latent_time: KeyError: 'latent_time'")
