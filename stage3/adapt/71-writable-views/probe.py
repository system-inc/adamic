#!/usr/bin/env python3
"""Add the table's exact eight body probes to a no-output scratch overlay."""
import json
import pathlib
import re
import sys

source, output, tree = [pathlib.Path(p).resolve() for p in sys.argv[1:]]
output.mkdir(parents=True, exist_ok=False)
metadata = json.loads((source / 'overlay.json').read_text())
keys = [key for key in metadata['Replace'] if key.endswith('/internal/load/latent_hook.go')]
assert len(keys) == 1
key = keys[0]
text = pathlib.Path(metadata['Replace'][key]).read_text()
needle = 'func (p *Program) LatentDiagnosticsIn(node *ast.Node) []string {'
assert text.count(needle) == 1
owners = [('checker.ts', 'createTypeChecker'), ('factory/nodeFactory.ts', 'createNodeFactory'),
          ('emitter.ts', 'createPrinter'), ('transformers/es2015.ts', 'transformES2015'),
          ('program.ts', 'createProgram'), ('binder.ts', 'createBinder'),
          ('transformers/classFields.ts', 'transformClassFields'), ('scanner.ts', 'createScanner')]
locations = []
for relative, name in owners:
    file = tree / 'src/compiler' / relative
    lines = file.read_text().splitlines()
    matches = [i + 1 for i, line in enumerate(lines)
               if re.match(r'(?:export )?function ' + name + r'\(', line)
               and not line.rstrip().endswith(';')]
    assert len(matches) == 1, (relative, name, matches)
    locations.append(f'/src/compiler/{relative}:{matches[0]}:1')
# The production loader and nil-IR guards remain those of the original overlay.
# Only the eight named bodies bypass eligibility; raw diagnostics remain visible.
insertion = '\n\tif node != nil && node.Parent != nil && os.Getenv("UNIT71_TABLE_PROBE") == "1" {\n'
insertion += '\t\tswitch p.Where(node.Parent) {\n\t\tcase ' + ', '.join('os.Getenv("UNIT71_TABLE_TREE") + ' + json.dumps(s) for s in locations) + ':\n\t\t\treturn nil\n\t\t}\n\t}\n'
text = text.replace(needle, needle + insertion, 1)
hook = output / 'probe_load_hook.go'
hook.write_text(text)
metadata['Replace'][key] = str(hook)
(output / 'overlay.json').write_text(json.dumps(metadata, indent=2) + '\n')
(output / 'owners.json').write_text(json.dumps(locations, indent=2) + '\n')
