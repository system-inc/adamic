from pathlib import Path
import json
import subprocess
import tempfile

root = Path.cwd()
with tempfile.TemporaryDirectory(prefix='never-baseline-') as directory:
    scratch = Path(directory)
    replacements = {}
    for name in ('expression.go', 'statements.go'):
        destination = scratch / name
        destination.write_bytes(subprocess.check_output([
            'git', 'show', '4e0bfda50a19c705a1aac0d9932e08483806d61c:internal/lower/' + name]))
        replacements[str(root / 'internal/lower' / name)] = str(destination)
    empty = scratch / 'empty_never.go'
    empty.write_text('package lower\n')
    replacements[str(root / 'internal/lower/never.go')] = str(empty)
    probe = scratch / 'counts_probe_test.go'
    probe.write_text('''package oracle
import "testing"
func TestNeverBaselineCounts(t *testing.T) {
 for _, path := range []string{"internal/oracle/testdata/user_iterators.a", "internal/oracle/testdata/optional_widening_boolean_as_number_array.a", "internal/oracle/testdata/optional_widening_boolean_as_number_return.a"} {
  t.Log(counted(t, path, false, nil, false, false))
 }
}
''')
    replacements[str(root / 'internal/oracle/never_baseline_counts_test.go')] = str(probe)
    overlay = scratch / 'overlay.json'
    overlay.write_text(json.dumps({'Replace': replacements}))
    result = subprocess.run(['go', 'test', '-overlay=' + str(overlay), './internal/oracle',
                             '-run', '^TestNeverBaselineCounts$', '-v', '-count=1'])
    raise SystemExit(result.returncode)
