from pathlib import Path
import shutil,subprocess,os,json,hashlib
s=Path('/tmp/resolve-host-1612');tree=s/'build-mutant';tree.mkdir()
for p in (s/'build-change').iterdir():
 if p.is_file():shutil.copy2(p,tree/p.name)
shutil.copytree(s/'build-change/src',tree/'src');(tree/'node_modules').symlink_to(s/'TypeScript/node_modules',target_is_directory=True)
with (s/'evidence/restoring-mutant.log').open('w') as f:subprocess.run(['node',str(s/'work/stage3/adapt/99-wasm-crypto-import/resolve-host-mutant.cjs'),str(tree)],check=True,stdout=f,stderr=subprocess.STDOUT)
a=(tree/'src/compiler/sys.ts').read_bytes();b=(s/'base-sys.ts').read_bytes();assert a==b
(s/'evidence/restoring-mutant.json').write_text(json.dumps({'exactOriginalBytes':True,'SHA256':hashlib.sha256(a).hexdigest(),'preservesSiblingAnnotation':b': ModuleImportResult =>' in a},indent=2)+'\n')
env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',CENSUS_MODULE_FORMAT='CommonJSBundle',LATENT_ONLY_FILE=str(tree/'src/compiler/sys.ts'))
with (s/'evidence/census-mutant-accepted.log').open('w') as f:r=subprocess.run([str(s/'census'),str(tree/'src/compiler'),str(s/'evidence/census-mutant-accepted.jsonl')],env=env,stdout=f,stderr=subprocess.STDOUT)
assert r.returncode==0
