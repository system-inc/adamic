import json,os,pathlib,shutil,subprocess,time
root=pathlib.Path(__file__).resolve().parent
env=dict(os.environ,PATH='/workspace/adamic-tools/bin:/workspace/adamic-tools/llvm/bin:/workspace/adamic-tools/go/bin',ADAMIC_BUILD_STORE='off')
assert shutil.which('git',path=env['PATH']) is None
start=time.monotonic()
with (root/'no-git.log').open('w') as log:
 r=subprocess.run(['/usr/bin/timeout','60','/tmp/adamic-recut4.test','-test.run=^TestCommandRuntimeArchivesHaveOneBuilder$','-test.timeout=55s','-test.v'],cwd='cmd/adamic',env=env,stdout=log,stderr=subprocess.STDOUT)
row=dict(path=env['PATH'],git_found=False,seconds=time.monotonic()-start,exit=r.returncode)
(root/'no-git.json').write_text(json.dumps(row,indent=2)+'\n');print(row)
assert r.returncode==0
