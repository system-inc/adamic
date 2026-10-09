#!/usr/bin/env python3
"""Generate the exact diff from pinned upstream and adaptation proof data."""
import argparse
import difflib
import json
from pathlib import Path
import subprocess

lane = Path(__file__).resolve().parent
parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('typescript_git', type=Path, help='pinned checkout or the apply cache bare repository')
parser.add_argument('--output', type=Path, default=lane / 'sanctioned-api.json')
args = parser.parse_args()
proof = json.loads(subprocess.check_output(['node', str(lane / 'generate-api.cjs'),
                                          str(args.typescript_git.resolve())], text=True))
original = proof.pop('original')
expected = proof.pop('expected')
proof['baseline_diffs'] = ['api/typescript.d.ts']
proof['diff_lines'] = list(difflib.unified_diff(original.splitlines(True), expected.splitlines(True),
                                              fromfile='reference/api/typescript.d.ts',
                                              tofile='local/api/typescript.d.ts'))
args.output.write_text(json.dumps(proof, indent=2) + '\n')
print(f"Generated {args.output}: {proof['counts']}")
