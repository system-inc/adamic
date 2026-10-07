#!/usr/bin/env python3
"""Hold rule.json named listeners to production Go and module declarations."""
from pathlib import Path
import json, re, subprocess, sys
own = Path(__file__).resolve().parent
repo = own.parents[3]
out = Path(sys.argv[1]).resolve()
out.mkdir(parents=True, exist_ok=True)
subjects = own / 'evidence/listeners/subjects.json'
with (out / 'go.stdout').open('wb') as stdout, (out / 'go.stderr').open('wb') as stderr:
    subprocess.run(['go', 'run', str(own / 'testdata/listeners.go'), str(subjects), str(repo / 'cohere/internal/lint/rules'), str(out / 'expected.json')], cwd=repo, stdout=stdout, stderr=stderr, check=True)
rows = json.loads((out / 'expected.json').read_text())
for row in rows:
    manifest = own / 'listeners' / Path(row['File']).stem / 'rule.json'
    actual = json.loads(manifest.read_text())
    def agrees(candidate):
        return candidate['name'] == row['Name'] and candidate['kinds'] == row['KindNames']
    assert agrees(actual), str(manifest)
    assert all(type(k) is str for k in actual['kinds']), str(manifest)
    module = (manifest.parent / actual['module']).resolve()
    assert module == (own.parent / row['File']).resolve(), str(manifest)
    match = re.search(r'export const listenerKinds: readonly string\[\] = (\[[^;]+\]);', module.read_text())
    assert match and json.loads(match[1]) == row['KindNames'], str(module)
    mutant = dict(actual, kinds=['Unknown'] + actual['kinds'][1:])
    assert not agrees(mutant), 'named metadata mutant survived: ' + row['Name']
    print('PASS ' + row['Name'] + ': Go/module/rule.json agree; named-kind mutant rejected')
print('PASS nine rule.json declarations and nine comparison-only metadata mutants')
