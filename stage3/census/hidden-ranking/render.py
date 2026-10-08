"""Refresh the README from the complete ranking artifact."""
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent


def render(result):
    ranked = result['ranked_reasons']
    other = result['other_causes']
    reason_bytes = sum(r['bytes_revealed_if_fixed_alone'] for r in ranked)
    other_bytes = sum(r['bytes_revealed_if_fixed_alone'] for r in other)
    compiler = result['provenance']['compiler_commit']
    measured = json.loads((ROOT.parent / 'hidden' / 'RESULT.json').read_text())
    delta = json.loads((ROOT / 'DELTA.json').read_text())
    any_rank = next(i for i, r in enumerate(ranked, 1) if r['kind'] == 'NotYet' and r['reason'] == 'a function returning any')
    text = f'''- Step 05: reran hidden source and ranking on compiler {compiler[:8]}.
- Hidden: **{result['hidden_bytes']:,} / {measured['total_bytes']:,} bytes ({measured['hidden_share']:.6%})**; delta **{delta['hidden_bytes']['delta']:+,} bytes** from ed6e2975.
- Refusal/NotYet reasons receive {reason_bytes:,} bytes; other causes receive {other_bytes:,} bytes.
- Four synthetic accounting tests pass; the nested-byte double-count mutant is caught.
- Owners use the pinned tables where supplied; NotYet ownership and actual fix-alone outcomes remain unavailable.

[RESULT.json](RESULT.json) contains all {len(ranked):,} refusal/NotYet reasons, including
zero-credit reasons, all other causes, every raw boundary/skip record and every
attributed half-open byte segment. This unit reruns the complete
[hidden-source measurement](../hidden/README.md), compiler `{compiler}`,
on area-next-drop 784b577a. The compiler is built in an isolated checkout;
no compiler features are merged into this reporting branch. All 82 adapted input
file hashes and byte lengths match the original ed6e2975 run.

[DELTA.json](DELTA.json) contains every file and exact reason's before/after
values and ranks, including disappeared and newly observed reasons. The original
artifacts remain at report commit 6c4fc1af.

| Metric | ed6e2975 | {compiler[:8]} | Delta |
|---|---:|---:|---:|
| Hidden bytes | {delta['hidden_bytes']['before']:,} | {delta['hidden_bytes']['after']:,} | {delta['hidden_bytes']['delta']:+,} |
| Hidden share | {delta['hidden_share']['before']:.6%} | {delta['hidden_share']['after']:.6%} | {delta['hidden_share_percentage_point_delta']:+.6f} percentage points |
| Blocked union bytes | {delta['blocked_union_bytes']['before']:,} | {delta['blocked_union_bytes']['after']:,} | {delta['blocked_union_bytes']['delta']:+,} |
| Independently examined bytes | {delta['independently_examined_bytes']['before']:,} | {delta['independently_examined_bytes']['after']:,} | {delta['independently_examined_bytes']['delta']:+,} |
| Any-return boundaries | 2,060 | 15 | -2,045 |
| Any-return credited bytes | 89,627 | 7,290 | -82,337 |
| Any-return rank | 8 | {any_rank} | {any_rank - 8:+d} |

The any-return bucket reproduces the compiler worker's 15 boundaries / 7,290
bytes. Its reduction is not the total hidden-byte delta: other reasons can stop
newly exposed statements. Neither change proves successful native lowering.
The total delta compares the entire ed6e2975 and 69501280 compiler pins; it does
not isolate the mapper fix from the intervening area changes.

Every one of the {measured['counts']['boundary_records']:,} Boundary records is matched by exact diagnostic text to
its recorded Refused, NotYet, panic/error or SkippedDependency cause. Panic/error
records sharing the same panic text are one panic cause. The {measured['counts']['checker_skipped_bodies']:,} checker-skipped
body records and {measured['counts']['skipped_dependency_records']:,} dependency body records are included as checker causes,
matching the current hidden-source union. Current per-file hidden ranges are
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
    text += f'''
Reproduce from the repository root:

```sh
# Build at the exact compiler pin in an isolated checkout; share its pinned cohere.
export GOPROXY='https://proxy.golang.org|direct'
source /workspace/adamic-tools/env.sh
export TMPDIR=/workspace/hidden-step05-tmp
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/hidden-step05-overlay > "$TMPDIR/overlay.log" 2>&1
gofmt -w /tmp/hidden-step05-overlay/*.go
go build -buildvcs=false -overlay=/tmp/hidden-step05-overlay/overlay.json -o "$TMPDIR/census" ./stage3/census/latent/tool > "$TMPDIR/build.log" 2>&1
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 "$TMPDIR/census" /tmp/hidden-adapted/src/compiler "$TMPDIR/full.jsonl" > "$TMPDIR/run.log" 2>&1
# From this reporting branch, using the frozen input verified against the original manifest:
NODE_PATH="$HOME/.cache/adamic-stage3/api/node_modules" node stage3/census/hidden/units.cjs /tmp/hidden-adapted/src/compiler "$TMPDIR/stock.json" > "$TMPDIR/stock.log" 2>&1
python3 stage3/census/hidden-ranking/step05.py /tmp/hidden-adapted/src/compiler "$TMPDIR/full.jsonl" "$TMPDIR/stock.json" --commit {compiler} > "$TMPDIR/result.log" 2>&1
python3 stage3/census/hidden-ranking/test_ranking.py > "$TMPDIR/ranking-tests.log" 2>&1
python3 stage3/census/hidden-ranking/test_step05.py > "$TMPDIR/delta-tests.log" 2>&1
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
command reproduced every current hidden range and partitioned {result['hidden_bytes']:,} bytes.
The new known-answer delta fixture also catches a headline +1 mutant and a changed
source-hash mutant, plus a changed-denominator mutant. The hidden arithmetic suite,
its nine mutants, independent byte-mask audit and real nested recovery witness are rerun on the corrected binary.
Logs are retained in evidence/. No native/oracle fixture or counts.md changed.
No whole packages or full gate were run.

Setup on the reporting branch reached Go/Node 0.022s, submodules 0.053s,
markdown 0.060s, clang 0.121s, then failed during Go dependency export with a
full /tmp filesystem. Scratch setup also rejected its shared cohere symlink:
`expected submodule path cohere not to be a symbolic link`. The workaround
retains the initialized matching submodule and tools, and places Go scratch and
new measurement output on /workspace. Logs are retained in evidence/step05.
`nproc` is 5; CPU quota is 4. No production compiler files were edited.
'''
    return text


if __name__ == '__main__':
    result = json.loads((ROOT / 'RESULT.json').read_text())
    (ROOT / 'README.md').write_text(render(result))
