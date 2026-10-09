import pathlib,subprocess,os,time,json,shlex
p=pathlib.Path('review/test-audit/internal-oracle-stage3_front');names=['TestStage3EnumFallthrough','TestStage3FixtureHook','TestStatementsSmallRulings','TestStatementsSmallTypeScriptIncrement','TestSwitchCaseDeadZoneStop','TestSwitchCaseDeclarationMutants','TestSwitchCaseOverloadIsNotYet','TestSwitchTrailingEmptyMutant','TestSyntaxModuleDeclarationMutants','TestSyntaxSubstrMutants','TestTypedArrayWriteStopIsPinned','TestTypedArrayWorkersStatsMatchesPlatforms'];rs=[]
family='^TestNativeAgreesWithNode$/internal/oracle/testdata/(switch_case_declarations|93122fa_s[12]|switch_empty_neighbors|syntax_module_declarations|syntax_substr|typed_arrays_(stop|workers_stats))'
for name in names+['TestNativeAgreesWithNode-subset']:
 for i in range(3):
  env=os.environ.copy();env['ADAMIC_GATE_UNCACHED']='1';pattern=family if name.endswith('-subset') else '^'+name+'$'
  if name=='TestStage3FixtureHook':env['ADAMIC_STAGE3_FIXTURE']=str(pathlib.Path('internal/oracle/testdata/switch_case_declarations/fallthrough.a').resolve());env['ADAMIC_STAGE3_RESULT']='/tmp/u070/hook-timing-'+str(i)+'.json'
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',pattern];t=time.monotonic()
  with open(p/f'timing-{name}-{i}.log','w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,env=env)
  rs.append(dict(test=name,trial=i,command='ADAMIC_GATE_UNCACHED=1 '+shlex.join(cmd),env={k:v for k,v in env.items() if k in ['ADAMIC_STAGE3_FIXTURE','ADAMIC_STAGE3_RESULT']},exit=r.returncode,wall_seconds=time.monotonic()-t));(p/'timing.json').write_text(json.dumps(rs,indent=2));print(name,i,r.returncode,flush=True)
  if r.returncode:raise SystemExit(1)
(p/'scope.json').write_text(json.dumps(dict(rows=names,extra_pattern=family),indent=2))
