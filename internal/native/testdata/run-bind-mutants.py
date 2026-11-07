#!/usr/bin/env python3
"""Plant one bind fault at a time; require its focused check to fail semantically."""
import json
import os
from pathlib import Path
import shlex
import subprocess

root = Path(__file__).resolve().parents[3]
os.chdir(root)
logs = Path('/tmp/adamic-bind-mutants')
logs.mkdir(exist_ok=True)
lower = Path('internal/lower/function_bind.go')
emit = Path('internal/native/emit_expressions.go')
flow = Path('internal/flow/async_normalize.go')
originals = {path: path.read_text() for path in [lower, emit, flow]}

def change(text, old, new):
    if text.count(old) != 1:
        raise RuntimeError(f'expected one mutation site: {old!r}, found {text.count(old)}')
    return text.replace(old, new, 1)

def repeated_fixed(text):
    return change(text, 'arguments := append([]ir.Expression{}, reads[2:]...)', 'arguments := append([]ir.Expression{}, values[2:]...)')

def widened_path(text):
    text = change(text, 'method = ast.IsIdentifier(receiver) && ast.IsIdentifier(argument) && l.symbol(receiver) == l.symbol(argument)', 'method = sourceExpression(receiver) == sourceExpression(argument)')
    text = change(text, 'if !ast.IsIdentifier(receiver) {', 'if !ast.IsIdentifier(receiver) && receiver.Kind != ast.KindPropertyAccessExpression {')
    text = change(text, 'same := ast.IsIdentifier(argument) && l.symbol(receiver) == l.symbol(argument)', 'same := sourceExpression(receiver) == sourceExpression(argument)')
    text = change(text, 'return ast.IsIdentifier(receiver) && ast.IsIdentifier(argument) && l.symbol(receiver) == l.symbol(argument) && l.stableBindReceiver(receiver) && l.bindReceiverRule(outer) == nil', 'return sourceExpression(receiver) == sourceExpression(argument) && l.stableBindReceiver(receiver) && l.bindReceiverRule(outer) == nil')
    return change(text, 'func (l *lowering) stableBindReceiver(node *ast.Node) bool {', 'func (l *lowering) stableBindReceiver(node *ast.Node) bool {\n if node.Kind == ast.KindPropertyAccessExpression { return l.stableBindReceiver(node.AsPropertyAccessExpression().Expression) }')

def dropped_links(text):
    return change(text, 'func (l *lowering) closureLinks(closure closureRecord) []int {', 'func (l *lowering) closureLinks(closure closureRecord) []int {\n if l.result.Functions[closure.function].Source.Role == "bound_function" { return nil }')

def late_lookup(text):
    return change(text, 'invoke.BoundName = target.Name().Text()', 'invoke.Bound = false\n invoke.Closure = ir.Property{Object: reads[1], Name: target.Name().Text(), Of: ir.Closure, Method: true}')

def dropped_optional(text):
    return change(text, 'if !expression.Optional {\n\t\treturn makeValue()\n\t}', 'if true {\n\t\treturn makeValue()\n\t}')

def allowed_suspension(text):
    return change(text, '\tcase ir.BoundMethod:\n\t\t// The lookup and nullish guard must precede a suspending fixed operand.\n\t\t// Generic operand normalization would move that operand before both.\n\t\tn.failure = n.errorAt(value, "suspension inside method binding operands is not yet proven")\n\t\treturn value\n', '')

mutants = [
 ('repeated-fixed', lower, repeated_fixed, '^TestFunctionBind$', 'javascript: stdout differs'),
 ('widened-path', lower, widened_path, '^TestMethodBindRefusals$/^path$', 'got <nil>'),
 ('dropped-links', lower, dropped_links, '^TestMethodBind$/^bind_cycle_cache_bound[.]a$', 'ERROR: LeakSanitizer'),
 ('dropped-optional', emit, dropped_optional, '^TestMethodBind$/^method_bind[.]a$', 'native stdout differs'),
 ('late-lookup', lower, late_lookup, '^TestMethodBind$/^method_bind[.]a$', 'javascript stdout differs'),
 ('allowed-suspension', flow, allowed_suspension, '^TestMethodBindRefusals$/^async_fixed$', 'got <nil>'),
]
results = []
try:
    for name, path, mutate, focused, catcher in mutants:
        path.write_text(mutate(originals[path]))
        log = logs / (name + '.log')
        command = 'source /workspace/adamic-tools/env.sh\nADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run ' + shlex.quote(focused) + ' -count=1 -timeout 10m'
        try:
            with log.open('w') as output:
                status = subprocess.run(['bash', '-lc', command], stdout=output, stderr=subprocess.STDOUT).returncode
        finally:
            path.write_text(originals[path])
        subprocess.run(['git', 'diff', '--exit-code'], check=True, stdout=subprocess.DEVNULL)
        output = log.read_text()
        caught = status != 0 and catcher in output and '[build failed]' not in output and 'error: ' not in output
        evidence = next((line.strip() for line in output.splitlines() if catcher in line), '')
        result = dict(mutant=name, test=focused, exit=status, caught=caught, evidence=evidence, log=str(log))
        results.append(result)
        (logs / 'results.json').write_text(json.dumps(results, indent=2) + '\n')
        print(json.dumps(result), flush=True)
finally:
    for path, source in originals.items():
        path.write_text(source)
    subprocess.run(['git', 'diff', '--exit-code'], check=True, stdout=subprocess.DEVNULL)
if not all(result['caught'] for result in results):
    raise SystemExit('one or more mutants were not caught by the required semantic check')
print(f'{len(mutants)} bind mutants caught; every tracked source restored byte for byte', flush=True)
