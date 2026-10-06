#!/usr/bin/env python3
"""Run real mutations only in an isolated checkout and restore every source."""
import os,pathlib,subprocess,sys
root=pathlib.Path(sys.argv[1]).resolve()
if root==pathlib.Path.cwd().resolve(): raise SystemExit('Use an isolated checkout')
mutants=[
 ('backend-gate-omitted','internal/regexp/native.go','if err := p.NativeCompatibility(); err != nil {','if err := p.NativeCompatibility(); false && err != nil {','./internal/lower','^TestRegExpV8Refusals$','expected named V8/spec refusal'),
 ('singleton-fold-refusal-omitted','internal/regexp/v8.go','case *ClassString:\n\t\treturn compileClass(expr, flags, properties)','case *ClassString:\n\t\tset, err := compileClass(expr, flags, properties); return matchedCharacters(set, flags), err','./internal/regexp','^TestV8CompatibilityRefusals$','expected V8 compatibility refusal'),
 ('modifier-leak-refusal-omitted','internal/regexp/v8.go','parserFlags = flags','parserFlags = tree.Flags','./internal/regexp','^TestV8CompatibilityRefusals$','expected V8 compatibility refusal'),
 ('empty-mixed-refusal-omitted','internal/regexp/v8.go','if len(set.ranges) > 0 && slices.ContainsFunc','if false && len(set.ranges) > 0 && slices.ContainsFunc','./internal/regexp','^TestV8CompatibilityRefusals$','expected V8 compatibility refusal'),
 ('class-string-fold-omitted','internal/regexp/sets.go','copyS[j] = canonicalize(c, f)','copyS[j] = c','./internal/oracle','TestNativeAgreesWithNode/internal/oracle/testdata/regexp_v8_folding.a','stdout differs'),
 ('go-skips-pair-interior','internal/regexp/matcher.go','start = uint64(position + 1)','_, next, ok := readCharacter(input, position, 1, unicodeMode(p.flags)); if !ok { break }; start = uint64(next)','./internal/regexp','^TestV8SurrogateAssertionsNode$','DISAGREEMENT'),
 ('native-skips-pair-interior','internal/native/runtime/regexp.c','start = (size_t)at + 1;','uint32_t point; ptrdiff_t next; if (!regex_read(input, length, at, 1, p->flags & 4, &point, &next)) break; start = (size_t)next;','./internal/oracle','TestNativeAgreesWithNode/internal/oracle/testdata/regexp_v8_surrogates.a','stdout differs'),
 ('go-consumes-pair-half','internal/regexp/matcher.go','if unicode && pos > 0 && pos < len(input) && high(input[pos-1]) && low(input[pos]) {','if false && unicode && pos > 0 && pos < len(input) && high(input[pos-1]) && low(input[pos]) {','./internal/regexp','^TestV8SurrogateAssertionsNode$','DISAGREEMENT'),
 ('native-consumes-pair-half','internal/native/runtime/regexp.c','if (unicode && at > 0 && (size_t)at < length && high(input[at - 1]) && low(input[at]))','if (false && unicode && at > 0 && (size_t)at < length && high(input[at - 1]) && low(input[at]))','./internal/oracle','TestNativeAgreesWithNode/internal/oracle/testdata/regexp_v8_surrogates.a','stdout differs'),
 ('sticky-retry-omitted','internal/regexp/matcher.go','stickyEnd++','stickyEnd += 0','./internal/regexp','^TestV8SurrogateAssertionsNode$','DISAGREEMENT'),
]
for name,file,old,new,package,test,witness in mutants:
 if len(sys.argv)>2 and name not in sys.argv[2:]:continue
 target=root/file;original=target.read_text()
 if original.count(old)!=1:raise SystemExit(f'{name}: mutation site count {original.count(old)}')
 log=pathlib.Path('/tmp')/f'regex-v8-mutant-{name}.log'
 try:
  target.write_text(original.replace(old,new))
  with log.open('w') as output:
   r=subprocess.run(['go','test',package,'-run',test,'-count=1','-v','-parallel','2','-timeout','10m'],cwd=root,env=dict(os.environ,GOFLAGS='-buildvcs=false'),stdout=output,stderr=subprocess.STDOUT)
  found=[line for line in log.read_text().splitlines() if witness in line]
  if r.returncode==0 or not found:raise SystemExit(f'{name}: NOT CAUGHT exit={r.returncode}; {log}')
  print(f'{name}: CAUGHT exit={r.returncode}; {log}',flush=True);print(found[0],flush=True)
 finally:target.write_text(original)
