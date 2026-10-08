#!/usr/bin/env python3
"""Check production loader assertions with Go source overlays, never runtime guards."""
import json
from pathlib import Path
import subprocess
import tempfile

repository = Path(__file__).resolve().parents[2]
mutants = {
    'replace-project-options': ('project_loader.go', 'return options, project, nil', 'return compilerOptions(), project, nil'),
    'erase-recorded-sites': ('load.go', 'sites = append(sites, site)', '_ = site'),
    'waive-optional-sites': ('load.go', 'if loadedFiles[site.File] {', 'if loadedFiles[site.File] && site.Options[0] != "exactOptionalPropertyTypes" {'),
    'erase-mixed-file-refusal': ('load.go', 'if _, isAdamic := fs.adamicFile(file.FileName()); isAdamic {', 'if _, isAdamic := fs.adamicFile(file.FileName()); isAdamic && false {'),
}
logs = Path(tempfile.mkdtemp(prefix='stricter-production-mutants-', dir='/tmp'))
for name, (file, before, after) in mutants.items():
    source = repository / 'internal/load' / file
    original = source.read_text()
    if original.count(before) != 1:
        raise SystemExit(f'{name}: expected exactly one mutation location')
    replacement = logs / (name + '.go')
    replacement.write_text(original.replace(before, after))
    overlay = logs / (name + '.json')
    overlay.write_text(json.dumps({'Replace': {str(source): str(replacement)}}))
    log = logs / (name + '.log')
    with log.open('w') as output:
        result = subprocess.run(['go', 'test', '-overlay', str(overlay), './internal/load', '-run', 'TestProduction', '-count=1', '-timeout', '5m'], cwd=repository, stdout=output, stderr=subprocess.STDOUT)
    text = log.read_text()
    if result.returncode == 0 or '--- FAIL: TestProduction' not in text or '[build failed]' in text:
        raise SystemExit(f'{name}: survived or failed outside assertions; see {log}')
    print(f'{name}: caught by production assertions, exit {result.returncode}; {log}', flush=True)
print('4 production loader mutants caught; runtime conversion is not tested by this runner', flush=True)
