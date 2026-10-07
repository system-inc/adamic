#!/usr/bin/env python3
"""Isolated C mutants: require clean executions and only a Node stdout difference."""
import os,pathlib,subprocess,sys
root=pathlib.Path(__file__).resolve().parents[3]
compiler=pathlib.Path(os.environ['ADAMIC_REGEX_MUTANT_COMPILER'])
objects=pathlib.Path(os.environ['ADAMIC_REGEX_MUTANT_OBJECTS'])
work=pathlib.Path(os.environ['ADAMIC_REGEX_MUTANT_WORK']);work.mkdir(parents=True,exist_ok=True)
include=root/'internal/native/runtime'
flags=['-std=c11','-Wall','-Wextra','-Werror','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-O1','-g','-fsanitize=address,undefined','-fno-sanitize-recover=all']
env=dict(os.environ,ASAN_OPTIONS='detect_leaks=1:abort_on_error=1:halt_on_error=1',UBSAN_OPTIONS='halt_on_error=1:abort_on_error=1')
def command(args):
 p=subprocess.run(list(map(str,args)),capture_output=True,env=env,timeout=60)
 if p.returncode or p.stderr:raise RuntimeError(f'{args}: exit {p.returncode}, {p.stderr.decode(errors="replace")}')
 return p.stdout
runtime=(include/'regexp.c').read_text()
def changed(text,old,new):
 assert text.count(old)==1,(old,text.count(old));return text.replace(old,new)
def function_change(name,old,new):
 at=runtime.index(name);return runtime[:at]+changed(runtime[at:],old,new)
# Scoped changes use function boundaries to avoid changing other methods.
def scoped(name,end,old,new):
 a=runtime.index(name);b=runtime.index(end,a) if end else len(runtime)
 return runtime[:a]+changed(runtime[a:b],old,new)+runtime[b:]
variants={
 'protocol':scoped('double adamic_regex_search(', 'adamic_array *adamic_regex_split(', 'return result;', 'return result + 1;'),
 'string':scoped('adamic_string *adamic_regex_to_string(',None,'adamic_string_concat(4,','adamic_string_concat(3,'),
 'replacement_callback':scoped('adamic_string *adamic_regex_replace_callback(', '// The proven intrinsic object', 'adamic_string *result = adamic_array_join(pieces, &adamic_string_empty, adamic_join_strings);','adamic_string *result = adamic_retain(input);'),
 'compile':scoped('adamic_object *adamic_regex_compile(', 'static const adamic_regex_program *regex_program', 'regex->slots[1].number = 0;', '/* mutant: retain lastIndex */'),
 'surrogate_methods':scoped('adamic_array *adamic_regex_split(', 'static void regex_piece(', 'at = default_limit || (p->flags & 64) ? at + 1\n\t\t\t\t: regex_advance(units, length, at, p->flags & 4);', 'at = regex_advance(units, length, at, p->flags & 4);'),
}
normal=work/'normal.o';command(['clang',*flags,'-I',include,'-c',include/'regexp.c','-o',normal])
others=sorted(p for p in objects.glob('*.o') if p.name!='regexp.o');assert len(others)>20
for name in ['errors','protocol','string','replacement_callback','compile','metadata','surrogate_methods']:
 source=root/'internal/oracle/testdata'/f'regexp_{name}.a';c=work/f'{name}.c';c.write_bytes(command([compiler,'c',source]))
 node=work/f'{name}.mts';node.write_bytes(source.read_bytes());expected=command(['node','--disable-warning=ExperimentalWarning',node]);(work/f'{name}.node.stdout').write_bytes(expected)
 binary=work/f'{name}.normal';command(['clang',*flags,'-I',include,c,normal,*others,'-lm','-o',binary]);assert command([binary])==expected,name
 if name=='errors':
  text=c.read_text();assert 'adamic_builtin_error_new(2,' in text;c.write_text(text.replace('adamic_builtin_error_new(2,','adamic_builtin_error_new(1,'));obj=normal
 elif name=='metadata':
  text=c.read_text();old='adamic_string_from_number((0x0p+00))';assert old in text;c.write_text(text.replace(old,'adamic_string_from_number((0x1p+00))',1));obj=normal
 else:
  src=work/f'{name}.mutant.c';src.write_text(variants[name]);obj=work/f'{name}.mutant.o';command(['clang',*flags,'-I',include,'-c',src,'-o',obj])
 binary=work/f'{name}.mutant';command(['clang',*flags,'-I',include,c,obj,*others,'-lm','-o',binary]);actual=command([binary]);(work/f'{name}.mutant.stdout').write_bytes(actual);assert actual!=expected,name
 print(f'{name}: normal and Node agree; mutant exits 0, empty stderr, Node stdout comparison catches it',flush=True)
