#!/usr/bin/env python3
"""Exercise the inherited serializer against unmodified upstream rule findings."""
import json
from pathlib import Path
import subprocess
import tempfile

repository = Path(__file__).resolve().parents[5]
owned = Path(__file__).resolve().parent
selection = '''package main
import (
 "github.com/system-inc/cohere/internal/lint/rule"
 rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)
type registered struct { subject rule.Rule; options func([]string) any }
func registeredRules() []registered {
 return []registered{
  {rules.NoNonNullAssertedOptionalChain, func([]string) any { return nil }},
  {rules.NoNonNullAssertion, func([]string) any { return nil }},
  {rules.NoThisAlias, func([]string) any { return nil }},
 }
}
'''
with tempfile.TemporaryDirectory(prefix='wave109-contract-') as directory:
    scratch = Path(directory)
    adapter = scratch / 'selection.go'
    adapter.write_text(selection)
    main_source = scratch / 'oracle.go'
    inherited = (repository / 'stage1/cohere/lint/testdata/oracle.go').read_text()
    assert inherited.count('FileName: path,') == 1
    main_source.write_text(inherited.replace('FileName: path,', 'FileName: strings.TrimSuffix(path, ".txt"),'))
    virtual_main = repository / 'cohere/wave109_oracle.go'
    virtual_selection = repository / 'cohere/wave109_selection.go'
    overlay = scratch / 'overlay.json'
    overlay.write_text(json.dumps({'Replace': {
        str(virtual_main): str(main_source),
        str(virtual_selection): str(adapter),
    }}))
    binary = scratch / 'oracle'
    result = subprocess.run(['go', 'build', '-overlay=' + str(overlay), '-o', str(binary),
                             str(virtual_main), str(virtual_selection)],
                            cwd=repository / 'cohere', capture_output=True)
    assert result.returncode == 0, result.stderr
    cases = [('optional-chain', '@typescript-eslint/no-non-null-asserted-optional-chain', 'unexpected suggestion shape'),
             ('non-null', '@typescript-eslint/no-non-null-assertion', 'unexpected suggestion shape'),
             ('this-alias', '@typescript-eslint/no-this-alias', None)]
    for name, rule, failure in cases:
        manifest = scratch / 'manifest'
        manifest.write_text(str(owned / (name + '.ts.txt')) + '\t' + rule + '\n')
        count = subprocess.run([str(binary), '--manifest', str(manifest), '--count'], capture_output=True)
        assert count.returncode == 0 and count.stdout == b'1\n', count
        print(name, 'independent Go findings', count.stdout.decode().strip())
        result = subprocess.run([str(binary), '--manifest', str(manifest)], capture_output=True)
        (owned / (name + '-go-rule.log')).write_bytes(result.stdout + result.stderr)
        print(name, 'inherited formatter exit', result.returncode)
        if failure:
            assert result.returncode != 0 and failure.encode() in result.stderr
            print('Observed:', failure)
        else:
            assert result.returncode == 0
    manifest.write_text(str(owned / 'this-alias.js.txt') + '\t@typescript-eslint/no-this-alias\n')
    result = subprocess.run([str(binary), '--manifest', str(manifest), '--count'], capture_output=True)
    (owned / 'this-alias-js-count.log').write_bytes(result.stdout + result.stderr)
    assert result.returncode == 0 and result.stdout == b'0\n'
    print('same this-alias source with .js filename: independent Go findings 0')
