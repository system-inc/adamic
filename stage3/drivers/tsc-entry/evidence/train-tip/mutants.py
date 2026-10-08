"""Known-answer fixtures and corrupt-receipt mutants for the pending report."""
import ast,importlib.util,json,shutil,sys,tempfile
from pathlib import Path
HERE=Path(__file__).resolve().parent
sys.dont_write_bytecode=True
spec=importlib.util.spec_from_file_location('train_verify',HERE/'verify.py');module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
# A known checker diagnostic outside compiler must retain its owner path and UTF-16 coordinates.
fixture='adamic: /tmp/two-file/src/tsc/entry.ts:4:2: error TS18048: value is possibly undefined.\n'
assert module.diagnostic(fixture)==('src/tsc/entry.ts',4,2,'error TS18048: value is possibly undefined.')
functions=ast.parse((HERE/'collect.py').read_text()).body
units=next(n for n in functions if isinstance(n,ast.FunctionDef) and n.name=='units')
scope={};exec(compile(ast.Module(body=[units],type_ignores=[]),str(HERE/'collect.py'),'exec'),scope)
assert len(scope['units']('a\U0001f642b'))==4
try:
    assert len('a\U0001f642b')==4
except AssertionError:print('UTF-16 scalar-count mutant: caught by known-answer fixture')
else:raise AssertionError('UTF-16 mutant survived')
module.verify(HERE,Path('/tmp/tsc-train-build-sparse/adapted'))
with tempfile.TemporaryDirectory(prefix='tsc-train-mutants-') as scratch:
    for name in ['pending-status','stop-population','split-byte','node-byte','owner','outside-file','source-hash']:
        copy=Path(scratch)/name;shutil.copytree(HERE,copy)
        if name=='pending-status':
            file=copy/'provenance.json';data=json.loads(file.read_text());data['status']='pass';file.write_text(json.dumps(data))
        elif name=='stop-population':
            file=copy/'stops.json';data=json.loads(file.read_text());file.write_text(json.dumps(data[:-1]))
        elif name=='split-byte':
            file=copy/'logs/01-split-1.stderr';file.write_bytes(file.read_bytes()+b'changed\n')
        elif name=='node-byte':(copy/'logs/16-symbol-array-node.stdout').write_text('changed\n')
        elif name=='owner':
            file=copy/'stops.json';data=json.loads(file.read_text());data[0]['owner']['candidate_branch']='unknown';file.write_text(json.dumps(data))
        elif name=='outside-file':
            file=copy/'closure.json';data=json.loads(file.read_text());data['files'].remove('src/tsc/_namespaces/ts.ts');file.write_text(json.dumps(data))
        else:
            file=copy/'source-hashes.json';data=json.loads(file.read_text());data['src/tsc/tsc.ts']='0'*64;file.write_text(json.dumps(data))
        try:module.verify(copy,Path('/tmp/tsc-train-build-sparse/adapted'))
        except AssertionError as error:print(f'{name}: caught ({error})')
        else:raise AssertionError(f'{name} survived')
print('Known-answer fixtures and eight mutants caught; no rehearsal pass claimed')
