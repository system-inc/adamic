import os, subprocess, time, json, hashlib
from pathlib import Path
root=Path('/workspace/tsgo-build-flags')
old=Path('/workspace/tsgo-speed/validation-pass')
env=dict(os.environ, TMPDIR=str(root),PATH=str(root/'wrapper')+':'+os.environ['PATH'])
def command(name,args,extra={}):
 with (root/(name+'.stdout')).open('wb') as out,(root/(name+'.stderr')).open('wb') as err:
  start=time.perf_counter_ns()
  p=subprocess.run(args,env=dict(env,**extra),stdout=out,stderr=err)
  elapsed=time.perf_counter_ns()-start
 if p.returncode: raise RuntimeError((name,p.returncode))
 return elapsed
command('stage0-build',['go','build','-o',str(root/'adamic'),'./cmd/adamic'])
for suite in ['coverage','volume']:
 for mode in ['release','sanitized']:
  archive=old/('checker.a' if mode=='release' else 'checker-asan.a')
  args=[str(root/'adamic'),'build',f'stage1/cohere/typeaware/{suite}_suite.ts','-o',str(root/f'{suite}-{mode}'),'--tsgo',str(archive)]
  if mode=='sanitized': args+=['--sanitize']
  command(f'{suite}-{mode}-build',args,{'CLANG_CAPTURE':str(root/f'{suite}-{mode}.clang')})
  print('built',suite,mode,flush=True)
records=[]
config='/tmp/tsgo-typescript/src/compiler/tsconfig.json'
manifest='/tmp/tsgo-profile/final/compiler.manifest'
for suite in ['coverage','volume']:
 expected=None
 for mode in ['go','release','sanitized']:
  binary=old/f'{suite}-oracle' if mode=='go' else root/f'{suite}-{mode}'
  # Full streams check findings and fixes, not merely counts.
  args=[str(binary),config,manifest]
  elapsed=command(f'{suite}-{mode}-run',args,{'ADAMIC_TSGO_TIMING':'1'})
  output=(root/f'{suite}-{mode}-run.stdout').read_bytes()
  if expected is None: expected=output
  assert output==expected, (suite,mode,'different findings')
  golden=old/('039-compiler-coverage-native.stdout' if suite=='coverage' else '043-compiler-volume-native.stdout')
  assert output==golden.read_bytes(),(suite,mode,'different historical findings')
  records.append(dict(suite=suite,mode=mode,command=args,process_ns=elapsed,bytes=len(output),sha256=hashlib.sha256(output).hexdigest(),stderr=(root/f'{suite}-{mode}-run.stderr').read_text()))
  (root/'results.json').write_text(json.dumps(records,indent=2)+'\n')
  print(suite,mode,elapsed/1e9,'byte-identical',flush=True)
# Exact previous headline mode; build remains uncounted.
for mode in ['go','release','sanitized']:
 binary=old/'coverage-oracle' if mode=='go' else root/f'coverage-{mode}'
 args=[str(binary),config,manifest,'--count']
 elapsed=command(f'headline-{mode}',args,{'ADAMIC_TSGO_TIMING':'1'})
 assert (root/f'headline-{mode}.stdout').read_bytes()==b'findings 16589\n'
 records.append(dict(suite='headline-count-only',mode=mode,command=args,process_ns=elapsed,stderr=(root/f'headline-{mode}.stderr').read_text()))
 (root/'results.json').write_text(json.dumps(records,indent=2)+'\n')
 print('headline',mode,elapsed/1e9,flush=True)
