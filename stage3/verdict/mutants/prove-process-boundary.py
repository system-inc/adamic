#!/usr/bin/env python3
"""Use a stock reference and a no-output candidate to check path reuse."""
import json
import os
from pathlib import Path
import subprocess
import sys
import time


SELF = Path(__file__).resolve()
BUILD_INFO = 'tsconfig.tsbuildinfo'


def late_writer(project, payload):
    marker = project.parent / 'candidate' / 'ready'
    deadline = time.monotonic() + 5
    while time.monotonic() < deadline:
        if marker.exists():
            (project / BUILD_INFO).write_bytes(payload.read_bytes())
            return
        time.sleep(0.01)


def reference(detached, argv):
    subprocess.run(['node', os.environ['STAGE3_VERDICT_NODE_TSC'], *argv], check=True)
    project = Path.cwd()
    payload = project.parent / 'reference' / 'payload'
    payload.write_bytes((project / BUILD_INFO).read_bytes())
    subprocess.Popen([sys.executable, str(SELF), 'late-writer', str(project), str(payload)],
                     start_new_session=detached)


def candidate():
    # The marker is a capture-side handshake, outside the compiler output tree.
    project = Path.cwd()
    (project.parent / 'candidate' / 'ready').touch()
    deadline = time.monotonic() + 0.5
    while time.monotonic() < deadline and not (project / BUILD_INFO).exists():
        time.sleep(0.01)


def proof(output):
    sys.path.insert(0, str(SELF.parents[1]))
    from run import compare_case

    success = True
    for mode in ('stock', 'late-writer', 'detached-late-writer'):
        command = ['node', os.environ['STAGE3_VERDICT_NODE_TSC']] if mode == 'stock' else [
            sys.executable, str(SELF), 'reference', mode,
        ]
        folder = output / mode
        result = compare_case(SELF, command, folder, {'emit.ts': 'export const n = 1;\n'},
                              {'noEmit': True, 'incremental': True})
        output_present = (folder / 'candidate/project' / BUILD_INFO).is_file()
        exits = [(folder / side / 'actual.exit').read_text().strip() for side in ('reference', 'candidate')]
        print(json.dumps({'case': mode, 'compiler_exits': exits, 'candidate_build_info_present': output_present,
                          'differences': result}), flush=True)
        success &= (exits == ['0', '0'] and not output_present
                    and result.get('candidate_build_info', {}).get('path') == BUILD_INFO
                    and result.get('files', {}).get(BUILD_INFO, {}).get('kind') == 'missing')
    return 0 if success else 1


if __name__ == '__main__':
    mode, *args = sys.argv[1:]
    if mode == 'late-writer':
        late_writer(*map(Path, args))
    elif mode == 'reference':
        reference(args[0] == 'detached-late-writer', args[1:])
    elif mode == '--project':
        candidate()
    else:
        raise SystemExit(proof(Path(mode).resolve()))
