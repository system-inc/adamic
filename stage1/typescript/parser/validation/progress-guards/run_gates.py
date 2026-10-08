"""Record complete gates independently of the executor's interactive connection."""
import json
import os
import pathlib
import subprocess
import time

root = pathlib.Path(__file__).resolve().parent
repo = root.parents[4]
environment = dict(os.environ, ADAMIC_TYPESCRIPT_SOURCE='/workspace/scratch/typescript-6.0.3', ADAMIC_PARSER_BENCH='1', ADAMIC_LINT_BENCH='1', ADAMIC_RECOVERY_ARTIFACTS='/tmp/adamic-parser-progress-guards')

def box():
    return {'hostname': os.uname().nodename, 'nproc': subprocess.check_output(['nproc'], text=True).strip(), 'quota': pathlib.Path('/sys/fs/cgroup/cpu.max').read_text().strip(), 'load': os.getloadavg(), 'memory_events': pathlib.Path('/sys/fs/cgroup/memory.events').read_text()}

results = {}
for label, package, selector in [('parser', './stage1/typescript/parser', None), ('rules', './stage1/cohere/lint', '^TestRulesAgree$'), ('lintcases', './stage1/typescript/parser', '^TestLintCases$')]:
    command = ['go', 'test', package, '-count=1', '-json', '-v', '-failfast=false', '-timeout=90m']
    if selector:
        command += ['-run', selector]
    before = box()
    started = time.monotonic()
    with (root / f'{label}.jsonl').open('w') as log:
        result = subprocess.run(command, cwd=repo, env=environment, stdout=log, stderr=subprocess.STDOUT)
    results[label] = {'command': command, 'exit': result.returncode, 'wall_seconds': time.monotonic()-started, 'before': before, 'after': box()}
    (root / 'gate-results.json').write_text(json.dumps(results, indent=2)+'\n')
    print(label, results[label], flush=True)
(root / 'finished.json').write_text(json.dumps({'successful': all(r['exit'] == 0 for r in results.values())})+'\n')
