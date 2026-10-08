#!/usr/bin/env python3
"""Who hears a fast gate's verdict, and what they read, with a stand-in router and a stand-in ahra."""
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

script = Path(__file__).with_name('verdict-notify.py')
SHA = 'ab' * 20
ROUTER = '''
def route(branch):
    if branch.startswith('codex/held-'):
        return 'hold', 'area-routes.tsv holds codex/held-*: ask @system_adamic_compiler'
    return ('compiler' if branch.startswith('codex/compiler-') else 'nowhere'), 'prefix'
def areaNames():
    return {'compiler', 'stage3'}
def fleetArea(fleet):
    return {'compiler': 'compiler', 'stage3': 'stage3'}.get(fleet)
'''


class VerdictTests(unittest.TestCase):
    def send(self, branch, log, toIntegration=False):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root / 'routes').mkdir()
            (root / 'routes/areas.tsv').write_text('# area\towner\toracle\ncompiler\tsystem_adamic_compiler\tnone\nstage3\tsystem_adamic_typescript\tnone\n')
            (root / 'routes/area-route.py').write_text(ROUTER)
            (root / 'bin').mkdir()
            (root / 'bin/ahra').write_text('#!/bin/bash\nprintf "%s|%s\\n" "$3" "$4" >> "$SENDS"\n')
            (root / 'bin/ahra').chmod(0o755)
            (root / 'state').mkdir()
            (root / 'roster.json').write_text('[{"label": "scout-map-keys", "fleet": "compiler"}, {"label": "scout-x", "fleet": "stage1-scouts"}]')
            if toIntegration:
                (root / 'state/verdicts-to-integration').touch()
            (root / 'gate.log').write_text(log)
            env = dict(os.environ, PATH=str(root / 'bin') + ':' + os.environ['PATH'], SENDS=str(root / 'sends'),
                       ADAMIC_VERDICT_ROUTES_DIR=str(root / 'routes'), ADAMIC_FAST_GATE_WATCH_STATE=str(root / 'state'),
                       ADAMIC_FAST_GATE_AHRA_DIR=tmp, ADAMIC_FLEET_ROSTER=str(root / 'roster.json'))
            subprocess.run(['python3', str(script), branch, SHA, str(root / 'gate.log')], env=env, check=True, capture_output=True)
            sends = root / 'sends'
            return [line.split('|', 1) for line in sends.read_text().splitlines()] if sends.exists() else []

    red = ('fast gate: %s against area/compiler 2f024dda2787de4168550bdd6bc419536dc6a924, tools x, on server, class S\n'
           'FIRST FAILURE (tests, at 153.4 s):\ngithub.com/system-inc/adamic/internal/fuzz TestOwnershipShapes\n'
           'red: %s fast gate, first failure at tests after 153.4 s (session 01a11b49-cdae-7122-96bf-971d73d43002), 2 fail, 1586 pass\n'
           'published gate-logs/ababababab/20261008T160758Z/fast (670f)\n') % (SHA, SHA)

    def test_worker_red_reaches_its_area_owner_only(self):
        sends = self.send('codex/compiler-loops', self.red)
        self.assertEqual([name for name, _ in sends], ['system_adamic_compiler'])
        text = sends[0][1]
        self.assertIn('Fast gate red: codex/compiler-loops abababababab against area/compiler 2f024dd', text)
        self.assertIn('first failure at tests: internal/fuzz TestOwnershipShapes (2 failed, 1586 passed)', text)
        self.assertIn('Log: gate-logs/ababababab/20261008T160758Z/fast.', text)
        self.assertIn('Worker session 01a11b49-cdae-7122-96bf-971d73d43002.', text)

    def test_area_reaches_owner_and_integration(self):
        log = 'green: %s fast gate in 630.1 s (stuff)\n' % SHA
        self.assertEqual([name for name, _ in self.send('area/stage3', log)], ['system_adamic_typescript', 'system_adamic_integration'])

    def test_held_or_unrouted_worker_goes_to_integration_with_why(self):
        sends = self.send('codex/held-views', self.red)
        self.assertEqual([name for name, _ in sends], ['system_adamic_integration', 'system_adamic_compiler'])
        self.assertIn('Routed to integration: area-routes.tsv holds codex/held-*', sends[0][1])
        self.assertEqual([name for name, _ in self.send('codex/other', self.red)], ['system_adamic_integration'])

    def test_unrouted_branch_routes_by_its_session_label(self):
        self.assertEqual([name for name, _ in self.send('codex/scout-map-keys', self.red)], ['system_adamic_compiler'])
        self.assertEqual([name for name, _ in self.send('codex/scout-x', self.red)], ['system_adamic_integration'])
        self.assertEqual([name for name, _ in self.send('codex/held-views', self.red)][0], 'system_adamic_integration')

    def test_switch_copies_integration_and_void_says_nothing(self):
        self.assertEqual([name for name, _ in self.send('codex/compiler-x', self.red, toIntegration=True)],
                         ['system_adamic_compiler', 'system_adamic_integration'])
        self.assertEqual(self.send('codex/compiler-x', 'void: box lacks wasmtime\n'), [])

    def test_the_tools_branch_is_left_to_its_canaries(self):
        self.assertEqual(self.send('devtools/fast-gate', self.red), [])
        self.assertEqual([name for name, _ in self.send('devtools/gate-shards', self.red)], ['system_adamic_developer_tools'])

    def test_no_all_caps_word_survives(self):
        log = self.red.replace('TestOwnershipShapes', 'TestJSONShapes')
        self.assertIn('Testjsonshapes', self.send('codex/compiler-x', log)[0][1])


if __name__ == '__main__':
    unittest.main()
