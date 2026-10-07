#!/usr/bin/env python3
"""Reconcile scalar destruction visits with uninstrumented branch edges and assembly."""
import argparse
import json
import re
from pathlib import Path
from read_profile import read_profile


def reconcile(directory, mutate_count=False):
    profile = read_profile(directory / 'baseline.callgrind')
    row = (directory / 'counter.stderr').read_text()
    values = {key: int(value) for key, value in re.findall(r'(\w+)=(\d+)', row)}
    if mutate_count:
        values['scalar_class'] += 1
    parts = []
    for kind, address, per_visit in [('class', '0x6e7a8', 7), ('plain', '0x6e7e1', 6)]:
        records = [r for r in profile['instructions'] if r['function'] == 'adamic_object_free_children' and r['address'] == address]
        taken = sum(r['taken'] for r in records)
        visits = sum(r['costs'][profile['events'].index('Bc')] for r in records)
        assert taken == values['scalar_' + kind], f'{kind}: scalar counter/profile mismatch'
        assert visits == values['scalar_' + kind] + values['reference_' + kind], f'{kind}: total counter/profile mismatch'
        parts.append(dict(kind=kind, visits=taken, instructions_per_visit=per_visit, instructions=taken*per_visit,
                          branch_address=address, all_visits=visits,
                          type_branch_mispredictions=sum(r['costs'][profile['events'].index('Bcm')] for r in records)))
    instructions = sum(r['instructions'] for r in parts)
    totals = dict(zip(profile['events'], profile['totals']))
    return dict(totals=totals, header_delta=profile['header_delta'], counters=values, scalar_paths=parts,
                scalar_instructions=instructions, scalar_parse_fraction=instructions/totals['Ir'],
                stop_under_two_percent=instructions/totals['Ir'] < .02)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('directory', type=Path)
    parser.add_argument('--mutate-count', action='store_true')
    args = parser.parse_args()
    result = reconcile(args.directory, args.mutate_count)
    if not args.mutate_count:
        (args.directory/'result.json').write_text(json.dumps(result, indent=2)+'\n')
    print(json.dumps(result, indent=2))


if __name__ == '__main__':
    main()
