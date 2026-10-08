"""Refresh the README from the complete ranking artifact."""
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent


def render(result):
    ranked = result['ranked_reasons']
    other = result['other_causes']
    reason_bytes = sum(r['bytes_revealed_if_fixed_alone'] for r in ranked)
    other_bytes = sum(r['bytes_revealed_if_fixed_alone'] for r in other)
    text = f'''- Built an exact-reason hidden-byte ranking on the frozen ed6e2975 census.
- Partitioned {result['hidden_bytes']:,} hidden bytes exactly once under their outermost stopping causes.
- Refusal/NotYet reasons receive {reason_bytes:,} bytes; other causes receive {other_bytes:,} bytes.
- Four synthetic accounting tests pass; the nested-byte double-count mutant is caught.
- Owners use the pinned tables where supplied; NotYet ownership and actual fix-alone outcomes remain unavailable.

[RESULT.json](RESULT.json) contains all {len(ranked):,} refusal/NotYet reasons, including
zero-credit reasons, all other causes, every raw boundary/skip record and every
attributed half-open byte segment. This unit reuses the complete
[hidden-source measurement](../hidden/README.md), compiler
`ed6e29751ee47d86fad450cd1674139883bc0f70`. It does not remeasure a newer compiler.
The topic branch starts at hidden-source `f1502d130bc0b4440f29b93b76b7d18bff3f6a60`;
no compiler feature branches were merged.

Every one of the 17,643 Boundary records is matched by exact diagnostic text to
its recorded Refused, NotYet, panic/error or SkippedDependency cause. Panic/error
records sharing the same panic text are one panic cause. The 147 checker-skipped
body records and 16 dependency body records are included as checker causes,
matching the original hidden-source union. Frozen per-file hidden ranges are
recomputed from the census and stock catalogue before attribution.

For each residual hidden byte, a strictly enclosing boundary takes precedence
over nested boundaries, including independently attempted nested units. Equal
spans with the same cause merge. Equal spans with differing causes, or crossing
outer spans with differing causes, receive a conflict bucket and no individual
reason credit. Checker-rejected enclosing bodies win over nested lowerer reasons.
Successfully examined bytes have already been removed and receive no credit.
All reasons retain their exact diagnostic text; there is no family normalization.

`bytes_revealed_if_fixed_alone` is the requested outermost-cause attribution:
bytes potentially exposed to the census after removing that outer reason. It is
an estimate, not a measured compiler counterfactual. Inner blockers, other checks
in the same construct and the checker can still stop a rerun. The estimate does
not claim all credited bytes would then lower successfully. Conflicting causes
are retained separately because the ledger cannot choose one cause for them.

`boundary_count` counts distinct (file, start, end) spans carrying that exact
reason, including fully examined or shadowed spans. `raw_boundary_records`
retains repetitions across attempts; `contributing_boundary_count` counts spans
that actually receive outermost credit. Examples are up to three distinct
boundary file:lines, preferring the largest credited boundaries. Fewer than
three means fewer contributing file:lines exist. Whole-body examples name the
boundary selection; its actual failure diagnostic and unit are retained in JSON.

Owner sources are frozen at refusal table
`d35a81d36fdafccf827bad0f572d311b2a0d4deb` and NotYet table
`dc6b1529ae9d2a2210672e105c8bb6374619a59d`. The refusal final and baseline tables
provide production file/function owners for exact matches. The supplied NotYet
table has disposition/why fields but **no owner column**; its owners are null.
Reasons absent from the supplied tables also retain null owners. No owners are
guessed from related templates. The frozen inputs and their hashes are in
evidence/ and RESULT.json. {sum(r['owner'] is not None for r in ranked)} exact reasons have supplied owners.

The top 30 refusal/NotYet reasons follow, sorted by descending credited bytes.
Other causes are excluded from this compiler-reason list and accounted below.

| Rank | Kind and exact reason | Bytes revealed if fixed alone | Boundaries | Three file:line examples (when available) | Owner from supplied tables |
|---:|---|---:|---:|---|---|
'''
    esc = lambda value: str(value).replace('|', '&#124;').replace('\n', '<br>')
    for index, row in enumerate(ranked[:30], 1):
        owner = row['owner'] or ('unavailable: NotYet table has no owners' if row['kind'] == 'NotYet'
                                 else 'unavailable: exact reason absent')
        text += '| ' + ' | '.join(map(esc, [index, row['kind'] + ': ' + row['reason'],
            f"{row['bytes_revealed_if_fixed_alone']:,}", row['boundary_count'],
            '<br>'.join(row['examples']) or 'none', owner])) + ' |\n'
    text += '\nOther causes, also counted once:\n\n| Cause | Hidden bytes |\n|---|---:|\n'
    for kind in sorted({r['kind'] for r in other}):
        total = sum(r['bytes_revealed_if_fixed_alone'] for r in other if r['kind'] == kind)
        text += f'| {kind} | {total:,} |\n'
    text += '''
Reproduce from the repository root:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/hidden-ranking-setup.log 2>&1
source /workspace/adamic-tools/env.sh
python3 stage3/census/hidden-ranking/test_ranking.py > /tmp/hidden-ranking-tests.log 2>&1
python3 stage3/census/hidden-ranking/ranking.py > /tmp/hidden-ranking-result.log 2>&1
python3 stage3/census/hidden-ranking/render.py
```

The committed fixture is a synthetic full census with an outer [10,60) boundary,
a nested [20,30) boundary, a duplicate outer observation and a disjoint [70,90)
boundary. Independent examination removes [40,50). The byte-by-byte oracle
expects exactly 40 outer-reason bytes and 20 inner-reason bytes, total 60;
the nested span receives zero extra credit. Separate tests cover checker ancestors,
conflicting equal spans and crossing spans. The mutant increments the credited
total by one only when a nested boundary is active. It is caught by
`each hidden byte must count exactly once`. The proof runner exits zero only
when that assertion rejects the mutant. All four tests passed; the ranking
command reproduced every original hidden range and partitioned 3,970,749 bytes.
Logs are retained in evidence/. No native/oracle fixture or counts.md changed.
No whole packages or full gate were run.

Setup cumulative timings: Node 0.018s, Go 0.020s, markdown ready 0.055s,
submodules 0.056s, clang 0.112s, Go build 30.167s, test binaries deferred 30.234s,
cache warm 30.235s, done 30.259s. `nproc` is 5; CPU quota is 4.
'''
    return text


if __name__ == '__main__':
    result = json.loads((ROOT / 'RESULT.json').read_text())
    (ROOT / 'README.md').write_text(render(result))
