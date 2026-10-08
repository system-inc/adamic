"""Run independent comparator mutants in disposable package copies."""
from pathlib import Path
import json
import shutil
import subprocess
import tempfile

here = Path(__file__).resolve().parent
original = (here / 'compare.go').read_text()
mutants = {
    'ignore-record': ('if !bytes.Equal(l, r) {', 'if !bytes.Equal(l, r) && false {'),
    'ignore-eof': ('if !bytes.Equal(l, r) {', 'if len(l) == 0 || len(r) == 0 { return nil, nil }; if !bytes.Equal(l, r) {'),
    'ignore-final-newline': ('if !bytes.Equal(l, r) {', 'if !bytes.Equal(bytes.TrimSuffix(l, []byte("\\n")), bytes.TrimSuffix(r, []byte("\\n"))) {'),
    'do-not-reset-preorder': ('c.node, c.diagnostic, c.jsdoc = 0, 0, 0', 'c.diagnostic, c.jsdoc = 0, 0'),
}
results = []
with tempfile.TemporaryDirectory(prefix='step24-comparator-mutants-') as tmp:
    root = Path(tmp)
    (root / 'go.mod').write_text('module step24mutants\n\ngo 1.27\n')
    shutil.copyfile(here / 'compare_test.go', root / 'compare_test.go')
    for name, (before, after) in mutants.items():
        assert original.count(before) == 1
        (root / 'compare.go').write_text(original.replace(before, after))
        log = here.parent / 'evidence' / (name + '.log')
        with log.open('w') as stream:
            result = subprocess.run(['go', 'test', '-count=1', '.'], cwd=root, stdout=stream, stderr=subprocess.STDOUT)
        content = log.read_text()
        assert result.returncode == 1 and '--- FAIL:' in content and '[build failed]' not in content, content
        results.append({'mutant': name, 'exit': result.returncode, 'caught_by': [l.strip() for l in content.splitlines() if '--- FAIL:' in l]})
(here.parent / 'evidence/comparator-mutants.json').write_text(json.dumps(results, indent=2)+'\n')
print(json.dumps(results, indent=2))
