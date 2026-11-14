"""One full file at a time, CommonJSBundle, 60-minute ceiling, no output."""
import argparse,json,os,pathlib,subprocess,sys
p=argparse.ArgumentParser();p.add_argument('road');p.add_argument('tree',type=pathlib.Path);p.add_argument('output',type=pathlib.Path);p.add_argument('--file');a=p.parse_args()
heat=json.loads(pathlib.Path('/workspace/scratch/receiver-heat.json').read_text())
files=[a.file] if a.file else list(dict.fromkeys(s['file'] for s in heat['sites'] if s['public']))
a.output.mkdir(parents=True,exist_ok=True)
for file in files:
 key=file.removeprefix('src/compiler/').removesuffix('.ts').replace('/','__');stem=a.output/(a.road+'__'+key)
 if stem.with_suffix('.json').exists() and json.loads(stem.with_suffix('.json').read_text())['exit']==0:continue
 env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',CENSUS_MODULE_FORMAT='CommonJSBundle',LATENT_ONLY_FILE=str(a.tree/file))
 raw=a.output/(a.road+'__'+key+'.jsonl')
 cmd=[sys.executable,'/workspace/scratch/receiver-measure.py','--out',str(stem),'--cwd','/workspace/scratch/receiver-calls-price','--seconds','3600','--','/workspace/scratch/receiver-latent',str(a.tree/'src/compiler'),str(raw)]
 result=subprocess.run(cmd,env=env);print(json.dumps({'road':a.road,'file':file,'exit':result.returncode}),flush=True)
 if result.returncode:raise SystemExit(result.returncode)
