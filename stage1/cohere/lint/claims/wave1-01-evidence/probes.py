"""Reproduce registration and suggestion blockers without changing production files."""
import json
import pathlib
import shutil
import subprocess
import tempfile

repository = pathlib.Path(__file__).resolve().parents[5]
cohere = repository / 'cohere'
scratch = pathlib.Path(tempfile.mkdtemp(prefix='lint-wave1-01-probes-'))
shutil.copytree(repository / 'stage1/cohere/lint/rules/no-debugger', scratch / 'rules/no-debugger')
(scratch / 'rules/no-debugger/rule.ts').rename(scratch / 'rules/no-debugger/rule.a')
result = subprocess.run(['go', 'run', './cmd/lint-registry', '-root', str(scratch)], cwd=repository)
print('extension exit=' + str(result.returncode), flush=True)
assert result.returncode == 1
selection = scratch / 'selection.go'
selection.write_text('''package main
import (
 "github.com/system-inc/cohere/internal/lint/rule"
 rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)
type registeredRule struct { subject rule.Rule; options func([]string) any }
func registeredRules() []registeredRule { return []registeredRule{{rules.ConsistentTypeAssertions, func(fields []string) any { options, err := rules.DecodeConsistentTypeAssertionsOptions([]byte(fields[5])); if err != nil { panic(err) }; return options }}} }
''')
virtual = [cohere / 'adamic_wave1_probe.go', cohere / 'adamic_wave1_selection.go']
overlay = scratch / 'overlay.json'
overlay.write_text(json.dumps({'Replace': {
 str(virtual[0]): str(repository / 'stage1/cohere/lint/testdata/oracle.go'),
 str(virtual[1]): str(selection),
}}))
witness = scratch / 'witness.ts.txt'
witness.write_text('const x = {} as Foo;')
manifest = scratch / 'manifest.txt'
manifest.write_text(str(witness) + '\t@typescript-eslint/consistent-type-assertions\t\t\tfalse\t'
                    + json.dumps({'assertionStyle': 'as', 'objectLiteralTypeAssertions': 'never'}) + '\n')
result = subprocess.run(['go', 'build', '-overlay=' + str(overlay), '-o', str(scratch / 'oracle'),
                         *[str(path) for path in virtual]], cwd=cohere)
print('build exit=' + str(result.returncode), flush=True)
assert result.returncode == 0
if result.returncode == 0:
 result = subprocess.run([str(scratch / 'oracle'), '--manifest', str(manifest)], capture_output=True)
 print(result.stdout.decode(), end='')
 print(result.stderr.decode(), end='')
 assert result.returncode == 2
 assert b'panic: unexpected suggestion shape' in result.stderr
 print('suggestion exit=' + str(result.returncode), flush=True)
