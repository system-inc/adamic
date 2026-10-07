#!/usr/bin/env python3
"""Hold owned manifests to production Go and the exported numeric declarations."""
from pathlib import Path
import hashlib
import json
import re
import runpy
import sys

OWN = Path(__file__).resolve().parent
REFERENCE = runpy.run_path(str(OWN / 'validation-wave16-listeners/check.py'))
OUT = Path(sys.argv[1]).resolve()
OUT.mkdir(parents=True, exist_ok=True)
results = []
for row in REFERENCE['results']:
    module = row['module']
    go = (REFERENCE['ROOT'] / 'cohere/internal/lint/rules' /
          (REFERENCE['RULES'][module] + '.go')).read_text()
    names = re.findall(r'\bName:\s*"([^"]+)"', go)
    if len(names) != 1:
        raise AssertionError(module + ': unrecognized production name')
    manifest = OWN / 'wave16_listeners' / module / 'rule.json'
    actual = json.loads(manifest.read_text())
    expected = {'name': names[0], 'module': '../../' + module + '.a',
                'kinds': row['numeric_kinds']}

    def compare(candidate):
        if candidate != expected or any(type(k) is not int for k in candidate['kinds']):
            raise AssertionError(f'{module}: Go/module/manifest mismatch: {candidate}')
        if (manifest.parent / candidate['module']).resolve() != OWN / (module + '.a'):
            raise AssertionError(module + ': module outside the owned rule')

    compare(actual)
    mutations = {
        'kind': dict(actual, kinds=actual['kinds'][:-1] + [actual['kinds'][-1] + 1]),
        'name': dict(actual, name=actual['name'] + '-mutant'),
        'module': dict(actual, module='../../no_such_wave16_rule.a'),
    }
    for label, mutant in mutations.items():
        try:
            compare(mutant)
        except AssertionError:
            print(f'PASS {actual["name"]}: {label} JSON mutant caught by independent metadata comparison')
        else:
            raise AssertionError(module + ': ' + label + ' JSON mutant survived')
    results.append({'manifest': str(manifest.relative_to(OWN)), 'expected': expected,
                    'mutants': mutations, 'mutants_caught': list(mutations),
                    'manifest_sha256': hashlib.sha256(manifest.read_bytes()).hexdigest(),
                    'go_source_sha256': row['go_source_sha256'],
                    'adamic_source_sha256': row['adamic_source_sha256']})
(OUT / 'manifests.json').write_text(json.dumps(results, indent=2) + '\n')
print('PASS 15 rule.json manifests and 45 comparison-only metadata mutants')
