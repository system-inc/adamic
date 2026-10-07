"""Capture each claimed helper's real consumer inputs through a Go overlay."""
import json
import os
from pathlib import Path
import subprocess
import tempfile
HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[5]
COHERE = ROOT / 'cohere'
LEDGER = HERE.parents[1] / 'readiness.json'
SYMBOLS = {'github.com/system-inc/cohere/internal/lint/ecmascript/regexp.'+name for name in ['checkGroupQuantifier','errNothingToRepeat','checkGroupConstruct']}
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
    original = COHERE / 'internal/lint/ecmascript/regexp/rewrite.go'
    text = original.read_text()
    injections = [
        ('func errNothingToRepeat(quantifier string) error {', 'AdamicRecord(AdamicCase{Source:quantifier})'),
        ('func checkGroupQuantifier(kind groupKind, rest string, options rewriteOptions) error {', 'AdamicRecord(AdamicCase{Source:rest,Kind:int(kind),Unicode:options.unicode})'),
        ('func checkGroupConstruct(source string) error {', 'AdamicRecord(AdamicCase{Source:source})')]
    for signature, injection in injections:
        assert text.count(signature)==1
        text=text.replace(signature,signature+'\n '+injection+'\n')
    side = scratch / 'rewrite.go'; side.write_text(text)
    replacements[str(original)] = str(side)
    replacements[str(COHERE / 'internal/lint/ecmascript/regexp/adamic_slot04_wave13_exports.go')] = str(HERE / 'exports.go')
    overlay.write_text(json.dumps({'Replace': replacements}))
    env = os.environ | {'COHERE_DOCS_CAPTURE': str(scratch / 'capture'), 'ADAMIC_SLOT04_ASSERTIONS': str(scratch / 'calls.jsonl')}
    for family in ['next', 'typescript', 'core']:
        with (HERE.parent / 'evidence' / ('capture-' + family + '.log')).open('w') as log:
            subprocess.run(['go', 'test', '-overlay='+str(overlay), './internal/lint/rules/'+family, '-count=1', '-timeout=10m'], cwd=COHERE, env=env, stdout=log, stderr=subprocess.STDOUT, check=True)
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
    print('captured', len(rows), 'unique inputs from', len(consumers), 'consumers')

    calls=[json.loads(line) for line in (scratch/'calls.jsonl').read_text().splitlines()]
    unique={json.dumps(row,sort_keys=True):row for row in calls}
    (HERE/'calls.json').write_text(json.dumps([unique[key] for key in sorted(unique)],indent=2)+'\n')
    print('captured',len(unique),'distinct helper inputs from',len(calls),'calls')
