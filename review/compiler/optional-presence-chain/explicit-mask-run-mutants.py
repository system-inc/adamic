import json
from pathlib import Path
import subprocess

root = Path.cwd()
out = root / 'review/compiler/optional-presence-chain/explicit-mask-source-mutants'
out.mkdir(exist_ok=True)
mutants = [
 ('drop-bool-reservation', 'internal/lower/optional_fields.go', '(slotless(of) && of != ir.MaybeBoolean)', 'slotless(of)', './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/optional_field_checked_view.a$', 'field write failed: flag'),
 ('drop-string-retain', 'internal/native/emit_statements.go', 'keptValue = e.kept(value)', 'keptValue = value', './internal/oracle', 'TestOptionalFieldCheckedViewWrites', 'Sanitizer'),
 ('cast-origin', 'internal/lower/optional_fields.go', 'return l.enumerationDataObject(node.AsAsExpression().Expression, depth+1)', 'return true', './internal/lower', 'TestOptionalDeletionRequiresPlainStorage', 'plain storage boundary'),
 ('plain-origin', 'internal/lower/optional_fields.go', 'if remove && !l.enumerationDataObject(receiver, 0) {', 'if false {', './internal/lower', 'TestOptionalDeletionRequiresPlainStorage', 'plain storage boundary'),
 ('hasown-origin', 'internal/lower/library_object.go', 'return l.exactObject(node, 0)', 'return true', './internal/lower', 'TestObjectUnprovenShapesStayNotYet', 'want NotYet'),
 ('binder-registration', 'internal/oracle/taste_stage3_test.go', '"stage3/fixtures/taste/17_binder_flow.a", true, false', '"stage3/fixtures/taste/17_binder_flow.a", false, false', './internal/oracle', 'TestNativeAgreesWithNode/stage3/fixtures/taste/17_binder_flow', 'want stage 0 to refuse'),
]
for name, file, before, after, package, test, catcher in mutants:
 source = (root / file).read_text()
 assert source.count(before) == 1, name
 replacement = out / (name + '.go.txt')
 replacement.write_text(source.replace(before, after, 1))
 overlay = out / (name + '.json')
 overlay.write_text(json.dumps({'Replace': {str(root / file): str(replacement)}}))
 with (out / (name + '.log')).open('w') as log:
  result = subprocess.run(['go', 'test', '-overlay', str(overlay), package, '-run', test, '-count=1', '-v', '-timeout', '90s'], stdout=log, stderr=subprocess.STDOUT, timeout=120)
 text = (out / (name + '.log')).read_text()
 assert result.returncode == 1 and catcher in text and 'build failed' not in text, (name, text)
 print(name + ': caught by ' + catcher)

# Apply the catalog patch to scratch source strings. Keep mutant sources non-compilable by census.
patch = (root / 'verify/catalog/09-literal-undefined-field.patch').read_text()
replacements = {}
for section in patch.split('--- a/')[1:]:
 file, remainder = section.split('\n', 1)
 source = (root / file).read_text()
 for hunk in remainder.split('@@')[2::2]:
  lines = hunk.splitlines()[1:]
  before = ''.join(line[1:] + '\n' for line in lines if line.startswith((' ', '-')))
  after = ''.join(line[1:] + '\n' for line in lines if line.startswith((' ', '+')))
  assert source.count(before) == 1, (file, before)
  source = source.replace(before, after, 1)
 replacement = out / ('catalog-' + Path(file).name + '.txt')
 replacement.write_text(source)
 replacements[str(root / file)] = str(replacement)
overlay = out / 'catalog-overlay.json'
overlay.write_text(json.dumps({'Replace': replacements}))
with (out / 'catalog.log').open('w') as log:
 result = subprocess.run(['go', 'test', '-overlay', str(overlay), './internal/oracle', '-run', 'TestNativeAgreesWithNode/internal/oracle/testdata/e4eec87_u01_undefined_field_widened', '-count=1', '-v', '-timeout', '90s'], stdout=log, stderr=subprocess.STDOUT, timeout=120)
text = (out / 'catalog.log').read_text()
assert result.returncode == 1 and 'stdout differs' in text and 'build failed' not in text, text
print('literal-undefined catalog: caught by stdout differs')
