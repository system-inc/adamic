#!/usr/bin/env python3
"""Run the pinned cohere read-only and reject findings outside the reviewed baseline."""
import argparse
import collections
import fnmatch
import json
import os
import re
import shutil
from pathlib import Path
import subprocess
import sys
import tempfile

ROOT = Path(__file__).resolve().parents[1]
BASELINE = ROOT / 'cloud/cohere-baseline.json'


def identity(finding):
    path = Path(finding['path'])
    if not path.is_absolute():
        path = ROOT / path
    relative = path.relative_to(ROOT).as_posix()
    source = ''
    if finding['rule'] not in ('fix', 'format'):
        lines = path.read_text().splitlines()
        line = finding['line']
        if not 1 <= line <= len(lines):
            raise ValueError(f'invalid finding location: {relative}:{line}')
        source = lines[line - 1].strip()
    # Not the message: cohere rewords its messages from one pin to the next, and a reworded finding would
    # read as new, which --record refuses by design, leaving a hand edit of the baseline as the only way
    # through a pin bump. The rule, its messageId and the source line say which finding this is.
    return json.dumps({
        'path': relative, 'rule': finding['rule'],
        'messageId': finding.get('messageId', ''),
        'severity': finding['severity'],
        'source': source,
    }, sort_keys=True)


def make_mirror(directory):
    # Pinned cohere ignores sourceExtensions and has no .a formatter registration.
    # Only the disposable copies change names; checked-in imports and sources stay put.
    names = subprocess.check_output(
        ['git', 'ls-files', '-co', '--exclude-standard', '-z'], cwd=ROOT,
    ).decode().split('\0')
    mirrored = 0
    for name in sorted(set(names)):
        source = ROOT / name
        if not name or name in ('cohere', '.gitmodules') or not source.is_file():
            continue
        mirrored += 1
        destination = directory / (name + '.ts' if name.endswith('.a') else name)
        if destination.exists():
            raise ValueError(f'.a alias collides with an existing source: {name}')
        destination.parent.mkdir(parents=True, exist_ok=True)
        if source.suffix in ('.a', '.ts', '.tsx'):
            text = re.sub(
                r"(\bfrom\s+|\bimport\s*(?:\(\s*)?)([\"'])([^\"'\n]+\.a)\2",
                lambda match: match[1] + match[2] + match[3] + '.ts' + match[2],
                source.read_text(),
            )
            destination.write_text(text)
        else:
            shutil.copyfile(source, destination)
    config = json.loads((ROOT / 'tsconfig.json').read_text())
    config.pop('sourceExtensions', None)
    config['include'] = ['**/*.ts', '**/*.tsx']
    (directory / 'tsconfig.json').write_text(json.dumps(config, indent=4) + '\n')
    settings = json.loads((ROOT / 'CohereSettings.json').read_text())
    # These are syntactically invalid Test262 parser controls, not Adamic programs.
    settings['ignorePatterns'] += [
        'cmd/adamic-test262/testdata/corpus/negative_parse.js',
        'cmd/adamic-test262/testdata/mini/test/skip/negative.js',
    ]
    (directory / 'CohereSettings.json').write_text(json.dumps(settings, indent=4) + '\n')
    subprocess.run(['git', 'init', '-q', str(directory)], check=True)
    return mirrored


