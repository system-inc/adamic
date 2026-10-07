import sys,shutil,subprocess
from pathlib import Path
base=Path(sys.argv[1]);target=Path(sys.argv[2]);target.mkdir(exist_ok=True)
reference=next((arg for arg in sys.argv[3:] if not arg.startswith('--')),None)
for p in base.iterdir():
 if p.suffix in ('.c','.h','.ts') or p.name in ('oracle','compiler.txt'):shutil.copy2(p,target/p.name)
for name in ('heap.c','string_build_impl.h','string_search_impl.h','string_slice_impl.h'):
 p=Path('/workspace/adamic/internal/native/runtime')/name
 if reference:
  data=subprocess.check_output(['git','-C','/workspace/adamic','show',reference+':internal/native/runtime/'+name]);(target/name).write_bytes(data)
 else:shutil.copy2(p,target/name)
flags=['-std=c11','-Wall','-Wextra','-Werror','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls']
units=[str(target/'main.c')]+[str(p) for p in sorted(target.glob('*.c')) if p.name!='main.c']
builds=[('scanner',['-O2','-g']),('counted',['-O2','-DADAMIC_COUNT'])]
if '--sanitize' in sys.argv:builds.append(('sanitized',['-O1','-g','-fsanitize=address,undefined','-fno-sanitize-recover=all']))
for name,opts in builds:
 with (target/(name+'-build.log')).open('wb') as log:subprocess.run(['clang',*flags,*opts,'-o',str(target/name),*units,'-lm'],stdout=log,stderr=log,check=True)
shutil.copy2(target/'scanner',target/'profiled')
