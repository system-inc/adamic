"""Production Go authority for require-await reporting only, not its verdict."""
import argparse
import json
import subprocess
from pathlib import Path

parser = argparse.ArgumentParser()
parser.add_argument('--stage0', required=True)
parser.add_argument('--checker', required=True)
parser.add_argument('--checker-asan', required=True)
parser.add_argument('--oracle', required=True)
parser.add_argument('--artifacts', required=True)
args = parser.parse_args()
owned = Path(__file__).resolve().parent
repository = owned.parents[4]
artifacts = Path(args.artifacts).resolve()
artifacts.mkdir(parents=True, exist_ok=True)
cases = json.loads((owned / 'reporting_controls.json').read_text())
for name, text in cases.items():
    (artifacts / name).write_text(text)
manifest = artifacts / 'manifest'
manifest.write_text(''.join(str(artifacts / name) + '\n' for name in sorted(cases)))
config = artifacts / 'tsconfig.json'
config.write_text('{"compilerOptions":{"strict":true,"target":"ES2022"}}\n')


def run(label, command):
    result = subprocess.run(command, cwd=repository, capture_output=True)
    (artifacts / (label + '.stdout')).write_bytes(result.stdout)
    (artifacts / (label + '.stderr')).write_bytes(result.stderr)
    assert result.returncode == 0, (label, result.returncode, result.stderr)
    return result


truth = run('go', [args.oracle, str(config), str(manifest)]).stdout
for sanitize in [False, True]:
    label = 'report-asan' if sanitize else 'report'
    binary = artifacts / label
    command = [args.stage0, 'build', str(owned / 'reporting_controls.a'), '-o', str(binary), '--tsgo', args.checker_asan if sanitize else args.checker]
    if sanitize:
        command.append('--sanitize')
    run(label + '-build', command)
    observed = run(label + '-run', [str(binary), str(manifest)])
    assert observed.stdout == truth and observed.stderr == b'', 'full reporting comparison'
    print(label, '32 reporting candidates, all finding/fix/suggestion bytes PASS', len(truth), flush=True)
mutant = artifacts / 'mutant-source'
for source in owned.glob('*.a'):
    text = source.read_text()
    if source.name == 'reporting.a':
        before = 'context.byte(range.end),'
        assert text.count(before) == 1
        text = text.replace(before, 'context.byte(range.end + 1),', 1)
    text = text.replace("'../../../../typescript/", "'" + str(repository / 'stage1/typescript') + "/")
    text = text.replace("'../symbol_description/context.a'", "'" + str(owned.parent / 'symbol_description/context.a') + "'")
    for module in ['diagnostic.ts', 'suggestion.ts', 'repair.ts', 'unary_minus.ts']:
        text = text.replace("'../../" + module + "'", "'" + str(repository / 'stage1/cohere/typeaware' / module) + "'")
    mutant.mkdir(exist_ok=True)
    (mutant / source.name).write_text(text)
binary = artifacts / 'report-mutant'
run('mutant-build', [args.stage0, 'build', str(mutant / 'reporting_controls.a'), '-o', str(binary), '--tsgo', args.checker])
observed = run('mutant-run', [str(binary), str(manifest)])
assert observed.stdout != truth and observed.stderr == b'', 'reporting mutant survived'
byte = next((i for i, (a, b) in enumerate(zip(observed.stdout, truth)) if a != b), min(len(observed.stdout), len(truth)))
print('reporting range mutant exits 0, empty stderr, comparison catches byte', byte, flush=True)
