import pathlib,subprocess,time,json,hashlib,shutil,os,sys
root=pathlib.Path('/workspace/or3-report');out=root/'stage3/census/overload-result-02d48211'
children=[]
def launch(name,cmd,cwd=root,env=None):
 log=open('/tmp/or3-'+name+'.log','wb')
 p=subprocess.Popen(cmd,cwd=cwd,stdout=log,stderr=subprocess.STDOUT,env=env);log.close();children.append(p)
 print('start '+name+' pid='+str(p.pid),flush=True);return p

def finish(name,p):
 rc=p.wait();print(name+' child_exit='+str(rc),flush=True)
 if rc:raise RuntimeError(name+' failed')
def run(name,cmd,cwd=root,env=None):finish(name,launch(name,cmd,cwd,env))
def git(*args):
 p=subprocess.run(['git',*args],cwd=root,capture_output=True,text=True);print('git child_exit='+str(p.returncode),flush=True);p.check_returncode();return p.stdout.strip()
try:
 base=git('rev-parse','19ab1c58');head=git('rev-parse','02d48211');old=git('rev-parse','efe88776')
 assert git('rev-parse','02d48211^')==base
 anc=subprocess.run(['git','merge-base','--is-ancestor',old,base],cwd=root).returncode
 paths=['stage3/source.json','stage3/apply.py','stage3/apply.sh','stage3/api','stage3/adapt']
 changes=git('diff','--name-only',old,base,'--',*paths).splitlines()
 data={'parent':base,'head':head,'old_parent':old,'base_descends_from_old':anc==0,'adaptation_inputs_identical':not changes,'changed_adaptation_inputs':changes,'reuse_parent_streams':False,'reason':'Changed adaptation inputs require a fresh parent compiler and census at 19ab1c58.','branch_tips':pathlib.Path('/tmp/or3-remote-tips.log').read_text(),'upstream':git('-C','/workspace/or3-upstream','rev-parse','HEAD')}
 assert data['upstream']=='050880ce59e30b356b686bd3144efe24f875ebc8'
 (out/'base-check.json').write_text(json.dumps(data,indent=2)+'\n')
 (out/'adaptation-input-diff.txt').write_text(git('diff',old,base,'--',*paths)+'\n')
 prep={}
 for label in ['parent','head']:
  script='''set -u
source /workspace/adamic-tools/env.sh
git -c submodule.cohere.url=https://github.com/system-inc/cohere.git submodule update --init --recursive --depth 1 > /tmp/or3-LABEL-submodules.log 2>&1
rc=$?; echo submodules_child_exit=$rc; test "$rc" = 0 || exit "$rc"
bash /workspace/or3-build.sh LABEL
rc=$?; echo build_child_exit=$rc; exit "$rc"
'''.replace('LABEL',label)
  prep[label]=launch(label+'-prepare',['bash','-c',script],'/workspace/or3-'+label)
 applyscript='''set -u
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/or3-upstream
npm ci --prefix stage3/api > /tmp/or3-npm.log 2>&1
rc=$?; echo npm_child_exit=$rc; test "$rc" = 0 || exit "$rc"
python3 /workspace/or3-run.py /tmp/or3-apply.log /tmp/or3-apply.time.json bash stage3/apply.sh /workspace/or3-adapted
rc=$?; echo apply_child_exit=$rc; exit "$rc"
'''
 apply=launch('apply-prepare',['bash','-c',applyscript],'/workspace/or3-parent')
 finish('apply-prepare',apply)
 manifest={str(p.relative_to('/workspace/or3-adapted')):hashlib.sha256(p.read_bytes()).hexdigest() for p in pathlib.Path('/workspace/or3-adapted/src').rglob('*') if p.is_file()}
 (out/'adapted-source-hashes.json').write_text(json.dumps(manifest,indent=2)+'\n');shutil.copyfile('/workspace/or3-adapted/patch-set.md',out/'patch-set.md')
 oldmanifest=json.loads((root/'stage3/census/overload-result-10b6c33/adapted-source-hashes.json').read_text())
 (out/'adapted-difference-from-old.json').write_text(json.dumps({k:{'old':oldmanifest.get(k),'new':manifest.get(k)} for k in sorted(manifest.keys()|oldmanifest.keys()) if manifest.get(k)!=oldmanifest.get(k)},indent=2)+'\n')
 run('checkpoint-apply',['python3','/workspace/or3-checkpoint.py','02d48211 apply complete'])
 for label,p in prep.items():finish(label+'-prepare',p)
 print('waiting for setup completion',flush=True)
 while 'setup: done on ' not in pathlib.Path('/tmp/or3-setup.log').read_text():time.sleep(5)
 for label in ['parent','head']:
  source=pathlib.Path('/workspace/or3-'+label);build=pathlib.Path('/workspace/or3-build-'+label)
  data[label+'_submodules']=git('-C',str(source),'submodule','status','--recursive')
  assert not git('-C',str(source),'status','--porcelain','--untracked-files=no')
  data[label+'_binary_sha256']=hashlib.sha256((build/'entry-census').read_bytes()).hexdigest()
 (out/'provenance.json').write_text(json.dumps(data,indent=2)+'\n')
 run('box',['python3','/workspace/or3-parent/stage3/meter/box.py',str(out/'box.json'),'start'])
 active={}
 for label in ['parent','head']:
  build=pathlib.Path('/workspace/or3-build-'+label)
  env=dict(os.environ,LATENT_ASSERT_NO_OUTPUT='1',LATENT_FULL='0',LATENT_TIMES=str(build/'latent-times.jsonl'),ADAMIC_TYPESCRIPT_SOURCE='/workspace/or3-upstream')
  for k in list(env):
   if k.startswith('LATENT_') and k not in ['LATENT_ASSERT_NO_OUTPUT','LATENT_FULL','LATENT_TIMES']:del env[k]
  active[label]=launch(label+'-measure',['python3','/workspace/or3-run.py',str(build/'latent.log'),str(build/'latent.time.json'),str(build/'entry-census'),'/workspace/or3-adapted/src/tsc/tsc.ts',str(build/'latent.jsonl')],env=env)
 while active:
  for label,p in list(active.items()):
   if p.poll() is not None:
    finish(label+'-measure',p);del active[label]
    run('checkpoint-'+label,['python3','/workspace/or3-checkpoint.py','02d48211 '+label+' run complete'])
  if active:time.sleep(5)
 run('pair',['python3','/workspace/or3-run.py',str(out/'pair.log'),str(out/'pair.time.json'),'python3','/workspace/or3-parent/stage3/census/pair/pair.py','/workspace/or3-build-parent/latent.jsonl','/workspace/or3-build-head/latent.jsonl','/workspace/or3-adapted','--typescript','/workspace/or3-parent/stage3/api/node_modules/typescript','--json',str(out/'delta.json'),'--report',str(out/'pair-report.md')])
 print('paired streams complete; ready for unit analysis',flush=True)
except BaseException:
 for p in children:
  if p.poll() is None:
   p.terminate();p.wait()
 raise
