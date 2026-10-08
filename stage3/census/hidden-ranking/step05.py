"""Recount fixed-mapper evidence on the original adapted bytes and compare ed6e2975."""
import argparse
import gzip
import json
import re
from pathlib import Path
import subprocess
import sys

import ranking

ROOT = Path(__file__).resolve().parent
BASE_REPORT = '6c4fc1af'


def original(path):
    return json.loads(subprocess.check_output(['git', 'show', f'{BASE_REPORT}:{path}'], cwd=ROOT))


def delta(before, after):
    assert before['total_bytes'] == after['total_bytes'], 'same denominator'
    assert {k: (v['bytes'], v['sha256']) for k, v in before['files'].items()} == {
        k: (v['bytes'], v['sha256']) for k, v in after['files'].items()}, 'same source bytes'
    result = {key: {'before': before[key], 'after': after[key], 'delta': after[key] - before[key]}
              for key in ('total_bytes', 'hidden_bytes', 'hidden_share', 'blocked_union_bytes', 'independently_examined_bytes')}
    result['hidden_share_percentage_point_delta'] = 100 * result['hidden_share']['delta']
    result['files'] = {k: {'before': before['files'][k]['hidden_bytes'], 'after': v['hidden_bytes'],
                         'delta': v['hidden_bytes'] - before['files'][k]['hidden_bytes']} for k, v in after['files'].items()}
    assert sum(v['delta'] for v in result['files'].values()) == result['hidden_bytes']['delta'], 'file deltas reconcile'
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('compiler', type=Path)
    parser.add_argument('census', type=Path)
    parser.add_argument('stock', type=Path)
    parser.add_argument('--commit', required=True)
    args = parser.parse_args()
    source = ROOT.parent / 'hidden'
    before = original('stage3/census/hidden/RESULT.json')
    old_ranking = original('stage3/census/hidden-ranking/RESULT.json')
    subprocess.run([sys.executable, str(source / 'hidden.py'), str(args.compiler), str(args.census),
                    str(args.stock), str(source / 'RESULT.json'), '--commit', args.commit], check=True)
    for path, name in [(args.census, 'full.jsonl.gz'), (args.stock, 'stock.json.gz')]:
        with gzip.GzipFile(filename=str(source / 'evidence' / name), mode='wb', mtime=0) as stream:
            stream.write(path.read_bytes())
    subprocess.run([sys.executable, str(ROOT / 'ranking.py')], check=True)
    after = json.loads((source / 'RESULT.json').read_text())
    new_ranking = json.loads((ROOT / 'RESULT.json').read_text())
    comparison = delta(before, after)
    comparison['baseline_compiler'] = before['provenance']['compiler_commit']
    comparison['compiler_commit'] = args.commit
    comparison['baseline_report_commit'] = BASE_REPORT
    comparison['source_hashes_identical'] = True
    comparison['counts'] = {'before': before['counts'], 'after': after['counts']}
    old = {(r['kind'], r['reason']): (i, r) for i, r in enumerate(old_ranking['ranked_reasons'], 1)}
    new = {(r['kind'], r['reason']): (i, r) for i, r in enumerate(new_ranking['ranked_reasons'], 1)}
    comparison['reason_deltas'] = []
    for key in sorted(old.keys() | new.keys()):
        before_rank, b = old.get(key, (None, {})); after_rank, a = new.get(key, (None, {}))
        comparison['reason_deltas'].append(dict(kind=key[0], reason=key[1], before_rank=before_rank,
            after_rank=after_rank, before_bytes=b.get('bytes_revealed_if_fixed_alone', 0),
            after_bytes=a.get('bytes_revealed_if_fixed_alone', 0),
            delta_bytes=a.get('bytes_revealed_if_fixed_alone', 0) - b.get('bytes_revealed_if_fixed_alone', 0),
            before_boundaries=b.get('boundary_count', 0), after_boundaries=a.get('boundary_count', 0)))
    any_return = new[('NotYet', 'a function returning any')][1]
    assert (any_return['boundary_count'], any_return['bytes_revealed_if_fixed_alone']) == (15, 7290), 'compiler reported any-return result'
    (ROOT / 'DELTA.json').write_text(json.dumps(comparison, indent=2) + '\n')
    new_ranking['step05_delta'] = {'artifact': 'DELTA.json', 'baseline_compiler': comparison['baseline_compiler'],
                                 'hidden_bytes_delta': comparison['hidden_bytes']['delta'], 'source_hashes_identical': True}
    (ROOT / 'RESULT.json').write_text(json.dumps(new_ranking, indent=2) + '\n')
    subprocess.run([sys.executable, str(ROOT / 'render.py')], check=True)
    # Keep the first region table current while retaining the original report verbatim below.
    baseline_readme = subprocess.check_output(['git', 'show', f'{BASE_REPORT}:stage3/census/hidden/README.md'], cwd=ROOT, text=True)
    text = f'''- Reran hidden source on `{args.commit[:8]}` with the corrected census mapper.
- Hidden: **{after['hidden_bytes']:,} / {after['total_bytes']:,} bytes ({after['hidden_share']:.6%})**.
- Delta from ed6e2975: **{comparison['hidden_bytes']['delta']:+,} bytes**, {comparison['hidden_share_percentage_point_delta']:+.6f} percentage points.
- All 82 source hashes are identical; independent coverage is subtracted before attribution.
- Ranking and complete deltas are in [hidden-ranking](../hidden-ranking/README.md).

## Step 05 current regions

Compiler: `{args.commit}`. Same frozen adapted TypeScript 6.0.3 input as ed6e2975.
Current [RESULT.json](RESULT.json) and evidence/full.jsonl.gz replace the original artifacts;
the original remains available at commit 6c4fc1af.

{after['blocked_union_bytes']:,} blocked union bytes minus {after['independently_examined_bytes']:,}
independently examined bytes = {after['hidden_bytes']:,} hidden bytes.

| File:lines | Hidden bytes | Stopping reasons (full diagnostics in JSON) |
|---|---:|---|
'''
    for r in after['largest_regions']:
        summaries = set()
        for category, reason, failure, unit in r['causes']:
            if category.startswith('checker'):
                codes = sorted(set(re.findall(r'TS[0-9]+', failure)))
                summaries.add('Checker-rejected body: ' + ', '.join(codes))
            else:
                summaries.add(failure.split(': stage 0 ')[-1].splitlines()[0].replace(str(args.compiler) + '/', ''))
        reasons = '<br>'.join(sorted(summaries)).replace('|', '&#124;')
        text += f"| `{r['file']}:{r['start_line']}-{r['end_line']}` | {r['bytes']:,} | {reasons} |\n"
    (source / 'README.md').write_text(text + '\n## Historical ed6e2975 report\n\n' + baseline_readme)
    print('PASS step05:', json.dumps(comparison['hidden_bytes']), 'any-return:', any_return['boundary_count'], any_return['bytes_revealed_if_fixed_alone'])


if __name__ == '__main__':
    main()
