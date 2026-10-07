from pathlib import Path
import subprocess, os, json
root=Path('/workspace/adamic'); scratch=Path('/workspace/utf16-measure');directory=scratch/'fixtures'; directory.mkdir(exist_ok=True)
env=os.environ.copy();env['VALGRIND_LIB']=str(scratch/'valgrind/usr/libexec/valgrind')
flags=['-std=c11','-Wall','-Wextra','-Werror','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-O2','-g']
rows=[]
for name in ['strings','strings_more','string_positions','string_index','string_append','shared_slices','lone_surrogates']:
 source=directory/(name+'.c')
 with source.open('wb') as out, (directory/(name+'-emit.log')).open('wb') as err:
  subprocess.run([str(scratch/'adamic'),'c',str(root/'internal/oracle/testdata'/(name+'.a'))],cwd=root,stdout=out,stderr=err,check=True)
 row={'fixture':name}
 answers=[]
 for snapshot in ['before','final']:
  runtime=scratch/snapshot;binary=directory/(name+'-'+snapshot);log=directory/(name+'-'+snapshot+'.log')
  with log.open('wb') as err:
   subprocess.run(['clang',*flags,'-I',str(runtime),str(source),*map(str,sorted(p for p in runtime.glob('*.c') if p.name!='main.c')),'-lm','-o',str(binary)],stdout=err,stderr=err,check=True)
  profile=directory/(name+'-'+snapshot+'.callgrind')
  with (directory/(name+'-'+snapshot+'.stdout')).open('wb') as out,log.open('ab') as err:
   subprocess.run([str(scratch/'valgrind/usr/bin/valgrind'),'--tool=callgrind','--callgrind-out-file='+str(profile),str(binary)],stdout=out,stderr=err,env=env,check=True)
  row[snapshot]=next(int(line.split()[1]) for line in profile.read_text().splitlines() if line.startswith('summary:'))
  answers.append((directory/(name+'-'+snapshot+'.stdout')).read_bytes())
 if answers[0]!=answers[1]:raise RuntimeError(name+' answers differ')
 row['change_percent']=100*(row['final']/row['before']-1);rows.append(row);print(row,flush=True)
(directory/'measurements.json').write_text(json.dumps(rows,indent=2)+'\n')
