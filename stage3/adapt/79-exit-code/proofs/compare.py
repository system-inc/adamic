#!/usr/bin/env python3
"""Check actual full-lane artifacts against main plus the sole ruled API pair."""
import difflib
import hashlib
import json
from pathlib import Path
import subprocess
import sys

repo, control, results, output = map(Path, sys.argv[1:])
tree = results / 'adapted-tree'
report = json.loads((results / 'report.json').read_text())
oracle = json.loads((results / 'oracle/report.json').read_text())
assert oracle['counts'] == dict(passing=106366, failing=1, pending=0), oracle['counts']
assert report['status'] == 'pass', report['errors']
assert {name: row['exit'] for name, row in oracle['phases'].items()} == dict(install=0, build=0, tests=1)
pin = '23ff023b9ed986834d20404d43e2c363667edf9b'
assert subprocess.check_output(['git', '-C', str(repo), 'merge-base', 'HEAD', pin], text=True, timeout=10).strip() == pin
# Main's stage3 is unchanged from the control's source revision.
subprocess.run(['git', '-C', str(repo), 'diff', '--exit-code', '141fec77', pin, '--', 'stage3'], check=True, timeout=10)
old = b'exit(exitCode?: number): void;'
new = b'exit(exitCode?: number): number | undefined;'
reference = (control / 'tests/baselines/reference/api/typescript.d.ts').read_bytes()
assert reference == (tree / 'tests/baselines/reference/api/typescript.d.ts').read_bytes()
main_api = (control / 'built/local/typescript.d.ts').read_bytes()
assert main_api.count(old) == 1
expected_api = main_api.replace(old, new)
assert (tree / 'built/local/typescript.d.ts').read_bytes() == expected_api
assert (tree / 'tests/baselines/local/api/typescript.d.ts').read_bytes() == expected_api
actual_diff = (results / 'oracle/baseline.diff').read_bytes()
expected_diff = ''.join(difflib.unified_diff(reference.decode().splitlines(True), expected_api.decode().splitlines(True), fromfile='reference/api/typescript.d.ts', tofile='local/api/typescript.d.ts')).encode()
assert actual_diff == expected_diff, 'API diff is not exactly main plus exit pair'
manifest_path = 'stage3/lane/sanctioned-api.json'
previous_manifest = json.loads(subprocess.check_output(['git', '-C', str(repo), 'show', f'{pin}:{manifest_path}'], timeout=10))
manifest = json.loads((repo / manifest_path).read_text())
previous = {row['declaration']: row for row in previous_manifest['normalized_changes']}
current = {row['declaration']: row for row in manifest['normalized_changes']}
added = current.keys() - previous.keys()
assert added == {'member:["ts","System","exit"]#0'}, added
assert all(current[key] == row for key, row in previous.items())
# Actual byte predicate is proved with an extra-line mutant.
baseline_mutant_caught = actual_diff + b'\nextra API mutant\n' != expected_diff
assert baseline_mutant_caught
source_names = sorted(p.relative_to(control / 'src') for p in (control / 'src').rglob('*') if p.is_file())
assert source_names == sorted(p.relative_to(tree / 'src') for p in (tree / 'src').rglob('*') if p.is_file())
changed = [str(p) for p in source_names if (control / 'src' / p).read_bytes() != (tree / 'src' / p).read_bytes()]
wanted = ['compiler/executeCommandLine.ts', 'compiler/sys.ts', 'harness/fakesHosts.ts', 'harness/harnessLanguageService.ts', 'testRunner/unittests/helpers/tscWatch.ts', 'testRunner/unittests/helpers/virtualFileSystemWithWatch.ts', 'testRunner/unittests/tscWatch/incremental.ts', 'testRunner/unittests/tsserver/session.ts']
assert changed == wanted, changed
source_patch = ''.join(''.join(difflib.unified_diff((control / 'src' / f).read_text().splitlines(True), (tree / 'src' / f).read_text().splitlines(True), fromfile='a/src/'+f, tofile='b/src/'+f)) for f in changed)
assert source_patch == (Path(__file__).parent / 'exit.patch').read_text(), 'numbered apply differs from the proved edits'
compiler_artifacts=[]
for p in sorted((control / 'built/local').glob('*.js')):
    if p.name == 'run.js':
        continue  # The harness bundle contains the ruled guards and dedicated stubs.
    a=p.read_bytes(); b=(tree / 'built/local' / p.name).read_bytes()
    assert a == b, p.name
    compiler_artifacts.append(dict(file=p.name, bytes=len(b), sha256=hashlib.sha256(b).hexdigest()))
assert (control / 'built/local/run.js').read_bytes() != (tree / 'built/local/run.js').read_bytes()
result = dict(status='pass', main=pin, counts=oracle['counts'], laneResultLine=(results / 'verdict.txt').read_text().strip(), apiAddition=dict(before=old.decode(), after=new.decode()), baselineDiffSha256=hashlib.sha256(actual_diff).hexdigest(), baselineDiffBytes=len(actual_diff), baselineMutantCaught=baseline_mutant_caught, changedSourceFiles=changed, compilerArtifacts=compiler_artifacts, harnessBundleChanged=True)
output.write_text(json.dumps(result, indent=2)+'\n')
(output.parent / 'baseline.diff').write_bytes(actual_diff)
print(json.dumps(result))
