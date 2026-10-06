from pathlib import Path
import subprocess

repository = Path(__file__).resolve().parents[2]
output = Path('/tmp/enums-coverage-builds')
output.mkdir(exist_ok=True)
programs = sorted((repository / 'internal/oracle/testdata').glob('enums_coverage*.a'))
programs += sorted((repository / 'internal/oracle/testdata/enums_coverage_modules').glob('*.a'))
for program in programs:
    name = program.stem if program.parent.name != 'enums_coverage_modules' else 'enums_coverage_modules_' + program.stem
    relative = str(program.relative_to(repository))
    node = subprocess.run(['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', str(program)], cwd=repository, capture_output=True)
    build = subprocess.run(['go', 'run', './cmd/adamic', 'build', relative, '-o', str(output / name)], cwd=repository, capture_output=True)
    (output / (name + '.build')).write_bytes(build.stdout + build.stderr)
    assert build.returncode == 0, build.stderr.decode()
    native = subprocess.run([str(output / name)], capture_output=True)
    (output / (name + '.node')).write_bytes(node.stdout)
    (output / (name + '.native')).write_bytes(native.stdout)
    assert (node.returncode, node.stdout, node.stderr) == (native.returncode, native.stdout, native.stderr), name
    print(f'{relative}: Node and CLI native agree, exit {native.returncode}', flush=True)
    print(native.stdout.decode(), end='', flush=True)
