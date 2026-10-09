"""Run from the repository root with the toolchain environment sourced."""
from pathlib import Path
import subprocess
import sys

out = Path(sys.argv[1] if len(sys.argv) > 1 else '/tmp/graph-walk-mutants')
out.mkdir(parents=True, exist_ok=True)
flow = Path('internal/lower/graph_flow.go')
ir = Path('internal/ir/ir.go')
original_flow = flow.read_text()
original_ir = ir.read_text()
try:
    for kind, following in [('Ptr', 'Map'), ('Map', 'Struct')]:
        start = original_flow.index('\t\tcase reflect.' + kind + ':')
        end = original_flow.index('\t\tcase reflect.' + following + ':', start)
        flow.write_text(original_flow[:start] + '\t\tcase reflect.' + kind + ':\n\t\t\treturn value\n' + original_flow[end:])
        with (out / ('skip-' + kind.lower() + '.log')).open('w') as log:
            result = subprocess.run(['go', 'test', './internal/lower', '-run', '^TestGraphAllocationSiteContainers$', '-count=1', '-v'], stdout=log, stderr=subprocess.STDOUT)
        print('skip-' + kind.lower(), result.returncode, flush=True)
        if result.returncode == 0:
            raise RuntimeError('traversal mutant survived: ' + kind)
        flow.write_text(original_flow)
    # A valid future allocation carrying ownership metadata must demand a new
    # traversal probe, even before lowering ever emits the new expression.
    ir.write_text(original_ir + '\ntype GraphWalkFutureAllocation struct { GraphTypes []int }\nfunc (GraphWalkFutureAllocation) Type() Type { return Object }\n')
    with (out / 'new-allocation.log').open('w') as log:
        result = subprocess.run(['go', 'test', './internal/lower', '-run', '^TestGraphAllocationSiteContainers$', '-count=1', '-v'], stdout=log, stderr=subprocess.STDOUT)
    print('new-allocation', result.returncode, flush=True)
    if result.returncode == 0:
        raise RuntimeError('new allocation bypassed the inventory guard')
finally:
    flow.write_text(original_flow)
    ir.write_text(original_ir)
