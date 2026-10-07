#!/usr/bin/env python3
"""Independent Go regexp versus the native instruction VM."""
import itertools,json,os,subprocess,sys
from pathlib import Path
repo=Path(__file__).resolve().parents[4]; source=Path(__file__).resolve().parent; d=Path(sys.argv[1]);d.mkdir(exist_ok=True)
def run(name,args,expected=0):
 with (d/(name+'.stdout')).open('wb') as out,(d/(name+'.stderr')).open('wb') as err: p=subprocess.run([str(a) for a in args],cwd=repo,stdout=out,stderr=err)
 assert p.returncode==expected,(name,p.returncode)
 return (d/(name+'.stdout')).read_bytes()
patterns=[r'\Afoo\z',r'^foo$',r'.',r'(?s).',r'(?i)^foo$',r'^[^_]+$',r'(?i)k',r'(?i)σ',r'(?i)[a-z]',r'\p{Greek}+',r'\pL+',r'\P{L}',r'\bfoo\b',r'\Bfoo\B',r'(?m)^foo$',r'a{0,3}',r'(ab|a)*b',r'(a?)*',r'[^a-z]',r'\d+',r'\s',r'\w',r'a+?',r'\x{1F30D}',r'(?i)ſ',r'(?U)a+',r'(?P<name>foo)',r'\A\z',r'[[:alpha:]]+',r'a\n?b',r'a{2,4}',r'(?:foo|bar)$',r'(?m)\Afoo\z',r'\Q.$[]\E','']
texts=['','foo','FOO','foo\n','a','ab','aaab','baab','afoo','foo_bar',' foo ','xfooX','\r','\n','\t','Σ','σ','ς','K','k','K','S','s','ſ','世界','🌍','.$[]','ab\nfoo\nb','12','_']
rows=list(itertools.product(patterns,texts));(d/'inputs.json').write_text(json.dumps(rows,ensure_ascii=False))
def frame(text): return str(len(text.encode('utf-16-le'))//2)+'\n'+text
(d/'inputs.frames').write_text(frame(str(len(rows)))+''.join(frame(x)+frame(y) for x,y in rows))
(d/'anchor.a').write_text('export {};\n');(d/'config.json').write_text('{"compilerOptions":{"strict":true,"target":"ESNext"},"files":["anchor.a"]}')
compiler=d.parent/'wave29-regex-controls/adamic';archive=d.parent/'wave29-regex-controls/checker.a';asan=d.parent/'wave29-regex-controls/asan-checker.a'
run('build',[compiler,'build',source/'regex_runner.a','-o',d/'native','--tsgo',archive])
truth=run('go',['go','run',source/'testdata/regexp_oracle.go',d/'inputs.json'])
got=run('native',[d/'native',d/'config.json',d/'anchor.a',d/'inputs.frames'])
assert truth==got,'native regex byte mismatch';print(len(rows),'independent regex decisions agree')
run('asan-build',[compiler,'build',source/'regex_runner.a','-o',d/'asan','--tsgo',asan,'--sanitize'])
os.environ['ASAN_OPTIONS']='detect_leaks=1:halt_on_error=1';os.environ['UBSAN_OPTIONS']='halt_on_error=1'
assert run('asan',[d/'asan',d/'config.json',d/'anchor.a',d/'inputs.frames'])==truth
assert not (d/'asan.stderr').read_bytes();print('regex ASan/UBSan/LSan agree, empty stderr')
# Mutate only the regex VM context gate; syntax facts and the Go oracle are unchanged.
text=(source.parent/'regexp_program.a').read_text();before='if((context & instruction.argument) === instruction.argument)';assert text.count(before)==1
text=text.replace(before,'if(context >= 0)')
for dep in ['rules.ts','frames.ts']: text=text.replace("'./"+dep+"'","'"+str(source.parent/dep)+"'")
(d/'regexp_mutant.a').write_text(text)
entry=(source/'regex_runner.a').read_text().replace("'../regexp_program.a'","'"+str(d/'regexp_mutant.a')+"'")
for dep in ['unary_minus.ts','rules.ts','frames.ts','id_match.a','id_denylist.a']:entry=entry.replace("'../"+dep+"'","'"+str(source.parent/dep)+"'")
entry=entry.replace("'../../../typescript/","'"+str(repo/'stage1/typescript')+'/');(d/'mutant_runner.a').write_text(entry)
run('mutant-build',[compiler,'build',d/'mutant_runner.a','-o',d/'mutant','--tsgo',archive])
assert run('mutant',[d/'mutant',d/'config.json',d/'anchor.a',d/'inputs.frames'])!=truth
assert not (d/'mutant.stderr').read_bytes();print('VM context mutant compiled, exit 0, empty stderr, killed only by bytes')
