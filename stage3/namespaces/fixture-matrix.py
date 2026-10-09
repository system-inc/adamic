"""Build the same 12 real-source fixtures at every namespace commit and the merge."""
from pathlib import Path
import argparse
import io
import json
import os
import re
import subprocess
import tarfile

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--scratch', type=Path, required=True)
parser.add_argument('--output', type=Path, required=True)
parser.add_argument('--mutant', action='store_true', help='Build a Parser countNode mutation and prove the Node equality check rejects it')
args = parser.parse_args()
repo = Path(__file__).resolve().parents[2]

def verify(expected, actual):
    if expected != actual:
        raise RuntimeError(f'Node disagreement: expected {expected!r}, observed {actual!r}')

if args.mutant:
    saved = json.loads(args.output.read_text())
    fixture = repo / 'stage3/fixtures/namespaces/07_parser_singleton.a'
    source = fixture.read_text()
    assert source.count('nodeCount++;') == 1
    mutated = args.scratch / 'parser-count-mutant.a'
    mutated.write_text(source.replace('nodeCount++;', 'nodeCount += 2;'))
    binary = args.scratch / 'parser-count-mutant'
    compiler = args.scratch / f"adamic-{len(saved['stages']) - 1}"
    build = subprocess.run([str(compiler), 'build', str(mutated), '-o', str(binary), '--sanitize'], cwd=repo, capture_output=True)
    assert build.returncode == 0, build.stderr.decode()
    native = subprocess.run([str(binary)], capture_output=True, env=dict(os.environ, ASAN_OPTIONS='detect_leaks=1', UBSAN_OPTIONS='halt_on_error=1'))
    observed = {'stdout': native.stdout.decode(), 'stderr': native.stderr.decode(), 'exit': native.returncode}
    (args.scratch / 'parser-count-mutant.json').write_text(json.dumps(observed, indent=2) + '\n')
    assert observed['exit'] == 0 and observed['stderr'] == '', observed
    verify(saved['node'][fixture.name], observed)
    raise RuntimeError('Parser countNode mutant survived')

scratch = args.scratch.resolve()
scratch.mkdir(parents=True, exist_ok=True)
fixtures = sorted((repo / 'stage3/fixtures/namespaces').glob('[0-9][0-9]_*.a'))
assert len(fixtures) == 12

def run(command, cwd, name, env=None):
    result = subprocess.run(command, cwd=cwd, capture_output=True, env=env, timeout=300)
    observed = {'stdout': result.stdout.decode(), 'stderr': result.stderr.decode(), 'exit': result.returncode}
    (scratch / (name + '.json')).write_text(json.dumps(observed, indent=2) + '\n')
    return observed

node = {}
recorded = {row['file']: row['node'] for row in json.loads((repo / 'stage3/fixtures/namespaces/status.json').read_text())}
for fixture in fixtures:
    observed = run(['node', '--disable-warning=ExperimentalWarning', str(repo / 'oracle/node.mjs'), str(fixture)], repo, fixture.stem + '.node')
    verify(recorded[fixture.name], observed)
    assert observed['exit'] == 0 and observed['stderr'] == ''
    node[fixture.name] = observed

points = [('base', 'adc45ca'), ('1 state', 'aff39ff'), ('2 enums', 'b0e9b15'),
          ('3 bodies', 'f5161d8'), ('4 exports', '083e7a9'), ('5 escape refusal', '2f40563'),
          ('6 merge refusals', 'ce8a2ac')]
results = {'fixture_commit': 'a75ba0023a70e01829f1b17850240f60446c1fb9', 'node': node, 'stages': []}
for index, (label, revision) in enumerate(points + [('fixture branch merge', 'working-tree')]):
    stage_dir = scratch / f'stage-{index}'
    if revision == 'working-tree':
        checkout = repo
        sha = subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=repo, text=True).strip()
    else:
        sha = subprocess.check_output(['git', 'rev-parse', revision], cwd=repo, text=True).strip()
        checkout = stage_dir / 'source'
        checkout.mkdir(parents=True, exist_ok=True)
        archive = subprocess.check_output(['git', 'archive', sha, 'cmd', 'internal', 'go.mod'], cwd=repo)
        with tarfile.open(fileobj=io.BytesIO(archive)) as tree:
            tree.extractall(checkout, filter='data')
        if not (checkout / 'cohere').exists():
            (checkout / 'cohere').symlink_to(repo / 'cohere', target_is_directory=True)
        (checkout / 'go.work').write_text(f'go 1.27\n\nuse (\n\t.\n\t{repo / "cohere/TypeScript/tsc"}\n)\n')
    cli = scratch / f'adamic-{index}'
    build = run(['go', 'build', '-o', str(cli), './cmd/adamic'], checkout, f'stage-{index}.compiler')
    assert build['exit'] == 0, build
    stage = {'label': label, 'compiler_commit': sha, 'fixtures': []}
    for fixture in fixtures:
        binary = scratch / f'{index}-{fixture.stem}'
        build = run([str(cli), 'build', str(fixture), '-o', str(binary)], checkout, f'{index}-{fixture.stem}.build')
        if build['exit'] == 0:
            outcome = 'Compiles'
        elif "stage 0 can't lower" in build['stderr']:
            outcome = 'NotYet'
        elif 'Adamic 0.1 refuses' in build['stderr']:
            outcome = 'Refused'
        elif re.search(r'error TS[0-9]+:', build['stderr']):
            outcome = 'Checker'
        else:
            raise RuntimeError(build)
        row = {'file': fixture.name, 'outcome': outcome, 'build': build}
        if outcome == 'Compiles':
            row['native'] = run([str(binary)], checkout, f'{index}-{fixture.stem}.native')
            verify(node[fixture.name], row['native'])
            row['matchesNode'] = True
            if revision == 'working-tree':
                sanitized = scratch / f'{index}-{fixture.stem}.sanitized'
                build = run([str(cli), 'build', str(fixture), '-o', str(sanitized), '--sanitize'], checkout, f'{index}-{fixture.stem}.sanitized-build')
                assert build['exit'] == 0, build
                environment = dict(os.environ, ASAN_OPTIONS='detect_leaks=1', UBSAN_OPTIONS='halt_on_error=1')
                row['sanitized'] = run([str(sanitized)], checkout, f'{index}-{fixture.stem}.sanitized', environment)
                verify(node[fixture.name], row['sanitized'])
        stage['fixtures'].append(row)
    stage['counts'] = {kind: sum(row['outcome'] == kind for row in stage['fixtures']) for kind in ['Compiles', 'NotYet', 'Refused', 'Checker']}
    results['stages'].append(stage)
    args.output.write_text(json.dumps(results, indent=2) + '\n')
    print(label, sha[:7], stage['counts'], flush=True)
