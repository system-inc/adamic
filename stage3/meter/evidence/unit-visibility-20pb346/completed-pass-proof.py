"""Completed cross-file fixture proof with actual before/after census binaries.

Usage: python3 completed-pass-proof.py BINARY_DIRECTORY OUTPUT_DIRECTORY
The binary directory contains before/after-entry, before/after-compiler and census.
"""
from pathlib import Path
import hashlib
import json
import os
import shutil
import subprocess
import sys
import time

repo = Path(__file__).resolve().parents[4]
sys.path.insert(0, str(repo / 'stage3/meter'))
from report import report_pair

binaries, output = (Path(arg).resolve() for arg in sys.argv[1:])
output.mkdir(parents=True, exist_ok=True)
tree = output / 'source'
compiler = tree / 'src/compiler'
entry = tree / 'src/tsc/tsc.ts'
compiler.mkdir(parents=True, exist_ok=True)
entry.parent.mkdir(parents=True, exist_ok=True)
(compiler / 'a.ts').write_text('import {caller} from "./z";\nexport function unused<T>(value:T):T { return value; }\nexport function called<T>(value:T):T { return value; }\nexport const bad: number = "wrong";\n')
(compiler / 'z.ts').write_text('import {called} from "./a";\nexport function caller():number { return called(7); }\n')
entry.write_text('import {caller} from "../compiler/z";\nexport const answer = caller();\n')

results = []
for side in ('before', 'after'):
    run = output / side / 'main'
    (run / 'tsc').mkdir(parents=True, exist_ok=True)
    for mode, source, folder in [('entry', entry, run / 'tsc'), ('compiler', compiler, run)]:
        for latent in (False, True):
            binary = binaries / (side + '-' + mode if latent else 'census')
            path = folder / ('latent.jsonl' if latent else 'census.jsonl')
            env = dict(os.environ)
            env.pop('LATENT_CPU_PROFILE', None)
            if latent:
                env.update(LATENT_FULL='1', LATENT_ASSERT_NO_OUTPUT='1', LATENT_TIMES=str(folder / 'latent-times.jsonl'))
            start = time.monotonic()
            with path.with_suffix('.log').open('w') as log:
                subprocess.run([str(binary), str(source), str(path)], env=env,
                               stdout=log, stderr=subprocess.STDOUT, check=True, timeout=120)
            results.append(dict(side=side, mode=mode, latent=latent,
                                wall_seconds=time.monotonic()-start))
    shutil.copytree(run, run.parent / 'area', dirs_exist_ok=True)

comparisons = []
for relative in ('latent.jsonl', 'tsc/latent.jsonl'):
    before = (output / 'before/main' / relative).read_bytes()
    after = (output / 'after/main' / relative).read_bytes()
    assert before == after, relative
    rows = [json.loads(line) for line in before.splitlines()]
    units = [unit for row in rows[1:] for unit in row['units']]
    called = [u for u in units if u.get('name') == 'called']
    unused = [u for u in units if u.get('name') == 'unused']
    assert len(called) == len(unused) == 1
    assert called[0]['instantiated'] and not unused[0]['instantiated']
    comparisons.append(dict(stream=relative, bytes=len(before),
                            sha256=hashlib.sha256(before).hexdigest(),
                            units=len(units), cross_file_witness=True))

# Independent wall/CPU durations necessarily differ. Hold the unchanged ordinary
# census and compiler timing inputs fixed to compare deterministic report bytes.
# Entry sidecars retain their independent timings and are ignored by report.py.
for label in ('main', 'area'):
    for relative in ('census.jsonl', 'tsc/census.jsonl', 'latent-times.jsonl'):
        before_path = output / 'before' / label / relative
        after_path = output / 'after' / label / relative
        if relative.endswith('census.jsonl'):
            def untimed(path):
                return [{key: value for key, value in json.loads(line).items() if key != 'seconds'}
                        for line in path.read_text().splitlines()]
            assert untimed(before_path) == untimed(after_path), relative
        shutil.copyfile(after_path, after_path.with_name(after_path.stem + '.measured.jsonl'))
        shutil.copyfile(before_path, after_path)
sha = subprocess.run(['git', 'rev-parse', 'origin/main'], cwd=repo,
                     check=True, text=True, stdout=subprocess.PIPE).stdout.strip()
for side in ('before', 'after'):
    run = output / side
    report_pair(tree, tree, run, '20pb346-completed-fixture', sha, sha, output, compiler_commit=sha)
for name in ('report.json', 'report.md'):
    before = (output / 'before' / name).read_bytes()
    after = (output / 'after' / name).read_bytes()
    assert before == after, name
    comparisons.append(dict(stream=name, bytes=len(before),
                            sha256=hashlib.sha256(before).hexdigest()))
(output / 'proof.json').write_text(json.dumps(dict(runs=results, comparisons=comparisons), indent=2)+'\n')
