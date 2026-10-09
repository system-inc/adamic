import os, subprocess, sys, time, json, resource
from pathlib import Path
root = Path('/tmp/css-printer-proof')
variant = sys.argv[1]
for limit in (2, 8):
    name = f'{variant}-{limit}'
    dump = root / name
    dump.mkdir(exist_ok=True)
    env = dict(os.environ, CSS_PROOF_DUMP=str(dump), ADAMIC_CSS_KEEP=str(dump / 'cases.txt'))
    before = resource.getrusage(resource.RUSAGE_CHILDREN)
    started = time.perf_counter()
    with (root / f'{name}.log').open('w') as log:
        result = subprocess.run([str(root / f'{variant}.test'), '-test.run=^TestCSSPrinterAgreesWithGo$', f'-test.parallel={limit}', '-test.timeout=0', '-test.v'], stdout=log, stderr=subprocess.STDOUT, env=env)
    elapsed = time.perf_counter() - started
    after = resource.getrusage(resource.RUSAGE_CHILDREN)
    report = dict(wall=elapsed, user=after.ru_utime-before.ru_utime, system=after.ru_stime-before.ru_stime, exit=result.returncode, parallel=limit)
    (root / f'{name}.time').write_text(json.dumps(report)+'\n')
    print(name, report, flush=True)
    if result.returncode: sys.exit(result.returncode)
