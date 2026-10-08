#!/usr/bin/env python3
"""Collect stopping evidence without copying any modified source tree."""
import argparse
import bisect
import json
from pathlib import Path
import re
import shutil

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('progress', type=Path)
args = parser.parse_args()
here = Path(__file__).resolve().parent
repository = here.parents[2]
evidence = here / 'evidence'
progress = args.progress.resolve()
records = json.loads((progress / 'stops.json').read_text())
run = sorted((repository / 'stage3/meter/runs').glob('*/report.json'))[-1]
meter = json.loads(run.read_text())['trees']['main']
reasons = meter['latent_lowering']['per_reason']
pristine = (evidence / 'build-0.stderr').read_text()
provenance = json.loads((evidence / 'provenance.json').read_text())
original_tree = Path(provenance['tree'])
maps = {}
for record in records:
    file = record['file']
    if file not in maps:
        raw = (original_tree / file).read_bytes().decode('utf8').encode('utf-16-le')
        maps[file] = [raw, list(range(len(raw) // 2)), raw]
    current, mapping, original = maps[file]
    line_starts = [0] + [i + 1 for i in range(len(current) // 2) if current[2*i:2*i+2] == b'\n\x00']
    offset = line_starts[record['line'] - 1] + record['column'] - 1
    original_offset = mapping[offset]
    original_starts = [0] + [i + 1 for i in range(len(original) // 2) if original[2*i:2*i+2] == b'\n\x00']
    original_line = bisect.bisect_right(original_starts, original_offset) - 1
    record['original_line'] = original_line + 1
    record['original_column'] = original_offset - original_starts[original_line] + 1
    ordinal = record['ordinal']
    probe = next((here / 'probes').glob(f'{ordinal:02d}-*.a'))
    stem = probe.stem
    record['probe'] = str(probe.relative_to(here))
    record['phase'] = 'checker' if record['message'].startswith('error TS') else 'lowering'
    text = re.sub(r'^error TS\d+: ', '', record['message'])
    record['meter_lowering_message_matches'] = [reason for reason in reasons
        if re.sub(r'^(NotYet|Refused): ', '', reason) == text]
    record['message_already_in_pristine_diagnostics'] = record['message'] in pristine
    record['node_exit'] = int((evidence / f'{stem}-node.exit').read_text())
    record['node_stdout'] = (evidence / f'{stem}-node.stdout').read_text()
    diagnostic = (evidence / f'{stem}-build.stderr').read_text().splitlines()[0]
    record['probe_message'] = re.sub(r'^.+:\d+:\d+: ', '', diagnostic)
    record['probe_exact_message_match'] = record['message'] == record['probe_message']
    replacement = progress / f'{ordinal:02d}-replacement.json'
    if replacement.exists():
        data = json.loads(replacement.read_text())
        record['replaced_function'] = data['function']
        start, end = data['start'], data['end']
        if current[start * 2:end * 2].decode('utf-16-le') != data['original']:
            raise RuntimeError('replacement does not match replayed source')
        replacement_bytes = data['replacement'].encode('utf-16-le')
        replacement_map = [mapping[start]] * (len(replacement_bytes) // 2)
        maps[file][0] = current[:start * 2] + replacement_bytes + current[end * 2:]
        mapping[start:end] = replacement_map
    for split in (0, 1):
        for suffix in ('stdout', 'stderr', 'exit'):
            shutil.copyfile(progress / f'{ordinal:02d}-split-{split}.{suffix}',
                evidence / f'{ordinal:02d}-split-{split}.{suffix}')
    if replacement.exists():
        shutil.copyfile(replacement, evidence / replacement.name)
(evidence / 'stops.json').write_text(json.dumps(records, indent=2) + '\n')
summary = {'meter_run': str(run.relative_to(repository)), 'meter_compiler': meter['adamic_commit'],
    'matched_stop_sites': sum(bool(record['meter_lowering_message_matches']) for record in records),
    'unmatched_stop_sites': sum(not record['meter_lowering_message_matches'] for record in records),
    'unmatched_distinct_messages': len({record['message'] for record in records if not record['meter_lowering_message_matches']}),
    'new_observed_lowering_reasons': len({record['message'] for record in records if record['phase'] == 'lowering' and not record['meter_lowering_message_matches']}),
    'checker_stop_sites': sum(record['phase'] == 'checker' for record in records),
    'not_in_pristine_diagnostics': [record['ordinal'] for record in records if not record['message_already_in_pristine_diagnostics']]}
(evidence / 'meter-comparison.json').write_text(json.dumps(summary, indent=2) + '\n')
print(json.dumps(summary))
