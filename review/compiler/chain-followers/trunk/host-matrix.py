import pathlib,subprocess,json,os
root=pathlib.Path.cwd();out=root/'review/compiler/chain-followers/trunk';scratch=pathlib.Path('/tmp/chain-followers-host-matrix');scratch.mkdir(exist_ok=True)
files=list((root/'stage3/fixtures/host').glob('*.a'))
files += [root/'stage3/fixtures'/p for p in ['assertions/16_unknown_chain.a','assertions/17_brand_upcast.a','cycles/04_directory_callback/main.a','cycles/05_safe_load_read/main.a','cycles/06_import_order_mutant/main.a','namespaces/02_jsx_names.a','namespaces/03_react_names.a']]
results=[]
def run(args,label):
 r=subprocess.run(args,cwd=root if args[0].startswith('/tmp/chain-followers-trunk-') else scratch,stdout=subprocess.PIPE,stderr=subprocess.PIPE,timeout=90,env={**os.environ,'ASAN_OPTIONS':'detect_leaks=1:halt_on_error=1','UBSAN_OPTIONS':'halt_on_error=1'})
 (out/(label+'.stdout')).write_bytes(r.stdout);(out/(label+'.stderr')).write_bytes(r.stderr);return r
for i,p in enumerate(files):
 node=run(['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(p)],f'host-{i}-node');base=run(['/tmp/chain-followers-trunk-base','js',str(p)],f'host-{i}-base');js=run(['/tmp/chain-followers-trunk-tip','js',str(p)],f'host-{i}-jsbuild');binary=scratch/f'{i}.bin';built=run(['/tmp/chain-followers-trunk-tip','build',str(p),'-o',str(binary),'--sanitize'],f'host-{i}-build');row={'fixture':str(p.relative_to(root)),'node_exit':node.returncode,'main_exit':base.returncode,'tip_js_exit':js.returncode,'tip_native_build_exit':built.returncode}
 assert (js.returncode==0)==(built.returncode==0),row
 if js.returncode==0:
  script=scratch/f'{i}.mjs';script.write_bytes(js.stdout);backend=run(['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(script)],f'host-{i}-js');native=run([str(binary)],f'host-{i}-native');row['backend_exit']=backend.returncode;row['native_exit']=native.returncode
  if True:
   assert (node.returncode,node.stdout,node.stderr)==(backend.returncode,backend.stdout,backend.stderr)==(native.returncode,native.stdout,native.stderr),row;row['agrees']=True
 results.append(row);(out/'host-matrix.json').write_text(json.dumps(results,indent=2));print(row,flush=True)
