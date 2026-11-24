#!/usr/bin/env python3
"""The daily meter's committed record: its summary and the rows readers read, never its raw files.

    python3 -I slim.py <run-directory> <out-directory> [--job J-xxxxxxx]

It reads a finished meter run, and never writes to it, and writes into <out-directory>, which must not exist yet,
what submit.sh commits under stage3/meter/runs/<stamp>/ (#jnx6n47, Kirk's ruling of Nov 24): README.md and report.md,
main/report.md and area/report.md when the run has them, a slim report.json holding every key a reader of a committed
record reads, and kept.json, which names every file the run held that the commit leaves out, with its bytes and
sha256 and the job that kept it (null without --job), so every number quoted has a commit behind it. The raw files
stay with the job: its stdout kept by hash, and the run's tarball on the box.

It refuses, exiting 1 and writing nothing, when the run has no report.md or report.json, when report.json lacks a key
the record keeps, or when the record would hold a `.log` or `.log.gz` file or any file over 1 MB.
"""
import hashlib
import json
import os
from pathlib import Path
import re
import sys

# A committed record's largest file: a slim report.json past it means the summary has grown bulk again.
LIMIT = 1_048_576
# The Markdown a record commits as it is, each small. report.md is required, as submit.sh requires it.
MARKDOWN = ('README.md', 'report.md', 'main/report.md', 'area/report.md')
# Every key of report.json a reader of a committed record reads, True for the value whole, and who reads it:
#   report.py's write_readme sets each day's README beside the latest earlier record, reading trees.main's
#     tree_commit, its totals (checker_whole_program, checker_own_file, lowering, source_files) and
#     tsc_entry.whole_program_diagnostics, and only a report holding trees counts as a record;
#   stage3/drivers/tsc-entry/collect.py reads trees.main.adamic_commit and trees.main.latent_lowering.per_reason;
#   report_test.py's recorded-family tests read the root adamic_commit, and relation_report_complete, totals and
#     reason_rows of trees.main.latent_lowering and of trees.main.tsc_entry.lowering_census (null when the checker
#     blocked the entry, and kept as null);
#   outcome.py's verdict reads relation_budget, and job.sh's checked line the root totals and latent_lowering.totals
#     beside it, so both read the same from the commit as from the run;
#   timestamp_utc names the record.
# Left out: trees.area (the main measurement again), per-file rows, unit times, readiness probes, site lists and the
# unowned tables, which report.md shows and reason_rows' owners rebuild.
KEPT = {
    'timestamp_utc': True,
    'adamic_commit': True,
    'totals': True,
    'latent_lowering': {'totals': True},
    'relation_budget': True,
    'trees': {'main': {
        'tree_commit': True,
        'adamic_commit': True,
        'totals': True,
        'latent_lowering': {'relation_report_complete': True, 'totals': True, 'per_reason': True, 'reason_rows': True},
        'tsc_entry': {
            'whole_program_diagnostics': True,
            'lowering_census': {'relation_report_complete': True, 'totals': True, 'reason_rows': True},
        },
    }},
}


def kept(report, spec, where):
    """The keys spec names, from report, refusing one that is missing."""
    if not isinstance(report, dict):
        raise ValueError(f'{where} is not an object, so the record would lose what its readers read')
    result = {}
    for key, inner in spec.items():
        if key not in report:
            raise ValueError(f'{where} has no {key}, which a committed record keeps for its readers')
        value = report[key]
        result[key] = value if inner is True or value is None else kept(value, inner, f'{where}.{key}')
    return result


def encode(value, indent=''):
    """JSON indented as report.py writes it, but an object of scalars inside a list on one line: a reason row is one
    line of the diff, and the Nov 10 record's 1,319 reason rows in each census fit under the limit, where its slim report.json
    indented whole came to 1.08 MB."""
    inner = indent + '  '
    if isinstance(value, dict) and value:
        fields = [f'{json.dumps(key)}: {encode(item, inner)}' for key, item in value.items()]
        return '{\n' + ',\n'.join(inner + field for field in fields) + '\n' + indent + '}'
    if isinstance(value, list) and value:
        rows = [json.dumps(item) if isinstance(item, dict) and not any(isinstance(field, (dict, list)) for field in item.values())
                else encode(item, inner) for item in value]
        return '[\n' + ',\n'.join(inner + row for row in rows) + '\n' + indent + ']'
    return json.dumps(value)


def sha256(path):
    digest = hashlib.sha256()
    with open(path, 'rb') as source:
        for block in iter(lambda: source.read(1 << 20), b''):
            digest.update(block)
    return digest.hexdigest()


def held(run):
    """Every file under run, by its path relative to run, sorted; a symbolic link is named with its target."""
    entries = []
    for directory, directories, files in os.walk(run):
        directories.sort()
        # os.walk never descends a linked directory, so it is named here rather than lost.
        for name in [*(name for name in directories if os.path.islink(os.path.join(directory, name))), *files]:
            path = os.path.join(directory, name)
            relative = os.path.relpath(path, run).replace(os.sep, '/')
            if os.path.islink(path):
                entries.append({'path': relative, 'symlink': os.readlink(path)})
            else:
                entries.append({'path': relative, 'bytes': os.path.getsize(path), 'sha256': sha256(path)})
    return sorted(entries, key=lambda entry: entry['path'])


def record(run, job):
    """The committed record, each file's relative path to its bytes, and how many files it leaves out; refused whole
    if any file may not be committed."""
    for name in ('report.md', 'report.json'):
        if not (run / name).is_file():
            raise ValueError(f'{run} holds no {name}, so it is no record')
    with open(run / 'report.json') as source:
        report = json.load(source)
    files = {name: (run / name).read_bytes() for name in MARKDOWN if (run / name).is_file()}
    files['report.json'] = (encode(kept(report, KEPT, 'report.json')) + '\n').encode()
    # The full report.json is left out too: its slim copy keeps the same name.
    left = [entry for entry in held(run) if entry['path'] not in files or entry['path'] == 'report.json']
    files['kept.json'] = (encode({'job': job, 'files': left}) + '\n').encode()
    refused = [f'{name} is a log' for name in files if name.endswith(('.log', '.log.gz'))]
    refused += [f'{name} is {len(data):,} bytes, past {LIMIT:,}' for name, data in files.items() if len(data) > LIMIT]
    if refused:
        raise ValueError('the record may not be committed: ' + '; '.join(refused))
    return files, len(left)


def main(arguments):
    job = None
    if len(arguments) == 4 and arguments[2] == '--job' and re.fullmatch(r'J-[0-9a-z]+', arguments[3]):
        job = arguments[3]
    elif len(arguments) != 2:
        print('usage: slim.py <run-directory> <out-directory> [--job J-xxxxxxx]', file=sys.stderr)
        return 2
    run, out = Path(arguments[0]), Path(arguments[1])
    if os.path.lexists(out):
        print(f'slim: {out} is there already; it must not exist yet', file=sys.stderr)
        return 1
    try:
        files, left = record(run, job)
    except (OSError, ValueError) as error:
        print(f'slim: {error}. Nothing written.', file=sys.stderr)
        return 1
    for name, data in files.items():
        (out / name).parent.mkdir(parents=True, exist_ok=True)
        (out / name).write_bytes(data)
    print(f"slim: {out} holds {', '.join(files)}, {sum(map(len, files.values())):,} bytes; kept.json names the {left} "
          'files left with the job')
    return 0


if __name__ == '__main__':
    sys.exit(main(sys.argv[1:]))
