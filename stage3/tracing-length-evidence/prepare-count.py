from pathlib import Path
from run import raw,root,run
import json,shutil
source=Path('/workspace/scratch/tracing-length-after');product=raw/'counted-product';tree=product/'tree'
assert (source/'built/local/run.js').is_file()
if not tree.exists():shutil.copytree(source,tree,symlinks=True,ignore=shutil.ignore_patterns('node_modules'))
files=sorted(str(p.relative_to(source)) for p in (source/'src').rglob('*') if p.is_file())
assert all((source/n).read_bytes()==(tree/n).read_bytes() for n in files)
js=sorted(p.name for p in (source/'built/local').glob('*.js'))
assert all((source/'built/local'/n).read_bytes()==(tree/'built/local'/n).read_bytes() for n in js)
run('count-instrument',['node',raw/'instrument.cjs',tree])
p=root/'stage3/lane/adaptedtree/adaptedtree.go';text=p.read_text();start=text.index('func Product(t testing.TB) string {');end=text.index('\n// build makes',start)
overlay=raw/'lane-product.go.txt';overlay.write_text(text[:start]+'func Product(t testing.TB) string { return '+json.dumps(str(product))+' }\n'+text[end:])
replace={str(p):str(overlay)}
p=root/'internal/toolchain/toolchain.go';text=p.read_text();start=text.index('func Env(extra ...string) []string {');end=text.index('\n// GitEnv',start);body=text[start:end]
needle='return append(environment, extra...)';assert body.count(needle)==1
body=body.replace(needle,'environment = append(environment, '+json.dumps('NODE_OPTIONS=--require='+str(raw/'preload.cjs'))+', '+json.dumps('TRACING_LENGTH_COUNTER_DIR='+str(raw/'lane-counts'))+')\n\t'+needle)
overlay=raw/'lane-environment.go.txt';overlay.write_text(text[:start]+body+text[end:]);replace[str(p)]=str(overlay)
(raw/'lane-overlay.json').write_text(json.dumps({'Replace':replace},indent=2)+'\n');(raw/'lane-counts').mkdir()
(raw/'counted-source.json').write_text(json.dumps({'sourceFilesIdentical':len(files),'normalJavaScriptFilesIdenticalBeforeInstrumentation':len(js),'privateProduct':str(product)},indent=2)+'\n')
print(tree)
