import json, os, pathlib, re, shlex, subprocess, time
root=pathlib.Path(__file__).resolve().parents[4]
evidence=root/'stage1/cohere/lint/check-evidence'
evidence.mkdir(exist_ok=True)
pathlib.Path('/tmp/adamic-lint-original.go').write_bytes(subprocess.check_output(['git','show','f4d98cab50048692781da3599131317dc569d466:stage1/cohere/lint/lint_test.go'],cwd=root))
replace={str(root/'stage1/cohere/lint/lint_test.go'):'/tmp/adamic-lint-original.go'}
for name in ('cache_test.go','cache_mutants_test.go','rule_check_test.go','selection_test.go'):
    replace[str(root/'stage1/cohere/lint'/name)]=''
pathlib.Path('/tmp/adamic-lint-before-overlay.json').write_text(json.dumps({'Replace':replace}))
base_env=os.environ.copy()
base_env['GOCACHE']=subprocess.check_output(['go','env','GOCACHE'],cwd=root,text=True).strip()
metadata={
 'commit':subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip(),
 'nproc':subprocess.check_output(['nproc'],text=True).strip(),
 'cpu.max':pathlib.Path('/sys/fs/cgroup/cpu.max').read_text().strip(),
 'go':subprocess.check_output(['go','version'],cwd=root,text=True).strip(),
 'clang':subprocess.check_output(['clang','--version'],text=True).splitlines()[0],
 'node':subprocess.check_output(['node','--version'],text=True).strip(),
 'native_flags':'-std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all',
 'GOCACHE':base_env['GOCACHE'],
 'baseline':'original f4d98cab lint_test.go through Go overlay; all new cache/check test files excluded',
}
(evidence/'build-flags.json').write_text(json.dumps(metadata,indent=2)+'\n')
results=[]
for round_number in range(1,4):
 for slug in ('no-var','no-empty','eqeqeq'):
  cache=pathlib.Path('/tmp')/f'adamic-lint-measure-{os.getpid()}-{round_number}-{slug}'
  cache.mkdir()
  env=base_env.copy(); env['XDG_CACHE_HOME']=str(cache); env.pop('ADAMIC_GATE_UNCACHED',None)
  mutant=json.loads((root/'stage1/cohere/lint/rules'/slug/'mutant.json').read_text())['name']
  mutant_pattern='^'+re.escape(re.sub(r'\s','_',mutant))+'$'
  before=['go','test','-overlay=/tmp/adamic-lint-before-overlay.json','./stage1/cohere/lint','-run','^(TestRulesAgree|TestOwnedWitnesses|TestMutants)$/'+mutant_pattern,'-count=1','-timeout','30m','-v']
  cold=['go','test','./stage1/cohere/lint','-run','^TestRule$','-count=1','-timeout','30m','-v','-args','-rule',slug]
  warm=['go','test','./stage1/cohere/lint','-run','^TestRule$','-count=1','-timeout','30m','-v','-args','-rule',slug,'-rule-byte-change']
  unchanged=cold.copy()
  for mode,command in [('before',before),('cold',cold),('warm-byte-change',warm),('warm-unchanged',unchanged)]:
   load_before=pathlib.Path('/proc/loadavg').read_text().strip()
   log=evidence/f'{slug}-{round_number}-{mode}.log'
   with log.open('wb') as output:
    started=time.perf_counter()
    outcome=subprocess.run(command,cwd=root,env=env,stdout=output,stderr=output)
    seconds=time.perf_counter()-started
   load_after=pathlib.Path('/proc/loadavg').read_text().strip()
   row={'slug':slug,'round':round_number,'mode':mode,'seconds':seconds,'exit':outcome.returncode,'load_before':load_before,'load_after':load_after,'instrument':shlex.join(command),'cache':str(cache),'cache_mode':'original uncached harness' if mode=='before' else ('empty lint cache' if mode=='cold' else 'cached'),'log':log.name,**metadata}
   results.append(row)
   (evidence/'measurements.json').write_text(json.dumps(results,indent=2)+'\n')
   print(f'{slug} round={round_number} {mode}: {seconds:.3f}s exit={outcome.returncode}',flush=True)
   if outcome.returncode: raise SystemExit(f'failed: {log}')
