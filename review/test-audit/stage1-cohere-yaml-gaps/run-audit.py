import pathlib,subprocess,os,time,json,difflib
p=pathlib.Path('/tmp/u152');root=pathlib.Path('/workspace/adamic');records=[]
prod=['TestLexerGaps','TestStructuralPositionRefusal','TestClosedStringPresenceGap','TestClosedLexerGaps','TestClosedValuePresenceGap','TestLexerMatchesGo','TestPropsMatchGo','TestSharedSliceAppendMatchesNode']
pattern='^('+'|'.join(prod)+')$'
env=os.environ.copy();env['ADAMIC_YAML_LIBRARY']=str(p/'library')
def run(id,cmd,environment=env):
 start=time.monotonic()
 with (p/(id+'.log')).open('w') as out:r=subprocess.run(cmd,env=environment,stdout=out,stderr=subprocess.STDOUT,cwd=root)
 rec={'id':id,'command':' '.join(cmd),'wall':round(time.monotonic()-start,3),'exit':r.returncode};records.append(rec);(p/'runs.json').write_text(json.dumps(records,indent=2));print(rec,flush=True);return r.returncode
def test(id,pat):
 e=env.copy();e['ADAMIC_BUILD_CACHE_DIR']=str(p/'cache'/id)
 return run(id,['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/yaml/','-run',pat],e)
def change(id,path,transform,vet=None,cfile=None,pat=pattern):
 f=root/path;old=f.read_text();new=transform(old);assert new!=old
 diff=''.join(difflib.unified_diff(old.splitlines(True),new.splitlines(True),fromfile='a/'+path,tofile='b/'+path))
 (p/(id+'.diff')).write_text(diff)
 assert run(id+'-apply-check',['git','apply','--check',str(p/(id+'.diff'))])==0
 try:
  f.write_text(new)
  if vet:assert run(id+'-vet',['timeout','90','go','vet',vet])==0
  if cfile:
   flags=['-std=c11','-Wall','-Wextra','-Werror','-Wcast-function-type-strict','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-O2']
   assert run(id+'-clang',['clang',*flags,'-I','internal/native/runtime','-c',cfile,'-o',str(p/(id+'.o'))])==0
  test(id,pat)
 finally:f.write_text(old)
def replace(a,b):
 def transform(s):assert s.count(a)==1,(a,s.count(a));return s.replace(a,b,1)
 return transform
change('M1','internal/lower/object.go',replace('if name == "push" {\n\t\tif len(arguments) != 1 {','if name == "push" {\n\t\tif len(arguments) == 1 {'),vet='./internal/lower/')
change('M2','internal/native/runtime/string_build_impl.h',replace('adamic_string adamic_string_true = ADAMIC_STRING("true");','adamic_string adamic_string_true = ADAMIC_STRING("TRUE");'),cfile='internal/native/runtime/string.c')
change('M3','internal/native/runtime/string_append.c',replace('\tresult->length = written;\n',''),cfile='internal/native/runtime/string_append.c')
change('M4','stage1/cohere/yaml/lexer.ts',replace('return unit === 44 || unit === 91','return unit === 45 || unit === 91'))
for id,file,field,testname in [('W1','lexer_test.go','output','TestLexerMutants'),('W2','props_test.go','out','TestPropsMutants'),('W3','scalar_test.go','out','TestScalarMutants')]:
 def transform(s,field=field,file=file):
  a=f'if bytes.Equal(side.{field}, expected) {{';assert s.count(a)==1
  s=s.replace(a,'if true {',1)
  if file=='scalar_test.go':s=s.replace('\t"bytes"\n','')
  return s
 change(id,'stage1/cohere/yaml/'+file,transform,vet='./stage1/cohere/yaml/',pat='^'+testname+'$')
def empty_lower(s):
 a=s.index('\n',s.index('func Lower('));b=s.index('\n}\n',a)
 return (s[:a]+'\n\treturn nil, nil'+s[b:]).replace('\t"fmt"\n','').replace('\t"path/filepath"\n','')
change('P1','internal/lower/lower.go',empty_lower,vet='./internal/lower/',pat='^('+ '|'.join(prod[:5]) +')$')
f=root/'internal/lower/lower.go';old=f.read_text()
try:
 f.write_text(empty_lower(old));test('P1-shared','^TestSharedSliceAppendMatchesNode$')
finally:f.write_text(old)
for id,driver,testname in [('P2','lex_main.ts','TestLexerMatchesGo'),('P3','props_main.ts','TestPropsMatchGo')]:
 def empty_main(s):return s[:s.index('const path = programArguments()')]
 change(id,'stage1/cohere/yaml/'+driver,empty_main,pat='^'+testname+'$')
print('AUDIT COMPLETE',flush=True)
