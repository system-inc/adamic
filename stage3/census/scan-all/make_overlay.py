"""Fail-closed measurement overlay; writes only the supplied scratch directory."""
import json, pathlib, re, subprocess, sys
repo = pathlib.Path(__file__).resolve().parents[3]
pins = json.loads((pathlib.Path(__file__).parent/'anchors.json').read_text())
import hashlib
for name, digest in pins.items():
    if hashlib.sha256((repo/name).read_bytes()).hexdigest() != digest:
        raise SystemExit('scan-all anchor drift: entire '+name)
if sys.argv[1:] == ['--check']:
    print('scan-all anchors match main'); raise SystemExit(0)
out = pathlib.Path(sys.argv[1]).resolve()
if out == repo or repo in out.parents:
    raise SystemExit('scan-all output must be outside the checkout')
out.mkdir(parents=True, exist_ok=True)
replace = {}
def put(name, source):
    target = out / name.replace('/', '_')
    target.write_text(source)
    replace[str(repo / name)] = str(target)
def once(source, old, new):
    if source.count(old) != 1:
        raise SystemExit('scan-all anchor must match exactly once: '+old)
    return source.replace(old, new)
s = (repo/'internal/lower/lower.go').read_text()
s = once(s, 'defer func() { err = lowering.hostMemberPending(err) }()', 'defer func() { if err != nil { stop := scoutDescribe(err); scoutUnderlying = &stop }; err = lowering.hostMemberPending(err) }()')
s = once(s, 'func LowerWithOptions(', 'func scoutLowerWithOptions(')
s = once(s, 'return LowerWithOptions(ctx, program, Options{', 'return scoutLowerWithOptions(ctx, program, Options{')
s = once(s, 'func Lower(', 'func scoutLower(')
s = once(s, 'lowering.refuse(module)', 'lowering.scoutRefuse(module)')
s = once(s, 'modules, err := lowering.moduleOrder(entry)', 'scoutPhase = "lowering.moduleOrder"\nmodules, err := lowering.moduleOrder(entry)')
s = once(s, 'body, err := lowering.moduleStatements(module)', 'scoutPhase = "lowering.moduleStatements"\nbody, err := lowering.moduleStatements(module)')
s = once(s, 'if lowering.unlowerable != nil {', 'scoutPhase = "lowering.unlowerable"\nif lowering.unlowerable != nil {')

# Name the actual pipeline call at each error boundary. Preserve its abort.
pattern = r'(\t*)(if err := ([^;\n]+); err != nil \{)'
for match in list(re.finditer(pattern, s)):
    anchor = match.group(0)
    s = once(s, anchor, match[1]+'scoutPhase = '+json.dumps(match[3].split('(')[0])+'\n'+anchor)

s += '''
func Lower(ctx context.Context, p *load.Program) (*ir.Program,error) { return nil, fmt.Errorf("scan-all: measurement only; no IR output") }
func LowerWithOptions(ctx context.Context, p *load.Program, o Options) (*ir.Program,error) { return Lower(ctx,p) }
'''
put('internal/lower/lower.go', s)
refusal = out/'refusals.go.txt'
subprocess.run(['go','run','./stage3/census/latent/refusalrewrite/cmd','-input',str(repo/'internal/lower/refusals.go'),'-output',str(refusal)],cwd=repo,check=True)
s = refusal.read_text()
# Its diagnostic-body exclusions belong to per-unit census only. All input to
# this probe has passed Load; walk every body in the coherent program.
for visitor, binding in (('contracts','contractError'), ('visit','found')):
    anchor = visitor+' = func(node *ast.Node) (latentStop bool) {'
    assert s.count(anchor) == 1, 'scan-all anchor: '+anchor
    begin = s.index(anchor)
    start = s.index('\n', begin)+1
    anchor = '\t\tdefer func() {\n\t\t\tlatentRecord('+binding+')'
    assert s.count(anchor) == 1, 'scan-all anchor: '+anchor
    end = s.index(anchor, start)
    assert 'latentFullEnabled()' in s[start:end]
    s = s[:start]+s[end:]

# Only the clone is instrumented, leaving refuseNode and original refuse untouched.
anchor = 'func (l *lowering) latentRefuse('
assert s.count(anchor) == 1, 'scan-all anchor: '+anchor
pos = s.index(anchor)
s = s[:pos] + once(s[pos:], 'prune, err := rule.check(l, node)', 'scoutPhase = "refusal_scan/" + rule.name\nprune, err := rule.check(l, node)')
# Name non-registry scans too; do not carry a previous validator's label.
head, clone = s[:pos], s[pos:]
clone = once(clone, 'for _, pragma := range module.Pragmas {', 'scoutPhase = "refusal_scan/checking-pragma"\nfor _, pragma := range module.Pragmas {')
clone = once(clone, 'contracts = func(node *ast.Node) (latentStop bool) {', 'contracts = func(node *ast.Node) (latentStop bool) {\nscoutPhase = "refusal_scan/predicateArguments"')
for match in list(re.finditer(r'if err := l\.(\w+)\(module\); err != nil \{', clone)):
    clone = once(clone, match[0], 'scoutPhase = "refusal_scan/'+match[1]+'"\n'+match[0])
s = head + clone
put('internal/lower/refusals.go', s)
put('internal/lower/scout.go', (pathlib.Path(__file__).parent/'scout.go.txt').read_text())
put('stage3/census/scan-all/tool/main.go', (pathlib.Path(__file__).parent/'main.go.txt').read_text())
(out/'overlay.json').write_text(json.dumps({'Replace':replace},indent=2)+'\n')
print(out/'overlay.json')
