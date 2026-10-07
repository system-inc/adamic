import json, pathlib, subprocess, sys
root = pathlib.Path(__file__).resolve().parents[2]
notes = pathlib.Path(__file__).resolve().parent
results = []
for path in sorted((root / 'internal/oracle/testdata').glob('coverage_*.a')):
    binary = '/tmp/adamic-gate/' + path.stem
    relative = str(path.relative_to(root))
    command = ['go', 'run', './cmd/adamic', 'build', relative, '-o', binary]
    build = subprocess.run(command, cwd=root, text=True, capture_output=True)
    node_command = ['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', relative]
    node = subprocess.run(node_command, cwd=root, text=True, capture_output=True)
    native = subprocess.run([binary], cwd=root, text=True, capture_output=True) if build.returncode == 0 else build
    passed = build.returncode == 0 and (node.returncode, node.stdout, node.stderr) == (native.returncode, native.stdout, native.stderr)
    results.append(dict(file=relative, command=command, node_command=node_command, native_command=[binary], build_exit=build.returncode, build_output=build.stdout+build.stderr, node=dict(exit=node.returncode, stdout=node.stdout, stderr=node.stderr), native=dict(exit=native.returncode, stdout=native.stdout, stderr=native.stderr), passed=passed))
    print(('PASS ' if passed else 'FAIL ') + relative)
(notes / 'build-results.json').write_text(json.dumps(results, indent=2) + '\n')
sys.exit(any(not result['passed'] for result in results))
