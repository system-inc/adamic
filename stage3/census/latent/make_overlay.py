"""Build a scratch-only lower overlay. Production Lower is disabled in this binary."""
import json
import pathlib
import sys

repository = pathlib.Path(sys.argv[1]).resolve()
scratch = pathlib.Path(sys.argv[2]).resolve()
territory = pathlib.Path(__file__).resolve().parent
scratch.mkdir(parents=True, exist_ok=True)
replace = {}

def overlay(name, text):
    output = scratch / pathlib.Path(name).name
    output.write_text(text)
    replace[str(repository / name)] = str(output)

original = (repository / 'internal/lower/lower.go').read_text()
start = original.index('\tfiles := program.Files()')
end = original.index('\n}\n\ntype lowering', start)
original = original[:start] + '\treturn nil, fmt.Errorf("latent census: measurement only; no IR output")' + original[end:]
original = original.replace('\n\t"path/filepath"', '')
overlay('internal/lower/lower.go', original)
refuse = (repository / 'internal/lower/refusals.go').read_text()
start = refuse.index('func (l *lowering) refuse(')
end = refuse.index('\n// called ', start)
body = refuse[start:end].replace('func (l *lowering) refuse(', 'func (l *lowering) latentRefuse(')
body = body.replace('\t\tif found != nil {\n\t\t\treturn true\n\t\t}\n', '')
body = body.replace('return true', 'latentRecord(found)\n\t\t\tfound = nil\n\t\t\tnode.ForEachChild(visit)\n\t\t\treturn false')
# Directives are module metadata rather than visitor nodes; visit every directive.
body = body.replace('if len(module.CommentDirectives) > 0 {\n\t\tdirective := module.CommentDirectives[0]', 'for _, directive := range module.CommentDirectives {')
body = body.replace('return &Refused{Where:', 'latentRecord(&Refused{Where:', 1)
body = body.replace('Fix: "remove it and fix the type error"}', 'Fix: "remove it and fix the type error"})', 1)
overlay('internal/lower/refusals.go', refuse[:end] + '\n' + body + refuse[end:])
overlay('internal/lower/latent_hook.go', (territory / 'lower.go.txt').read_text())
overlay('stage3/census/latent/tool/main.go', (territory / 'main.go.txt').read_text())
(scratch / 'overlay.json').write_text(json.dumps({'Replace': replace}, indent=2) + '\n')
print(scratch / 'overlay.json')
