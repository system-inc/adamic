import json
import subprocess
from pathlib import Path

root = Path(__file__).resolve().parents[3]
review = Path(__file__).resolve().parent
selection = '^(TestInheritedLibraryReadsNeverLoadOwnFields|TestPrototypeMethodsAreRefusedWithReasons|TestNullishPrototypeReadsAreRejectedByChecker|TestPrototypeHazardsBehindObjectViewsAreNotYet|TestUnrepresentedPrototypeCallsAreNotYet|TestIsPrototypeOfReadsExplainThePrototypeRefusal)$'
results = []

def run(name, target, changed, select):
    path = root / target
    original = path.read_bytes()
    try:
        path.write_bytes(changed(original))
        patch = subprocess.run(['git', 'diff', '--', target], cwd=root, capture_output=True, check=True).stdout
        (review / (name + '.patch')).write_bytes(patch)
        with (review / (name + '.jsonl')).open('w') as log:
            command = ['go', 'test', './internal/lower', '-run', select, '-count=1', '-timeout', '90s', '-json']
            result = subprocess.run(command, cwd=root, stdout=log, stderr=subprocess.STDOUT, timeout=180)
        results.append({'mutant': name, 'command': command, 'exit': result.returncode})
        print(name, 'exit', result.returncode, flush=True)
    finally:
        path.write_bytes(original)

regex = root / 'internal/lower/regexp.go'
patch = (review / 'M20.patch').read_text()
old = b'escaped, sets := false, strings.Contains(flags, "v")'
new = b'escaped, sets := false, strings.Contains(flags, "u")'
assert ('-\t' + old.decode()) in patch and ('+\t' + new.decode()) in patch
assert regex.read_bytes().count(old) == 1
run('M20', 'internal/lower/regexp.go', lambda source: source.replace(old, new), selection)
old = b'Fix: "Adamic has no observable prototype chain; use instanceof for class identity or an explicit discriminant"'
new = b'Fix: "use instanceof for class identity or an explicit discriminant"'
assert (root / 'internal/lower/prototype.go').read_bytes().count(old) == 2
run('own-refusal-reason', 'internal/lower/prototype.go', lambda source: source.replace(old, new, 1), '^TestIsPrototypeOfReadsExplainThePrototypeRefusal$')
(review / 'mutants.json').write_text(json.dumps(results, indent=2) + '\n')
assert results[0]['exit'] == 0, 'M20 should survive this refusal-only scope'
assert results[1]['exit'] == 1, 'the existing refusal explanation assertion must catch the diagnostic mutant'
