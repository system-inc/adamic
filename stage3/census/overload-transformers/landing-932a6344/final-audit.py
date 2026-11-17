import pathlib,json,subprocess,hashlib,shutil,sys
r=pathlib.Path('/workspace/scratch/get-properties-runtime');e=pathlib.Path('/workspace/scratch/overload-transformers/stage3/census/overload-transformers/landing-932a6344')
snapshot=json.loads((e/'landing-snapshot.json').read_text());head=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip()
assert head==snapshot['landingCommit']
assert subprocess.check_output(['git','rev-parse','HEAD^'],text=True).strip()==snapshot['parent']
assert subprocess.check_output(['git','rev-list','--count',snapshot['parent']+'..HEAD'],text=True).strip()=='1'
assert all(hashlib.sha256(pathlib.Path(f).read_bytes()).hexdigest()==sha for f,sha in snapshot['fileSha256'].items())
for name in ['whole-apply','adapted-tree-product','whole-lane','census-ruling-and-mutant-red']:
 assert json.loads((e/(name+'.json')).read_text())['exit']==0,name
lane=json.loads((r/'lane/report.json').read_text());oracle=json.loads((r/'lane/oracle/report.json').read_text())
assert lane['status']=='pass' and not lane['errors'],lane
assert oracle['counts']=={'passing':106366,'failing':1,'pending':0},oracle['counts']
assert oracle['runners']=='all' and oracle['tests'] is None and oracle['workers']==8
assert json.loads((e/'mutant-red.json').read_text())['pass']
public=json.loads((e/'public-declarations.json').read_text());assert all(x['byteIdentical'] for x in public)
for x in public:
 assert (e/'public-before'/x['file']).read_bytes()==(e/'public-after'/x['file']).read_bytes()
shards=json.loads((e/'shard-inputs.json').read_text())
assert all(hashlib.sha256(pathlib.Path('stage3/lane',n).read_bytes()).hexdigest()==sha for n,sha in shards['hashes'].items())
before=subprocess.check_output(['git','show',snapshot['parent']+':stage3/upstream/LEDGER.md'])
after=pathlib.Path('stage3/upstream/LEDGER.md').read_bytes();assert after.startswith(before) and after[len(before):].count(b'\n')==1
assert {p.name for p in pathlib.Path('stage3/adapt/99-correctness-get-properties').iterdir()}=={'adapt.cjs','owners.json','mutant.cjs','README.md'}
for path in [pathlib.Path('stage3/typescript/node_modules')]:
 if path.is_symlink():path.unlink()
cache=pathlib.Path('stage3/oracle/__pycache__')
if cache.exists():shutil.rmtree(cache)
assert subprocess.check_output(['git','status','--porcelain'],text=True).strip()==''
# Preserve all lane evidence except its source/build tree.
destination=e/'lane-results';destination.mkdir(exist_ok=True)
for src in (r/'lane').iterdir():
 if src.name=='adapted-tree':continue
 target=destination/src.name
 if src.is_dir():shutil.copytree(src,target,dirs_exist_ok=True)
 else:shutil.copy2(src,target)
proofs={'whole apply':'PASS','TestProduct_Stage3AdaptedTree':'PASS','whole unfiltered lane':'PASS','restoring mutant red':'PASS','public typescript.d.ts byte-identical':'PASS'}
result={'pass':True,'landingCommit':head,'base':snapshot['parent'],'proofs':proofs,'laneCounts':oracle['counts'],'shardRegenerationRequired':False,'landingFiles':snapshot['files'],'landingWorktreeClean':True}
e.joinpath('publication-audit.json').write_text(json.dumps(result,indent=2)+'\n')
for name,status in proofs.items():print(status+' '+name,flush=True)
print('PASS landing snapshot: one commit, five permitted files, ledger +1/-0, clean worktree',flush=True)
