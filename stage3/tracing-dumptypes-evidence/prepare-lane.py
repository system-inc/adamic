from pathlib import Path
import shutil,subprocess,json
S=Path(__file__).resolve().parent;R=Path('/workspace/adamic');base=Path('/workspace/scratch/dp24r5m-base');product=S/'counted-product';tree=product/'tree'
shutil.copytree(base,tree,symlinks=True,ignore=lambda directory,entries: [name for name in ['node_modules','.git'] if name in entries] if Path(directory)==base else [])
shutil.copytree(base/'.git',tree/'.git',symlinks=True)
with (S/'instrumented-files.log').open('w') as log:
 p=subprocess.run(['node',S/'instrument.cjs',tree/'built/local'],stdout=log,stderr=subprocess.STDOUT)
print('instrument child exit:',p.returncode,flush=True);assert p.returncode==0
p=R/'stage3/lane/adaptedtree/adaptedtree.go';text=p.read_text();a=text.index('func Product(t testing.TB) string {');b=text.index('\n// build makes',a)
f=S/'lane-product.go.txt';f.write_text(text[:a]+'func Product(t testing.TB) string { return '+json.dumps(str(product))+' }\n'+text[b:]);replace={str(p):str(f)}
p=R/'internal/toolchain/toolchain.go';text=p.read_text();a=text.index('func Env(extra ...string) []string {');b=text.index('\n// GitEnv',a);body=text[a:b];needle='return append(environment, extra...)';assert body.count(needle)==1
counts=S/'lane-counts';counts.mkdir();body=body.replace(needle,'environment = append(environment, '+json.dumps('NODE_OPTIONS=--require='+str(S/'preload.cjs'))+', '+json.dumps('DUMP_COUNTER_DIR='+str(counts))+')\n\t'+needle)
f=S/'lane-environment.go.txt';f.write_text(text[:a]+body+text[b:]);replace[str(p)]=str(f)
(S/'lane-overlay.json').write_text(json.dumps({'Replace':replace},indent=2)+'\n')
