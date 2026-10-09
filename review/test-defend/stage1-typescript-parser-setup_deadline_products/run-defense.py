import pathlib,json,time,subprocess,os
p=pathlib.Path(__file__).resolve().parent
while True:
 runs=json.loads((p/'clean-runs.json').read_text()) if (p/'clean-runs.json').exists() else []
 if any(r['exit'] for r in runs): raise SystemExit('Clean bounded baseline failed; no mutation planted')
 if len(runs)==len(json.loads((p/'groups.json').read_text())):
  perf=p/'clean-performance-enabled.log'
  events=[json.loads(s) for s in perf.read_text().splitlines() if s.startswith('{')] if perf.exists() else []
  done=[e for e in events if 'Test' not in e and e.get('Action') in ['pass','fail']]
  if done and done[-1]['Action']=='fail': raise SystemExit('Enabled performance baseline failed')
  if done: break
 time.sleep(2)
plan=json.loads((p/'mutant-plan.json').read_text())['D1']
f=pathlib.Path(plan['file']);text=f.read_text();assert text.count(plan['from'])==1
f.write_text(text.replace(plan['from'],plan['to'],1))
with (p/'D1.diff').open('w') as out: subprocess.check_call(['git','diff','HEAD','--',str(f)],stdout=out)
env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/parser-defense/cache/D1';env['ADAMIC_TYPESCRIPT_SOURCE']='/tmp/u156-typescript'
with (p/'D1-native-build.log').open('w') as out:
 started=time.monotonic();rc=subprocess.call(['timeout','120','go','run',str(p/'build-probe.go'),'/tmp/parser-defense/D1-native'],env=env,stdout=out,stderr=subprocess.STDOUT)
print('D1 native build',rc,round(time.monotonic()-started,3),flush=True)
if rc: raise SystemExit('Native mutant compile failed')
fixture=pathlib.Path('/tmp/parser-defense/with.ts');fixture.write_text('with (x) y();')
for label,cmd in [('native',['/tmp/parser-defense/D1-native',str(fixture),'--whole']),('Node',['node','--disable-warning=ExperimentalWarning','oracle/node.mjs','stage1/typescript/parser/main.ts',str(fixture),'--whole'])]:
 with (p/('D1-witness-'+label+'.log')).open('w') as out: subprocess.check_call(cmd,env=env,stdout=out,stderr=subprocess.STDOUT)
# Clean whole-package run exceeded 90 s. Replay all current rows in bounded groups.
subprocess.check_call(['python3',str(p/'run-matrix.py'),'D1'],env=env)
env['ADAMIC_PARSER_BENCH']='1'
for row in ['TestPerformance','TestWholePerformance']:
 with (p/('D1-'+row+'-enabled.log')).open('w') as out:
  rc=subprocess.call(['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/typescript/parser/','-run','^'+row+'$'],env=env,stdout=out,stderr=subprocess.STDOUT)
 print('D1 enabled performance',row,rc,flush=True)
f.write_text(text)
print('Production source restored',flush=True)
