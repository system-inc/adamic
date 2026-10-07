#!/usr/bin/env python3
"""Validate census coverage and emit the morning/evening checker/lowering meter."""
import json
from collections import Counter
from html import escape
import re
from pathlib import Path
import subprocess
import sys


MEASUREMENT = 'measured on a checker-rejected program'


def reason_owner(reason, owners):
    # The variance family owns every such message, including ones that also
    # begin with a mapped type or generic-return prefix.
    if ' seen as ' in reason and 'seen as' in owners:
        return owners['seen as']
    if reason in owners:
        return owners[reason]
    for prefix in sorted(owners, key=lambda key: (-len(key), key)):
        if prefix != 'seen as' and reason.startswith(prefix):
            return owners[prefix]
    return 'OWNER BLANK'


def latent_summary(tree, run):
    records = [json.loads(line) for line in (run / 'latent.jsonl').read_text().splitlines()]
    if not records or records[0].get('measurement') != MEASUREMENT or records[0].get('status') != 'measurement':
        raise ValueError('invalid latent measurement header')
    expected = {str(path.resolve()) for path in (tree / 'src/compiler').rglob('*')
                if path.is_file() and path.suffix in ('.ts', '.a')}
    seen = set()
    sites = set()
    events = 0
    for record in records[1:]:
        name = str(Path(record['file']).resolve())
        if name not in expected or name in seen or record.get('measurement') != MEASUREMENT:
            raise ValueError('invalid or duplicate latent census file')
        seen.add(name)
        for finding in record.get('findings') or []:
            if finding.get('measurement') != MEASUREMENT or finding['kind'] not in ('NotYet', 'Refused', 'SkippedDependency', 'error', 'panic'):
                raise ValueError('invalid latent finding')
            sites.add(tuple(finding[key] for key in ('kind', 'where', 'reason', 'text')))
            events += 1
    if seen != expected:
        raise ValueError('latent census coverage incomplete')
    counts = Counter(site[0] for site in sites)
    reasons = Counter((site[0], site[2]) for site in sites if site[0] in ('NotYet', 'Refused'))
    ranked = sorted(reasons, key=lambda key: (-reasons[key], key))
    owners = json.loads(Path(__file__).with_name('owners.json').read_text())
    groups = {}
    for (kind, reason), count in reasons.items():
        row = groups.setdefault(reason, {'reason': reason, 'owner': reason_owner(reason, owners),
                                         'NotYet': 0, 'Refused': 0, 'count': 0})
        row[kind] += count
        row['count'] += count
    reason_rows = sorted(groups.values(), key=lambda row: (-row['count'], row['reason']))
    return {
        'measurement': MEASUREMENT,
        'checker_rejected': records[0]['checker_rejected'],
        'count_definition': 'unique (kind, where, reason, text) sites across all attempts',
        'totals': {kind: counts[kind] for kind in ('NotYet', 'Refused', 'SkippedDependency', 'error', 'panic')},
        'per_reason': {kind + ': ' + reason: reasons[kind, reason] for kind, reason in sorted(reasons)},
        'reason_rows': reason_rows,
        'unowned_reasons': [row for row in reason_rows if row['owner'] == 'OWNER BLANK'],
        'top_reasons': [{'kind': kind, 'reason': reason, 'count': reasons[kind, reason], 'owner': reason_owner(reason, owners)}
                        for kind, reason in ranked[:10]],
        'source_files': len(seen),
        'recorded_events': events,
    }


def latent_table(label, summary):
    totals = summary['totals']
    lines = ['', f'Latent lowering, {label}: {MEASUREMENT}.',
             f"NotYet: {totals['NotYet']}; Refused: {totals['Refused']}.", '',
             '| Reason | Owner | NotYet | Refused | Total |', '| --- | --- | ---: | ---: | ---: |']
    for item in summary['reason_rows'][:10]:
        reason = escape(item['reason']).replace('|', '&#124;').replace('\n', ' ')
        lines.append(f"| {reason} | {item['owner']} | {item['NotYet']} | {item['Refused']} | {item['count']} |")
    if not summary['top_reasons']:
        lines.append('| No NotYet or Refused findings | | 0 | 0 | 0 |')
    lines += ['', 'Counts are unique finding sites, not attempt events. Skipped dependencies, errors and panics',
              'are retained separately in JSON. This measurement does not establish successful lowering or native output.']
    return lines


