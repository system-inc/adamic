import json,os,subprocess,time
from pathlib import Path
P=Path('review/test-audit/bridge-tsgo-registered_units');S=Path('/tmp/u004');cache=Path('/home/agent/.cache/adamic-build')
archive=max(cache.glob('*/tsgo.a'),key=lambda p:p.stat().st_mtime); native=max(cache.glob('*/native'),key=lambda p:p.stat().st_mtime)
cmd=['clang','-std=c11','-Wall','-Wextra','-Werror','-pedantic','-O1','-g','-Ibridge/tsgo',str(P/'reset-witness.c'),str(archive),'-lpthread','-ldl','-lm','-o',str(S/'reset-witness')];s=time.monotonic();subprocess.run(cmd,check=True,stdout=open(S/'reset-build.log','w'),stderr=subprocess.STDOUT)
result=[{'build_command':' '.join(cmd),'build_seconds':time.monotonic()-s}]
manifest=S/'survivor.tsv';sample=str(Path('bridge/tsgo/testdata/sample.ts').resolve());manifest.write_text(sample+'\t14\n');config=str(Path('bridge/tsgo/testdata/tsconfig.json').resolve())
for id,cmd in [('M06',[str(S/'reset-witness')]),('M07',[str(native),config,str(manifest),sample])]:
 for selector in ['',id]:
  env=dict(os.environ,ADAMIC_MUTANT=selector,ADAMIC_TSGO_TIMING='1');s=time.monotonic();r=subprocess.run(cmd,capture_output=True,text=True,env=env,timeout=90);result.append({'mutant':id,'selector':selector,'command':'ADAMIC_MUTANT='+repr(selector)+' ADAMIC_TSGO_TIMING=1 '+' '.join(cmd),'exit':r.returncode,'stdout':r.stdout,'stderr':r.stderr,'seconds':time.monotonic()-s})
(P/'survivor-witnesses.json').write_text(json.dumps(result,indent=2));print(json.dumps(result,indent=2))
