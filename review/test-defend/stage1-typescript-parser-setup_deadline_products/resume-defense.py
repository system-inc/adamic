import pathlib,json,time,subprocess,os
p=pathlib.Path(__file__).resolve().parent
plan=json.loads((p/'mutant-plan.json').read_text())['D1']
f=pathlib.Path(plan['file']);text=subprocess.check_output(['git','show','HEAD:'+str(f)],text=True)
assert f.read_text()==text.replace(plan['from'],plan['to'],1)
assert 'sanitized native build seconds' in (p/'D1-native-build.log').read_text()
env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/parser-defense/cache/D1';env['ADAMIC_TYPESCRIPT_SOURCE']='/tmp/u156-typescript'
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
