from pathlib import Path
import subprocess
repo = Path('/workspace/adamic')
source = repo/'internal/native/stack_checks.go'
original = source.read_text()
mutants = [
    ('drop-one-member', 'checks[member] = recursive', 'checks[member] = recursive && member != 1', 'TestRecursiveStackComponentsAgainstReachability'),
    ('ignore-unknown', 'if unresolved {', 'if unresolved && false {', 'TestStackCheckCallTargets'),
    ('static-virtual-only', 'targets = program.CallTargets(call)', 'targets = []int{call.Function}', 'TestStackCheckCallTargets'),
    ('ignore-map-callback', 'case ir.ArrayMap, ir.ArrayVisit,', 'case ir.ArrayVisit,', 'TestStackCheckBuiltinCallbacks'),
]
try:
    for name, before, after, test in mutants:
        assert original.count(before) == 1
        source.write_text(original.replace(before, after))
        path = Path('/workspace/scratch/stack-check-scc')/(name+'.log')
        with path.open('w') as out:
            result = subprocess.run(['go', 'test', './internal/native', '-run', '^'+test+'$', '-count=1'], cwd=repo, stdout=out, stderr=subprocess.STDOUT)
        source.write_text(original)
        log = path.read_text()
        if result.returncode != 1 or '--- FAIL:' not in log:
            raise RuntimeError(name+' did not fail its assertion: '+log)
        print(name, 'caught', result.returncode)
finally:
    source.write_text(original)
