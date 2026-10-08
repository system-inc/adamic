#!/usr/bin/env python3
"""Reconcile observed witness logs with the assigned ledger, without counting skips."""
import json
import pathlib
import re
import subprocess

root = pathlib.Path(__file__).resolve().parent
sites = json.loads((root / 'sites.json').read_text())
logs = sorted(pathlib.Path('/tmp').glob('stricter-indexed-b-group*-final.log'))
proven = set()
blocked = {}
for log in logs:
    text = log.read_text()
    proven.update(re.findall(r'PROVEN (D\d+):', text))
    for site, reason in re.findall(r'BLOCKED (D\d+): ([^\n]+)', text):
        blocked[site] = re.sub(r'/tmp/[^ ]+/main.ts:', 'main.ts:', reason)
remaining = {site['id'] for site in sites} - proven - blocked.keys()
sha = subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=root, text=True).strip()
lines = [
    f'Built 27 isolated project-.ts fixture templates and a Node/native/JavaScript witness harness; {len(proven)} sites proven.',
    f'Base 390af985; latest preceding commit {sha}; branch codex/stricter-indexed-b only.',
    f'Logged filtered go tests: {len(proven)} proven, {len(blocked)} blocked, {len(remaining)} remaining; final gate recorded below.',
    f'Every proven site has an emitted-C erase-panic mutant that builds under sanitizers and loses the pinned exit-70 observation.',
    'Whole-program compilation, sparse holes, typed arrays and records are not covered; refused shapes are never counted as proven.',
    '',
    '## Scope and assumptions',
    '',
    'The exact 27 rows D195 through D221 are retained in sites.json from a1a16427',
    'stage3/ledger/checker-259/rows.csv. All receivers are arrays. Read shapes were',
    'traced through the referenced cohere/TypeScript submodule source at',
    'tsc/testdata/fixtures/compiler/transformers/generators.ts; no implementation',
    'was copied from cohere. The fixtures are independently written minimal shapes.',
    'Statement and declaration arrays use small object elements; clause and CodeBlock',
    'arrays preserve discriminated object unions. readonly array receivers are preserved.',
    'Label is number; labelNumbers is number[][]; blockOffsets is number[].',
    '',
    'D196/D197 share variables[i]. D200 uses one caseBlock.clauses[i] read;',
    'D201-D205 and D207/D208 share another such read. D213/D214 share blockStack[i].',
    'These are ledger diagnostic rows, not a claim of 27 unique upstream reads.',
    'Each row has an isolated witness so a prior trap cannot hide its observation.',
    'Minimal witnesses stop at the read, before downstream property uses, allowing',
    'source Node to observe undefined rather than throw on a later dereference.',
    '',
    'The specific .ts witness requirement supersedes the general new-.a convention.',
    'Following the base worker, tracked .ts.txt templates materialize as main.ts in',
    'a temporary project with its own tsconfig. noUncheckedIndexedAccess is omitted',
    '(project default false); strictNullChecks and strictFunctionTypes stay enabled.',
    'Each source is run directly through Node 24 type stripping, outside Adamic.',
    '',
    'Present input prints present; an empty receiver with the same numeric index prints',
    'undefined on source Node. Both have empty stderr and exit 0. Native release and',
    'ASan/UBSan and the JavaScript backend must match present input. For absent input',
    'all backends must produce empty stdout, exit 70, and exactly',
    '`adamic: panic: indexed read is absent: <main.ts>:<read line>:<read column>\\n`.',
    'The expected location is calculated independently from source text, and the',
    'same location must appear in the inserted guard. The actual --explain-checks CLI',
    'must list exactly one checked indexed-presence site, count one, and trust zero.',
    '',
    '## Site results',
    '',
    '| Row | Upstream line | Source expression | Receiver | Result |',
    '|---|---:|---|---|---|',
]
for site in sites:
    sid = site['id']
    n = int(sid[1:])
    kind = ('number[]' if n in (206,209,210,220) else 'number[][]' if n == 219
            else 'object union array' if n in list(range(200,206))+[207,208,211]+list(range(212,218))+[221]
            else 'object array')
    result = 'proven, erase-panic caught' if sid in proven else ('blocked: '+blocked[sid] if sid in blocked else 'remaining')
    lines.append(f'| {sid} | {site["line"]} | `{site["source_expression"]}` | {kind} | {result} |')
lines += ['', '## Commands, setup and limits', '',
          'All test output is written to log files, never piped. Current proof logs:', '']
lines += [f'- `{log}`' for log in logs]
setup = pathlib.Path('/tmp/stricter-indexed-b-setup.log').read_text()
lines += ['', 'Setup: `export GOPROXY=\'https://proxy.golang.org|direct\'; bash cloud/setup.sh`',
          'with `/workspace/adamic-tools/env.sh` sourced for builds and tests. nproc=5.', '', '```text']
lines += [line for line in setup.splitlines() if line.startswith('setup:')]
lines += ['```', '',
          'The first test attempt preceded submodule checkout and failed with missing',
          'cohere/TypeScript/tsc/go.mod. It was rerun after checkout completed.',
          'No compiler implementation files are edited. Unsupported receivers are',
          'reported with the observed refusal; this unit does not extend them.',
          'The full repository gate is not claimed. The final focused gate and vet',
          'results are appended after the last group.']
(root / 'REPORT.md').write_text('\n'.join(lines)+'\n')
print(f'proven={len(proven)} blocked={len(blocked)} remaining={len(remaining)}')
