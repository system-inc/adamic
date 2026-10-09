#!/usr/bin/env python3
"""Validate the final scanner scratch measurement, including failed prerequisites."""
import json
from pathlib import Path
import sys


def validate(summary):
    errors = []
    merges = summary.get('merges', [])
    conflicted = [m for m in merges if m.get('conflicts')]
    failed = [m for m in merges if m.get('exit') != 0]
    status = summary.get('status')
    if conflicted and status != 'merge-conflict':
        errors.append('merge conflict cannot be reported as ' + str(status))
    if status == 'merge-conflict':
        if not conflicted:
            errors.append('merge-conflict needs exact conflicting paths')
        if summary.get('compiler') or summary.get('scanner'):
            errors.append('compiler/scanner must not run after a merge conflict')
    if status == 'pass':
        if failed:
            errors.append('pass requires every merge to exit zero')
        if summary.get('compiler', {}).get('exit') != 0:
            errors.append('pass requires a successful compiler build')
        modes = summary.get('scanner', [])
        if len(modes) != 2 or {m.get('split') for m in modes} != {0, 1}:
            errors.append('pass requires split 0 and 1')
        for mode in modes:
            r = mode.get('report') or {}
            for key, expected in [('build_exit', 0), ('node_exit', 0), ('native_exit', 0),
                                  ('native_diff_exit', 0), ('native_byte_mutant_diff_exit', 1),
                                  ('comparison_control_exit', 0), ('end_mutant_diff_exit', 1)]:
                if r.get(key) != expected:
                    errors.append(f"split {mode.get('split')}: {key} must be {expected}")
            if mode.get('exit') != 0 or mode.get('full_tree_diff_exit') != 0:
                errors.append('pass requires successful runner and full-tree comparison')
        proof = summary.get('scanner_native_evidence') or {}
        for key in ['adamic_sha', 'source_sha', 'run_directory', 'node_sha256', 'native_sha256', 'mutant', 'comparison']:
            if not proof.get(key):
                errors.append('pass requires scanner_native evidence: ' + key)
        if proof.get('node_sha256') != proof.get('native_sha256'):
            errors.append('scanner_native evidence hashes must match')
        if proof.get('mutant', {}).get('comparison_exit') != 1:
            errors.append('scanner_native evidence needs a caught byte mutant')
        if summary.get('error') or summary.get('ordered_stops'):
            errors.append('pass cannot contain an error or discovery stops')
    if status not in {'pass', 'merge-conflict', 'compiler-failed', 'scanner-blocked', 'failed'}:
        errors.append('unknown final status')
    return errors


if __name__ == '__main__':
    data = json.loads(Path(sys.argv[1]).read_text())
    problems = validate(data)
    print(json.dumps({'valid': not problems, 'errors': problems}))
    raise SystemExit(1 if problems else 0)
