import pathlib,json,subprocess,time,os,re
D=pathlib.Path('review/test-audit/stage1-cohere-estree-gaps');names=['TestPostfixValueGap','TestMethodReplacementGap','TestRawInputGap','TestParserRecoveryGap','TestInterfaceDefaultGap','TestInterfaceTypeMethodGap','TestJSXAgreement','TestJSXAgreementPlantedDisagreement','TestJSXOriginalLibraries','TestJSXMutant','TestJSXMutantPlantedSurvivor','TestMiscShardUnionRejectsInvalidEnumeration','TestMiscShardGrowthKeepsAssignments','TestRecoveryCacheInputKeys','TestRecoveryNativeRecipe'];listed=(D/'test-list.log').read_text().splitlines();assert all(x in listed for x in names);os.environ['ADAMIC_ESTREE_LIBRARY']='/tmp/u086/library';records=[]
for name in names:
 regex='^TestJSXAgreement(_[0-9]{3})?$' if name=='TestJSXAgreement' else '^'+name+'$'
 for i in range(3):
  label='time-'+name+'-'+str(i+1);args=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/estree/','-run',regex];t=time.monotonic()
  with (D/(label+'.log')).open('w') as f:rc=subprocess.call(args,stdout=f,stderr=subprocess.STDOUT)
  records.append(dict(label=label,command=' '.join(args),wall_seconds=round(time.monotonic()-t,3),exit=rc));(D/'timing-commands.json').write_text(json.dumps(records,indent=2));print(label,rc,flush=True)
  if rc:raise RuntimeError('clean isolated run failed: '+label)
