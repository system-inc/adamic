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
SYMBOLS = {'github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.*UtilityEvaluator.Compile','github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.*UtilityEvaluator.compile','github.com/system-inc/cohere/internal/lint/rules/react.DecodeCompilerRuleOptions'}
PREFIX = 'tailwind-'
consumers = {r['rule'] for r in json.loads(LEDGER.read_text())['remaining'] if SYMBOLS.intersection(r['remaining_helpers'])}
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
        for folder,file,signature,injection in [
         ('internal/lint/rules/tailwind/collapse','utility.go','func (evaluator *UtilityEvaluator) Compile(candidate *ParsedCandidate) ([]*Node, bool) {','AdamicRecordCompile(evaluator,candidate)'),
         ('internal/lint/rules/react','compiler_rule_options.go','func DecodeCompilerRuleOptions(raw []byte) (any, error) {','adamicRecordOptions(raw)')]:
            original=COHERE / folder / file
            text=original.read_text()
            assert text.count(signature)==1
            adapted=scratch / file
            adapted.write_text(text.replace(signature,signature+'\n '+injection+'\n'))
            replacements[str(original)]=str(adapted)
        replacements[str(COHERE / 'internal/lint/rules/tailwind/collapse/adamic_slot04_wave22_exports.go')]=str(HERE / 'exports.go')
        replacements[str(COHERE / 'internal/lint/rules/react/adamic_slot04_wave22_exports.go')]=str(HERE / 'react_exports.go')
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
    env = os.environ | {'COHERE_DOCS_CAPTURE': str(scratch / 'capture'), 'ADAMIC_SLOT04_COMPILE': str(scratch / 'compile.jsonl'),'ADAMIC_SLOT04_OPTIONS': str(scratch / 'options.jsonl')}
    for family in ['tailwind','react']:
        with (HERE.parent / 'evidence' / (PREFIX + 'capture-' + family + '.log')).open('w') as log:
            selected = '^Test(Enforce|NoConflicting|NoUnknown|CanonicalFixturesActuallyRan|ConflictFixturesActuallyRan|UnknownClassFixturesActuallyRan)' if family=='tailwind' else '^Test(Config|ErrorBoundaries|IncompatibleLibrary|VoidUseMemo|CompilerRuleOptions)'
            subprocess.run(['go','test','-overlay='+str(overlay),'./internal/lint/rules/'+family,'-count=1','-timeout=10m','-run',selected],cwd=COHERE,env=env,stdout=log,stderr=subprocess.STDOUT,check=True)

    unique = {}
    for file in sorted((scratch / 'capture').glob('*.jsonl')):
        for line in file.read_text().split('\n'):
            if not line: continue
            row = json.loads(line)
            if row['rule'] in consumers:
                key = (row['rule'], row['file'], row['source'])
                unique[key] = {'name': row['rule'] + ':' + row['file'].split('/')[-1], 'source': row['source'], 'OptionRaw': json.dumps(row.get('options', {}), ensure_ascii=True, separators=(',', ':')) if row['rule'].startswith('react-hooks/') else '{}'}
    rows = [unique[key] for key in sorted(unique)]
    missing = consumers - {row['name'].split(':')[0] for row in rows}
    assert not missing, sorted(missing)
    (HERE / 'consumers.json').write_text(json.dumps(rows, ensure_ascii=True, indent=2)+'\n')
    print('captured', len(rows), 'unique inputs from', len(consumers), 'consumers')

    captured = [json.loads(line) for file in ['compile.jsonl','options.jsonl'] if (scratch / file).exists() for line in (scratch / file).read_text().split('\n') if line]
    unique = {json.dumps(row, sort_keys=True, ensure_ascii=True): row for row in captured}
    rows = [unique[key] for key in sorted(unique)]
    (HERE / 'calls.json').write_text(json.dumps(rows, ensure_ascii=True, indent=2)+'\n')
    print('captured', len(rows), 'distinct real helper inputs from', len(captured), 'entry calls')
