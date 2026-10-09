#!/usr/bin/env python3
"""Capture the complete candidate population without claiming unfinished proofs."""
import collections
import gzip
import hashlib
import json
from pathlib import Path
import subprocess
import sys
from audit_contracts import candidates

HERE = Path(__file__).resolve().parent
census, tree, metadata = map(Path, sys.argv[1:])
if metadata.exists():
    rows = json.loads(metadata.read_text())
else:
    rows = candidates(census)
    metadata.write_text(json.dumps(rows))
    print('candidate input written; run inspect_contracts.cjs, then run this script with its output')
    sys.exit(0)
counts = dict(sorted(collections.Counter(str(r['code']) for r in rows).items()))
classes = ['a union property access without a discriminant', 'an iterator element contract',
           'an overload choice', 'a catch variable of type unknown',
           'a generic factory specialization', 'another']
for index, row in enumerate(rows, 1):
    row['id'] = f'C{index:04d}'
    row['class'] = 'another'
    row['classification_status'] = 'unreviewed'
    row['scope'] = 'included' if row['code'] != 2345 else 'pending-index-origin-review'
    row['root_cause'] = None
    row['smallest_edit'] = None
    row['javascript_effect'] = 'unresolved'
    row['edit_status'] = 'not-established'
    row['why_unfinished'] = 'The full diagnostic and source context are recorded; a behavior-identical edit and its contract proof have not been established.'
    if row['code'] == 2488:
        row.update({'class': 'an iterator element contract', 'classification_status': 'diagnostic-chain',
                    'root_cause': 'The destructured iterator-derived value includes undefined; the tuple itself is iterable.'})
    if row['code'] == 2769:
        row.update({'class': 'an overload choice', 'classification_status': 'diagnostic-chain',
                    'root_cause': 'No declared overload accepts this argument; the final-overload chain is preserved below. Root owner review is still required.'})
    if row['code'] == 18046:
        if row['file'].endswith('transformers/jsx.ts'):
            row.update({'class': 'an iterator element contract', 'classification_status': 'source-reviewed',
                        'root_cause': 's is the callback element of arrayFrom(importSpecifiersMap.values()); it is not a catch binding. Earlier map-entry destructuring loses the element contract.'})
        else:
            row.update({'class': 'a catch variable of type unknown', 'classification_status': 'source-reviewed',
                        'root_cause': 'The catch binding e is unknown, but the original code directly reads message or code.'})
            row['why_unfinished'] = 'An instanceof Error check is insufficient for Node errno.code and arbitrary host throws. String conversion or a fallback changes the original property-read behavior. No verified narrowing that preserves all original error paths was established.'
    if row['code'] == 2538:
        row.update({'classification_status': 'diagnostic-chain',
                    'root_cause': 'The lookup key may be undefined; prove its origin and bounds before changing the indexed lookup.'})
    if row['code'] == 2339:
        if "| undefined'" in row['chain']:
            row.update({'classification_status': 'diagnostic-chain',
                        'root_cause': 'A possibly absent value is destructured or accessed. This is not evidence of a union missing a discriminant.'})
        else:
            row['root_cause'] = 'The receiver remains a union at the property read. Several of these sites already test kind; indexed absence can prevent narrowing. Discriminant absence is not established.'
    if row['code'] == 2345 and row.get('argument', {}).get('kind') == 'ElementAccessExpression':
        # An arbitrary-specialization ending is also an absence chain when T|undefined is passed to T.
        absent = ("Type 'undefined' is not assignable" in row['chain'] or
                  ("| undefined' is not assignable to parameter of type" in row['chain'] and
                   'could be instantiated' in row['chain']) or
                  "Argument of type 'undefined' is not assignable" in row['chain'])
        if absent:
            row.update({'scope': 'excluded-direct-index', 'classification_status': 'AST-and-chain',
                        'root_cause': 'The diagnosed argument is directly an element access, and its expanded mismatch is undefined-from-an-index.',
                        'why_unfinished': 'Excluded from requested TS2345 scope, but retained so the census partition cannot silently lose a row.'})
    if row['code'] == 2345 and not ("Type 'undefined' is not assignable" in row['chain'] or
            "Argument of type 'undefined' is not assignable" in row['chain'] or
            ('could be instantiated' in row['chain'] and '| undefined' in row['chain'])):
        row['scope'] = 'included'
    if row['file'].endswith('utilities.ts') and row['line'] == 11201 and row['code'] == 2345:
        row.update({'scope': 'included', 'class': 'an overload choice', 'classification_status': 'source-reviewed',
                    'root_cause': 'Saving overloaded String.prototype.replace and calling its .call selects the last Symbol.replace overload; the actual search argument is a string.',
                    'smallest_edit': 'Declare const stringReplace: (this: string, searchValue: string, replaceValue: string) => string = String.prototype.replace; Leave stringReplace.call(s, "*", replacement) unchanged.',
                    'javascript_effect': 'type-only (byte-identical JavaScript after erasure)', 'edit_status': 'proposed',
                    'why_unfinished': None})
    if row['file'].endswith('sys.ts') and row['code'] == 2322 and "Type 'unknown'" in row['chain']:
        row.update({'class': 'a catch variable of type unknown', 'classification_status': 'source-reviewed',
                    'root_cause': 'nodeSystem.require returns the caught unknown value unchanged, while ModuleImportResult promises an object with stack/message.',
                    'smallest_edit': 'In types.ts ModuleImportResult failure member, change error: { stack?: string; message?: string; } to error: unknown. Retain the returned error value. Recheck every consumer before accepting this owner edit.',
                    'javascript_effect': 'type-only (byte-identical JavaScript after erasure)', 'edit_status': 'proposed',
                    'why_unfinished': None})

