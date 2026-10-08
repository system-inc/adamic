#!/usr/bin/env python3
"""Kill each selector defect independently in native and JavaScript release runs."""
import pathlib,subprocess
root=pathlib.Path.cwd();lane=root/'stage3/interface-downcasts/lane4';logs=lane/'logs'
c=root/'internal/native/runtime/view_unions_mixed.c';js=root/'internal/javascript/view_unions_mixed.go'
original={p:p.read_bytes() for p in [c,js]}
mutants={
 'skip-member-check':{
 c:('if (value->kind == adamic_view_union_unknown || member->kind != value->kind) { continue; }','/* mutant: no member kind test */'),
 js:("if (snapshot.kind === 'unknown' || member.kind !== snapshot.kind) continue;",'/* mutant: no member kind test */')},
 'take-first-member':{
 c:('    for (size_t index = 0; index < count; index++) {','    return 0; /* mutant: first member without testing */\n    for (size_t index = 0; index < count; index++) {'),
 js:(' for (let index = 0; index < members.length; index++) {',' return 0; /* mutant: first member without testing */\n for (let index = 0; index < members.length; index++) {')},
 'drop-transitive-check':{
 c:('if (member->contract == 0 || match == NULL || !match(context, member, value)) { continue; }','/* mutant: skip object member contract */'),
 js:("if (!member.contract || match === undefined || !match(member, snapshot)) continue;",'/* mutant: skip object member contract */')},
}
try:
 for name,changes in mutants.items():
  for target,(before,after) in changes.items():
   text=original[target].decode();assert text.count(before)==1;(target).write_text(text.replace(before,after))
   backend='native' if target==c else 'javascript';log=logs/f'selection-mutant-{name}-{backend}.log'
   with log.open('wb') as output:
    result=subprocess.run(['go','test','./internal/oracle','-run','^TestCheckedViewMixedSelection$','-count=1','-v','-timeout','10m'],cwd=root,stdout=output,stderr=subprocess.STDOUT)
   observed=log.read_text()
   assert result.returncode!=0 and '--- FAIL: TestCheckedViewMixedSelection' in observed and 'native: clang failed' not in observed,(name,backend,observed)
   assert 'exitCode:0' in observed or 'stdout' in observed,(name,backend,observed)
   print(f'{name}/{backend}: caught by semantic output or pinned exit-70 assertion',flush=True)
   target.write_bytes(original[target])
finally:
 for target,data in original.items():target.write_bytes(data)
