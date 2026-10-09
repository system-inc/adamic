"""Check the extracted rows against the pinned, unmodified historical ranking."""
import argparse
import json
from pathlib import Path
import subprocess

PIN = '6c4fc1afb019d0a16fea97082e45fef8fcbe1746'
SOURCE = 'stage3/census/hidden-ranking/RESULT.json'


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--mutant', choices=('bytes', 'roots', 'witnesses', 'boundaries'))
    args = parser.parse_args()
    repo = Path(__file__).resolve().parents[2]
    original = json.loads(subprocess.check_output(['git', 'show', f'{PIN}:{SOURCE}'], cwd=repo))
    extracted = json.loads(Path(__file__).with_name('hidden.json').read_text())
    row = extracted['rows'][0]
    if args.mutant == 'bytes':
        row['bytes_revealed_if_fixed_alone'] += 1
    elif args.mutant == 'roots':
        row['roots'] += 1
    elif args.mutant == 'witnesses':
        row['diagnostic_witnesses'][0] = 'invented.a:1:1'
    elif args.mutant == 'boundaries':
        row['boundaries'].pop()
    ranked = {(r['kind'], r['reason']): r for r in original['ranked_reasons']}
    for row in extracted['rows']:
        key = row['kind'], row['reason']
        expected = [b for b in original['boundaries'] if (b['kind'], b['reason']) == key]
        assert row['boundaries'] == expected, 'boundary extraction changed'
        assert row['bytes_revealed_if_fixed_alone'] == ranked[key]['bytes_revealed_if_fixed_alone'], 'historical byte credit changed'
        assert sum(b['attributed_hidden_bytes'] for b in expected) == row['bytes_revealed_if_fixed_alone'], 'attribution sum changed'
        sites = {b['where'] for b in expected}
        assert row['roots'] == len(sites), 'diagnostic root deduplication changed'
        assert row['units'] == len({b['unit'] for b in expected}), 'attempted unit deduplication changed'
        assert len(row['diagnostic_witnesses']) == min(3, len(sites)), 'witness count changed'
        assert len(set(row['diagnostic_witnesses'])) == len(row['diagnostic_witnesses']) and set(row['diagnostic_witnesses']) <= sites, 'witness provenance changed'
    print(f"{len(extracted['rows'])} historical reasons agree with pinned ranking")


if __name__ == '__main__':
    main()
