"""Paired whole/split author-command timings; clang wrapper counts real object builds."""
import collections, json, os, pathlib, shlex, shutil, subprocess, time
root = pathlib.Path(__file__).resolve().parents[4]
evidence = pathlib.Path(__file__).resolve().parent
work = pathlib.Path('/tmp/adamic-lint-split-measure')
work.mkdir(exist_ok=True)
bin_dir = work / 'bin'; bin_dir.mkdir(exist_ok=True)
real_clang = shutil.which('clang')
events = work / 'clang-events.jsonl'
wrapper = bin_dir / 'clang'
wrapper.write_text('#!/usr/bin/python3\nimport json,os,sys,time\n'
    + f'fd=os.open({str(events)!r},os.O_WRONLY|os.O_CREAT|os.O_APPEND,0o600)\n'
    + 'os.write(fd,(json.dumps({"argv":sys.argv[1:],"cwd":os.getcwd(),"time":time.time()})+"\\n").encode());os.close(fd)\n'
    + f'os.execv({real_clang!r},[{real_clang!r}]+sys.argv[1:])\n')
wrapper.chmod(0o755)
env = os.environ.copy()
env['GOCACHE'] = subprocess.check_output(['go','env','GOCACHE'],text=True).strip()
env['GOMAXPROCS'] = '4'
env['PATH'] = str(bin_dir) + ':' + env['PATH']
env.pop('ADAMIC_GATE_UNCACHED',None)
metadata = {name: subprocess.check_output(command,text=True).strip().splitlines()[0] for name,command in {
 'commit':['git','rev-parse','HEAD'], 'nproc':['nproc'], 'go':['go','version'],
 'clang':[real_clang,'--version'], 'node':['node','--version']}.items()}
metadata.update({'cpu.max':pathlib.Path('/sys/fs/cgroup/cpu.max').read_text().strip(),
 'GOMAXPROCS':'4','jobs':'4 per split build; two builds concurrent',
 'flags':'-std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all',
 'GOCACHE':env['GOCACHE'],'clang_wrapper':str(wrapper)})
(evidence/'build-flags.json').write_text(json.dumps(metadata,indent=2)+'\n')
results=[]
for round_number in range(1,4):
 for slug in ('no-var','no-empty','eqeqeq'):
  caches={compiler: work/f'cache-{os.getpid()}-{round_number}-{slug}-{compiler}' for compiler in ('whole','split')}
  for mode,extra in [('cold',[]),('warm-byte-edit',['-rule-byte-change']),('warm-unchanged',[]),('helper-edit',['-rule-helper-change'])]:
   for compiler in ('whole','split'):
    run_env=env.copy();run_env['XDG_CACHE_HOME']=str(caches[compiler])
    command=['go','test','./stage1/cohere/lint','-run','^TestRule$','-count=1','-timeout','30m','-v','-args','-rule',slug]+extra+(['-rule-whole'] if compiler=='whole' else [])
    events.write_text('')
    load_before=pathlib.Path('/proc/loadavg').read_text().strip()
    log=evidence/f'{slug}-{round_number}-{mode}-{compiler}.log'
    with log.open('wb') as output:
     started=time.perf_counter();outcome=subprocess.run(command,cwd=root,env=run_env,stdout=output,stderr=output);seconds=time.perf_counter()-started
    records=[json.loads(line) for line in events.read_text().splitlines()]
    trace=log.with_suffix('.clang.json');trace.write_text(json.dumps(records,indent=2)+'\n')
    compiled=[item for item in records if '-c' in item['argv'] and 'cpp-output' in item['argv']]
    groups=collections.Counter(str(pathlib.Path(item['cwd']).parent) for item in compiled)
    row={**metadata,'slug':slug,'round':round_number,'mode':mode,'compiler':compiler,'seconds':seconds,
      'exit':outcome.returncode,'load_before':load_before,'load_after':pathlib.Path('/proc/loadavg').read_text().strip(),
      'instrument':shlex.join(command),'cache':str(caches[compiler]),'cache_mode':'empty lint/native caches' if mode=='cold' else 'cached',
      'log':log.name,'trace':trace.name,'compiled_objects':len(compiled),'compiled_by_build':dict(groups)}
    results.append(row);(evidence/'measurements.json').write_text(json.dumps(results,indent=2)+'\n')
    print(f'{slug} round={round_number} {mode} {compiler}: {seconds:.3f}s objects={len(compiled)} exit={outcome.returncode}',flush=True)
    if outcome.returncode:raise SystemExit(f'failed: {log}')