ledger = dict(status='incomplete', merged_tree_commit=subprocess.check_output(['git', 'rev-parse', 'HEAD'], text=True).strip(),
              adaptation_commits=['779f303d929c8bf9dbbac54130358e21310b8618', 'e3535e702e285a5dcb4d06364c4f889b8178ea0c'],
              census_sha256=hashlib.sha256(census.read_bytes()).hexdigest(),
              census='evidence/census.jsonl.gz', source_tree=str(tree), candidate_counts=counts,
              scope_counts=dict(collections.Counter(r['scope'] for r in rows)),
              class_counts=dict(collections.Counter(r['class'] for r in rows)),
              proposed_edits=sum(r['edit_status'] == 'proposed' for r in rows), rows=rows)
(HERE / 'contracts.json').write_text(json.dumps(ledger, indent=2) + '\n')
text = '''# Remaining checker contracts: incomplete review

This is a complete **candidate inventory**, not a completed adaptation ledger.
All seven requested codes, including every TS2345 candidate, are captured with
full expanded chains and exact adapted-tree UTF-16 locations. Two source-reviewed
type-only edits are proposed. Other edits are explicitly unresolved. Group labels
with diagnostic-chain or unreviewed status are routing information, not completed
root-cause proofs. In particular, no generic factory specialization has yet been
confirmed; a generic T fed by an indexed read is not enough to assign that class.

The TS2345 partition conservatively excludes only direct element-access arguments
whose full chain establishes missing undefined. Indirect index origins and
iterator-propagated absence still need review. Those rows stay visible. Thus the
included count is not a final non-index TS2345 count. No source edits were made.

## Observations and reproduction

The tree starts at origin/main ef3d907, merges adaptation 10 at 779f303 and
adaptation 20 at e3535e7, and is measured before any upstream build regeneration.
Apply generated diagnostics and changed 73 files, 4,125 lines replaced overall:
10 changed 72 files/3,719 lines; 20 changed 26 files/406 lines. Apply's incidental
stage3/patch-set.md write is restored, since this unit owns only stage3/triage/.

Setup succeeded: Go 1.27.1, clang 20.1.8, Node 24.19.0. Go, clang, Node and
submodules ready at 0s; build cache warm and total 118s; nproc 5; CPU quota 4.
Setup warms test binaries without running tests. The census completed with 82
records: 81 per-file attempts and the final 78-root program, 2,164 whole-program
diagnostics, 143.924 seconds summed gate time. Its extra TS2307 is the missing
source-map-support dependency; no upstream dependencies were installed. No full gate or TypeScript
oracle suite was run: this unit makes no source adaptation.

Commands (all run output was redirected to the matching evidence log):

```sh
bash cloud/setup.sh > /tmp/stage3-contracts-setup.log 2>&1
source /workspace/adamic-tools/env.sh
stage3/apply.sh /tmp/stage3-contracts-tree > /tmp/stage3-contracts-apply.log 2>&1
go build -o /tmp/stage3-contracts-census ./stage3/census/tool > /tmp/stage3-contracts-build.log 2>&1
/tmp/stage3-contracts-census /tmp/stage3-contracts-tree/src/compiler /tmp/stage3-contracts-census.jsonl > /tmp/stage3-contracts-census.log 2>&1
python3 stage3/triage/audit_contracts.py /tmp/stage3-contracts-census.jsonl stage3/triage/contracts.json > /tmp/stage3-contracts-audit.log 2>&1
```

The final raw record is the whole program. Per-entry duplicate diagnostics are
retained in compressed census evidence but do not inflate row counts. The stock
6.0.3 AST inspector supplies source and argument-origin evidence; its argument
types are labeled stock and do not replace Adamic's checker measurement. The
census loader itself formats the complete message chain recursively.

## Candidate counts and scope

'''
text += '| Code | Candidate rows |\n|---|---:|\n' + ''.join(f'| TS{code} | {n} |\n' for code, n in counts.items())
text += '\nScope partition: ' + json.dumps(ledger['scope_counts'], sort_keys=True) + '.\n'
text += '''
## Rebuilding these artifacts

Source excerpts are from TypeScript 6.0.3, Apache-2.0, commit
050880ce59e30b356b686bd3144efe24f875ebc8. The source hash map is in evidence/.
To regenerate from the retained census and the same adapted source tree, use
fresh scratch output paths for the candidate and inspection files:

```sh
gzip -dc stage3/triage/evidence/census.jsonl.gz > /tmp/contracts-census.jsonl
python3 stage3/triage/build_contracts.py /tmp/contracts-census.jsonl /tmp/stage3-contracts-tree /tmp/contracts-candidates.json > /tmp/contracts-extract.log 2>&1
NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules node stage3/triage/inspect_contracts.cjs /tmp/stage3-contracts-tree /tmp/contracts-candidates.json /tmp/contracts-inspected.json > /tmp/contracts-inspect.log 2>&1
python3 stage3/triage/build_contracts.py /tmp/contracts-census.jsonl /tmp/stage3-contracts-tree /tmp/contracts-inspected.json > /tmp/contracts-build.log 2>&1
python3 stage3/triage/audit_contracts.py /tmp/contracts-census.jsonl stage3/triage/contracts.json > /tmp/contracts-audit.log 2>&1
```

The first generator invocation writes candidates if its input does not yet exist;
the last consumes AST-inspected rows and rewrites only this unit's two ledgers.
Adjust NODE_PATH to the stock 6.0.3 tooling installed by apply in your cache.
The census hash intentionally pins exact retained raw bytes, including scratch
paths and per-entry durations. A fresh census gives a new hash and new ledgers.

## Coverage audit and mutant

`audit_contracts.py` compares the complete multiset of (file, line, column, code,
full chain) against the final census record, then checks stored counts, unique
IDs and census hash. Excluded and pending TS2345 rows participate too. It has no
hardcoded expected population that can accidentally hide newly exposed rows.
See evidence/audit.log for the green run and evidence/mutant.log for a separate
copy with one row removed. The mutant command must exit nonzero. This proves row
coverage only; it does not prove the unfinished classification or edit semantics.

## What remains unfinished and why

I did not establish a smallest behavior-identical edit for every row, complete
the indirect TS2345 index-origin partition, or prove all class assignments.
Diagnostic chains establish incompatibility, not array density, cache presence,
host exception shapes or generic specialization invariants. Checked unwraps can
change the original thrown error; assertions can erase a real type lie. Claiming
that either is equivalent from these observations would be unsupported. The
source contexts and argument declarations below make that remaining review
concrete. The two proposed type-only edits have not been applied, checker-tested,
or checked against every affected consumer; their JavaScript classification
follows from erasing their type annotations, not a native-output experiment.

## Rows grouped for follow-up

'''
for group in classes:
    subset = [r for r in rows if r['class'] == group]
    text += f'\n### {group} ({len(subset)} candidates)\n\n'
    if not subset:
        text += 'No confirmed rows; unfinished review must not be read as a zero census count.\n'
    for row in subset:
        text += f'#### {row["id"]}: {row["file"]}:{row["line"]}:{row["column"]}, TS{row["code"]}\n\n'
        text += f'Scope: {row["scope"]}. Classification: {row["classification_status"]}.\n\n'
        text += '```text\n' + row['chain'] + '\n```\n\n'
        text += '```typescript\n' + row.get('context', '') + '\n```\n\n'
        text += 'Cause: ' + (row['root_cause'] or 'Not established; source and full chain retained for review.') + '\n\n'
        text += 'Smallest edit: ' + (row['smallest_edit'] or 'Not established.') + '\n\n'
        text += 'JavaScript: ' + row['javascript_effect'] + '.\n\n'
        if row.get('why_unfinished'):
            text += row['why_unfinished'] + '\n\n'
(HERE / 'contracts.md').write_text(text.rstrip() + '\n')
print(json.dumps({k: ledger[k] for k in ['status', 'candidate_counts', 'scope_counts', 'proposed_edits']}))
