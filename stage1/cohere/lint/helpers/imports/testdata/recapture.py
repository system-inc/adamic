"""Run the same Go captures as TestImportsAgreementAndMutants before recording provenance."""
import gzip, hashlib, json, pathlib, subprocess, sys
HERE = pathlib.Path(__file__).resolve().parent
ROOT = HERE.parents[5]
sys.path.insert(0, str(ROOT / 'stage1/cohere/lint/helpers/testdata'))
from pin import capture_pin
PIN = capture_pin(ROOT)
COHERE = ROOT / 'cohere'
OUT = pathlib.Path(sys.argv[1]).resolve()
OUT.mkdir(parents=True, exist_ok=True)
for mode in ['bindings', 'filename', 'imported', 'call', 'segment']:
    directory = HERE / mode
    dest = OUT / mode
    dest.mkdir(exist_ok=True)
    virtual = COHERE / 'adamic_imports_oracle.go'
    replacements = {str(virtual): str(directory / 'oracle.go')}
    if mode == 'bindings':
        replacements[str(COHERE / 'internal/lint/ecmascript/react/adamic_imports.go')] = str(directory / 'react_export.go')
    if mode == 'filename':
        replacements[str(COHERE / 'TypeScript/tsc/internal/ast/adamic_imports.go')] = str(directory / 'filename_ast.go')
        replacements[str(COHERE / 'TypeScript-shim/ast/adamic_imports.go')] = str(directory / 'filename_shim.go')
    overlay = dest / 'overlay.json'
    overlay.write_text(json.dumps({'Replace': replacements}))
    binary = dest / 'oracle'
    subprocess.run(['go', 'build', '-overlay='+str(overlay), '-o', str(binary), str(virtual)], cwd=COHERE, check=True)
    if mode in ['bindings', 'filename', 'imported']:
        args = [str(ROOT), str(dest)]
        if mode == 'bindings': args.append('github.com/system-inc/cohere/internal/lint/ecmascript/imports.BindingsOf')
        if mode == 'imported': args.append('imported')
        subprocess.run([str(binary)] + args, check=True)
    elif mode == 'segment':
        with (dest / 'want.txt').open('wb') as output:
            subprocess.run([str(binary), str(directory / 'sources.jsonl.gz'), str(dest / 'cases.json')], stdout=output, check=True)
    else:
        with gzip.open(directory / 'sources.jsonl.gz', 'rt') as source:
            records = [json.loads(line) for line in source if line.strip()]
        inputs = records
        config = dest / 'sources.json'
        config.write_text(json.dumps(inputs))
        with (dest / 'raw.json').open('wb') as output:
            subprocess.run([str(binary), str(config)], stdout=output, check=True)
        corpus = json.loads((dest / 'raw.json').read_text())
        want = corpus.pop('Want')
        (dest / 'want.txt').write_text(''.join(line+'\n' for line in want.splitlines() if line.startswith('call:')))
        (dest / 'cases.json').write_text(json.dumps(corpus))
    (dest / 'capture-pin.json').write_text(json.dumps({'pin': PIN})+'\n')
    print(mode, len((dest / 'want.txt').read_bytes()), 'answer bytes')
assert capture_pin(ROOT) == PIN
metadata = json.loads((HERE / 'provenance.json').read_text())
metadata['goPin'] = PIN
metadata['captureAnswers'] = {mode: {'lines': (OUT / mode / 'want.txt').read_bytes().count(b'\n'), 'sha256': hashlib.sha256((OUT / mode / 'want.txt').read_bytes()).hexdigest()} for mode in ['bindings', 'filename', 'imported', 'call', 'segment']}
(HERE / 'provenance.json').write_text(json.dumps(metadata, indent=2)+'\n')
