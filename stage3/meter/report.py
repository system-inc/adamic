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
ENTRY_MEASUREMENT = 'measured on a checker-clean entry-root program'
REJECTED_ENTRY_MEASUREMENT = 'measured on a checker-rejected entry-root program'


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
    expected = {str(path.resolve()) for path in (tree / 'src/compiler').rglob('*')
                if path.is_file() and path.suffix in ('.ts', '.a')}
    return lowering_summary(records, expected, MEASUREMENT)


def lowering_summary(records, expected, measurement):
    if not records or records[0].get('measurement') != measurement or records[0].get('status') != 'measurement':
        raise ValueError('invalid latent measurement header')
    seen = set()
    sites = set()
    events = 0
    for record in records[1:]:
        name = str(Path(record['file']).resolve())
        if name not in expected or name in seen or record.get('measurement') != measurement:
            raise ValueError('invalid or duplicate latent census file')
        seen.add(name)
        for finding in record.get('findings') or []:
            if finding.get('measurement') != measurement or finding['kind'] not in ('NotYet', 'Refused', 'SkippedDependency', 'error', 'panic', 'Boundary'):
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
        'measurement': measurement,
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
        'latent_mode': records[0].get('latent_mode', 'first-error'),
        'boundaries': counts['Boundary'],
        'excluded_nested_functions': [dict(unit, file=record['file'], bytes=unit.get('body_end',0)-unit.get('body_start',0))
                                      for record in records[1:] for unit in record.get('units',[])
                                      if unit.get('depth',0)>0 and unit['status']=='split_checker_body'],
    }


def entry_root_matches(record, entry):
    roots = record.get('roots')
    return (isinstance(roots, list) and len(roots) == 1 and isinstance(roots[0], str)
            and Path(roots[0]).resolve() == entry)


def entry_summary(tree, run):
    entry = (tree / 'src/tsc/tsc.ts').resolve()
    if not entry.is_file():
        raise ValueError(f'tsc entry missing: {entry}')
    records = [json.loads(line) for line in (run / 'census.jsonl').read_text().splitlines()]
    # A single file produces its ordinary per-file and aggregate observations.
    if len(records) != 2 or any(not entry_root_matches(record, entry) for record in records):
        raise ValueError('invalid tsc entry census roots or coverage')
    comparable = [{key: value for key, value in record.items() if key != 'seconds'} for record in records]
    if comparable[0] != comparable[1]:
        raise ValueError('tsc entry observations disagree')
    record = records[-1]
    if record.get('kind') not in ('accepted', 'checker', 'NotYet', 'Refused'):
        raise ValueError('invalid tsc entry checker outcome: ' + str(record.get('kind')))
    checked = record['kind'] != 'checker'
    diagnostics = record.get('diagnostics', [])
    if (checked and diagnostics) or (not checked and not diagnostics):
        raise ValueError('tsc entry checker outcome disagrees with diagnostics')
    latent = [json.loads(line) for line in (run / 'latent.jsonl').read_text().splitlines()]
    if not latent or not entry_root_matches(latent[0], entry):
        raise ValueError('invalid tsc entry lowering roots')
    header = latent[0]
    if header.get('checker_rejected') != (not checked) or sorted(header.get('diagnostics', [])) != sorted(diagnostics):
        raise ValueError('tsc entry checker and lowering diagnostics disagree')
    result = {'root': 'src/tsc/tsc.ts', 'checker_whole_program': checked,
              'whole_program_diagnostics': len(diagnostics), 'diagnostics': diagnostics,
              'outcome': record['kind'], 'lowering_attempted': checked, 'lowering_census': None}
    if not checked and header.get('latent_mode') != 'full':
        if len(latent) != 1 or header.get('status') != 'blocked':
            raise ValueError('tsc entry lowering ran despite checker diagnostics')
        return result
    sources = header.get('sources', [])
    expected = {str(Path(name).resolve()) for name in sources}
    if len(expected) != len(sources) or str(entry) not in expected or any(not Path(name).is_file() for name in expected):
        raise ValueError('invalid tsc entry resolved reach')
    measurement = ENTRY_MEASUREMENT if checked else REJECTED_ENTRY_MEASUREMENT
    result['lowering_census'] = lowering_summary(latent, expected, measurement)
    result['measurement_attempted'] = True
    return result


