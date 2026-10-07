"""Compare both frozen populations and retain phase timings and canonical bytes."""
import hashlib, json, os, pathlib, subprocess, sys, time
native, oracle, output = map(pathlib.Path, sys.argv[1:4])
output.mkdir(parents=True, exist_ok=True)
results=[]
for name,config,manifest in [('compiler','/workspace/wave13-corpus/src/compiler/tsconfig.json','/workspace/wave-13-compiler.manifest'),('repository','/workspace/adamic/tsconfig.json','/workspace/wave-13-repository.manifest')]:
    streams=[];item={'corpus':name,'roots':len(pathlib.Path(manifest).read_text().splitlines())}
    for mode,binary in [('go',oracle),('native',native)]:
        out,err=output/(name+'-'+mode+'.stdout'),output/(name+'-'+mode+'.stderr')
        env=dict(os.environ,ADAMIC_TSGO_TIMING='1');started=time.perf_counter()
        with out.open('wb') as stdout,err.open('wb') as stderr:
            result=subprocess.run([str(binary),config,manifest],stdout=stdout,stderr=stderr,env=env)
        assert result.returncode==0,(name,mode,result.returncode,err.read_text())
        item[mode+'_seconds']=time.perf_counter()-started;streams.append(out.read_bytes())
    assert streams[0]==streams[1],name+' finding bytes differ'
    item['bytes']=len(streams[0]);item['findings']=streams[0].splitlines()[-1].decode();item['sha256']=hashlib.sha256(streams[0]).hexdigest()
    results.append(item);print(json.dumps(item),flush=True)
(output/'corpora-results.json').write_text(json.dumps(results,indent=2)+'\n')
