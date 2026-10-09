"""Remove completed-function memoization; active cycles still terminate."""
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[2]
subject = root / 'internal/lower/namespaces_call_graph.go'
original = subject.read_text()
needle = 'if cached := g.functions[function]; cached != nil {'
assert needle in original
try:
    subject.write_text(original.replace(needle, 'if cached := g.functions[function]; cached != nil && cached.active {', 1))
    log = Path('/tmp/namespaces-graph-no-memo.log')
    with log.open('w') as output:
        result = subprocess.run(['go', 'test', './internal/lower', '-run', '^TestNamespaceCallGraphLinearWork$/12$', '-count=1', '-v'], cwd=root, stdout=output, stderr=subprocess.STDOUT, timeout=120)
    observed = log.read_text()
    assert result.returncode != 0 and 'nonlinear function walks: got 8191, want 13' in observed, observed
    assert 'build failed' not in observed, observed
    print('drop memo caught by body-expansion count: 8191 rather than 13;', log)
finally:
    subject.write_text(original)

# A cycle must share the union, rather than caching each member's local reads.
try:
    subject.write_text(original.replace('member.reaches = reaches', 'member.reaches = member.reads', 1))
    log = Path('/tmp/namespaces-graph-partial-cycle.log')
    with log.open('w') as output:
        result = subprocess.run(['go', 'test', './internal/lower', '-run', '^TestNamespaceCallGraphCycleUnion$', '-count=1', '-v'], cwd=root, stdout=output, stderr=subprocess.STDOUT, timeout=120)
    observed = log.read_text()
    assert result.returncode != 0 and 'partial cycle reach' in observed and 'build failed' not in observed, observed
    print('partial cycle union caught by reach-set comparison;', log)
finally:
    subject.write_text(original)

# Integration's module analysis cannot silently erase the namespace enum boundary.
needle = '\t\t\tif declaration := g.lowering.namespaceRuntimeEnum(node); declaration != nil {\n\t\t\t\tentry.reads[declaration] = node\n\t\t\t}\n'
assert needle in original
try:
    subject.write_text(original.replace(needle, '', 1))
    log = Path('/tmp/namespaces-graph-early-enum.log')
    with log.open('w') as output:
        result = subprocess.run(['go', 'test', './internal/lower', '-run', '^TestNamespaceEnumInitializationIndependentOfModuleAnalysis$', '-count=1', '-v'], cwd=root, stdout=output, stderr=subprocess.STDOUT, timeout=120)
    observed = log.read_text()
    assert result.returncode != 0 and 'namespace enum lost its independent refusal' in observed and 'build failed' not in observed, observed
    print('lost early enum reach caught by independent namespace preflight;', log)
finally:
    subject.write_text(original)