def entry_table(label, entry):
    summary = entry['lowering_census']
    if summary is None:
        return ['', f'tsc entry lowering, {label}: blocked by checker diagnostics.']
    lines = ['', f"tsc entry lowering, {label}: {summary['measurement']}.",
             f"Resolved source files: {summary['source_files']}; "
             f"NotYet: {summary['totals']['NotYet']}; Refused: {summary['totals']['Refused']}.",
             f"Errors: {summary['totals']['error']}; panics: {summary['totals']['panic']}; "
             f"skipped dependencies: {summary['totals']['SkippedDependency']}.", '',
             '| Reason | Owner | NotYet | Refused | Total |', '| --- | --- | ---: | ---: | ---: | ---: |']
    for row in summary['reason_rows'][:10]:
        reason = escape(row['reason']).replace('|', '&#124;').replace('\n', ' ')
        lines.append(f"| {reason} | {row['owner']} | {row['NotYet']} | {row['Refused']} | {row['count']} |")
    if not summary['reason_rows']:
        lines.append('| No NotYet or Refused findings | | 0 | 0 | 0 |')
    lines += ['', 'Counts cover the entry root and its resolved implementation dependencies, excluding declarations.',
              'Errors and panics are retained separately in JSON. This census does not establish native output.']
    return lines


