import sys,subprocess,shutil,json,shlex,hashlib
from pathlib import Path
clang='/workspace/adamic-tools/llvm/bin/clang'
common=['-std=c11','-Wall','-Wextra','-Werror','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls']
cases=[('before8','both8'),('before4','both4'),('after8','split8'),('after4','split4'),('original8','flags-original8')]
records=[]
for name,source in cases:
 src=Path('/workspace/lint-'+source);dst=Path('/workspace/lint-flags-'+name);dst.mkdir(exist_ok=True)
 if dst!=src:
  for p in src.iterdir():
   if p.suffix in ('.c','.h','.ts') or p.name in ('oracle','compiler.txt'):shutil.copy2(p,dst/p.name)
 units=['main.c']+sorted(p.name for p in dst.glob('*.c') if p.name!='main.c')
 record={'case':name,'cwd':str(dst),'clang_version':subprocess.check_output([clang,'--version']).decode(),'commands':{},'input_sha256':{p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in dst.iterdir() if p.suffix in ('.c','.h')},'go_build_metadata':subprocess.check_output(['go','version','-m',str(dst/'oracle')]).decode()}
 for label,flags in [('scanner',['-O2','-g']),('sanitized',['-O1','-g','-fsanitize=address,undefined','-fno-sanitize-recover=all'])]:
  argv=[clang,*common,*flags,'-o',label,*units,'-lm'];record['commands'][label]={'argv':argv,'shell':shlex.join(argv)}
  with (dst/(label+'-build.log')).open('wb') as log:subprocess.run(argv,cwd=dst,stdout=log,stderr=log,check=True)
 records.append(record)
 Path('/workspace/lint-flags-builds.json').write_text(json.dumps(records,indent=2)+'\n')
 print('built',name,flush=True)
Path('/workspace/lint-flags-commands.txt').write_text('\n\n'.join(r['case']+'\n'+r['clang_version']+'cd '+shlex.quote(r['cwd'])+'\n'+'\n'.join(v['shell'] for v in r['commands'].values())+'\n'+r['go_build_metadata'] for r in records))
