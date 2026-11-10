"""Keep the complete committed API sanction equal to canonical generation."""
import itertools
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

LANE = Path(__file__).resolve().parent


def differing_keys(committed, generated, key='$'):
    if type(committed) is not type(generated):
        yield key
    elif isinstance(committed, dict):
        for name in sorted(committed.keys() | generated.keys()):
            child = f'{key}.{name}'
            if name not in committed or name not in generated:
                yield child
            else:
                yield from differing_keys(committed[name], generated[name], child)
    elif isinstance(committed, list):
        for index, (before, after) in enumerate(zip(committed, generated)):
            child = f'{key}[{index}]'
            if isinstance(before, dict) and 'file' in before:
                child += f'({before["file"]})'
            yield from differing_keys(before, after, child)
        if len(committed) != len(generated):
            yield f'{key}.length'
    elif committed != generated:
        yield key


class SanctionedApiFreshnessTests(unittest.TestCase):
    def test_lane_committed_sanction_matches_full_tree_generation(self):
        upstream = os.environ.get('ADAMIC_TYPESCRIPT_SOURCE')
        self.assertTrue(upstream, 'ADAMIC_TYPESCRIPT_SOURCE must name the pinned TypeScript Git input')
        environment = dict(os.environ, NODE_PATH=str(LANE.parent / 'api/node_modules'),
                           TSC_ADAPT_TYPESCRIPT=str(LANE.parent / 'api/node_modules/typescript'))
        with tempfile.TemporaryDirectory(prefix='lane-api-freshness-') as directory:
            output = Path(directory) / 'sanctioned-api.json'
            result = subprocess.run(
                [sys.executable, '-B', str(LANE / 'generate-api.py'), upstream, '--output', str(output)],
                env=environment, capture_output=True, text=True)
            self.assertEqual(result.returncode, 0,
                             f'canonical generator exit {result.returncode}\n{result.stdout}{result.stderr}')
            committed = (LANE / 'sanctioned-api.json').read_bytes()
            generated = output.read_bytes()
        if committed != generated:
            keys = list(itertools.islice(differing_keys(json.loads(committed), json.loads(generated)), 8))
            self.fail('sanctioned-api.json differs from full tree generation; first differing keys: '
                      + (', '.join(keys) if keys else '(JSON encoding or whitespace)')
                      + '\nRegenerate from the repository root:\n'
                      + 'NODE_PATH="$PWD/stage3/api/node_modules" python3 stage3/lane/generate-api.py '
                        '"$ADAMIC_TYPESCRIPT_SOURCE"')


if __name__ == '__main__':
    unittest.main()
