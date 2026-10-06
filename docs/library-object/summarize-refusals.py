"""Join real-tsc observations and classify the runner's first refusal, without guessing blockers."""
import collections
import gzip
import hashlib
import html
import json
from pathlib import Path
import re
import sys

root = Path(__file__).resolve().parent
captures = [json.loads(line) for line in Path(sys.argv[1]).read_text().splitlines()]
observations = {}
for name in sys.argv[2:]:
    for line in Path(name).read_text().splitlines():
        row = json.loads(line)
        observations[row['path']] = row
programs = {item['result']['path']: item['program'] for item in captures}
wanted = {item['result']['path'] for item in captures if item['result']['kind'] == 'refused'}
assert set(observations) == wanted
for row in observations.values():
    if row['tsc']:
        assert row['sameCode'], row
        row['bucket'] = 'a'
        row['feature'] = re.search(r'error TS\d+', row['reason'])[0]
    else:
        method = re.search(r'(?:not yet: |refuses )(Object\.[\w]+)', row['reason'])
        feature = method[1] if method else None
        # A builtin constructor or builtin value is a library dependency, even when the first
        # diagnostic only names the expression form rather than its identifier.
        builtin = re.search(r'not yet: (?:reading |)(Object|Array|String|Number|Boolean|Date|RegExp|Error|EvalError|RangeError|ReferenceError|SyntaxError|TypeError|URIError|Math|JSON|Function)(?: as a value|$)', row['reason'])
        if builtin:
            feature = builtin[1] + ' builtin value'
        if row['reason'] == 'not yet: Number.prototype':
            feature = 'Number.prototype builtin value'
        if row['reason'] == 'refuses isPrototypeOf':
            feature = 'Object.prototype.isPrototypeOf'
        if row['reason'] == 'not yet: new an Identifier':
            body = programs[row['path']][programs[row['path']].rfind('// Copyright'):]
            constructor = re.search(r'\bnew\s+(String|Number|Boolean|Date|SyntaxError|Object)\b', body)
            assert constructor, row
            feature = constructor[1] + ' constructor'
        row['bucket'] = 'b' if feature else 'c'
        row['feature'] = feature if feature else row['reason']
    row['sourceSha256'] = hashlib.sha256(programs[row['path']].encode()).hexdigest()
rows = sorted(observations.values(), key=lambda row: row['path'])
(root / 'classification.json').write_text(json.dumps(rows, indent=2) + '\n')
with gzip.open(root / 'adapted-sources.jsonl.gz', 'wb') as output:
    for item in captures:
        output.write((json.dumps(item) + '\n').encode())
counts = collections.Counter(row['bucket'] for row in rows)
text = '''# Fresh Object refusal classification

The initial classification was committed and pushed before implementation as `c48709e`, from main `5d4c801`, with real `typescript@6.0.3` and test262 `5992dc3b60faf62a48fd6be8a40ae9d9a8c84d81`. Adaptation is enabled. The old 0-pass measurement predates existing Object support.

| Measurement | Pass | Disagreement | Refused | Crashed | Skipped | Total |
|---|---:|---:|---:|---:|---:|---:|
| Fresh before | 17 | 1 | 2317 | 0 | 1076 | 3411 |

## Every refusal checked

'''
text += '| Bucket | Count | Meaning |\n|---|---:|---|\n'
for bucket, meaning in [('a','Real tsc rejects with the same TS diagnostic code'),('b','Real tsc accepts; first refusal names a library feature or dependency'),('c','Real tsc accepts; first refusal names a language or representation feature')]:
    text += f'| {bucket} | {counts[bucket]} | {meaning} |\n'
