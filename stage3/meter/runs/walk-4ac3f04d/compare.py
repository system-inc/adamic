#!/usr/bin/env python3
"""Recompute comparison tables from committed raw streams and time -v logs.

Site normalization and set counting reuse count-predicates.py from
origin/typescript/stage3-meter-daily-20261026-tzd3gjg:
strip through /src/, then count set(where), never finding tuples or events.
Gate statistics require census.jsonl.gz, which latent does not contain.
"""
import collections
import gzip
import json
from pathlib import Path
import re
import statistics
import sys

HERE = Path(__file__).resolve().parent

def norm(value):
    return re.sub(r'/tmp/[^\s:]*?/src/', 'src/', value)

def site(where):
    if '/src/' in where:
        where = where[where.index('/src/') + 1:]
    return where

def canonical(value):
    if isinstance(value, str):
        return norm(value)
    if isinstance(value, list):
        return [canonical(item) for item in value]
    if isinstance(value, dict):
        return {key: canonical(item) for key, item in value.items()}
    return value

def read(path):
    with gzip.open(path, 'rt') as handle:
        return [json.loads(line) for line in handle]

def table(headers, rows):
    def cell(value):
        return str(value).replace('|', '&#124;').replace('\n', '<br>')
    return '\n'.join(['| ' + ' | '.join(map(cell, headers)) + ' |',
                      '| ' + ' | '.join('---' for _ in headers) + ' |'] +
                     ['| ' + ' | '.join(map(cell, row)) + ' |' for row in rows])

def findings(rows):
    return [f for row in rows for f in row.get('findings', []) or []]

def analyze(run):
    latent = read(run / 'main/latent.jsonl.gz')
    census = read(run / 'main/census.jsonl.gz')
    entry = read(run / 'main/tsc/latent.jsonl.gz')
    sources = [r for r in census if len(r['roots']) == 1 and
               Path(r['roots'][0]).suffix in ('.ts', '.a')]
    diagnosed = set()
    for row in census:
        for diagnostic in row.get('diagnostics', []):
            match = re.match(r'^(.+?):[0-9]+:[0-9]+: error TS[0-9]+:', diagnostic)
            if match:
                diagnosed.add(match[1])
    reasons = collections.defaultdict(set)
    kinds = collections.defaultdict(set)
    first = {}
    for row in latent[1:]:
        fs = row.get('findings', []) or []
        # The file refusal scan is attempted before independent top-level units.
        # Its first finding is therefore the file's recorded first stop.
        first[site(row['file'])] = (norm(fs[0]['kind'] + ' ' + fs[0].get('where', '') + ': ' + fs[0].get('reason', ''))
                                   if fs else 'no recorded stop')
        for f in fs:
            kinds[f['kind']].add(tuple(norm(f.get(k, '')) for k in ('kind', 'where', 'reason', 'text')))
            if f['kind'] in ('NotYet', 'Refused'):
                reasons[norm(f['reason'])].add(site(f.get('where', '')))
    passed = ('accepted', 'NotYet', 'Refused')
    head = [f"{sum(r['kind'] in passed for r in sources)}/{len(sources)}",
            f"{sum(r['kind'] != 'error' and r['roots'][0] not in diagnosed for r in sources)}/{len(sources)}",
            f"{sum(r['kind'] == 'accepted' for r in sources)}/{len(sources)}",
            len(entry[0].get('diagnostics', []))] + [len(kinds[k]) for k in ('NotYet', 'Refused', 'panic', 'error')]
    return head, reasons, first

def seconds(value):
    result = 0.0
    for part in value.split(':'):
        result = result * 60 + float(part)
    return result

def time_row(sha, scope):
    samples = []
    for index in range(1, 4):
        path = HERE / 'timing' / f'{sha}-{scope}-{index}.time'
        text = path.read_text()
        def field(label):
            return re.search(r'^\s*' + re.escape(label) + r':\s*(.+)$', text, re.M).group(1)
        if field('Exit status') != '0':
            raise ValueError(f'failed timing: {path}')
        wall = seconds(field('Elapsed (wall clock) time (h:mm:ss or m:ss)'))
        cpu = float(field('User time (seconds)')) + float(field('System time (seconds)'))
        rss = int(field('Maximum resident set size (kbytes)'))
        samples.append((wall, cpu, rss))
    return [sha[:8], scope, *(f'{statistics.median(s[i] for s in samples):.2f}' if i < 2
                             else str(int(statistics.median(s[i] for s in samples))) for i in range(3)), max(s[2] for s in samples)]

