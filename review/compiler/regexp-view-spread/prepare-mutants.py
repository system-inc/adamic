from pathlib import Path
import json
root = Path.cwd()
evidence = root / 'review/compiler/regexp-view-spread'
source = root / 'internal/lower/lower.go'
text = source.read_text()
block = '\tif err := lowering.checkLibrarySpreads(modules); err != nil {\n\t\treturn nil, err\n\t}\n'
assert text.count(block) == 1
mutant = evidence / 'drop-refusal.go.txt'
mutant.write_text(text.replace(block, ''))
(evidence / 'drop-refusal-overlay.json').write_text(json.dumps({'Replace': {str(source): str(mutant)}}, indent=2) + '\n')
source = root / 'internal/lower/library_spread.go'
text = source.read_text()
block = 'if graph.plainSpreadSources(spread.source, map[int]bool{}) {'
assert text.count(block) == 1
mutant = evidence / 'drop-plain-proof.go.txt'
mutant.write_text(text.replace(block, 'if false && graph.plainSpreadSources(spread.source, map[int]bool{}) {'))
(evidence / 'drop-plain-proof-overlay.json').write_text(json.dumps({'Replace': {str(source): str(mutant)}}, indent=2) + '\n')
