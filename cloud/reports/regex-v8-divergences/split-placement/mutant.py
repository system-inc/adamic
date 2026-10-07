#!/usr/bin/env python3
"""Prove the split witness catches code-point candidate scanning, with clean ASan exits."""
import pathlib, subprocess, os
root=pathlib.Path.cwd();work=pathlib.Path('/tmp/regex-placement-mutant');work.mkdir(exist_ok=True)
runtime=root/'internal/native/runtime';text=(runtime/'regexp.c').read_bytes()
cache=pathlib.Path(os.environ.get('XDG_CACHE_HOME') or pathlib.Path.home()/'.cache')/'adamic/runtime'
found=[]
for directory in cache.iterdir():
 p=directory/'regexp.c'
 if p.exists() and p.read_bytes()==text and (directory/'regexp.o').exists():
  n=subprocess.run(['nm',str(directory/'regexp.o')],capture_output=True)
  if b'__asan' in n.stdout:found.append(directory)
assert found, 'no current sanitized runtime cache'
objects=found[0]
flags=['-std=c11','-Wall','-Wextra','-Werror','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-O1','-g','-fsanitize=address,undefined','-fno-sanitize-recover=all']
env=dict(os.environ,ASAN_OPTIONS='detect_leaks=1:abort_on_error=1:halt_on_error=1',UBSAN_OPTIONS='halt_on_error=1:abort_on_error=1')
def run(args):
 p=subprocess.run(list(map(str,args)),capture_output=True,env=env,timeout=90)
 assert p.returncode==0 and not p.stderr,(args,p.returncode,p.stderr)
 return p.stdout
source=root/'internal/oracle/testdata/regexp_surrogate_methods.a';c=work/'fixture.c';c.write_bytes(run(['/tmp/regex-placement-adamic','c',source]))
node=work/'fixture.mts';node.write_bytes(source.read_bytes());expected=run(['node','--disable-warning=ExperimentalWarning',node])
others=sorted(p for p in objects.glob('*.o') if p.name!='regexp.o')
binary=work/'normal';run(['clang',*flags,'-I',runtime,c,objects/'regexp.o',*others,'-lm','-o',binary]);assert run([binary])==expected
old='at = default_limit || (p->flags & 64) ? at + 1\n\t\t\t\t: regex_advance(units, length, at, p->flags & 4);'
s=text.decode();assert s.count(old)==1;s=s.replace(old,'at = regex_advance(units, length, at, p->flags & 4);')
mutant=work/'regexp.c';mutant.write_text(s);obj=work/'mutant.o';run(['clang',*flags,'-I',runtime,'-c',mutant,'-o',obj]);binary=work/'mutant';run(['clang',*flags,'-I',runtime,c,obj,*others,'-lm','-o',binary]);actual=run([binary]);assert actual!=expected
report=root/'cloud/reports/regex-v8-divergences/split-placement';(report/'mutant-node.stdout').write_bytes(expected);(report/'mutant-native.stdout').write_bytes(actual)
print('Normal native and Node agree. Mutant advances failed split candidates by code points: exit 0, empty sanitizer stderr, stdout differs from Node.')
