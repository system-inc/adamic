#!/usr/bin/env python3
"""Render a complete, disjoint read/write handoff from checker-bound evidence."""
import collections
import json
from pathlib import Path
import sys
report = json.loads(Path(sys.argv[1]).read_text())
out = Path(sys.argv[2])
rows = report['rows']
assert len(rows) == 391
assert len({(r['File'], r['Line'], r['Column'], r['Kind'], r['Text']) for r in rows}) == 391
assert report['counts'] == {'total':391,'read_only':sum(r['classification']=='read-only' for r in rows),'write':sum(r['classification']=='write' for r in rows)}
reads = [r for r in rows if r['classification'] == 'read-only']
writes = [r for r in rows if r['classification'] == 'write']
assert len(reads) + len(writes) == 391
assert all(r['writes'] for r in writes)
assert not any(r['writes'] for r in reads)
def key(w): return w['where'], w['access'], w['property'], w['written']
operations = {}
for r in writes:
    for w in r['writes']: operations.setdefault(key(w), w)
ids = {k: f'W{i:03}' for i, k in enumerate(sorted(operations), 1)}
def cell(value):
    return str(value).replace('|', '&#124;').replace('\r', '').replace('\n', '<br>').replace('`', '&#96;')
lines = [
    f'Read-only: **{len(reads)} sites** wait for the checked view, with no new adaptation.',
    f'Write-reaching: **{len(writes)} sites** go to @system_adamic; the complete list follows.',
    'Partition: all **391** remaining census locations, counted once each.',
    'Method: stock TypeScript 6.0.3 checker symbols plus an alias-flow fixed point, not source-text matching.',
    'Existing adaptation 75 stays in place; full-apply validation waits for the announced relocation.',
    '',
    '# Read or write handoff', '',
    'This supersedes the owner-family grouping request and the earlier decline recommendations.',
    'It applies the October 7 04:35 ruling: a wider-view read checks absence or assignability',
    'at the read and panics on an incompatible present value; a wider-view write needs',
    'the source declaration. This unit adds no further source adaptations.',
    '',
    '## What the counts mean', '',
    'These are static, conservative **may-write** counts for the pinned compiler project,',
    'not counts of assignments that execute on every run. A site goes in the write bucket',
    'if its widened value, a returned value, or a callback parameter can reach a matching',
    'write through aliases. Branches and calling contexts are joined, so a write witness',
    'may require path-sensitive review by @system_adamic. No site is counted in both buckets.',
    '',
    'The census records only its first missing property. The analysis also checks other',
    'missing optional members of that same bound target, so a read of the first property',
    'does not hide a write to another missing member. The witness states which member is',
    'written. Writes to a field already declared on the non-nullable source are excluded',
    'from those additional-member checks. Seven sites write a different missing member',
    'than their first census property and are therefore retained in the write bucket.',
    '',
    'Some census sources print as `never`, while the stock checker resolves the expression',
    'to an inhabited type with the offending member already declared. They are **not**',
    'automatically classified as dead code. The census source/target are retained verbatim;',
    'the bound stock source and member declarations are in the machine evidence. These',
    'disagreements need compiler review rather than another declaration adaptation.',
    '',
    '## Symbol and alias method', '',
    'The analyzer loads the actual compiler config path, including Node declarations, and',
    'retains strict, exactOptionalPropertyTypes, noUncheckedIndexedAccess,',
    'verbatimModuleSyntax, with erasableSyntaxOnly false and no house-style return or fallthrough options. It examines diagnosed bodies too.',
    'It locates each inventory AST node by file, position and kind, then resolves the',
    'source, contextual target and target member with the checker. Variable, parameter,',
    'import and member identities come from checker symbols and getRootSymbols.',
    '',
    'Forward edges cover variable initialization, assignments, destructuring, casts,',
    'conditional values, compiler-call arguments/results, member stores/loads, array',
    'elements, standard collection storage and callback implementations. Known function',
    'identities are propagated to a fixed point, including callbacks invoked through',
    'an aliased variable. Member writes are projected into the receiver type and matched',
    'by checker root-symbol identity; unrelated same-name members do not suffice.',
    'Computed keys use checker literal/union/string types. A broad string key that can',
    'name a missing member is conservatively a possible write.',
    '',
    'The write forms include assignment, compound assignment, increment/decrement,',
    'deletion and resolved Object.assign/Object.defineProperty calls. Array copies are',
    'distinguished from transforming map callbacks; method receivers propagate to this.',
    'This is a flow-insensitive compiler-source analysis. It does not prove branch',
    'feasibility or arbitrary effects of external user callbacks, reflective code or',
    'native dependencies. Read-only means no matching write in this analyzed compiler',
    'graph. It is a compiler handoff inventory, not a whole-program native soundness proof.',
    '',
    '## Validation and reproduction', '',
    'The alias fixture has eight widening sites: seven writes and one read-only site.',
    'It catches `c.p = 2` through three variable aliases and `alias.p = 3` through',
    'an aliased callback. A different receiver with its own `p` symbol does not pollute',
    'the read-only site. Replacing the first alias initializer by a separate object',
    'changes the fixture to six writes and two reads, proving that the alias edge matters.',
    'Additional probes detect another missing member, a computed keyof write and',
    'an Object.assign write through an alias, an untyped alias write, and a write',
    'through the receiver returned by Object.assign.',
    'All probe commands exited 0. Evidence includes exact write member declarations.',
    '',
    '```sh',
    'NODE_PATH=STOCK_6_0_3_NODE_MODULES node stage3/adapt/75-optional-widening/read-write.cjs TREE stage3/adapt/75-optional-widening/evidence/wave2/remaining-sites.json result.json > analysis.log 2> analysis-errors.log',
    'NODE_PATH=STOCK_6_0_3_NODE_MODULES node stage3/adapt/75-optional-widening/probe-read-write.cjs > probes.json 2> probes.log',
    'python3 stage3/adapt/75-optional-widening/render-read-write.py result.json stage3/adapt/75-optional-widening/READ_WRITE.md > render.log 2>&1',
    '```', '',
    'TREE is the saved wave-2 sanctioned tree at upstream commit',
    '`050880ce59e30b356b686bd3144efe24f875ebc8`, with adaptation commits',
    '`bbb99fad` and `ecb85529`. The refusal inventory came from the scratch-only',
    'merge of `f1c9173`. This classification does not re-run or replace the earlier',
    'oracle. Full apply will be validated after the user confirms the slice-only',
    'directories have moved; no attempt to work around that pending integration is made.',
    '',
    'Machine evidence: [analysis](evidence/read-write/analysis.json.gz),',
    '[counts](evidence/read-write/counts.json), [probes](evidence/read-write/probes.json).',
    '',
    '## Complete write-site list for @system_adamic', '',
    f'**{len(writes)} sites; {len(operations)} distinct assignment witnesses.** Each W identifier',
    'links to its exact write location, receiver, assigned expression and checker type below.',
    'Sites retain the census line numbers on the wave-2 tree. Columns distinguish repeated',
    'relations on one line. A repeated witness is still listed against every site it reaches.',
    '',
    '| Site in src/compiler | Census source type | Census target type | First refused p | What is written |',
    '|---|---|---|---|---|',
]
for r in writes:
    refs = []
    for w in r['writes']:
        label = ids[key(w)]
        refs.append(f"[{label}](#{label.lower()}): {cell(w['property'])} {cell(w['operation'])} {cell(w['written'])}")
    lines.append(f"| {r['File']}:{r['Line']}:{r['Column']} | {cell(r['Source'])} | {cell(r['Target'])} | {cell(r['Property'])} | " + '<br>'.join(refs) + ' |')
