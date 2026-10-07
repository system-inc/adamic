#!/usr/bin/env python3
"""Keep raw diagnostic ordering and compare feature profiles as multisets."""
from collections import Counter
import gzip
import hashlib
import json
from pathlib import Path
import re
import shutil
import subprocess
import sys

scratch = Path(sys.argv[1]).resolve()
output = Path(__file__).resolve().parent / 'evidence'
closure = json.loads((output / 'closure.json').read_text())
values = json.loads((output / 'value-closure.json').read_text())
profiles = {}
baseline = None
for name in ['main', 'taste', 'flags', 'namespaces', 'nested', 'combined']:
    record = json.loads((scratch / (name + '-integration.json')).read_text())
    record['head'] = subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=scratch / name, text=True).strip()
    diff = subprocess.check_output(['git', 'diff', '--', 'internal/lower', 'internal/load'], cwd=scratch / name)
    (output / (name + '-scratch-resolution.patch')).write_bytes(diff)
    record['scratchResolutionPatchSha256'] = hashlib.sha256(diff).hexdigest()
    record['cohereGitlink'] = subprocess.check_output(['git', 'ls-tree', 'HEAD', 'cohere'], cwd=scratch / name, text=True).strip()
    (output / (name + '-integration.json')).write_text(json.dumps(record, indent=2) + '\n')
    shutil.copyfile(scratch / (name + '.log'), output / (name + '-integration.log'))
    loader = (scratch / (name + '-load.go.txt')).read_text()
    assert 'NoImplicitReturns:          core.TSTrue,' in loader
    assert 'NoFallthroughCasesInSwitch: core.TSTrue,' in loader
    shutil.copyfile(scratch / (name + '-load.go.txt'), output / (name + '-load.go.txt'))
    profile = {'buildExit': record['buildExit']}
    if record['buildExit'] == 0:
        parser = json.loads((scratch / (name + '-parser.ts.json')).read_text())
        scanner = json.loads((scratch / (name + '-scanner.ts.json')).read_text())
        assert record['entryResults']['parser.ts']['exit'] == record['entryResults']['scanner.ts']['exit'] == 0
        profile['outcome'] = parser['outcome']
        diagnostics = [s.replace('/tmp/parser-adapted10/', '') for s in parser.get('diagnostics', [])]
        scanner_diagnostics = [s.replace('/tmp/parser-adapted10/', '') for s in scanner.get('diagnostics', [])]
        profile['sameScannerDiagnosticsInOrder'] = diagnostics == scanner_diagnostics
        profile['scannerDiagnostics'] = len(scanner_diagnostics)
        profile['scannerEntryTextDifferences'] = {'parserOnly': list((Counter(diagnostics) - Counter(scanner_diagnostics)).elements()), 'scannerOnly': list((Counter(scanner_diagnostics) - Counter(diagnostics)).elements())}
        profile['sameScannerSitesAndCodes'] = [s.split(': error TS', 1)[0] + ':TS' + re.search(r'error TS(\d+):', s).group(1) for s in diagnostics] == [s.split(': error TS', 1)[0] + ':TS' + re.search(r'error TS(\d+):', s).group(1) for s in scanner_diagnostics]
        profile['diagnostics'] = len(diagnostics)
        profile['counts'] = dict(Counter(re.search(r'error TS(\d+):', s).group(1) for s in diagnostics))
        if baseline is None:
            baseline = diagnostics
        old, new = Counter(baseline), Counter(diagnostics)
        profile['removedFromMain'] = list((old - new).elements())
        profile['addedToMain'] = list((new - old).elements())
        if diagnostics:
            profile['first'] = diagnostics[0]
        own = [s for s in diagnostics if s.startswith('src/compiler/parser.ts:')]
        profile['parserFileCounts'] = dict(Counter(re.search(r'error TS(\d+):', s).group(1) for s in own))
        (output / (name + '-parser-file.json')).write_text(json.dumps(own, indent=2) + '\n')
        data = (json.dumps({'entry': 'src/compiler/parser.ts', 'outcome': parser['outcome'], 'diagnostics': diagnostics}, indent=2) + '\n').encode()
        (output / (name + '-ordered-diagnostics.json.gz')).write_bytes(gzip.compress(data, mtime=0))
        profile['orderedDiagnosticsSha256'] = hashlib.sha256(data).hexdigest()
        def belongs(diagnostic, files):
            return any(diagnostic.startswith(file + ':') for file in files)
        literal_only = [s for s in diagnostics if belongs(s, closure['parserBeyondScanner'])]
        value_only = [s for s in diagnostics if belongs(s, values['parserBeyondScanner'])]
        profile['fileClosureExclusiveDiagnostics'] = len(literal_only)
        profile['valueClosureExclusiveFileDiagnostics'] = len(value_only)
        (output / (name + '-value-exclusive-diagnostics.json.gz')).write_bytes(gzip.compress((json.dumps(value_only, indent=2)+'\n').encode(), mtime=0))
    profiles[name] = profile
# These are gate observations, not a guessed order of latent lowering refusals.
(output / 'feature-summary.json').write_text(json.dumps(profiles, indent=2) + '\n')
for name, profile in profiles.items():
    print(name, {key: value for key, value in profile.items() if key not in ['removedFromMain', 'addedToMain']}, 'removed', len(profile.get('removedFromMain', [])), 'added', len(profile.get('addedToMain', [])))
