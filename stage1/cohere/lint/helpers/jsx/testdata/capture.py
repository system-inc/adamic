"""Capture every asserted upstream case of every JSX helper consumer at the pinned Go revision."""
import json, os, subprocess, sys
from pathlib import Path
here = Path(__file__).resolve().parent
root = here.parents[5]
cohere = root / 'cohere'
out = Path(sys.argv[1])
sys.path.insert(0, str(root / 'stage1/cohere/lint/helpers/testdata'))
from pin import capture_pin
pin = capture_pin(root)
(out / 'capture-pin.json').write_text(json.dumps({'pin': pin}) + '\n')
ledger = json.loads((here.parents[1] / 'comments/readiness.json').read_text())
consumers = {r['rule'] for r in ledger['remaining'] if any('/jsx.' in h for h in r['remaining_helpers'])}
harness = cohere / 'internal/lint/testing/rule_testing.go'
anchor = 'return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}'
s = harness.read_text()
assert s.count(anchor) == 1
side = out / 'harness.go'
side.write_text(s.replace(anchor, 'result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\n RecordAssertedCase(t, result)\n return result'))
overlay = out / 'capture-overlay.json'
overlay.write_text(json.dumps({'Replace': {str(harness): str(side)}}))
env = os.environ | {'COHERE_DOCS_CAPTURE': str(out / 'capture')}
for family in ['next', 'react', 'structure']:
    log = out / ('capture-' + family + '.log')
    with log.open('w') as f:
        p = subprocess.run(['go', 'test', '-overlay='+str(overlay), './internal/lint/rules/'+family, '-count=1', '-timeout=10m'], cwd=cohere, env=env, stdout=f, stderr=subprocess.STDOUT)
    if p.returncode:
        print(log.read_text()); sys.exit(p.returncode)
rows = {}
for path in sorted((out / 'capture').glob('*.jsonl')):
    for line in path.read_text().split('\n'):
        if not line: continue
        row = json.loads(line)
        if row['rule'] in consumers:
            key = (row['rule'], row['file'], row['source'])
            rows[key] = {'rule':row['rule'], 'file':row['file'], 'source':row['source']}
assert consumers <= {r['rule'] for r in rows.values()}, sorted(consumers - {r['rule'] for r in rows.values()})
(out / 'sources.json').write_text(json.dumps(list(rows.values()), ensure_ascii=True))
assert len(rows) == 1284, ('upstream capture count drift',len(rows))
print('captured',len(rows),'distinct upstream inputs from',len(consumers),'JSX consumers')
for rule in sorted(consumers):
    print(rule, sum(r['rule']==rule for r in rows.values()))
