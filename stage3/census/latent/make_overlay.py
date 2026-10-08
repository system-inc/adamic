"""Build a scratch-only lower overlay. Production Lower is disabled in this binary."""
import json
import pathlib
import sys
import subprocess

repository = pathlib.Path(sys.argv[1]).resolve()
scratch = pathlib.Path(sys.argv[2]).resolve()
territory = pathlib.Path(__file__).resolve().parent
scratch.mkdir(parents=True, exist_ok=True)
replace = {}

def overlay(name, text):
    output = scratch / name.replace('/', '_')
    output.write_text(text)
    replace[str(repository / name)] = str(output)

original = (repository / 'internal/lower/lower.go').read_text()
start = original.index('\tfiles := program.Files()')
end = original.index('\n}\n\ntype lowering', start)
original = original[:start] + '\treturn nil, fmt.Errorf("latent census: measurement only; no IR output")' + original[end:]
original = original.replace('\n\t"path/filepath"', '')
original = original.replace('type lowering struct {', 'type lowering struct {\n latentDeclarations map[int]*ast.Node\n latentReady map[int]bool\n latentAncestors map[*ast.Node]bool')
overlay('internal/lower/lower.go', original)
# Generate typed snapshots of all lowering-owned state; unknown shapes fail loudly.
state_output = scratch / 'internal_lower_latent_state.go'
subprocess.run(['go', 'run', str(territory / 'statecopy'), str(repository),
                str(scratch / 'internal_lower_lower.go'), str(state_output)], check=True, cwd=territory)
replace[str(repository / 'internal/lower/latent_state.go')] = str(state_output)
statement_output = scratch / 'internal_lower_statements.go'
subprocess.run(['go', 'run', str(territory / 'statementrewrite'),
                str(repository / 'internal/lower/statements.go'), str(statement_output)], check=True, cwd=territory)
replace[str(repository / 'internal/lower/statements.go')] = str(statement_output)
overlay('internal/lower/latent_full.go', (territory / 'full.go.txt').read_text())
overlay('internal/lower/latent_units.go', (territory / 'units.go.txt').read_text())
overlay('internal/lower/latent_replay.go', (territory / 'replay.go.txt').read_text())
overlay('internal/lower/latent_provenance.go', (territory / 'provenance.go.txt').read_text())

# Parse the exact refusal function and its visitor scopes; fail on unsupported shapes.
refusal_output = scratch / 'internal_lower_refusals.go'
subprocess.run(['go', 'run', str(territory / 'refusalrewrite/cmd'),
                '-input', str(repository / 'internal/lower/refusals.go'),
                '-output', str(refusal_output)], check=True, cwd=territory)
