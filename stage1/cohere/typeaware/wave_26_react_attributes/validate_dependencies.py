"""Owned Node-source, stale-handle and bridge-mutant checks."""
import pathlib, subprocess, os, json, time
ROOT=pathlib.Path(__file__).resolve().parents[4];OWN=pathlib.Path(__file__).resolve().parent
S=pathlib.Path('/workspace/wave-26-attributes');records=[]
def run(name,command,expected=0,env=None):
    started=time.monotonic()
    with (S/(name+'.stdout')).open('wb') as out,(S/(name+'.stderr')).open('wb') as err:
        result=subprocess.run([str(x) for x in command],cwd=ROOT,stdout=out,stderr=err,env=env)
    records.append(dict(name=name,command=[str(x) for x in command],exit=result.returncode,seconds=time.monotonic()-started))
    (S/'dependency-runs.json').write_text(json.dumps(records,indent=2)+'\n')
    assert result.returncode==expected,(name,result.returncode,(S/(name+'.stderr')).read_text())
    return (S/(name+'.stdout')).read_bytes()
config=S/'tsconfig.json';manifest=S/'controls.manifest'
channels=[S/'node-input.fifo',S/'node-output.fifo']
for file in channels:
    if not file.exists():os.mkfifo(file)
fds=[os.open(file,os.O_RDWR) for file in channels]
with (S/'checker-server.stderr').open('wb') as err:
    server=subprocess.Popen([str(S/'checker-server')],stdin=fds[0],stdout=fds[1],stderr=err,cwd=ROOT)
    try:
        env=dict(os.environ,ADAMIC_REACT_INPUT=str(channels[0]),ADAMIC_REACT_OUTPUT=str(channels[1]))
        actual=run('source-node',['node','--disable-warning=ExperimentalWarning',OWN/'testdata/node_oracle.mjs',OWN/'suite.a',config,manifest],env=env)
        assert actual==(S/'controls-go.stdout').read_bytes()
        assert (S/'source-node.stderr').read_bytes()==b''
        assert server.wait(timeout=10)==0
    finally:
        if server.poll() is None:server.terminate();server.wait(timeout=10)
        for fd in fds:os.close(fd)
print('source Node: complete default control bytes match production Go',flush=True)
run('js-emission',['/workspace/wave-26-rawhir/adamic','js',OWN/'suite.a'],1)
assert b'unlinked typescript-go library call' in (S/'js-emission.stderr').read_bytes()
print('shared emitted-JavaScript checker-link gap: explicit refusal preserved',flush=True)
probe=OWN/'testdata/released.a'
run('released-build',['/workspace/wave-26-rawhir/adamic','build',probe,'-o',S/'released','--tsgo',S/'checker.a'])
file=pathlib.Path(manifest.read_text().splitlines()[0])
outputs={}
for question in ['jsx-structure','symbol-locations']:
    outputs[question]=run(question+'-released',[S/'released',config,file,len(file.read_bytes()),question],70)
    assert (S/(question+'-released.stderr')).read_bytes()==b'adamic: panic: invalid or released checker handle\n'
registry=ROOT/'bridge/tsgo/archive/main.go';source=registry.read_text();needle='delete(programs.live, uint64(handle))';assert source.count(needle)==1
(S/'retained.go').write_text(source.replace(needle,'// Mutant retains the released handle.'))
overlay=S/'retained.json';overlay.write_text(json.dumps({'Replace':{str(registry):str(S/'retained.go')}}))
run('retained-archive',['go','build','-overlay',overlay,'-buildmode=c-archive','-o',S/'retained.a','./bridge/tsgo/archive'])
run('retained-build',['/workspace/wave-26-rawhir/adamic','build',probe,'-o',S/'retained','--tsgo',S/'retained.a'])
for question,output in outputs.items():
    assert run(question+'-retained',[S/'retained',config,file,len(file.read_bytes()),question])==output+output
    assert (S/(question+'-retained.stderr')).read_bytes()==b''
print('both new questions: copied data survives release, stale queries panic 70, retaining-handle mutant exits 0 and is caught',flush=True)
