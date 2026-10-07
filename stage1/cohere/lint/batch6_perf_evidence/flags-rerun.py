from pathlib import Path
import subprocess,shlex,json,time,hashlib,statistics,os
p=Path('/workspace/scratch/batch6-perf/flags-rerun'); parent=p.parent
common=['-std=c11','-Wall','-Wextra','-Werror','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls']
commands={}
for name in ['baseline','optimized']:
 for mode,flags in [('release',['-O2']),('sanitized',['-O1','-g','-fsanitize=address,undefined','-fno-sanitize-recover=all'])]:
  key=name+'-'+mode; command=['/workspace/adamic-tools/llvm/bin/clang']+common+flags+['-I','internal/native/runtime',str(parent/(name+'.c'))]+sorted(str(x) for x in Path('internal/native/runtime').glob('*.c'))+['-lm','-o',str(p/key)]
  commands[key]=command
(p/'build-commands.json').write_text(json.dumps(commands,indent=2)+'\n')
(p/'build-commands.sh').write_text('\n'.join(shlex.join(c) for c in commands.values())+'\n')
for key,command in commands.items():
 with (p/(key+'-build.log')).open('wb') as log: subprocess.run(command,stdout=log,stderr=log,check=True)
run_commands={key:[str(p/key),'--manifest',str(parent/'compiler.txt')] for key in commands}
run_commands['Go']=[str(parent/'go-oracle'),'--manifest',str(parent/'compiler.txt')]
env=os.environ.copy(); env.update(ASAN_OPTIONS='detect_leaks=1:halt_on_error=1',UBSAN_OPTIONS='halt_on_error=1')
want=None
for key,command in run_commands.items():
 with (p/(key+'-findings.txt')).open('wb') as out,(p/(key+'-correctness.log')).open('wb') as err: subprocess.run(command,stdout=out,stderr=err,env=env,check=True)
 data=(p/(key+'-findings.txt')).read_bytes()
 if want is None: want=data
 assert data==want,key
 assert not (p/(key+'-correctness.log')).read_bytes(),key
results={key:[] for key in run_commands}
for r in range(5):
 order=list(run_commands); order=order[r%len(order):]+order[:r%len(order)]
 for key in order:
  start=time.perf_counter()
  with (p/(key+'-count.txt')).open('wb') as out,(p/(key+'-stderr.log')).open('wb') as err: subprocess.run(run_commands[key]+['--count'],stdout=out,stderr=err,env=env,check=True)
  elapsed=time.perf_counter()-start
  assert (p/(key+'-count.txt')).read_bytes()==b'481\n'
  assert not (p/(key+'-stderr.log')).read_bytes()
  results[key].append(elapsed)
  print(r+1,key,elapsed,flush=True)
summary={'files':77,'findings':481,'identical_bytes':len(want),'runtime_counting_build':False,'runner_output_mode':'--count','seconds':results,'best_seconds':{k:min(v) for k,v in results.items()},'best_findings_per_second':{k:481/min(v) for k,v in results.items()},'median_seconds':{k:statistics.median(v) for k,v in results.items()},'run_commands':{k:v+['--count'] for k,v in run_commands.items()},'build_commands':commands,'sha256':{k:hashlib.sha256(Path(v[0]).read_bytes()).hexdigest() for k,v in run_commands.items()},'manifest_sha256':hashlib.sha256((parent/'compiler.txt').read_bytes()).hexdigest()}
(p/'results.json').write_text(json.dumps(summary,indent=2)+'\n'); print(json.dumps(summary,indent=2))
