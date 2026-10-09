#!/usr/bin/env python3
"""Run the scanner proof, retaining failures and proving comparison sensitivity."""
import argparse
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys

here = Path(__file__).resolve().parent
repository = here.parents[2]
parser = argparse.ArgumentParser()
parser.add_argument('output', type=Path, help='new scratch output directory')
parser.add_argument('--tree', type=Path, help='reuse a tree already made by stage3/apply.sh')
parser.add_argument('--compiler', type=Path, help='existing Adamic binary, otherwise go run')
parser.add_argument('--compiler-cwd', type=Path, help='compiler checkout containing pinned Node declarations')
parser.add_argument('--node-only', action='store_true')
parser.add_argument('--oracle', type=Path, help='full-tree Node stdout to compare byte for byte')
parser.add_argument('--inputs', type=Path, help='fixed compiler corpus for comparing source adaptations')
args = parser.parse_args()
out = args.output.resolve()
if out.exists():
    sys.exit(f'refusing to replace existing output: {out}')
out.mkdir(parents=True)

def run(command, stem, cwd=repository, env=None):
    with (out / (stem + '.stdout')).open('wb') as stdout, (out / (stem + '.stderr')).open('wb') as stderr:
        return subprocess.run(command, cwd=cwd, env=env, stdout=stdout, stderr=stderr).returncode

if args.tree:
    tree = args.tree.resolve()
    (out / 'adapted').symlink_to(tree, target_is_directory=True)
else:
    tree = out / 'adapted'
    if not (repository / 'stage3/apply.sh').exists():
        sys.exit('stage3/apply.sh is missing: integrate codex/stage3-base and the adaptations first')
    # October 7: adaptation 20 is excluded until its upstream-build fix lands.
    # Copy the pipeline so apply still owns construction and patch-set accounting.
    pipeline = out / 'pipeline'
    pipeline.mkdir()
    for name in ['apply.sh', 'apply.py', 'source.json']:
        shutil.copyfile(repository / 'stage3' / name, pipeline / name)
    shutil.copytree(repository / 'stage3/api', pipeline / 'api')
    (pipeline / 'adapt').mkdir()
    for name in ['00-setup', '10-type-imports', '42-scanner-any', '50-temporary-scanner-implicit-returns', '51-temporary-scanner-fallthrough']:
        source = repository / 'stage3/adapt' / name
        if source.exists():
            shutil.copytree(source, pipeline / 'adapt' / name)
    if run(['bash', str(pipeline / 'apply.sh'), str(tree)], 'apply'):
        sys.exit('adapted tree failed; see apply.stderr')
# Adaptation 10 edits the generator owner. Regenerate its output afterward,
# as upstream's build does, so the newly annotated type import is present.
if not (tree / 'slice.json').exists() and run(['node', 'scripts/processDiagnosticMessages.mjs', 'src/compiler/diagnosticMessages.json'], 'generate', tree):
    sys.exit('diagnostic generation failed; see generate.stderr')
shutil.copyfile(here / 'main.a', out / 'main.a')
# Every regular file recursively, including JSON inputs and generated diagnostics.
inputs = args.inputs.resolve() if args.inputs else tree
files = sorted(p.relative_to(inputs).as_posix() for p in (inputs / 'src/compiler').rglob('*') if p.is_file())
(out / 'files.json').write_text(json.dumps(files, indent=2) + '\n')
cache = Path(os.environ.get('STAGE3_CACHE', str(Path.home() / '.cache/adamic-stage3')))
api = cache / 'api/node_modules/typescript/lib/typescript.js'
if not api.exists():
    sys.exit(f'stock TypeScript API not found: {api}; run stage3/apply.sh with adaptations')
env = dict(os.environ, SCANNER_TYPESCRIPT=str(api.resolve()), SCANNER_RUNTIME=str(repository / 'oracle/adamic.mjs'))
if run(['node', str(here / 'inventory.cjs'), str(tree), str(out)], 'inventory', repository, env):
    sys.exit('source inventory failed; see inventory.stderr')
