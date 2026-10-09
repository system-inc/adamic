#!/usr/bin/env python3
"""Corrupt scratch upstream bytes and prove their pins stop execution."""
import argparse
from pathlib import Path
import tempfile
from run import ROOT, baseline_suite
import json


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('scratch_upstream', type=Path, help='disposable checkout, never the active measurement tree')
    args = parser.parse_args()
    tree = args.scratch_upstream.resolve()
    manifest = json.loads((ROOT / 'selection.json').read_text())
    binary = ROOT / 'standins/empty.sh'
    for field, message in [('source', 'source hash mismatch'), ('baseline', 'baseline hash mismatch')]:
        index, row = next((index, row) for index, row in enumerate(manifest['cases']) if row[field])
        path = tree / row[field]
        original = path.read_bytes()
        try:
            path.write_bytes(original + b'\n')
            with tempfile.TemporaryDirectory(prefix='verdict-pin-mutant-') as scratch:
                try:
                    baseline_suite(binary, tree, Path(scratch) / 'baselines', index + 1)
                except RuntimeError as error:
                    if message not in str(error):
                        raise
                    print(field + ' mutant caught: ' + str(error), flush=True)
                else:
                    raise RuntimeError(field + ' mutant survived')
        finally:
            path.write_bytes(original)
    print('both artifact pin mutants caught and restored', flush=True)


if __name__ == '__main__':
    main()
