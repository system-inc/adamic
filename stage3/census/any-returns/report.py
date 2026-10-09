"""Aggregate frozen boundary credit by the actual failing declaration."""
from collections import defaultdict
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent


def calculate(inspection):
    proposals = json.loads((ROOT / 'return-candidates.json').read_text())
    groups = defaultdict(list)
    for observation in inspection['observations']:
        groups[observation['callee']].append(observation)
    ranked = []
    for key, observations in groups.items():
        callee = inspection['callees'][key]
        declaration_owner = callee['declaration_owner']
        sources = {s['owner'] for e in callee['return_evidence'] for s in e['any_sources']}
        if not callee['stock_return_is_any']:
            origin = 'latent context unresolved: stock return is not any'
        elif declaration_owner != 'tsc':
            origin = declaration_owner
        elif sources == {'lib.d.ts'}:
            origin = 'lib.d.ts'
        elif sources == {'Node typings'}:
            origin = 'Node typings'
        else:
            origin = 'tsc'
        candidate = proposals['callees'].get(key)
        if candidate is None:
            candidate = dict(type=(callee['body_return_candidate'] if callee['stock_return_is_any'] else callee['stock_return'])
                if callee['body_return_candidate'] else None,
                evidence='Existing stock signature and return-expression evidence; no annotation rewrite proposed.'
                    if callee['body_return_candidate'] else 'Body evidence contains any or no body; no truthful replacement inferred.')
        examples = []
        for row in sorted(observations, key=lambda row: (-row['attributed_hidden_bytes'], row['where'])):
            example = row['where'].rsplit(':', 1)[0]
            if example not in examples:
                examples.append(example)
        unique = {(row['file'], row['start'], row['end']) for row in observations}
        credited = {(row['file'], row['start'], row['end']) for row in observations if row['attributed_hidden_bytes']}
        ranked.append(dict(**callee, any_origin=origin, boundary_count=len(unique), raw_observations=len(observations),
            hidden_bytes=sum(row['attributed_hidden_bytes'] for row in observations),
            credited_boundary_count=len(credited), examples=examples[:3], return_candidate=candidate))
        for observation in observations:
            observation['declaration_owner'] = declaration_owner
            observation['any_origin'] = origin
    ranked.sort(key=lambda row: (-row['hidden_bytes'], -row['boundary_count'], row['declaration']))
    def totals(field):
        result = {}
        for row in ranked:
            bucket = result.setdefault(row[field], dict(callees=0, boundaries=0, raw_observations=0, hidden_bytes=0))
            for name, value in [('callees',1), ('boundaries',row['boundary_count']),
                                ('raw_observations',row['raw_observations']), ('hidden_bytes',row['hidden_bytes'])]:
                bucket[name] += value
        for owner in ('tsc', 'lib.d.ts', 'Node typings'):
            result.setdefault(owner, dict(callees=0,boundaries=0,raw_observations=0,hidden_bytes=0))
        return result
    spans = {(row['file'], row['start'], row['end']) for row in inspection['observations']}
    assert sum(row['boundary_count'] for row in ranked) == len(spans), 'one callee per distinct boundary'
    return dict(boundaries=len(spans), raw_observations=len(inspection['observations']),
        hidden_bytes=sum(row['hidden_bytes'] for row in ranked), callees=len(ranked),
        declaration_owners=totals('declaration_owner'), any_origins=totals('any_origin'),
        return_type_definitions=proposals['definitions'], ranked_callees=ranked,
        boundary_observations=inspection['observations'],
        stock_typescript=inspection['typescript'], stock_diagnostics=inspection['diagnostics'])


