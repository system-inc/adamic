import subprocess,pathlib,time,json,statistics,hashlib,os,argparse
parser=argparse.ArgumentParser()
for option in ['artifacts','compiler-config','repository-config','compiler-manifest','repository-manifest','output']:
 parser.add_argument('--'+option, required=True)
args=parser.parse_args()
root=pathlib.Path(args.artifacts); records=[]
for corpus,config in [('compiler',args.compiler_config),('repository',args.repository_config)]:
 manifest=getattr(args,corpus+'_manifest')
 expected=None
 for round in range(3):
  for variant in (['go','native'] if round%2==0 else ['native','go']):
   binary=root/('wave08-oracle' if variant=='go' else 'wave08')
   start=time.perf_counter_ns()
   with open(root/f'timing-{corpus}-{round}-{variant}.stdout','wb') as out,open(root/f'timing-{corpus}-{round}-{variant}.stderr','wb') as err:
    p=subprocess.run([str(binary),config,manifest],stdout=out,stderr=err,env=dict(os.environ,ADAMIC_TSGO_TIMING='1'))
   elapsed=time.perf_counter_ns()-start
   assert p.returncode==0
   actual=(root/f'timing-{corpus}-{round}-{variant}.stdout').read_bytes()
   if expected is None:expected=actual
   assert actual==expected
   records.append(dict(corpus=corpus,round=round,variant=variant,elapsed_ns=elapsed,bytes=len(actual),sha256=hashlib.sha256(actual).hexdigest(),stderr=(root/f'timing-{corpus}-{round}-{variant}.stderr').read_text()))
   print(corpus,round,variant,elapsed/1e9,flush=True)
summary={}
for corpus in ['compiler','repository']:
 med={v:statistics.median(r['elapsed_ns'] for r in records if r['corpus']==corpus and r['variant']==v)/1e9 for v in ['go','native']}
 summary[corpus]=dict(median_seconds=med,native_over_go=med['native']/med['go'])
pathlib.Path(args.output).write_text(json.dumps(dict(records=records,summary=summary),indent=2)+'\n')
print(json.dumps(summary),flush=True)
