import pathlib,subprocess,os,json,time
r=pathlib.Path('/workspace/adamic/review/test-audit/stage1-cohere-typeaware-volume_profile_Controls');env=os.environ.copy();env.update(ADAMIC_TYPESCRIPT_SOURCE='/tmp/u147/typescript',ADAMIC_VOLUME_REPOSITORY_MANIFEST='/tmp/u147/repository.manifest',ADAMIC_VOLUME_COMPILER_MANIFEST='/tmp/u147/compiler.manifest')
rows=[('lower','^TestVolumeProfileControlsLower$'),('controls','^TestVolumeProfileControls(Union|_[0-9]+)$'),('missing-union','^TestShadowIndexMissingBindingUnion$'),('corpus-setup','^TestVolumeProfileCorpora_Setup$'),('corpus-bounded','^TestVolumeProfileCorpora(Union|_0[0-5][0-9]|_06[0-3])$'),('mutants','^TestVolumeProfileMutants(Union|_[0-9]+)$'),('overlays','^TestVolumeProfileOverlays(Union|_[0-9]+)$'),('partition','^TestVolumeProfilePartition$'),('agreement-bounded','^TestVolumeAgreementAndMutants(Union|_014|_030)$'),('config-setup','^TestVolumeConfigGuardAndMutant$'),('symbol','^TestVolumeTypeSymbol(Union|_000)$'),('symbol-planted','^TestVolumeTypeSymbolPlantedSurvivor$'),('deadline','^TestVolumeTypeSymbolCommandDeadline$')]
results=[]
for name,pattern in rows:
 for count in [1,2,3]:
  command=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/typeaware/','-run',pattern];start=time.monotonic()
  with (r/'logs'/f'timing-{name}-{count}.log').open('w') as f:rc=subprocess.run(command,stdout=f,stderr=subprocess.STDOUT,env=env).returncode
  results.append(dict(row=name,run=count,pattern=pattern,wall=time.monotonic()-start,returncode=rc));(r/'timing-runs.json').write_text(json.dumps(results,indent=2)+'\n')
  if rc:
   log=(r/'logs'/f'timing-{name}-{count}.log').read_text()
   if 'test timed out' not in log and rc!=124:
    (r/'BASELINE_RED.txt').write_text(f'{name}: non-timeout failure; stop audit.\n');raise SystemExit(1)
   break
