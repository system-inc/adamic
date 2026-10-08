#!/usr/bin/env python3
"""Measure every newly eligible configuration and kill one byte mutant per group."""
import argparse
import collections
import json
import os
from pathlib import Path
import subprocess
import sys
from run import ROOT, baseline_suite, write_json


def group(reason):
    if reason.startswith('diagnostics outside'):
        return 'library-placeholders'
    if reason.startswith('absolute source'):
        return 'absolute-source'
    if reason.startswith('absolute package'):
        return 'absolute-package'
    if '@suppressoutputpathcheck' in reason:
        return 'output-path'
    if reason.startswith('Windows'):
        return 'drive-paths'
    if reason.startswith('case-insensitive'):
        return 'case-host'
    if reason.startswith('filename:'):
        return 'filename'
    if reason.startswith('option variants') or reason.startswith('variant errors'):
        return 'option-variants'
    if reason.startswith('declaration') or reason.startswith('TS18027'):
        return 'declaration-emit'
    if reason.startswith('references'):
        return 'references'
    if reason.startswith('global'):
        return 'global-diagnostics'
    if reason.startswith('syntax') or reason.startswith('stock parser'):
        return 'syntax-only'
    return 'compiler-directives'


def subset(manifest, rows):
    return {**manifest, 'cases': rows, 'selected': len({r['source'] for r in rows}),
            'configurations': len(rows), 'excluded': 0, 'exclusions': [],
            'total': len({r['source'] for r in rows}), 'reasons': {}, 'configuration_exclusions': []}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('old', type=Path)
    parser.add_argument('new', type=Path)
    parser.add_argument('tree', type=Path)
    parser.add_argument('output', type=Path)
    parser.add_argument('--group')
    args = parser.parse_args()
    old, new = [json.loads(p.read_text()) for p in (args.old, args.new)]
    old_cases = {(r['source'], r.get('configuration', '')) for r in old['cases']}
    old_reasons = {r['source']: r['reason'] for r in old['exclusions']}
    old_config_reasons = {(r['source'], r['configuration']): r['reason']
                         for r in old.get('configuration_exclusions', [])}
    groups = collections.defaultdict(list)
    for row in new['cases']:
        identity = (row['source'], row.get('configuration', ''))
        if identity not in old_cases:
            reason = old_reasons.get(row['source']) or old_config_reasons.get(identity, 'additional configuration')
            groups['library-placeholders' if row.get('library_placeholders') else group(reason)].append(row)
    if args.group:
        groups = {args.group: groups[args.group]}
    args.output.mkdir(parents=True)
    evidence = []
    for name, rows in sorted(groups.items(), key=lambda item: -len(item[1])):
        print(name, 'inputs', len({r['source'] for r in rows}), 'configurations', len(rows), flush=True)
        folder = args.output / name
        folder.mkdir()
        manifest = subset(new, rows)
        write_json(folder / 'selection.json', manifest)
        original = baseline_suite(ROOT / 'standins/node.sh', args.tree.resolve(), folder / 'A', None, manifest)
        print(name, 'A', original['passed'], '/', original['total'], flush=True)
        if original['failed']:
            raise RuntimeError(f'{name}: A has {original["failed"]} differences; see {folder}/A/report.json')
        target = next((r for r in rows if r['baseline']), None)
        target_tree = args.tree.resolve()
        target_manifest = new
        if target is None or name == 'case-host':
            if name != 'case-host':
                raise RuntimeError(f'{name}: no diagnostic case to mutate')
            from census import digest, summary_bytes
            from cases import parse, configurations
            target_tree = ROOT / 'fixtures/case-host'
            raw = (target_tree / 'input.a').read_bytes()
            units, settings, roots = parse(raw, 'host.ts')
            options = configurations(settings)[0][1]
            target = {'source': 'input.a', 'name': 'host.ts', 'configuration': '', 'options': options, 'effective_options': options, 'project': None, 'baseline': 'reference.txt', 'source_sha256': digest(raw), 'baseline_sha256': digest((target_tree / 'reference.txt').read_bytes()), 'expected_sha256': digest(summary_bytes((target_tree / 'reference.txt').read_bytes()))}
            target_manifest = subset(new, [target])
            fixture = baseline_suite(ROOT / 'standins/node.sh', target_tree, folder / 'fixture-A', None, target_manifest)
            if fixture['passed'] != 1:
                raise RuntimeError('case-host diagnostic fixture does not pass A')
        # A single new diagnostic configuration proves this group's comparison.
        mutated = baseline_suite(ROOT / 'standins/mutant.sh', target_tree, folder / 'B', None,
                                 subset(target_manifest, [target]))
        if mutated['failed'] != 1 or set(mutated['failures'][0]['differences']) != {'stdout'}:
            raise RuntimeError(f'{name}: diagnostic byte mutant escaped or hit another check')
        capture = next((folder / 'B').glob('00001_*'))
        if target in rows:
            original_index = rows.index(target) + 1
            before = next((folder / 'A').glob(f'{original_index:05d}_*'))
        else:
            before = next((folder / 'fixture-A').glob('00001_*'))
        # Absolute scratch roots differ; compare formatted diagnostic captures.
        a, b = [(p / 'actual.diagnostics').read_bytes() for p in (before, capture)]
        if len(a) != len(b) or sum(x != y for x, y in zip(a, b)) != 1:
            raise RuntimeError('mutant did not change exactly one diagnostic byte')
        argv = json.loads((capture / 'command.json').read_text())
        cwd = Path(json.loads((capture / 'working-directory.json').read_text()))
        with (capture / 'control.stdout').open('wb') as stdout, (capture / 'control.stderr').open('wb') as stderr:
            control_argv = list(argv)
            control_argv[control_argv.index(str(ROOT / 'standins/mutant.sh'))] = str(ROOT / 'standins/node.sh')
            completed = subprocess.run(control_argv, cwd=cwd,
                                       stdout=stdout, stderr=stderr, timeout=60)
        raw_a = (capture / 'control.stdout').read_bytes()
        raw_b = (capture / 'actual.stdout').read_bytes()
        if (len(raw_a) != len(raw_b) or sum(x != y for x, y in zip(raw_a, raw_b)) != 1
                or (capture / 'control.stderr').read_bytes() != (capture / 'actual.stderr').read_bytes()
                or f'{completed.returncode}\n' != (capture / 'actual.exit').read_text()):
            raise RuntimeError('raw mutant witness changed more than one stdout byte')
        evidence.append({'group': name, 'inputs': manifest['selected'], 'configurations': len(rows),
                         'A_passed': original['passed'], 'A_failed': 0, 'B_total': 1, 'B_failed': 1,
                         'fixture_A_passed': 1 if target not in rows else None,
                         'mutant_source': target['source'], 'mutant_configuration': target['configuration'],
                         'catch': 'stdout diagnostic bytes only, exactly one byte changed'})
        write_json(args.output / 'proof.json', {'groups': evidence, 'proved': True})
        print(name, 'B 0/1: one diagnostic byte caught', flush=True)
    print('all group assertions passed', flush=True)


if __name__ == '__main__':
    main()
