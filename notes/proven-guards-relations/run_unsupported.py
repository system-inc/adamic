import json, pathlib, subprocess, sys
root = pathlib.Path(__file__).resolve().parents[2]
notes = pathlib.Path(__file__).resolve().parent
results = []
for name, expected in [('tuple_to_array', 'a tuple is held as an object'), ('exact_optional_tuple', 'optional field element.count'), ('nullable_primitives', 'number | null'), ('unknown_guard', 'a value of type unknown'), ('generic_guard', 'true return narrows to T, not T & Leaf'), ('array_from_iterable', 'Array.from with other than { length } and a callback'), ('array_constructor', 'new an Identifier'), ('primitive_filter_guard', 'an array of string | number')]:
    file = str((notes / (name + '.a')).relative_to(root))
    command = ['go', 'run', './cmd/adamic', 'build', file, '-o', '/tmp/adamic-gate/' + name]
    build = subprocess.run(command, cwd=root, text=True, capture_output=True)
    node_command = ['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', file]
    node = subprocess.run(node_command, cwd=root, text=True, capture_output=True)
    diagnostic = build.stdout + build.stderr
    passed = build.returncode != 0 and expected in diagnostic and node.returncode == 0
    results.append(dict(file=file, command=command, node_command=node_command, build_exit=build.returncode, diagnostic=diagnostic, node=dict(exit=node.returncode, stdout=node.stdout, stderr=node.stderr), expected=expected, passed=passed))
    print(('PASS ' if passed else 'FAIL ') + file)
(notes / 'unsupported-results.json').write_text(json.dumps(results, indent=2) + '\n')
sys.exit(any(not result['passed'] for result in results))
