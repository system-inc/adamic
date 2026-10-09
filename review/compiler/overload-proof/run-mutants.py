#!/usr/bin/env python3
"""Run each mutation independently and restore the compiler even on failure."""
import difflib
import pathlib
import subprocess

root = pathlib.Path(__file__).resolve().parents[3]
evidence = root / 'review/compiler/overload-proof'
body = root / 'internal/lower/overload_body_proof.go'
entry = root / 'internal/lower/census_overload_proof.go'
body_original = body.read_text()
entry_original = entry.read_text()

start = entry_original.index('func (l *lowering) censusProveOverloadResult(')
end = entry_original.index('\nfunc (l *lowering) censusReturnProof(', start)
unsound = entry_original[:start] + '''func (l *lowering) censusProveOverloadResult(implementation, overload *ast.Node) bool {
    promised := l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(overload))
    accepted := false
    var scan ast.Visitor
    scan = func(node *ast.Node) bool {
        if ast.IsFunctionLike(node) { return false }
        if node.Kind == ast.KindReturnStatement {
            expression := node.AsReturnStatement().Expression
            if expression != nil && l.censusRelated(l.checker.GetTypeAtLocation(expression), promised) { accepted = true }
            return false
        }
        node.ForEachChild(scan)
        return false
    }
    implementation.Body().ForEachChild(scan)
    return accepted
}
''' + entry_original[end:]

mutants = [
    ('throw-falls-through', body, body_original.replace('operand can only leave abruptly. Neither outcome returns a value.\n\t\treturn nil, true', 'operand can only leave abruptly. Neither outcome returns a value.\n\t\treturn states, true', 1),
     './internal/lower', '^TestOverloadBodyObligationEarlyThrow$', 'proof=false, want true'),
    ('hidden-accessor-is-data', body, body_original.replace(' || p.possibleAccessor(access.Name().Text())', '').replace(' || p.possibleAccessor(current.Name().Text())', ''),
     './internal/lower', '^TestOverloadBodyObligationHiddenAccessor$', 'proof=true, want false'),
    ('full-implementation-types', body, body_original.replace('state[symbol] = given', 'state[symbol] = l.checker.GetTypeAtLocation(parameter.Name())', 1),
     './internal/oracle', '^TestOverloadBodyProof(Block|Evaluate)$', 'FAIL: TestOverloadBodyProof'),
    ('trust-overload-annotation', entry, unsound,
     './internal/oracle', '^TestOverloadBodyProofRefusesNumericReturn$', 'numeric return must be refused, got program=true error=<nil>'),
]
try:
    for name, target, mutated, package, test, catcher in mutants:
        original = target.read_text()
        if original == mutated:
            raise SystemExit(f'{name}: mutation did not change the source')
        (evidence / (name + '.patch')).write_text(''.join(difflib.unified_diff(original.splitlines(True), mutated.splitlines(True), fromfile=str(target.relative_to(root)), tofile=str(target.relative_to(root)))))
        target.write_text(mutated)
        command = ['go', 'test', package, '-run', test, '-count=1', '-timeout', '90s', '-v']
        log = evidence / (name + '.log')
        try:
            with log.open('w') as output:
                ran = subprocess.run(command, cwd=root, stdout=output, stderr=subprocess.STDOUT, timeout=120)
            observed = log.read_text()
            if ran.returncode != 1 or catcher not in observed or '[build failed]' in observed:
                raise SystemExit(f'{name}: not killed by its intended assertion; exit={ran.returncode}, log={log}')
            print(f'{name}: caught, exit 1, {catcher}', flush=True)
        finally:
            target.write_text(original)
finally:
    body.write_text(body_original)
    entry.write_text(entry_original)
