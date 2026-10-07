#!/usr/bin/env python3
"""Restore the production Go tests' input projects for the isolated byte gate."""
import argparse
import gzip
import json
import pathlib

parser = argparse.ArgumentParser()
parser.add_argument('destination', type=pathlib.Path)
args = parser.parse_args()
destination = args.destination.resolve()
fixture = pathlib.Path(__file__).parent / 'testdata/upstream.json.gz'
projects = json.loads(gzip.decompress(fixture.read_bytes()))
for project in projects:
    case = destination / project['name']
    for name, contents in project['files'].items():
        path = case / name
        if not path.resolve().is_relative_to(case.resolve()):
            raise ValueError('fixture path escapes project')
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(contents)
    (case / 'manifest').write_text(''.join(str(case / name) + '\n' for name in project['roots']))
print('Restored', len(projects), 'upstream projects')
