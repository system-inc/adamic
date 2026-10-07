#!/usr/bin/env python3
"""Run every requested parity check independently, retaining full diagnostics."""
from concurrent.futures import ThreadPoolExecutor, as_completed
import json
import os
from pathlib import Path
import re
import subprocess
import time

root = Path(__file__).resolve().parent.parent
report = root / 'cloud/reports/css-gate-inputs'
scratch = Path('/workspace/css-parity-proof');scratch.mkdir(exist_ok=True)
tests = [('css', name) for name in ['TestCSSParserOptimizedMatchesNode', 'TestCSSPrinterAgreesWithGo',
         'TestCSSPrinterOptimizedMatchesGo', 'TestCompositionMatchesGo', 'TestThePortParsesAsGoCohereDoes']]
tests += [('cssnumbers', 'TestCSSNumbers'), ('cssstrings', 'TestCSSStrings'), ('markdowninline', 'TestMarkdownInline')]
results = []

def run(item):
    package, name = item
    environment = dict(os.environ, ADAMIC_GATE_UNCACHED='1',
                       ADAMIC_CSS_KEEP=str(scratch / (name + '.cases')),
                       ADAMIC_CSS_KEEP_RAW=str(scratch / (name + '.answers')))
    for kind in ['CSSNUMBERS', 'CSSSTRINGS', 'MARKDOWNINLINE']:
        environment['ADAMIC_' + kind + '_KEEP'] = str(scratch / (name + '.' + kind + '.cases'))
    flags = dict(commit=subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=root).decode().strip(),
                 nproc=subprocess.check_output(['nproc']).decode().strip(),
                 cpu_max=Path('/sys/fs/cgroup/cpu.max').read_text().strip(),
                 go=subprocess.check_output(['go', 'version']).decode().strip(),
                 clang=subprocess.check_output(['clang', '--version']).decode().splitlines()[0],
                 node=subprocess.check_output(['node', '--version']).decode().strip(),
                 GOFLAGS=os.environ.get('GOFLAGS', ''), cached=False,
                 load_before=Path('/proc/loadavg').read_text().strip())
    args = ['go', 'test', './stage1/cohere/' + package, '-run', '^' + name + '$', '-count=1', '-timeout=30m', '-v']
    start = time.monotonic()
    with (report / (name + '.log')).open('w') as log:
        process = subprocess.run(args, cwd=root, env=environment, stdout=log, stderr=subprocess.STDOUT, timeout=1860)
    elapsed = time.monotonic() - start
    flags['load_after'] = Path('/proc/loadavg').read_text().strip()
    text = (report / (name + '.log')).read_text()
    corpus = [line.strip() for line in text.splitlines() if 'corpus files:' in line or 'corpus:' in line]
    recorded = [line.strip() for line in text.splitlines() if 'recorded discrepancy occurrences' in line or 'proved surrogate gap' in line]
    degraded = [line for line in text.splitlines() if 'external library not checked' in line or 'Prettier fixtures absent' in line]
    root_pass = bool(re.search(r'^--- PASS: ' + name + r' ', text, re.M))
    verdict = 'PASS' if process.returncode == 0 and root_pass else 'FAIL'
    coverage = package != 'css' or any('.css:158 .less:43 .scss:90' in line for line in corpus)
    if degraded or not coverage:verdict = 'FAIL (coverage)'
    result = dict(test=name, package=package, verdict=verdict, exit=process.returncode, seconds=elapsed,
                  corpus_lines=corpus, degraded=degraded, recorded_library_differences=recorded, expected_coverage=coverage,
                  instrument=args, build_flags=flags)
    print(name, verdict, 'exit', process.returncode, f'{elapsed:.3f}s', corpus, flush=True)
    return result

with ThreadPoolExecutor(max_workers=2) as executor:
    for future in as_completed([executor.submit(run, item) for item in tests]):
        results.append(future.result())
        (report / 'verdicts.json').write_text(json.dumps(results, indent=2) + '\n')
assert len(results) == 8
