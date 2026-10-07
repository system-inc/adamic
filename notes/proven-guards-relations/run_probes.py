import json, pathlib, subprocess, sys
root = pathlib.Path(__file__).resolve().parents[2]
notes = pathlib.Path(__file__).resolve().parent
results = []
for case in json.loads((notes / 'refusals.json').read_text()):
    command = ['go', 'run', './cmd/adamic', 'build', case['file'], '-o', '/tmp/adamic-gate/proven-refusal']
    run = subprocess.run(command, cwd=root, text=True, capture_output=True)
    output = run.stdout + run.stderr
    results.append(dict(case, command=command, exit=run.returncode, output=output, passed=run.returncode != 0 and case['expected'] in output))
(notes / 'refusal-results.json').write_text(json.dumps(results, indent=2) + '\n')
for result in results:
    print(('PASS ' if result['passed'] else 'FAIL ') + result['file'])
    if not result['passed']: print(result['output'])
sys.exit(any(not result['passed'] for result in results))
