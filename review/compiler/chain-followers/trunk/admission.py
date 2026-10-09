import pathlib,subprocess,json,os
root=pathlib.Path.cwd();out=root/'review/compiler/chain-followers/trunk';scratch=pathlib.Path('/tmp/chain-followers-trunk-admission');scratch.mkdir(exist_ok=True)
files=subprocess.check_output(['git','diff','--name-only','--diff-filter=A','origin/main','--','internal/oracle/testdata'],text=True).splitlines();results=[]
def run(args,label):
 r=subprocess.run(args,stdout=subprocess.PIPE,stderr=subprocess.PIPE,timeout=90,env={**os.environ,'ASAN_OPTIONS':'detect_leaks=1:halt_on_error=1','UBSAN_OPTIONS':'halt_on_error=1'})
 (out/(label+'.stdout')).write_bytes(r.stdout);(out/(label+'.stderr')).write_bytes(r.stderr)
 return r
for i,path in enumerate(files):
 p=str(root/path);base=run(['/tmp/chain-followers-trunk-base','js',p],f'admit-{i}-base');tip=run(['/tmp/chain-followers-trunk-tip','js',p],f'admit-{i}-tip');assert tip.returncode==0,(path,tip.stderr)
 row={'fixture':path,'main_exit':base.returncode,'tip_exit':tip.returncode,'new_admission':base.returncode!=0}
 if row['new_admission']:
  node=run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',p],f'admit-{i}-node');js=scratch/f'{i}.mjs';js.write_bytes(tip.stdout);backend=run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(js)],f'admit-{i}-js');binary=scratch/f'{i}.bin';built=run(['/tmp/chain-followers-trunk-tip','build',p,'-o',str(binary),'--sanitize'],f'admit-{i}-build');assert built.returncode==0,(path,built.stderr);native=run([str(binary)],f'admit-{i}-native');assert (node.returncode,node.stdout,node.stderr)==(backend.returncode,backend.stdout,backend.stderr)==(native.returncode,native.stdout,native.stderr),(path,node,backend,native)
  row['node_js_sanitized_native_agree']=True
 results.append(row);(out/'admission.json').write_text(json.dumps(results,indent=2));print(path,row,flush=True)
