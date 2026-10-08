"""Exercise compiler selection with real Git refs and small compiled Go drivers."""
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest

from report import report_pair


DRIVER = r'''package main
import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "meterprobe/internal/compiler"
)
func main() {
    root, output := os.Args[1], os.Args[2]
    file := filepath.Join(root, "input.a")
    stream, err := os.Create(output)
    if err != nil { panic(err) }
    defer stream.Close()
    encoder := json.NewEncoder(stream)
    fmt.Fprintf(os.Stderr, "compiled with %s\n", compiler.Name)
    if filepath.Base(root) == "tsc.ts" {
        entry := root
        dependency := filepath.Clean(filepath.Join(filepath.Dir(entry), "../executeCommandLine.a"))
        diagnostics := []string{}
        if compiler.Name == "area" { diagnostics = append(diagnostics, dependency+":1:1: error TS2322: entry compiler witness") }
        if LATENT {
            if len(diagnostics) > 0 {
                encoder.Encode(map[string]any{"roots":[]string{entry}, "status":"blocked", "checker_rejected":true, "diagnostics":diagnostics})
            } else {
                label := "measured on a checker-clean entry-root program"
                encoder.Encode(map[string]any{"roots":[]string{entry}, "status":"measurement", "checker_rejected":false, "diagnostics":diagnostics, "measurement":label, "sources":[]string{entry,dependency}})
                encoder.Encode(map[string]any{"file":entry, "measurement":label, "findings":[]any{}})
                encoder.Encode(map[string]any{"file":dependency, "measurement":label, "findings":[]any{
                    map[string]any{"measurement":label, "kind":"NotYet", "where":dependency+":1:1", "reason":"compiled with "+compiler.Name, "text":"entry compiler provenance"},
                }})
            }
        } else {
            kind := "accepted"
            if len(diagnostics) > 0 { kind = "checker" }
            for index := 0; index < 2; index++ { encoder.Encode(map[string]any{"roots":[]string{entry}, "kind":kind, "diagnostics":diagnostics}) }
        }
        return
    }
    if LATENT {
        encoder.Encode(map[string]any{"measurement":"measured on a checker-rejected program", "status":"measurement", "checker_rejected":true})
        encoder.Encode(map[string]any{"measurement":"measured on a checker-rejected program", "file":file, "findings":[]any{
            map[string]any{"measurement":"measured on a checker-rejected program", "kind":"NotYet", "where":file+":1:1", "reason":"compiled with "+compiler.Name, "text":"compiler provenance"},
        }})
    } else {
        kind := "accepted"
        diagnostics := []string{}
        if compiler.Name == "area" {
            kind = "checker"
            diagnostics = append(diagnostics, file+":1:1: error TS2322: area compiler witness")
        }
        // Two roots distinguish per-file observations from the aggregate record.
        other := filepath.Join(root, "other.a")
        encoder.Encode(map[string]any{"roots":[]string{file}, "kind":kind, "diagnostics":diagnostics})
        encoder.Encode(map[string]any{"roots":[]string{other}, "kind":kind, "diagnostics":diagnostics})
        encoder.Encode(map[string]any{"roots":[]string{file,other}, "kind":kind, "diagnostics":diagnostics})
    }
}
'''
# Latent coverage must include both source files too.
DRIVER = DRIVER.replace('    } else {\n        kind :=', '''        other := filepath.Join(root, "other.a")
        encoder.Encode(map[string]any{"measurement":"measured on a checker-rejected program", "file":other, "findings":[]any{}})
    } else {
        kind :=''')


