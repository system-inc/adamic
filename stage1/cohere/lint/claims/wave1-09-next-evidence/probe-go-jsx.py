#!/usr/bin/env python3
"""Use the real Go parser in JSX mode to distinguish rejection from reinterpretation."""
import json
from pathlib import Path
import subprocess
import tempfile

repository = Path(__file__).resolve().parents[5]
root = repository / 'cohere/TypeScript/tsc'
owned = Path(__file__).resolve().parent
with tempfile.TemporaryDirectory(prefix='lint-wave1-09-jsx-') as directory:
    scratch = Path(directory)
    source = (repository / 'stage1/typescript/parser/testdata/oracle.go').read_text()
    assert source.count('core.ScriptKindTS)') == 1
    assert source.count('FileName: "/source.ts"') == 1
    source = source.replace('core.ScriptKindTS)', 'core.ScriptKindTSX)')
    source = source.replace('FileName: "/source.ts"', 'FileName: "/source.tsx"')
    adapter = scratch / 'oracle.go'
    adapter.write_text(source)
    virtual = root / 'adamic_wave109_jsx_probe.go'
    overlay = scratch / 'overlay.json'
    overlay.write_text(json.dumps({'Replace': {str(virtual): str(adapter)}}))
    binary = scratch / 'oracle'
    result = subprocess.run(['go', 'build', '-overlay=' + str(overlay), '-o', str(binary), str(virtual)],
                            cwd=root, text=True, capture_output=True)
    print('upstream JSX parser build exit', result.returncode)
    print(result.stdout + result.stderr, end='')
    assert result.returncode == 0
    for name in ['tailwind-control', 'tailwind-jsx', 'description-control', 'description-jsx', 'google-font-display']:
        fixture = owned / (name + '.ts.txt')
        result = subprocess.run([str(binary), str(fixture), '--whole'], text=True, capture_output=True)
        (owned / (name + '-go.log')).write_text(result.stdout + result.stderr)
        print(name, 'Go JSX parser exit', result.returncode)
        assert result.returncode == 0
        if name in ['tailwind-jsx', 'description-jsx', 'google-font-display']:
            assert ' Jsx' in result.stdout
        if name == 'description-jsx':
            native = (owned / (name + '-native.log')).read_text()
            assert result.stdout != native
            assert 'JsxElement' in result.stdout and 'JsxElement' not in native
            assert 'TypeAssertionExpression' in native and 'TypeAssertionExpression' not in result.stdout
            print('Observed: successful Adamic parse reinterprets JSX as a type assertion and regexp.')
