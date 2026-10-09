import pathlib,subprocess,os,json,difflib,time
repo=pathlib.Path('/workspace/adamic');r=repo/'review/test-audit/stage1-cohere-typeaware-volume_profile_Controls';plan=json.loads((r/'plan.json').read_text());results=[]
# Actual reached family leaves; construction and witnesses are separate experiments.
pattern='^(TestVolumeProfileControls(Union|_[0-9]+)|TestVolumeProfileCorpora(Union|_0[0-5][0-9]|_06[0-3])|TestVolumeAgreementAndMutants(Union|_014|_030))$'
for m in plan['mutants']:
 p=repo/m['file'];s=p.read_text();changed=s.replace(m['from_'],m['to'],1);assert changed!=s
 (r/(m['id']+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),changed.splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file'])))
 p.write_text(changed);env=os.environ.copy();env.update(ADAMIC_TYPESCRIPT_SOURCE='/tmp/u147/typescript',ADAMIC_VOLUME_REPOSITORY_MANIFEST='/tmp/u147/repository.manifest',ADAMIC_VOLUME_COMPILER_MANIFEST='/tmp/u147/compiler.manifest',ADAMIC_BUILD_CACHE_DIR='/tmp/u147/cache/'+m['id'])
 try:
  # Phased source build before matrix. Same source key as all profile products.
  phases=[('build','^TestVolumeProfileControlsLower$'),('matrix',pattern)]
  for phase,selection in phases:
   cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/typeaware/','-run',selection];t=time.monotonic();log=r/'logs'/(m['id']+'-'+phase+'.log')
   with log.open('w') as f:rc=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,env=env).returncode
   results.append(dict(id=m['id'],phase=phase,pattern=selection,wall=time.monotonic()-t,returncode=rc,cache=env['ADAMIC_BUILD_CACHE_DIR']));(r/'matrix-runs.json').write_text(json.dumps(results,indent=2)+'\n')
   if phase=='build' and rc:break
   if phase=='matrix' and ('test timed out' in log.read_text() or rc==124):
    bounded='^(TestVolumeProfileControls(Union|_[0-9]+)|TestVolumeProfileCorpora(Union|_0[0-5][0-9]|_06[0-3]))$'
    cmd[-1]=bounded;t=time.monotonic()
    with (r/'logs'/(m['id']+'-bounded.log')).open('w') as f:rc=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,env=env).returncode
    results.append(dict(id=m['id'],phase='bounded',pattern=bounded,wall=time.monotonic()-t,returncode=rc,cache=env['ADAMIC_BUILD_CACHE_DIR']));(r/'matrix-runs.json').write_text(json.dumps(results,indent=2)+'\n')
 finally:p.write_text(s)
