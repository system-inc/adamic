from pathlib import Path
import subprocess,json,shlex,sys,hashlib
root=Path('/workspace/lint-cost-'+sys.argv[1]);clang='/workspace/adamic-tools/llvm/bin/clang'
flags=['-std=c11','-Wall','-Wextra','-Werror','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-O2','-g']
units=['main.c']+sorted(p.name for p in root.glob('*.c') if p.name!='main.c');cmd=[clang,*flags,'-o','scanner',*units,'-lm']
with (root/'build.log').open('wb') as f:subprocess.run(cmd,cwd=root,stdout=f,stderr=f,check=True)
(root/'build.json').write_text(json.dumps({'cwd':str(root),'command':cmd,'shell':shlex.join(cmd),'clang_version':subprocess.check_output([clang,'--version']).decode(),'sources':{p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in root.iterdir() if p.suffix in ['.c','.h']}},indent=2)+'\n')