lines += ['', '## Exact write witnesses', '']
for k in sorted(operations):
    w = operations[k]
    lines += [f"### {ids[k]}", '', f"Location: `{w['where']}`. Missing member: `{w['property']}`.",
              f"Receiver type: `{w['receiver_type']}`. Operation: `{w['operation']}`.",
              '', '```typescript', w['access'] + ' ' + w['operation'] + ' ' + w['written'], '```', '',
              'Assigned value type: `' + w.get('written_type', 'mutation of presence/value') + '`.',
              'Bound property declarations: ' + ', '.join('`' + x + '`' for x in w['property_declarations']) + '.', '']
lines += ['## Complete read-only list', '',
          f'**{len(reads)} sites.** Wait for the compiler checked view; no new owner edit proposed.', '',
          '| Site in src/compiler | Census source type | Census target type | First refused p |',
          '|---|---|---|---|']
for r in reads:
    lines.append(f"| {r['File']}:{r['Line']}:{r['Column']} | {cell(r['Source'])} | {cell(r['Target'])} | {cell(r['Property'])} |")
assert '\u2014' not in '\n'.join(lines), 'repository prose must not contain em-dashes'
out.write_text('\n'.join(lines) + '\n')
print(json.dumps({'total':len(rows),'read_only':len(reads),'write':len(writes),'distinct_write_witnesses':len(operations)}))
