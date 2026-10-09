"""Reapply this unit's two layers and compare every nondependency tree byte."""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys

unit = Path(__file__).resolve().parent
root = Path(sys.argv[1]).resolve()
assert (root / 'src/compiler/parser.ts').is_file()


def hashes():
    result = {}
    for directory, dirs, files in os.walk(root):
        dirs[:] = sorted(name for name in dirs if name not in {'.git', 'node_modules'})
        for name in sorted(files):
            file = Path(directory) / name
            if file.is_file():
                result[str(file.relative_to(root))] = hashlib.sha256(file.read_bytes()).hexdigest()
    return result


before = hashes()
reports = []
for script in [unit / 'adapt.cjs', unit.parent / '46-fix-pragma-empty-argument/adapt.cjs']:
    completed = subprocess.run(['node', str(script), str(root)], text=True, capture_output=True, check=True)
    report = json.loads(completed.stdout)
    assert report['files'] == report['edits'] == 0, report
    reports.append(report)
assert hashes() == before
print(json.dumps({'status': 'pass', 'files_hashed': len(before), 'reports': reports,
                  'identical_bytes': True, 'excluded': ['.git', 'node_modules']}, indent=2))
