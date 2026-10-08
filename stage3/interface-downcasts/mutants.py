#!/usr/bin/env python3
"""Run one compiler mutation at a time, restore the exact source, retain test logs."""
import os
import subprocess
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
SOURCE = ROOT / 'internal/lower/interface_cast.go'
LOGS = Path(os.environ.get('ADAMIC_INTERFACE_MUTANT_LOGS', '/workspace/interface-downcast-mutants')).resolve()
assert LOGS != ROOT and ROOT not in LOGS.parents, 'run output must stay outside the checkout'


def main():
    LOGS.mkdir(parents=True, exist_ok=True)
    original = SOURCE.read_text()
    tests = [
        ('tag-only', 'if err := l.interfaceConstructions(node, target, field, literal); err != nil {', 'if err := error(nil); err != nil {', './internal/oracle', '^TestInterfaceCastRefusesMalformed$', 'construction check must refuse', 'accepted mutant emitted and built valid C'),
        ('missing-field', 'fail(node, "matching kind lacks required field "+name)', 'continue', './internal/lower', '^TestInterfaceConstructionProof$/^missing_payload$', 'want refusal', 'got <nil>'),
        ('wrong-field-type', '(name != field && !l.checker.IsTypeAssignableTo(actual, wanted))', '(name != field && wanted == nil)', './internal/lower', '^TestInterfaceConstructionProof$/^wrong_payload$', 'want refusal', 'got <nil>'),
        ('ignore-writes', 'l.interfaceWrite(binary.Left, checked, fail)', '// mutant: ignore writes', './internal/lower', '^TestInterfaceConstructionProof$/^alias_write$', 'want refusal', 'got <nil>'),
        ('entry-only', 'for _, module := range modules {', 'for _, module := range modules[len(modules)-1:] {', './internal/oracle', '^TestInterfaceCastImportedConstruction$', 'imported malformed factory', '<nil>'),
    ]
    try:
        for name, old, new, package, select, expected, detail in tests:
            assert original.count(old) == 1, (name, original.count(old))
            SOURCE.write_text(original.replace(old, new, 1))
            log = LOGS / ('mutant-' + name + '.log')
            with log.open('w') as output:
                result = subprocess.run(['go', 'test', package, '-run', select, '-v', '-count=1', '-timeout', '30m'], cwd=ROOT, stdout=output, stderr=subprocess.STDOUT, env={**os.environ, 'ADAMIC_GATE_UNCACHED': '1'})
            text = log.read_text()
            if result.returncode == 0 or expected not in text or detail not in text or '[build failed]' in text or 'Sanitizer' in text:
                raise RuntimeError(f'{name}: mutant was not killed by its intended semantic assertion; inspect {log}')
            print(f'{name}: caught by intended assertion, go test exit {result.returncode}; {log}', flush=True)
            SOURCE.write_text(original)
    finally:
        SOURCE.write_text(original)


if __name__ == '__main__':
    main()
