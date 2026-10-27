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

def k1t2wp9():
    before_sha = '4ac3f04d767ec7d2d4e9b41ae9cdc9c4c9f4f288'
    after_sha = '8502a46e9289d6abb673f1359aa12d66be659b55'
    evidence = HERE / 'k1t2wp9'
    runs = [(before_sha, HERE.parent / '20261027T025412Z.HCmJYO'),
            (after_sha, HERE.parent / Path((evidence / 'run-path.txt').read_text().strip()).name)]
    aa = [analyze(run) for sha, run in runs]
    target = 'checked predicate membership by an enum tag another member can hold'
    named = [target, 'checked predicate membership over an open tag domain']
    b, a = aa[0][1], aa[1][1]
    changed = sorted([(r, len(b[r]), len(a[r]), len(a[r])-len(b[r]))
                      for r in b.keys() | a.keys() if len(b[r]) != len(a[r])], key=lambda x: (-abs(x[3]), x[0]))
    sections = ['# K1T2WP9: 4ac3f04d → 8502a46e', '',
        'One foreground meter invocation at 8502a46e, compared with the committed 4ac3f04d run. Main only is counted; area repeats the same pin. First-error findings are floors, and latent lowering on a checker-rejected program does not establish native output.', '',
        '## Headline', '', table(['SHA', 'Whole program', 'Own file', 'Lowering', 'tsc entry diagnostics', 'NotYet findings', 'Refused findings', 'Panic findings', 'Error findings'], [[sha[:8], *x[0]] for (sha, run), x in zip(runs, aa)]), '',
        'Gate columns come from census; finding columns count unique (kind, where, reason, text), as in COMPARISON.md.', '',
        '## Named reasons', '', table(['Reason', 'Before distinct sites', 'After distinct sites', 'Difference'], [(r,len(b[r]),len(a[r]),len(a[r])-len(b[r])) for r in named]), '',
        '## Changed lowering reasons', '', table(['Reason', 'Before', 'After', 'Difference'], changed), '',
        'Counts use distinct normalized file:line:column sites, not finding events.', '']
    rows = [read(run / 'main/latent.jsonl.gz') for sha,run in runs]
    dropped = b[target] - a[target]
    old = collections.defaultdict(set)
    new = collections.defaultdict(list)
    for f in findings(rows[0]):
        if f.get('reason') == target and site(f.get('where','')) in dropped:
            old[site(f['where'])].add(site(f['unit']))
    for f in findings(rows[1]):
        if f.get('phase') == 'lowering':
            new[site(f.get('unit',''))].append(f)
    after_units={site(u['where']) for row in rows[1] for u in row.get('units', [])}
    assert all(u in after_units for units in old.values() for u in units)
    details=[]
    moved=collections.Counter()
    no_finding=0
    for s in sorted(dropped):
        later = {(f['kind'], f.get('reason',''), site(f.get('where',''))) for u in old[s] for f in new[u]}
        if later:
            for reason in {reason for kind,reason,where in later}: moved[reason]+=1
            result='; '.join(f'{kind}: {reason} at {where}' for kind,reason,where in sorted(later))
        else:
            no_finding+=1
            result='no lowering finding in owning unit(s)'
        details.append((s, ', '.join(sorted(old[s])),result))
    sections += ['## Split of removed enum-tag sites', '',
        f'{len(dropped)} removed sites: {no_finding} reach no lowering finding in their owning units; {len(dropped)-no_finding} move to a later finding. {len(a[target]-b[target])} newly appearing sites for the original reason. A no-finding unit is not proof of native output. Later findings are matched by normalized owning unit, using the lowering phase; refusal-scan findings are not substituted for unit stops. A site with multiple owning units is classified as moved if any has a later finding. Reason totals can overlap if such units differ. Multiple removed sites can reach the same later callee site, so these attribution counts differ from the distinct-site reason changes above; the getter reason has no net distinct-site change.', '',
        table(['Later reason', 'Removed sites reaching it'], sorted(moved.items(),key=lambda x:(-x[1],x[0]))), '',
        '<details><summary>Every removed site and its owning units</summary>', '', table(['Before site','Owning unit(s)','After'],details), '', '</details>', '',
        '## Files with a changed first stop', '',
        'First recorded file finding, including the file refusal scan; this is distinct from independent unit stops.', '',
        table(['File','Before','After'],[(f,aa[0][2].get(f,'absent'),aa[1][2].get(f,'absent')) for f in sorted(aa[0][2].keys()|aa[1][2].keys()) if aa[0][2].get(f)!=aa[1][2].get(f)]) or 'No changed first stops.', '', '## Panics', '']
    panic=collections.defaultdict(set)
    for sha,run in runs:
        for label in ('main','area'):
            for stream in ('latent.jsonl.gz','tsc/latent.jsonl.gz'):
                for f in findings(read(run/label/stream)):
                    if f['kind']=='panic': panic[(site(f.get('where','')),f.get('text') or f.get('reason',''))].add(sha[:8]+'/'+label+'/'+stream)
    for (where,msg),labels in sorted(panic.items()): sections += [f'### {where}', '', ', '.join(sorted(labels)), '', '```text',msg,'```','']
    sections += ['## Timing on this box', '', 'GNU time -v; checker.ts is the sole root, with LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1. These are end-to-end latent driver measurements, including checker work and JSON/fsync, not isolated production lowering. Samples alternate before/after, sequentially.', '']
    samples=[]
    for i in range(1,4):
        for sha,run in runs:
            p=evidence/'timing'/f'{sha}-checker-{i}.time'
            if not p.exists(): continue
            txt=p.read_text()
            def field(label): return re.search(r'^\s*'+re.escape(label)+r':\s*(.+)$',txt,re.M).group(1)
            assert field('Exit status')=='0',p
            sample=read(p.with_suffix('.jsonl.gz'))
            assert len(sample)==2 and sample[0]['latent_mode']=='first-error'
            assert site(sample[1]['file'])=='src/compiler/checker.ts'
            samples.append([sha[:8],i,seconds(field('Elapsed (wall clock) time (h:mm:ss or m:ss)')),round(float(field('User time (seconds)'))+float(field('System time (seconds)')),2),field('Maximum resident set size (kbytes)')])
    sections += [table(['SHA','Sample','Wall (s)','User+system (s)','Max RSS (KiB)'],samples) if samples else 'Timing pending or skipped; see timing status.', '', '## Box', '', '```text',(evidence/'box.txt').read_text().strip(),'```','', '## Reproduction and checks', '',
        'Run `python3 stage3/meter/runs/walk-4ac3f04d/compare.py --k1t2wp9`. The original comparison mode remains available. Build, measurement and timing commands and raw logs are retained under k1t2wp9/.', '']
    for sha,run in runs:
        pin=json.loads((run/'compiler-mode.json').read_text())
        assert pin['commits']=={'main':sha,'area':sha} and pin['mode']=='per-ref' and pin['latent_mode']=='first-error'
        sections += [f'- {sha[:8]}: per-ref pin and first-error metadata verified.']
        for stream in ('latent.jsonl.gz','census.jsonl.gz','tsc/latent.jsonl.gz','tsc/census.jsonl.gz'):
            pair=[canonical(read(run/label/stream)) for label in ('main','area')]
            for records in pair:
                for record in records: record.pop('seconds',None)
            assert pair[0]==pair[1],(sha,stream,'label mismatch')
        sections += [f'- {sha[:8]}: main/area streams agree after source-path and census-seconds normalization.']
    if len(samples)==6:
        hashes=[line.split()[0] for line in (evidence/'timing/binaries.sha256').read_text().splitlines()]
        expected_after=next(line.split()[0] for line in (runs[1][1]/'main/binaries.sha256').read_text().splitlines() if line.endswith('/latent-census'))
        expected_before=next(line.split()[0] for line in (evidence/'before-build/binaries.sha256').read_text().splitlines() if line.endswith('/latent-census'))
        original_before=next(line.split()[0] for line in (runs[0][1]/'main/binaries.sha256').read_text().splitlines() if line.endswith('/latent-census'))
        assert expected_before==original_before
        assert hashes==[expected_before,expected_after]*3
        sections += ['- All six timing streams contain checker.ts alone; exit statuses are zero. Binary hashes match the retained after meter binary and the separately rebuilt before binary. The rebuilt before latent binary also exactly matches the SHA-256 of the committed baseline run (9ed61cdc7c4d8f0a60d4805f10f0d32ca8a2f18e643b14144ee5075496e05169).', '', '### Timing tool', '', '```text',(evidence/'timing/tool.txt').read_text().strip(),'```']
    sections += ['', '### Run and build status', '', '```text', (evidence/'run-status.txt').read_text().strip()]
    if (evidence/'before-build-status.txt').exists(): sections += [(evidence/'before-build-status.txt').read_text().strip()]
    sections += ['```', '', 'All 746 adapted src files are byte-identical between the independently generated timing trees (input-check.txt). /usr/bin/time was absent; GNU time was extracted from the official Debian time_1.9-0.2_amd64.deb package, as in the prior run. Tool and package hashes are retained in timing/. The cgroup memory limit is 17179869184 bytes (16 GiB). The 8502a46e meter artifacts were committed and pushed immediately on exit as c4ea5bdd7, before comparison or rebuilding the before binary. The run took 4286 seconds. Both completed compiler source worktrees were removed during the latent pass to free tmpfs at the memory limit, matching the prior run cleanup. Binaries, overlays, adapted trees and logs were retained. No OOM occurred. Only test fixtures and a meter report change intervene between the pins; stage3 adaptation, census drivers, overlay generators and twice-daily.sh are unchanged; report.py and its tests/evidence changed.']
    (HERE/'K1T2WP9.md').write_text('\n'.join(sections)+'\n')
    print('\n'.join(sections[:14]))

if __name__ == '__main__':
    if '--k1t2wp9' in sys.argv:
        k1t2wp9()
    else:
        main()
