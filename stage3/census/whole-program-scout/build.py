"""Fail-closed measurement overlay; writes only the supplied scratch directory."""
import json, pathlib, re, subprocess, sys
repo = pathlib.Path(__file__).resolve().parents[3]
out = pathlib.Path(sys.argv[1]).resolve()
out.mkdir(parents=True, exist_ok=True)
replace = {}
def put(name, source):
    target = out / name.replace('/', '_')
    target.write_text(source)
    replace[str(repo / name)] = str(target)
def once(source, old, new):
    assert source.count(old) == 1, old
    return source.replace(old, new)
s = (repo/'internal/lower/lower.go').read_text()
s = once(s, 'graphNodeArrays []graphNodeArraySite', 'graphNodeArrays []graphNodeArraySite\n graphRecords []graphRecordSite')
s = once(s, 'defer func() { err = lowering.hostMemberPending(err) }()', 'defer func() { if err != nil { stop := scoutDescribe(err); scoutUnderlying = &stop }; err = lowering.hostMemberPending(err) }()')
s = once(s, 'func LowerWithOptions(', 'func scoutLowerWithOptions(')
s = once(s, 'return LowerWithOptions(ctx, program, Options{', 'return scoutLowerWithOptions(ctx, program, Options{')
s = once(s, 'func Lower(', 'func scoutLower(')
s = once(s, 'lowering.refuse(module)', 'lowering.scoutRefuse(module)')
# Name the actual pipeline call at each error boundary. Preserve its abort.
s = re.sub(r'(\t*)(if err := ([^;\n]+); err != nil \{)', lambda m: m[1]+'scoutPhase = '+json.dumps(m[3].split('(')[0])+ '\n'+m[1]+m[2], s)
s += '''
func Lower(ctx context.Context, p *load.Program) (*ir.Program,error) { return nil, fmt.Errorf("scout: measurement only; no IR output") }
func LowerWithOptions(ctx context.Context, p *load.Program, o Options) (*ir.Program,error) { return Lower(ctx,p) }
'''
put('internal/lower/lower.go', s)
# Language's exact head compiler file, no edits; production's other files stay at main.
put('internal/lower/function_bind.go', (pathlib.Path(__file__).parent/'heads/function_bind.go.txt').read_text())
for name in ('cycles.go', 'graph_records.go', 'records.go'):
    put('internal/lower/'+name, (pathlib.Path(__file__).parent/'heads'/(name+'.txt')).read_text())
refusal = out/'refusals.go.txt'
subprocess.run(['go','run','./stage3/census/latent/refusalrewrite/cmd','-input',str(repo/'internal/lower/refusals.go'),'-output',str(refusal)],cwd=repo,check=True)
s = refusal.read_text()
# Its diagnostic-body exclusions belong to per-unit census only. All input to
# this probe has passed Load; walk every body in the coherent program.
for visitor, binding in (('contracts','contractError'), ('visit','found')):
    begin = s.index(visitor+' = func(node *ast.Node) (latentStop bool) {')
    start = s.index('\n', begin)+1
    end = s.index('\t\tdefer func() {\n\t\t\tlatentRecord('+binding+')', start)
    assert 'latentFullEnabled()' in s[start:end]
    s = s[:start]+s[end:]

# Only the clone is instrumented, leaving refuseNode and original refuse untouched.
pos = s.index('func (l *lowering) latentRefuse(')
s = s[:pos] + s[pos:].replace('prune, err := rule.check(l, node)', 'scoutPhase = "refusal_scan/" + rule.name\nprune, err := rule.check(l, node)')
put('internal/lower/refusals.go', s)
put('internal/lower/scout.go', (pathlib.Path(__file__).parent/'scout.go.txt').read_text())
put('stage3/census/whole-program-scout/tool/main.go', (pathlib.Path(__file__).parent/'main.go.txt').read_text())
(out/'overlay.json').write_text(json.dumps({'Replace':replace},indent=2)+'\n')
print(out/'overlay.json')