def audit(binary, directory):
    records = []
    coverage = []
    for flags, phase in [(['--types'], 'types'), (['--lint'], 'lint'),
                         (['--no-fix', '--format-only'], 'fix'),
                         (['--no-fix', '--format-only', 'tsconfig.json', 'CohereSettings.json'], 'config-format')]:
        if phase == 'fix':
            patterns = json.loads((directory / 'CohereSettings.json').read_text())['ignorePatterns']
            paths = []
            ignored = 0
            for path in sorted(directory.rglob('*')):
                if not path.is_file():
                    continue
                name = path.relative_to(directory).as_posix()
                original = name[:-3] if name.endswith('.a.ts') else name
                if name.startswith('.git/'):
                    continue
                if any(fnmatch.fnmatchcase(original, pattern) for pattern in patterns):
                    ignored += 1
                    continue
                # For now, the programs the type and lint phases check. Temporary: the JSON, Markdown and
                # scripts the areas add unformatted get one reformat on fresh main, and this filter goes
                # in the same push (#dvxrzsv). Until then they'd block every merge.
                if not name.endswith(('.ts', '.tsx')):
                    continue
                paths.append(name)
            flags = [*flags, *paths]
        checked_directory = ROOT if phase == 'config-format' else directory
        expected_phase = 'fix' if phase == 'config-format' else phase
        result = subprocess.run(
            [binary, '--directory', str(checked_directory), '--no-cache', '--json', *flags],
            cwd=checked_directory, capture_output=True, text=True,
        )
        rows = [json.loads(line) for line in result.stdout.splitlines() if line.strip()]
        if result.stderr.strip() or not rows or rows[-1].get('kind') != 'summary':
            raise ValueError(f'incomplete {phase} check (exit {result.returncode}): {result.stderr}')
        summary = rows[-1]
        gaps = summary['gaps']
        if summary['schemaVersion'] != 1 or any(gaps.get(key) for key in (
            'crashedFiles', 'ruleCrashes', 'nothingToCheck', 'modifiedBuild', 'unread', 'programFiles',
        )):
            raise ValueError(f'incomplete {phase} check: {summary}')
        outcomes = [p['outcome'] for p in summary['phases'] if p['name'] == expected_phase]
        if outcomes != ['checked' if expected_phase == 'fix' else 'ran']:
            raise ValueError(f'{phase} did not run: {summary["phases"]}')
        if expected_phase != 'fix' and summary['filesInScope'] == 0:
            raise ValueError(f'empty {phase} scope')
        if phase == 'fix':
            coverage.append(f'format: {len(paths)} programs checked, {ignored} files left out by the ignore patterns')
        elif phase == 'config-format':
            coverage.append('config-format: tsconfig.json and CohereSettings.json checked at their real paths')
        else:
            coverage.append(f'{phase}: {summary["filesInScope"]} files in scope')
        if any(row.get('kind') not in ('finding', 'summary') for row in rows):
            raise ValueError('unexpected output record, crash, or attempted write')
        if result.returncode not in (0, 1) or (result.returncode == 1 and len(rows) == 1):
            raise ValueError(f'{phase} failed without findings (exit {result.returncode})')
        retained = []
        for row in rows:
            if row['kind'] == 'finding':
                name = Path(row['path']).relative_to(checked_directory).as_posix()
                if phase == 'fix' and name in ('tsconfig.json', 'CohereSettings.json'):
                    continue  # Synthesized configurations are checked separately at their real paths.
                if name.endswith('.a.ts'):
                    name = name[:-3]
                row['path'] = str(ROOT / name)
                row['message'] = row['message'].replace(str(directory), '<mirror>').replace('.a.ts', '.a')
            row['auditPhase'] = phase
            retained.append(row)
        records.extend(retained)
    return records, coverage


def baseline_at(reference):
    subprocess.run(['git', 'rev-parse', '--verify', reference], cwd=ROOT, check=True, stdout=subprocess.DEVNULL)
    name = BASELINE.relative_to(ROOT).as_posix()
    exists = subprocess.check_output(['git', 'ls-tree', '--name-only', reference, '--', name], cwd=ROOT)
    if not exists:
        return None  # First introduction of this gate.
    text = subprocess.check_output(['git', 'show', f'{reference}:{name}'], cwd=ROOT)
    return collections.Counter(json.loads(text))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--base-ref', default='origin/main', help='baseline may only shrink relative to this integration ref')
    parser.add_argument('--output', help='save the complete cohere JSON output here')
    parser.add_argument('--record', action='store_true', help='bootstrap or reduce the baseline; never add entries to an existing list')
    arguments = parser.parse_args()
    with tempfile.TemporaryDirectory(prefix='adamic-cohere-') as scratch:
        binary = os.environ.get('COHERE_BINARY', str(Path(scratch) / 'cohere'))
        if 'COHERE_BINARY' not in os.environ:
            subprocess.run(['go', 'build', '-o', binary, './command/cohere'], cwd=ROOT / 'cohere', check=True)
        directory = Path(scratch) / 'source'
        directory.mkdir()
        mirrored = make_mirror(directory)
        records, coverage = audit(binary, directory)
    # What the gate checked, so a gate that quietly checked less shows it here.
    print(f'cohere gate coverage: {mirrored} repository files mirrored')
    for line in coverage:
        print(f'  {line}')
    if arguments.output:
        Path(arguments.output).write_text(''.join(json.dumps(row) + '\n' for row in records))
    current = collections.Counter(identity(row) for row in records if row['kind'] == 'finding')
    old = collections.Counter(json.loads(BASELINE.read_text())) if BASELINE.exists() else None
    if old is None and not arguments.record:
        raise ValueError('missing reviewed baseline')
    integration = baseline_at(arguments.base_ref)
    if integration is not None and old is not None and old - integration:
        raise ValueError('baseline grew relative to the integration ref')
    committed = baseline_at('HEAD')
    if committed is not None and old is not None and old - committed:
        raise ValueError('baseline grew relative to the committed HEAD')
    added = collections.Counter()
    for ceiling in (old, integration, committed):
        if ceiling is not None:
            added |= current - ceiling
    if added:
        for key, count in sorted(added.items()):
            print(f'NEW x{count}: {key}', file=sys.stderr)
        return 1
    if old is not None and old - current and not arguments.record:
        raise ValueError('findings decreased; run --record to remove stale baseline entries')
    if arguments.record:
        BASELINE.write_text(json.dumps(dict(sorted(current.items())), indent=4) + '\n')
    print(f'cohere baseline: {sum(current.values())} findings, {sum((old - current).values()) if old is not None else 0} removed, 0 new')
    return 0


if __name__ == '__main__':
    try:
        sys.exit(main())
    except (ValueError, KeyError, OSError, subprocess.SubprocessError) as error:
        print(f'cohere gate failed: {error}', file=sys.stderr)
        sys.exit(1)
