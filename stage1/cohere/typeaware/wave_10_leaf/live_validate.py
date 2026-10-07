"""Live owned handed-node adapters against unchanged production Go rules."""
import argparse
import json
import statistics
import subprocess
import time
from pathlib import Path

parser = argparse.ArgumentParser()
parser.add_argument('--stage0', required=True)
parser.add_argument('--checker', required=True)
parser.add_argument('--checker-asan', required=True)
parser.add_argument('--isolated', required=True)
parser.add_argument('--artifacts', required=True)
parser.add_argument('--compiler-config', required=True)
parser.add_argument('--compiler-manifest', required=True)
parser.add_argument('--repository-manifest', required=True)
args = parser.parse_args()
owned = Path(__file__).resolve().parent
repository = owned.parents[3]
artifacts = Path(args.artifacts).resolve()
artifacts.mkdir(parents=True, exist_ok=True)
isolated = Path(args.isolated).resolve()


def run(label, command):
    started = time.perf_counter()
    result = subprocess.run(command, cwd=repository, capture_output=True)
    elapsed = time.perf_counter() - started
    (artifacts / (label + '.stdout')).write_bytes(result.stdout)
    (artifacts / (label + '.stderr')).write_bytes(result.stderr)
    assert result.returncode == 0, (label, result.returncode, result.stderr)
    return result, elapsed


binaries = {}
for sanitize in [False, True]:
    label = 'live-asan' if sanitize else 'live'
    binary = artifacts / label
    command = [args.stage0, 'build', str(owned / 'symbol_description/live.a'), '-o', str(binary), '--tsgo', args.checker_asan if sanitize else args.checker]
    if sanitize:
        command.append('--sanitize')
    run(label + '-build', command)
    binaries[sanitize] = binary

catalog = str(isolated / 'catalog.txt')
commands = {}
truths = {}
for label, name in [('symbol', 'symbol_description'), ('hook', 'react_hook_no_any_type')]:
    oracle = isolated / (label + '-oracle')
    datasets = [
        ('controls', str(isolated / 'tsconfig.json'), str(isolated / (name + '.manifest'))),
        ('compiler', args.compiler_config, args.compiler_manifest),
        ('repository', str(repository / 'tsconfig.json'), args.repository_manifest),
    ]
    for dataset, config, manifest in datasets:
        key = label + '-' + dataset
        go_command = [str(oracle), config, manifest]
        native_command = [str(binaries[False]), config, manifest, label, catalog]
        truth, _ = run(key + '-go', go_command)
        truths[key] = truth.stdout
        for sanitize in [False, True]:
            command = [str(binaries[sanitize]), config, manifest, label, catalog]
            observed, _ = run(key + ('-asan' if sanitize else '-native'), command)
            assert observed.stdout == truth.stdout and observed.stderr == b'', (key, 'byte comparison')
        commands[key] = (go_command, native_command)
        print(key, 'full finding/fix/suggestion bytes PASS', len(truth.stdout), flush=True)
    config = str(isolated / 'tsconfig.json')
    manifest = str(isolated / (name + '.manifest'))
    binary = artifacts / (label + '-mutant')
    source = isolated / (label + '-mutant-source') / 'symbol_description/live.a'
    run(label + '-mutant-build', [args.stage0, 'build', str(source), '-o', str(binary), '--tsgo', args.checker])
    observed, _ = run(label + '-mutant-run', [str(binary), config, manifest, label, catalog])
    assert observed.stderr == b'' and observed.stdout != truths[label + '-controls'], (label, 'mutant survives')
    truth = truths[label + '-controls']
    byte = next((i for i, (a, b) in enumerate(zip(observed.stdout, truth)) if a != b), min(len(observed.stdout), len(truth)))
    print(label, 'live mutant exits 0, empty stderr, comparison catches byte', byte, flush=True)
    for sanitize in [False, True]:
        result = subprocess.run([str(binaries[sanitize]), config, manifest, label, catalog, '--released'], capture_output=True)
        prefix = label + ('-released-asan' if sanitize else '-released')
        (artifacts / (prefix + '.stdout')).write_bytes(result.stdout)
        (artifacts / (prefix + '.stderr')).write_bytes(result.stderr)
        assert result.returncode == 70 and result.stderr == b'adamic: panic: invalid or released checker handle\n', (prefix, result.returncode, result.stderr)
    print(label, 'released handle normal/sanitized rejected at exit 70', flush=True)

observations = {}
for key, (go_command, native_command) in commands.items():
    go_times = []
    native_times = []
    for round_index in range(3):
        for implementation, command, samples in [('go', go_command, go_times), ('native', native_command, native_times)]:
            result, elapsed = run(key + '-bench-' + implementation + '-' + str(round_index), command)
            assert result.stdout == truths[key]
            if implementation == 'native':
                assert result.stderr == b''
            samples.append(elapsed)
    go_median = statistics.median(go_times)
    native_median = statistics.median(native_times)
    observations[key] = {'go_seconds': go_times, 'native_seconds': native_times, 'go_median': go_median, 'native_median': native_median, 'native_over_go': native_median / go_median}
    print(key, 'native', native_median, 'Go', go_median, 'ratio', native_median / go_median, flush=True)
(artifacts / 'timings.json').write_text(json.dumps(observations, indent=2) + '\n')
