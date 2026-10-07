import json,os,pathlib,shutil,subprocess,tempfile
repo=pathlib.Path('/workspace/adamic');reports=repo/'cloud/reports/node-pin';root=pathlib.Path(tempfile.mkdtemp(prefix='node-path-proof-',dir='/tmp/adamic-gate'));fixture=root/'repository';(fixture/'cloud').mkdir(parents=True)
shutil.copyfile(repo/'cloud/setup.sh',fixture/'cloud/setup.sh');(fixture/'go.mod').write_text('module node-setup-proof\n\ngo 1.27\n');(fixture/'main.go').write_text('package main\nfunc main() {}\n')
tools=pathlib.Path('/tmp/adamic-gate/node-pin-tools')
for name in ['go','llvm']:
 (tools/name).symlink_to(pathlib.Path('/workspace/adamic-tools')/name)
env=dict(os.environ,PATH='/tmp/adamic-gate/node-drift-tools/bin:/workspace/adamic-tools/go/bin:'+os.environ['PATH'],GOCACHE='/home/agent/.cache/go-build',GOMODCACHE='/tmp/adamic-gate/setup-modules-proof/with',GONOPROXY='github.com/klauspost/compress',XDG_CACHE_HOME='/tmp/adamic-gate/yaml-input-user-cache',ADAMIC_TOOLS=str(tools),ADAMIC_SETUP_REPOSITORY=str(fixture))
env.pop('ADAMIC_GATE_UNCACHED',None)
def run(args,name,cwd=fixture,environment=env,expected=0):
 with (reports/name).open('wb') as output:r=subprocess.run(args,cwd=cwd,env=environment,stdout=output,stderr=subprocess.STDOUT,timeout=600)
 text=(reports/name).read_text();assert (r.returncode==0)==(expected==0),(args,r.returncode,text);return text
run(['git','init'],'fixture-git.log');run(['git','add','.'],'fixture-add.log');run(['git','-c','user.name=Node proof','-c','user.email=proof@example.invalid','commit','-m','Setup fixture'],'fixture-commit.log')
assert run(['node','--version'],'before-node.log').strip()=='v24.21.0'
text=run(['go','test','./internal/oracle','-run','^$','-count=1'],'oracle-drift-refused.log',repo,expected=1)
assert 'got v24.21.0, want v24.19.0' in text,text
(tools/'node.stamp').unlink()
for number in range(2):
 text=run(['bash',str(repo/'cloud/setup.sh')],f'full-setup-{number}.log',repo)
 assert text.splitlines()[-1]=='setup: node v24.19.0',text
 assert 'node=v24.19.0' in text,text
 if number:assert 'node v24.19.0 skipped' in text,text
text=run(['bash','-c','source "$ADAMIC_TOOLS/env.sh"; node --version; go test ./internal/oracle -run "^TestTheOracleCatchesOneByte$" -count=1 -v'],'oracle-after-setup.log',repo)
assert text.splitlines()[0]=='v24.19.0',text
assert 'oracle: node v24.19.0 at /tmp/adamic-gate/node-pin-tools/bin/node' in text,text
print('PASS: real v24.21.0 on PATH refused; setup installs pinned version; warm skips; env/oracle select v24.19.0; final output line correct')
