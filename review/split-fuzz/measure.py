#!/usr/bin/env python3
"""Serial per-unit cold runtime timing with Python time.monotonic().
Usage: python3 review/split-fuzz/measure.py before|after OUTPUT_DIRECTORY
Source cloud/setup.sh's environment first. Go compilation is shared setup; every
unit gets a fresh process and native runtime cache. The before binary uses base
test sources through an overlay, leaving the checkout unchanged.
"""
import os, re, json, subprocess, time, pathlib, sys, tempfile
root=pathlib.Path(__file__).resolve().parents[4]
out=pathlib.Path(sys.argv[2]).resolve()
out.mkdir(parents=True, exist_ok=True)
phase=sys.argv[1]
binary=out / (phase+'.test')
build_start=time.monotonic()
build=['go','test','-c','-o',str(binary)]
if phase == 'before':
    base='54cbc125422d4e1d64c1ffe782445b2cbc2bc5b8'
    replacements={}
    for p in (root/'internal/fuzz').glob('*test.go'):
        relative=p.relative_to(root).as_posix()
        original=subprocess.check_output(['git','show',base+':'+relative],cwd=root)
        saved=out/('baseline-'+p.name)
        saved.write_bytes(original)
        replacements[str(p)]=str(saved)
    overlay=out/'baseline-overlay.json'
    overlay.write_text(json.dumps({'Replace':replacements},indent=2)+'\n')
    build.append('-overlay='+str(overlay))
build.append('./internal/fuzz')
subprocess.run(build,cwd=root,check=True)
print('binary build seconds',time.monotonic()-build_start,flush=True)
if phase=='before':
    names=re.findall(r'^func (Test\w+)\(', '\n'.join(p.read_text() for p in (root/'internal/fuzz').glob('*test.go')),re.M)
else:
    names=[]
    for p in (root/'internal/fuzz').glob('*test.go'):
        for block in re.split(r'(?=^func Test)',p.read_text(),flags=re.M)[1:]:
            parent=re.match(r'func (Test\w+)',block)[1]
            leaves=re.findall(r'\{"(seeds-[0-9]+-[0-9]+|vocabulary)",',block)
            if parent == 'TestReduceKeepsTheSignature': leaves=['signature','reduction']
            if parent == 'TestFuzzerSharesRuntimeLibrary': leaves=['library','program']
            names.extend([parent+'/'+leaf for leaf in leaves] if leaves else [parent])
rows=[]
for name in names:
    cache=tempfile.mkdtemp(prefix=phase+'-cache-',dir=out)
    env=dict(os.environ,XDG_CACHE_HOME=cache,GOMAXPROCS='4',ADAMIC_GATE_UNCACHED='1')
    # Keep the prepared Go action cache explicit; XDG changes must not accidentally rebuild Go.
    env['GOCACHE']=subprocess.check_output(['go','env','GOCACHE']).decode().strip()
    log=out/(phase+'-'+name.replace('/','--')+'.log')
    pattern='/'.join('^'+re.escape(part)+'$' for part in name.split('/'))
    start=time.monotonic()
    with log.open('w') as f:
        run=subprocess.run([str(binary),'-test.v','-test.count=1','-test.parallel=4','-test.timeout=10m','-test.run='+pattern],cwd=root/'internal/fuzz',env=env,stdout=f,stderr=subprocess.STDOUT)
    runs=re.findall(r'^=== RUN   (.+)$',log.read_text(),re.M)
    if '/' in name:
        assert runs == [name.split('/')[0],name], (name,runs)
    row=dict(test=name,wall=time.monotonic()-start,exit=run.returncode)
    rows.append(row)
    (out/(phase+'.json')).write_text(json.dumps(rows,indent=2)+'\n')
    print(row,flush=True)
