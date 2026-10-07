#!/usr/bin/env python3
"""Capture actual cohere test inputs without editing its worktree."""
import argparse
import collections
import gzip
import json
import os
from pathlib import Path
import re
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[5]
HELPERS = ROOT / 'stage1/cohere/lint/helpers'
SYMBOLS = [
    'github.com/system-inc/cohere/internal/lint/ecmascript/jsx.AttributeName',
    'github.com/system-inc/cohere/internal/lint/ecmascript/property.Name',
]
parser = argparse.ArgumentParser()
parser.add_argument('--reuse-capture', type=Path)
args = parser.parse_args()
ledger = json.loads((HELPERS / 'readiness.json').read_text())['remaining']
consumers = {symbol: sorted(row['rule'] for row in ledger if symbol in row['remaining_helpers']) for symbol in SYMBOLS}
rules = set().union(*consumers.values())
inventory = json.loads((HELPERS / '../inventory/inventory.json').read_text())
paths = sorted({path for row in inventory['rules'] if row['name'] in rules for path in row['tests']['files']})
tests = sorted({name for path in paths for name in re.findall(r'^func (Test\w+)\(', (ROOT / path).read_text(), re.M)})
packages = sorted({str(Path(path).parent).removeprefix('cohere/') for path in paths})

with tempfile.TemporaryDirectory(prefix='lint-helpers-02-capture-') as scratch_name:
    scratch = Path(scratch_name)
    capture = args.reuse_capture or scratch / 'records'
    if args.reuse_capture is None:
        capture.mkdir()
        # This consumer's custom multi-file harness bypasses docs capture. The
        # overlay records its actual project sources after the rule ran.
        original = ROOT / 'cohere/internal/lint/rules/nexus/localization_no_untranslated_value_test.go'
        source = original.read_text()
        source = source.replace('"github.com/system-inc/cohere/internal/lint/rule"',
                                '"github.com/system-inc/cohere/internal/lint/rule"\n "github.com/system-inc/cohere/internal/docsdata/capture"')
        anchor = 'return rule_testing.Result{Diagnostics: diagnostics, SourceFile: subject}'
        if source.count(anchor) != 1:
            raise RuntimeError('localization capture anchor changed')
        source = source.replace(anchor, '''for _, sourceFile := range graph.ProjectFiles() {
 rel,err := filepath.Rel(directory,sourceFile.FileName());if err!=nil{t.Fatal(err)}
 if err:=capture.Write(capture.Record{Rule:LocalizationNoUntranslatedValue.Name,File:filepath.ToSlash(rel),Source:sourceFile.Text(),Outcome:"Input"});err!=nil{t.Fatal(err)}
 }
 ''' + anchor)
        side = scratch / 'localization.go'
        side.write_text(source)
        overlay = scratch / 'overlay.json'
        overlay.write_text(json.dumps({'Replace': {str(original): str(side)}}))
        command = ['go', 'test', '-overlay=' + str(overlay), '-count=1', '-timeout=20m', '-run', '^(' + '|'.join(tests) + ')$'] + ['./' + package for package in packages]
        env = dict(os.environ, COHERE_DOCS_CAPTURE=str(capture))
        log_path = HELPERS / 'slot02/evidence/capture.log'
        log_path.parent.mkdir(parents=True, exist_ok=True)
        with log_path.open('w') as log:
            subprocess.run(command, cwd=ROOT / 'cohere', env=env, stdout=log, stderr=subprocess.STDOUT, check=True)
    counts = collections.Counter()
    unique = set()
    for path in sorted(capture.glob('capture-*.jsonl')):
        for line in path.read_text().splitlines():
            row = json.loads(line)
            if row['rule'] not in rules:
                continue
            counts[row['rule']] += 1
            # Only the suffix affects parsing for these helpers. Stable file
            # identities avoid baking temporary test-directory names into data.
            suffix = Path(row['file']).suffix
            if suffix not in ('.ts', '.tsx', '.js', '.jsx', '.a'):
                raise RuntimeError('unsupported captured source extension ' + suffix)
            unique.add((row['rule'], '/fixture' + suffix, row['source']))
    missing = rules - counts.keys()
    if missing:
        raise RuntimeError('no actual captured inputs for ' + ', '.join(sorted(missing)))
    rows = [{'Rule': rule, 'File': file, 'Source': source} for rule, file, source in sorted(unique)]
    data = json.dumps(rows, ensure_ascii=True, separators=(',', ':')).encode() + b'\n'
    (HELPERS / 'testdata/slot02_inputs.json.gz').write_bytes(gzip.compress(data, mtime=0))
    coverage = {'cohere_commit': subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=ROOT / 'cohere', text=True).strip(),
                'consumers': consumers, 'captured_records': dict(sorted(counts.items())),
                'unique_inputs': dict(sorted(collections.Counter(row['Rule'] for row in rows).items())),
                'test_files': paths, 'test_functions': tests}
    (HELPERS / 'testdata/slot02_coverage.json').write_text(json.dumps(coverage, indent=2, sort_keys=True) + '\n')
    print(f'{len(rules)} consumers; {sum(counts.values())} captured records; {len(rows)} unique rule/file/source inputs')
