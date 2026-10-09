"""Create a bounded load-path ledger from ledger.cjs's verified trace."""
import collections
import gzip
import json
from pathlib import Path
import sys

BUCKET = Path(__file__).resolve().parent
scratch = Path(sys.argv[1])
trace = json.loads((scratch / 'trace.json').read_text())
metadata = json.loads((scratch / 'metadata.json').read_text())
order = json.loads((scratch / 'order.json').read_text())
rank = {file: index + 1 for index, file in enumerate(order['order'])}
end = {event['file']: index for index, event in enumerate(trace['events']) if event['event'] == 'end'}

def original(statement):
    return statement and statement.startswith('src/compiler/') and '.generated.ts:' not in statement

observed = [r for r in trace['reads'] if original(r['statement'])]
assert all(r['providerEnded'] for r in observed)
# Preserve every occurrence, including generated input, without a large textual repetition.
with (BUCKET / 'trace.json.gz').open('wb') as raw:
    with gzip.GzipFile(fileobj=raw, mode='wb', mtime=0) as compressed:
        compressed.write(json.dumps(trace, separators=(',', ':')).encode())
(BUCKET / 'order.json').write_text(json.dumps(order, indent=2) + '\n')
seen = {r['location'] for r in trace['reads']}
unobserved = [m for m in metadata if original(m['directStatement']) and m['location'] not in seen]
rows = collections.OrderedDict()
for read in observed:
    statement = read['statement']
    row = rows.setdefault(statement, dict(statement=statement, bodyOrder=rank[statement.rsplit(':', 2)[0]],
                                         decision='accept by module order', bindings=collections.OrderedDict()))
    key = (read['binding'], read['provider'], read['declaration'])
    binding = row['bindings'].setdefault(key, dict(binding=read['binding'], provider=read['provider'],
         declaration=read['declaration'], providerBodyOrder=rank[read['provider']], providerEndEvent=end[read['provider']],
         firstReadEvent=read['sequence'], lastReadEvent=read['sequence'], occurrences=0, readLocations=[]))
    binding['occurrences'] += 1
    binding['lastReadEvent'] = read['sequence']
    if read['location'] not in binding['readLocations']:
        binding['readLocations'].append(read['location'])
    assert binding['providerEndEvent'] < binding['firstReadEvent']
for row in rows.values():
    row['bindings'] = list(row['bindings'].values())
    row['schedulingWitness'] = '05_safe_load_read/main.a'
# Static direct reads complement the exercised load path. Function bodies called at load
# are covered dynamically; this does not enumerate their unexecuted platform branches.
for read in unobserved:
    statement = read['directStatement']
    assert rank[read['provider']] < rank[statement.rsplit(':', 2)[0]]
    rows[statement] = dict(statement=statement, bodyOrder=rank[statement.rsplit(':', 2)[0]],
        decision='accept by module order; conditional read not observed', schedulingWitness='05_safe_load_read/main.a',
        bindings=[dict(binding=read['binding'], provider=read['provider'], declaration=read['declaration'],
                       providerBodyOrder=rank[read['provider']], readLocations=[read['location']])])
result = dict(method='TypeScript 6.0.3 compiler API plus instrumented Node ESM order and independent DFS',
    sourceCommit='050880ce59e30b356b686bd3144efe24f875ebc8',
    entry='src/tsc/tsc.ts', node='v24.19.0', platform='linux', argument='--version',
    originalStatementsObserved=len({r['statement'] for r in observed}), originalReadOccurrences=len(observed),
    originalReadSites=len({(r['statement'],r['location'],r['binding']) for r in observed}),
    readsBeforeProviderBodyBegins=sum(not r['providerBegun'] for r in trace['reads']),
    readsBeforeProviderBodyEnds=sum(not r['providerEnded'] for r in trace['reads']),
    staticDirectReadsNotObserved=len(unobserved), statements=list(rows.values()),
    limits=['Unexecuted branches inside functions called at load time are not certified.',
            'Windows/macOS, development hooks, optional packages and environment-dependent paths were not run.',
            'No cohere cycle rule was run. This is not an exhaustive all-platform hazard ledger.'])
(BUCKET / 'ledger.json').write_text(json.dumps(result, indent=2) + '\n')
lines = ['# Load-time imported-value ledger', '',
    'Method: stock TypeScript 6.0.3 compiler API, with imported value references instrumented before emit. '
    'Node evaluates the emitted ES modules from `src/tsc/tsc.ts`. A separate uninstrumented emit has the same '
    'dependency graph and output. Independent depth-first traversal matches all 80 observed module body starts. '
    'Cohere was not run.', '',
    'Observed on Linux, Node v24.19.0, `--version`: 57 original top-level statements, 914 imported-value reads, '
    '658 distinct `(originating statement, read location, binding)` sites. All providers finished before the '
    'reads. The generated diagnostic initializer is separate in `trace.json.gz`: 2,130 reads, all after '
    '`types.ts` finished. There are zero observed reads before provider begin or end.', '',
    'The API also identifies one direct conditional read not executed here: `sys.ts:1979` reads `Debug` '
    'when `sys` is absent. `debug.ts` runs third and `sys.ts` tenth, so that direct read is safe by order.', '',
    '**Bound:** this certifies the exercised module-load path and inventories direct imported reads in '
    'other top-level branches. It does not certify unexecuted branches inside functions invoked at load, '
    'other operating systems, development hooks or optional host packages. I did not produce an exhaustive '
    'all-platform ledger or propose a source adaptation on this evidence.', '',
    'Each row is accepted for cycle scheduling, represented by fixture 05; fixtures 09 and 10 separately '
    'cover copying imported function values and reading hoisted functions before provider body execution. '
    'This is a scheduling witness, not a claim to reproduce the full initializer in every row. Other '
    'features in those initializers remain their own workers\' scope.', '',
    '`ledger.json` gives each binding, its declaration and every observed read location, its provider end '
    'event, and the first/last read event. Provider end is strictly earlier than first read. '
    '`trace.json.gz` preserves every occurrence and originating top-level statement, including reads '
    'performed inside called functions; `order.json` preserves the graph and body order.', '',
    '| Top-level statement | Imported bindings (provider body order) | Evidence / decision |',
    '| --- | --- | --- |']
for row in rows.values():
    names = ', '.join(f"`{b['binding']}` ({b['providerBodyOrder']})" for b in row['bindings'])
    evidence = f"reader body {row['bodyOrder']}; " + row['decision']
    lines.append(f"| `{row['statement']}` | {names} | {evidence} |")
lines.extend(['', '## Evaluation order', '', '| Body order | Module |', '| ---: | --- |'])
lines.extend(f'| {rank[file]} | `{file}` |' for file in order['order'])
(BUCKET / 'LEDGER.md').write_text('\n'.join(lines) + '\n')
print(f"Ledger: {len(rows)} statements (57 observed plus one direct conditional), {len(observed)} original reads; all observed provider ends precede reads.")
