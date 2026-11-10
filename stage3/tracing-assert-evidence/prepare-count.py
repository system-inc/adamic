from pathlib import Path
import json,shutil,subprocess
r=Path('/workspace/adamic');s=Path('/workspace/scratch/baobab-evidence')
prepared=json.loads(Path('/tmp/baobab-prepare-shards.log').read_text())
source=Path(prepared['tree']);product=s/'instrumented';tree=product/'tree'
shutil.copytree(source,tree,symlinks=True)
subprocess.run(['node',str(s/'instrument.cjs'),str(tree)],check=True,stdout=(s/'instrumented-files.json').open('w'))
p=r/'stage3/lane/adaptedtree/adaptedtree.go';text=p.read_text();start=text.index('func Product(t testing.TB) string {');end=text.index('\n// build makes',start)
replacement=text[:start]+'func Product(t testing.TB) string { return '+json.dumps(str(product))+' }\n'+text[end:]
# Only the measurement copy substitutes the temporary counted tree for the product.
overlay=s/'lane-product.go.txt';overlay.write_text(replacement)
(s/'lane-overlay.json').write_text(json.dumps({'Replace':{str(p):str(overlay)}},indent=2)+'\n')
(s/'lane-counts').mkdir();(s/'control-counts').mkdir()
(s/'prepared.json').write_text(json.dumps(prepared,indent=2)+'\n')
print(tree)
toolchain=r/'internal/toolchain/toolchain.go';text=toolchain.read_text();start=text.index('func Env(extra ...string) []string {');end=text.index('\n// GitEnv',start)
body=text[start:end];needle='return append(environment, extra...)';assert body.count(needle)==1
body=body.replace(needle,'environment = append(environment, '+json.dumps('NODE_OPTIONS=--require='+str(s/'preload.cjs'))+', '+json.dumps('BAOBAB_COUNTER_DIR='+str(s/'lane-counts'))+')\n\t'+needle)
file=s/'lane-environment.go.txt';file.write_text(text[:start]+body+text[end:])
overlay=json.loads((s/'lane-overlay.json').read_text());overlay['Replace'][str(toolchain)]=str(file);(s/'lane-overlay.json').write_text(json.dumps(overlay,indent=2)+'\n')
