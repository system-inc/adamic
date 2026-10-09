"""Exercise independent initialization pins; restore every mutation afterward."""
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[3]

def mutant(name, relative, before, after, pattern, evidence):
    subject = root / relative
    original = subject.read_text()
    assert original.count(before) == 1, (name, original.count(before))
    log = Path('/tmp/enum-init-' + name + '.log')
    try:
        subject.write_text(original.replace(before, after, 1))
        with log.open('w') as output:
            result = subprocess.run(['go', 'test', './internal/lower' if name in ('unreaching', 'no-memo') else './internal/oracle', '-run', pattern, '-count=1', '-v'], cwd=root, stdout=output, stderr=subprocess.STDOUT, timeout=120)
        observed = log.read_text()
        assert result.returncode != 0 and evidence in observed and 'build failed' not in observed, observed
        print(name, 'caught:', evidence, log)
    finally:
        subject.write_text(original)

mutant('unreaching', 'internal/lower/namespaces_call_graph.go',
       '\t\t\tif declaration := g.lowering.namespaceRuntimeEnum(node); declaration != nil {\n\t\t\t\tentry.reads[declaration] = node\n\t\t\t}\n', '',
       '^TestEnumInitializationReach/reaching-', 'pending enum read on line 2 must stay refused: <nil>')
mutant('no-memo', 'internal/lower/namespaces_call_graph.go',
       'if cached := g.functions[function]; cached != nil {',
       'if cached := g.functions[function]; cached != nil && cached.active {',
       '^TestEnumInitializationGraphMemo/12$', 'nonlinear enum function walks: got 8191, want 13')
mutant('no-readiness', 'internal/lower/locals.go',
       'Checked: l.result.Locals[local].NamespaceState || l.checkedModuleRead(node, local)',
       'Checked: (l.result.Locals[local].NamespaceState || l.checkedModuleRead(node, local)) && l.namespaceRuntimeEnum(node) == nil',
       '^TestEnumInitializationUnknownPinned$', '--- FAIL: TestEnumInitializationUnknownPinned')
mutant('wrong-callable-branch', 'internal/lower/literal_callable_branch.go',
       'case ast.KindTrueKeyword:\n\t\treturn expression.WhenTrue',
       'case ast.KindTrueKeyword:\n\t\treturn expression.WhenFalse',
       '^TestEnumInitializationNode/callable-selection$', 'stdout differs')
