#!/usr/bin/env python3
"""Prove the qualified-call checks with Go overlays, never editing the checkout.

Prepare oracle's native runtime products first, then run this under the same
product/cache environment as the ordinary qualified-predicate tests.
"""
import argparse
import json
import os
from pathlib import Path
import signal
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[3]
MUTANTS = [
    ('bypass-qualified-check', 'internal/lower/expression.go',
     'return l.censusOverloadResult(call, ir.Call{Function: function, Arguments: arguments, Spread: spread, Returns: l.result.Functions[function].Returns})',
     'if ast.SkipParentheses(call.Expression).Kind == ast.KindPropertyAccessExpression { return ir.Call{Function: function, Arguments: arguments, Spread: spread, Returns: l.result.Functions[function].Returns}, nil }\n\treturn l.censusOverloadResult(call, ir.Call{Function: function, Arguments: arguments, Spread: spread, Returns: l.result.Functions[function].Returns})',
     './internal/oracle', '^TestQualifiedPredicate(NamespaceAssertions|ReexportAssertions|ModuleAssertions)$'),
    ('identifier-only-false-walk', 'internal/lower/predicate_false_flow.go',
     'if f.l.directPredicateCallee(callee, declaration) {',
     'if ast.IsIdentifier(callee) && f.l.directPredicateCallee(callee, declaration) {',
     './internal/lower', '^TestQualifiedPredicateFalseSideReadLowers$'),
]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output', type=Path)
    args = parser.parse_args()
    output = args.output or Path(tempfile.mkdtemp(prefix='qualified-predicate-mutants-'))
    output.mkdir(parents=True, exist_ok=True)
    rows = []
    for name, relative, before, after, package, pattern in MUTANTS:
        directory = output / name
        directory.mkdir(exist_ok=True)
        original = ROOT / relative
        source = original.read_text()
        if source.count(before) != 1:
            raise RuntimeError(f'{name}: original text occurs {source.count(before)} times')
        mutated = directory / original.name
        mutated.write_text(source.replace(before, after))
        overlay = directory / 'overlay.json'
        overlay.write_text(json.dumps({'Replace': {str(original): str(mutated)}}))
        command = ['go', 'test', '-json', '-overlay=' + str(overlay), package,
                   '-run', pattern, '-count=1', '-p', '1', '-parallel', '2', '-timeout', '3m']
        log = directory / 'test.jsonl'
        with log.open('w') as stream:
            process = subprocess.Popen(command, cwd=ROOT, stdout=stream, stderr=stream, start_new_session=True)
            try:
                code = process.wait(timeout=240)
            except subprocess.TimeoutExpired:
                os.killpg(process.pid, signal.SIGKILL)
                process.wait()
                code = 124
        records = []
        for line in log.read_text().splitlines():
            try:
                records.append(json.loads(line))
            except json.JSONDecodeError:
                pass
        text = ''.join(record.get('Output', '') for record in records)
        failed = [record['Test'] for record in records
                  if record.get('Action') == 'fail' and 'Test' in record and '/' not in record['Test']]
        if name == 'bypass-qualified-check':
            expected = ['TestQualifiedPredicate' + suffix for suffix in
                        ['NamespaceAssertions', 'ReexportAssertions', 'ModuleAssertions']]
            caught = all(test in failed for test in expected) and all(
                backend + ' qualified lie did not stop' in text for backend in ['JavaScript', 'ASan/UBSan native'])
        else:
            caught = 'TestQualifiedPredicateFalseSideReadLowers' in failed and "stage 0 can't lower" in text and 'false' in text
        caught = caught and code not in (0, 124)
        rows.append(dict(name=name, exit=code, caught=caught, failed_tests=failed, command=command, log=str(log)))
        print(json.dumps(rows[-1]), flush=True)
    (output / 'results.json').write_text(json.dumps(rows, indent=2) + '\n')
    return 0 if all(row['caught'] for row in rows) else 1


if __name__ == '__main__':
    raise SystemExit(main())