text += '''
No tsc rejection had a different code. A match means the runner's first normalized TS code appears among real tsc's diagnostics, not that every diagnostic or its order is identical. Every refused adapted program was checked separately against `lib.es2024.d.ts` and the exact `internal/load/prelude.d.ts`, and Adamic's pinned declaration files from `cohere/TypeScript/tsc/internal/bundled/libs` with every `regexp_library.go` correction, with all `compilerOptions()` from `internal/load/load.go`. The checker implementation is the npm TypeScript package, not typescript-go. Skipped tests are outside the refusal classification.

[classification.json](classification.json) lists every refused test, the full real-tsc diagnostics, feature, bucket and source hash. [adapted-sources.jsonl.gz](adapted-sources.jsonl.gz) preserves the runner's exact sources and per-test outcomes. [classify-tsc.cjs](classify-tsc.cjs) and [summarize-refusals.py](summarize-refusals.py) reproduce the tsc pass and tables. The scratch runner was a copy of `cmd/adamic-test262`, with only a capture after each result. The official runner's aggregate result and the capture run match. No runner source was edited.

Counts describe the first blocker observed. Closing it may expose another language or library blocker, so bucket (b) counts are candidate tests, not promised passes. Bucket (c) includes deliberately unsupported language features such as any, delete and detached methods, not just implementation work owed.
'''
for bucket in 'abc':
    reasons = collections.Counter(row['feature'] for row in rows if row['bucket'] == bucket)
    text += f'\n## Bucket ({bucket}) reasons\n\n| Reason | Count |\n|---|---:|\n'
    for reason, count in reasons.most_common():
        text += f'| `{reason.replace("|", " / ")}` | {count} |\n'
text += '''
## Classification refinement

The first claim used npm's es2024 declarations and classified generic builtin constructor/value diagnostics as representation gaps: (a) 1679, (b) 388, (c) 250. The final real-tsc pass uses the checker's exact pinned declaration text and RegExp corrections. Constructor/value expressions were inspected and named as library dependencies rather than generic language blockers; the final tables reflect that refinement. The accepted/rejected partition and matching diagnostic codes are checked again, not inferred. These dependencies may belong to other library slices and do not authorize edits there.

## Claim and build order

1. The largest library refusal, `Object.defineProperty` (185), and `defineProperties` (24) require observable descriptors, accessors and shape mutation. These remain explicit refusals: CLAUDE.md binds the approved fixed-shape memory and type contract, and docs/0.1.md forbids this mutation. Implementing them needs a representation and soundness design beyond this slice. Prototype mutation has the same constraint. Descriptor reads currently return any, adding a language blocker even if their first refusal were removed.
2. Build the largest remaining family: `Object.seal` (38), `Object.isExtensible` (35), `Object.isSealed` (28), then `Object.preventExtensions` (9). Port V8's integrity tests for represented fixed plain objects and its primitive fast paths. Unknown host constructors, arrays needing integrity mutation, descriptor/prototype mutation, and unrepresented values remain refused. Existing Object.freeze must interoperate with the integrity queries.
3. Investigate own-name reflection (16) and exact-shape keys (14) after integrity, retaining explicit refusals whenever the shape or presence is unproven.
4. Close the observed baseline disagreement `Object.is(undefined, null)`: the existing null/undefined representation cannot be silently compared as equal. Refuse ambiguous comparisons unless this slice can prove the exact result.

## Language handoff for @system_adamic

Each row is a representative adapted test body on one line. To reproduce, prepend the unchanged runner prelude from `cmd/adamic-test262/prelude.go`. The full per-test sources and diagnostics are in the linked artifacts; no language changes are claimed here.

| First language/representation refusal | Tests | One-line reproducer | Representative test |
|---|---:|---|---|
'''
reasons = collections.Counter(row['feature'] for row in rows if row['bucket'] == 'c')
for reason, count in reasons.most_common():
    row = next(row for row in rows if row['feature'] == reason and row['bucket'] == 'c')
    body = programs[row['path']]
    # Prelude ends immediately before the test copyright; use the exact body, not a guessed reduction.
    body = body[body.find('// Copyright'): ] if '// Copyright' in body else body[body.find('/* Copyright'):]
    body = re.sub(r'/\*.*?\*/', '', body, flags=re.S)
    body = re.sub(r'//[^\n]*', '', body)
    body = html.escape(' '.join(body.split())).replace('|', '&#124;').replace('`', '&#96;')
    text += f'| {reason.replace("|", " / ")} | {count} | <code>{body}</code> | `{row["path"]}` |\n'
(root / 'refusals.md').write_text(text)
print(dict(counts))
