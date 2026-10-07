#!/usr/bin/env python3
"""Apply the reviewed origin manifest to the reproducible baseline inventory."""
import collections
import json
from pathlib import Path
import sys
H = Path(__file__).resolve().parent
ledger = json.loads((H / 'contracts.json').read_text())
traces = {r['id']: r for r in json.loads(Path(sys.argv[1]).read_text())}
manifest = json.loads((H / 'evidence/origin-review.json').read_text())
assert len(manifest) == len(traces) == 434
for review in manifest:
    row = next(r for r in ledger['rows'] if r['id'] == review['id'])
    assert all(row[k] == review[k] for k in ('file', 'line', 'column'))
    indexed = review['adaptation'] == 30
    row.update(origin_review_cohort='original-434', origin_class=review['origin_class'],
               origin_trace=traces[row['id']], origin_review=review['review'],
               classification_status='value-origin-reviewed', root_cause=review['review'],
               scope='adaptation-30-index' if indexed else 'included',
               smallest_edit=None, javascript_effect='unresolved', edit_status='not-established',
               why_unfinished='Routed to adaptation 30; no edit proposed in this unit.' if indexed else
               'Class and value origin recorded; no verified type-only edit proposed.')
    row['class'] = review['origin_class']
classes = ['undefined from an indexed read', 'an optional field passed to a required parameter',
           'a union without a discriminant', 'a generic inference shape', 'other']
counts = collections.Counter(r['origin_class'] for r in ledger['rows'] if r.get('origin_review_cohort'))
ledger['origin_review'] = dict(population=434, counts={k: counts[k] for k in classes},
                             status='classified; edits unresolved')
for field, key in [('scope_counts', 'scope'), ('class_counts', 'class')]:
    ledger[field] = dict(collections.Counter(r[key] for r in ledger['rows']))
(H / 'contracts.json').write_text(json.dumps(ledger, indent=2) + '\n')
old = (H / 'contracts.md').read_text()
observations = old[old.index('## Observations and reproduction'):old.index('## Candidate counts and scope')]
text = '# Remaining checker contracts and TS2345 value origins\n\n'
text += 'All 921 candidate diagnostics retain their full chains and locations. The original 434 pending TS2345 arguments now have value-origin classifications and source traces. No source edits were made. Classification does not establish a safe edit. The other diagnostic buckets retain the earlier incomplete contract review.\n\n'
text += '## Original 434 TS2345 origin counts\n\n| Class | Count |\n|---|---:|\n'
text += ''.join(f'| {k} | {counts[k]} |\n' for k in classes)
text += '\nThe 377 indexed origins route to adaptation 30, alongside the previously identified 298 direct indexed arguments. The 43 generic inference shapes are collection iterator/key/value and callback element shapes, not proven generic factory specializations. Optional fields written inside whole object arguments belong to other; the optional-field class is reserved for an actual field read passed as an argument. Existing discriminants are recorded rather than counted as missing. No new edits are proposed. The two previous proposals remain unaccepted, requiring consumer/checker review; whole-file TypeScript 6.0.3 ES2024/ESNext transpilation produced byte-identical JavaScript for both in-memory proposals (evidence/erasure.log). This excludes source maps and does not prove checker acceptance or consumer contracts.\n\n'
text += observations
text += '''## Reproduction and checks

The raw census is retained in evidence/census.jsonl.gz. Follow the baseline extraction and AST inspection in build_contracts.py, then finish the classification:

```sh
NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules node stage3/triage/trace_origins.cjs /tmp/stage3-contracts-tree stage3/triage/contracts.json /tmp/contracts-origins.json > /tmp/contracts-origins.log 2>&1
python3 stage3/triage/classify_origins.py /tmp/contracts-origins.json > /tmp/contracts-classification.log 2>&1
python3 stage3/triage/audit_contracts.py /tmp/stage3-contracts-census.jsonl stage3/triage/contracts.json > /tmp/contracts-audit.log 2>&1
```

The tracing pass records symbol initializers, assignments, destructuring, callback collections and function returns. It is a navigation aid, not a proof of runtime density or path reachability. Cycles and depth limits remain visible; evidence/origin-review.json supplies reviewed classification and additional source chains for helper parameters and regex captures. Each row below lists that review and every recorded source step. The audit checks all 921 census chains and the exact original 434 review locations and class counts. A dropped-row mutant with recomputed counts fails, as does a reclassified-row mutant with recomputed counts. Logs are retained in evidence/.

The adapted source hashes were checked against the previous census tree. No new apply/census was needed because that tree is unchanged. No full gate or oracle suite was run. No code-touching edit, unwrap, cast or runtime guard is proposed. The rest of the contract review remains unresolved because source origins do not prove those changes preserve behavior.

## Rows grouped for follow-up

'''
for group in sorted(ledger['class_counts']):
    rows = [r for r in ledger['rows'] if r['class'] == group]
    text += f'### {group} ({len(rows)} candidates)\n\n'
    for r in rows:
        text += f'#### {r["id"]}: {r["file"]}:{r["line"]}:{r["column"]}, TS{r["code"]}\n\nScope: {r["scope"]}. Classification: {r["classification_status"]}.\n\n```text\n{r["chain"]}\n```\n\n'
        text += f'```typescript\n{r.get("context", "")}\n```\n\nCause: {r["root_cause"] or "Not established."}\n\n'
        if r.get('origin_trace'):
            text += 'Value trace (branches are alternatives, not a sequential execution):\n\n'
            for step in r['origin_trace']['steps']:
                snippet = step['text'].splitlines()[0][:160].replace('`', "'")
                text += f'- {step["via"]}: {step["file"]}:{step["line"]}:{step["column"]}, `{snippet}`.\n'
            text += '\nTerminal observations: ' + ', '.join(sorted({t['reason'] for t in r['origin_trace']['terminals']})) + '.\n\n'
        text += f'Smallest edit: {r["smallest_edit"] or "Unresolved; no edit proposed."}\n\nJavaScript: {r["javascript_effect"]}.\n\n'
        if r.get('why_unfinished'): text += r['why_unfinished'] + '\n\n'
(H / 'contracts.md').write_text(text.rstrip() + '\n')
print(json.dumps(ledger['origin_review']))
