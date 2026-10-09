import pathlib,subprocess,os,difflib,json,time
repo=pathlib.Path('/workspace/adamic');r=repo/'review/test-audit/stage1-cohere-typeaware-volume_profile_Controls';p=repo/'stage1/cohere/typeaware/volume_suite.ts';s=p.read_text();v=s[:s.index('const args =')]+'/* empty-answer probe: entry body is disabled\n'+s[s.index('const args ='):]+'\n*/\n';(r/'Pport.diff').write_text(''.join(difflib.unified_diff(s.splitlines(True),v.splitlines(True),fromfile='a/stage1/cohere/typeaware/volume_suite.ts',tofile='b/stage1/cohere/typeaware/volume_suite.ts')))
env=os.environ.copy();env.update(ADAMIC_BUILD_CACHE_DIR='/tmp/u147/cache/Pport',ADAMIC_TYPESCRIPT_SOURCE='/tmp/u147/typescript',ADAMIC_VOLUME_REPOSITORY_MANIFEST='/tmp/u147/repository.manifest',ADAMIC_VOLUME_COMPILER_MANIFEST='/tmp/u147/compiler.manifest')
# Preserve adapter source markers inside a block comment, so preparation succeeds while the native module entry has no executable body.
pattern='^(TestVolumeProfileControls(Union|_[0-9]+)|TestVolumeProfileCorpora(Union|_0[0-5][0-9]|_06[0-3])|TestVolumeAgreementAndMutants(Union|_014|_030))$';p.write_text(v)
try:
 start=time.monotonic()
 with (r/'logs/Pport.log').open('w') as f:rc=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/typeaware/','-run',pattern],env=env,stdout=f,stderr=subprocess.STDOUT).returncode
 (r/'probe-runs.json').write_text(json.dumps([dict(id='Pport',pattern=pattern,wall=time.monotonic()-start,returncode=rc)],indent=2))
finally:p.write_text(s)
