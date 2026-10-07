import hashlib,json,os,pathlib,re,statistics,subprocess,time
root=pathlib.Path('/workspace/wave20-validation/complete')
results={}
for corpus,config in [('repository','/workspace/adamic/tsconfig.json'),('compiler','/workspace/wave20-typescript/src/compiler/tsconfig.json')]:
    runs={'native':[],'go':[]}
    for round in range(3):
        hashes=[]
        for mode in (['native','go'] if round%2==0 else ['go','native']):
            binary=root/('wave20' if mode=='native' else 'wave20-oracle')
            stem=root/f'bench-{corpus}-{round+1}-{mode}'
            with open(str(stem)+'.stdout','wb') as out,open(str(stem)+'.stderr','wb') as err:
                started=time.perf_counter_ns()
                subprocess.run([str(binary),config,str(root/f'{corpus}.manifest')],stdout=out,stderr=err,env=dict(os.environ,ADAMIC_TSGO_TIMING='1'),check=True)
                process_ns=time.perf_counter_ns()-started
            data=pathlib.Path(str(stem)+'.stdout').read_bytes()
            digest=hashlib.sha256(data).hexdigest();hashes.append(digest)
            phases={key:int(value) for key,value in re.findall(r'(\w+)=(\d+)',pathlib.Path(str(stem)+'.stderr').read_text())}
            runs[mode].append(dict(process_ns=process_ns,phases=phases,sha256=digest,bytes=len(data)))
        assert hashes[0]==hashes[1],f'{corpus} timing byte disagreement'
    results[corpus]=dict(runs=runs,medians={mode:dict(process_seconds=statistics.median(row['process_ns'] for row in rows)/1e9,load_seconds=statistics.median(row['phases']['load_ns'] for row in rows)/1e9,run_seconds=statistics.median(row['phases']['run_ns'] for row in rows)/1e9) for mode,rows in runs.items()})
(root/'bench.json').write_text(json.dumps(results,indent=2)+'\n')
for corpus,value in results.items(): print(corpus,json.dumps(value['medians']))
