#!/usr/bin/env python3
"""Check the narrow snapshot compatibility adapter, including a guard mutant."""
from pathlib import Path
import subprocess
import tempfile

run = Path(__file__).resolve().parents[1]
scratch = Path(tempfile.mkdtemp(prefix='meter-statecopy-'))
source = scratch / 'internal/lower'
source.mkdir(parents=True)
generator = scratch / 'statecopy'
def execute(name, command, expected=0):
    log = run / 'validation' / (name + '.log')
    with log.open('w') as output:
        result = subprocess.run(command, stdout=output, stderr=output, timeout=180)
    text = log.read_text()
    assert result.returncode == expected, (name, result.returncode, text)
    return text
generator_source = scratch / 'statecopy.go'
generator_source.write_text((run / 'validation/statecopy-run.go.txt').read_text())
execute('statecopy-adapter-build', ['go', 'build', '-o', str(generator), str(generator_source)])
input_file = source / 'lower.go'
input_file.write_text('package lower\ntype lowering struct { impliedTargetQuery func(*ast.Node) *checker.Type }\n')
output_file = scratch / 'copier.go'
execute('statecopy-adapter-generate', [str(generator), str(scratch), str(input_file), str(output_file)])
text = output_file.read_text().replace('package lower', 'package main')
(scratch / 'main.go').write_text('''package main
import (
 "github.com/microsoft/TypeScript/tsc/shim/ast"
 "github.com/microsoft/TypeScript/tsc/shim/checker"
)
type lowering struct { impliedTargetQuery func(*ast.Node) *checker.Type }
func main() {
 state := lowering{}
 if latentCopy_lowering(state,map[any]any{}).impliedTargetQuery != nil { panic("nil query changed") }
 state.impliedTargetQuery = func(*ast.Node) *checker.Type {return nil}
 defer func(){ if recover()==nil {panic("populated query guard did not fail")}; println("nil preserved; populated query refused") }()
 latentCopy_lowering(state,map[any]any{})
}
''')
output_file.write_text(text)
observed = execute('statecopy-adapter-witness', ['go', 'run', str(scratch / 'main.go'), str(output_file)])
assert 'nil preserved; populated query refused' in observed
old = 'if value.impliedTargetQuery != nil {\n\t\tpanic("latent state copy: populated impliedTargetQuery is unsupported")\n\t}'
assert text.count(old) == 1
output_file.write_text(text.replace(old, ''))
observed = execute('statecopy-adapter-mutant', ['go', 'run', str(scratch / 'main.go'), str(output_file)], 1)
assert 'populated query guard did not fail' in observed
input_file.write_text('package lower\ntype lowering struct { unrelated func() }\n')
unknown = scratch / 'unknown.go'
observed = execute('statecopy-adapter-unknown', [str(generator), str(scratch), str(input_file), str(unknown)], 2)
assert 'unsupported field type func()' in observed and not unknown.exists()
print('PASS: nil preserved, populated hook rejected, unrelated function shape rejected; permissive-hook mutant caught')
