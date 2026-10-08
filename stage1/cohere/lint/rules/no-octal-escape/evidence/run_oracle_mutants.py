"""Run adapter mutations through scratch-only Go overlays."""
import json
from pathlib import Path
import subprocess
import tempfile

root = Path(__file__).resolve().parents[6]
package = root / 'stage1/cohere/lint'
shared = (package / 'shared_test.go').read_text()
driver = (package / 'testdata/oracle.go').read_text()
logs = Path('/tmp/octal-oracle-mutants')
logs.mkdir(exist_ok=True)
for name, run, witness in [
    ('capture-drops-recovery', 'TestCapturedOracleRecovery', 'OctalEscape.ts'),
    ('oracle-refuses-unsupported', 'TestCapturedOracleRecovery', 'unsupported.ts'),
    ('overwrite-port-boundary', 'TestRecoveryClassificationPreservesModes', 'classification changed'),
    ('missing-flags-accepted', 'TestRecoveryClassificationPreservesModes', 'missing diagnostic flags were accepted'),
]:
    with tempfile.TemporaryDirectory(prefix='octal-oracle-mutant-') as temporary:
        scratch = Path(temporary)
        mutated = shared
        if name == 'capture-drops-recovery':
            mutated = mutated.replace('return classifyRecoveryRows(rows, strings.Fields(string(flags)))', '_ = flags\n\treturn rows, nil', 1)
        elif name == 'oracle-refuses-unsupported':
            side = scratch / 'oracle.go'
            side.write_text(driver.replace(' || fields[6] == "unsupported-recovery"', '').replace(' && fields[6] != "unsupported-recovery"', ''))
            anchor = 'side, err := filepath.Abs(filepath.Join(packageDirectory, "testdata/oracle.go"))'
            assert mutated.count(anchor) == 1
            mutated = mutated.replace(anchor, 'side := ' + json.dumps(str(side)) + '\n\terr = nil', 1)
        elif name == 'overwrite-port-boundary':
            mutated = mutated.replace('if fields[6] == "" {', 'if true {', 1)
        else:
            mutated = mutated.replace('return nil, fmt.Errorf("diagnostics answered %d rows of %d", len(flags), len(rows))', 'return rows, nil', 1)
        assert mutated != shared
        side = scratch / 'shared_test.go'
        side.write_text(mutated)
        overlay = scratch / 'overlay.json'
        overlay.write_text(json.dumps({'Replace': {str(package / 'shared_test.go'): str(side)}}))
        log = logs / (name + '.log')
        with log.open('w') as output:
            result = subprocess.run(['go', 'test', '-v', '-count=1', '-timeout', '30m', '-overlay=' + str(overlay), '-run', '^' + run + '$', './stage1/cohere/lint'], cwd=root, stdout=output, stderr=subprocess.STDOUT)
        text = log.read_text()
        assert result.returncode == 1 and witness in text and ('--- FAIL: ' + run) in text, (name, str(log), text[-1000:])
        print(f'{name}: caught by {run}, naming {witness}; {log}', flush=True)
