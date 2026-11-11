from control import *
name=sys.argv[1]
if name=='mutant':
 tree=root/'head'; adapted=root/'adapted-mutant'; binary=root/'head-census'
 subprocess.run(['cp','-a','--reflink=auto',str(root/'adapted-head'),str(adapted)],check=True)
 for p in (root/'adapted-head/src').rglob('*'):
  if p.is_file(): assert p.read_bytes()==(adapted/p.relative_to(root/'adapted-head')).read_bytes()
 patch=tree/'stage3/adapt/99-correctness-core-each-node/restoring-mutant.patch'
 before=(adapted/'src/compiler/debug.ts').read_bytes()
 (root/'debug-head.ts.txt').write_bytes(before)
 code=run('apply-mutant',['git','apply','--check',str(patch)],adapted)
 if code: sys.exit(code)
 code=run('apply-mutant-write',['git','apply',str(patch)],adapted)
 if code: sys.exit(code)
 # Reverse patch must restore the exact adapted bytes.
 check=subprocess.run(['git','apply','--reverse',str(patch)],cwd=adapted).returncode
 assert check==0 and (adapted/'src/compiler/debug.ts').read_bytes()==before
 subprocess.run(['git','apply',str(patch)],cwd=adapted,check=True)
 (root/'debug-mutant.ts.txt').write_bytes((adapted/'src/compiler/debug.ts').read_bytes())
 assert (adapted/'src/compiler/debug.ts').read_bytes()==(root/'adapted-parent/src/compiler/debug.ts').read_bytes(), 'restoring patch did not produce exact parent debug.ts bytes'
 checkpoint('apply-mutant')
else: tree=root/name; adapted=root/('adapted-'+name); binary=root/(name+'-census')
extra={'LATENT_FULL':'1','LATENT_ASSERT_NO_OUTPUT':'1','LATENT_ONLY_FILE':str(adapted/'src/compiler/debug.ts')}
code=run('debug-'+name,[binary,adapted/'src/tsc/tsc.ts',root/('debug-'+name+'.jsonl')],tree,extra)
checkpoint('debug-'+name)
if name=='mutant':
 subprocess.run(['git','apply','--reverse',str(patch)],cwd=adapted,check=True)
 assert (adapted/'src/compiler/debug.ts').read_bytes()==before
 sys.exit(code)
if code: sys.exit(code)
extra={'LATENT_FULL':'','LATENT_ASSERT_NO_OUTPUT':'1','LATENT_ONLY_FILE':''}
code=run('entry-'+name,[binary,adapted/'src/tsc/tsc.ts',root/('entry-'+name+'.jsonl')],tree,extra)
checkpoint('entry-'+name)
sys.exit(code)
