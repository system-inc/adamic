"""Source-bound function coverage and conservative differential selection."""
import gzip
import base64
import zlib
import hashlib
import io
import json
import os
from pathlib import Path
import subprocess
import tarfile
import tempfile
import time

ROOT = Path(__file__).resolve().parent
SAMPLE_SIZE = 8
SEED = 'step34-differential-v1'


def sha(raw):
    return hashlib.sha256(raw).hexdigest()


def identity(row):
    return row['source'] + '\0' + row.get('configuration', '')


def output_hash(wanted):
    return sha(b''.join(len(wanted[name]).to_bytes(8, 'big') + wanted[name]
                        for name in ('stdout', 'stderr', 'exit')))


def pack_functions(indices, count):
    bitmap = bytearray((count + 7) // 8)
    for index in indices:
        bitmap[index // 8] |= 1 << (index % 8)
    return base64.b64encode(zlib.compress(bytes(bitmap))).decode()


def covered_indices(row):
    if 'functions' in row:
        return row['functions']
    bitmap = zlib.decompress(base64.b64decode(row['functions_bitmap']))
    return [byte * 8 + bit for byte, value in enumerate(bitmap) for bit in range(8) if value & (1 << bit)]


def compact_map(mapping):
    count = len(mapping['inventory']['functions'])
    for row in mapping['configurations'].values():
        if 'functions' in row:
            row['functions_bitmap'] = pack_functions(row.pop('functions'), count)
    return mapping


def load_map(path):
    with gzip.open(path, 'rt') as stream:
        result = json.load(stream)
    checksum = result.pop('map_sha256', None)
    if checksum != sha(json.dumps(result, sort_keys=True, separators=(',', ':')).encode()):
        raise RuntimeError('coverage map checksum mismatch')
    return result


def save_map(path, result):
    result['map_sha256'] = sha(json.dumps(result, sort_keys=True, separators=(',', ':')).encode())
    with gzip.GzipFile(filename=str(path), mode='wb', mtime=0) as stream:
        stream.write(json.dumps(result, separators=(',', ':')).encode())


def inventory(source, output, reference=None):
    with (output.parent / (output.name + '.log')).open('wb') as log:
        argv = ['node', str(ROOT / 'function_map.cjs'), 'inventory', str(source), str(output)]
        if reference is not None:
            reference_file = output.parent / 'reference-inventory.json'
            reference_file.write_text(json.dumps(reference))
            argv.append(str(reference_file))
        subprocess.run(argv,
                       stdout=log, stderr=log, check=True)
    return json.loads(output.read_text())


def validate_source_map(mapping, source, base):
    # The map belongs to the base source, not the intentionally changed candidate.
    raw = subprocess.check_output(['git', '-C', str(source), 'archive', base, 'src/compiler', 'src/tsc'])
    files = {}
    with tarfile.open(fileobj=io.BytesIO(raw)) as archive:
        for member in archive:
            if member.isfile() and member.name.endswith('.ts') and not member.name.endswith('.d.ts'):
                files[member.name] = sha(archive.extractfile(member).read())
    recorded = mapping['inventory']['files']
    generated = 'src/compiler/diagnosticInformationMap.generated.ts'
    if generated in recorded:
        # This ignored file is part of tsc, but not a Git blob. Recreate it
        # with the base commit's own generator and relative input spelling.
        with tempfile.TemporaryDirectory(prefix='verdict-base-') as scratch:
            root = Path(scratch)
            for name in ('scripts/processDiagnosticMessages.mjs', 'src/compiler/diagnosticMessages.json'):
                target = root / name
                target.parent.mkdir(parents=True, exist_ok=True)
                target.write_bytes(subprocess.check_output(['git','-C',str(source),'show',base+':'+name]))
            with (root/'generator.log').open('wb') as log:
                subprocess.run(['node','scripts/processDiagnosticMessages.mjs','src/compiler/diagnosticMessages.json'],cwd=root,stdout=log,stderr=log,check=True)
            files[generated] = sha((root/generated).read_bytes())
    if files != {name: item['sha256'] for name, item in recorded.items()}:
        raise RuntimeError('stale coverage map: base source hash mismatch')
    aggregate = sha(json.dumps([[name, item['sha256']] for name, item in recorded.items()],
                               separators=(',', ':'), ensure_ascii=False).encode())
    if aggregate != mapping['inventory']['source_sha256']:
        raise RuntimeError('stale coverage map: source hash mismatch')


def validate_binding(binary, output, reference=None):
    path = Path(str(binary) + '.source.json')
    if not path.is_file():
        raise RuntimeError('binary source provenance missing: ' + str(path))
    binding = json.loads(path.read_text())
    if binding['binary_sha256'] != sha(binary.read_bytes()):
        raise RuntimeError('binary hash mismatch')
    for name, expected in binding.get('artifacts', {}).items():
        if sha(Path(name).read_bytes()) != expected:
            raise RuntimeError('binary dependency hash mismatch: ' + name)
    source = Path(binding['source_root'])
    current = inventory(source, output / 'candidate-inventory.json', reference)
    if current['source_sha256'] != binding['source_sha256']:
        raise RuntimeError('binary source hash mismatch')
    return source, current, binding


def select(mapping, current, rows):
    before = {item['id']: item for item in mapping['inventory']['functions']}
    after = {item['id']: item for item in current['functions']}
    changed = {key for key in before.keys() | after.keys()
               if key not in before or key not in after or before[key]['sha256'] != after[key]['sha256']}
    old_files, new_files = mapping['inventory']['files'], current['files']
    outside = [name for name in old_files.keys() | new_files.keys()
               if name not in old_files or name not in new_files or
               old_files[name]['outside_sha256'] != new_files[name]['outside_sha256']]
    changed_indices = {before[key]['index'] for key in changed if key in before}
    coverage = mapping['configurations']
    affected = []
    covered_changes = set()
    for row in rows:
        if not changed_indices:
            break
        record = coverage[identity(row)]
        if 'functions' in record:
            hits = changed_indices.intersection(record['functions'])
        else:
            bitmap = zlib.decompress(base64.b64decode(record['functions_bitmap']))
            hits = {index for index in changed_indices if bitmap[index // 8] & (1 << (index % 8))}
        if hits:
            affected.append(row)
            covered_changes.update(hits)
    affected_ids = {identity(row) for row in affected}
    rest = [row for row in rows if identity(row) not in affected_ids]
    sample = sorted(rest, key=lambda row: sha((SEED + identity(row)).encode()))[:SAMPLE_SIZE]
    gaps = sorted(key for key in changed if key not in before or before[key]['index'] not in covered_changes)
    return affected, sample, {'changed_functions': sorted(changed), 'outside_function_changes': sorted(outside),
                              'coverage_gaps': gaps, 'sample_seed': SEED, 'sample_size': len(sample)}


def record_coverage(binary, source, destination, output, resume=None):
    from run import baseline_suite, upstream_tree, write_json
    from cases import expected_exit
    source = source.resolve() if source else upstream_tree(output)
    data = json.loads((binary.parent / 'inventory.json').read_text())
    current = inventory(source, output / 'source-inventory.json')
    if current != data:
        raise RuntimeError('instrumented compiler/source inventory mismatch')
    selection = json.loads((ROOT / 'selection.json').read_text())
    observations = {}
    def observe(row, folder, cwd, wanted, differences):
        capture = cwd / '.verdict-coverage.json'
        if not capture.exists():
            capture = cwd / '.verdict-capture/function-coverage.json'
        covered = json.loads(capture.read_text())
        if not covered or any(type(index) is not int or index < 0 or index >= len(data['functions']) for index in covered):
            raise RuntimeError('invalid or empty function coverage')
        observations[identity(row)] = {'functions': covered, 'expected_output_sha256': output_hash(wanted),
                                       'expected_stdout_sha256': row['expected_sha256']}
    resumed = 0
    if resume:
        resume = resume.resolve()
        previous = json.loads((resume / 'source-inventory.json').read_text())
        if previous != current:
            raise RuntimeError('stale resumed coverage: source inventory mismatch')
        for index, row in enumerate(selection['cases'], 1):
            folder = resume / 'baselines' / f'{index:05d}_{Path(row["source"]).stem}'
            if not (folder / 'working-directory.json').exists():
                continue
            cwd = Path(json.loads((folder / 'working-directory.json').read_text()))
            capture = cwd / '.verdict-coverage.json'
            if not capture.exists():
                capture = cwd / '.verdict-capture/function-coverage.json'
            if not capture.exists() or not (folder / 'actual.diagnostics').exists():
                continue
            wanted_stdout = (folder / 'expected.stdout').read_bytes()
            if sha(wanted_stdout) != row['expected_sha256']:
                raise RuntimeError('stale resumed expected output')
            if sha((source / row['source']).read_bytes()) != row['source_sha256']:
                raise RuntimeError('stale resumed input')
            if row['baseline'] and sha((source / row['baseline']).read_bytes()) != row['baseline_sha256']:
                raise RuntimeError('stale resumed baseline')
            wanted = {'stdout':wanted_stdout, 'stderr':b'', 'exit':f"{expected_exit(row.get('effective_options',row['options']),wanted_stdout)}\n".encode()}
            if ((folder / 'actual.diagnostics').read_bytes() != wanted['stdout'] or
                (folder / 'actual.stderr').read_bytes() != wanted['stderr'] or
                (folder / 'actual.exit').read_bytes() != wanted['exit']):
                raise RuntimeError('resumed capture does not pass its independent reference')
            observe(row, folder, cwd, wanted, {})
            resumed += 1
        write_json(output / 'resume.json', {'source':str(resume), 'reused_passing_configurations':resumed,
                                          'source_sha256':current['source_sha256']})
    remaining = [row for row in selection['cases'] if identity(row) not in observations]
    from coverage_worker import Workers
    with Workers(binary.parent / 'server/tsc.cjs', output / 'workers') as workers:
        report = baseline_suite(binary, source, output / 'baselines', None, selected_rows=remaining, observer=observe, command_runner=workers.execute)
    report = {**report, 'measured_this_invocation':report['total'], 'resumed':resumed,
              'seconds_scope':'new configurations in this invocation',
              'total':report['total']+resumed, 'passed':report['passed']+resumed, 'deferred':0}
    write_json(output / 'coverage-report.json', report)
    if report['failed']:
        raise RuntimeError('instrumented Node differs from pinned reference baselines')
    if len(observations) != len(selection['cases']):
        raise RuntimeError('coverage configuration population mismatch')
    mapping = {'schema':1, 'inventory':data, 'selection_sha256':sha((ROOT / 'selection.json').read_bytes()),
               'configurations':observations, 'node_version':subprocess.check_output(['node','--version'],text=True).strip(),
               'instrumented_binary_sha256':sha(binary.read_bytes()),
               'instrumented_cli_sha256':sha((binary.parent/'tsc.cjs').read_bytes()),
               'worker_cli_sha256':sha((binary.parent/'server/tsc.cjs').read_bytes()),
               'module_state':'fresh per configuration', 'resumed_configurations':resumed}
    save_map(destination,compact_map(mapping))
    print(f'coverage: {len(observations)} configurations; {resumed} resumed; {report["measured_this_invocation"]} new in {report["seconds"]} seconds',flush=True)
    return 0


def differential(binary, base, map_path, output):
    from run import baseline_suite, upstream_tree, write_json
    started = time.monotonic()
    mapping = load_map(map_path)
    if mapping['selection_sha256'] != sha((ROOT / 'selection.json').read_bytes()):
        raise RuntimeError('stale coverage map: selection hash mismatch')
    source, current, binding = validate_binding(binary,output,mapping['inventory'])
    validate_source_map(mapping,source,base)
    selection = json.loads((ROOT / 'selection.json').read_text())
    if set(mapping['configurations']) != {identity(row) for row in selection['cases']}:
        raise RuntimeError('coverage configuration population mismatch')
    for row in selection['cases']:
        if mapping['configurations'][identity(row)]['expected_stdout_sha256'] != row['expected_sha256']:
            raise RuntimeError('stale coverage map: expected output hash mismatch')
    affected, sample, details = select(mapping,current,selection['cases'])
    selected = affected + sample
    tree = upstream_tree(output)
    def observe(row, folder, cwd, wanted, differences):
        if output_hash(wanted) != mapping['configurations'][identity(row)]['expected_output_sha256']:
            raise RuntimeError('stale coverage map: expected streams hash mismatch')
    report = baseline_suite(binary,tree,output/'baselines',None,selected_rows=selected,observer=observe)
    # Recheck artifact binding after execution, rejecting source/binary movement.
    if sha(binary.read_bytes()) != binding['binary_sha256'] or any(sha(Path(name).read_bytes()) != expected for name,expected in binding.get('artifacts',{}).items()):
        raise RuntimeError('compiler artifacts changed during differential')
    gap = bool(details['coverage_gaps'] or details['outside_function_changes'])
    result = {**details, 'schema':1, 'mode':'differential', 'base':base, 'tsc':str(binary),
              'affected':len(affected), 'sampled':len(sample), 'selected':len(selected),
              'deferred':len(selection['cases'])-len(selected), 'baseline_report':report,
              'seconds':round(time.monotonic()-started,3), 'success':not gap and not report['failed'],
              'scope':'13,693 baseline configurations; acceptance and tiny suites are deferred'}
    write_json(output/'differential.json',result)
    print(json.dumps({key:result[key] for key in ('affected','sampled','selected','deferred','seconds','success','coverage_gaps','outside_function_changes')}),flush=True)
    return 2 if gap else 1 if report['failed'] else 0
