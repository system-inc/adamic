from pathlib import Path
import subprocess

output = Path('/tmp/string-views-builds')
output.mkdir(exist_ok=True)
for program in sorted(Path('internal/oracle/testdata').glob('string_views_*.a')):
    binary = output / program.stem
    command = ['go', 'run', './cmd/adamic', 'build', str(program), '-o', str(binary)]
    print(' '.join(command), flush=True)
    subprocess.run(command, check=True)
    node = subprocess.run(['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', str(program.resolve())], capture_output=True)
    native = subprocess.run([str(binary)], capture_output=True)
    for label, result in [('node', node), ('native', native)]:
        binary.with_suffix('.' + label + '.stdout').write_bytes(result.stdout)
        binary.with_suffix('.' + label + '.stderr').write_bytes(result.stderr)
    assert (node.returncode, node.stdout, node.stderr) == (native.returncode, native.stdout, native.stderr), program
    assert node.returncode == 0, program
    print(f'PASS: identical stdout ({len(node.stdout)} bytes), empty stderr, exit 0', flush=True)
