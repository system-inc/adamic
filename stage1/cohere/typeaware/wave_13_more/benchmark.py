"""Alternate full-output runs after builds and tests; report three-round medians."""
import hashlib,json,os,pathlib,statistics,subprocess,sys,time
native,oracle,output=map(pathlib.Path,sys.argv[1:4]);output.mkdir(parents=True,exist_ok=True)
rounds=[];summary={}
for corpus,config,manifest in [('compiler','/workspace/wave13-corpus/src/compiler/tsconfig.json','/workspace/wave-13-compiler.manifest'),('repository','/workspace/adamic/tsconfig.json','/workspace/wave-13-repository.manifest')]:
    values={'native':[],'go':[]}
    for number in range(1,4):
        order=[('native',native),('go',oracle)]
        if number==2:order.reverse()
        expected=None;record={'corpus':corpus,'round':number}
        for mode,binary in order:
            out,err=output/(f'{corpus}-{number}-{mode}.stdout'),output/(f'{corpus}-{number}-{mode}.stderr')
            started=time.perf_counter()
            with out.open('wb') as stdout,err.open('wb') as stderr:
                result=subprocess.run([str(binary),config,manifest],stdout=stdout,stderr=stderr,env=dict(os.environ,ADAMIC_TSGO_TIMING='1'))
            elapsed=time.perf_counter()-started;assert result.returncode==0,(corpus,mode,result.returncode)
            data=out.read_bytes()
            if expected is not None:assert expected==data,corpus+' finding bytes differ'
            expected=data;values[mode].append(elapsed);record[mode]={'seconds':elapsed,'sha256':hashlib.sha256(data).hexdigest(),'bytes':len(data),'timing':err.read_text()}
        rounds.append(record)
    medians={mode:statistics.median(v) for mode,v in values.items()};summary[corpus]=dict(native_median_seconds=medians['native'],go_median_seconds=medians['go'],ratio=medians['native']/medians['go'])
(output/'measurements.json').write_text(json.dumps(dict(rounds=rounds,summary=summary),indent=2)+'\n')
print(json.dumps(summary,indent=2))
