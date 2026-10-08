"""Run serially from the repository root; compiler/test sources are restored."""
from pathlib import Path
import subprocess

out = Path('/tmp/graph-review-mutants')
out.mkdir(parents=True, exist_ok=True)
flow = Path('internal/lower/graph_flow.go')
native = Path('internal/native/graph_regions_test.go')
original_flow = flow.read_text()
original_native = native.read_text()
def caught(name, package, pattern):
    with (out / (name + '.log')).open('w') as log:
        result = subprocess.run(['go', 'test', package, '-run', pattern, '-count=1', '-v', '-timeout=10m'], stdout=log, stderr=subprocess.STDOUT)
    print(name, result.returncode, flush=True)
    if result.returncode == 0:
        raise RuntimeError('mutant survived: ' + name)
try:
    old = '''\t\tcase ir.Call:
\t\t\tfor _, target := range program.CallTargets(value) {
\t\t\t\tdemand(resultNode(target))
\t\t\t}'''
    assert original_flow.count(old) == 1
    flow.write_text(original_flow.replace(old, '\t\tcase ir.Call:\n\t\t\tdemand(resultNode(value.Function))'))
    caught('base-only-guard', './internal/ir', '^TestCallTargetReaders$')
    caught('base-only-classification', './internal/lower', '^TestGraphAllocationFlowClassification/override$')
    caught('base-only-oracle', './internal/oracle', '^TestGraphAllocationFlowIsLeakClean/override$')
    flow.write_text(original_flow)
    old = 'if (strcmp(arguments[1], "leak") != 0) { adamic_release(root); }'
    assert original_native.count(old) == 1
    native.write_text(original_native.replace(old, 'adamic_release(root);'))
    caught('anchor-mutant-disabled', './internal/native', '^TestGraphRegionsRuntime$|^TestGraphUnreleasedAnchorCounted$')
finally:
    flow.write_text(original_flow)
    native.write_text(original_native)
