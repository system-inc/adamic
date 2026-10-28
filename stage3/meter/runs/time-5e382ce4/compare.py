#!/usr/bin/env python3
"""Verify checker-only evidence and regenerate TIMING.md from GNU time logs."""
import collections
import gzip
import json
from pathlib import Path
import re
import statistics

HERE = Path(__file__).resolve().parent
SCRATCH = '/workspace/time-5e382ce4/'
PINS = {'parent': '22026c7392f55556e721b6f425dd369ceafdf14f', 'change': '5e382ce4baca2f75cc9a452f9b438c8318f375a5'}

def canonical(value):
    if isinstance(value, str):
        for label in PINS:
            value = value.replace(SCRATCH + label + '-adapted/', '')
        return value
    if isinstance(value, list):
        return [canonical(x) for x in value]
    if isinstance(value, dict):
        return {k:canonical(v) for k,v in value.items()}
    return value

def table(headers, rows):
    def cell(x): return str(x).replace('|', '&#124;').replace('\n', '<br>')
    return '\n'.join('| ' + ' | '.join(map(cell, row)) + ' |' for row in [headers, ['---']*len(headers), *rows])

def read(label, index):
    with gzip.open(HERE/'timing'/f'{label}-{index}.jsonl.gz', 'rt') as f:
        rows = canonical([json.loads(line) for line in f])
    assert len(rows)==2 and rows[0]['latent_mode']=='first-error'
    assert rows[1]['file']=='src/compiler/checker.ts'
    return rows

def findings(rows):
    return rows[1].get('findings', []) or []

def reasons(rows):
    result=collections.defaultdict(set)
    for f in findings(rows):
        if f['kind'] in ('NotYet','Refused'):
            result[f['reason']].add(f.get('where',''))
    return result

samples=[]
streams={}
for index in range(1,6):
    for label in PINS:
        text=(HERE/'timing'/f'{label}-{index}.time').read_text()
        def field(name):
            return re.search(r'^\s*'+re.escape(name)+r':\s*(.+)$',text,re.M).group(1)
        assert field('Exit status')=='0'
        wall=0.
        for part in field('Elapsed (wall clock) time (h:mm:ss or m:ss)').split(':'):
            wall=wall*60+float(part)
        cpu=float(field('User time (seconds)'))+float(field('System time (seconds)'))
        rss=int(field('Maximum resident set size (kbytes)'))
        samples.append((label,index,wall,cpu,rss))
        rows=read(label,index)
        if label in streams: assert rows==streams[label], (label,index,'nonrepeatable stream')
        streams[label]=rows
hashes=[line.split()[0] for line in (HERE/'timing/binaries.sha256').read_text().splitlines()]
expected={label:next(line.split()[0] for line in (HERE/label/'binaries.sha256').read_text().splitlines() if line.endswith('/latent-census')) for label in PINS}
assert hashes==[expected[label] for index in range(5) for label in PINS]
a,b=(streams[label] for label in PINS)
identical=a==b
same_findings=findings(a)==findings(b)
before,after=reasons(a),reasons(b)
changes=[]
for reason in sorted(before.keys()|after.keys()):
    removed=before[reason]-after[reason]
    added=after[reason]-before[reason]
    if removed or added:
        changes.append((reason,len(before[reason]),len(after[reason]),len(removed),len(added)))
summary=[]
for label in PINS:
    values=[s for s in samples if s[0]==label]
    summary.append((label,*(statistics.median(s[i] for s in values) for i in (2,3,4)), min(s[2] for s in values), max(s[2] for s in values)))
