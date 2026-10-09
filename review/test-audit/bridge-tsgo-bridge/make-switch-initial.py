from pathlib import Path
import subprocess,json,difflib
root=Path('/workspace/adamic'); p=root/'review/test-audit/bridge-tsgo-bridge'; (p/'diffs').mkdir(exist_ok=True)
plan=[
('M1','production','internal/lower/tsgo.go','if !l.program.TSGoEnabled() {','if l.program.TSGoEnabled() {','flip condition'),
('M2','production','internal/lower/tsgo.go','\t\tdeclared.Body = append([]ir.Statement{ir.Panic{Message: ir.StringConstant{Index: message}}}, declared.Body...)\n','','drop statement'),
('M3','production','internal/native/tsgo.go','\t\t\treturn true\n','\t\t\treturn false\n','change constant'),
('C1','construction','bridge/tsgo/units_test.go','count := min(400, len(text))','count := min(399, len(text))','change constant'),
('C2','construction','bridge/tsgo/units_test.go','index%count == shard','index%(count+1) == shard','off-by-one bound'),
('C3','construction','bridge/tsgo/units_test.go','return file == "" || (file == "sample.ts")','return file == "" && (file == "sample.ts")','flip condition'),
('C4','construction','bridge/tsgo/product_cache_test.go','\tif err == nil {\n\t\terr = bridgeProductCheck(directory)\n\t}\n','','drop statement'),
('C5','construction','bridge/tsgo/product_cache_test.go','if len(hashes) != len(bridgeProductNames) {','if false && len(hashes) != len(bridgeProductNames) {','change boolean constant to disable count guard'),
('C6','construction','bridge/tsgo/product_cache_test.go','if hashes[name] != fmt.Sprintf("%x", sha256.Sum256(data)) {','if false && hashes[name] != fmt.Sprintf("%x", sha256.Sum256(data)) {','change boolean constant to disable digest guard'),
('P1','probe','internal/lower/lower.go','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {\n\tif true { return nil, nil }','empty Lower at entry'),
('P2','probe','internal/native/tsgo.go','func UsesTSGo(program *ir.Program) bool {','func UsesTSGo(program *ir.Program) bool {\n\tif true { return false }','empty UsesTSGo at entry'),
]
originals={f:subprocess.check_output(['git','show','HEAD:'+f],cwd=root,text=True) for _,_,f,*_ in plan}
entries=[]
for id,kind,f,before,after,menu in plan:
 original=originals[f]; assert original.count(before)==1,(id,original.count(before)); changed=original.replace(before,after,1); line=original[:original.index(before)].count('\n')+1
 diff=''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/'+f,tofile='b/'+f))
 (p/'diffs'/(id+'.diff')).write_text(diff)
 entries.append(dict(id=id,kind=kind,file=f,line=line,before=before,after=after,menu=menu))
(p/'plan.json').write_text(json.dumps(entries,indent=2)+'\n')
(p/'code-and-oracles.md').write_text('Starting commit: '+subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip()+'\n\nCODE UNDER TEST: Adamic Lower, lower.tsgo, native.UsesTSGo. Suite construction: bridgeShard, bridgeFileActive, bridgeCaseActive, bridgeRoots, bridgePositions, bridgeProductGet, bridgeProductCheck, bridgeCases, bridgeBindings, bridgeProductNames. See reached-functions.txt for the 302 covered production functions, including load preparation and package initializers. Test files are not instrumented by Go production coverage; the construction function list comes from reading units_test.go, registered_units_test.go and product_cache_test.go.\n\nORACLE: self for all three rows. Linkage checks only non-nil error for the unlinked case, nil error and UsesTSGo for the linked case, with no diagnostic identity assertion. Coverage compares self-written counts, names, positions and shard ownership. Cache compares self-written build count/path equality and non-nil errors for corrupt bytes and extra hash entries; it does not check error identity.\n\nMutant plan fixed before mutant outcomes: three production changes and six construction edits permitted by setup-check exception. The two probes are separate, never verdict evidence. Construction rows have no single production entry, so their vacuity field stays null. Matrix narrowed to requested rows after clean whole package timed out building sanitizer archive. No family member expansion is needed for these three standalone tests.\n')
# Write switched scratch source, with same selector everywhere.
for f,original in originals.items():
 changed=original
 for id,kind,path,before,after,menu in plan:
  if path != f: continue
  condition='os.Getenv("ADAMIC_MUTANT") == "'+id+'"'
  if id=='M1': replacement='if (!l.program.TSGoEnabled()) != ('+condition+') {'
  elif id=='M2': replacement='\t\tif !('+condition+') {\n'+before+'\t\t}\n'
  elif id=='M3': replacement='\t\t\treturn !('+condition+')\n'
  elif id=='C1': replacement='count := min(400, len(text))\n\tif '+condition+' { count = min(399, len(text)) }'
  elif id=='C2': replacement='index%(count + func() int { if '+condition+' { return 1 }; return 0 }()) == shard'
  elif id=='C3': replacement='return func() bool { if '+condition+' { return file == "" && (file == "sample.ts") == (os.Getenv("ADAMIC_TSGO_CORPUS") == "") }; return file == "" || (file == "sample.ts") == (os.Getenv("ADAMIC_TSGO_CORPUS") == "") }() //'
  elif id=='C4': replacement='\tif !('+condition+') {\n'+before+'\t}\n'
  elif id=='C5': replacement='if !('+condition+') && len(hashes) != len(bridgeProductNames) {'
  elif id=='C6': replacement='if !('+condition+') && hashes[name] != fmt.Sprintf("%x", sha256.Sum256(data)) {'
  elif id in ('P1','P2'): replacement=after.replace('if true', 'if '+condition)
  assert changed.count(before)==1,(id,'switch target')
  changed=changed.replace(before,replacement,1)
 if 'os.Getenv' in changed and '"os"' not in changed: changed=changed.replace('import (','import (\n\t"os"',1)
 (root/f).write_text(changed)
subprocess.run(['gofmt','-w',*originals.keys()],cwd=root,check=True)
(p/'scratch-switch.diff').write_text(subprocess.check_output(['git','diff','--',*originals.keys()],cwd=root,text=True))
print('Saved plan, 11 standalone diffs and scratch switch')
