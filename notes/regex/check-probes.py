"""Record CLI observations for all coverage programs without changing testdata."""
import json
import pathlib
import subprocess

root = pathlib.Path(__file__).resolve().parents[2]
output = pathlib.Path('/tmp/regex-coverage')
output.mkdir(exist_ok=True)
paths = sorted((root / 'internal/oracle/testdata').glob('regex_coverage_*.a'))
paths += sorted((root / 'internal/oracle/testdata/regex_coverage_refused').glob('*.a'))
paths += sorted((root / 'notes/regex').glob('*.a'))
rows = []
for path in paths:
    program = str(path.relative_to(root))
    row = {'program': program}
    commands = {
        'node': ['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', program],
        'build': ['go', 'run', './cmd/adamic', 'build', program, '-o', str(output / path.stem)],
        'jsbuild': ['go', 'run', './cmd/adamic', 'js', program],
    }
    observations = {}
    for name, command in commands.items():
        result = subprocess.run(command, cwd=root, capture_output=True, timeout=180)
        row[name] = result.returncode
        observations[name] = result
        (output / (path.stem + '.' + name + '.stdout')).write_bytes(result.stdout)
        (output / (path.stem + '.' + name + '.stderr')).write_bytes(result.stderr)
    if row['build'] == 0:
        result = subprocess.run([str(output / path.stem)], capture_output=True, timeout=15)
        row['native'] = result.returncode
        row['agrees'] = (result.returncode, result.stdout, result.stderr) == (observations['node'].returncode, observations['node'].stdout, observations['node'].stderr)
        (output / (path.stem + '.native.stdout')).write_bytes(result.stdout)
        (output / (path.stem + '.native.stderr')).write_bytes(result.stderr)
    if row['jsbuild'] == 0:
        javascript = output / (path.stem + '.mjs')
        javascript.write_bytes(observations['jsbuild'].stdout)
        result = subprocess.run(['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', str(javascript)], cwd=root, capture_output=True, timeout=15)
        row['js'] = result.returncode
        row['jsAgrees'] = (result.returncode, result.stdout, result.stderr) == (observations['node'].returncode, observations['node'].stdout, observations['node'].stderr)
        (output / (path.stem + '.js.stdout')).write_bytes(result.stdout)
        (output / (path.stem + '.js.stderr')).write_bytes(result.stderr)
    rows.append(row)
    print(json.dumps(row), flush=True)
(output / 'direct.json').write_text(json.dumps(rows, indent=2) + '\n')
