"""Use an independent source to verify binder extraction and generic hatch scopes."""
import json
from pathlib import Path
import subprocess
import sys
import tempfile

with tempfile.TemporaryDirectory(prefix='step16-parameters-') as scratch:
    root = Path(scratch)
    sources = root / 'source'
    sources.mkdir()
    (sources / 'probe.a').write_text('''interface SourceFile { readonly text: string; }
function generic<SourceFile>(value: unknown): SourceFile { return value as SourceFile; }
function other<T>(value: unknown): T { return value as T; }
function concrete(value: unknown): SourceFile { return value as SourceFile; }
''')
    script = Path(__file__).with_name('parameters.cjs').read_text()
    if '--mutant' in sys.argv:
        script = script.replace('ts.ScriptTarget.Latest, true)', 'ts.ScriptTarget.Latest, false)')
    runner = root / 'parameters.cjs'
    runner.write_text(script)
    data = json.loads(subprocess.check_output(['node', str(runner), sys.argv[1], str(sources)]))
    assert data['names'] == ['SourceFile', 'T'], data['names']
    assert len(data['scopes']['probe.a']) == 2, data['scopes']
    assert len(data['genericUses']['probe.a']) == 2, data['genericUses']
    assert data['genericUses']['probe.a'][0]['start'][0] == 2, data['genericUses']
    assert data['genericUses']['probe.a'][1]['start'][0] == 3, data['genericUses']
    print('PASS: concrete same-name type excluded; both generic casts retain enclosing binders')