def unowned_table(trees):
    rows = {}
    for label, tree in trees.items():
        for item in tree['latent_lowering']['unowned_reasons']:
            row = rows.setdefault(item['reason'], {'reason': item['reason'], 'owner': 'OWNER BLANK',
                                                  'main_NotYet': 0, 'main_Refused': 0,
                                                  'area_NotYet': 0, 'area_Refused': 0})
            for kind in ('NotYet', 'Refused'):
                row[label + '_' + kind] = item[kind]
    def count(row):
        return max(row['main_NotYet'] + row['main_Refused'], row['area_NotYet'] + row['area_Refused'])
    ordered = sorted(rows.values(), key=lambda row: (-count(row), row['reason']))
    visible = [row for row in ordered if count(row) >= 10]
    tail = [row for row in ordered if count(row) < 10]
    lines = ['', '### Unowned', '', MEASUREMENT + '.',
             'OWNER BLANK rows go to @system_adamic. Counts are grouped by exact reason.', '']
    def table(items):
        result = ['| Reason | Owner | Main NotYet | Main Refused | Area NotYet | Area Refused |',
                  '| --- | --- | ---: | ---: | ---: | ---: |']
        for row in items:
            reason = escape(row['reason']).replace('|', '&#124;').replace('\n', ' ')
            result.append(f"| {reason} | OWNER BLANK | {row['main_NotYet']} | {row['main_Refused']} | {row['area_NotYet']} | {row['area_Refused']} |")
        return result
    if visible:
        lines += table(visible)
    else:
        lines.append('No unowned reasons with at least 10 sites on either tree.')
    tail_sites = sum(row[key] for row in tail for key in
                     ('main_NotYet', 'main_Refused', 'area_NotYet', 'area_Refused'))
    lines += ['', f'{len(tail)} more unowned reasons, {tail_sites} sites in all']
    return ordered, lines


def report(tree, run, stamp):
    compiler = tree / 'src/compiler'
    expected = {str(path.resolve()): path for path in compiler.rglob('*') if path.is_file()}
    records = [json.loads(line) for line in (run / 'census.jsonl').read_text().splitlines()]
    # A root imports other files. Attribute diagnostics to their primary location,
    # never to the root that happened to load them or to an elaboration line.
    located = {}
    unlocated = set()
    external = set()
    for record in records:
        for diagnostic in record.get('diagnostics', []):
            match = re.match(r'^(.+?):[0-9]+:[0-9]+: error TS[0-9]+:', diagnostic)
            if match is None:
                if not diagnostic.startswith('error TS'):
                    raise ValueError(f'unrecognized diagnostic format: {diagnostic}')
                unlocated.add(diagnostic)
                continue
            name = str(Path(match[1]).resolve())
            if name in expected:
                located.setdefault(name, set()).add(diagnostic)
            else:
                external.add(diagnostic)
    rows = {}
    whole = None
    sources = sorted(name for name in expected if Path(name).suffix in ('.ts', '.a'))
    for record in records:
        roots = [str(Path(root).resolve()) for root in record['roots']]
        if len(roots) != 1:
            if whole is not None or sorted(roots) != sources:
                raise ValueError('invalid or duplicate whole-program census record')
            whole = record
            continue
        name = roots[0]
        if name not in expected or name in rows:
            raise ValueError(f'unexpected or duplicate census file: {name}')
        kind = record['kind']
        if kind not in ('accepted', 'checker', 'Refused', 'NotYet', 'error'):
            raise ValueError(f'unknown census kind: {kind}')
        checked = kind in ('accepted', 'Refused', 'NotYet')
        rows[name] = {
            'file': str(expected[name].relative_to(tree)),
            'source': Path(name).suffix in ('.ts', '.a'),
            'checker': checked,
            'checker_whole_program': checked,
            'checker_own_file': Path(name).suffix in ('.ts', '.a') and kind != 'error' and name not in located,
            'own_file_diagnostics': len(located.get(name, set())),
            'lowering': kind == 'accepted',
            'lowering_attempted': checked,
            'outcome': kind,
        }
    if set(rows) != set(expected) or whole is None:
        raise ValueError('census coverage incomplete: per-file or whole-program rows missing')
    files = [rows[name] for name in sorted(rows)]
    totals = {
        'files': len(files),
        'source_files': sum(row['source'] for row in files),
        'checker': sum(row['checker'] for row in files),
        'lowering': sum(row['lowering'] for row in files),
        'checker_whole_program': sum(row['checker_whole_program'] for row in files),
        'checker_own_file': sum(row['checker_own_file'] for row in files),
    }
    result = {
        'timestamp_utc': stamp,
        'adamic_commit': subprocess.check_output(['git', 'rev-parse', 'HEAD'], text=True).strip(),
        'profile': 'unmodified stage 0 checker options and prelude',
        'adapted_tree': str(tree),
        'unlocated_diagnostics': len(unlocated),
        'external_diagnostics': len(external),
        'files': files,
        'totals': totals,
        'whole_program': {key: value for key, value in whole.items() if key != 'diagnostics'},
        'whole_program_diagnostics': len(whole.get('diagnostics', [])),
    }
    (run / 'report.json').write_text(json.dumps(result, indent=2) + '\n')
    lines = [f"Whole program: {totals['checker_whole_program']}/{totals['source_files']}",
             f"Own file: {totals['checker_own_file']}/{totals['source_files']}", '',
             f'Stage 3 meter {stamp}', '', '| File | Whole program | Own file | Lowering |',
             '| --- | --- | --- | --- |']
    for row in files:
        lowering = 'pass' if row['lowering'] else 'fail' if row['lowering_attempted'] else 'blocked'
        lines.append(f"| {row['file']} | {'pass' if row['checker'] else 'fail'} | {'pass' if row['checker_own_file'] else 'fail'} | {lowering} |")
    lines += ['', f"Checker: {totals['checker']}/{totals['source_files']} source files. "
              f"Lowering: {totals['lowering']}/{totals['source_files']} source files.",
              f"All files: {totals['files']}; JSON inputs fail the source-extension gate.",
              'Blocked means lowering was not attempted because the checker/input gate failed.', '']
    (run / 'report.md').write_text('\n'.join(lines))
    return result


