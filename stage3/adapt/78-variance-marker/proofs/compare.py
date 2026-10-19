#!/usr/bin/env python3
"""Check the successful whole lane against the current-main control artifacts."""
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
assert report['status'] == 'pass', report
assert oracle['counts'] == dict(passing=106366, failing=1, pending=0), oracle['counts']
assert {name: row['exit'] for name, row in oracle['phases'].items()} == dict(install=0, build=0, tests=1)
assert (results / 'execution.json').exists()
assert json.loads((results / 'execution.json').read_text())['apply_exit'] == 0
pin = '141fec7702d4aed21429724b2afd7efcd980f277'
assert subprocess.check_output(['git', '-C', str(repo), 'merge-base', 'HEAD', 'origin/main'], text=True, timeout=10).strip() == pin
assert subprocess.check_output(['git', '-C', str(repo), 'rev-parse', 'typescript/stage3-adapt78-void-truths'], text=True, timeout=10).strip() == '12432d746daee34aa8ebd2fe88f104fc744882ec'
manifest = 'stage3/lane/sanctioned-api.json'
assert (repo / manifest).read_bytes() == subprocess.check_output(['git', '-C', str(repo), 'show', f'{pin}:{manifest}'], timeout=10)

def digest(data):
    return hashlib.sha256(data).hexdigest()

artifacts = []
for name in sorted(p.name for p in (control / 'built/local').glob('*.js')) + ['typescript.d.ts', 'tsserverlibrary.d.ts']:
    before = (control / 'built/local' / name).read_bytes()
    after = (tree / 'built/local' / name).read_bytes()
    assert before == after, name
    artifacts.append(dict(file=name, bytes=len(after), sha256=digest(after)))
relative = 'tests/baselines/reference/api/typescript.d.ts'
reference = (control / relative).read_bytes()
assert reference == (tree / relative).read_bytes(), 'reference changed'
main_api = (control / 'built/local/typescript.d.ts').read_bytes()
main_diff = ''.join(difflib.unified_diff(reference.decode().splitlines(True), main_api.decode().splitlines(True),
                                      fromfile='reference/api/typescript.d.ts', tofile='local/api/typescript.d.ts')).encode()
actual_diff = (results / 'oracle/baseline.diff').read_bytes()
assert actual_diff == main_diff, 'API diff changed from main'
# Prove the byte-identity check rejects extra baseline text.
try:
    assert actual_diff + b'\nmarker baseline mutant\n' == main_diff
except AssertionError:
    baseline_mutant_caught = True
else:
    raise AssertionError('baseline mutant survived')
source_names = sorted(p.relative_to(control / 'src') for p in (control / 'src').rglob('*') if p.is_file())
assert source_names == sorted(p.relative_to(tree / 'src') for p in (tree / 'src').rglob('*') if p.is_file())
changed_sources = [str(p) for p in source_names if (control / 'src' / p).read_bytes() != (tree / 'src' / p).read_bytes()]
assert changed_sources == ['compiler/checker.ts'], changed_sources
old_checker = (control / 'src/compiler/checker.ts').read_bytes()
old_annotation = b'((onlyUnreliable: boolean) => void)'
assert old_checker.count(old_annotation) == 1
expected_checker = old_checker.replace(old_annotation, b'((onlyUnreliable: boolean) => boolean)')
assert (tree / 'src/compiler/checker.ts').read_bytes() == expected_checker, 'numbered apply differs from the proved rewrite'
result = dict(status='pass', main=pin, blockedBranchUnchanged=True, manifestUnchanged=True, counts=oracle['counts'],
              laneResultLine=(results / 'verdict.txt').read_text().strip(), artifacts=artifacts,
              baselineDiffSha256=digest(actual_diff), mainBaselineDiffSha256=digest(main_diff), baselineDiffBytes=len(actual_diff),
              baselineMutantCaught=baseline_mutant_caught, numberedApplyMatchesProof=True, changedSourceFiles=changed_sources)
output.write_text(json.dumps(result, indent=2) + '\n')
(output.parent / 'baseline.diff').write_bytes(actual_diff)
print(json.dumps(result))
