#!/usr/bin/env python3
"""Prove option attribution fails when a responsible flag or site guard is erased.

Go overlays keep the working tree unchanged, including during concurrent setup.
These are checker-analysis mutants, not runtime-check mutants.
"""
import json
from pathlib import Path
import subprocess
import tempfile

repository = Path(__file__).resolve().parents[2]
source = repository / 'internal/load/project_options.go'
original = source.read_text()
mutants = {
    'erase-index-option': ('strict.NoUncheckedIndexedAccess = core.TSTrue', 'strict.NoUncheckedIndexedAccess = core.TSFalse'),
    'erase-optional-option': ('strict.ExactOptionalPropertyTypes = core.TSTrue', 'strict.ExactOptionalPropertyTypes = core.TSFalse'),
    'erase-catch-option': ('strict.UseUnknownInCatchVariables = core.TSTrue', 'strict.UseUnknownInCatchVariables = core.TSFalse'),
    'erase-ordinary-membership': ('ordinaryKeys := diagnosticKeys(ordinary)', 'ordinaryKeys := map[optionDiagnosticKey]bool{}'),
    'overwrite-project-lib': ('base := config.CompilerOptions().Clone()', 'base := config.CompilerOptions().Clone(); base.Lib = []string{"lib.es2024.d.ts"}'),
    'erase-site-position': ('position: diagnostic.Pos()', 'position: 0'),
}
logs = Path(tempfile.mkdtemp(prefix='stricter-options-mutants-', dir='/tmp'))
for name, (before, after) in mutants.items():
    if original.count(before) != 1:
        raise SystemExit(f'{name}: mutation must identify exactly one source location')
    replacement = logs / (name + '.go')
    replacement.write_text(original.replace(before, after))
    overlay = logs / (name + '.json')
    overlay.write_text(json.dumps({'Replace': {str(source): str(replacement)}}))
    log = logs / (name + '.log')
    with log.open('w') as output:
        result = subprocess.run(['go', 'test', '-overlay', str(overlay), './internal/load', '-run', 'TestProjectOption', '-count=1', '-timeout', '5m'], cwd=repository, stdout=output, stderr=subprocess.STDOUT)
    text = log.read_text()
    if result.returncode == 0 or '--- FAIL: TestProjectOption' not in text or '[build failed]' in text:
        raise SystemExit(f'{name}: survived or failed outside the intended assertions; see {log}')
    print(f'{name}: caught by project-option assertions, exit {result.returncode}; {log}', flush=True)
print(f'{len(mutants)} analysis mutants caught; runtime conversion is not tested here', flush=True)