def render(result):
    text = f'''- Ranked {result['callees']} failing callees for the frozen ed6e2975 any-return reason.
- Reconciled {result['boundaries']:,} distinct boundaries, {result['raw_observations']:,} observations and {result['hidden_bytes']:,} credited bytes.
- Stock TypeScript disagrees with latent any for 167 callees and 82,337 bytes; these are not labelled source adaptations.
- The three-callee stock-checker fixture passes; misattributing JSON.parse to tsc is caught.
- Exact latent instantiation provenance and production fixes remain outside this report.

[RESULT.json](RESULT.json) contains every callee and every boundary observation.
The top 20 by credited hidden bytes are below. The diagnostic's declaration
location identifies the callee; the boundary location identifies where recovery
stopped. They often differ because lowering enters an imported helper or prepares
a signature. Repetitions across attempts are retained as observations, while
boundary counts deduplicate (file, start, end) per callee. All 2,060 distinct
spans resolve to exactly one callee. Existing outermost-cause byte credit is
preserved from ranking `6c4fc1af`: nested or already examined spans have zero
credit. This is not a sum of full boundary lengths.

The unit branches from main `3ffb1a835184713998a34874e86326cd21db971f` and imports
only frozen measurement evidence, without merging the previous topic branch.
Compiler measurement remains `ed6e29751ee47d86fad450cd1674139883bc0f70`.
All adapted compiler file hashes are verified before inspecting them. Stock
TypeScript 6.0.3 resolves declarations and signatures independently. Its options
match the census's strict ES2024/Bundler settings, with Node typings explicitly
loaded to identify their provenance. This inspection reports 172 stock diagnostics;
these are observations on a rejected program, not a successful compiler build.
External library/Node declaration hashes are retained in inspection.json.

Declaration ownership and any provenance are separate. All failing declarations
are tsc's own. An explicit tsc return annotation is an adaptation candidate;
`tryParseJson` also exposes the library producer JSON.parse. The local ambient
setTimeout declaration is tsc-owned, with no body; it is not relabelled as Node's
typings. No recorded failing signature belongs directly to lib.d.ts or Node.
The fixture nevertheless verifies both external declaration classes.

167 declarations have non-any stock return contracts, such as memoize's `() => T`
and forEach's `U | undefined`. The census signature path calls `l.concrete` on
the checker's return type before printing this reason. The original ledger stores
neither the pre-substitution type nor the mapper nor the transitive call stack.
The source of those any results therefore remains **latent checker/substitution/
context unresolved**, rather than an invented explicit any in tsc. Matching calls
within each failed span carry independently resolved stock return types and any
argument sources. They are syntactic evidence, not a recorded failing invocation;
some boundaries have no direct matching call because preparation or transitive
lowering failed. There are 2,322 observations with matching calls in their spans; only two matching
calls have an any stock result (258 credited bytes). The other 21 observations
have no direct matching call. Fixing this discrepancy requires a targeted
lowering probe.

Totals by declaration owner:

| Owner | Callees | Boundaries | Observations | Credited bytes |
|---|---:|---:|---:|---:|
'''
    for field in ('declaration_owners', 'any_origins'):
        if field == 'any_origins':
            text += '\nTotals by any origin (library wrapper provenance included):\n\n| Origin | Callees | Boundaries | Observations | Credited bytes |\n|---|---:|---:|---:|---:|\n'
        for owner, counts in result[field].items():
            text += f"| {owner} | {counts['callees']} | {counts['boundaries']:,} | {counts['raw_observations']:,} | {counts['hidden_bytes']:,} |\n"
    text += '\nTop 20 callees. Examples name boundaries; the callee declaration is a separate column.\n\n| Callee and declaration | Boundaries | Credited bytes | Boundary file:line examples | Any origin | Existing or proposed return contract |\n|---|---:|---:|---|---|---|\n'
    esc = lambda value: str(value).replace('|', '&#124;').replace('\n', '<br>')
    for row in result['ranked_callees'][:20]:
        values = [row['callee']+'<br>'+row['declaration'], row['boundary_count'], f"{row['hidden_bytes']:,}",
            '<br>'.join(row['examples']), row['any_origin'], row['return_candidate']['type'] or 'not evident']
        text += '| ' + ' | '.join(map(esc, values)) + ' |\n'
    text += '''
The eight actual-any tsc declarations follow. These candidates come from inspected
bodies and are separate from recorded checker types. They require contract and
call-site validation before implementation; no source annotation was changed.
JSON recovery objects can contain undefined-valued properties on invalid input.
Use `JsonRecoveryValue = string | number | boolean | null | JsonRecoveryValue[] |
JsonRecoveryObject` and `JsonRecoveryObject = { [key: string]: JsonRecoveryValue |
undefined }`. For strict JSON.parse without a reviver, `JsonValue` recursively
contains only JSON primitives, arrays and objects; it excludes undefined.

| Callee | Bytes | Candidate and body evidence |
|---|---:|---|
'''
    for row in result['ranked_callees']:
        if row['stock_return_is_any']:
            candidate = row['return_candidate']
            text += '| ' + ' | '.join(map(esc,[row['callee']+'<br>'+row['declaration'],row['hidden_bytes'],
                (candidate['type'] or 'not evident')+': '+candidate['evidence']])) + ' |\n'
    text += '''
Reproduce using the frozen adapted tree from the hidden-source measurement:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/any-returns-setup.log 2>&1
source /workspace/adamic-tools/env.sh
export NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules
python3 stage3/census/any-returns/test_fixture.py > /tmp/any-returns-fixture.log 2>&1
node stage3/census/any-returns/inspect.cjs corpus /tmp/hidden-adapted stage3/census/any-returns/evidence/boundaries.json stage3/census/any-returns/evidence/inspection.json > /tmp/any-returns-inspection.log 2>&1
python3 stage3/census/any-returns/report.py > /tmp/any-returns-report.log 2>&1
```

The three-callee fixture is committed as `.a.txt` and copied to a scratch `.a`.
Stock TypeScript accepts it with zero diagnostics. Its known answer is own -> tsc,
JSON.parse -> lib.d.ts and require -> Node typings; own has an explicit any return
but its literal body supports number. Counts are one boundary per callee, with
5 own-call bytes, 16 JSON.parse-call bytes and 20 require-call bytes, total 41.
These are synthetic full call-span credits, not a new latent census run.
The lib-to-tsc mutant changes the real
classifier's result for JSON.parse. It still has zero stock diagnostics and is
caught by `three-callee declaration provenance`, not compilation failure.
Fixture and mutant artifacts and logs are in evidence/.

No native/oracle fixture or counts.md changed. No whole packages, gate or compiler
semantic oracle were run. We did not identify an exact source of any for the
167 stock-signature discrepancies, or prove proposed contracts against all callers.
Setup cumulative timing: Go 0.016s, Node 0.018s, markdown 0.048s, submodules 0.059s,
clang 0.117s, Go build 32.855s, test binaries deferred 32.928s, cache warm 32.929s,
done 32.951s. `nproc` is 5; CPU quota is 4.
'''
    return text


def main():
    inspection = json.loads((ROOT / 'evidence/inspection.json').read_text())
    result = calculate(inspection)
    assert (result['boundaries'], result['raw_observations'], result['hidden_bytes']) == (2060,2343,89627), 'frozen any-return ranking totals'
    result['provenance'] = dict(branch_base='3ffb1a835184713998a34874e86326cd21db971f',
        compiler_commit='ed6e29751ee47d86fad450cd1674139883bc0f70', ranking_commit='6c4fc1af',
        input_hashes={str(p.relative_to(ROOT)):hashlib.sha256(p.read_bytes()).hexdigest()
            for p in [ROOT / 'evidence/boundaries.json', ROOT / 'evidence/inspection.json',
                      ROOT / 'evidence/source-manifest.json', ROOT / 'return-candidates.json']})
    (ROOT / 'RESULT.json').write_text(json.dumps(result,indent=2)+'\n')
    (ROOT / 'README.md').write_text(render(result))
    print(f"PASS: {result['callees']} callees; {result['boundaries']} boundaries; {result['raw_observations']} observations; {result['hidden_bytes']} bytes")
    print(json.dumps(result['any_origins'],sort_keys=True))


if __name__ == '__main__':
    main()
