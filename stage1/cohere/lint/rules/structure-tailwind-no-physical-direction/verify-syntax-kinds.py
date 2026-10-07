#!/usr/bin/env python3
"""Compare every owned numeric listener declaration with the external parser."""
import argparse
import json
from pathlib import Path
import re
import shutil
import subprocess

parser = argparse.ArgumentParser()
parser.add_argument('--scratch', type=Path, required=True)
args = parser.parse_args()
owned = Path(__file__).resolve().parent
repository = owned.parents[4]
scratch = args.scratch.resolve()
scratch.mkdir(parents=True, exist_ok=True)
slugs = ['next-no-assign-module-variable', 'typescript-default-param-last', 'structure-tailwind-no-physical-direction', 'typescript-no-unnecessary-type-constraint', 'typescript-prefer-as-const', 'typescript-prefer-enum-initializers', 'nexus-import-require-node-namespace', 'structure-network-no-invalidate-cache-literal-key', 'structure-network-no-string-literal-query', 'no-multi-str', 'no-nonoctal-decimal-escape', 'no-octal']
def run(command, name):
    result = subprocess.run(list(map(str, command)), cwd=repository, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    (scratch / (name + '.log')).write_bytes(result.stdout)
    (scratch / (name + '.stderr')).write_bytes(result.stderr)
    if result.returncode or result.stderr:
        raise RuntimeError(name + ': execution failed')
    return result.stdout

oracle = scratch / 'kinds.go'
oracle.write_text('''package main
import ("encoding/json"; "os"; "strings"; "github.com/microsoft/TypeScript/tsc/shim/ast")
func main() {
 values := map[string]int{}
 for kind := ast.Kind(0); kind < ast.KindCount; kind++ { values[strings.TrimPrefix(kind.String(), "Kind")] = int(kind) }
 if err := json.NewEncoder(os.Stdout).Encode(values); err != nil { panic(err) }
}
''')
kinds = json.loads(run(['go', 'run', oracle], 'Go-kind-map'))
expected = b''
for slug in slugs:
    descriptor = json.loads((owned.parent / slug / 'rule.json').read_text())
    expected += (slug + ' ' + ','.join(str(kinds[name]) for name in descriptor['kinds']) + '\n').encode()
(scratch / 'expected.log').write_bytes(expected)
virtual = repository / 'wave15_syntax_kind_build.go'
overlay = scratch / 'builder.json'
overlay.write_text(json.dumps({'Replace': {str(virtual): str(owned / 'standalone-build.go.txt')}}))
entry = owned / 'verify-syntax-kinds.a'
binary = scratch / 'declarations'
run(['go', 'run', '-overlay=' + str(overlay), virtual, entry, binary], 'build')
runner = repository / 'oracle/node.mjs'
for side, command in [('Node', ['node', '--disable-warning=ExperimentalWarning', runner, entry]), ('JavaScript', ['node', '--disable-warning=ExperimentalWarning', runner, str(binary) + '.mjs']), ('native', [binary])]:
    if run(command, side) != expected:
        raise RuntimeError(side + ': listener kinds differ from Go parser')
print('All 12 exported numeric listener declarations match the pinned Go parser on source Node, emitted JavaScript and sanitized native.', flush=True)

# One declaration is deliberately wrong; current rule dispatch ignores these
# exports, so ordinary rule output would miss it. The new external check catches it.
tree = scratch / 'mutant'
for source in (repository / 'stage1').rglob('*'):
    if source.suffix in ['.ts', '.a']:
        target = tree / source.relative_to(repository)
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(source, target)
target = tree / 'stage1/cohere/lint/rules/no-octal/rule.a'
source = target.read_text()
anchor = 'export const syntaxKinds: readonly number[] = [8];'
if source.count(anchor) != 1:
    raise RuntimeError('nonunique numeric declaration mutant')
target.write_text(source.replace(anchor, 'export const syntaxKinds: readonly number[] = [9];'))
candidate = tree / entry.relative_to(repository)
mutant = scratch / 'mutated-declarations'
run(['go', 'run', '-overlay=' + str(overlay), virtual, candidate, mutant], 'mutant-build')
for side, command in [('Node', ['node', '--disable-warning=ExperimentalWarning', runner, candidate]), ('JavaScript', ['node', '--disable-warning=ExperimentalWarning', runner, str(mutant) + '.mjs']), ('native', [mutant])]:
    if run(command, 'mutant-' + side) == expected:
        raise RuntimeError(side + ': numeric listener mutant survived')
    print('NumericLiteral 8 -> 9 mutant caught on ' + side + ' only by declaration output comparison; compile and exit successful.', flush=True)
