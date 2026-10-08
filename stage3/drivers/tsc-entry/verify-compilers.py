#!/usr/bin/env python3
"""Recompute the compiler comparison from retained build and snapshot streams."""
import json
from pathlib import Path
import re
import sys

here = Path(__file__).resolve().parent
baseline = here / 'evidence/main-efe9f404'
old = json.loads((baseline / 'stops.json').read_text())
hashes = json.loads((baseline / 'provenance.json').read_text())['sources']
pins = {'area-b68b2fe1': 'b68b2fe1af2dd6ad369ad8dfbfedf66be5a42861',
        'front3-a36d1c04': 'a36d1c0472649ef2ee549cc2505b93a84e893b59'}


def check(root):
    for name, pin in pins.items():
        directory = root / name
        report = json.loads((directory / 'comparison.json').read_text())
        assert report['compiler_commit'] == pin, 'compiler pin'
        binary = (directory / 'binary.log').read_text()
        assert f'vcs.revision={pin}\n' in binary and 'vcs.modified=false\n' in binary, 'binary provenance'
        assert report['source_hashes'] == hashes, 'adapted source hashes'
        assert len(report['stops']) == 15, 'stop population'
        streams = [(directory / f'build-{split}.stderr').read_bytes() for split in (0, 1)]
        exits = [int((directory / f'build-{split}.exit').read_text()) for split in (0, 1)]
        assert streams[0] == streams[1] and report['split_streams_equal'], 'split byte comparison'
        assert (directory / 'build-0.stdout').read_bytes() == (directory / 'build-1.stdout').read_bytes(), 'split stdout comparison'
        assert exits == [report['split_0_exit'], report['split_1_exit']], 'build exits'
        assert report['first_stop'] == (streams[0].decode().splitlines()[0] if streams[0] else None), 'first stop'
        for record, previous in zip(report['stops'], old, strict=True):
            ordinal = previous['ordinal']
            assert record['ordinal'] == ordinal, 'stop order'
            assert record['old_message'] == previous['message'], 'baseline diagnostic'
            text = (directory / f'{ordinal:02d}-snapshot-types.stderr').read_text()
            suffix = f"/{previous['file'].removeprefix('src/')}:{previous['line']}:{previous['column']}: "
            diagnostics = [line.split(suffix, 1)[1] for line in text.splitlines() if suffix in line]
            assert record['snapshot_diagnostics'] == diagnostics, 'snapshot diagnostics'
            status = 'remain' if previous['message'] in diagnostics else ('changed' if diagnostics else 'disappear')
            assert record['status'] == status, 'stop classification'
            prefix = f"{report['tree']}/{previous['file']}:{previous['original_line']}:{previous['original_column']}: "
            pristine = [line[len(prefix):] for line in streams[0].decode().splitlines() if line.startswith(prefix)]
            assert record['pristine_diagnostics'] == pristine, 'pristine diagnostics'
            assert record['placeholder_induced'] == (not previous['message_already_in_pristine_diagnostics']), 'placeholder provenance'
            assert record['node_exit'] == int((directory / f'{ordinal:02d}-probe-node.exit').read_text()) == 0, 'Node exit'
            stdout = (directory / f'{ordinal:02d}-probe-node.stdout').read_text()
            assert record['node_stdout'] == stdout == previous['node_stdout'], 'Node observation'
            for action in ('types', 'build'):
                assert record[f'probe_{action}_exit'] == int((directory / f'{ordinal:02d}-probe-{action}.exit').read_text()), 'probe exit'
            probe_stderr = (directory / f'{ordinal:02d}-probe-build.stderr').read_text()
            first = probe_stderr.splitlines()[0] if probe_stderr else None
            assert record['probe_first_stop'] == first, 'probe first stop'
            if record['probe_types_exit']:
                assert previous['probe_message'] in (directory / f'{ordinal:02d}-probe-types.stderr').read_text(), 'probe checker diagnostic'
                assert previous['probe_message'] in probe_stderr, 'native probe checker diagnostic'
    return 'Two compiler pins verified; 30 saved stops classified; all Node witnesses agree'


if __name__ == '__main__':
    try:
        print(check(Path(sys.argv[1]) if len(sys.argv) > 1 else here / 'evidence/compiler-tips'))
    except (AssertionError, KeyError, ValueError, FileNotFoundError) as error:
        print(f'comparison verification failed: {error}')
        sys.exit(1)
