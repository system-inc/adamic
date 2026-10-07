"""Quiet, alternating fresh-process native/Go timings for the three-rule profile."""
import argparse,json,os,pathlib,statistics,subprocess,time
parser=argparse.ArgumentParser()
for name in ['artifacts','native','go','compiler-root']:
 parser.add_argument('--'+name,required=True)
args=parser.parse_args();source=pathlib.Path(__file__).resolve().parent;repo=source.parents[3];root=pathlib.Path(args.artifacts).resolve();root.mkdir(parents=True,exist_ok=True)
records=[]
for corpus,prefix,config in [('compiler',args.compiler_root,pathlib.Path(args.compiler_root)/'src/compiler/tsconfig.json'),('repository',repo,repo/'tsconfig.json')]:
 manifest=root/(corpus+'.manifest');manifest.write_text(''.join(str(pathlib.Path(prefix)/path)+'\n' for path in (source.parent/'validation-coverage'/f'{corpus}.manifest').read_text().splitlines()))
 truth=None
 for round in range(3):
  for variant in (['go','native'] if round%2==0 else ['native','go']):
   binary=args.native if variant=='native' else args.go
   with open(root/f'{corpus}-{round}-{variant}.stdout','wb') as out,open(root/f'{corpus}-{round}-{variant}.stderr','wb') as err:
    start=time.perf_counter();p=subprocess.run([binary,str(config),str(manifest),'--all'],stdout=out,stderr=err,env=dict(os.environ,ADAMIC_TSGO_TIMING='1'));elapsed=time.perf_counter()-start
   output=(root/f'{corpus}-{round}-{variant}.stdout').read_bytes();errors=(root/f'{corpus}-{round}-{variant}.stderr').read_text()
   assert p.returncode==0
   if truth is None:truth=output
   assert output==truth
   records.append(dict(corpus=corpus,round=round,variant=variant,seconds=elapsed,stderr=errors))
 med={v:statistics.median(r['seconds'] for r in records if r['corpus']==corpus and r['variant']==v) for v in ['go','native']}
 print(corpus,med,'native/Go',med['native']/med['go'],flush=True)
(root/'results.json').write_text(json.dumps(records,indent=2)+'\n');print('PASS',flush=True)