def latent_table(label, summary):
    totals = summary['totals']
    lines = ['', f'Latent lowering, {label}: {MEASUREMENT}.',
             f"NotYet: {totals['NotYet']}; Refused: {totals['Refused']}.",
             f"Recovery boundaries: {summary['boundaries']}; named nested checker exclusions: {len(summary['excluded_nested_functions'])}.", '',
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


ADAPTATION_ROOTS = {'src/compiler/hostErrors.ts': '47-host-errors'}
ROOT_TARGET = 79


def root_inventory(files):
    sources = [row for row in files if row['source']]
    created = [dict(file=row['file'], adaptation=ADAPTATION_ROOTS[row['file']])
               for row in sources if row['file'] in ADAPTATION_ROOTS]
    upstream = [row for row in sources if row['file'] not in ADAPTATION_ROOTS]
    return {'target': ROOT_TARGET, 'total': len(sources), 'tsc_original': len(upstream),
            'adaptation_created': created,
            'tsc_whole_program': sum(row['checker_whole_program'] for row in upstream),
            'tsc_own_file': sum(row['checker_own_file'] for row in upstream)}


def root_lines(trees):
    lines = ['Roots (target 79/79): ' + '; '.join(
        f"{label}: {tree['roots']['total']} total, {tree['roots']['tsc_original']} tsc original"
        for label, tree in trees.items())]
    for label, tree in trees.items():
        roots = tree['roots']
        lines.append(f"Tsc original roots, {label}: whole program {roots['tsc_whole_program']}/{roots['tsc_original']}; own file {roots['tsc_own_file']}/{roots['tsc_original']}.")
        for row in roots['adaptation_created']:
            lines.append(f"Adaptation-created root, {label}: {row['file']} (adaptation {row['adaptation']}).")
    return lines


def report(tree, run, stamp, compiler_commit=None):
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
        'adamic_commit': compiler_commit or subprocess.check_output(['git', 'rev-parse', 'HEAD'], text=True).strip(),
        'profile': 'unmodified stage 0 checker options and prelude',
        'adapted_tree': str(tree),
        'unlocated_diagnostics': len(unlocated),
        'external_diagnostics': len(external),
        'files': files,
        'roots': root_inventory(files),
        'totals': totals,
        'whole_program': {key: value for key, value in whole.items() if key != 'diagnostics'},
        'whole_program_diagnostics': len(whole.get('diagnostics', [])),
    }
    (run / 'report.json').write_text(json.dumps(result, indent=2) + '\n')
    lines = [f"Whole program: {totals['checker_whole_program']}/{totals['source_files']}",
             f"Own file: {totals['checker_own_file']}/{totals['source_files']}", '',
             f'Stage 3 meter {stamp}', '', '| File | Whole program | Own file | Lowering |',
             '| --- | --- | --- | --- |']
    lines[3:3] = root_lines({'tree': result}) + ['']
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
    metadata_path = run / 'compiler-mode.json'
    metadata = json.loads(metadata_path.read_text()) if metadata_path.exists() else None
    mode = metadata['mode'] if metadata else 'single'
    if mode not in ('single', 'per-ref'):
        raise ValueError('invalid compiler mode')
    if metadata:
        commits = metadata['commits']
        if set(commits) != {'main', 'area'} or any(not re.fullmatch(r'[0-9a-f]{40}', sha) for sha in commits.values()):
            raise ValueError('invalid compiler commits')
        if mode == 'per-ref' and commits != {'main': main_commit, 'area': area_commit}:
            raise ValueError('per-ref compiler commits do not match source pins')
        if mode == 'single' and commits['main'] != commits['area']:
            raise ValueError('single compiler commits differ')
    trees = {}
    for label, tree, commit in [('main', main_tree, main_commit), ('area', area_tree, area_commit)]:
        trees[label] = report(tree, run / label, stamp, metadata['commits'][label] if metadata else None)
        trees[label]['tree_ref'] = 'origin/main' if label == 'main' else 'origin/area/stage3'
        trees[label]['tree_commit'] = commit
        trees[label]['latent_lowering'] = latent_summary(tree, run / label)
        trees[label]['tsc_entry'] = entry_summary(tree, run / label / 'tsc')
        if metadata and 'latent_mode' in metadata:
            if trees[label]['latent_lowering']['latent_mode'] != metadata['latent_mode']:
                raise ValueError('latent overlay mode does not match requested mode')
            entry_census=trees[label]['tsc_entry']['lowering_census']
            if entry_census and entry_census['latent_mode'] != metadata['latent_mode']:
                raise ValueError('entry overlay mode does not match requested mode')
        with (run / label / 'report.md').open('a') as output:
            output.write('\n'.join(entry_table(label, trees[label]['tsc_entry']) +
                                   latent_table(label, trees[label]['latent_lowering'])) + '\n')
        (run / label / 'report.json').write_text(json.dumps(trees[label], indent=2) + '\n')
    # Preserve the existing area's single-tree fields for JSON consumers.
    unowned, unowned_lines = unowned_table(trees)
    result = dict(trees['area'], trees=trees, unowned_reasons=unowned, compiler_mode=mode)
    (run / 'report.json').write_text(json.dumps(result, indent=2) + '\n')
    lines = []
    for key, title in [('checker_whole_program', 'Whole program'), ('checker_own_file', 'Own file')]:
        values = [f"{label}: {trees[label]['totals'][key]}/{trees[label]['totals']['source_files']}" for label in trees]
        lines.append(f"{title}: " + '; '.join(values))
    lines.append('tsc entry: ' + '; '.join(f"{label}: {'pass' if tree['tsc_entry']['checker_whole_program'] else 'fail'}"
                                        for label, tree in trees.items()))
    lines.append('tsc entry diagnostics: ' + '; '.join(f"{label}: {tree['tsc_entry']['whole_program_diagnostics']}"
                                                    for label, tree in trees.items()))
    lines += root_lines(trees)
    if mode == 'single':
        compilers = f"Both measured with Adamic {result['adamic_commit']} and ordinary stage 0 options."
    else:
        compilers = (f"Main measured with Adamic {trees['main']['adamic_commit']}; "
                     f"area measured with Adamic {trees['area']['adamic_commit']}. "
                     'Each compiler is built from its pinned ref with ordinary stage 0 options.')
    lines += ['', f'Stage 3 meter {stamp}: main {main_commit[:8]}', '',
              f'Main: origin/main at {main_commit}. Area: origin/area/stage3 at {area_commit}.',
              compilers, '',
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
    for label in trees:
        lines += entry_table(label, trees[label]['tsc_entry'])
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