def main():
    manifest = json.loads((HERE / 'inputs.json').read_text())
    runs = [(sha, HERE.parent / name) for sha, name in manifest['runs']]
    analyses = [analyze(run) for sha, run in runs]
    sections = ['# False-direction walk: 76ea8b17 → 4ac3f04d', '',
                'First-error counts are floors: each unit contributes only its first stop. '
                'The latent stream measures lowering on a checker-rejected program; it does not establish native output. '
                'The main and area labels repeat the same pin. Tables use main once, without adding area duplicates.', '',
                '## Headline', '',
                table(['SHA', 'Whole program', 'Own file', 'Lowering', 'tsc entry diagnostics', 'NotYet findings', 'Refused findings', 'Panic findings', 'Error findings'],
                      [[sha[:8], *a[0]] for (sha, run), a in zip(runs, analyses)]), '',
                'Whole program is the number of per-file roots whose entire resolved program passes the checker. '
                'Own file excludes files with a located checker diagnostic. Lowering is ordinary successful IR lowering. '
                'These gate statistics come from census.jsonl.gz; latent.jsonl.gz alone cannot reconstruct them. '
                'NotYet/Refused/panic/error columns use the meter definition: unique (kind, where, reason, text) findings.', '',
                '## Changed lowering reasons', '',
                'Exact reason strings are grouped after normalizing scratch prefixes to src/. '
                'Counts use distinct file:line:column sites, reusing the earlier count-predicates.py set logic. '
                'Only NotYet and Refused are lowering reasons in this table; other errors and panics are separate. A decrease may reflect an earlier stop hiding a prior reason, rather than successful lowering.', '']
    before, after = analyses[0][1], analyses[1][1]
    changed = [(reason, len(before[reason]), len(after[reason]), len(after[reason])-len(before[reason]))
               for reason in before.keys() | after.keys() if len(before[reason]) != len(after[reason])]
    changed.sort(key=lambda r: (-abs(r[3]), r[0]))
    sections += [table(['Reason', 'Before', 'After', 'Difference'], changed) if changed else table(['Reason', 'Before', 'After', 'Difference'], [('None; all distinct-site counts match', '—', '—', 0)]), '',
                 '## Files with a changed first stop', '',
                 'The first finding in each file record is its recorded first stop: the file refusal scan runs before independent top-level-unit attempts. '
                 'This table does not claim to list every unit whose stop changed.', '']
    b, a = analyses[0][2], analyses[1][2]
    changed_files = [(f, b.get(f, 'absent'), a.get(f, 'absent')) for f in sorted(b.keys() | a.keys()) if b.get(f) != a.get(f)]
    sections += [table(['File', 'Before', 'After'], changed_files) if changed_files else 'No files changed first stop.', '', '## Panics', '']
    panics = collections.defaultdict(set)
    for sha, run in runs:
        for label in ('main', 'area'):
            for stream in ('latent.jsonl.gz', 'tsc/latent.jsonl.gz'):
                for f in findings(read(run / label / stream)):
                    if f['kind'] == 'panic':
                        panics[(site(f.get('where', '')), f.get('text') or f.get('reason', ''))].add(sha[:8] + '/' + label + '/' + stream)
    if panics:
        for (where, message), occurrences in sorted(panics.items()):
            sections += [f'### {where}', '', 'Recorded in: ' + ', '.join(sorted(occurrences)), '', '```text', message, '```', '']
    else:
        sections += ['No panics recorded in either run (including both labels and entry streams).', '']
    sections += ['## Latent driver timing', '', 'The system /usr/bin/time was absent and system installation was denied by Unix permissions. The first timing attempt exited 127 before launching the latent binary. GNU time was instead extracted from the official Debian time_1.9-0.2_amd64.deb package and invoked at /workspace/adamic-tools/gnu-time/extracted/usr/bin/time with -v. This is a tool-path deviation from the requested command; the measured compiler binaries and flags are unchanged. Failed prerequisite attempts are preserved under timing/prerequisite-failure/. Package and tool provenance follows.', '', '```text', (HERE / 'timing/tool.txt').read_text().rstrip(), '```', '',
                 'Each invocation uses the exact retained main-label latent-census binary built by twice-daily.sh. '
                 'LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 is set. The checker.ts scope supplies that file as the sole root; '
                 'a whole-program scope would supply the adapted src/compiler directory, but is omitted below. Imports are loaded and checked. '
                 'This is end-to-end latent driver time, including walking input paths, loading/checking, the no-output assertions, '
                 'declaration collection, refusal scans, independent unit lowering, JSON encoding and fsync; it is not isolated production Lower time. '
                 'All timing invocations run sequentially. Each reported value is the median of three samples; '
                 'CPU is user+system per sample and RSS is the per-invocation maximum, then median across samples.', '',
                 table(['SHA', 'Root scope', 'Median wall (s)', 'Median user+sys (s)', 'Median max RSS (KiB)', 'Peak RSS across samples (KiB)'],
                       [time_row(sha, scope) for sha, run in runs for scope in manifest['timing_scopes']]), '',
                 manifest['timing_note'], '', '## Box and run elapsed time', '', '```text', (HERE / 'box.txt').read_text().rstrip(), '```', '',
                 table(['SHA', 'Run directory', 'Meter elapsed (s)'],
                       [[sha[:8], run.name, json.loads((run / 'timing.json').read_text())['elapsed_seconds']] for sha, run in runs]), '',
                 'Meter elapsed includes fetching, compiler/overlay builds, adaptation, census and reporting. Preparation/cache states differed (see the apply logs), so these totals are not isolated lowering times.', '', '## Attribution and reproducibility', '',
                 '4ac3f04d is the direct child of 76ea8b17; there is no other intervening commit. '
                 'There are no stage3 or submodule changes between these pins. '
                 'Checker gate differences cannot be explained by this lowering-only walk and must be treated separately. '
                 'The comparison script checks repeated labels and input metadata below rather than pooling them.', '',
                 'The latent step has no per-file timing fields. Its header records checker diagnostics and spans; each LatentFile records declarations, unit statuses, findings and optional boundaries. Ordinary census seconds measure a different gated path and are not latent lowering times. '
                 'Its stderr reports file and count summaries only. See the pinned main.go.txt and lower.go.txt and the unmodified twice-daily.sh build commands.', '',
                 'Run `python3 stage3/meter/runs/walk-4ac3f04d/compare.py` from any directory to regenerate this file. '
                 'The script reads committed compressed latent and census streams plus raw time -v logs. '
                 'See commands.sh for the exact setup, run, timing and publication commands, timing/commands.log for every timed invocation, '
                 'and run-status.txt for external elapsed times and exit codes. Initial setup overlapped a branch switch and failed; '
                 'the preserved initial log is not a meter failure. Setup was repeated successfully on the stable target checkout.', '']
    checks = []
    timed_hashes = [line.split()[0] for line in (HERE / 'timing/binaries.sha256').read_text().splitlines()]
    assert len(timed_hashes) == len(runs)
    for run_index, (sha, run) in enumerate(runs):
        expected_hash = next(line.split()[0] for line in (run / 'main/binaries.sha256').read_text().splitlines() if line.endswith('/latent-census'))
        assert timed_hashes[run_index] == expected_hash
        for sample_index in range(1, 4):
            sample = read(HERE / 'timing' / f'{sha}-checker-{sample_index}.jsonl.gz')
            assert len(sample) == 2 and sample[0]['latent_mode'] == 'first-error'
            assert site(sample[1]['file']) == 'src/compiler/checker.ts'
        checks.append(f'{sha[:8]}: timed binary SHA-256 {expected_hash} matches the meter; each of the three timing streams contains only the checker.ts file record.')
        pin = json.loads((run / 'compiler-mode.json').read_text())
        assert pin['mode'] == 'per-ref' and pin['latent_mode'] == 'first-error'
        assert pin['commits'] == {'main': sha, 'area': sha}
        for stream in ('latent.jsonl.gz', 'census.jsonl.gz', 'tsc/latent.jsonl.gz', 'tsc/census.jsonl.gz'):
            left = json.loads(json.dumps(read(run / 'main' / stream)).replace('/main-adapted/', '/adapted/'))
            right = json.loads(json.dumps(read(run / 'area' / stream)).replace('/area-adapted/', '/adapted/'))
            # Census includes wall seconds, which need not match repeated measurements.
            for records in (left, right):
                for record in records:
                    record.pop('seconds', None)
            assert left == right, (run, stream, 'label mismatch')
        checks.append(f'{sha[:8]}: pin metadata verified; main/area raw streams agree after path and census-seconds normalization.')
    left = runs[0][1]
    right = runs[1][1]
    for stream in ('latent.jsonl.gz', 'census.jsonl.gz', 'tsc/census.jsonl.gz', 'tsc/latent.jsonl.gz'):
        pair = [read(run / 'main' / stream) for run in (left, right)]
        for records in pair:
            for record in records:
                record.pop('seconds', None)
        same = canonical(pair[0]) == canonical(pair[1])
        checks.append(f'Before/after {stream}: ' + ('identical after source-path and census-seconds normalization.' if same else 'DIFFERS; inspect the retained raw streams. Checker gate differences would be outside the lowering-only walk.'))
    checks.append('The live origin/main advanced during the second fetch; both explicit run pins remained fixed, as verified by compiler-mode.json. No intervening commit was included in this comparison.')
    sections += ['- ' + check for check in checks]
    (HERE / 'COMPARISON.md').write_text('\n'.join(sections) + '\n')
    print(table(['Reason', 'Before', 'After', 'Difference'], changed or [('None; all distinct-site counts match', '—', '—', 0)]))

if __name__ == '__main__':
    main()
