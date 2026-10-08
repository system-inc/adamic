"""Run a writer accepting equal hashes alone; the missing-mutant check must kill it."""
import ast
import os
from pathlib import Path
import subprocess
import sys

source = Path(__file__).resolve().parent
destination = Path(sys.argv[1]).resolve()
destination.mkdir(parents=True, exist_ok=True)
tree = ast.parse((source / 'progress.py').read_text())
functions = [node for node in tree.body if isinstance(node, ast.FunctionDef) and node.name == 'validate_evidence']
assert len(functions) == 1, 'expected validate_evidence function'
functions[0].body = ast.parse("return record.get('node_sha256') == record.get('native_sha256'), 'equal hashes alone'").body
mutant = destination / 'equal-hashes-only.py'
mutant.write_text(ast.unparse(ast.fix_missing_locations(tree)) + '\n')
env = dict(os.environ, METER_PROGRESS_WRITER_UNDER_TEST=str(mutant))
log = destination / 'equal-hashes-only.log'
with log.open('w') as output:
    result = subprocess.run([sys.executable, '-m', 'unittest',
        'progress_test.ProgressTests.test_missing_mutant'], cwd=source, env=env,
        stdout=output, stderr=subprocess.STDOUT)
text = log.read_text()
assert result.returncode == 1 and 'FAIL: test_missing_mutant' in text and 'True is not false' in text, text
print('equal-hashes-only writer mutant caught by missing native-output mutant assertion; exit=1')
