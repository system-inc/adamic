"""Capture actual asserted Go consumer inputs without changing any rule."""
import json
import os
from pathlib import Path
import subprocess
import tempfile

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[5]
COHERE = ROOT / 'cohere'
ledger = json.loads((HERE.parents[1] / 'readiness.json').read_text())
consumers = {r['rule'] for r in ledger['remaining'] if any('/comments.' in h for h in r['remaining_helpers'])}
with tempfile.TemporaryDirectory(prefix='adamic-comments-capture-') as scratch:
    scratch = Path(scratch)
    harness = COHERE / 'internal/lint/testing/rule_testing.go'
    original = 'return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}'
    source = harness.read_text()
    assert source.count(original) == 1
    replacement = 'result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\n RecordAssertedCase(t, result)\n return result'
    side = scratch / 'harness.go'
    side.write_text(source.replace(original, replacement))
    overlay = scratch / 'overlay.json'
    overlay.write_text(json.dumps({'Replace': {str(harness): str(side)}}))
    env = os.environ | {'COHERE_DOCS_CAPTURE': str(scratch / 'capture')}
    patterns = {
        'core': 'RequireDescription|ArrowBodyStyle|DefaultCase|MaxLines|NoExtraBind|NoExtraLabel|NoFallthrough|NoInlineComments|NoInvalidThis|NoIrregularWhitespace|NoUnusedLabels|NoUselessRename|OperatorAssignment|DotNotation|NoEmptyFunction',
        'typescript': 'BanTslintComment|NoInvalidThis|PreferFunctionType|TripleSlashReference',
        'react': 'JsxCurlyBracePresence|JsxKey|NoDeprecated',
        'structure': 'ConsistencyOrganizeImports|ReactHookRequireEffectComment',
    }
    for family, pattern in patterns.items():
        with (HERE.parent / 'evidence' / ('capture-' + family + '.log')).open('w') as log:
            subprocess.run(['go', 'test', '-overlay='+str(overlay), './internal/lint/rules/'+family, '-run', '^Test('+pattern+')', '-count=1', '-timeout=10m'], cwd=COHERE, env=env, stdout=log, stderr=subprocess.STDOUT, check=True)
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
    (HERE / 'consumers.json').write_text(json.dumps(rows, ensure_ascii=True, indent=2)+'\n')
    print('captured', len(rows), 'unique inputs from', len(consumers), 'of', len(consumers), 'consumers')
