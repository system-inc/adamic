from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[3]
evidence = root / 'review/compiler/cohere-dry-guard'
target = root / 'internal/flow/effects.go'
original = target.read_bytes()
scratch = root / 'internal/dry_guard_mutant.go'
assert not scratch.exists()
cases = {
    'unlisted-header': '// Lifted from cohere\npackage internal\n',
    'unlisted-rule': 'package internal\nconst diagnostic = "refused (adamic/invariant-mutable)"\n',
    'stale-entry': 'package flow\n',
}
for name, source in cases.items():
    (evidence / (name + '.go.txt')).write_text(source)
    try:
        if name == 'stale-entry':
            target.write_text(source)
        else:
            scratch.write_text(source)
        with (evidence / (name + '.log')).open('w') as log:
            result = subprocess.run(['go', 'test', './internal/dry', '-run', '^TestCohereMirrors$', '-count=1', '-timeout', '90s'], cwd=root, stdout=log, stderr=subprocess.STDOUT, timeout=100)
        output = (evidence / (name + '.log')).read_text()
        expected = 'stale lift mirror' if name == 'stale-entry' else ('cohere source provenance is unlisted' if name == 'unlisted-header' else 'emits cohere rule adamic/invariant-mutable')
        if result.returncode != 1 or expected not in output:
            raise RuntimeError(f'{name}: wrong result {result.returncode}: {output}')
        print(f'{name}: exit 1, TestCohereMirrors caught {expected}', flush=True)
    finally:
        target.write_bytes(original)
        if scratch.exists():
            scratch.unlink()
