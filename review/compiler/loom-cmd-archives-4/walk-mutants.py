import difflib,pathlib,subprocess
root=pathlib.Path(__file__).resolve().parent
path=pathlib.Path('cmd/adamic/runtime_owners_test.go');original=path.read_text()
mutants=[
 ('ignored-directory','if match {','if false && match {','decoy.go'),
 ('fixture-directory','case "testdata", "vendor", "node_modules", ".git":','case "vendor", "node_modules", ".git":','testdata/decoy.go'),
 ('submodule-directory','if slices.Contains(submodules, name) {','if false && slices.Contains(submodules, name) {','cohere/decoy.go'),
 ('missing-producer','files = append(files, name)','if !strings.HasPrefix(name, "ordinary/") { files = append(files, name) }','census builder packages'),
]
try:
 for name,old,new,message in mutants:
  assert original.count(old)==1,(name,original.count(old))
  changed=original.replace(old,new)
  (root/(name+'.patch')).write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/'+str(path),tofile='b/'+str(path),n=0)))
  path.write_text(changed)
  with (root/(name+'.log')).open('w') as log:r=subprocess.run(['timeout','90','go','test','./cmd/adamic','-run','^TestCommandRuntimeCensusWalksSourceTree$','-count=1','-timeout','55s'],stdout=log,stderr=subprocess.STDOUT)
  output=(root/(name+'.log')).read_text();assert r.returncode==1 and message in output,(name,r.returncode,output)
  print(name+': caught by TestCommandRuntimeCensusWalksSourceTree',flush=True);path.write_text(original)
finally:path.write_text(original)
# Reintroduce the original checkout assumption and run its real ownership test
# with exactly the no-Git PATH used for the positive proof.
import json,os,shutil
old='return runtimeBuilderPackagesIn(t, filepath.Join("..", ".."))'
new='''command := exec.Command("git", "ls-files", "-z", "--", "*.go")
 command.Dir = filepath.Join("..", "..")
 if _, err := command.Output(); err != nil { t.Fatal(err) }
 '''+old
assert original.count(old)==1
changed=original.replace(old,new)
(root/'git-required.patch').write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/'+str(path),tofile='b/'+str(path),n=0)))
try:
 path.write_text(changed)
 with (root/'git-required-build.log').open('w') as log:
  r=subprocess.run(['timeout','120','go','test','-c','-o','/tmp/adamic-recut4-git-mutant.test','./cmd/adamic'],stdout=log,stderr=subprocess.STDOUT)
 assert r.returncode==0
 env=dict(os.environ,PATH='/workspace/adamic-tools/bin:/workspace/adamic-tools/llvm/bin:/workspace/adamic-tools/go/bin',ADAMIC_BUILD_STORE='off')
 assert shutil.which('git',path=env['PATH']) is None
 with (root/'git-required.log').open('w') as log:
  r=subprocess.run(['/usr/bin/timeout','60','/tmp/adamic-recut4-git-mutant.test','-test.run=^TestCommandRuntimeArchivesHaveOneBuilder$','-test.timeout=55s','-test.v'],cwd='cmd/adamic',env=env,stdout=log,stderr=subprocess.STDOUT)
 assert r.returncode==1 and 'executable file not found' in (root/'git-required.log').read_text()
 print('git-required: caught by TestCommandRuntimeArchivesHaveOneBuilder with no git in PATH',flush=True)
finally:path.write_text(original)
