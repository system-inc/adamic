"""Publish milestones only from retained full-output comparisons and mutants."""
import argparse
import copy
import hashlib
import json
import os
from pathlib import Path
import re
import tempfile


SUPPORTED = {'adamic_sha', 'source_sha', 'run_directory', 'node_sha256',
             'native_sha256', 'mutant'}


def read(path):
    return json.loads(path.read_text())


def artifact(directory, name):
    if not isinstance(name, str) or not name:
        raise ValueError('missing artifact path')
    path = (directory / name).resolve()
    path.relative_to(directory.resolve())
    if not path.is_file():
        raise ValueError('missing artifact: ' + name)
    return path


def run_path(repository, name):
    if not isinstance(name, str) or not name:
        raise ValueError('missing run_directory')
    path = (repository / name).resolve()
    path.relative_to(repository.resolve())
    if not path.is_dir():
        raise ValueError('missing run directory: ' + name)
    return path


def sha256(path):
    digest = hashlib.sha256()
    with path.open('rb') as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b''):
            digest.update(chunk)
    return digest.hexdigest()


def required_fields(progress):
    fields = set()
    for description in progress['evidence_required']:
        fields.update(description.split(':', 1)[0].split(' and '))
    if fields != SUPPORTED:
        raise ValueError('unsupported evidence_required fields: ' + repr(sorted(fields)))
    return fields


def comparison(directory, record, expected):
    if not isinstance(record, dict) or type(record.get('exit')) is not int or record['exit'] != expected:
        raise ValueError('missing comparison with exit ' + str(expected))
    artifact(directory, record.get('log'))


def outputs(repository, record):
    for name in ('adamic_sha', 'source_sha'):
        if not isinstance(record.get(name), str) or not re.fullmatch('[0-9a-f]{40}', record[name]):
            raise ValueError('missing or invalid ' + name)
    directory = run_path(repository, record.get('run_directory'))
    paths = [artifact(directory, record.get(name + '_output')) for name in ('node', 'native')]
    hashes = [sha256(path) for path in paths]
    for name, digest in zip(('node', 'native'), hashes):
        if record.get(name + '_sha256') != digest:
            raise ValueError(name + '_sha256 does not cover the full output')
    return directory, paths, hashes


def changed_bytes(left, right):
    if left.stat().st_size != right.stat().st_size:
        return -1
    count = 0
    with left.open('rb') as a, right.open('rb') as b:
        while chunk := a.read(1024 * 1024):
            count += sum(x != y for x, y in zip(chunk, b.read(len(chunk))))
            if count > 1:
                return count
    return count


def validate_evidence(repository, record, required):
    try:
        for field in sorted(required):
            if field not in record:
                raise ValueError('missing ' + field)
        directory, paths, hashes = outputs(repository, record)
        if hashes[0] != hashes[1]:
            raise ValueError('full Node and native outputs differ')
        comparison(directory, record.get('comparison'), 0)
        mutant = record['mutant']
        if not isinstance(mutant, dict):
            raise ValueError('missing native-output mutant')
        mutated = artifact(directory, mutant.get('output'))
        if changed_bytes(paths[1], mutated) != 1:
            raise ValueError('mutant must change exactly one native-output byte')
        comparison(directory, mutant.get('comparison'), 1)
        if sha256(mutated) == hashes[0]:
            raise ValueError('comparison did not catch native-output mutant')
        return True, 'full outputs match; one-byte native-output mutant caught'
    except (ValueError, KeyError, TypeError) as error:
        return False, str(error)


def validate_regression(repository, record):
    try:
        directory, _, hashes = outputs(repository, record)
        comparison(directory, record.get('comparison'), 1)
        if hashes[0] == hashes[1]:
            raise ValueError('regression outputs still match')
        return True, 'regression: full Node and native outputs differ'
    except (ValueError, KeyError, TypeError) as error:
        return False, str(error)