@unittest.skipUnless(shutil.which('go'), 'Go is required for real compiler selection probes')
class CompilerSelectionTests(unittest.TestCase):
    def setUp(self):
        self.scratch = tempfile.TemporaryDirectory()
        self.addCleanup(self.scratch.cleanup)
        self.root = Path(self.scratch.name)
        self.repository = self.root / 'repository'
        self.repository.mkdir()
        self.env = dict(os.environ, GIT_CONFIG_NOSYSTEM='1', GIT_CONFIG_GLOBAL=os.devnull,
                        GIT_AUTHOR_NAME='Meter test', GIT_AUTHOR_EMAIL='meter@example.invalid',
                        GIT_COMMITTER_NAME='Meter test', GIT_COMMITTER_EMAIL='meter@example.invalid',
                        GIT_TERMINAL_PROMPT='0', TMPDIR=str(self.root), GOWORK='off')
        self.env.pop('STAGE3_METER_COMPILER', None)
        self.env.pop('STAGE3_METER_RUNS', None)
        self.command('git', 'init', '-b', 'main')
        source = Path(__file__).resolve().parent
        for name in ('report.py', 'owners.json'):
            self.write('stage3/meter/' + name, (source / name).read_text())
        script = Path(os.environ.get('METER_SCRIPT_UNDER_TEST', source / 'twice-daily.sh'))
        self.write('stage3/meter/twice-daily.sh', script.read_text())
        self.write('go.mod', 'module meterprobe\n\ngo 1.21\n')
        self.write('stage3/apply.sh', 'set -eu\nmkdir -p "$1/src/compiler" "$1/src/tsc"\nprintf "const x: number = 1;\\n" > "$1/src/compiler/input.a"\nprintf "const y: number = 2;\\n" > "$1/src/compiler/other.a"\nprintf "export {};\\n" > "$1/src/tsc/tsc.ts"\nprintf "export {};\\n" > "$1/src/executeCommandLine.a"\n')
        self.write('stage3/census/tool/main.go', DRIVER.replace('LATENT', 'false'))
        self.write('stage3/census/latent/tool/main.go', DRIVER.replace('LATENT', 'true'))
        self.write('stage3/census/latent/make_overlay.py',
                   'import json, pathlib, sys\np=pathlib.Path(sys.argv[2]); p.mkdir(parents=True)\n'
                   '(p/"empty.go").write_text("package main\\n")\n'
                   '(p/"overlay.json").write_text(json.dumps({"Replace":{}, "compiler_root":sys.argv[1]}))\n')
        # A small entry driver provides distinct compiler witnesses in these refs.
        self.write('stage3/meter/entry_overlay.py',
                   'import json, pathlib, sys\ns=pathlib.Path(sys.argv[1]); p=pathlib.Path(sys.argv[2]); p.mkdir(parents=True)\n'
                   'metadata=json.loads((s/"overlay.json").read_text())\n'
                   '(p/"entry.go").write_text(' + repr(DRIVER.replace('LATENT', 'true')) + ')\n'
                   '(p/"overlay.json").write_text(json.dumps({"Replace":{str(pathlib.Path(metadata["compiler_root"])/"stage3/census/latent/tool/main.go"):str(p/"entry.go")}}))\n')
        self.write('internal/compiler/name.go', 'package compiler\nconst Name = "main"\n')
        self.command('git', 'add', '.')
        self.command('git', 'commit', '-m', 'Main compiler witness')
        self.main = self.command('git', 'rev-parse', 'HEAD').strip()
        self.command('git', 'checkout', '-b', 'area/stage3')
        self.write('internal/compiler/name.go', 'package compiler\nconst Name = "area"\n')
        self.command('git', 'commit', '-am', 'Area compiler witness')
        self.area = self.command('git', 'rev-parse', 'HEAD').strip()
        remote = self.root / 'remote.git'
        self.command('git', 'clone', '--bare', str(self.repository), str(remote))
        self.command('git', 'remote', 'add', 'origin', str(remote))
        self.command('git', 'checkout', '-b', 'worker')
        self.write('internal/compiler/name.go', 'package compiler\nconst Name = "checkout"\n')
        self.command('git', 'commit', '-am', 'Checkout compiler witness')
        self.checkout = self.command('git', 'rev-parse', 'HEAD').strip()

    def write(self, name, text):
        path = self.repository / name
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(text)

    def command(self, *args):
        # All subprocess output is written to a file before reading it.
        with (self.root / 'commands.log').open('w+') as log:
            result = subprocess.run(args, cwd=self.repository, env=self.env,
                                    stdout=log, stderr=log, timeout=120)
            log.seek(0)
            output = log.read()
        self.assertEqual(result.returncode, 0, output)
        return output

    def measure(self, mode=None):
        if mode:
            self.env['STAGE3_METER_COMPILER'] = mode
        self.command('bash', 'stage3/meter/twice-daily.sh')
        run, = (self.repository / 'stage3/meter/runs').iterdir()
        report = json.loads((run / 'report.json').read_text())
        for label in ('main', 'area'):
            tree = report['trees'][label]
            self.assertEqual(tree['tree_commit'], self.main if label == 'main' else self.area)
        return run, report

    def test_default_uses_checkout_compiler_for_both_trees(self):
        run, result = self.measure()
        self.assertEqual(result['compiler_mode'], 'single')
        for label in ('main', 'area'):
            tree = result['trees'][label]
            self.assertEqual(tree['adamic_commit'], self.checkout)
            self.assertEqual(tree['totals']['checker_own_file'], 2)
            self.assertEqual(tree['latent_lowering']['per_reason'], {'NotYet: compiled with checkout': 1})
            self.assertTrue(tree['tsc_entry']['checker_whole_program'])
            self.assertEqual(tree['tsc_entry']['lowering_census']['per_reason'], {'NotYet: compiled with checkout': 1})
        self.assertTrue((run / 'build.log').exists())
        self.assertIn('Both measured with Adamic ' + self.checkout, (run / 'report.md').read_text())

    def test_per_ref_builds_and_runs_each_pinned_compiler(self):
        run, result = self.measure('per-ref')
        self.assertEqual(result['compiler_mode'], 'per-ref')
        self.assertEqual(result['adamic_commit'], self.area)
        for label, sha, count in [('main', self.main, 2), ('area', self.area, 1)]:
            tree = result['trees'][label]
            self.assertEqual(tree['adamic_commit'], sha)
            self.assertEqual(tree['totals']['checker_own_file'], count)
            self.assertEqual(tree['latent_lowering']['per_reason'], {'NotYet: compiled with ' + label: 1})
            self.assertIn('compiled with ' + label, (run / label / 'census.log').read_text())
            entry = tree['tsc_entry']
            self.assertEqual(entry['checker_whole_program'], label == 'main')
            self.assertEqual(entry['whole_program_diagnostics'], 0 if label == 'main' else 1)
            if label == 'main':
                self.assertEqual(entry['lowering_census']['source_files'], 2)
            else:
                self.assertIsNone(entry['lowering_census'])
            self.assertTrue((run / label / 'tsc/census.log').exists())
            self.assertIn('compiled with ' + label, (run / label / 'latent.log').read_text())
            self.assertTrue((run / label / 'build.log').exists())
        self.assertFalse((run / 'build.log').exists())
        text = (run / 'report.md').read_text()
        self.assertEqual(text.splitlines()[:2], ['Whole program: main: 2/2; area: 0/2', 'Own file: main: 2/2; area: 1/2'])
        self.assertIn('tsc entry: main: pass; area: fail', text)
        self.assertIn('tsc entry diagnostics: main: 0; area: 1', text)
        self.assertIn('Main measured with Adamic ' + self.main, text)
        self.assertIn('area measured with Adamic ' + self.area, text)
        # A claimed per-ref compiler must agree with the actual source pin.
        metadata = json.loads((run / 'compiler-mode.json').read_text())
        metadata['commits']['main'] = self.checkout
        (run / 'compiler-mode.json').write_text(json.dumps(metadata))
        with self.assertRaisesRegex(ValueError, 'do not match source pins'):
            report_pair(Path(result['trees']['main']['adapted_tree']),
                        Path(result['trees']['area']['adapted_tree']), run, 'stamp', self.main, self.area)

    def test_invalid_mode_fails_before_creating_a_run(self):
        self.env['STAGE3_METER_COMPILER'] = 'typo'
        with (self.root / 'invalid.log').open('w') as log:
            result = subprocess.run(['bash', 'stage3/meter/twice-daily.sh'], cwd=self.repository,
                                    env=self.env, stdout=log, stderr=log, timeout=10)
        self.assertEqual(result.returncode, 2)
        self.assertIn('invalid STAGE3_METER_COMPILER', (self.root / 'invalid.log').read_text())
        self.assertFalse((self.repository / 'stage3/meter/runs').exists())


if __name__ == '__main__':
    unittest.main()
