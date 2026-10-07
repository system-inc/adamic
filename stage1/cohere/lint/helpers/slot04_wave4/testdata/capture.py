"""Capture each claimed helper's real consumer inputs through a Go overlay."""
import json
import os
from pathlib import Path
import subprocess
import tempfile
import sys
HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[5]
COHERE = ROOT / 'cohere'
LEDGER = HERE.parents[1] / 'readiness.json'
SYMBOL = 'github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.sortedKeys'
PREFIX = 'tailwind-'
consumers = {r['rule'] for r in json.loads(LEDGER.read_text())['remaining'] if SYMBOL in r['remaining_helpers']}
with tempfile.TemporaryDirectory(prefix='adamic-slot04-capture-') as scratch:
    scratch = Path(scratch)
    harness = COHERE / 'internal/lint/testing/rule_testing.go'
    anchor = 'return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}'
    source = harness.read_text()
    assert source.count(anchor) == 1
    side = scratch / 'harness.go'
    side.write_text(source.replace(anchor, 'result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\n RecordAssertedCase(t, result)\n return result'))
    overlay = scratch / 'overlay.json'
    replacements = {str(harness): str(side)}
    if PREFIX:
        literals = COHERE / 'internal/lint/rules/tailwind/collapse/design_system.go'
        text = literals.read_text()
        anchor = 'keys := make([]string, 0, len(set))'
        assert text.count(anchor) == 1
        adapted = scratch / 'class_literals.go'
        adapted.write_text(text.replace(anchor, 'AdamicRecordSortedKeys(set)\n '+anchor))
        replacements[str(literals)] = str(adapted)
        replacements[str(COHERE / 'internal/lint/rules/tailwind/collapse/adamic_slot04_sort_exports.go')] = str(HERE / 'sort_exports.go')
        fixture = scratch / 'fixture'
        with (HERE.parent / 'evidence/tailwind-install.log').open('w') as log:
            subprocess.run(['npm', 'install', '--ignore-scripts', '--no-audit', '--no-fund', '--prefix', str(fixture), 'tailwindcss@4.3.3'], stdout=log, stderr=subprocess.STDOUT, check=True)
        styles = fixture / 'app/_theme/styles'
        styles.mkdir(parents=True)
        (styles / 'theme.css').write_text('@import "tailwindcss";\n')
        for name in ['no_unknown_classes_test.go', 'enforce_consistent_class_order_test.go']:
            original = COHERE / 'internal/lint/rules/tailwind' / name
            adapted = scratch / name
            text = original.read_text()
            assert '/Users/kirkouimet/Projects/ahra/app/_theme/styles' in text
            adapted.write_text(text.replace('/Users/kirkouimet/Projects/ahra/app/_theme/styles', str(styles)))
            replacements[str(original)] = str(adapted)

    overlay.write_text(json.dumps({'Replace': replacements}))
    env = os.environ | {'COHERE_DOCS_CAPTURE': str(scratch / 'capture'), 'ADAMIC_SLOT04_SORTED_KEYS': str(scratch / 'settings.jsonl')}
    for family in (['tailwind'] if PREFIX else ['next', 'react', 'structure']):
        with (HERE.parent / 'evidence' / (PREFIX + 'capture-' + family + '.log')).open('w') as log:
            subprocess.run(['go', 'test', '-overlay='+str(overlay), './internal/lint/rules/'+family, '-count=1', '-timeout=10m'] + (['-run', '^Test(Enforce|No|Class(Literal|Segment|Template|Values|Order(Message|Fix|Reports|Dimensions|Variant|Depth|Readings|Roots|Markers|Fixtures|Sorts|Leaves|Options))|CanonicalFixturesActuallyRan|ConflictFixturesActuallyRan|UnknownClassFixturesActuallyRan)'] if PREFIX else []), cwd=COHERE, env=env, stdout=log, stderr=subprocess.STDOUT, check=True)
    unique = {}
    for file in sorted((scratch / 'capture').glob('*.jsonl')):
        for line in file.read_text().split('\n'):
            if not line: continue
            row = json.loads(line)
            if row['rule'] in consumers:
                key = (row['rule'], row['file'], row['source'])
                unique[key] = {'name': row['rule'] + ':' + row['file'].split('/')[-1], 'source': row['source']}
    rows = [unique[key] for key in sorted(unique)]
    missing = consumers - {row['name'].split(':')[0] for row in rows}
    assert not missing, sorted(missing)
    (HERE / 'sort-source-consumers.json').write_text(json.dumps(rows, ensure_ascii=True, indent=2)+'\n')
    print('captured', len(rows), 'unique inputs from', len(consumers), 'consumers')

    captured = [json.loads(line) for line in (scratch / 'settings.jsonl').read_text().split('\n') if line]
    unique = {json.dumps(row, sort_keys=True, ensure_ascii=True): row for row in captured}
    rows = [unique[key] for key in sorted(unique)]
    (HERE / 'sort-consumers.json').write_text(json.dumps(rows, ensure_ascii=True, indent=2)+'\n')
    print('captured', len(rows), 'distinct real sortedKeys maps from', len(captured), 'entry calls')