p,c=summary
wall_overlap=max(p[4],c[4])<=min(p[5],c[5])
cpu_ranges=[(min(s[3] for s in samples if s[0]==label),max(s[3] for s in samples if s[0]==label)) for label in PINS]
cpu_overlap=max(r[0] for r in cpu_ranges)<=min(r[1] for r in cpu_ranges)
percent=(c[1]/p[1]-1)*100
out=['# checker.ts timing: 22026c73 → 5e382ce4','',
     f'Parent: `{PINS["parent"]}`. Change: `{PINS["change"]}` (its direct child).', '',
     'Five samples per revision, alternating parent then change, sequentially on one box. All builds and timed invocations exited zero. checker.ts was the sole root; no whole-directory latent run was made.', '',
     'This follows the commands.sh, K1T2WP9.md and compare.py method on typescript/stage3-meter-walk-4ac3f04d. GNU time -v measures end-to-end latent driver time: imports/loading/checking, declaration collection, no-output assertions, refusal scans, independent unit lowering, JSON and fsync. It does not isolate production Lower time. LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1; first-error findings are floors, and checker-rejected latent lowering does not establish native output.', '',
     table(['Revision','Sample','Wall (s)','User + system (s)','Max RSS (KiB)'],[(label,index,f'{wall:.2f}',f'{cpu:.2f}',rss) for label,index,wall,cpu,rss in samples]), '',
     table(['Revision','Median wall (s)','Median CPU (s)','Median RSS (KiB)','Wall min (s)','Wall max (s)'], [(r[0], f'{r[1]:.2f}', f'{r[2]:.2f}', r[3], f'{r[4]:.2f}', f'{r[5]:.2f}') for r in summary]),'',
     f'Median wall change: {percent:+.2f}%. Wall ranges {"overlap" if wall_overlap else "do not overlap"}; CPU ranges {"overlap" if cpu_overlap else "do not overlap"} ({cpu_ranges[0][0]:.2f}–{cpu_ranges[0][1]:.2f} vs {cpu_ranges[1][0]:.2f}–{cpu_ranges[1][1]:.2f} s).', '',
     ('The wall difference is not clearly past noise in these five samples: the observed ranges overlap.' if wall_overlap else 'The wall difference is past the observed sample noise on this box: the two ranges do not overlap. Five samples support this local comparison, not a universal effect size.'), '',
     f'Normalized full checker.ts latent records are {"identical" if identical else "different"}. Findings are {"identical" if same_findings else "different"}. Every revision reproduces its normalized stream across all five samples. Normalization replaces only the two adapted-tree path prefixes; all JSON fields and array order are retained.', '',
     table(['Changed reason','Parent distinct sites','Change distinct sites','Removed sites','Added sites'],changes) if changes else 'No lowering reason has a changed distinct-site set.', '']
if changes:
    details=[]
    for reason, *_ in changes:
        for site in sorted(before[reason]-after[reason]): details.append(('removed',reason,site))
        for site in sorted(after[reason]-before[reason]): details.append(('added',reason,site))
    out += [table(['Change','Reason','Distinct site'],details),'']
if not identical:
    out += ['Full normalized JSON differences are retained in normalized-diff.txt.','']
    import difflib
    (HERE/'normalized-diff.txt').write_text(''.join(difflib.unified_diff(json.dumps(a,indent=2,sort_keys=True).splitlines(True),json.dumps(b,indent=2,sort_keys=True).splitlines(True),fromfile='parent',tofile='change')))
out += ['Both revisions were built in separate detached worktrees with the exact build_census sequence from twice-daily.sh (census, latent overlay, entry overlay, entry-census, latent-census), then stage3/apply.sh. '+(HERE/'input-check.txt').read_text().strip(), '',
        'Each binary SHA-256 is in parent/binaries.sha256 or change/binaries.sha256; timed binary hashes were checked against those manifests. Submodule pins and clean tracked-source checks are retained. /usr/bin/time was absent, so GNU time was extracted from the official Debian time_1.9-0.2_amd64.deb package; tool and package hashes are retained.', '',
        '```text', (HERE/'box.txt').read_text().strip(), '```', '',
        'Reproduce with commands.sh after successful setup (see its preamble). Regenerate this report with `python3 stage3/meter/runs/time-5e382ce4/compare.py`. All ten raw streams are retained under timing/; parent-1.jsonl.gz and change-1.jsonl.gz are the compared pair. Timing commands, stdout/stderr, GNU time logs, setup, build, overlay, package and apply logs are retained. No compiler code was changed.', '']
(HERE/'TIMING.md').write_text('\n'.join(out))
print('\n'.join(out[:22]))
