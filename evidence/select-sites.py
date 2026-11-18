"""Select candidate source units in a supplemental overlay; leave the census builds intact."""
import json
import sys
from pathlib import Path

source, original, destination = map(lambda name: Path(name).resolve(), sys.argv[1:])
destination.mkdir(parents=True, exist_ok=True)
overlay = json.loads(original.read_text())
replacements = overlay['Replace']

def modify(filename, changes):
    path = str(source/'internal/lower'/filename)
    content = Path(replacements[path]).read_text()
    for old, new in changes:
        assert content.count(old) == 1, (filename, old)
        content = content.replace(old, new)
    output = destination / filename
    output.write_text(content)
    replacements[path] = str(output)

for file, loop in [
    ('latent_hook.go', 'for _, node := range file.Statements.Nodes {'),
    ('latent_units.go', 'node := candidate.node'),
]:
    if file == 'latent_hook.go':
        outer = 'for _, file := range program.Files() {\n\t\tlatentFindings = nil'
        outer = (outer, outer.replace('\n\t\tlatentFindings', '\n if !censusCandidateFile(program.FileName(file)) { continue }\n\t\tlatentFindings'))
    else:
        outer = ('if !latentFileSelected(program.FileName(file)) {', 'if !latentFileSelected(program.FileName(file)) || !censusCandidateFile(program.FileName(file)) {')
    after = '\n\t\t\t' + ('latentFindingOwner, latentFindingPhase = program.Where(node), "lowering"' if file == 'latent_hook.go' else 'if selection != nil && node != selection.node {')
    modify(file, [outer, (loop + after, loop + '\n if !censusCandidateUnit(program.FileName(file), file, node) { continue }' + after)])

helper = destination/'census_candidate_select.go'
helper.write_text(r'''package lower

import (
    "encoding/json"
    "os"

    "github.com/microsoft/TypeScript/tsc/shim/ast"
    "github.com/microsoft/TypeScript/tsc/shim/scanner"
)

var censusCandidatePaths = func() map[string][]int {
    path := os.Getenv("CENSUS_CANDIDATE_MANIFEST")
    if path == "" {
        return nil
    }
    raw, err := os.ReadFile(path)
    if err != nil {
        panic(err)
    }
    var result map[string][]int
    if err := json.Unmarshal(raw, &result); err != nil {
        panic(err)
    }
    return result
}()

func censusCandidateFile(path string) bool {
    return censusCandidatePaths == nil || len(censusCandidatePaths[path]) > 0
}

func censusCandidateUnit(path string, file *ast.SourceFile, node *ast.Node) bool {
    if censusCandidatePaths == nil {
        return true
    }
    first, _ := scanner.GetLineAndCharacterOfPosition(file, node.Pos())
    last, _ := scanner.GetLineAndCharacterOfPosition(file, node.End())
    for _, line := range censusCandidatePaths[path] {
        if line >= first + 1 && line <= last + 1 {
            return true
        }
    }
    return false
}
''')
replacements[str(source/'internal/lower/census_candidate_select.go')]=str(helper)
(destination/'overlay.json').write_text(json.dumps(overlay))
