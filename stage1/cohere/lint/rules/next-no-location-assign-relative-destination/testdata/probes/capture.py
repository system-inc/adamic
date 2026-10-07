"""Run original Go assertions and retain filename-sensitive corpus and positive controls."""
import json, os, pathlib, subprocess, sys
root = pathlib.Path(__file__).resolve().parents[7]
repo = root / 'cohere'
work = pathlib.Path(sys.argv[1]).resolve()
work.mkdir(parents=True, exist_ok=True)
names = ['NoHeadElement','NoHeadImportInDocument','NoHtmlLinkForPages','NoImgElement','NoLocationAssignRelativeDestination','NoPageCustomFont','NoStyledJsxInDocument','NoSyncScripts','NoTypos','NoUnwantedPolyfillio']
harness = repo / 'internal/lint/testing/rule_testing.go'
text = harness.read_text()
anchor = 'return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}'
assert text.count(anchor) == 1
replacement = 'result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\n RecordAssertedCase(t, result)\n return result'
side = work / 'rule_testing.go'
side.write_text(text.replace(anchor, replacement))
probe = pathlib.Path(__file__).with_name('boundary_test.go')
overlay = work / 'overlay.json'
overlay.write_text(json.dumps({'Replace': {str(harness): str(side), str(repo / 'internal/lint/rules/next/adamic_filename_boundary_test.go'): str(probe)}}))
capture = work / 'capture'
env = dict(os.environ, COHERE_DOCS_CAPTURE=str(capture))
with (work / 'upstream.log').open('w') as log:
    subprocess.run(['go','test','-overlay='+str(overlay),'./internal/lint/rules/next','-run','^Test('+'|'.join(names+['AdamicFilenameBoundary'])+')','-count=1','-v','-timeout=10m'],cwd=repo,env=env,stdout=log,stderr=log,check=True)
records = {}
for path in sorted(capture.glob('*.jsonl')):
    for line in path.read_text().splitlines():
        record = json.loads(line)
        if not record['rule'].startswith('@next/next/'):
            continue
        key = json.dumps([record['rule'], record['file'], record['source'], record.get('options')])
        records[key] = record
rows = sorted(records.values(),key=lambda x:(x['rule'],x['file'],x['source']))
assert rows
(work / 'corpus.json').write_text(json.dumps(rows,indent=2)+'\n')
for name in sorted(set(r['rule'] for r in rows)):
    selected = [r for r in rows if r['rule']==name]
    findings = [r for r in selected if r.get('findings')]
    assert findings, name+' has no positive control'
    print(name, 'cases='+str(len(selected)), 'positive='+str(len(findings)))
print('unique filename/source/rule/options cases',len(rows))
