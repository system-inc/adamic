"""A single bounded control; binaries are built before tests, never by tests."""
import json, os, pathlib, subprocess, sys, tempfile
name = sys.argv[1]
here = pathlib.Path(__file__).resolve().parent
binary = os.environ['SCAN_ALL_BINARY']
production = os.environ['SCAN_ALL_PRODUCTION']
source = (here/'fixtures'/(name+'.ts')).read_text()
def call(binary, path):
    return json.loads(subprocess.check_output([binary,str(path)],text=True,timeout=40))
with tempfile.TemporaryDirectory() as directory:
    path = pathlib.Path(directory)/'control.ts'
    path.write_text(source.removeprefix('// @ts-check\n'))
    plain = call(production,path)
    print('plain production '+name+': '+json.dumps(plain), flush=True)
    positive = name.endswith('-stop')
    target = ('bind of a function that observes' if name.startswith('count') else
              'bind of a rest-parameter function' if name.startswith('rest') else
              'a record written into a graph slot')
    assert plain['Completed'] == (not positive), ('production completion',name,plain)
    if positive: assert target in plain['Error'], ('production target',name,plain)
    path.write_text(source)
    r = call(binary,path)
    stops = r['Stops'] or []
    assert any('@ts-check' in s['Words'] and s['Kind']=='scan stop' for s in stops), r
    assert any(target in s['Words'] for s in stops) == positive, ('continuation-target',name,r)
    assert r['Completed'] == (not positive), ('continuation-completion',name,r)
    if positive:
        phase = ('lowering.checkBoundArguments' if name.startswith('count') else
                 'lowering.moduleStatements' if name.startswith('rest') else 'lowering.findCycles')
        assert any(target in s['Words'] and s['Kind']=='first declaration/body/tail stop' and s['Rule']==phase for s in stops), ('pipeline-boundary',name,r)
    assert not r['Boundary'].startswith('panic'),r
print(name+' production and scan-continuation control passed')
