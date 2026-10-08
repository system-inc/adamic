"""Run the bounded fixture gate and prove its comparison and pending-native checks fail."""
import argparse
import pathlib
import subprocess
parser = argparse.ArgumentParser()
parser.add_argument('--runtime-js-compiler', default='')
args = parser.parse_args()
root = pathlib.Path(__file__).resolve().parent
repository = root.parents[5]
logs = root / 'evidence'
logs.mkdir(exist_ok=True)
command = ['go', 'run', str(root / 'gate.go')]
extra = ['-runtime-js-compiler', args.runtime_js_compiler] if args.runtime_js_compiler else []
def run(name, arguments):
    path = logs / (name + '.log')
    with path.open('wb') as output:
        result = subprocess.run(command + arguments, cwd=repository, stdout=output, stderr=subprocess.STDOUT)
    return result.returncode, path.read_text()
code, baseline = run('gate', extra)
if code != 0:
    raise SystemExit('fixture gate failed; see evidence/gate.log')
code, mutant = run('mutant', ['-mutant'])
if code == 0 or 'base/consistency_require_pagination_argument_name.go:27: source Node fixture mismatch' not in mutant:
    raise SystemExit('translation mutant was not killed by its match comparison')
code, forced = run('forced-native', ['-force-native'])
if 'awaits codex/regex-runtime-compiler=107' in baseline:
    if code == 0 or 'forced native requirement caught refusal' not in forced:
        raise SystemExit('forced native check did not expose the named pending refusal')
elif code != 0:
    raise SystemExit('runtime compiler now accepts the path, so native is required, not pending')
print(baseline.strip())
print('PASS one-character translation mutant and forced native requirement controls')
