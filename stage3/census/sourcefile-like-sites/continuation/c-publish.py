import pathlib,sys,json,hashlib,gzip,shutil,subprocess
p=pathlib.Path(__file__).resolve().parent;out=p.parent;repo=out.parents[2]
old='7944b20a23c7e2e40c00e925603c20bd7267a569'
assert sys.argv[1:]==['publish-after-review']
assert json.loads(pathlib.Path('/workspace/scratch/baobab-c-ready.json').read_text())['audit']['audit']=='passed'
def git(*args):return subprocess.check_output(['git',*args],cwd=repo,text=True).strip()
def invoke(args,label):
 log=pathlib.Path('/tmp/baobab-c-publish-'+label+'.log')
 with log.open('wb') as target:child=subprocess.run(args,cwd=repo,stdout=target,stderr=subprocess.STDOUT)
 record={'command':args,'child_exit':child.returncode};log.with_suffix('.status.json').write_text(json.dumps(record,indent=2)+'\n');print(json.dumps(record),flush=True)
 if child.returncode:raise SystemExit(child.returncode)
assert git('branch','--show-current')=='typescript/sourcefile-like-sites'
assert git('rev-parse','HEAD')==old
# All controllers have finished before their final logs are copied or sealed.
for log in pathlib.Path('/tmp').glob('baobab-c-*.log'):
 if log.name.startswith('baobab-c-publish-'):continue
 shutil.copyfile(log,p/'preparation'/log.name)
for file in sorted(out.rglob('*')):
 if not file.is_file() or file.suffix=='.gz' or file.stat().st_size<=1000000:continue
 packed=pathlib.Path(str(file)+'.gz')
 if packed.exists() and gzip.decompress(packed.read_bytes())==file.read_bytes():file.unlink();continue
 with file.open('rb') as source,gzip.open(packed,'wb') as target:shutil.copyfileobj(source,target)
 file.unlink()
manifest=out/'SHA256SUMS'
manifest.write_text(''.join(hashlib.sha256(f.read_bytes()).hexdigest()+'  '+f.relative_to(out).as_posix()+'\n' for f in sorted(out.rglob('*')) if f.is_file() and f!=manifest))
invoke(['python3',str(out/'checkpoint.py'),'named-LineMapSource-final'],'checkpoint')
paths=git('diff','--cached','--name-only').splitlines();assert paths and all(f.startswith('stage3/census/sourcefile-like-sites/') for f in paths)
assert not git('diff','--name-only')
tracked=set(git('ls-files',str(out.relative_to(repo))).splitlines());actual={str(f.relative_to(repo)) for f in out.rglob('*') if f.is_file()};assert tracked==actual
for line in manifest.read_text().splitlines():
 digest,name=line.split('  ',1);assert hashlib.sha256((out/name).read_bytes()).hexdigest()==digest,name
assert not [f for f in out.rglob('*') if f.is_file() and f.suffix!='.gz' and f.stat().st_size>1000000]
invoke(['git','commit','-m','census: measure named LineMapSource and public SDK boundary (baobab)'],'commit')
assert git('rev-parse','HEAD^')==old
assert git('rev-list','--count','d690f276749f91e6b36fe3a9726a5a453fcd7d39..HEAD')=='3'
assert not git('status','--porcelain')
invoke(['git','push','origin','HEAD:refs/heads/typescript/sourcefile-like-sites'],'push')
sha=git('rev-parse','HEAD');checkpoint=git('rev-parse','typescript/sourcefile-like-sites-checkpoint')
remote=git('ls-remote','origin','refs/heads/typescript/sourcefile-like-sites','refs/heads/typescript/sourcefile-like-sites-checkpoint')
assert sha+'\trefs/heads/typescript/sourcefile-like-sites' in remote
assert checkpoint+'\trefs/heads/typescript/sourcefile-like-sites-checkpoint' in remote
result={'sha':sha,'checkpoint':checkpoint,'branch':'typescript/sourcefile-like-sites','audit':json.loads((p/'c-audit-result.json').read_text())}
pathlib.Path('/workspace/scratch/baobab-c-published.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)