def resolve(repository, milestone, record, seen=None):
    if 'evidence_file' not in record:
        return record
    directory = run_path(repository, record.get('run_directory'))
    path = artifact(directory, record['evidence_file'])
    seen = set() if seen is None else seen
    if path in seen:
        raise ValueError('cyclic milestone evidence reference')
    seen.add(path)
    return resolve(repository, milestone, read(path)[milestone], seen)


def blocked_reason(repository, record):
    directory = run_path(repository, record.get('run_directory'))
    report = read(artifact(directory, record['blocked_report']))
    if report.get('native') != 'blocked at compilation' or not report.get('build_exit'):
        raise ValueError('blocked report does not show native build failure')
    log = artifact(directory, record['build_log'])
    stop = next((line.strip() for line in log.read_text().splitlines() if line.strip()), None)
    if stop is None:
        raise ValueError('native build log has no first stop')
    stop = re.sub(r'^adamic: .*?/src/compiler/', '', stop)
    return 'native build stops before emitting: ' + stop + ' (run ' + record['run_directory'] + ')'


def write_json(path, data):
    path.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.NamedTemporaryFile(mode='w', dir=path.parent, delete=False) as stream:
        temporary = Path(stream.name)
        json.dump(data, stream, indent=2)
        stream.write('\n')
    os.replace(temporary, path)


def update(repository, run, meter_exit=0):
    repository, run = repository.resolve(), run.resolve()
    progress = copy.deepcopy(read(repository / 'stage3/progress.json'))
    required = required_fields(progress)
    current = read(run / 'milestones.json') if (run / 'milestones.json').exists() else {}
    catalog = repository / 'stage3/meter/progress-inputs.json'
    registered = read(catalog) if catalog.exists() else {}
    names = set(progress['milestones'])
    if any(type(value) is not bool for value in progress['milestones'].values()):
        raise ValueError('milestones must be Boolean')
    if set(current) - names or set(registered) - names:
        raise ValueError('unknown milestone in evidence inputs')
    reasons = progress.setdefault('reasons', {})
    observations = progress.setdefault('observations', {})
    progress['last_meter_run'] = os.path.relpath(run, repository)
    progress['last_meter_exit'] = meter_exit
    for name in progress['milestones']:
        candidate = current.get(name, registered.get(name))
        accepted = False
        regression = False
        reason = 'no complete native-output comparison and mutant evidence in this run'
        record = None
        if candidate is not None:
            try:
                record = resolve(repository, name, candidate)
                if 'blocked_report' in record:
                    reason = blocked_reason(repository, record)
                    if name in current and record.get('status') == 'regression':
                        for field in ('adamic_sha', 'source_sha'):
                            if not isinstance(record.get(field), str) or not re.fullmatch('[0-9a-f]{40}', record[field]):
                                raise ValueError('missing or invalid ' + field)
                        regression = True
                        reason = 'regression: ' + reason
                elif name in current and record.get('status') == 'regression':
                    regression, reason = validate_regression(repository, record)
                else:
                    accepted, reason = validate_evidence(repository, record, required)
                observations[name] = dict(record, validation=reason)
            except (ValueError, KeyError, TypeError) as error:
                reason = str(error)
        if accepted:
            progress['milestones'][name] = True
            progress['evidence'][name] = record
            reason += ' (run ' + record['run_directory'] + ')'
        elif regression:
            progress['milestones'][name] = False
            progress['evidence'][name] = record
            reason += ' (run ' + record['run_directory'] + ')'
        elif progress['milestones'][name]:
            previous = progress['evidence'].get(name) or {}
            reason = 'retained prior proof from ' + previous.get('run_directory', 'previous run') + '; no verified regression in this run'
        reasons[name] = reason
    # A run carries its own immutable snapshot; stage3/progress.json is the latest.
    write_json(run / 'progress.json', progress)
    write_json(repository / 'stage3/progress.json', progress)
    print(json.dumps({'milestones': progress['milestones'], 'reasons': reasons}, indent=2))
    return progress


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('repository', type=Path)
    parser.add_argument('run', type=Path)
    parser.add_argument('--meter-exit', type=int, default=0)
    arguments = parser.parse_args()
    update(arguments.repository, arguments.run, arguments.meter_exit)
