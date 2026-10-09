import json, os, pathlib, re, subprocess, tempfile, time, difflib
root = pathlib.Path.cwd()
evidence = root / 'review/compiler/bridge-products'
source = root / 'bridge/tsgo/products_test.go'
text = source.read_text()
pairs = re.findall(r'func TestProduct_(\w+)\(t \*testing.T\) \{\s*t.Parallel\(\)\s*bridgeProduct\(t, bridgeRepository\(t\), "([^"]+)"\)', (root/'bridge/tsgo/product_units_test.go').read_text())
assert len(pairs) == 19
scratch = pathlib.Path('/workspace/bridge-products-overlays')
scratch.mkdir(exist_ok=True)
needle = '\tfor file := range files {'
assert text.count(needle) == 1
cache = '/workspace/bridge-products-measured-cache'
go_cache = subprocess.check_output(['go', 'env', 'GOCACHE'], text=True).strip()
base = dict(os.environ, GOMAXPROCS='4', ADAMIC_GATE_UNCACHED='1', ADAMIC_BUILD_CACHE_DIR=cache, GOCACHE=go_cache)
rows=[]
def run(label, name, overlay=None, cold=False, expected=0):
    command=['go','test','./bridge/tsgo','-run','^'+name+'$','-count=1','-json','-timeout','90s']
    if overlay: command[2:2]=['-overlay',str(overlay)]
    with tempfile.TemporaryDirectory(prefix='bridge-products-cold-') as directory:
        env=dict(base, ADAMIC_BUILD_LOG=str(evidence/(label+'.builds.txt')))
        if cold: env['XDG_CACHE_HOME']=directory
        started=time.monotonic()
        with (evidence/(label+'.jsonl')).open('w') as log:
            result=subprocess.run(command,env=env,stdout=log,stderr=subprocess.STDOUT)
        row=dict(label=label,test=name,command=command,exit=result.returncode,expected_exit=expected,wall_seconds=round(time.monotonic()-started,3),cold_runtime_cache=cold)
        rows.append(row)
        (evidence/'results.json').write_text(json.dumps(rows,indent=2)+'\n')
        print(json.dumps(row),flush=True)
        assert result.returncode==expected, label
        assert row['wall_seconds']<60, label
for suffix, product in pairs:
    run('input-'+suffix, 'TestBridgeProductInput_'+suffix)
    mutant=text.replace(needle,'\tif name == '+json.dumps(product)+' { delete(files, "bridge/tsgo/products_test.go") }\n'+needle)
    replacement=scratch/(suffix+'.go')
    replacement.write_text(mutant)
    overlay=scratch/(suffix+'.json')
    overlay.write_text(json.dumps({'Replace':{str(source):str(replacement)}}))
    (evidence/(product+'.patch')).write_text(''.join(difflib.unified_diff(text.splitlines(True),mutant.splitlines(True),fromfile='a/bridge/tsgo/products_test.go',tofile='b/bridge/tsgo/products_test.go',n=0)))
    run('mutant-'+suffix, 'TestBridgeProductInput_'+suffix,overlay=overlay,expected=1)
    log=(evidence/('mutant-'+suffix+'.jsonl')).read_text()
    assert 'served stale product' in log, suffix
for suffix, product in pairs:
    run('product-'+suffix, 'TestProduct_'+suffix, cold=True)
for name in ['TestBridgeProductUnitsCoverEveryProduct','TestBridgeUnitsCoverEveryPiece','TestBridgeProductCacheIsVerified','TestBridgeABI','TestBridgeOracleSample']:
    run(name,name)
run('two-fresh-leaves','(?:TestBridgeABI|TestBridgeOracleSample)')
