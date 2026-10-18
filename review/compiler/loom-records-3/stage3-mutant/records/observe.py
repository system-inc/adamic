"""Record Node and current stage-0 observations; run from the repository root."""
import json
import re
import subprocess
from pathlib import Path

ROOT = Path(__file__).resolve().parent
REPO = ROOT.parents[2]
LOGS = Path('/tmp/records-observations')
LOGS.mkdir(exist_ok=True)
rows = []
for file in sorted(ROOT.glob('*.a')):
    relative = file.relative_to(REPO)
    node = subprocess.run(['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', str(relative)], cwd=REPO, capture_output=True, text=True)
    (LOGS / (file.stem + '.node.log')).write_text(node.stdout + node.stderr)
    build = subprocess.run(['go', 'run', './cmd/adamic', 'build', str(relative), '-o', str(LOGS / file.stem)], cwd=REPO, capture_output=True, text=True)
    diagnostic = build.stdout + build.stderr
    (LOGS / (file.stem + '.stage0.log')).write_text(diagnostic)
    if build.returncode == 0:
        outcome = 'Compiles'
        native = subprocess.run([str(LOGS / file.stem)], capture_output=True, text=True)
        (LOGS / (file.stem + '.native.log')).write_text(native.stdout + native.stderr)
        if (native.stdout, native.stderr, native.returncode) != (node.stdout, node.stderr, node.returncode):
            raise RuntimeError('SILENT MISCOMPILE: ' + file.name)
    elif 'not yet' in diagnostic.lower() or 'notyet' in diagnostic.lower():
        outcome = 'NotYet'
    elif 'refuse' in diagnostic.lower():
        outcome = 'Refused'
    elif re.search(r'error TS\d+', diagnostic):
        outcome = 'Checker'
    else:
        raise RuntimeError('Unclassified build failure: ' + diagnostic)
    rows.append({'file': file.name, 'tsc': re.findall(r'// From TypeScript 6.0.3, (.*)', file.read_text()), 'reason': 'an index signature', 'node': {'stdout': node.stdout, 'stderr': node.stderr, 'exit': node.returncode}, 'stage0': {'outcome': outcome, 'what': diagnostic if build.returncode else ''}})
    print(file.name, 'node', node.returncode, 'stage0', outcome, flush=True)
(ROOT / 'status.json').write_text(json.dumps(rows, indent=2) + '\n')
