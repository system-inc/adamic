import pathlib,subprocess,time,sys,json,hashlib,gzip,shutil,os
sys.dont_write_bytecode=True
out=pathlib.Path(__file__).resolve().parents[1];follow=out/'continuation';repo=out.parents[2]
old='5a9cf385c47a257a1c8c6fcef353ef80a9420b42'
def invoke(command,log):
 log.parent.mkdir(parents=True,exist_ok=True)
 with log.open('wb') as target:child=subprocess.run(command,cwd=repo,stdout=target,stderr=subprocess.STDOUT)
 print('child exit='+str(child.returncode)+' '+repr(command),flush=True)
 if child.returncode:raise SystemExit(child.returncode)
def git(*args):return subprocess.check_output(['git',*args],cwd=repo,text=True).strip()
print('waiting for the bounded variant controller',flush=True)
wait_pid=int(sys.argv[1]) if len(sys.argv)>1 else None
while wait_pid and pathlib.Path('/proc/'+str(wait_pid)).exists():time.sleep(2)
for i in range(30):
 text=pathlib.Path('/tmp/baobab-variants-controller.log').read_text()
 if 'variant workflow child exit=' in text:break
 time.sleep(1)
assert 'variant workflow child exit=0' in text,'variant controller did not complete successfully'
assert git('rev-parse','HEAD')==old,'report branch advanced unexpectedly'
for i,command in enumerate([
 ['python3',str(out/'collect.py')],['node',str(out/'classify.cjs')],
 ['python3',str(follow/'compare-sites.py')],['python3',str(follow/'followup-report.py')],
 ['python3',str(out/'report.py')],['python3',str(out/'audit.py')],
 ['python3',str(follow/'audit-followup.py')],['python3',str(out/'seal-inputs.py')]]):invoke(command,follow/'preparation'/('final-'+str(i)+'.log'))
for source,dest in [('/tmp/baobab-long-controller.log','baseline-controller.log'),('/tmp/baobab-variants-controller.log','variant-controller.log'),('/tmp/baobab-src-deps.log','dependencies.log'),('/tmp/baobab-first-stop-build.log','first-stop-build.log'),('/tmp/baobab-original-source-audit.log','original-source-audit.log'),('/tmp/baobab-original-emit.log','original-public-emit.log')]:
 shutil.copyfile(source,follow/'preparation'/dest)
metadata={'first_report_sha':old,'main_sha':'d690f276749f91e6b36fe3a9726a5a453fcd7d39','typescript_sha':'050880ce59e30b356b686bd3144efe24f875ebc8','baseline_bound_seconds':1500,'baseline_max_concurrency':1,'variants_max_concurrency':2,'stock_exactOptionalPropertyTypes':False,'adamic_exactOptionalPropertyTypes':True,'followup_audit':json.loads((follow/'audit-result.json').read_text())}
(follow/'provenance.json').write_text(json.dumps(metadata,indent=2)+'\n')
for p in sorted(out.rglob('*')):
 if p.is_file() and p.suffix!='.gz' and p.stat().st_size>1000000:
  with p.open('rb') as src,gzip.open(str(p)+'.gz','wb') as target:shutil.copyfileobj(src,target)
  p.unlink()
manifest=out/'SHA256SUMS';manifest.write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+p.relative_to(out).as_posix()+'\n' for p in sorted(out.rglob('*')) if p.is_file() and p!=manifest))
invoke(['python3',str(out/'checkpoint.py'),'followup-final'],pathlib.Path('/tmp/baobab-followup-final-checkpoint.log'))
paths=git('diff','--cached','--name-only').splitlines();assert paths and all(p.startswith('stage3/census/sourcefile-like-sites/') for p in paths)
assert not git('diff','--name-only')
tracked=set(git('ls-files',str(out.relative_to(repo))).splitlines());existing={str(p.relative_to(repo)) for p in out.rglob('*') if p.is_file()};assert tracked==existing
for line in manifest.read_text().splitlines():
 digest,name=line.split('  ',1);assert hashlib.sha256((out/name).read_bytes()).hexdigest()==digest,name
invoke(['git','commit','-m','census: measure SourceFileLike road 2 variants (baobab)'],pathlib.Path('/tmp/baobab-followup-commit.log'))
assert git('rev-parse','HEAD^')==old
assert git('rev-list','--count','d690f276749f91e6b36fe3a9726a5a453fcd7d39..HEAD')=='2'
assert not git('status','--porcelain')
invoke(['git','push','origin','HEAD:refs/heads/typescript/sourcefile-like-sites'],pathlib.Path('/tmp/baobab-followup-push.log'))
sha=git('rev-parse','HEAD');checkpoint=git('rev-parse','typescript/sourcefile-like-sites-checkpoint')
remote=subprocess.check_output(['git','ls-remote','origin','refs/heads/typescript/sourcefile-like-sites','refs/heads/typescript/sourcefile-like-sites-checkpoint'],cwd=repo,text=True)
assert sha+'\trefs/heads/typescript/sourcefile-like-sites' in remote
assert checkpoint+'\trefs/heads/typescript/sourcefile-like-sites-checkpoint' in remote
result={'sha':sha,'checkpoint':checkpoint,'branch':'typescript/sourcefile-like-sites','variants':json.loads((follow/'variant-summary.json').read_text()),'baseline_audit':json.loads((out/'audit-result.json').read_text())}
pathlib.Path('/workspace/scratch/baobab-followup-published.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)
