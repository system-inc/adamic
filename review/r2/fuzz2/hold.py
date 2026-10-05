#!/usr/bin/env python3
# hold.py <tree> <program.a>...: run each program's source on Node through oracle/probe.mjs, and its
# JavaScript backend output on Node, and say whether the two stopped at the same check with the same
# stdout before it. (A checked program's backend run is byte for byte its native one: the fuzzer's own
# verdict.)
import os, subprocess, sys, tempfile
tree = sys.argv[1]
node = ['node', '--disable-warning=ExperimentalWarning']
for program in sys.argv[2:]:
    probe = subprocess.run(node + [os.path.join(tree, 'oracle', 'probe.mjs'), program], capture_output=True, timeout=60)
    js = subprocess.run([os.path.join(tree, 'adamic-cli'), 'js', program], capture_output=True).stdout
    with tempfile.TemporaryDirectory(dir=os.environ.get('TMPDIR')) as work:
        path = os.path.join(work, 'program.mjs')
        open(path, 'wb').write(js)
        backend = subprocess.run(node + [os.path.join(tree, 'oracle', 'node.mjs'), path], capture_output=True, timeout=60)
    same = probe.returncode == backend.returncode and probe.stdout == backend.stdout and probe.stderr.split(b'\n')[0] == backend.stderr.split(b'\n')[0]
    first = lambda run: run.stderr.split(b'\n')[0].decode(errors='replace')[:90]
    verdict = 'same' if same else 'DIFFERENT'
    print(f'{os.path.basename(program)}\t{verdict}\tprobe exit {probe.returncode} {first(probe)}\tbackend exit {backend.returncode} {first(backend)}')
