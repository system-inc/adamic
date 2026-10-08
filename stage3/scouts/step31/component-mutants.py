#!/usr/bin/env python3
"""Kill mutations in actual checker and emitter source with successful Node execution."""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
import subprocess

ROOT = Path(__file__).resolve().parent
parser = argparse.ArgumentParser()
parser.add_argument('tree', type=Path)
parser.add_argument('output', type=Path)
args = parser.parse_args()
args.output.mkdir(parents=True, exist_ok=False)
request = ROOT / 'fixtures/components.json'

def run(tree, mode, label):
    output = args.output / (label + '.stdout')
    errors = args.output / (label + '.stderr')
    with output.open('wb') as out, errors.open('wb') as err:
        result = subprocess.run(['node', str(ROOT / 'source.mjs'), str(tree), '--' + mode, str(request)],
            stdout=out, stderr=err, timeout=120)
    if result.returncode or errors.read_bytes():
        raise RuntimeError(label + ': execution failures do not count as semantic kills')
    return output.read_bytes()

report = []
for mode, function, before, after in [
    ('checker', 'checkTypeAssignableToAndOptionallyElaborate',
     'return checkTypeRelatedToAndOptionallyElaborate(source, target, assignableRelation, errorNode, expr, headMessage, containingMessageChain, /*errorOutputContainer*/ undefined);', 'return true;'),
    ('emitter', 'emitVariableDeclaration',
     'emitInitializer(node.initializer, node.type?.end ?? node.name.emitNode?.typeNode?.end ?? node.name.end, node, parenthesizer.parenthesizeExpressionForDisallowedComma);', '/* mutant omits the initializer */'),
]:
    baseline = run(args.tree.resolve(), mode, mode + '-baseline')
    if baseline != (ROOT / 'fixtures' / (mode + '-golden.jsonl')).read_bytes():
        raise RuntimeError(mode + ': source observation differs from stock golden')
    tree = args.output / mode
    shutil.copytree(args.tree / 'src', tree / 'src')
    source = tree / 'src/compiler' / (mode + '.ts')
    text = source.read_text()
    start = text.index('    function ' + function + '(')
    end = text.index('\n    function ', start + 1)
    body = text[start:end]
    if body.count(before) != 1:
        raise RuntimeError('mutant anchor changed')
    source.write_text(text[:start] + body.replace(before, after) + text[end:])
    actual = run(tree, mode, mode + '-mutant')
    if actual == baseline:
        raise RuntimeError(mode + ': mutant survived')
    left = [json.loads(line) for line in baseline.splitlines()]
    right = [json.loads(line) for line in actual.splitlines()]
    if mode == 'checker':
        before_codes = [d['code'] for row in left for d in row.get('diagnostics', [])]
        after_codes = [d['code'] for row in right for d in row.get('diagnostics', [])]
        if 2322 not in before_codes or 2322 in after_codes:
            raise RuntimeError('checker mutant did not remove the intended assignment diagnostic')
    else:
        if not any(row.get('record') == 'output' and row.get('text') != other.get('text') for row, other in zip(left, right)):
            raise RuntimeError('emitter mutant did not change emitted text')
    report.append({'mode': mode, 'function': function, 'before': before, 'after': after,
        'exit': 0, 'stderr_bytes': 0, 'caught_by': mode + ' bytes versus stock Node golden',
        'baseline_sha256': hashlib.sha256(baseline).hexdigest(), 'mutant_sha256': hashlib.sha256(actual).hexdigest()})
    print(mode + ': actual source mutant caught by output bytes, exit 0, empty stderr', flush=True)
(args.output / 'report.json').write_text(json.dumps(report, indent=2) + '\n')
