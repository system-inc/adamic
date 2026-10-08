"""Remove actual IR checks in supported negative twins and require Node behavior."""
import argparse
import json
import os
from pathlib import Path
import tempfile
import sys

sys.dont_write_bytecode = True

from verify import ROOT, REPO, run, verify_observation, checked_write_accepts

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--tool', type=Path, required=True)
arguments = parser.parse_args()
manifest = json.loads((ROOT / 'manifest.json').read_text())
observations = json.loads((ROOT / 'observations.json').read_text())
by_name = {row['file']: row for row in observations['fixtures']}
results = []
with tempfile.TemporaryDirectory(prefix='checked-write-mutants-') as directory:
    scratch = Path(directory)
    for family in manifest['families']:
        name = family['files']['out']
        if not by_name[name]['checked_write_pass']:
            results.append(dict(family=family['family'], file=name, status='blocked',
                                reason='unmutated checked-write runtime contract has not passed'))
            continue
        source = scratch / name.replace('.a', '.ts')
        source.write_bytes((ROOT / name).read_bytes())
        executable = scratch / name.replace('.a', '')
        build = run([str(arguments.tool.resolve()), str(source), str(executable)])
        assert build['exit'] == 0 and 'actual IR store checks' in build['stdout'], build
        environment = dict(os.environ, ASAN_OPTIONS='detect_leaks=1', UBSAN_OPTIONS='halt_on_error=1')
        native = run([str(executable)], environment=environment)
        javascript = run(['node', '--disable-warning=ExperimentalWarning', str(REPO / 'oracle/node.mjs'), str(executable)+'.mjs'])
        for backend, actual in [('sanitized native', native), ('JavaScript', javascript)]:
            # It must survive type checking, clang, sanitizers and leaks, then
            # fail only the exit-70 write-check comparison by storing silently.
            verify_observation(actual, family['expected']['out']['node'], name + ' ' + backend + ' mutant')
            assert not checked_write_accepts(actual, family['expected']['out']['adamic_ts']), name
        results.append(dict(family=family['family'], file=name, status='caught',
                            caught_by='exit-70 checked-write expectation in both backends',
                            build=build, native=native, javascript=javascript))
        print('CAUGHT actual check-removal mutant:', family['family'], 'by native and JS exit-70 expectations')
report = dict(compiler_revision=observations['compiler_revision'], results=results,
              caught=sum(row['status']=='caught' for row in results),
              blocked=sum(row['status']=='blocked' for row in results))
assert report['caught'] > 0, 'no runtime check-removal mutant was tested'
(ROOT / 'runtime-mutants.json').write_text(json.dumps(report, indent=2)+'\n')
print('TOTALS', report['caught'], 'actual IR mutants caught;', report['blocked'], 'blocked families')
