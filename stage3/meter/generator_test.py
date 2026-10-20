"""Run generated nil-hook copies, including a permissive copier mutant."""
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[2]


class HookTests(unittest.TestCase):
    def command(self, args, cwd, log):
        with log.open('w') as output:
            result = subprocess.run(args, cwd=cwd, env=dict(os.environ, GOWORK='off'),
                                    stdout=output, stderr=output, timeout=120)
        return result.returncode, log.read_text()

    def test_nil_populated_unknown_and_permissive_mutant(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            folder = root / 'internal/lower'
            folder.mkdir(parents=True)
            source = folder / 'lower.go'
            output = root / 'copy.go'
            generator = Path(os.environ.get('STATECOPY_UNDER_TEST', ROOT / 'stage3/census/latent/statecopy'))
            field = 'impliedTargetQuery func(*ast.Node) *checker.Type'
            for declaration in (field, 'unknown func()', 'impliedTargetQuery func()'):
                source.write_text('package lower\ntype lowering struct { ' + declaration + ' }\n')
                output.unlink(missing_ok=True)
                code, log = self.command(['go', 'run', str(generator), str(root), str(source), str(output)], ROOT, root / 'generate.log')
                if declaration != field:
                    self.assertNotEqual(code, 0, log)
                    self.assertFalse(output.exists(), 'unknown hook published output')
                    continue
                self.assertEqual(code, 0, log)
                generated = output.read_text().replace('github.com/microsoft/TypeScript/tsc/shim/', 'probe/')
                (root / 'go.mod').write_text('module probe\n\ngo 1.23\n')
                for package, typename in [('ast', 'Node'), ('checker', 'Type')]:
                    (root / package).mkdir()
                    (root / package / 'stub.go').write_text('package ' + package + '\ntype ' + typename + ' struct{}\n')
                output.write_text(generated)
                (root / 'shape.go').write_text('package lower\nimport("probe/ast";"probe/checker")\ntype lowering struct { ' + field + ' }\n')
                (root / 'copy_test.go').write_text('''package lower
import("testing";"probe/ast";"probe/checker")
func TestNil(t *testing.T){result:=latentCopy_lowering(lowering{},map[any]any{});if result.impliedTargetQuery!=nil{t.Fatal("nil changed")}}
func TestPopulated(t *testing.T){defer func(){if recover()==nil{t.Fatal("populated hook accepted")}}();latentCopy_lowering(lowering{impliedTargetQuery:func(*ast.Node)*checker.Type{return nil}},map[any]any{})}
''')
                code, log = self.command(['go', 'test', '-count=1', '.'], root, root / 'copy.log')
                self.assertEqual(code, 0, log)
                start = generated.index('\tif value.impliedTargetQuery != nil {')
                end = generated.index('\n\t}', start) + len('\n\t}')
                output.write_text(generated[:start] + generated[end:])
                code, log = self.command(['go', 'test', '-count=1', '.'], root, root / 'mutant.log')
                self.assertNotEqual(code, 0, 'permissive hook mutant survived')
                self.assertIn('populated hook accepted', log)
