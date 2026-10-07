#!/usr/bin/env python3
"""Check the new .a driver against the unchanged nine-rule landing artifacts."""
import argparse
import hashlib
import json
from pathlib import Path
import subprocess

parser = argparse.ArgumentParser()
parser.add_argument('--scratch', type=Path, required=True)
parser.add_argument('--prior-six', type=Path, required=True)
parser.add_argument('--prior-three', type=Path, required=True)
args = parser.parse_args()
owned = Path(__file__).resolve().parent
repository = owned.parents[4]
args.scratch.mkdir(exist_ok=True, parents=True)
subjects = [
    (args.prior_six, owned.parent / 'typescript-no-unnecessary-type-constraint' / 'verify.a', [
        '@typescript-eslint/no-unnecessary-type-constraint', '@typescript-eslint/prefer-as-const',
        '@typescript-eslint/prefer-enum-initializers', 'nexus/import-require-node-namespace',
        'structure/network-no-invalidate-cache-literal-key', 'structure/network-no-string-literal-query']),
    (args.prior_three, owned.parent / 'no-multi-str' / 'verify.a', [
        'no-multi-str', 'no-nonoctal-decimal-escape', 'no-octal']),
]
observations = {}
for index, (prior, entry, names) in enumerate(subjects):
    manifest = args.scratch / ('group-' + str(index) + '.manifest')
    manifest.write_text(''.join(str(owned / 'standalone.a') + '\t' + name + '\n' for name in names))
    expected = None
    for side, command in [
        ('Go', [prior / 'oracle', manifest]),
        ('Node', ['node', '--disable-warning=ExperimentalWarning', repository / 'oracle/node.mjs', entry, manifest]),
        ('JavaScript', ['node', '--disable-warning=ExperimentalWarning', repository / 'oracle/node.mjs', str(prior / 'port') + '.mjs', manifest]),
        ('native', [prior / 'port', manifest]),
    ]:
        result = subprocess.run(list(map(str, command)), cwd=repository, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        path = args.scratch / ('group-' + str(index) + '-' + side + '.log')
        path.write_bytes(result.stdout)
        path.with_suffix('.stderr').write_bytes(result.stderr)
        if result.returncode or result.stderr:
            raise RuntimeError(str(path) + ': failed execution')
        if expected is None:
            expected = result.stdout
        if result.stdout != expected:
            raise RuntimeError(str(path) + ': byte comparison differs')
        observations[path.name] = dict(bytes=len(result.stdout), sha256=hashlib.sha256(result.stdout).hexdigest())
    print(str(len(names)) + ' unchanged rules agree on the new stage1 .a source on all four sides', flush=True)
(args.scratch / 'hashes.json').write_text(json.dumps(observations, indent=2) + '\n')
