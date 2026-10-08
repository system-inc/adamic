"""Missing-value producers stay Unknown and name their local or return frontier."""
import json
import pathlib
import subprocess
import tempfile

root = pathlib.Path.cwd()
source_path = root / 'internal/lower/shape_flow.go'
source = source_path.read_text()
mutants = {
 'local-cause-erased': ('unknown(fmt.Sprintf("local %d has a producer without a value", node-1))', 'unknown("missing expression")', 'local'),
 'result-cause-erased': ('unknown(fmt.Sprintf("function %d returns without a value", node-len(graph.program.Locals)-1))', 'unknown("missing expression")', 'result'),
 'nil-local-is-ignored': ('unknown(fmt.Sprintf("local %d has a producer without a value", node-1))', '_ = node', 'local'),
 'nil-result-is-ignored': ('unknown(fmt.Sprintf("function %d returns without a value", node-len(graph.program.Locals)-1))', '_ = node', 'result'),
}
with tempfile.TemporaryDirectory(prefix='shape-missing-producers-') as directory:
 scratch = pathlib.Path(directory)
 for name, (old, new, case) in mutants.items():
  assert source.count(old) == 1, name
  replacement = scratch / (name + '.go')
  replacement.write_text(source.replace(old, new))
  overlay = scratch / (name + '.json')
  overlay.write_text(json.dumps({'Replace': {str(source_path): str(replacement)}}))
  binary = scratch / (name + '.test')
  log_path = root / 'stage3/shape-conformance/overnight' / (name + '.log')
  with log_path.open('w') as log:
   build = subprocess.run(['go', 'test', '-c', '-overlay=' + str(overlay), '-o', str(binary), './internal/lower'], stdout=log, stderr=subprocess.STDOUT)
   assert build.returncode == 0, name + ' did not compile'
   result = subprocess.run([str(binary), '-test.run=^TestShapeMissingValueProducers$/^' + case + '$', '-test.count=1', '-test.v'], cwd=root / 'internal/lower', stdout=log, stderr=subprocess.STDOUT)
   assert result.returncode != 0, name + ' escaped'
  assert 'nonempty allocations must retain their missing-value frontier' in log_path.read_text(), name + ' failed for an unrelated reason'
  print(name + ': valid test binary; missing-value frontier assertion caught mutant', flush=True)