def report_pair(main_tree, area_tree, run, stamp, main_commit, area_commit):
    trees = {}
    for label, tree, commit in [('main', main_tree, main_commit), ('area', area_tree, area_commit)]:
        trees[label] = report(tree, run / label, stamp)
        trees[label]['tree_ref'] = 'origin/main' if label == 'main' else 'origin/area/stage3'
        trees[label]['tree_commit'] = commit
        trees[label]['latent_lowering'] = latent_summary(tree, run / label)
        with (run / label / 'report.md').open('a') as output:
            output.write('\n'.join(latent_table(label, trees[label]['latent_lowering'])) + '\n')
        (run / label / 'report.json').write_text(json.dumps(trees[label], indent=2) + '\n')
    # Preserve the existing area's single-tree fields for JSON consumers.
    unowned, unowned_lines = unowned_table(trees)
    result = dict(trees['area'], trees=trees, unowned_reasons=unowned)
    (run / 'report.json').write_text(json.dumps(result, indent=2) + '\n')
    lines = []
    for key, title in [('checker_whole_program', 'Whole program'), ('checker_own_file', 'Own file')]:
        values = [f"{label}: {trees[label]['totals'][key]}/{trees[label]['totals']['source_files']}" for label in trees]
        lines.append(f"{title}: " + '; '.join(values))
    lines += ['', f'Stage 3 meter {stamp}', '',
              f'Main: origin/main at {main_commit}. Area: origin/area/stage3 at {area_commit}.',
              f"Both measured with Adamic {result['adamic_commit']} and ordinary stage 0 options.", '',
              '| File | Main whole program | Main own file | Area whole program | Area own file |',
              '| --- | --- | --- | --- | --- |']
    rows = {label: {row['file']: row for row in tree['files']} for label, tree in trees.items()}
    for name in sorted(set(rows['main']) | set(rows['area'])):
        cells = []
        for label in trees:
            row = rows[label].get(name)
            for key in ['checker_whole_program', 'checker_own_file']:
                cells.append('absent' if row is None else 'pass' if row[key] else 'fail')
        lines.append('| ' + ' | '.join([name, *cells]) + ' |')
    lines += unowned_lines
    for label in trees:
        lines += latent_table(label, trees[label]['latent_lowering'])
    lines += ['', 'Whole program includes imported diagnostics. Own file uses the primary diagnostic location.',
              'Global and external diagnostics are counted separately in JSON; they have no compiler-file location.',
              'Non-source inputs fail the extension gate and are excluded from source denominators.', '']
    (run / 'report.md').write_text('\n'.join(lines))
    return result


if __name__ == '__main__':
    if len(sys.argv) == 4:
        report(Path(sys.argv[1]).resolve(), Path(sys.argv[2]).resolve(), sys.argv[3])
    else:
        report_pair(Path(sys.argv[1]).resolve(), Path(sys.argv[2]).resolve(),
                    Path(sys.argv[3]).resolve(), *sys.argv[4:7])
