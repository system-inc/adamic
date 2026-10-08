#!/usr/bin/env python3
"""Observe each real helper invocation in every consuming rule's upstream tests."""
import collections
import hashlib
import json
import os
from pathlib import Path
import subprocess
import tempfile

owned = Path(__file__).resolve().parents[1]
root = owned.parents[4]
cohere = root / 'cohere'
assert subprocess.check_output(['git', '-C', str(cohere), 'rev-parse', 'HEAD'], text=True).strip() == '7945d102a6c18dd36adf9114a758ce646e8b2359'
out = owned / 'testdata'
evidence = owned / 'evidence'
consumers = ['NoExplicitAny', 'BanTsComment', 'NoThisAlias', 'TripleSlashReference',
             'NoExtraNonNullAssertion', 'NoNonNullAssertion', 'NoNonNullAssertedOptionalChain']
with tempfile.TemporaryDirectory(prefix='typescript-helper-capture-') as temporary:
    scratch = Path(temporary)
    replacements = {}
    for file, symbol in [('no_explicit_any.go', 'isTypeScriptSourceFile'),
                         ('no_extra_non_null_assertion.go', 'nonNullAssertionOperatorRange')]:
        original = cohere / 'internal/lint/rules/typescript' / file
        text = original.read_text()
        anchor = 'func ' + symbol + '('
        assert text.count(anchor) == 1
        patched = scratch / file
        renamed = 'adamicOriginal' + symbol[0].upper() + symbol[1:]
        patched.write_text(text.replace(anchor, 'func ' + renamed + '(', 1))
        replacements[str(original)] = str(patched)
    directory = cohere / 'internal/lint/rules/typescript'
    replacements[str(directory / 'adamic_capture.go')] = str(out / 'capture.go.txt')
    replacements[str(directory / 'adamic_controls_test.go')] = str(out / 'controls_test.go.txt')
    overlay = scratch / 'overlay.json'
    overlay.write_text(json.dumps({'Replace': replacements}))
    cases = []
    coverage = {}
    for consumer in consumers + ['AdamicTypescriptHelperControls']:
        observations = scratch / (consumer + '.jsonl')
        environment = dict(os.environ, ADAMIC_TYPESCRIPT_HELPER_CAPTURE=str(observations))
        command = ['go', 'test', '-overlay=' + str(overlay), './internal/lint/rules/typescript',
                   '-count=1', '-json', '-timeout=5m', '-run', '^Test' + consumer]
        log = evidence / ('capture-' + consumer + '.jsonl')
        with log.open('w') as stream:
            result = subprocess.run(command, cwd=cohere, env=environment, stdout=stream, stderr=subprocess.STDOUT)
        if result.returncode:
            raise RuntimeError(str(command) + ' failed; see ' + str(log))
        events = [json.loads(line) for line in log.read_text().splitlines()]
        assert not any(e.get('Action') == 'skip' for e in events), consumer + ' skipped'
        rows = [json.loads(line) for line in observations.read_text().splitlines()]
        assert rows, consumer + ' made no helper calls'
        coverage[consumer] = {'invocations': len(rows), 'by_helper': dict(collections.Counter(r['kind'] for r in rows)),
                              'passed_tests': sum(e.get('Action') == 'pass' and bool(e.get('Test')) for e in events)}
        for row in rows:
            row['consumer'] = consumer
        cases.extend(rows)
    (out / 'cases.json').write_text(json.dumps(cases, ensure_ascii=False, indent=2) + '\n')
    (out / 'coverage.json').write_text(json.dumps(coverage, indent=2) + '\n')
    files = ['internal/lint/rules/typescript/' + name for name in ['no_explicit_any.go', 'no_extra_non_null_assertion.go', 'no_explicit_any_test.go', 'ban_ts_comment_test.go', 'no_this_alias_test.go', 'triple_slash_reference_test.go', 'no_extra_non_null_assertion_test.go', 'no_non_null_assertion_test.go', 'no_non_null_asserted_optional_chain_test.go']] + ['internal/types/sourcename/sourcename.go']
    metadata = {'cohere': '7945d102a6c18dd36adf9114a758ce646e8b2359', 'sha256': {name: hashlib.sha256((cohere/name).read_bytes()).hexdigest() for name in files}}
    (out/'provenance.json').write_text(json.dumps(metadata, indent=2) + '\n')
    print(json.dumps(coverage, indent=2))
    print('total helper invocations:', len(cases))
