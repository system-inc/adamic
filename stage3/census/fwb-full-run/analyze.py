#!/usr/bin/env python3
"""Summarize original ledger findings and unit outcomes without changing records."""
import collections as C
import gzip
import json
from pathlib import Path

root = Path(__file__).resolve().parent
kinds = ['NotYet', 'Refused', 'panic', 'error', 'Boundary', 'SkippedDependency']
results = {}
for label in ('main', 'fix'):
    records = [json.loads(line) for line in gzip.open(root / f'{label}.jsonl.gz', 'rt')]
    files = [r for r in records if 'units' in r]
    units = [(r['file'], u) for r in files for u in r['units']]
    findings = [f for r in files for f in r['findings']]
    by_unit = C.defaultdict(list)
    for r in files:
        for f in r['findings']:
            by_unit[(r['file'], f['unit'])].append(f)
    counts = C.Counter(f['kind'] for f in findings)
    statuses = C.Counter(u['status'] for _, u in units)
    unit_keys = {(file, u['where']) for file, u in units}
    outcomes = C.Counter()
    reason_units = C.defaultdict(set)
    stopped_reason_units = C.defaultdict(set)
    examples = {}
    stopped_examples = {}
    for file, u in units:
        key = (file, u['where'])
        fs = by_unit[key]
        if u['status'] != 'attempted':
            outcome = u['status']
        elif not fs:
            outcome = 'attempted with no finding'
        elif any(f['kind'] == 'panic' for f in fs):
            outcome = 'panic'
        elif any(f['kind'] in ('NotYet', 'Refused') for f in fs):
            outcome = 'NotYet or Refused'
        else:
            outcome = 'other findings'
        outcomes[outcome] += 1
        for f in fs:
            if f['kind'] in ('NotYet', 'Refused'):
                reason = (f['kind'], f.get('reason') or f.get('text', ''))
                reason_units[reason].add(key)
                if outcome == "NotYet or Refused":
                    stopped_reason_units[reason].add(key)
                examples.setdefault(reason, {'unit': u['where'], 'name': u.get('name', ''), 'site': f.get('where', ''), 'text': f.get('text', '')})
                if outcome == 'NotYet or Refused':
                    stopped_examples.setdefault(reason, {'unit': u['where'], 'name': u.get('name', ''), 'site': f.get('where', ''), 'text': f.get('text', '')})
    panic_counts = C.Counter(f.get('reason') or f.get('text', '') for f in findings if f['kind'] == 'panic')
    panic_examples = {reason: next({'unit': f['unit'], 'site': f.get('where', ''), 'phase': f.get('phase', '')} for f in findings if f['kind'] == 'panic' and (f.get('reason') or f.get('text', '')) == reason) for reason in panic_counts}
    results[label] = {'files': len(files), 'units': len(units), 'findings': dict(counts), 'statuses': dict(statuses), 'outcomes': dict(outcomes), 'panic_reasons': [{'reason': r, 'findings': n, 'example': panic_examples[r]} for r, n in panic_counts.most_common()], 'notyet_refused_reasons': [{'kind': k, 'reason': r, 'units': len(us), 'example': examples[(k, r)]} for (k, r), us in sorted(reason_units.items(), key=lambda item: (-len(item[1]), item[0]))], 'stopped_notyet_refused_reasons': [{'kind': k, 'reason': r, 'units': len(us), 'example': stopped_examples[(k, r)]} for (k, r), us in sorted(stopped_reason_units.items(), key=lambda item: (-len(item[1]), item[0]))], 'unmatched_findings': sum(1 for r in files for f in r['findings'] if (r['file'], f['unit']) not in unit_keys)}
(root / 'summary.json').write_text(json.dumps(results, indent=2) + '\n')

def clean(value):
    return str(value).replace('|', '\\|').replace('\n', ' ').replace(chr(0x2014), '-').replace(chr(0x2013), '-')

lines = ['# Full latent entry census tables', '', '| Metric | Main | Fix |', '| --- | ---: | ---: |']
for metric in ('files', 'units'):
    lines.append(f'| {metric} | {results["main"][metric]} | {results["fix"][metric]} |')
for kind in kinds:
    lines.append(f'| {kind} findings | {results["main"]["findings"].get(kind, 0)} | {results["fix"]["findings"].get(kind, 0)} |')
for label in ('main', 'fix'):
    data = results[label]
    lines += ['', f'## {label.capitalize()} panic reasons', '', '| Reason | Findings | Example unit |', '| --- | ---: | --- |']
    for p in data['panic_reasons']:
        lines.append(f'| {clean(p["reason"])} | {p["findings"]} | {clean(p["example"]["unit"])} |')
    if not data['panic_reasons']:
        lines.append('| None | 0 | |')
lines += ['', '## Fix unit outcomes', '', '| Outcome | Units |', '| --- | ---: |']
for outcome, count in results['fix']['outcomes'].items():
    lines.append(f'| {clean(outcome)} | {count} |')
lines += ['', 'Units are keyed by (file, where). Only attempted units with no findings owned by that key count as lowering with no finding. Deferred generics and checker splits remain separate. NotYet or Refused outcomes exclude panics. File refusal scan findings without a matching unit are counted in finding totals, not assigned to a unit.', '', '## Fix top 20 NotYet or Refused reasons among units without a panic', '', '| Kind | Reason | Distinct units | Example unit | Example site |', '| --- | --- | ---: | --- | --- |']
for r in results['fix']['stopped_notyet_refused_reasons'][:20]:
    lines.append(f'| {r["kind"]} | {clean(r["reason"])} | {r["units"]} | {clean(r["example"]["unit"])} | {clean(r["example"]["site"])} |')
lines += ['', 'These reasons are counted among units in the NotYet or Refused outcome, excluding every unit with a panic finding. Each reason counts distinct units; a unit can have several reasons. Full reason lists and examples are in summary.json.', '', '## Fix top 20 NotYet or Refused reasons across all units', '', 'This additional table includes units that also panic. It does not describe the 254 units that stop without a panic.', '', '| Kind | Reason | Distinct units | Example unit |', '| --- | --- | ---: | --- |']
for r in results['fix']['notyet_refused_reasons'][:20]:
    lines.append(f'| {r["kind"]} | {clean(r["reason"])} | {r["units"]} | {clean(r["example"]["unit"])} |')
lines.append('')
(root / 'TABLES.md').write_text('\n'.join(lines))
print(json.dumps({label: {k: v for k, v in data.items() if k not in ('notyet_refused_reasons', 'stopped_notyet_refused_reasons')} for label, data in results.items()}, indent=2))
