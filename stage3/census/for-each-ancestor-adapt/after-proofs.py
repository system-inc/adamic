from pathlib import Path
import subprocess,sys,json,shutil,re
out=Path(__file__).resolve().parent;repo=Path('/workspace/ancestor-adapt');tree=Path('/workspace/scratch/ancestor-production/after');before=Path('/workspace/scratch/ancestor/baseline');scratch=tree.parent
assert json.loads((out/'apply.status.json').read_text())['child_exit']==0
link=tree/'node_modules';assert not link.exists();link.symlink_to(repo/'stage3/typescript/node_modules',target_is_directory=True)
def run(label,args,expected=0):
 code=subprocess.run([sys.executable,str(out/'run.py'),str(out/label),*map(str,args)],cwd=repo).returncode
 print(label+' wrapper child exit='+str(code),flush=True)
 if code!=expected:raise RuntimeError(label+': unexpected child exit '+str(code))
run('replay',['env','NODE_PATH='+str(repo/'stage3/api/node_modules'),'node',out/'replay.cjs',tree,out/'replay.json'])
run('existing-pins',['env','NODE_PATH='+str(repo/'stage3/api/node_modules'),'node',out/'check-adapter-pins.cjs',before,tree,out/'existing-pins.json'])
run('types-after-driver',[sys.executable,out/'check-types.py',tree,'types-after'],expected=1)
a=(out/'types-before-extended.log').read_text().replace(str(before),'TREE');b=(out/'types-after-extended.log').read_text().replace(str(tree),'TREE');assert a==b
assert json.loads((out/'types-after.status.json').read_text())['child_exit']==0
(out/'extra-adamic-extended-comparison.json').write_text(json.dumps({'before_child_exit':1,'after_child_exit':1,'diagnostics_identical_after_root_normalization':True,'diagnostics':len(re.findall(r': error TS\d+:',a)),'scope':'extra 80-root Adamic probe; primary 79-root check is clean'},indent=2)+'\n')
run('census-after',['env','LATENT_FULL=1','LATENT_ONLY_FILE='+str(tree/'src/compiler/path.ts'),'LATENT_ASSERT_NO_OUTPUT=1',scratch/'latent',tree/'src/compiler',out/'census-after.jsonl'])
mutant=scratch/'mutant';shutil.copytree(tree/'src',mutant/'src');pins=json.loads((repo/'stage3/adapt/99-correctness-for-each-ancestor/owners.json').read_text())[0];p=mutant/'src/compiler/path.ts';raw=p.read_bytes();after=pins['after'].replace('\n','\r\n').encode() if b'\r\n' in raw else pins['after'].encode();original=pins['before'].replace('\n','\r\n').encode() if b'\r\n' in raw else pins['before'].encode();assert raw.count(after)==1;p.write_bytes(raw.replace(after,original));assert p.read_bytes()==(before/'src/compiler/path.ts').read_bytes()
run('census-mutant',['env','LATENT_FULL=1','LATENT_ONLY_FILE='+str(p),'LATENT_ASSERT_NO_OUTPUT=1',scratch/'latent',mutant/'src/compiler',out/'census-mutant.jsonl'])
run('census-delta',[sys.executable,out/'compare-census.py'])
print('After proof sequence completed',flush=True)
