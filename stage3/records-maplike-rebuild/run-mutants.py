#!/usr/bin/env python3
"""Rerun record refusal checks with Go overlays, without changing production files."""
from pathlib import Path
import json
import os
import subprocess

root = Path(__file__).resolve().parents[2]
logs = Path(os.environ.get('RECORDS_REBUILD_MUTANT_LOGS', '/tmp/records-maplike-rebuild-mutants'))
logs.mkdir(parents=True, exist_ok=True)


def text(file):
    return (root / ('internal/lower/' + file)).read_text()


def replace(file, before, after):
    source = text(file)
    assert before in source, (file, before)
    return source.replace(before, after, 1)


def early(file, function, statement):
    source = text(file)
    at = source.index('func (l *lowering) ' + function + '(')
    insert = source.index('\n', at) + 1
    return source[:insert] + '\t' + statement + '\n' + source[insert:]


records = text('records.go')
proto = replace('records.go', 'if read && (n.Kind == ast.KindStringLiteral', 'if false && (n.Kind == ast.KindStringLiteral')
proto = proto.replace('if read && recordPrototypeName(name)', 'if false')
proto = proto.replace('if recordPrototypeName(name) {', 'if false {')
cycles = text('cycles.go')
start = cycles.index('\tcase l.recordElement(holder) != nil:')
end = cycles.index('\tcase l.isLibraryType(holder, "ReadonlyMap")', start)
cycles = cycles[:start] + '\tcase l.recordElement(holder) != nil:\n' + cycles[end:]
invariant = replace('invariance.go', 'if source, target := l.recordElement(from), l.recordElement(to); source != nil && target != nil {', 'if source, target := l.recordElement(from), l.recordElement(to); source != nil && target != nil {\n\t\treturn nil // deliberate mutant')
spread = replace('records.go', 'if widened := l.widened(element, t, map[[2]*checker.Type]bool{}); widened != nil {', 'if widened := l.widened(element, t, map[[2]*checker.Type]bool{}); false && widened != nil {')
nullable = replace('expression.go', ' || of == ir.Record || of == ir.String', ' || of == ir.String')
nullable = nullable.replace('if shared == ir.Record && l.includesNull(proven) {', 'if false {')
mutants = [
    ('readonly-index', {'records.go': replace('records.go', ' || infos[0].IsReadonly() ||', ' || false ||')}, 'TestRecordRefusals', 'want "signature"'),
    ('numeric-index', {'records.go': replace('records.go', 'infos[0].KeyType().Flags() != checker.TypeFlagsString || infos[0].IsReadonly()', 'false || infos[0].IsReadonly()')}, 'TestRecordRefusals', 'want "signature"'),
    ('prototype-literal', {'records.go': proto}, 'TestRecordPrototypeLiteralNames', 'want literal member refusal'),
    ('storage-views', {'records.go': early('records.go', 'sameRecordStorage', 'return true')}, 'TestRecordRefusals', 'want "seen as"'),
    ('mutable-invariance', {'invariance.go': invariant}, 'TestRecordRefusals', 'want "invariant-mutable"'),
    ('record-cycles', {'cycles.go': cycles}, 'TestRecordRefusals', 'want "cycle"'),
    ('spread-invariance', {'records.go': spread}, 'TestRecordRefusals', 'want "invariant-mutable"'),
    ('logical-record-conversion', {'expression.go': replace('expression.go', 'if left.Type() == ir.Record && of != ir.Record && of != ir.Union && !of.IsMaybe() {', 'if false {')}, 'TestRecordRefusals', 'want "logical record operand"'),
    ('nullable-record-container', {'expression.go': nullable}, 'TestRecordRefusals', 'want "value of type"'),
    ('opaque-own-argument', {'detached_own.go': replace('detached_own.go', 'if contextual != nil && contextual.Flags()&checker.TypeFlagsNonPrimitive != 0 && l.detachedOwnOpaqueArgument(at) {', 'if false && contextual != nil && contextual.Flags()&checker.TypeFlagsNonPrimitive != 0 && l.detachedOwnOpaqueArgument(at) {')}, 'TestDetachedOwnRepresentation', 'must not compile'),
    ('shorthand-own-escape', {'detached_own.go': replace('detached_own.go', 'if node.Kind == ast.KindShorthandPropertyAssignment {', 'if false {')}, 'TestDetachedOwnRefusals', 'want detached-method refusal'),
    ('type-only-index-admission', {'refusals.go': replace('refusals.go', 'var refusals = map[ast.Kind]refusal{', 'var refusals = map[ast.Kind]refusal{\n\tast.KindIndexSignature: {"an index signature", "deliberate mutant"},')}, 'TestRecordForms', 'refuses'),
]
for name, files, test, witness in mutants:
    replacements = {}
    for file, source in files.items():
        assert source != text(file), name
        modified = logs / (name + '-' + file)
        modified.write_text(source)
        replacements[str(root / ('internal/lower/' + file))] = str(modified)
    overlay = logs / (name + '.json')
    overlay.write_text(json.dumps({'Replace': replacements}))
    log = logs / (name + '.log')
    with log.open('w') as output:
        result = subprocess.run(['go', 'test', '-overlay', str(overlay), './internal/lower', '-run', '^' + test + '$', '-count=1', '-timeout', '10m'], cwd=root, stdout=output, stderr=subprocess.STDOUT)
    observed = log.read_text()
    if result.returncode == 0 or witness not in observed:
        raise AssertionError((name, result.returncode, observed))
    print(name + ': caught by ' + test + ' (exit ' + str(result.returncode) + ')', flush=True)