replace[str(repository / 'internal/lower/refusals.go')] = str(refusal_output)
# Keep all checker diagnostics and populated roots, but expose them only to measurement.
loader = (repository / 'internal/load/load.go').read_text()
loader = loader.replace('type Program struct {', 'type Program struct {\n latentSites []*ast.Diagnostic\n latentDiagnostics []string')
needle = 'return nil, &CheckError{Diagnostics: diagnostics}'
assert loader.count(needle) == 1
loader = loader.replace(needle, 'loaded.latentDiagnostics = diagnostics')
loader = loader.replace('formatted := make([]string, 0, len(all))', 'p.latentSites = all\n formatted := make([]string, 0, len(all))')
# Rename the permissive API; ordinary Load and LoadOverlay cannot expose rejected programs.
loader = loader.replace('func Load(paths []string)', 'func LatentLoad(paths []string)')
loader = loader.replace('func LoadOverlay(paths []string, overlay map[string]string)', 'func latentLoadOverlay(paths []string, overlay map[string]string)')
loader += '\nfunc Load(paths []string) (*Program, error) { return nil, fmt.Errorf("latent census: measurement loader only; no output path") }\nfunc LoadOverlay(paths []string, overlay map[string]string) (*Program, error) { return Load(paths) }\n'
overlay('internal/load/load.go', loader)
overlay('internal/load/latent_hook.go', (territory / 'load.go.txt').read_text())
# Direct sibling calls need signatures, not sibling bodies. Prepare signatures lazily.
expressions = (repository / 'internal/lower/expression.go').read_text()
needle = 'func (l *lowering) callFunction(call *ast.CallExpression, function int) (ir.Expression, error) {'
assert needle in expressions
expressions = expressions.replace(needle, needle + '\n if err := l.latentSignature(function); err != nil { return nil, err }')
needle = 'func (l *lowering) functionValue(node *ast.Node, target int) (ir.Expression, error) {'
assert needle in expressions
expressions = expressions.replace(needle, needle + '\n if err := l.latentSignature(target); err != nil { return nil, err }')
needle = 'return nil, l.notYet(node, "reading "+node.Text())'
assert expressions.count(needle) == 1
expressions = expressions.replace(needle, 'return nil, l.latentReadNotYet(node, l.symbol(node), node.Text())')
overlay('internal/lower/expression.go', expressions)
objects = (repository / 'internal/lower/object.go').read_text()
needle = 'return nil, l.notYet(property, "reading "+property.Name().Text())'
assert objects.count(needle) == 1
objects = objects.replace(needle, 'return nil, l.latentReadNotYet(property, symbol, property.Name().Text())')
overlay('internal/lower/object.go', objects)
functions = (repository / 'internal/lower/functions.go').read_text()
needle = 'l.signed[index] = signed{this: this, defaults: defaults, patterns: patterns}'
assert needle in functions
functions = functions.replace(needle, needle + '\n if l.latentReady == nil { l.latentReady = map[int]bool{} }; l.latentReady[index] = true')
# Generic dependency instantiations must also obey body skip policy.
needle = 'func (l *lowering) lowerFunction(index int, declaration *ast.Node, this int) error {'
functions = functions.replace(needle, needle + '\n if len(l.latentBodyDiagnostics(declaration.Body())) > 0 { return &LatentDependencySkipped{Where:l.program.Where(declaration)} }')
overlay('internal/lower/functions.go', functions)
locals_source = (repository / 'internal/lower/locals.go').read_text()
needle = 'local, isLocal := l.locals[symbol]'
assert locals_source.count(needle) == 1
locals_source = locals_source.replace(needle, 'if latentFullEnabled() { l.latentLexicalLocal(identifier, symbol) }\n' + needle)
needle = 'func (l *lowering) declareLocal(name *ast.Node) (int, error) {'
assert locals_source.count(needle) == 1
locals_source = locals_source.replace(needle, needle + '\n l.latentDeclaredName(name)')
needle = 'func (l *lowering) variables(list *ast.Node) ([]ir.Statement, error) {'
assert locals_source.count(needle) == 1
locals_source = locals_source.replace(needle, needle + '\n for _, declaration := range list.AsVariableDeclarationList().Declarations.Nodes { l.latentDeclaredName(declaration.Name()) }')
overlay('internal/lower/locals.go', locals_source)

hook = (territory / 'lower.go.txt').read_text()
if (repository / 'internal/lower/enums.go').exists():
    hook = hook.replace('// LATENT_ENUM_REGISTRATION', """case ast.KindEnumDeclaration:
        if !ast.HasSyntacticModifier(node, ast.ModifierFlagsConst) {
            func() {
                defer func() { _ = recover() }()
                if local, err := l.enumLocal(node); err == nil { l.result.Locals[local].Global = true }
            }()
        }""")
overlay('internal/lower/latent_hook.go', hook)
overlay('stage3/census/latent/tool/main.go', (territory / 'main.go.txt').read_text())
overlay('stage3/census/latent/replay/worker/main.go', (territory / 'replay/main.go.txt').read_text())
(scratch / 'overlay.json').write_text(json.dumps({'Replace': replace}, indent=2) + '\n')
print(scratch / 'overlay.json')