paths = [str(inputs / p) for p in files]
node = run(['node', '--disable-warning=ExperimentalWarning', str(here / 'node.mjs'), str(out / 'main.a'), *paths], 'node', out, env)
if node:
    sys.exit(f'Node oracle failed ({node}); see node.stderr')
if args.oracle:
    oracle_diff = run(['diff', '-u', str(args.oracle.resolve()), str(out / 'node.stdout')], 'full-tree-diff')
    if oracle_diff:
        sys.exit(f'slice Node differs from full-tree oracle (diff exit {oracle_diff})')
# The native binary is unavailable while lowering is blocked. Prove the exact
# comparison with a scratch control copied from Node, then mutate Node's end.
control = out / 'comparison-control.stdout'
shutil.copyfile(out / 'node.stdout', control)
assert run(['diff', '-u', str(control), str(out / 'node.stdout')], 'control-diff') == 0
mutant = out / 'node-end-mutant.stdout'
with (out / 'node.stdout').open('rb') as source, mutant.open('wb') as target:
    while True:
        line = source.readline()
        if not line:
            sys.exit('oracle did not produce a token')
        first = line.split(b'\t')
        if len(first) == 7 and first[0] != b'error':
            break
        target.write(line)
    first[3] = str(int(first[3]) + 1).encode()
    target.write(b'\t'.join(first))
    shutil.copyfileobj(source, target)
mutation = run(['diff', '-u', str(control), str(mutant)], 'mutant-diff')
if mutation != 1:
    sys.exit(f'end mutant was not caught by comparison (diff exit {mutation})')
dump = (out / 'node.stdout').read_bytes()
import hashlib
rows = dump.splitlines()
tokens = sum(len(row.split(b'\t')) == 7 and not row.startswith(b'error\t') for row in rows)
errors = sum(row.startswith(b'error\t') for row in rows)
report = {'tokens': tokens, 'errors': errors, 'sha256': hashlib.sha256(dump).hexdigest(), 'bytes': len(dump), 'input_tree': str(inputs), 'files': len(files), 'node_exit': node, 'comparison_control_exit': 0, 'end_mutant_diff_exit': mutation,
          'native': 'not attempted' if args.node_only else 'pending'}
if not args.node_only:
    command = [str(args.compiler.resolve())] if args.compiler else ['go', 'run', './cmd/adamic']
    compiled = run([*command, 'build', str(out / 'main.a'), '-o', str(out / 'scanner-native')], 'build',
                   args.compiler_cwd.resolve() if args.compiler_cwd else repository)
    report['build_exit'] = compiled
    if compiled:
        report['native'] = 'blocked at compilation'
    else:
        native = run([str(out / 'scanner-native'), *paths], 'native', out)
        difference = run(['diff', '-u', str(out / 'node.stdout'), str(out / 'native.stdout')], 'native-diff')
        report.update(native='executed', native_exit=native, native_diff_exit=difference)
        if native or difference:
            report['failure'] = 'native exit or token bytes differ'
        else:
            native_mutant = out / 'native-byte-mutant.stdout'
            with (out / 'native.stdout').open('rb') as source, native_mutant.open('wb') as target:
                first = source.read(1)
                if not first:
                    sys.exit('native output empty: cannot plant byte mutant')
                target.write(bytes([first[0] ^ 1]))
                shutil.copyfileobj(source, target)
            killed = run(['diff', '-u', str(out / 'node.stdout'), str(native_mutant)], 'native-byte-mutant-diff')
            report['native_byte_mutant_diff_exit'] = killed
            if killed != 1:
                report['failure'] = 'native byte mutant was not caught by comparison'
(out / 'report.json').write_text(json.dumps(report, indent=2) + '\n')
print(json.dumps(report))
if report.get('build_exit', 0) or report.get('failure'):
    sys.exit(1)
