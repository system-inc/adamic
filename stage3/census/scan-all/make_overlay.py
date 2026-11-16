"""Scratch-only measurement overlay; --check reads anchors without running Go."""
import argparse
import json
import pathlib
import re
import subprocess

repo = pathlib.Path(__file__).resolve().parents[3]
LOWER_REWRITES = [
    ('defer func() { err = lowering.hostMemberPending(err) }()', 'defer func() { if err != nil { stop := scoutDescribe(err); scoutUnderlying = &stop }; err = lowering.hostMemberPending(err) }()'),
    ('func LowerWithOptions(', 'func scoutLowerWithOptions('),
    ('return LowerWithOptions(ctx, program, Options{', 'return scoutLowerWithOptions(ctx, program, Options{'),
    ('func Lower(', 'func scoutLower('),
    ('lowering.refuse(module)', 'lowering.scoutRefuse(module)'),
    ('modules, err := lowering.moduleOrder(entry)', 'scoutPhase = "lowering.moduleOrder"\nmodules, err := lowering.moduleOrder(entry)'),
    ('body, err := lowering.moduleStatements(module)', 'scoutPhase = "lowering.moduleStatements"\nbody, err := lowering.moduleStatements(module)'),
    ('if lowering.unlowerable != nil {', 'scoutPhase = "lowering.unlowerable"\nif lowering.unlowerable != nil {'),
]
PIPELINE_PATTERN = r'(\t*)(if err := ([^;\n]+); err != nil \{)'
PREFLIGHT_PATTERN = r'if err := l\.(\w+)\(module\); err != nil \{'
REFUSE_SIGNATURE = 'func (l *lowering) refuse(module *ast.SourceFile) error {'
PRAGMA_ANCHOR = 'for _, pragma := range module.Pragmas {'
RULE_ANCHOR = 'prune, err := rule.check(l, node)'


def exactly_once(source, anchor, location):
    count = source.count(anchor)
    if count != 1:
        raise SystemExit(f'scan-all anchor {location}: expected exactly once (found {count}): {anchor}')
    return source.index(anchor)


def once(source, old, new):
    exactly_once(source, old, 'generated source')
    return source.replace(old, new)


def sites(pattern, source, location, required=False):
    matches = list(re.finditer(pattern, source))
    if required and not matches:
        raise SystemExit('scan-all anchor '+location+': regex found no sites: '+pattern)
    for match in matches:
        exactly_once(source, match[0], location)
    return matches


def refuse_source(source):
    start = exactly_once(source, REFUSE_SIGNATURE, 'internal/lower/refusals.go')
    # Balance Go braces while ignoring comments and quoted/raw strings. Do not
    # depend on the following function's name or on unrelated file contents.
    tokens = re.compile(r"//[^\n]*|/\*[\s\S]*?\*/|\"(?:\\.|[^\"\\])*\"|'(?:\\.|[^'\\])*'|`[^`]*`|[{}]")
    depth = 0
    for token in tokens.finditer(source, start):
        if token[0] == '{':
            depth += 1
        elif token[0] == '}':
            depth -= 1
            if depth == 0:
                return source[start:token.end()]
    raise SystemExit('scan-all anchor internal/lower/refusals.go: unclosed '+REFUSE_SIGNATURE)


def check_sources(root):
    lower = (root/'internal/lower/lower.go').read_text()
    for anchor, _ in LOWER_REWRITES:
        exactly_once(lower, anchor, 'internal/lower/lower.go')
    sites(PIPELINE_PATTERN, lower, 'internal/lower/lower.go pipeline-site', required=True)
    refusal = refuse_source((root/'internal/lower/refusals.go').read_text())
    for anchor in (PRAGMA_ANCHOR, RULE_ANCHOR,
                   'contracts = func(node *ast.Node) bool {',
                   'visit = func(node *ast.Node) bool {'):
        exactly_once(refusal, anchor, 'internal/lower/refusals.go refuse')
    preflights = sites(PREFLIGHT_PATTERN, refusal, 'internal/lower/refusals.go refuse preflight')
    return lower, preflights


parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--check', action='store_true')
parser.add_argument('output', nargs='?', type=pathlib.Path)
args = parser.parse_args()
lower_source, refusal_preflights = check_sources(repo)
if args.check:
    print('scan-all source anchors match exactly once; pipeline sites present')
    raise SystemExit(0)
if args.output is None:
    parser.error('output is required unless --check is used')
out = args.output.resolve()
if out == repo or repo in out.parents:
    raise SystemExit('scan-all output must be outside the checkout')
out.mkdir(parents=True, exist_ok=True)
replace = {}


def put(name, source):
    target = out / name.replace('/', '_')
    target.write_text(source)
    replace[str(repo / name)] = str(target)


s = lower_source
for old, new in LOWER_REWRITES:
    s = once(s, old, new)

# Name the actual pipeline call at each error boundary. Preserve its abort.
for match in sites(PIPELINE_PATTERN, s, 'generated lower.go pipeline-site', required=True):
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
    begin = exactly_once(s, anchor, 'post-rewrite '+visitor+' signature')
    start = s.index('\n', begin)+1
    anchor = '\t\tdefer func() {\n\t\t\tlatentRecord('+binding+')'
    end = exactly_once(s, anchor, 'post-rewrite '+visitor+' recorder defer')
    exactly_once(s[start:end], 'latentFullEnabled()', 'post-rewrite '+visitor+' exclusions')
    s = s[:start]+s[end:]

# Only the clone is instrumented, leaving refuseNode and original refuse untouched.
anchor = 'func (l *lowering) latentRefuse('
pos = exactly_once(s, anchor, 'post-rewrite clone')
s = s[:pos] + once(s[pos:], RULE_ANCHOR, 'scoutPhase = "refusal_scan/" + rule.name\nprune, err := rule.check(l, node)')
# Name non-registry scans too; do not carry a previous validator's label.
head, clone = s[:pos], s[pos:]
clone = once(clone, PRAGMA_ANCHOR, 'scoutPhase = "refusal_scan/checking-pragma"\nfor _, pragma := range module.Pragmas {')
clone = once(clone, 'contracts = func(node *ast.Node) (latentStop bool) {', 'contracts = func(node *ast.Node) (latentStop bool) {\nscoutPhase = "refusal_scan/predicateArguments"')
for match in refusal_preflights:
    clone = once(clone, match[0], 'scoutPhase = "refusal_scan/'+match[1]+'"\n'+match[0])
s = head + clone
put('internal/lower/refusals.go', s)
put('internal/lower/scout.go', (pathlib.Path(__file__).parent/'scout.go.txt').read_text())
put('stage3/census/scan-all/tool/main.go', (pathlib.Path(__file__).parent/'main.go.txt').read_text())
(out/'overlay.json').write_text(json.dumps({'Replace':replace},indent=2)+'\n')
print(out/'overlay.json')
