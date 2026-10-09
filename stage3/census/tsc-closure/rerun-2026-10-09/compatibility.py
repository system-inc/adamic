"""Minimal scratch compatibility for f1502d13 tools on the measured main."""
import argparse
from pathlib import Path
import shutil

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('tools', type=Path)
parser.add_argument('fixture_runner', type=Path)
args = parser.parse_args()
full = args.tools / 'overlay/internal_lower_latent_full.go'
text = full.read_text()
needle = '\t\t\t\tif !result.Field(i).CanSet() {\n'
assert text.count(needle) == 1
# This private field is a derived backend cache. Leave it nil in the new
# snapshot; the Program accessor recomputes it from copied public state.
# All other unexported fields still fail the pinned copier's guard.
text = text.replace(needle, needle + '''\t\t\t\t\tif v.Type() == reflect.TypeOf(ir.Program{}) && v.Type().Field(i).Name == "argumentFacts" {
\t\t\t\t\t\tcontinue
\t\t\t\t\t}
''')
full.write_text(text)
args.fixture_runner.mkdir()
original = Path(__file__).resolve().parent.parent
for name in ['test_fixture.py', 'measure.py', 'closure.cjs']:
    shutil.copyfile(original / name, args.fixture_runner / name)
for name in ['src/compiler/entry.a', 'src/outside.a']:
    source = original / 'fixtures' / name
    text = source.read_text()
    header = '// a-check: type error TS2322\n'
    assert text.startswith(header)
    target = args.fixture_runner / 'fixtures' / name
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_text(text.removeprefix(header))
print('Scratch only: invalidate Program.argumentFacts on snapshot; restore fixture bodies before header metadata')
