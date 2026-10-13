import json, os, pathlib, subprocess, time
root = pathlib.Path(__file__).resolve().parent
results = []
for patch in sorted(root.glob('*.patch')):
    name = patch.stem
    package = 'load' if name.startswith('load-') else 'fuzz'
    command = ['timeout', '89', 'go', 'test', './internal/' + package, '-json', '-count=1', '-timeout', '85s']
    if package == 'fuzz':
        command += ['-run', '^(TestOctoberFeaturesAppear|TestUndefinedNumbersShapes)$']
    subprocess.run(['git', 'apply', '--check', str(patch)], check=True)
    subprocess.run(['git', 'apply', str(patch)], check=True)
    start = time.monotonic()
    try:
        env = dict(os.environ, ADAMIC_GATE_UNCACHED='1', ADAMIC_BUILD_CACHE_DIR='/tmp/audit-del-a-cache-' + name)
        with (root / (name + '.log')).open('w') as log:
            run = subprocess.run(command, stdout=log, stderr=subprocess.STDOUT, env=env)
        rows = []
        for line in (root / (name + '.log')).read_text().splitlines():
            try:
                event = json.loads(line)
            except ValueError:
                continue
            if event.get('Action') == 'fail' and event.get('Test'):
                rows.append(event['Test'])
        result = dict(mutant=name, command=command, exit=run.returncode, wall_seconds=round(time.monotonic()-start, 3), caught_by=rows)
        results.append(result)
        (root / 'replay-results.json').write_text(json.dumps(results, indent=2) + '\n')
        print(json.dumps(result), flush=True)
        assert run.returncode == 1 and rows, 'Mutant was not caught by a remaining test'
    finally:
        subprocess.run(['git', 'apply', '-R', str(patch)], check=True)
